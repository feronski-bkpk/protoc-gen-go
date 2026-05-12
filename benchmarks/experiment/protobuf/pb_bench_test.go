package protobuf

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

// ============================================================
// Сборка пакета через Protobuf
// ============================================================

func buildPacketPB(numExt int, payloadSize int) []byte {
	extProto := []uint32{0, 44, 60, 43}

	base := &Ipv6BaseHeader{
		VersionTc:    0x60,
		TcFlow:       0,
		FlowLabelLow: 0,
		HopLimit:     64,
		SrcAddr:      []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		DstAddr:      []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	}
	if numExt > 0 {
		base.NextHeader = extProto[0]
	} else {
		base.NextHeader = 59
	}

	extHeaders := make([][]byte, 0, numExt)
	headerSize := 40

	for i := 0; i < numExt; i++ {
		nextHdr := uint32(59)
		if i < numExt-1 {
			nextHdr = extProto[i+1]
		}

		var msg proto.Message
		switch extProto[i] {
		case 0:
			msg = &Ipv6HopByHop{NextHeader: nextHdr, Options: make([]byte, 6)}
		case 44:
			msg = &Ipv6Fragment{NextHeader: nextHdr, Identification: 1}
		case 60:
			msg = &Ipv6DestOpts{NextHeader: nextHdr, Options: make([]byte, 6)}
		case 43:
			msg = &Ipv6Routing{NextHeader: nextHdr, Reserved: make([]byte, 4)}
		}

		hdrBytes, _ := proto.Marshal(msg)
		headerSize += len(hdrBytes)
		extHeaders = append(extHeaders, hdrBytes)
	}

	payloadMsg := &Payload{
		DataLen: uint32(payloadSize),
		Data:    make([]byte, payloadSize),
	}
	payloadBytes, _ := proto.Marshal(payloadMsg)

	base.PayloadLength = uint32(headerSize - 40 + len(payloadBytes))
	baseBytes, _ := proto.Marshal(base)

	packet := make([]byte, 0, 40+int(base.PayloadLength))
	packet = append(packet, baseBytes...)
	for _, ext := range extHeaders {
		packet = append(packet, ext...)
	}
	packet = append(packet, payloadBytes...)

	return packet
}

func parsePacketPB(data []byte, numExt int) error {
	extProto := []uint32{0, 44, 60, 43}

	var base Ipv6BaseHeader
	if err := proto.Unmarshal(data, &base); err != nil {
		return err
	}
	offset := len(data) - int(base.PayloadLength)

	for i := 0; i < numExt; i++ {
		var msg proto.Message
		switch extProto[i] {
		case 0:
			msg = &Ipv6HopByHop{}
		case 44:
			msg = &Ipv6Fragment{}
		case 60:
			msg = &Ipv6DestOpts{}
		case 43:
			msg = &Ipv6Routing{}
		}
		if err := proto.Unmarshal(data[offset:], msg); err != nil {
			return err
		}
		offset += proto.Size(msg)
	}

	var p Payload
	if err := proto.Unmarshal(data[offset:], &p); err != nil {
		return err
	}
	return nil
}

// ============================================================
// Эксперимент А
// ============================================================

func BenchmarkPBBuild_64B_0ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(0, 64)
	}
}
func BenchmarkPBParse_64B_0ext(b *testing.B) {
	data := buildPacketPB(0, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 0)
	}
}

func BenchmarkPBBuild_256B_2ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(2, 256)
	}
}
func BenchmarkPBParse_256B_2ext(b *testing.B) {
	data := buildPacketPB(2, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 2)
	}
}

func BenchmarkPBBuild_1024B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(4, 1024)
	}
}
func BenchmarkPBParse_1024B_4ext(b *testing.B) {
	data := buildPacketPB(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 4)
	}
}

func BenchmarkPBBuild_4096B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(4, 4096)
	}
}
func BenchmarkPBParse_4096B_4ext(b *testing.B) {
	data := buildPacketPB(4, 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 4)
	}
}

func BenchmarkPBBuild_64K_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(4, 65536-40-32)
	}
}
func BenchmarkPBParse_64K_4ext(b *testing.B) {
	data := buildPacketPB(4, 65536-40-32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 4)
	}
}

// ============================================================
// Эксперимент Б
// ============================================================

func BenchmarkPBBuild_1024B_Ext0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(0, 1024)
	}
}
func BenchmarkPBBuild_1024B_Ext1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(1, 1024)
	}
}
func BenchmarkPBBuild_1024B_Ext2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(2, 1024)
	}
}
func BenchmarkPBBuild_1024B_Ext3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(3, 1024)
	}
}
func BenchmarkPBBuild_1024B_Ext4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = buildPacketPB(4, 1024)
	}
}

func BenchmarkPBParse_1024B_Ext0(b *testing.B) {
	data := buildPacketPB(0, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 0)
	}
}
func BenchmarkPBParse_1024B_Ext1(b *testing.B) {
	data := buildPacketPB(1, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 1)
	}
}
func BenchmarkPBParse_1024B_Ext2(b *testing.B) {
	data := buildPacketPB(2, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 2)
	}
}
func BenchmarkPBParse_1024B_Ext3(b *testing.B) {
	data := buildPacketPB(3, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 3)
	}
}
func BenchmarkPBParse_1024B_Ext4(b *testing.B) {
	data := buildPacketPB(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parsePacketPB(data, 4)
	}
}
