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
}

func TestParser_BitFields(t *testing.T) {
	input := `
protocol BitTest {
    id: 0x6000
    flags: bitstruct {
        ack: bit(7)
        error: bit(6)
        priority: bits[5:4]
        reserved: bits[3:0]
    }
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 1)

	bitField := proto.Fields[0].(*ast.BitStructField)
	assert.Equal(t, "flags", bitField.Name)
	assert.Len(t, bitField.Fields, 4)

	ack := bitField.Fields[0]
	assert.Equal(t, "ack", ack.Name)
	assert.Equal(t, 7, ack.Bit)
	assert.False(t, ack.IsRange)

	priority := bitField.Fields[2]
	assert.Equal(t, "priority", priority.Name)
	assert.True(t, priority.IsRange)
	assert.Equal(t, 5, priority.HighBit)
	assert.Equal(t, 4, priority.LowBit)
}

func TestParser_ArraySlice(t *testing.T) {
	input := `
protocol ArrayTest {
    id: 0x7000
    values: []float32
    points: []struct {
        x: float32
        y: float32
    }
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	array1 := proto.Fields[0].(*ast.ArrayField)
	assert.Equal(t, "values", array1.Name)
	assert.Equal(t, 0, array1.FixedLength)

	scalarElem := array1.ElementType.(*ast.ScalarField)
	assert.Equal(t, ast.FLOAT32, scalarElem.Type)

	array2 := proto.Fields[1].(*ast.ArrayField)
	assert.Equal(t, "points", array2.Name)

	structElem := array2.ElementType.(*ast.StructField)
	assert.Len(t, structElem.Struct.Fields, 2)
}

func TestParser_SliceScalar(t *testing.T) {
	input := `
protocol SliceTest {
    id: 0x9000
    count: uint16
    values: []float32 length: count
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	countField := proto.Fields[0].(*ast.ScalarField)
	assert.Equal(t, "count", countField.Name)
	assert.Equal(t, ast.UINT16, countField.Type)

	sliceField := proto.Fields[1].(*ast.ArrayField)
	assert.Equal(t, "values", sliceField.Name)
	assert.Equal(t, 0, sliceField.FixedLength)
	assert.Equal(t, "count", sliceField.LengthFrom)

	scalarElem := sliceField.ElementType.(*ast.ScalarField)
	assert.Equal(t, ast.FLOAT32, scalarElem.Type)
}

func TestParser_SliceStruct(t *testing.T) {
	input := `
protocol SliceStructTest {
    id: 0xA000
    count: uint8
    points: []struct {
        x: float32
        y: float32
        z: float32
    } length: count
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	sliceField := proto.Fields[1].(*ast.ArrayField)
	assert.Equal(t, "points", sliceField.Name)
	assert.Equal(t, 0, sliceField.FixedLength)
	assert.Equal(t, "count", sliceField.LengthFrom)

	structElem := sliceField.ElementType.(*ast.StructField)
	assert.Len(t, structElem.Struct.Fields, 3)
}

func TestParser_SliceWithCondition(t *testing.T) {
	input := `
protocol SliceCondTest {
    id: 0xB000
    flags: uint8
    count: uint16
    data: []uint32 length: count if flags == 1
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 3)

	sliceField := proto.Fields[2].(*ast.ArrayField)
	assert.Equal(t, "data", sliceField.Name)
	assert.NotNil(t, sliceField.Condition)
	assert.Equal(t, "flags", sliceField.Condition.Field)
	assert.Equal(t, "==", sliceField.Condition.Operator)
	assert.Equal(t, uint64(1), sliceField.Condition.Value)
}
