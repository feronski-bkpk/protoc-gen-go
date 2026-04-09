// internal/dsl/parser_test.go
package dsl

import (
	"testing"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_SimpleProtocol(t *testing.T) {
	input := `
protocol TemperatureReading {
    id: 0x1234
    sensor_id: uint16
    temperature: int16
    battery: uint8
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Equal(t, "TemperatureReading", proto.Name)
	assert.Equal(t, uint16(0x1234), proto.PacketID)
	assert.Len(t, proto.Fields, 3)

	// Проверяем поля
	field1 := proto.Fields[0].(*ast.ScalarField)
	assert.Equal(t, "sensor_id", field1.Name)
	assert.Equal(t, ast.UINT16, field1.Type)

	field2 := proto.Fields[1].(*ast.ScalarField)
	assert.Equal(t, "temperature", field2.Name)
	assert.Equal(t, ast.INT16, field2.Type)

	field3 := proto.Fields[2].(*ast.ScalarField)
	assert.Equal(t, "battery", field3.Name)
	assert.Equal(t, ast.UINT8, field3.Type)
}

func TestParser_StructField(t *testing.T) {
	input := `
protocol GPSData {
    id: 0x1000
    location: struct {
        latitude: float64
        longitude: float64
        altitude: int32
    }
    satellites: uint8
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	// Проверяем вложенную структуру
	structField := proto.Fields[0].(*ast.StructField)
	assert.Equal(t, "location", structField.Name)
	assert.NotNil(t, structField.Struct)
	assert.Len(t, structField.Struct.Fields, 3)

	// Проверяем поля структуры
	lat := structField.Struct.Fields[0].(*ast.ScalarField)
	assert.Equal(t, "latitude", lat.Name)
	assert.Equal(t, ast.FLOAT64, lat.Type)

	lon := structField.Struct.Fields[1].(*ast.ScalarField)
	assert.Equal(t, "longitude", lon.Name)
	assert.Equal(t, ast.FLOAT64, lon.Type)

	alt := structField.Struct.Fields[2].(*ast.ScalarField)
	assert.Equal(t, "altitude", alt.Name)
	assert.Equal(t, ast.INT32, alt.Type)
}

func TestParser_ArrayField(t *testing.T) {
	input := `
protocol DataPacket {
    id: 0x2000
    samples: []struct {
        timestamp: uint32
        value: float32
    } length_from: samples_count
    samples_count: uint16
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	// Проверяем массив
	arrayField := proto.Fields[0].(*ast.ArrayField)
	assert.Equal(t, "samples", arrayField.Name)
	assert.Equal(t, "samples_count", arrayField.LengthFrom)

	// Проверяем элемент массива
	structField := arrayField.Element.(*ast.StructField)
	assert.NotNil(t, structField.Struct)
	assert.Len(t, structField.Struct.Fields, 2)
}

func TestParser_ConditionalField(t *testing.T) {
	input := `
protocol ConditionalPacket {
    id: 0x3000
    flags: uint8
    extended_data: bytes(length: 8) if flags == 1
}
`
	parser := NewParser()
	proto, err := parser.ParseString(input)

	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	// Проверяем условное поле
	bytesField := proto.Fields[1].(*ast.BytesField)
	assert.Equal(t, "extended_data", bytesField.Name)
	assert.NotNil(t, bytesField.Condition)
	assert.Equal(t, "flags", bytesField.Condition.Field)
	assert.Equal(t, "==", bytesField.Condition.Operator)
	assert.Equal(t, uint64(1), bytesField.Condition.Value)
}
