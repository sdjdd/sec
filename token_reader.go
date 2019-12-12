package sec

import (
	"fmt"
	"io"
	"strings"
)

type (
	tokenType uint8

	SourceInfo struct {
		Row, col int
	}

	token struct {
		SourceInfo
		// token type
		typ tokenType
		// token text
		txt string

		afterBlank bool

		afterNewLine bool
	}

	tokenReader struct {
		SourceInfo
		// expression source
		src strings.Reader
		// store current token's text
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

func (s *SourceInfo) NextLine() {
	s.Row, s.col = s.Row+1, 1
}

func isAlpha(ch rune) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func isNumber(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isBlank(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func (s SourceInfo) String() string {
	return fmt.Sprintf("[%d, %d]", s.Row, s.col)
}

func (s SourceInfo) wrapErr(err error) error { return secError{s, err} }

func (t token) String() (str string) {
	switch t.typ {
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
	t.Row, t.col = 1, 1
}

func (t *tokenReader) read() (tk token, err error) {
	tk.SourceInfo = t.SourceInfo

	if t.src.Len() == 0 {
		err = io.EOF
		return
	}

	var ch rune
loop:
	for {
		ch, _, err = t.src.ReadRune()
		if err != nil {
			err = nil
			break
		}

		switch tk.typ {
		case initial:
			if isBlank(ch) {
				t.col++
			} else if ch == '\r' || ch == '\n' {
				if ch == '\r' {
					ch, _, _ = t.src.ReadRune()
					if ch != '\n' {
						err = tk.wrapErr(ErrUnexpected{'\r'})
						return
					}
				}
				t.Row++
				t.col = 1
				tk.Row, tk.col = t.Row, t.col
			} else if isAlpha(ch) || ch == '_' {
				t.text.WriteRune(ch)
				tk.typ = identifier
			} else if ch == '0' {
				t.text.WriteRune(ch)
				tk.typ = zero
			} else if ch >= '1' && ch <= '9' {
				t.text.WriteRune(ch)
				tk.typ = integer
			} else {
				t.text.WriteRune(ch)
				switch ch {
				case '+':
					tk.typ = plus
					break loop
				case '-':
					tk.typ = minus
					break loop
				case '*':
					tk.typ = star
				case '/':
					tk.typ = slash
				case '%':
					tk.typ = percent
					break loop
				case '(':
					tk.typ = lBracket
					break loop
				case ')':
					tk.typ = rBracket
					break loop
				case ',':
					tk.typ = comma
					break loop
				default:
					err = tk.wrapErr(ErrUnexpected{ch})
					return
				}
			}
		case star:
			if ch == '*' {
				t.text.WriteRune(ch)
				tk.typ = doubleStar
			} else {
				t.src.UnreadRune()
			}
			break loop
		case slash:
			if ch == '/' {
				t.text.WriteRune(ch)
				tk.typ = doubleSlash
			} else {
				t.src.UnreadRune()
			}
			break loop
		case identifier:
			if isAlpha(ch) || isNumber(ch) || ch == '_' {
				t.text.WriteRune(ch)
			} else {
				t.src.UnreadRune()
				break loop
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
					t.src.UnreadRune()
					break loop
				}
			}
		case binLiteralPrefix:
			if ch == '0' || ch == '1' {
				t.text.WriteRune(ch)
				tk.typ = binLiteral
			} else {
				t.src.UnreadRune()
				break loop
			}
		case octLiteralPrefix:
			if ch >= '0' && ch <= '7' {
				t.text.WriteRune(ch)
				tk.typ = octLiteral
			} else {
				t.src.UnreadRune()
				break loop
			}
		case hexLiteralPrefix:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				t.text.WriteRune(ch)
				tk.typ = hexLiteral
			} else {
				t.src.UnreadRune()
				break loop
			}
		case binLiteral:
			if ch == '0' || ch == '1' {
				t.text.WriteRune(ch)
			} else {
				t.src.UnreadRune()
				break loop
			}
		case octLiteral:
			if ch >= '0' && ch <= '7' {
				t.text.WriteRune(ch)
			} else {
				t.src.UnreadRune()
				break loop
			}
		case hexLiteral:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				t.text.WriteRune(ch)
			} else {
				t.src.UnreadRune()
				break loop
			}
		case integer:
			if isNumber(ch) {
				t.text.WriteRune(ch)
			} else if ch == '.' {
				t.text.WriteRune(ch)
				tk.typ = float
			} else {
				t.src.UnreadRune()
				break loop
			}
		case float:
			if isNumber(ch) {
				t.text.WriteRune(ch)
			} else {
				t.src.UnreadRune()
				break loop
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
			err = next.wrapErr(ErrUnexpected{[]rune(next.txt)[0]})
		} else if next.typ == integer {
			err = next.wrapErr(errInvalidDigitInLiteral{n, []rune(next.txt)[0]})
		}
	}

	return
}
