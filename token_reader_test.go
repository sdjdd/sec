package sec

import (
	"io"
	"strings"
	"testing"
)

func TestUnexpectedRune(t *testing.T) {
	var r tokenReader
	r.load("\n0b6")
	t.Log(r.read())
}

func TestReadToken(t *testing.T) {
	var r tokenReader
	if tk, err := r.read(); err != io.EOF || tk.typ != initial {
		t.Fatal("expect initial tokan and EOF error")
	}

	testTokens := []struct {
		txt string
		typ int
	}{
		{"id", identifier},
		{"123", integer}, {"0", integer},
		{"3.14", float}, {"0.000001", float},
		{"0b1101", binLiteral}, {"0B1001", binLiteral},
		{"0o7654", octLiteral}, {"0O3456", octLiteral}, {"04567", octLiteral},
		{"0xabcd", hexLiteral}, {"0XFFFF", hexLiteral},
		{"+", plus}, {"-", minus}, {"*", star}, {"/", slash},
		{"**", doubleStar}, {"//", doubleSlash},
	}

	var buf strings.Builder
	for i, tt := range testTokens {
		buf.WriteString(tt.txt)
		if i%10 == 0 {
			buf.WriteString("\r\n")
		} else {
			buf.WriteRune(' ')
		}
	}
	r.load(buf.String())

	for _, tt := range testTokens {
		tk, err := r.read()
		if err != nil {
			t.Fatal("expect no error")
		} else if tk.typ != tt.typ {
			t.Fatalf("expect %s %s", token{typ: tt.typ}, tt.txt)
		}
	}
}
