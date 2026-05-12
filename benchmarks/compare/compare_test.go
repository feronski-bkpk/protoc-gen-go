package compare

import (
	"testing"

	// Твой DSL
	ipv4_dsl "github.com/feronski-bkpk/protoc-gen-go/benchmarks/protocols/ipv4"
	sensor_dsl "github.com/feronski-bkpk/protoc-gen-go/benchmarks/protocols/sensor"

	// Protobuf
	ipv4_pb "github.com/feronski-bkpk/protoc-gen-go/benchmarks/compare/ipv4/pb/benchmarks/compare/ipv4"
	sensor_pb "github.com/feronski-bkpk/protoc-gen-go/benchmarks/compare/sensor/pb/benchmarks/compare/sensor"

	"google.golang.org/protobuf/proto"
)

// === SENSOR: DSL ===
func BenchmarkDSL_SensorMarshal(b *testing.B) {
	p := sensor_dsl.SensorData{
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

func BenchmarkDSL_SensorUnmarshal(b *testing.B) {
	p := sensor_dsl.SensorData{
		Device_id:    12345,
		Temperature:  23.5,
		Humidity:     60.0,
		Readings_len: 10,
		Readings:     [10]float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	data, _ := p.MarshalBinary()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored sensor_dsl.SensorData
		_ = restored.UnmarshalBinary(data)
	}
}

// === SENSOR: Protobuf ===
func BenchmarkPB_SensorMarshal(b *testing.B) {
	p := &sensor_pb.SensorData{
		DeviceId:    12345,
		Temperature: 23.5,
		Humidity:    60.0,
		ReadingsLen: 10,
		Readings:    []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(p)
	}
}

func BenchmarkPB_SensorUnmarshal(b *testing.B) {
	p := &sensor_pb.SensorData{
		DeviceId:    12345,make clean
tree
		Temperature: 23.5,
		Humidity:    60.0,
		ReadingsLen: 10,
		Readings:    []float32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	data, _ := proto.Marshal(p)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored sensor_pb.SensorData
		_ = proto.Unmarshal(data, &restored)
	}
}

// === IPv4: DSL ===
func BenchmarkDSL_IPv4Marshal(b *testing.B) {
	p := ipv4_dsl.Ipv4Header{
		Total_length:   1500,
		Identification: 12345,
		Ttl:            64,
		Proto:          6,
		Src_addr:       0xC0A80001,
		Dst_addr:       0xC0A80002,
	}
	p.SetVersion(4)
	p.SetIhl(5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.MarshalBinary()
	}
}

func BenchmarkDSL_IPv4Unmarshal(b *testing.B) {
	p := ipv4_dsl.Ipv4Header{
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
		var restored ipv4_dsl.Ipv4Header
		_ = restored.UnmarshalBinary(data)
	}
}

// === IPv4: Protobuf ===
func BenchmarkPB_IPv4Marshal(b *testing.B) {
	p := &ipv4_pb.Ipv4Header{
		VersionIhl:     0x45,
		TotalLength:    1500,
		Identification: 12345,
		Ttl:            64,
		Proto:          6,
		SrcAddr:        0xC0A80001,
		DstAddr:        0xC0A80002,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(p)
	}
}

func BenchmarkPB_IPv4Unmarshal(b *testing.B) {
	p := &ipv4_pb.Ipv4Header{
		VersionIhl:     0x45,
		TotalLength:    1500,
		Identification: 12345,
		Ttl:            64,
		Proto:          6,
		SrcAddr:        0xC0A80001,
		DstAddr:        0xC0A80002,
	}
	data, _ := proto.Marshal(p)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var restored ipv4_pb.Ipv4Header
		_ = proto.Unmarshal(data, &restored)
	}
}
