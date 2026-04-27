package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
)

type Parser struct {
	tokens  []Token
	pos     int
	aliases map[string]string
	context string
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (*ast.Protocol, error) {
	proto, err := p.parseProtocol()
	if err != nil {
		return nil, err
	}

	for _, field := range proto.Fields {
		if enumField, ok := field.(*ast.EnumField); ok {
			if proto.Enums == nil {
				proto.Enums = make(map[string]*ast.EnumType)
			}
			proto.Enums[enumField.EnumName] = enumField.Enum
		}
	}

	return proto, nil
}

func ParseFile(filename string) (*ast.Protocol, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла %s: %w", filename, err)
	}

	lexer := NewLexer(string(data))
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("ошибка лексера: %w", err)
	}

	p := NewParser(tokens)
	proto, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("ошибка парсера: %w", err)
	}

	return proto, nil
}

func ParseString(input string) (*ast.Protocol, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, fmt.Errorf("ошибка лексера: %w", err)
	}

	p := NewParser(tokens)
	return p.Parse()
}

func (p *Parser) setContext(ctx string) {
	p.context = ctx
}

func (p *Parser) parseProtocol() (*ast.Protocol, error) {
	p.setContext("протокола")

	if !p.match(TokenProtocol) {
		return nil, p.error("ожидалось 'protocol'")
	}

	nameTok := p.expect(TokenIdent)

	if !p.match(TokenLBrace) {
		return nil, p.error("ожидалась '{'")
	}

	p.setContext("идентификатора протокола")

	if p.current().Value != "id" {
		return nil, p.error("ожидалось 'id'")
	}
	p.advance()
	if !p.match(TokenColon) {
		return nil, p.error("ожидалось ':'")
	}

	idTok := p.expect(TokenHexNumber)
	packetID, err := strconv.ParseUint(idTok.Value[2:], 16, 16)
	if err != nil {
		return nil, fmt.Errorf("неверный ID пакета: %s", idTok.Value)
	}

	endian := "big"
	if p.current().Value == "endian" {
		p.setContext("настройки порядка байт")
		p.advance()
		if !p.match(TokenColon) {
			return nil, p.error("ожидалось ':'")
		}
		endianTok := p.expect(TokenIdent)
		endian = strings.ToLower(endianTok.Value)
	}

	p.aliases = make(map[string]string)
	for p.current().Value == "alias" {
		p.setContext("алиаса")
		p.advance()
		aliasName := p.expect(TokenIdent)
		if !p.match(TokenColon) {
			return nil, p.error("ожидалось ':'")
		}
		baseType := p.expect(TokenIdent)
		p.aliases[aliasName.Value] = baseType.Value
	}

	p.setContext("полей протокола")
	fields, err := p.parseFields()
	if err != nil {
		return nil, err
	}

	if !p.match(TokenRBrace) {
		return nil, p.error("ожидалась '}'")
	}

	return &ast.Protocol{
		Name:     nameTok.Value,
		PacketID: uint16(packetID),
		Fields:   fields,
		Types:    make(map[string]*ast.StructType),
		Aliases:  p.aliases,
		Endian:   endian,
	}, nil
}

