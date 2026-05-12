package handwritten

import (
	"encoding/binary"
	"fmt"
)

// ============================================================
// Ручные структуры (эквивалент DSL-описаний)
// ============================================================

// Ipv6BaseHeader — 40 байт
type Ipv6BaseHeader struct {
	VersionTc     uint8
	TcFlow        uint8
	FlowLabelLow  uint16
	PayloadLength uint16
	NextHeader    uint8
	HopLimit      uint8
	SrcAddr       [16]uint8
	DstAddr       [16]uint8
}

func (h *Ipv6BaseHeader) Marshal() []byte {
	buf := make([]byte, 40)
	buf[0] = h.VersionTc
	buf[1] = h.TcFlow
	binary.BigEndian.PutUint16(buf[2:], h.FlowLabelLow)
	binary.BigEndian.PutUint16(buf[4:], h.PayloadLength)
	buf[6] = h.NextHeader
	buf[7] = h.HopLimit
	copy(buf[8:24], h.SrcAddr[:])
	copy(buf[24:40], h.DstAddr[:])
	return buf
}

func (h *Ipv6BaseHeader) Unmarshal(data []byte) {
	h.VersionTc = data[0]
	h.TcFlow = data[1]
	h.FlowLabelLow = binary.BigEndian.Uint16(data[2:])
	h.PayloadLength = binary.BigEndian.Uint16(data[4:])
	h.NextHeader = data[6]
	h.HopLimit = data[7]
	copy(h.SrcAddr[:], data[8:24])
	copy(h.DstAddr[:], data[24:40])
}

func (h *Ipv6BaseHeader) Size() int { return 40 }

func (h *Ipv6BaseHeader) GetVersion() uint8  { return h.VersionTc >> 4 }
func (h *Ipv6BaseHeader) SetVersion(v uint8) { h.VersionTc = (v << 4) | (h.VersionTc & 0x0F) }

type Ipv6HopByHop struct {
	NextHeader uint8
	HdrExtLen  uint8
	Options    [6]uint8
}

func (h *Ipv6HopByHop) Marshal() []byte {
	buf := make([]byte, 8)
	buf[0] = h.NextHeader
	buf[1] = h.HdrExtLen
	copy(buf[2:], h.Options[:])
	return buf
}

func (h *Ipv6HopByHop) Unmarshal(data []byte) {
	h.NextHeader = data[0]
	h.HdrExtLen = data[1]
	copy(h.Options[:], data[2:8])
}

type Ipv6Fragment struct {
	NextHeader     uint8
	Reserved       uint8
	OffsetFlags    uint8
	OffsetLow      uint8
	Identification uint32
}

func (h *Ipv6Fragment) Marshal() []byte {
	buf := make([]byte, 8)
	buf[0] = h.NextHeader
	buf[1] = h.Reserved
	buf[2] = h.OffsetFlags
	buf[3] = h.OffsetLow
	binary.BigEndian.PutUint32(buf[4:], h.Identification)
	return buf
}

func (h *Ipv6Fragment) Unmarshal(data []byte) {
	h.NextHeader = data[0]
	h.Reserved = data[1]
	h.OffsetFlags = data[2]
	h.OffsetLow = data[3]
	h.Identification = binary.BigEndian.Uint32(data[4:])
}

type Ipv6DestOpts struct {
	NextHeader uint8
	HdrExtLen  uint8
	Options    [6]uint8
}

func (h *Ipv6DestOpts) Marshal() []byte {
	buf := make([]byte, 8)
	buf[0] = h.NextHeader
	buf[1] = h.HdrExtLen
	copy(buf[2:], h.Options[:])
	return buf
}

func (h *Ipv6DestOpts) Unmarshal(data []byte) {
	h.NextHeader = data[0]
	h.HdrExtLen = data[1]
	copy(h.Options[:], data[2:8])
}

type Ipv6Routing struct {
	NextHeader   uint8
	HdrExtLen    uint8
	RoutingType  uint8
	SegmentsLeft uint8
	Reserved     [4]uint8
}

