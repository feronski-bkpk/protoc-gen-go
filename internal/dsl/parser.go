package dsl

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
)

var (
	dslLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Keyword", Pattern: `protocol|struct|bitstruct|id|if|length_from|length|repeated`},
		{Name: "Slice", Pattern: `\[\]`},
		{Name: "Type", Pattern: `uint8|uint16|uint32|uint64|int8|int16|int32|int64|float32|float64|bytes`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "HexNumber", Pattern: `0x[0-9a-fA-F]+`},
		{Name: "Number", Pattern: `\d+`},
		{Name: "Operator", Pattern: `==|!=|<=|>=|<|>`},
		{Name: "Punct", Pattern: `[(){}\[\]:;,]`},
		{Name: "Comment", Pattern: `//[^\n]*`},
		{Name: "Whitespace", Pattern: `\s+`},
	})
)

type Parser struct {
	parser *participle.Parser[Protocol]
}

func NewParser() *Parser {
	parser := participle.MustBuild[Protocol](
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
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filename, err)
	}
	dslProto, err := p.parser.ParseBytes(filename, data)
	if err != nil {
		return nil, fmt.Errorf("parse file %s: %w", filename, err)
	}
	return dslProto.ToAST()
}

type Protocol struct {
	Name string   `parser:"'protocol' @Ident '{'"`
	ID   string   `parser:"'id' ':' @HexNumber"`
	Body []*Field `parser:"@@* '}'"`
}

func (p *Protocol) ToAST() (*ast.Protocol, error) {
	packetID, err := strconv.ParseUint(strings.TrimPrefix(p.ID, "0x"), 16, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid packet ID: %s", p.ID)
	}

	proto := &ast.Protocol{
		Name:     p.Name,
		PacketID: uint16(packetID),
		Types:    make(map[string]*ast.StructType),
		Fields:   make([]ast.Field, 0),
	}

	for _, f := range p.Body {
		field, err := f.ToAST()
		if err != nil {
			return nil, err
		}
		if field != nil {
			proto.Fields = append(proto.Fields, field)
		}
	}

	return proto, nil
}

type Field struct {
	Name      string     `parser:"@Ident ':'"`
	Type      *FieldType `parser:"@@"`
	LengthOpt *LengthOpt `parser:"@@?"`
	Condition *Condition `parser:"( 'if' @@ )?"`
}

type LengthOpt struct {
	LengthFrom *string `parser:"( 'length_from' ':' @Ident )"`
	Length     *string `parser:"| ( 'length' ':' @Ident )"`
}

func (f *Field) ToAST() (ast.Field, error) {
	if f.Type == nil {
		return nil, nil
	}

	condition, err := f.Condition.ToAST()
	if err != nil {
		return nil, err
	}

	lengthFrom := ""
	if f.LengthOpt != nil {
		if f.LengthOpt.LengthFrom != nil {
			lengthFrom = *f.LengthOpt.LengthFrom
		} else if f.LengthOpt.Length != nil {
			lengthFrom = *f.LengthOpt.Length
		}
	}

	if f.Type.Scalar != nil {
		return &ast.ScalarField{
			Name:      f.Name,
			Type:      ast.ScalarType(*f.Type.Scalar),
			Condition: condition,
		}, nil
	}

	if f.Type.Struct != nil {
		fields := make([]ast.Field, 0)
		for _, sf := range f.Type.Struct.Body {
			field, err := sf.ToAST()
			if err != nil {
				return nil, err
			}
			if field != nil {
				fields = append(fields, field)
			}
		}
		return &ast.StructField{
			Name:      f.Name,
			Struct:    &ast.StructType{Fields: fields},
			Condition: condition,
		}, nil
	}

	if f.Type.BitStruct != nil {
		bitFields := make([]*ast.BitFieldSpec, 0)
		for _, bf := range f.Type.BitStruct.Fields {
			spec := bf.ToAST()
			if spec != nil {
				bitFields = append(bitFields, spec)
			}
		}
		return &ast.BitStructField{
			Name:      f.Name,
			Fields:    bitFields,
			Condition: condition,
		}, nil
	}

	if f.Type.Bytes {
		return &ast.BytesField{
			Name:       f.Name,
			LengthFrom: lengthFrom,
			Condition:  condition,
		}, nil
	}

	if f.Type.SliceScalar != nil {
		elemField := &ast.ScalarField{
			Type: ast.ScalarType(*f.Type.SliceScalar),
		}
		return &ast.ArrayField{
			Name:        f.Name,
			ElementType: elemField,
			FixedLength: 0,
			LengthFrom:  lengthFrom,
			Condition:   condition,
		}, nil
	}

	if f.Type.SliceStruct != nil {
		fields := make([]ast.Field, 0)
		for _, sf := range f.Type.SliceStruct.Body {
			field, err := sf.ToAST()
			if err != nil {
				return nil, err
			}
			if field != nil {
				fields = append(fields, field)
			}
		}
		elemField := &ast.StructField{
			Struct: &ast.StructType{Fields: fields},
		}
		return &ast.ArrayField{
			Name:        f.Name,
			ElementType: elemField,
			FixedLength: 0,
			LengthFrom:  lengthFrom,
			Condition:   condition,
		}, nil
	}

	if f.Type.Repeated != nil && f.Type.RepeatedLength != nil {
		if f.Type.Repeated.Scalar != "" {
			elemField := &ast.ScalarField{
				Type: ast.ScalarType(f.Type.Repeated.Scalar),
			}
			return &ast.ArrayField{
				Name:        f.Name,
				ElementType: elemField,
				FixedLength: *f.Type.RepeatedLength,
				Condition:   condition,
			}, nil
		}
		if f.Type.Repeated.Struct != nil {
			fields := make([]ast.Field, 0)
			for _, sf := range f.Type.Repeated.Struct.Body {
				field, err := sf.ToAST()
				if err != nil {
					return nil, err
				}
				if field != nil {
					fields = append(fields, field)
				}
			}
			elemField := &ast.StructField{
				Struct: &ast.StructType{Fields: fields},
			}
			return &ast.ArrayField{
				Name:        f.Name,
				ElementType: elemField,
				FixedLength: *f.Type.RepeatedLength,
				Condition:   condition,
			}, nil
		}
	}

	return nil, nil
}

