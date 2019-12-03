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

func (s srcInfo) String() string {
	return fmt.Sprintf("[%d, %d]", s.row, s.col)
}

func (t token) wrapErr(err error) error { return tokenErr{t.srcInfo, err} }

func (t token) String() (str string) {
	switch t.typ {
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
	tk.row, tk.col = t.row, t.col

	if t.src.Len() == 0 {
		err = io.EOF
		return
	}

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
				tk.afterBlank = true
				tk.col++
			} else if ch == '\r' {
				ch, _, er := t.src.ReadRune()
				if er != nil || ch != '\n' {
					err = tk.wrapErr(errUnexpected('\r'))
					return
				}
				t.src.UnreadRune()
			} else if ch == '\n' {
				tk.afterNewLine = true
				t.lastRowCols = t.col
				t.row, t.col = t.row+1, 1
				tk.row, tk.col = t.row, t.col
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
				t.text.WriteRune(ch)
				switch ch {
				case '+':
					tk.typ = plus
					break readLoop
				case '-':
					tk.typ = minus
					break readLoop
				case '*':
					tk.typ = star
				case '/':
					tk.typ = slash
				case '%':
					tk.typ = percent
					break readLoop
				case '(':
					tk.typ = lBracket
					break readLoop
				case ')':
					tk.typ = rBracket
					break readLoop
				case ',':
					tk.typ = comma
					break readLoop
				default:
					err = tk.wrapErr(errUnexpected(ch))
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
			} else {
				switch ch {
				case '.':
					tk.typ = float
				case 'b', 'B':
					tk.typ = binLiteralPrefix
				case 'o', 'O':
					tk.typ = octLiteralPrefix
				case 'x', 'X':
					tk.typ = hexLiteralPrefix
				default:
					t.srcUnread()
					break readLoop
				}
				t.text.WriteRune(ch)
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

	tk.txt = t.text.String()

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
	err = tk.wrapErr(errLiteralHasNoDigits(n))

	next, er := r.read()
	if er == nil && !next.afterBlank && !next.afterNewLine {
		if tk.typ == hexLiteralPrefix {
			err = next.wrapErr(errUnexpected(next.txt[0]))
		} else if next.typ == integer {
			err = next.wrapErr(errInvalidDigitInLiteral{n, rune(next.txt[0])})
		}
	}

	return
}
