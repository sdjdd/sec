package sec

import (
	"fmt"
	"io"
	"strings"
)

type (
	tokenType int

	pos struct {
		Row, Col int
	}

	token struct {
		pos

		typ tokenType
		txt string

		afterBlank   bool
		afterNewLine bool
	}

	tokenReader struct {
		pos
		src strings.Reader
		// current token's text
		text strings.Builder
	}
)

const (
	initial tokenType = iota
	identifier
	zero
	integer
	float

	binLiteralPrefix // 0[bB]
	octLiteralPrefix // 0[oO]
	hexLiteralPrefix // 0[xX]
	binLiteral       // 0[bB][01]+
	octLiteral       // 0[oO]?[0-7]+
	hexLiteral       // 0[xX][0-9a-fA-F]+

	lBracket    // '('
	rBracket    // ')'
	comma       // ','
	plus        // '+'
	minus       // '-'
	star        // '*'
	slash       // '/'
	percent     // '%'
	doubleStar  // '**'
	doubleSlash // '//'

	EOF
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

func (s pos) String() string {
	return fmt.Sprintf("[%d, %d]", s.Row, s.Col)
}

func (s pos) wrapErr(err error) error { return secError{s, err} }

func (t tokenType) String() (str string) {
	switch t {
	case initial:
		str = "initial"
	case identifier:
		str = "identifier"
	case zero:
		str = "zero"
	case integer:
		str = "integer"
	case float:
		str = "float"
	case binLiteralPrefix:
		str = "bin-literal-prefix"
	case octLiteralPrefix:
		str = "oct-literal-prefix"
	case hexLiteralPrefix:
		str = "hex-literal-prefix"
	case binLiteral:
		str = "bin-literal"
	case octLiteral:
		str = "oct-literal"
	case hexLiteral:
		str = "hex-literal"
	case lBracket:
		str = "left-bracket"
	case rBracket:
		str = "right-bracket"
	case comma:
		str = "comma"
	case plus:
		str = "plus"
	case minus:
		str = "minus"
	case star:
		str = "star"
	case slash:
		str = "slash"
	case percent:
		str = "percent"
	case doubleStar:
		str = "double-star"
	case doubleSlash:
		str = "double-slash"
	default:
		str = "unknown"
	}
	return
}

func (t *tokenReader) load(src string) {
	t.src.Reset(src)
	t.text.Reset()
	t.Row, t.Col = 1, 1
}

func (t *tokenReader) read() (tk token, err error) {
	var unread, finish bool
	prevLineCols := t.Col

	for !finish {
		ch, _, er := t.src.ReadRune()
		if er != nil {
			break
		}
		t.Col++

		switch tk.typ {
		case initial:
			if isBlank(ch) {
				// skip blank
			} else if ch == '\r' || ch == '\n' {
				if ch == '\r' {
					if ch, _, _ = t.src.ReadRune(); ch != '\n' {
						err = tk.wrapErr(ErrUnexpected{'\r'})
						return
					}
				}
				t.Row++
				prevLineCols = t.Col
				t.Col = 1
			} else {
				tk.Row, tk.Col = t.Row, t.Col-1
				t.text.WriteRune(ch)
				if isAlpha(ch) || ch == '_' {
					tk.typ = identifier
				} else if ch >= '1' && ch <= '9' {
					tk.typ = integer
				} else {
					switch ch {
					case '0':
						tk.typ = zero
					case '+':
						tk.typ = plus
						finish = true
					case '-':
						tk.typ = minus
						finish = true
					case '*':
						tk.typ = star
					case '/':
						tk.typ = slash
					case '%':
						tk.typ = percent
						finish = true
					case '(':
						tk.typ = lBracket
						finish = true
					case ')':
						tk.typ = rBracket
						finish = true
					case ',':
						tk.typ = comma
						finish = true
					default:
						err = tk.wrapErr(ErrUnexpected{ch})
						return
					}
				}
			}
		case star:
			if ch == '*' {
				t.text.WriteRune(ch)
				tk.typ = doubleStar
			} else {
				unread = true
			}
			finish = true
		case slash:
			if ch == '/' {
				t.text.WriteRune(ch)
				tk.typ = doubleSlash
			} else {
				unread = true
			}
			finish = true
		case identifier:
			if isAlpha(ch) || isNumber(ch) || ch == '_' {
				t.text.WriteRune(ch)
			} else {
				unread = true
				finish = true
			}
		case zero:
			if isNumber(ch) {
				t.text.WriteRune(ch)
				tk.typ = octLiteral
			} else {
				switch ch {
				case '.':
					t.text.WriteRune(ch)
					tk.typ = float
				case 'b', 'B':
					t.text.WriteRune(ch)
					tk.typ = binLiteralPrefix
				case 'o', 'O':
					t.text.WriteRune(ch)
					tk.typ = octLiteralPrefix
				case 'x', 'X':
					t.text.WriteRune(ch)
					tk.typ = hexLiteralPrefix
				default:
					unread = true
					finish = true
				}
			}
		case binLiteralPrefix:
			if ch == '0' || ch == '1' {
				t.text.WriteRune(ch)
				tk.typ = binLiteral
			} else {
				unread = true
				finish = true
			}
		case octLiteralPrefix:
			if ch >= '0' && ch <= '7' {
				t.text.WriteRune(ch)
				tk.typ = octLiteral
			} else {
				unread = true
				finish = true
			}
		case hexLiteralPrefix:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				t.text.WriteRune(ch)
				tk.typ = hexLiteral
			} else {
				unread = true
				finish = true
			}
		case binLiteral:
			if ch == '0' || ch == '1' {
				t.text.WriteRune(ch)
			} else {
				unread = true
				finish = true
			}
		case octLiteral:
			if ch >= '0' && ch <= '7' {
				t.text.WriteRune(ch)
			} else {
				unread = true
				finish = true
			}
		case hexLiteral:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				t.text.WriteRune(ch)
			} else {
				unread = true
				finish = true
			}
		case integer:
			if isNumber(ch) {
				t.text.WriteRune(ch)
			} else if ch == '.' {
				t.text.WriteRune(ch)
				tk.typ = float
			} else {
				unread = true
				finish = true
			}
		case float:
			if isNumber(ch) {
				t.text.WriteRune(ch)
			} else {
				unread = true
				finish = true
			}
		default:
			panic("Not handling all possible cases")
		}
	}

	if unread {
		t.src.UnreadRune()
		if t.Col--; t.Col == 0 {
			t.Col = prevLineCols
			t.Row--
		}
	}

	tk.txt = t.text.String()
	t.text.Reset()

	switch tk.typ {
	case initial:
		tk.typ, err = EOF, io.EOF
	case zero:
		tk.typ = integer
	case binLiteralPrefix, octLiteralPrefix, hexLiteralPrefix:
		err = explainLiteralPrefixError(tk, t)
	}

	return
}

func explainLiteralPrefixError(tk token, r *tokenReader) (err error) {
	var n int
	switch tk.typ {
	case binLiteralPrefix:
		n = 2
	case octLiteralPrefix:
		n = 8
	case hexLiteralPrefix:
		n = 16
	default:
		return nil
	}
	err = tk.wrapErr(ErrLiteralNoDigit{n})

	next, er := r.read()
	if er == nil && !next.afterBlank && !next.afterNewLine {
		if tk.typ == hexLiteralPrefix {
			err = next.wrapErr(ErrUnexpected{[]rune(next.txt)[0]})
		} else if next.typ == integer {
			err = next.wrapErr(ErrInvalidDigitInLiteral{n, []rune(next.txt)[0]})
		}
	}

	return
}