func (p *Parser) parseFields() ([]ast.Field, error) {
	var fields []ast.Field

	for !p.isEOF() && p.current().Type != TokenRBrace {
		field, err := p.parseField()
		if err != nil {
			return nil, err
		}
		if field != nil {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func (p *Parser) parseField() (ast.Field, error) {
	nameTok := p.expect(TokenIdent)
	p.setContext("поля '" + nameTok.Value + "'")

	if !p.match(TokenColon) {
		return nil, p.error("ожидалось ':'")
	}

	if p.current().Value == "bytes" {
		p.advance()
		return p.parseBytesField(nameTok.Value)
	}

	switch {
	case p.match(TokenStruct):
		return p.parseStructField(nameTok.Value)

	case p.match(TokenBitStruct):
		return p.parseBitStructField(nameTok.Value)

	case p.match(TokenLBracket):
		return p.parseArrayField(nameTok.Value)

	case p.match(TokenEnum):
		return p.parseEnumField(nameTok.Value)

	case p.isType():
		typeName := p.current().Value
		originalType := typeName
		p.advance()

		if p.aliases != nil {
			if baseType, ok := p.aliases[typeName]; ok {
				typeName = baseType
			}
		}

		if typeName == "bytes" {
			return p.parseBytesField(nameTok.Value)
		}

		field := &ast.ScalarField{
			Name:         nameTok.Value,
			Type:         ast.ScalarType(typeName),
			OriginalType: originalType,
		}

		if p.match(TokenIf) {
			p.setContext("условия поля '" + nameTok.Value + "'")
			cond, err := p.parseCondition()
			if err != nil {
				return nil, err
			}
			field.Condition = cond
		}

		return field, nil
	}

	return nil, p.error("ожидался тип поля")
}

func (p *Parser) parseStructField(name string) (ast.Field, error) {
	p.setContext("структуры '" + name + "'")

	if !p.match(TokenLBrace) {
		return nil, p.error("ожидалась '{'")
	}

	fields, err := p.parseFields()
	if err != nil {
		return nil, err
	}

	if !p.match(TokenRBrace) {
		return nil, p.error("ожидалась '}'")
	}

	field := &ast.StructField{
		Name:   name,
		Struct: &ast.StructType{Fields: fields},
	}

	if p.match(TokenIf) {
		p.setContext("условия структуры '" + name + "'")
		cond, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		field.Condition = cond
	}

	return field, nil
}

func (p *Parser) parseBitStructField(name string) (ast.Field, error) {
	p.setContext("битовой структуры '" + name + "'")

	if !p.match(TokenLBrace) {
		return nil, p.error("ожидалась '{'")
	}

	var bitFields []*ast.BitFieldSpec

	for !p.isEOF() && p.current().Type != TokenRBrace {
		bitName := p.expect(TokenIdent)
		if !p.match(TokenColon) {
			return nil, p.error("ожидалось ':'")
		}

		if p.matchString("bit") {
			if !p.match(TokenLParen) {
				return nil, p.error("ожидалась '('")
			}
			bitNumTok := p.expect(TokenNumber)
			if !p.match(TokenRParen) {
				return nil, p.error("ожидалась ')'")
			}
			bitNum, _ := strconv.Atoi(bitNumTok.Value)
			bitFields = append(bitFields, &ast.BitFieldSpec{
				Name:    bitName.Value,
				Bit:     bitNum,
				IsRange: false,
			})
		} else if p.matchString("bits") {
			if !p.match(TokenLBracket) {
				return nil, p.error("ожидался '['")
			}
			hiTok := p.expect(TokenNumber)
			if !p.match(TokenColon) {
				return nil, p.error("ожидалось ':'")
			}
			loTok := p.expect(TokenNumber)
			if !p.match(TokenRBracket) {
				return nil, p.error("ожидался ']'")
			}
			hi, _ := strconv.Atoi(hiTok.Value)
			lo, _ := strconv.Atoi(loTok.Value)
			bitFields = append(bitFields, &ast.BitFieldSpec{
				Name:    bitName.Value,
				HighBit: hi,
				LowBit:  lo,
				IsRange: true,
			})
		} else {
			return nil, p.error("ожидалось 'bit' или 'bits'")
		}
	}

	if !p.match(TokenRBrace) {
		return nil, p.error("ожидалась '}'")
	}

	return &ast.BitStructField{
		Name:   name,
		Fields: bitFields,
	}, nil
}

func (p *Parser) parseArrayField(name string) (ast.Field, error) {
	p.setContext("массива '" + name + "'")

	if p.match(TokenRBracket) {
		return p.parseSliceField(name)
	}

	sizeTok := p.expect(TokenNumber)
	if !p.match(TokenRBracket) {
		return nil, p.error("ожидался ']'")
	}

	size, _ := strconv.Atoi(sizeTok.Value)

	if p.match(TokenStruct) {
		if !p.match(TokenLBrace) {
			return nil, p.error("ожидалась '{'")
		}
		fields, err := p.parseFields()
		if err != nil {
			return nil, err
		}
		if !p.match(TokenRBrace) {
			return nil, p.error("ожидалась '}'")
		}

		return &ast.ArrayField{
			Name: name,
			ElementType: &ast.StructField{
				Struct: &ast.StructType{Fields: fields},
			},
			FixedLength: size,
		}, nil
	}

	typeName := p.expectIdent()
	originalType := typeName

	if p.aliases != nil {
		if baseType, ok := p.aliases[typeName]; ok {
			typeName = baseType
		}
	}

	elemField := &ast.ScalarField{
		Type:         ast.ScalarType(typeName),
		OriginalType: originalType,
	}

	return &ast.ArrayField{
		Name:        name,
		ElementType: elemField,
		FixedLength: size,
	}, nil
}

func (p *Parser) parseSliceField(name string) (ast.Field, error) {
	p.setContext("слайса '" + name + "'")

	if p.match(TokenStruct) {
		if !p.match(TokenLBrace) {
			return nil, p.error("ожидалась '{'")
		}
		fields, err := p.parseFields()
		if err != nil {
			return nil, err
		}
		if !p.match(TokenRBrace) {
			return nil, p.error("ожидалась '}'")
		}

		field := &ast.ArrayField{
			Name: name,
			ElementType: &ast.StructField{
				Struct: &ast.StructType{Fields: fields},
			},
		}

		if p.match(TokenLength) || p.match(TokenLengthFrom) {
			if !p.match(TokenColon) {
				return nil, p.error("ожидалось ':'")
			}
			lenField := p.expect(TokenIdent)
			field.LengthFrom = lenField.Value
		}

		return field, nil
	}

	typeName := p.expectIdent()
	originalType := typeName

	if p.aliases != nil {
		if baseType, ok := p.aliases[typeName]; ok {
			typeName = baseType
		}
	}

	elemField := &ast.ScalarField{
		Type:         ast.ScalarType(typeName),
		OriginalType: originalType,
	}

	field := &ast.ArrayField{
		Name:        name,
		ElementType: elemField,
	}

	if p.match(TokenLength) || p.match(TokenLengthFrom) {
		if !p.match(TokenColon) {
			return nil, p.error("ожидалось ':'")
		}
		lenField := p.expect(TokenIdent)
		field.LengthFrom = lenField.Value
	}

	if p.match(TokenIf) {
		p.setContext("условия слайса '" + name + "'")
		cond, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		field.Condition = cond
	}

	return field, nil
}

func (p *Parser) parseBytesField(name string) (ast.Field, error) {
	p.setContext("bytes поля '" + name + "'")

	field := &ast.BytesField{
		Name: name,
	}

	if p.match(TokenLengthFrom) || p.match(TokenLength) {
		if !p.match(TokenColon) {
			return nil, p.error("ожидалось ':'")
		}
		lenField := p.expect(TokenIdent)
		field.LengthFrom = lenField.Value
	}

	if p.match(TokenIf) {
		p.setContext("условия bytes поля '" + name + "'")
		cond, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		field.Condition = cond
	}

	return field, nil
}

func (p *Parser) parseEnumField(name string) (ast.Field, error) {
	p.setContext("enum '" + name + "'")

	if !p.match(TokenLBrace) {
		return nil, p.error("ожидалась '{'")
	}

	enumType := &ast.EnumType{
		Name: name + "Enum",
	}

	for !p.isEOF() && p.current().Type != TokenRBrace {
		valueName := p.expectIdent()
		if !p.match(TokenEqAssign) {
			return nil, p.error("ожидался '='")
		}
		valueNum := p.expect(TokenNumber)
		val, _ := strconv.Atoi(valueNum.Value)

		enumType.Values = append(enumType.Values, &ast.EnumValue{
			Name:  valueName,
			Value: val,
		})
	}

	if !p.match(TokenRBrace) {
		return nil, p.error("ожидалась '}'")
	}

	return &ast.EnumField{
		Name:     name,
		EnumName: enumType.Name,
		Enum:     enumType,
	}, nil
}

func (p *Parser) parseCondition() (*ast.Condition, error) {
	var parts []string
	parts = append(parts, p.expectIdent())

	for p.match(TokenDot) {
		parts = append(parts, p.expectIdent())
	}

	fieldPath := strings.Join(parts, ".")

	var op string
	switch {
	case p.match(TokenEq):
		op = "=="
	case p.match(TokenNotEq):
		op = "!="
	case p.match(TokenLT):
		op = "<"
	case p.match(TokenGT):
		op = ">"
	case p.match(TokenLTE):
		op = "<="
	case p.match(TokenGTE):
		op = ">="
	default:
		return nil, p.error("ожидался оператор сравнения")
	}

	var val uint64
	var enumValue string
	tok := p.current()
	if tok.Type == TokenNumber {
		val, _ = strconv.ParseUint(tok.Value, 10, 64)
	} else if tok.Type == TokenHexNumber {
		val, _ = strconv.ParseUint(tok.Value[2:], 16, 64)
	} else if tok.Type == TokenIdent {
		enumValue = tok.Value
	} else {
		return nil, p.error("ожидалось число или идентификатор")
	}
	p.advance()

	return &ast.Condition{
		Field:     fieldPath,
		Operator:  op,
		Value:     val,
		EnumValue: enumValue,
	}, nil
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	p.pos++
}

func (p *Parser) match(t TokenType) bool {
	if p.current().Type == t {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) matchString(s string) bool {
	if p.current().Value == s {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) expect(t TokenType) Token {
	tok := p.current()
	if tok.Type != t {
		panic(p.error(fmt.Sprintf("ожидался %s, получен %s", t, tok.Type)))
	}
	p.advance()
	return tok
}

func (p *Parser) expectIdent() string {
	tok := p.expect(TokenIdent)
	return tok.Value
}

func (p *Parser) isType() bool {
	t := p.current()
	types := map[string]bool{
		"uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"int8": true, "int16": true, "int32": true, "int64": true,
		"float32": true, "float64": true,
	}
	if types[t.Value] {
		return true
	}
	if p.aliases != nil {
		if _, ok := p.aliases[t.Value]; ok {
			return true
		}
	}
	return false
}

func (p *Parser) isEOF() bool {
	return p.current().Type == TokenEOF
}

func (p *Parser) error(msg string) error {
	tok := p.current()
	ctx := ""
	if p.context != "" {
		ctx = " при парсинге " + p.context
	}
	return fmt.Errorf("%s%s (строка %d, колонка %d: '%s')",
		msg, ctx, tok.Line, tok.Col, tok.Value)
}
