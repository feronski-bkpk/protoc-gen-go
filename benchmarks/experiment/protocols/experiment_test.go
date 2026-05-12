package protocols

import (
	"fmt"
	"testing"
)

// ============================================================
// Сборка IPv6 пакета с заданным количеством расширений
// numExtensions: 0-4
// payloadSize: размер полезной нагрузки в байтах
// ============================================================
func BuildIPv6Packet(numExtensions int, payloadSize int) ([]byte, error) {
	baseHdr := Ipv6BaseHeader{
		Hop_limit: 64,
	}
	baseHdr.SetVersion(6)

	extProto := []uint8{0, 44, 60, 43}

	if numExtensions == 0 {
		baseHdr.Next_header = 59
	} else {
		baseHdr.Next_header = extProto[0]
	}

	extHeaders := make([][]byte, 0, numExtensions)
	headerSize := 40

	for i := 0; i < numExtensions; i++ {
		nextHdr := uint8(59)
		if i < numExtensions-1 {
			nextHdr = extProto[i+1]
		}

		var hdrBytes []byte
		var err error

		switch extProto[i] {
		case 0:
			h := Ipv6HopByHop{
				Next_header: nextHdr,
			}
			hdrBytes, err = h.MarshalBinary()
		case 44:
			h := Ipv6Fragment{
				Next_header:    nextHdr,
				Identification: 1,
			}
			hdrBytes, err = h.MarshalBinary()
		case 60:
			h := Ipv6DestOpts{
				Next_header: nextHdr,
			}
			hdrBytes, err = h.MarshalBinary()
		case 43:
			h := Ipv6Routing{
				Next_header: nextHdr,
			}
			hdrBytes, err = h.MarshalBinary()
		}

		if err != nil {
			return nil, fmt.Errorf("extension %d: %w", i, err)
		}
		headerSize += len(hdrBytes)
		extHeaders = append(extHeaders, hdrBytes)
	}

	p := Payload{
		Data_len: uint16(payloadSize),
		Data:     make([]byte, payloadSize),
	}
	payloadBytes, err := p.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("payload: %w", err)
	}

	baseHdr.Payload_length = uint16(headerSize - 40 + len(payloadBytes))

	baseBytes, err := baseHdr.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}

	packet := make([]byte, 0, 40+int(baseHdr.Payload_length))
	packet = append(packet, baseBytes...)
	for _, ext := range extHeaders {
		packet = append(packet, ext...)
	}
	packet = append(packet, payloadBytes...)

	return packet, nil
}

// ============================================================
// Разбор IPv6 пакета
// ============================================================
func ParseIPv6Packet(data []byte, numExtensions int) error {
	offset := 0

	var baseHdr Ipv6BaseHeader
	if err := baseHdr.UnmarshalBinary(data[offset:]); err != nil {
		return fmt.Errorf("base: %w", err)
	}
	offset += 40

	extProto := []uint8{0, 44, 60, 43}
	extSize := []int{8, 8, 8, 8}

	for i := 0; i < numExtensions; i++ {
		var err error
		switch extProto[i] {
		case 0:
			var h Ipv6HopByHop
			err = h.UnmarshalBinary(data[offset:])
		case 44:
			var h Ipv6Fragment
			err = h.UnmarshalBinary(data[offset:])
		case 60:
			var h Ipv6DestOpts
			err = h.UnmarshalBinary(data[offset:])
		case 43:
			var h Ipv6Routing
			err = h.UnmarshalBinary(data[offset:])
		}
		if err != nil {
			return fmt.Errorf("ext %d: %w", i, err)
		}
		offset += extSize[i]
	}

	var p Payload
	if err := p.UnmarshalBinary(data[offset:]); err != nil {
		return fmt.Errorf("payload: %w", err)
	}

	return nil
}

// ============================================================
// БЕНЧМАРКИ: Эксперимент А — Варьирование размера payload
// ============================================================

func BenchmarkBuild_64B_0ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(0, 64)
	}
}
func BenchmarkParse_64B_0ext(b *testing.B) {
	data, _ := BuildIPv6Packet(0, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 0)
	}
}

func BenchmarkBuild_256B_2ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(2, 256)
	}
}
func BenchmarkParse_256B_2ext(b *testing.B) {
	data, _ := BuildIPv6Packet(2, 256)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 2)
	}
}

func BenchmarkBuild_1024B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(4, 1024)
	}
}
func BenchmarkParse_1024B_4ext(b *testing.B) {
	data, _ := BuildIPv6Packet(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 4)
	}
}

func BenchmarkBuild_4096B_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(4, 4096)
	}
}
func BenchmarkParse_4096B_4ext(b *testing.B) {
	data, _ := BuildIPv6Packet(4, 4096)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 4)
	}
}

func BenchmarkBuild_64K_4ext(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(4, 65536-40-32)
	}
}
func BenchmarkParse_64K_4ext(b *testing.B) {
	data, _ := BuildIPv6Packet(4, 65536-40-32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 4)
	}
}

// ============================================================
// БЕНЧМАРКИ: Эксперимент Б — Варьирование сложности (количества расширений)
// Фиксированный размер payload = 1024 байта
// ============================================================

func BenchmarkBuild_1024B_Ext0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(0, 1024)
	}
}
func BenchmarkBuild_1024B_Ext1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(1, 1024)
	}
}
func BenchmarkBuild_1024B_Ext2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(2, 1024)
	}
}
func BenchmarkBuild_1024B_Ext3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(3, 1024)
	}
}
func BenchmarkBuild_1024B_Ext4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = BuildIPv6Packet(4, 1024)
	}
}

func BenchmarkParse_1024B_Ext0(b *testing.B) {
	data, _ := BuildIPv6Packet(0, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 0)
	}
}
func BenchmarkParse_1024B_Ext1(b *testing.B) {
	data, _ := BuildIPv6Packet(1, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 1)
	}
}
func BenchmarkParse_1024B_Ext2(b *testing.B) {
	data, _ := BuildIPv6Packet(2, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 2)
	}
}
func BenchmarkParse_1024B_Ext3(b *testing.B) {
	data, _ := BuildIPv6Packet(3, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 3)
	}
}
func BenchmarkParse_1024B_Ext4(b *testing.B) {
	data, _ := BuildIPv6Packet(4, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseIPv6Packet(data, 4)
	}
}

// ============================================================
// Проверка корректности (Roundtrip)
// ============================================================
func TestRoundtrip(t *testing.T) {
	sizes := []int{0, 64, 256, 1024, 4096}
	extCounts := []int{0, 1, 2, 3, 4}

	for _, payloadSize := range sizes {
		for _, numExt := range extCounts {
			t.Run(fmt.Sprintf("payload=%d_ext=%d", payloadSize, numExt), func(t *testing.T) {
				data, err := BuildIPv6Packet(numExt, payloadSize)
				if err != nil {
					t.Fatalf("Build: %v", err)
				}
				if err := ParseIPv6Packet(data, numExt); err != nil {
					t.Fatalf("Parse: %v", err)
				}
				t.Logf("OK: size=%d bytes, extensions=%d", len(data), numExt)
			})
		}
	}
}