func (h *Ipv6Routing) Marshal() []byte {
	buf := make([]byte, 8)
	buf[0] = h.NextHeader
	buf[1] = h.HdrExtLen
	buf[2] = h.RoutingType
	buf[3] = h.SegmentsLeft
	copy(buf[4:], h.Reserved[:])
	return buf
}

func (h *Ipv6Routing) Unmarshal(data []byte) {
	h.NextHeader = data[0]
	h.HdrExtLen = data[1]
	h.RoutingType = data[2]
	h.SegmentsLeft = data[3]
	copy(h.Reserved[:], data[4:8])
}

type Payload struct {
	DataLen uint16
	Data    []byte
}

func (p *Payload) Marshal() []byte {
	buf := make([]byte, 2+len(p.Data))
	binary.BigEndian.PutUint16(buf[0:], p.DataLen)
	copy(buf[2:], p.Data)
	return buf
}

func (p *Payload) Unmarshal(data []byte) {
	p.DataLen = binary.BigEndian.Uint16(data[0:])
	p.Data = make([]byte, p.DataLen)
	copy(p.Data, data[2:2+p.DataLen])
}

// ============================================================
// Сборка и разбор (эквивалент BuildIPv6Packet / ParseIPv6Packet)
// ============================================================

func BuildPacket(numExt int, payloadSize int) []byte {
	extProto := []uint8{0, 44, 60, 43}

	base := Ipv6BaseHeader{
		HopLimit: 64,
	}
	base.SetVersion(6)

	if numExt == 0 {
		base.NextHeader = 59
	} else {
		base.NextHeader = extProto[0]
	}

	extHeaders := make([][]byte, 0, numExt)
	headerSize := 40

	for i := 0; i < numExt; i++ {
		nextHdr := uint8(59)
		if i < numExt-1 {
			nextHdr = extProto[i+1]
		}

		var hdrBytes []byte
		switch extProto[i] {
		case 0:
			h := Ipv6HopByHop{NextHeader: nextHdr}
			hdrBytes = h.Marshal()
		case 44:
			h := Ipv6Fragment{NextHeader: nextHdr, Identification: 1}
			hdrBytes = h.Marshal()
		case 60:
			h := Ipv6DestOpts{NextHeader: nextHdr}
			hdrBytes = h.Marshal()
		case 43:
			h := Ipv6Routing{NextHeader: nextHdr}
			hdrBytes = h.Marshal()
		}
		headerSize += len(hdrBytes)
		extHeaders = append(extHeaders, hdrBytes)
	}

	p := Payload{
		DataLen: uint16(payloadSize),
		Data:    make([]byte, payloadSize),
	}
	payloadBytes := p.Marshal()

	base.PayloadLength = uint16(headerSize - 40 + len(payloadBytes))
	baseBytes := base.Marshal()

	packet := make([]byte, 0, 40+int(base.PayloadLength))
	packet = append(packet, baseBytes...)
	for _, ext := range extHeaders {
		packet = append(packet, ext...)
	}
	packet = append(packet, payloadBytes...)

	return packet
}

func ParsePacket(data []byte, numExt int) error {
	if len(data) < 40 {
		return fmt.Errorf("data too short")
	}

	var base Ipv6BaseHeader
	base.Unmarshal(data[:40])
	offset := 40

	extProto := []uint8{0, 44, 60, 43}

	for i := 0; i < numExt; i++ {
		switch extProto[i] {
		case 0:
			var h Ipv6HopByHop
			h.Unmarshal(data[offset:])
		case 44:
			var h Ipv6Fragment
			h.Unmarshal(data[offset:])
		case 60:
			var h Ipv6DestOpts
			h.Unmarshal(data[offset:])
		case 43:
			var h Ipv6Routing
			h.Unmarshal(data[offset:])
		}
		offset += 8
	}

	var p Payload
	p.Unmarshal(data[offset:])

	return nil
}
