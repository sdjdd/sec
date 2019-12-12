package sec

import (
	"io"
	"strings"
	"testing"
)

func TestReadEmptyText(t *testing.T) {
	var r tokenReader
	tk, err := r.read()
	if err != io.EOF || tk.typ != initial {
		t.Fatal("expect initial tokan and EOF error")
	}
}

func TestUnexpectedRune(t *testing.T) {
	var r tokenReader
	r.load("\r\n")
	for {
		tk, err := r.read()
		if err != nil {
			t.Log(err)
			break
		}
		t.Log(tk, tk.txt)
	}
}

func TestReadToken(t *testing.T) {
	var r tokenReader

	testTokens := []struct {
		txt string
		typ tokenType
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
			t.Fatal(tt, "expect no error, get:", err)
		} else if tk.typ != tt.typ {
			t.Fatalf("expect %s %s", token{typ: tt.typ}, tt.txt)
		}
	}
}
