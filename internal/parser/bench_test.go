package parser

import (
	"testing"
)

var complexDSL = `
protocol SensorData {
    id: 0xABCD
    device_id: uint32
    timestamp: uint64
    temperature: float32
    humidity: uint8
    pressure: uint32
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
        priority: bits[5:4]
        reserved: bits[3:0]
    }
    location: struct {
        latitude: float64
        longitude: float64
        altitude: int32
    }
    readings: [10]float32
    samples_len: uint16
    samples: []struct {
        x: float32
        y: float32
        z: float32
    } length: samples_len
    name_len: uint16
    name: bytes length_from: name_len
    error_msg_len: uint16
    error_msg: bytes length_from: error_msg_len if flags.error == 1
}
`

func BenchmarkLexer(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(complexDSL)
		lexer.Tokenize()
	}
}

func BenchmarkParser(b *testing.B) {
	lexer := NewLexer(complexDSL)
	tokens, _ := lexer.Tokenize()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser(tokens)
		p.Parse()
	}
}

func BenchmarkFullPipeline(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ParseString(complexDSL)
	}
}

func BenchmarkLexer_Simple(b *testing.B) {
	simple := `protocol Test { id: 0x1234 field: uint16 }`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		lexer := NewLexer(simple)
		lexer.Tokenize()
	}
}

func BenchmarkParser_Simple(b *testing.B) {
	simple := `protocol Test { id: 0x1234 field: uint16 }`
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ParseString(simple)
	}
}