type FieldType struct {
	Scalar         *string       `parser:"  @('uint8'|'uint16'|'uint32'|'uint64'|'int8'|'int16'|'int32'|'int64'|'float32'|'float64')"`
	Struct         *Struct       `parser:"| 'struct' @@"`
	BitStruct      *BitStructDef `parser:"| 'bitstruct' @@"`
	Bytes          bool          `parser:"| @'bytes'"`
	SliceScalar    *string       `parser:"| '[]' @('uint8'|'uint16'|'uint32'|'uint64'|'int8'|'int16'|'int32'|'int64'|'float32'|'float64')"`
	SliceStruct    *SliceStruct  `parser:"| '[]' 'struct' @@"`
	Repeated       *RepeatedDef  `parser:"| 'repeated' @@ 'length' ':' @Number"`
	RepeatedLength *int          `parser:""`
}

type RepeatedDef struct {
	Scalar string  `parser:"  @('uint8'|'uint16'|'uint32'|'uint64'|'int8'|'int16'|'int32'|'int64'|'float32'|'float64')"`
	Struct *Struct `parser:"| 'struct' @@"`
}

func (r *RepeatedDef) Capture(values []string) error { return nil }

type SliceStruct struct {
	Body []*Field `parser:"'{' @@* '}'"`
}

type Struct struct {
	Body []*Field `parser:"'{' @@* '}'"`
}

type BitStructDef struct {
	Fields []*BitFieldDef `parser:"'{' @@* '}'"`
}

type BitFieldDef struct {
	Name string      `parser:"@Ident ':'"`
	Spec *BitSpecDef `parser:"@@"`
}

type BitSpecDef struct {
	BitsHi *int `parser:"  'bits' '[' @Number"`
	BitsLo *int `parser:"  ':' @Number ']'"`
	Bit    *int `parser:"| 'bit' '(' @Number ')'"`
}

func (b *BitFieldDef) ToAST() *ast.BitFieldSpec {
	if b.Spec == nil {
		return nil
	}
	if b.Spec.BitsHi != nil && b.Spec.BitsLo != nil {
		return &ast.BitFieldSpec{
			Name:    b.Name,
			HighBit: *b.Spec.BitsHi,
			LowBit:  *b.Spec.BitsLo,
			IsRange: true,
		}
	}
	if b.Spec.Bit != nil {
		return &ast.BitFieldSpec{
			Name:    b.Name,
			Bit:     *b.Spec.Bit,
			IsRange: false,
		}
	}
	return nil
}

type Condition struct {
	Field string `parser:"@Ident"`
	Op    string `parser:"@('=='|'!='|'>'|'<'|'>='|'<=')"`
	Value string `parser:"@(Number|HexNumber)"`
}

func (c *Condition) ToAST() (*ast.Condition, error) {
	if c == nil {
		return nil, nil
	}
	var val uint64
	var err error
	if strings.HasPrefix(c.Value, "0x") {
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
