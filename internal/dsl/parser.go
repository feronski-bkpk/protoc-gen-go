// internal/dsl/parser.go
package dsl

import (
	"fmt"
	"strconv"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/yourusername/protoc-gen-go/internal/ast"
)

var (
	dslLexer = lexer.MustSimple([]lexer.SimpleRule{
		{"Keyword", `protocol|struct|id|bit|bits|if|length_from|length|max_length`},
		{"Ident", `[a-zA-Z_][a-zA-Z0-9_]*`},
		{"HexNumber", `0x[0-9a-fA-F]+`},
		{"Number", `\d+`},
		{"Punct", `[(){}\[\]:.,;=]|==|!=|<=|>=|<|>`},
		{"Comment", `//[^\n]*`},
		{"Whitespace", `\s+`},
	})
)

type Parser struct {
	parser *participle.Parser[dslProtocol]
}

func NewParser() *Parser {
	parser := participle.MustBuild[dslProtocol](
		participle.Lexer(dslLexer),
		participle.Elide("Comment", "Whitespace"),
	)
	return &Parser{parser: parser}
}

func (p *Parser) ParseString(input string) (*ast.Protocol, error) {
	dslProto, err := p.parser.ParseString("", input)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return dslProto.ToAST()
}

func (p *Parser) ParseFile(filename string) (*ast.Protocol, error) {
	dslProto, err := p.parser.ParseFile(filename)
	if err != nil {
		return nil, fmt.Errorf("parse file %s: %w", filename, err)
	}

	return dslProto.ToAST()
}

// DSL структуры для парсинга
type dslProtocol struct {
	Name   string      `parser:"'protocol' @Ident '{'"`
	ID     string      `parser:"'id' ':' @HexNumber"`
	Fields []*dslField `parser:"@@* '}'"`
}

func (p *dslProtocol) ToAST() (*ast.Protocol, error) {
	packetID, err := strconv.ParseUint(p.ID[2:], 16, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid packet ID: %s", p.ID)
	}

	proto := &ast.Protocol{
		Name:     p.Name,
		PacketID: uint16(packetID),
		Types:    make(map[string]*ast.StructType),
		Fields:   make([]ast.Field, 0),
	}

	for _, f := range p.Fields {
		field, err := f.ToAST()
		if err != nil {
			return nil, err
		}
		proto.Fields = append(proto.Fields, field)
	}

	return proto, nil
}

type dslField struct {
	Name      string        `parser:"@Ident ':'"`
	Type      *dslFieldType `parser:"@@"`
	Condition *dslCondition `parser:"( 'if' @@ )?"`
}

func (f *dslField) ToAST() (ast.Field, error) {
	astType, isStruct, structName := f.Type.ToAST()

	condition, err := f.Condition.ToAST()
	if err != nil {
		return nil, err
	}

	switch t := astType.(type) {
	case ast.ScalarType:
		return &ast.ScalarField{
			Name:      f.Name,
			Type:      t,
			Condition: condition,
		}, nil

	case *ast.ArrayField:
		t.Name = f.Name
		t.Condition = condition
		return t, nil

	case *ast.BytesField:
		t.Name = f.Name
		t.Condition = condition
		return t, nil

	case *ast.StructType:
		if isStruct {
			// Это вложенное определение структуры
			structName = f.Name + "_struct"
		}
		return &ast.StructField{
			Name:      f.Name,
			TypeRef:   structName,
			Struct:    t,
			Condition: condition,
		}, nil
	}

	return nil, fmt.Errorf("unknown field type for %s", f.Name)
}

type dslFieldType struct {
	Scalar *string       `parser:"  @('uint8'|'uint16'|'uint32'|'uint64'|'int8'|'int16'|'int32'|'int64'|'float32'|'float64')"`
	Bit    *dslBitField  `parser:"| @('bit'|'bits')"`
	Struct *dslStructDef `parser:"| 'struct' '{' @@ '}'"`
	Array  *dslArraySpec `parser:"| '[]' @@"`
	Bytes  *dslBytesSpec `parser:"| 'bytes' '(' @@? ')'"`
}

func (t *dslFieldType) ToAST() (interface{}, bool, string) {
	if t.Scalar != nil {
		return ast.ScalarType(*t.Scalar), false, ""
	}

	if t.Bit != nil {
		// Битовые поля будут обработаны позже
		return nil, false, ""
	}

	if t.Struct != nil {
		fields := make([]ast.Field, 0)
		for _, f := range t.Struct.Fields {
			field, _ := f.ToAST()
			fields = append(fields, field)
		}

		structType := &ast.StructType{
			Name:        "",
			Fields:      fields,
			IsBitPacked: t.Struct.IsBitPacked,
		}
		return structType, true, ""
	}

	if t.Array != nil {
		elemType, isStruct, structName := t.Array.Element.ToAST()

		var elemField ast.Field
		switch et := elemType.(type) {
		case ast.ScalarType:
			elemField = &ast.ScalarField{Type: et}
		case *ast.StructType:
			if isStruct {
				structName = "elem_struct"
			}
			elemField = &ast.StructField{TypeRef: structName, Struct: et}
		}

		array := &ast.ArrayField{
			Element: elemField,
		}

		if t.Array.LengthFrom != nil {
			array.LengthFrom = *t.Array.LengthFrom
		}
		if t.Array.FixedLength != nil {
			array.FixedLength = *t.Array.FixedLength
		}

		return array, false, ""
	}

	if t.Bytes != nil {
		bytes := &ast.BytesField{}
		if t.Bytes.LengthFrom != nil {
			bytes.LengthFrom = *t.Bytes.LengthFrom
		}
		if t.Bytes.MaxLength != nil {
			bytes.MaxLength = *t.Bytes.MaxLength
		}
		return bytes, false, ""
	}

	return nil, false, ""
}

type dslBitField struct {
	Type string `parser:"@('bit'|'bits')"`
	Bits int    `parser:"'(' @Number ')'"`
}

type dslStructDef struct {
	Fields      []*dslField `parser:"'{' @@* '}'"`
	IsBitPacked bool
}

type dslArraySpec struct {
	Element     *dslFieldType `parser:"@@"`
	LengthFrom  *string       `parser:"( 'length_from' ':' @Ident )?"`
	FixedLength *int          `parser:"| 'length' ':' @Number"`
}

type dslBytesSpec struct {
	LengthFrom *string `parser:"'length_from' ':' @Ident"`
	MaxLength  *int    `parser:"| 'max_length' ':' @Number"`
}

type dslCondition struct {
	Field string `parser:"@Ident"`
	Op    string `parser:"@('=='|'!='|'>'|'<'|'>='|'<=')"`
	Value string `parser:"@(Number|HexNumber)"`
}

func (c *dslCondition) ToAST() (*ast.Condition, error) {
	if c == nil {
		return nil, nil
	}

	var val uint64
	var err error

	if len(c.Value) > 2 && c.Value[:2] == "0x" {
		val, err = strconv.ParseUint(c.Value[2:], 16, 64)
	} else {
		val, err = strconv.ParseUint(c.Value, 10, 64)
	}

	if err != nil {
		return nil, fmt.Errorf("invalid condition value: %s", c.Value)
	}

	return &ast.Condition{
		Field:    c.Field,
		Operator: c.Op,
		Value:    val,
	}, nil
}
