package experiment

import (
	"fmt"
	"testing"

	base "github.com/feronski-bkpk/protoc-gen-go/benchmarks/experiment/protocols"
	dest "github.com/feronski-bkpk/protoc-gen-go/benchmarks/experiment/protocols"
	frag "github.com/feronski-bkpk/protoc-gen-go/benchmarks/experiment/protocols"
	hop "github.com/feronski-bkpk/protoc-gen-go/benchmarks/experiment/protocols"
	payload "github.com/feronski-bkpk/protoc-gen-go/benchmarks/experiment/protocols"
	routing "github.com/feronski-bkpk/protoc-gen-go/benchmarks/experiment/protocols"
)

func BuildIPv6Packet(numExtensions int, payloadSize int) ([]byte, error) {
	baseHdr := base.Ipv6BaseHeader{
		Hop_limit: 64,
	}
	baseHdr.SetVersion(6)

	extChain := []struct {
		nextHeader uint8
		protoType  uint8
	}{
		{0, 0},
		{44, 44},
		{60, 60},
		{43, 43},
	}

	switch numExtensions {
	case 0:
		baseHdr.Next_header = 59
	case 1:
		baseHdr.Next_header = extChain[0].protoType
	case 2:
		baseHdr.Next_header = extChain[0].protoType
		extChain[0].nextHeader = extChain[1].protoType
	case 3:
		baseHdr.Next_header = extChain[0].protoType
		extChain[0].nextHeader = extChain[1].protoType
		extChain[1].nextHeader = extChain[2].protoType
	case 4:
		baseHdr.Next_header = extChain[0].protoType
		extChain[0].nextHeader = extChain[1].protoType
		extChain[1].nextHeader = extChain[2].protoType
		extChain[2].nextHeader = extChain[3].protoType
	}

	headerSize := 40
	extHeaders := make([][]byte, 0, numExtensions)

	for i := 0; i < numExtensions; i++ {
		var hdrBytes []byte
		var err error

		switch extChain[i].protoType {
		case 0:
			h := hop.Ipv6HopByHop{
				Next_header: extChain[i].nextHeader,
			}
			hdrBytes, err = h.MarshalBinary()
		case 44:
			h := frag.Ipv6Fragment{
				Next_header:    extChain[i].nextHeader,
				Identification: 1,
			}
			hdrBytes, err = h.MarshalBinary()
		case 60:
			h := dest.Ipv6DestOpts{
				Next_header: extChain[i].nextHeader,
			}
			hdrBytes, err = h.MarshalBinary()
		case 43:
			h := routing.Ipv6Routing{
				Next_header: extChain[i].nextHeader,
			}
			hdrBytes, err = h.MarshalBinary()
		}

		if err != nil {
			return nil, fmt.Errorf("extension %d marshal: %w", i, err)
		}
		headerSize += len(hdrBytes)
		extHeaders = append(extHeaders, hdrBytes)
	}

	p := payload.Payload{
		Data_len: uint16(payloadSize),
		Data:     make([]byte, payloadSize),
	}
	payloadBytes, err := p.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("payload marshal: %w", err)
	}

	totalPayloadSize := headerSize - 40 + len(payloadBytes)
	baseHdr.Payload_length = uint16(totalPayloadSize)

	baseBytes, err := baseHdr.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("base marshal: %w", err)
	}

	packet := make([]byte, 0, 40+totalPayloadSize)
	packet = append(packet, baseBytes...)
	for _, ext := range extHeaders {
		packet = append(packet, ext...)
	}
	packet = append(packet, payloadBytes...)

	return packet, nil
}

func ParseIPv6Packet(data []byte, numExtensions int) error {
	offset := 0

	var baseHdr base.Ipv6BaseHeader
	if err := baseHdr.UnmarshalBinary(data[offset:]); err != nil {
		return fmt.Errorf("base unmarshal: %w", err)
	}
	offset += 40

	extTypes := []uint8{0, 44, 60, 43}
	extSizes := []int{8, 8, 8, 8}

	for i := 0; i < numExtensions; i++ {
		var err error
		switch extTypes[i] {
		case 0:
			var h hop.Ipv6HopByHop
			err = h.UnmarshalBinary(data[offset:])
		case 44:
			var h frag.Ipv6Fragment
			err = h.UnmarshalBinary(data[offset:])
		case 60:
			var h dest.Ipv6DestOpts
			err = h.UnmarshalBinary(data[offset:])
		case 43:
			var h routing.Ipv6Routing
			err = h.UnmarshalBinary(data[offset:])
		}
		if err != nil {
			return fmt.Errorf("extension %d unmarshal: %w", i, err)
		}
		offset += extSizes[i]
	}

	return nil
}

func BenchmarkBuild_64B_0ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(0, 64)
	}
}

func BenchmarkBuild_256B_2ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(2, 256)
	}
}

func BenchmarkBuild_1024B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(4, 1024)
	}
}

func BenchmarkBuild_4096B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(4, 4096)
	}
}

func BenchmarkBuild_64KB_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(4, 65536)
	}
}

func BenchmarkParse_64B_0ext(b *testing.B) {
	data, _ := BuildIPv6Packet(0, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 0)
	}
}

func BenchmarkParse_256B_2ext(b *testing.B) {
	data, _ := BuildIPv6Packet(2, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 2)
	}
}

func BenchmarkParse_1024B_4ext(b *testing.B) {
	data, _ := BuildIPv6Packet(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 4)
	}
}

func BenchmarkParse_4096B_4ext(b *testing.B) {
	data, _ := BuildIPv6Packet(4, 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 4)
	}
}

func BenchmarkParse_64KB_4ext(b *testing.B) {
	data, _ := BuildIPv6Packet(4, 65536)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 4)
	}
}
