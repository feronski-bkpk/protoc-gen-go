package golang

import (
	"encoding/binary"
	"math"
	"testing"
)

// === Ручная реализация SensorData ===

type HandSensorData struct {
	Device_id    uint32
	Temperature  float32
	Humidity     float32
	Readings_len uint16
	Readings     [10]float32
}

func (p *HandSensorData) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 54)
	binary.BigEndian.PutUint32(buf[0:], p.Device_id)
	binary.BigEndian.PutUint32(buf[4:], math.Float32bits(p.Temperature))
	binary.BigEndian.PutUint32(buf[8:], math.Float32bits(p.Humidity))
	binary.BigEndian.PutUint16(buf[12:], p.Readings_len)
	for i := 0; i < 10; i++ {
		binary.BigEndian.PutUint32(buf[14+i*4:], math.Float32bits(p.Readings[i]))
	}
	return buf, nil
}

func (p *HandSensorData) UnmarshalBinary(data []byte) error {
	p.Device_id = binary.BigEndian.Uint32(data[0:])
	p.Temperature = math.Float32frombits(binary.BigEndian.Uint32(data[4:]))
	p.Humidity = math.Float32frombits(binary.BigEndian.Uint32(data[8:]))
	p.Readings_len = binary.BigEndian.Uint16(data[12:])
	for i := 0; i < 10; i++ {
		p.Readings[i] = math.Float32frombits(binary.BigEndian.Uint32(data[14+i*4:]))
	}
	return nil
}

func BenchmarkHandSensorMarshal(b *testing.B) {
	p := HandSensorData{
		Device_id:    12345,
		Temperature:  23.5,
		Humidity:     60.0,
		Readings_len: 10,
		Readings:     [10]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.MarshalBinary()
	}
}

func BenchmarkHandSensorUnmarshal(b *testing.B) {
	p := HandSensorData{
		Device_id:    12345,
		Temperature:  23.5,
		Humidity:     60.0,
		Readings_len: 10,
		Readings:     [10]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	data, _ := p.MarshalBinary()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored HandSensorData
		_ = restored.UnmarshalBinary(data)
	}
}

// === Ручная реализация IPv4 ===

type HandIPv4Header struct {
	VersionIhl       uint8
	DscpEcn          uint8
	TotalLength      uint16
	Identification   uint16
	FlagsFragment    uint8
	FragmentOffset   uint8
	Ttl              uint8
	Proto            uint8
	HeaderChecksum   uint16
	SrcAddr          uint32
	DstAddr          uint32
}

func (p *HandIPv4Header) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 20)
	buf[0] = p.VersionIhl
	buf[1] = p.DscpEcn
	binary.BigEndian.PutUint16(buf[2:], p.TotalLength)
	binary.BigEndian.PutUint16(buf[4:], p.Identification)
	buf[6] = p.FlagsFragment
	buf[7] = p.FragmentOffset
	buf[8] = p.Ttl
	buf[9] = p.Proto
	binary.BigEndian.PutUint16(buf[10:], p.HeaderChecksum)
	binary.BigEndian.PutUint32(buf[12:], p.SrcAddr)
	binary.BigEndian.PutUint32(buf[16:], p.DstAddr)
	return buf, nil
}

func (p *HandIPv4Header) UnmarshalBinary(data []byte) error {
	p.VersionIhl = data[0]
	p.DscpEcn = data[1]
	p.TotalLength = binary.BigEndian.Uint16(data[2:])
	p.Identification = binary.BigEndian.Uint16(data[4:])
	p.FlagsFragment = data[6]
	p.FragmentOffset = data[7]
	p.Ttl = data[8]
	p.Proto = data[9]
	p.HeaderChecksum = binary.BigEndian.Uint16(data[10:])
	p.SrcAddr = binary.BigEndian.Uint32(data[12:])
	p.DstAddr = binary.BigEndian.Uint32(data[16:])
	return nil
}

func BenchmarkHandIPv4Marshal(b *testing.B) {
	p := HandIPv4Header{
		VersionIhl:     0x45,
		DscpEcn:        0,
		TotalLength:    1500,
		Identification: 12345,
		Ttl:            64,
		Proto:          6,
		SrcAddr:        0xC0A80001,
		DstAddr:        0xC0A80002,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.MarshalBinary()
	}
}

func BenchmarkHandIPv4Unmarshal(b *testing.B) {
	p := HandIPv4Header{
		VersionIhl:     0x45,
		DscpEcn:        0,
		TotalLength:    1500,
		Identification: 12345,
		Ttl:            64,
		Proto:          6,
		SrcAddr:        0xC0A80001,
		DstAddr:        0xC0A80002,
	}
	data, _ := p.MarshalBinary()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored HandIPv4Header
		_ = restored.UnmarshalBinary(data)
	}
}
