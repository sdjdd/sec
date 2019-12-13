package sec

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadEmptyText(t *testing.T) {
	var r tokenReader
	// r.load("")
	tk, err := r.read()
	if err != io.EOF || tk.typ != EOF {
		t.Fatal("expect EOF tokan and EOF error")
	}
}

func TestReadInvalidLiteral(t *testing.T) {
	var r tokenReader
	{
		var lerr ErrLiteralNoDigit
		for _, txt := range []string{"0b", "0B", "0o", "0O", "0x", "0X"} {
			r.load(txt)
			if _, err := r.read(); !errors.As(err, &lerr) {
				t.Fatal("expect ErrLiteralNoDigit error")
			}
		}
	}
	{
		var lerr ErrInvalidDigitInLiteral
		for _, txt := range []string{"0b2", "0B2", "0o8", "0O8"} {
			r.load(txt)
			if _, err := r.read(); !errors.As(err, &lerr) {
				t.Fatal("expect ErrInvalidDigitInLiteral error")
			}
		}
	}
}

func TestNewLine(t *testing.T) {
	var r tokenReader

	r.load("\r\ntest")
	if _, err := r.read(); err != nil {
		t.Fatal("expect no error")
	}

	r.load("\ntest")
	if _, err := r.read(); err != nil {
		t.Fatal("expect no error")
	}

	r.load("\rtest")
	var uerr ErrUnexpected
	if _, err := r.read(); !errors.As(err, &uerr) {
		t.Fatal("expect ErrUnexpected error")
	} else if uerr.Char != '\r' {
		t.Fatal("expect ErrUnexpected{'\\r'} error")
	}
}

func TestTokenPosition(t *testing.T) {
	tokenTexts := []string{"identifier", "0", "114514", "3.14159", "0b10101101",
		"0755", "0xFAFAFA", "-", "//"}
	positions := []struct {
		row, col     int
		blanksBefore int
	}{
		{1, 0, 3}, {1, 0, 5}, {1, 0, 7},
		{2, 0, 4}, {2, 0, 8}, {2, 0, 16},
		{3, 0, 1}, {3, 0, 100},
	}

	var r tokenReader
	buf := new(strings.Builder)
	for _, text := range tokenTexts {
		buf.Reset()
		row, col := 1, 1
		for i, p := range positions {
			for p.row > row {
				buf.WriteByte('\n')
				row++
				col = 1
			}
			for i := 0; i < p.blanksBefore; i++ {
				buf.WriteByte(' ')
				col++
			}
			positions[i].col = col
			buf.WriteString(text)
			col += len(text)
		}

		r.load(buf.String())

		for i, p := range positions {
			token, err := r.read()
			if err != nil {
				t.Fatal(i, "unexpected error:", err)
			} else if token.Row != p.row || token.Col != p.col {
				t.Fatalf("{%s; %d}: expect [%d,%d], get [%d,%d]",
					text, i, p.row, p.col, token.Row, token.Col)
			}
		}
	}
}

func TestReadToken(t *testing.T) {
	tokenGroup := map[tokenType][]string{
		identifier:  {"id", "_", "abc123", "_000"},
		integer:     {"0", "114514", "1919810"},
		float:       {"3.14159", "0.5"},
		binLiteral:  {"0b1001", "0B1101"},
		octLiteral:  {"0755", "0o1234", "0O4567"},
		hexLiteral:  {"0xFA", "0Xfa", "0xFa0", "0XfA1"},
		lBracket:    {"("},
		rBracket:    {")"},
		comma:       {","},
		plus:        {"+"},
		minus:       {"-"},
		star:        {"*"},
		percent:     {"%"},
		doubleStar:  {"**"},
		doubleSlash: {"//"},
		EOF:         {"", "\n  \n", "\r\n  \r\n"},
	}

	var r tokenReader
	for typ, tokens := range tokenGroup {
		r.load(strings.Join(tokens, " "))
		for i := 0; i <= len(tokens); i++ {
			token, err := r.read()
			if i < len(tokens) {
				if err != nil && token.typ != EOF {
					t.Fatal("expect no err, got", err, i)
				} else if token.typ != typ {
					t.Fatal("expect:", typ, ", got", token.typ)
				}
			} else if err != io.EOF {
				t.Fatal("expect EOF error, got", err)
			}
		}
	}
}
