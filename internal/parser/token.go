package parser

import "fmt"

// TokenType определяет тип токена
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenNumber
	TokenString
	TokenHexNumber

	// Ключевые слова
	TokenProtocol
	TokenStruct
	TokenBitStruct
	TokenID
	TokenIf
	TokenLengthFrom
	TokenLength

	// Знаки пунктуации
	TokenColon     // :
	TokenSemicolon // ;
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenLParen    // (
	TokenRParen    // )
	TokenComma     // ,

	// Операторы
	TokenEq    // ==
	TokenNotEq // !=
	TokenLT    // <
	TokenGT    // >
	TokenLTE   // <=
	TokenGTE   // >=
)

// Token представляет один токен
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
	TokenID:         "id",
	TokenIf:         "if",
	TokenLengthFrom: "length_from",
	TokenLength:     "length",
	TokenColon:      ":",
	TokenSemicolon:  ";",
	TokenLBrace:     "{",
	TokenRBrace:     "}",
	TokenLBracket:   "[",
	TokenRBracket:   "]",
	TokenLParen:     "(",
	TokenRParen:     ")",
	TokenComma:      ",",
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
