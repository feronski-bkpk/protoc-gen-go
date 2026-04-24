package analyzer

import (
	"testing"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzer_SimpleProtocol(t *testing.T) {
	proto := &ast.Protocol{
		Name:     "Test",
		PacketID: 0x1234,
		Fields: []ast.Field{
			&ast.ScalarField{Name: "field1", Type: ast.UINT16},
			&ast.ScalarField{Name: "field2", Type: ast.UINT32},
		},
	}

	a := NewAnalyzer(proto)
	err := a.Analyze()

	require.NoError(t, err)
	assert.Len(t, a.GetSymbolTable(), 2)

	offset, err := a.GetFieldOffset("field1")
	require.NoError(t, err)
	assert.Equal(t, 0, offset)

	offset, err = a.GetFieldOffset("field2")
	require.NoError(t, err)
	assert.Equal(t, 2, offset)
}

func TestAnalyzer_BitOverlap(t *testing.T) {
	proto := &ast.Protocol{
		Name:     "Test",
		PacketID: 0x1234,
		Fields: []ast.Field{
			&ast.BitStructField{
				Name: "flags",
				Fields: []*ast.BitFieldSpec{
					{Name: "a", Bit: 7, IsRange: false},
					{Name: "b", Bit: 7, IsRange: false},
				},
			},
		},
	}

	a := NewAnalyzer(proto)
	err := a.Analyze()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "уже используется")
}

func TestAnalyzer_InvalidLengthFrom(t *testing.T) {
	proto := &ast.Protocol{
		Name:     "Test",
		PacketID: 0x1234,
		Fields: []ast.Field{
			&ast.BytesField{Name: "data", LengthFrom: "nonexistent"},
		},
	}

	a := NewAnalyzer(proto)
	err := a.Analyze()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "несуществующее поле")
}
