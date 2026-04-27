package parser

import (
	"testing"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLexer_HexNumber(t *testing.T) {
	lexer := NewLexer("0x1234")
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)
	require.Len(t, tokens, 1)
	assert.Equal(t, TokenHexNumber, tokens[0].Type)
	assert.Equal(t, "0x1234", tokens[0].Value)
}

func TestLexer_Keywords(t *testing.T) {
	lexer := NewLexer("protocol struct bitstruct enum alias id if length_from length endian")
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)
	assert.Equal(t, 10, len(tokens))
	assert.Equal(t, TokenProtocol, tokens[0].Type)
	assert.Equal(t, TokenStruct, tokens[1].Type)
	assert.Equal(t, TokenBitStruct, tokens[2].Type)
	assert.Equal(t, TokenEnum, tokens[3].Type)
	assert.Equal(t, TokenIdent, tokens[4].Type) // alias
	assert.Equal(t, TokenIdent, tokens[5].Type) // id
	assert.Equal(t, TokenIf, tokens[6].Type)
	assert.Equal(t, TokenLengthFrom, tokens[7].Type)
	assert.Equal(t, TokenLength, tokens[8].Type)
	assert.Equal(t, TokenIdent, tokens[9].Type) // endian
}

func TestLexer_Operators(t *testing.T) {
	lexer := NewLexer("== != < > <= >=")
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)
	assert.Equal(t, TokenEq, tokens[0].Type)
	assert.Equal(t, TokenNotEq, tokens[1].Type)
	assert.Equal(t, TokenLT, tokens[2].Type)
	assert.Equal(t, TokenGT, tokens[3].Type)
	assert.Equal(t, TokenLTE, tokens[4].Type)
	assert.Equal(t, TokenGTE, tokens[5].Type)
}

func TestLexer_Comment(t *testing.T) {
	lexer := NewLexer("field // comment\nfield2")
	tokens, err := lexer.Tokenize()
	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Equal(t, "field", tokens[0].Value)
	assert.Equal(t, "field2", tokens[1].Value)
}

func TestParser_SimpleProtocol(t *testing.T) {
	input := `
protocol TestProtocol {
    id: 0x1234
    field1: uint16
    field2: int32
}`
	proto, err := ParseString(input)
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
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 1)

	structField := proto.Fields[0].(*ast.StructField)
	assert.Equal(t, "location", structField.Name)
	assert.Len(t, structField.Struct.Fields, 2)
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
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 1)

	bitField := proto.Fields[0].(*ast.BitStructField)
	assert.Equal(t, "flags", bitField.Name)
	assert.Len(t, bitField.Fields, 4)

	assert.Equal(t, "ack", bitField.Fields[0].Name)
	assert.Equal(t, 7, bitField.Fields[0].Bit)
	assert.False(t, bitField.Fields[0].IsRange)

	assert.Equal(t, "priority", bitField.Fields[2].Name)
	assert.True(t, bitField.Fields[2].IsRange)
	assert.Equal(t, 5, bitField.Fields[2].HighBit)
	assert.Equal(t, 4, bitField.Fields[2].LowBit)
}

func TestParser_BytesField(t *testing.T) {
	input := `
protocol Packet {
    id: 0x4000
    data_len: uint16
    data: bytes length_from: data_len
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	bytesField := proto.Fields[1].(*ast.BytesField)
	assert.Equal(t, "data", bytesField.Name)
	assert.Equal(t, "data_len", bytesField.LengthFrom)
}

func TestParser_ConditionalField(t *testing.T) {
	input := `
protocol ConditionalPacket {
    id: 0x5000
    flags: uint8
    extended: uint32 if flags == 1
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	condField := proto.Fields[1].(*ast.ScalarField)
	assert.Equal(t, "extended", condField.Name)
	require.NotNil(t, condField.Condition)
	assert.Equal(t, "flags", condField.Condition.Field)
	assert.Equal(t, "==", condField.Condition.Operator)
	assert.Equal(t, uint64(1), condField.Condition.Value)
}

func TestParser_FixedArray(t *testing.T) {
	input := `
protocol FixedArrayTest {
    id: 0x8000
    values: [10]float32
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 1)

	arr := proto.Fields[0].(*ast.ArrayField)
	assert.Equal(t, "values", arr.Name)
	assert.Equal(t, 10, arr.FixedLength)
	assert.Equal(t, ast.FLOAT32, arr.ElementType.(*ast.ScalarField).Type)
}

func TestParser_SliceScalar(t *testing.T) {
	input := `
protocol SliceTest {
    id: 0x9000
    count: uint16
    values: []float32 length: count
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	sliceField := proto.Fields[1].(*ast.ArrayField)
	assert.Equal(t, "values", sliceField.Name)
	assert.Equal(t, 0, sliceField.FixedLength)
	assert.Equal(t, "count", sliceField.LengthFrom)
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
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 2)

	sliceField := proto.Fields[1].(*ast.ArrayField)
	assert.Equal(t, "points", sliceField.Name)
	assert.Equal(t, "count", sliceField.LengthFrom)
}

func TestParser_FixedArrayStruct(t *testing.T) {
	input := `
protocol FixedArrayStructTest {
    id: 0xB000
    points: [5]struct {
        x: float32
        y: float32
    }
}`
	proto, err := ParseString(input)
	require.NoError(t, err)
	assert.Len(t, proto.Fields, 1)

	arr := proto.Fields[0].(*ast.ArrayField)
	assert.Equal(t, "points", arr.Name)
	assert.Equal(t, 5, arr.FixedLength)

	structElem := arr.ElementType.(*ast.StructField)
	assert.Len(t, structElem.Struct.Fields, 2)
}
