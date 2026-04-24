package parser

import (
	"fmt"
	"os"
	"strconv"

	"github.com/feronski-bkpk/protoc-gen-go/internal/ast"
)

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) Parse() (*ast.Protocol, error) {
	return p.parseProtocol()
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

func (p *Parser) parseProtocol() (*ast.Protocol, error) {
	if !p.match(TokenProtocol) {
		return nil, p.error("ожидалось 'protocol'")
	}

	nameTok := p.expect(TokenIdent)

	if !p.match(TokenLBrace) {
		return nil, p.error("ожидалась '{'")
	}

	if !p.match(TokenID) {
		return nil, p.error("ожидалось 'id'")
	}
	if !p.match(TokenColon) {
		return nil, p.error("ожидалось ':'")
	}

	idTok := p.expect(TokenHexNumber)
	packetID, err := strconv.ParseUint(idTok.Value[2:], 16, 16)
	if err != nil {
		return nil, fmt.Errorf("неверный ID пакета: %s", idTok.Value)
	}

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

	case p.isType():
		typeName := p.current().Value
		p.advance()

		field := &ast.ScalarField{
			Name: nameTok.Value,
			Type: ast.ScalarType(typeName),
		}

		if p.match(TokenIf) {
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
		cond, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		field.Condition = cond
	}

	return field, nil
}

func (p *Parser) parseBitStructField(name string) (ast.Field, error) {
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
	return &ast.ArrayField{
		Name: name,
		ElementType: &ast.ScalarField{
			Type: ast.ScalarType(typeName),
		},
		FixedLength: size,
	}, nil
}

func (p *Parser) parseSliceField(name string) (ast.Field, error) {
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
	field := &ast.ArrayField{
		Name: name,
		ElementType: &ast.ScalarField{
			Type: ast.ScalarType(typeName),
		},
	}

	if p.match(TokenLength) || p.match(TokenLengthFrom) {
		if !p.match(TokenColon) {
			return nil, p.error("ожидалось ':'")
		}
		lenField := p.expect(TokenIdent)
		field.LengthFrom = lenField.Value
	}

	if p.match(TokenIf) {
		cond, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		field.Condition = cond
	}

	return field, nil
}

func (p *Parser) parseBytesField(name string) (ast.Field, error) {
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
		cond, err := p.parseCondition()
		if err != nil {
			return nil, err
		}
		field.Condition = cond
	}

	return field, nil
}

func (p *Parser) parseCondition() (*ast.Condition, error) {
	fieldTok := p.expect(TokenIdent)

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
	tok := p.current()
	if tok.Type == TokenNumber {
		val, _ = strconv.ParseUint(tok.Value, 10, 64)
	} else if tok.Type == TokenHexNumber {
		val, _ = strconv.ParseUint(tok.Value[2:], 16, 64)
	} else {
		return nil, p.error("ожидалось число")
	}
	p.advance()

	return &ast.Condition{
		Field:    fieldTok.Value,
		Operator: op,
		Value:    val,
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
	return types[t.Value]
}

func (p *Parser) isEOF() bool {
	return p.current().Type == TokenEOF
}

func (p *Parser) error(msg string) error {
	tok := p.current()
	return fmt.Errorf("%s (строка %d, колонка %d: '%s')", msg, tok.Line, tok.Col, tok.Value)
}
