package golang

import (
	"testing"

	dns "github.com/feronski-bkpk/protoc-gen-go/benchmarks/protocols/dns"
	ipv4 "github.com/feronski-bkpk/protoc-gen-go/benchmarks/protocols/ipv4"
	sensor "github.com/feronski-bkpk/protoc-gen-go/benchmarks/protocols/sensor"
	tcp "github.com/feronski-bkpk/protoc-gen-go/benchmarks/protocols/tcp"
)

// --- Sensor Data ---

func BenchmarkSensorMarshal(b *testing.B) {
	p := sensor.SensorData{
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

func BenchmarkSensorUnmarshal(b *testing.B) {
	p := sensor.SensorData{
		Device_id:    12345,
		Temperature:  23.5,
		Humidity:     60.0,
		Readings_len: 10,
		Readings:     [10]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	data, _ := p.MarshalBinary()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored sensor.SensorData
		_ = restored.UnmarshalBinary(data)
	}
}

// --- TCP Header ---

func BenchmarkTCPMarshal(b *testing.B) {
	p := tcp.TcpHeader{
		Src_port:   443,
		Dst_port:   8080,
		Seq_num:    123456789,
		Ack_num:    987654321,
		Window:     65535,
		Checksum:   0xFFFF,
		Urgent_ptr: 0,
	}
	p.SetData_offset(5)
	p.SetSyn(true)
	p.SetAck(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.MarshalBinary()
	}
}

func BenchmarkTCPUnmarshal(b *testing.B) {
	p := tcp.TcpHeader{
		Src_port:   443,
		Dst_port:   8080,
		Seq_num:    123456789,
		Ack_num:    987654321,
		Window:     65535,
		Checksum:   0xFFFF,
		Urgent_ptr: 0,
	}
	p.SetData_offset(5)
	p.SetSyn(true)
	p.SetAck(true)
	data, _ := p.MarshalBinary()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored tcp.TcpHeader
		_ = restored.UnmarshalBinary(data)
	}
}

// --- DNS Message ---

func BenchmarkDNSMarshal(b *testing.B) {
	p := dns.DnsMessage{
		Transaction_id: 0xABCD,
		Qdcount:        1,
		Question_len:   32,
		Question:       make([]byte, 32),
	}
	p.SetQr(true)
	p.SetRd(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.MarshalBinary()
	}
}

func BenchmarkDNSUnmarshal(b *testing.B) {
	p := dns.DnsMessage{
		Transaction_id: 0xABCD,
		Qdcount:        1,
		Question_len:   32,
		Question:       make([]byte, 32),
	}
	p.SetQr(true)
	p.SetRd(true)
	data, _ := p.MarshalBinary()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored dns.DnsMessage
		_ = restored.UnmarshalBinary(data)
	}
}

// --- IPv4 Header ---

func BenchmarkIPv4Marshal(b *testing.B) {
	p := ipv4.Ipv4Header{
		Total_length:    1500,
		Identification:  12345,
		Ttl:             64,
		Proto:           6,
		Header_checksum: 0,
		Src_addr:        0xC0A80001,
		Dst_addr:        0xC0A80002,
	}
	p.SetVersion(4)
	p.SetIhl(5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.MarshalBinary()
	}
}

func BenchmarkIPv4Unmarshal(b *testing.B) {
	p := ipv4.Ipv4Header{
		Total_length:   1500,
		Identification: 12345,
		Ttl:            64,
		Proto:          6,
		Src_addr:       0xC0A80001,
		Dst_addr:       0xC0A80002,
	}
	p.SetVersion(4)
	p.SetIhl(5)
	data, _ := p.MarshalBinary()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored ipv4.Ipv4Header
		_ = restored.UnmarshalBinary(data)
	}
}
