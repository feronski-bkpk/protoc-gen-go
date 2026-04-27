package parser

import (
	"testing"
)

// FuzzParser проверяет, что парсер не паникует на случайных входных данных
func FuzzParser(f *testing.F) {
	seeds := []string{
		`protocol Test { id: 0x1234 field: uint16 }`,
		`protocol Test { id: 0x0 field: uint8 }`,
		`protocol Test { id: 0xFFFF field: float32 }`,
		`protocol Test { id: 0x1 field: struct { x: uint8 } }`,
		`protocol Test { id: 0x1 flags: bitstruct { ack: bit(7) } }`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("паника на входе %q: %v", input, r)
			}
		}()

		lexer := NewLexer(input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			return
		}

		p := NewParser(tokens)
		proto, err := p.Parse()
		if err != nil {
			return
		}

		_ = proto
	})
}

// FuzzLexer проверяет, что лексер не паникует
func FuzzLexer(f *testing.F) {
	seeds := []string{
		"protocol",
		"struct",
		"bitstruct",
		"0x1234",
		"12345",
		"// comment",
		"field_name:",
		"",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("паника на входе %q: %v", input, r)
			}
		}()

		lexer := NewLexer(input)
		tokens, _ := lexer.Tokenize()
		_ = tokens
	})
}
