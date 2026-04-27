package parser

import "fmt"

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenNumber
	TokenString
	TokenHexNumber

	TokenProtocol
	TokenStruct
	TokenBitStruct
	TokenEnum
	TokenAlias
	TokenEndian
	TokenID
	TokenIf
	TokenLengthFrom
	TokenLength

	TokenColon
	TokenSemicolon
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenLParen
	TokenRParen
	TokenComma
	TokenDot
	TokenEqAssign

	TokenEq
	TokenNotEq
	TokenLT
	TokenGT
	TokenLTE
	TokenGTE

	TokenAnd
	TokenOr
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)", t.Type, t.Value)
}

var tokenNames = map[TokenType]string{
	TokenEOF:        "EOF",
	TokenIdent:      "IDENT",
	TokenNumber:     "NUMBER",
	TokenString:     "STRING",
	TokenHexNumber:  "HEX",
	TokenProtocol:   "protocol",
	TokenStruct:     "struct",
	TokenBitStruct:  "bitstruct",
	TokenEnum:       "enum",
	TokenAlias:      "alias",
	TokenEndian:     "endian",
	TokenID:         "id",
	TokenIf:         "if",
	TokenLengthFrom: "length_from",
	TokenLength:     "length",
	TokenAnd:        "&&",
	TokenOr:         "||",
	TokenColon:      ":",
	TokenSemicolon:  ";",
	TokenLBrace:     "{",
	TokenRBrace:     "}",
	TokenLBracket:   "[",
	TokenRBracket:   "]",
	TokenLParen:     "(",
	TokenRParen:     ")",
	TokenComma:      ",",
	TokenDot:        ".",
	TokenEqAssign:   "=",
	TokenEq:         "==",
	TokenNotEq:      "!=",
	TokenLT:         "<",
	TokenGT:         ">",
	TokenLTE:        "<=",
	TokenGTE:        ">=",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Token(%d)", int(t))
}
