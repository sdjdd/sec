package sec

import (
	"fmt"
	"io"
	"strings"
)

type srcInfo struct {
	row, col int
}

type token struct {
	srcInfo

	// token type
	typ int

	// token text
	txt string

	afterBlank   bool
	afterNewLine bool
}

type tokenReader struct {
	// source script
	src strings.Reader

	// store current token's text
	text strings.Builder

	// readed tokens
	tokens []token

	// read position
	readPos int

	// current row and column
	srcInfo

	lastRowCols int
}

const (
	initial = iota
	identifier
	zero
	integer
	float

	binLiteralPrefix
	octLiteralPrefix
	hexLiteralPrefix
	binLiteral
	octLiteral
	hexLiteral

	lBracket    // '('
	rBracket    // ')'
	comma       // ','
	plus        // '+'
	minus       // '-'
	star        // '*'
	slash       // '/'
	doubleStar  // '**'
	doubleSlash // '//'
)

func isAlpha(ch rune) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func isNumber(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isBlank(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func (t token) String() string {
	var typ string
	switch t.typ {
	case identifier:
		typ = "identifier"
	case zero:
		typ = "zero"
	case integer:
		typ = "integer"
	case float:
		typ = "float"
	case binLiteralPrefix:
		typ = "bin-literal-prefix"
	case octLiteralPrefix:
		typ = "oct-literal-prefix"
	case hexLiteralPrefix:
		typ = "hex-literal-prefix"
	case binLiteral:
		typ = "bin-literal"
	case octLiteral:
		typ = "oct-literal"
	case hexLiteral:
		typ = "hex-literal"
	case lBracket:
		typ = "left-bracket"
	case rBracket:
		typ = "right-bracket"
	case comma:
		typ = "comma"
	case plus:
		typ = "plus"
	case minus:
		typ = "minus"
	case star:
		typ = "star"
	case slash:
		typ = "slash"
	case doubleStar:
		typ = "double-star"
	case doubleSlash:
		typ = "double-slash"
	}

	return fmt.Sprintf("<%-14s %q>", typ, t.txt)
}

func (t *tokenReader) load(src string) {
	t.src.Reset(src)
	t.tokens = []token{}
	t.row, t.col = 1, 1
}

func (t *tokenReader) srcUnread() {
	t.src.UnreadRune()
	if t.col == 0 {
		t.col = t.lastRowCols
		t.row--
	} else {
		t.col--
	}
}

func (t *tokenReader) read() (tk token, err error) {
	if t.readPos < len(t.tokens) {
		tk = t.tokens[t.readPos]
		t.readPos++
		return
	}

	if t.src.Len() == 0 {
		err = io.EOF
		return
	}

	var afterBlank, afterNewLine bool
readLoop:
	for {
		ch, _, er := t.src.ReadRune()
		if er != nil {
			break
		}
		t.col++

		switch tk.typ {
		case initial:
			if isBlank(ch) {
				afterBlank = true
			} else if ch == '\r' {
				nch, _, er := t.src.ReadRune()
				if er != nil || nch != '\n' {
					tk.row, tk.col = t.row, t.col-1 // update tk'a srcInfo
					err = tk.errorf(`unexpected '\r'`)
					return
				}
				t.src.UnreadRune()
			} else if ch == '\n' {
				afterNewLine = true
				t.row++
				t.col, t.lastRowCols = 1, t.col
			} else if isAlpha(ch) || ch == '_' {
				tk.typ = identifier
				t.text.WriteRune(ch)
			} else if isNumber(ch) {
				if ch == '0' {
					tk.typ = zero
				} else {
					tk.typ = integer
				}
				t.text.WriteRune(ch)
			} else {
				switch ch {
				case '+':
					tk.typ = plus
					t.text.WriteRune(ch)
					break readLoop
				case '-':
					tk.typ = minus
					t.text.WriteRune(ch)
					break readLoop
				case '*':
					tk.typ = star
					t.text.WriteRune(ch)
				case '/':
					tk.typ = slash
					t.text.WriteRune(ch)
				case '(':
					tk.typ = lBracket
					t.text.WriteRune(ch)
					break readLoop
				case ')':
					tk.typ = rBracket
					t.text.WriteRune(ch)
					break readLoop
				case ',':
					tk.typ = comma
					t.text.WriteRune(ch)
					break readLoop
				default:
					tk.row, tk.col = t.row, t.col-1
					err = tk.errorf("invalid character %q", ch)
					return
				}
			}
		case star:
			if ch == '*' {
				tk.typ = doubleStar
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
			}
			break readLoop
		case slash:
			if ch == '/' {
				tk.typ = doubleSlash
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
			}
			break readLoop
		case identifier:
			if isAlpha(ch) || isNumber(ch) || ch == '_' {
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case zero:
			if isNumber(ch) {
				tk.typ = octLiteral
				t.text.WriteRune(ch)
			} else if ch == 'b' || ch == 'B' {
				tk.typ = binLiteralPrefix
				t.text.WriteRune(ch)
			} else if ch == 'o' || ch == 'O' {
				tk.typ = octLiteralPrefix
				t.text.WriteRune(ch)
			} else if ch == 'x' || ch == 'X' {
				tk.typ = hexLiteralPrefix
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case binLiteralPrefix:
			if ch == '0' || ch == '1' {
				tk.typ = binLiteral
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case octLiteralPrefix:
			if ch >= '0' && ch <= '7' {
				tk.typ = octLiteral
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case hexLiteralPrefix:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				tk.typ = hexLiteral
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case binLiteral:
			if ch == '0' || ch == '1' {
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case octLiteral:
			if ch >= '0' && ch <= '7' {
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case hexLiteral:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case integer:
			if isNumber(ch) {
				t.text.WriteRune(ch)
			} else if ch == '.' {
				tk.typ = float
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case float:
			if isNumber(ch) {
				t.text.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		default:
			panic("Not handling all possible cases")
		}
	}

	tk.row = t.row
	tk.col = t.col - t.text.Len()
	tk.txt = t.text.String()
	tk.afterBlank, tk.afterNewLine = afterBlank, afterNewLine

	t.tokens = append(t.tokens, tk)
	t.readPos++

	t.text.Reset()

	switch tk.typ {
	case zero:
		tk.typ = integer
	case binLiteralPrefix, octLiteralPrefix, hexLiteralPrefix:
		err = explainLiteralPrefixError(tk, t)
	}

	return
}

func explainLiteralPrefixError(tk token, r *tokenReader) (err error) {
	var ltype string
	switch tk.typ {
	case binLiteralPrefix:
		ltype = "binary"
	case octLiteralPrefix:
		ltype = "octal"
	case hexLiteralPrefix:
		ltype = "hexadecimal"
	default:
		return nil
	}
	err = tk.errorf(ltype + " literal has no digits")

	next, er := r.read()
	if er == nil && !next.afterBlank && !next.afterNewLine {
		if tk.typ == hexLiteralPrefix {
			err = next.errorf("unexpected %q", next.txt)
		} else if next.typ == integer {
			err = next.errorf("invalid digit %q in binary literal", next.txt[0])
		}
	}

	return
}
