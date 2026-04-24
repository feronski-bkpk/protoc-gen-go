package parser

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer разбивает исходный текст на токены
type Lexer struct {
	input  []rune
	pos    int
	line   int
	col    int
	tokens []Token
}

// NewLexer создаёт новый лексер
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		line:  1,
		col:   1,
	}
}

// Tokenize разбирает весь входной поток на токены
func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.input) {
		ch := l.current()

		if unicode.IsSpace(ch) {
			l.advance()
			continue
		}

		if ch == '/' && l.peek() == '/' {
			l.skipComment()
			continue
		}

		if ch == '0' && l.peek() == 'x' {
			l.tokenizeHex()
			continue
		}
		if unicode.IsDigit(ch) {
			l.tokenizeNumber()
			continue
		}

		if unicode.IsLetter(ch) || ch == '_' {
			l.tokenizeIdent()
			continue
		}

		if ch == '"' {
			l.tokenizeString()
			continue
		}

		switch ch {
		case ':':
			l.emit(TokenColon, ":")
		case ';':
			l.emit(TokenSemicolon, ";")
		case '{':
			l.emit(TokenLBrace, "{")
		case '}':
			l.emit(TokenRBrace, "}")
		case '[':
			l.emit(TokenLBracket, "[")
		case ']':
			l.emit(TokenRBracket, "]")
		case '(':
			l.emit(TokenLParen, "(")
		case ')':
			l.emit(TokenRParen, ")")
		case ',':
			l.emit(TokenComma, ",")
		case '=':
			if l.peek() == '=' {
				l.advance()
				l.emit(TokenEq, "==")
			} else {
				return nil, fmt.Errorf("неожиданный символ '=' на %d:%d", l.line, l.col)
			}
		case '!':
			if l.peek() == '=' {
				l.advance()
				l.emit(TokenNotEq, "!=")
			} else {
				return nil, fmt.Errorf("неожиданный символ '!' на %d:%d", l.line, l.col)
			}
		case '<':
			if l.peek() == '=' {
				l.advance()
				l.emit(TokenLTE, "<=")
			} else {
				l.emit(TokenLT, "<")
			}
		case '>':
			if l.peek() == '=' {
				l.advance()
				l.emit(TokenGTE, ">=")
			} else {
				l.emit(TokenGT, ">")
			}
		default:
			return nil, fmt.Errorf("неожиданный символ '%c' на %d:%d", ch, l.line, l.col)
		}
	}

	return l.tokens, nil
}

func (l *Lexer) current() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) advance() {
	if l.current() == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	l.pos++
}

func (l *Lexer) emit(tokenType TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:  tokenType,
		Value: value,
		Line:  l.line,
		Col:   l.col,
	})
	l.advance()
}

func (l *Lexer) skipComment() {
	for l.pos < len(l.input) && l.current() != '\n' {
		l.advance()
	}
}

func (l *Lexer) tokenizeHex() {
	start := l.pos
	l.advance()
	l.advance()
	for l.pos < len(l.input) && (unicode.IsDigit(l.current()) ||
		(l.current() >= 'a' && l.current() <= 'f') ||
		(l.current() >= 'A' && l.current() <= 'F')) {
		l.advance()
	}
	value := string(l.input[start:l.pos])
	l.tokens = append(l.tokens, Token{Type: TokenHexNumber, Value: value, Line: l.line, Col: l.col})
}

func (l *Lexer) tokenizeNumber() {
	start := l.pos
	for l.pos < len(l.input) && unicode.IsDigit(l.current()) {
		l.advance()
	}
	value := string(l.input[start:l.pos])
	l.tokens = append(l.tokens, Token{Type: TokenNumber, Value: value, Line: l.line, Col: l.col})
}

func (l *Lexer) tokenizeIdent() {
	start := l.pos
	for l.pos < len(l.input) && (unicode.IsLetter(l.current()) || unicode.IsDigit(l.current()) || l.current() == '_') {
		l.advance()
	}
	value := string(l.input[start:l.pos])

	tokenType := TokenIdent
	switch strings.ToLower(value) {
	case "protocol":
		tokenType = TokenProtocol
	case "struct":
		tokenType = TokenStruct
	case "bitstruct":
		tokenType = TokenBitStruct
	case "id":
		tokenType = TokenID
	case "if":
		tokenType = TokenIf
	case "length_from":
		tokenType = TokenLengthFrom
	case "length":
		tokenType = TokenLength
	}

	l.tokens = append(l.tokens, Token{Type: tokenType, Value: value, Line: l.line, Col: l.col})
}

func (l *Lexer) tokenizeString() {
	l.advance()
	start := l.pos
	for l.pos < len(l.input) && l.current() != '"' {
		l.advance()
	}
	value := string(l.input[start:l.pos])
	l.advance()
	l.tokens = append(l.tokens, Token{Type: TokenString, Value: value, Line: l.line, Col: l.col})
}
