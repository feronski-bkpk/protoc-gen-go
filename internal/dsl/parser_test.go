package dsl

import (
	"testing"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_SimpleProtocol(t *testing.T) {
	input := `
protocol TestProtocol {
    id: 0x1234
    field1: uint16
    field2: int32
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Equal(t, "TestProtocol", proto.Name)
	assert.Equal(t, uint16(0x1234), proto.PacketID)
	assert.Len(t, proto.Fields, 2)

	field1 := proto.Fields[0].(*ast.ScalarField)
	assert.Equal(t, "field1", field1.Name)
	assert.Equal(t, ast.UINT16, field1.Type)

	field2 := proto.Fields[1].(*ast.ScalarField)
	assert.Equal(t, "field2", field2.Name)
	assert.Equal(t, ast.INT32, field2.Type)
}

func TestParser_StructField(t *testing.T) {
	input := `
protocol GPSData {
    id: 0x1000
    location: struct {
        latitude: float64
        longitude: float64
    }
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 1)

	structField := proto.Fields[0].(*ast.StructField)
	assert.Equal(t, "location", structField.Name)
	assert.NotNil(t, structField.Struct)
	assert.Len(t, structField.Struct.Fields, 2)

	lat := structField.Struct.Fields[0].(*ast.ScalarField)
	assert.Equal(t, "latitude", lat.Name)
	assert.Equal(t, ast.FLOAT64, lat.Type)

	lon := structField.Struct.Fields[1].(*ast.ScalarField)
	assert.Equal(t, "longitude", lon.Name)
	assert.Equal(t, ast.FLOAT64, lon.Type)
}

func TestParser_MultipleFields(t *testing.T) {
	input := `
protocol ComplexProtocol {
    id: 0x2000
    sensor_id: uint16
    temperature: int16
    humidity: uint8
    pressure: float32
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Equal(t, "ComplexProtocol", proto.Name)
	assert.Len(t, proto.Fields, 4)
}

func TestParser_ConditionalField(t *testing.T) {
	input := `
protocol ConditionalPacket {
    id: 0x5000
    flags: uint8
    extended: uint32 if flags == 1
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	condField := proto.Fields[1].(*ast.ScalarField)
	assert.Equal(t, "extended", condField.Name)
	assert.NotNil(t, condField.Condition)
	assert.Equal(t, "flags", condField.Condition.Field)
	assert.Equal(t, "==", condField.Condition.Operator)
	assert.Equal(t, uint64(1), condField.Condition.Value)
}

func TestParser_NestedStruct(t *testing.T) {
	input := `
protocol NestedData {
    id: 0x6000
    header: struct {
        version: uint8
        flags: uint8
    }
    payload: struct {
        data: uint32
        checksum: uint16
    }
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	header := proto.Fields[0].(*ast.StructField)
	assert.Equal(t, "header", header.Name)
	assert.Len(t, header.Struct.Fields, 2)

	payload := proto.Fields[1].(*ast.StructField)
	assert.Equal(t, "payload", payload.Name)
	assert.Len(t, payload.Struct.Fields, 2)
}

func TestParser_BytesField(t *testing.T) {
	input := `
protocol Packet {
    id: 0x4000
    data_len: uint16
    data: bytes length_from: data_len
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	bytesField := proto.Fields[1].(*ast.BytesField)
	assert.Equal(t, "data", bytesField.Name)
	assert.Equal(t, "data_len", bytesField.LengthFrom)
}
