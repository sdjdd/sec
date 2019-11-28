package sec

import (
	"fmt"
	"io"
	"strings"
)

type token struct {
	// token's begin column
	row, col int

	// token type
	typ int

	// token text
	txt string
}

type tokenReader struct {
	// source script
	src strings.Reader

	// store current token's text
	tbuf strings.Builder

	// readed tokens
	tokens []token

	// read position
	readPos int

	// current row and column
	row, col int

	lastRowCols int
}

func newTokenReader(src string) *tokenReader {
	t := new(tokenReader)
	t.row = 1
	t.col = 1
	t.load(src)
	return t
}

func (t *tokenReader) load(src string) {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	t.src.Reset(src)
	t.tokens = []token{}
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
				// do nothing
			} else if ch == '\n' {
				t.row++
				t.col, t.lastRowCols = 0, t.col
			} else if isAlpha(ch) || ch == '_' {
				tk.typ = identifier
				t.tbuf.WriteRune(ch)
			} else if isNumber(ch) {
				if ch == '0' {
					tk.typ = zero
				} else {
					tk.typ = integer
				}
				t.tbuf.WriteRune(ch)
			} else {
				switch ch {
				case '+':
					tk.typ = plus
					t.tbuf.WriteRune(ch)
					break readLoop
				case '-':
					tk.typ = minus
					t.tbuf.WriteRune(ch)
					break readLoop
				case '*':
					tk.typ = star
					t.tbuf.WriteRune(ch)
				case '/':
					tk.typ = slash
					t.tbuf.WriteRune(ch)
				case '(':
					tk.typ = lBracket
					t.tbuf.WriteRune(ch)
					break readLoop
				case ')':
					tk.typ = rBracket
					t.tbuf.WriteRune(ch)
					break readLoop
				case ',':
					tk.typ = comma
					t.tbuf.WriteRune(ch)
					break readLoop
				default:
					err = fmt.Errorf("[%d, %d]: invalid character %q", t.row, t.col, ch)
					return
				}
			}
		case star:
			if ch == '*' {
				tk.typ = doubleStar
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
			}
			break readLoop
		case slash:
			if ch == '/' {
				tk.typ = doubleSlash
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
			}
			break readLoop
		case identifier:
			if isAlpha(ch) || isNumber(ch) || ch == '_' {
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case zero:
			if isNumber(ch) {
				tk.typ = octLiteral
				t.tbuf.WriteRune(ch)
			} else if ch == 'b' || ch == 'B' {
				tk.typ = binLiteralPrefix
				t.tbuf.WriteRune(ch)
			} else if ch == 'o' || ch == 'O' {
				tk.typ = octLiteralPrefix
				t.tbuf.WriteRune(ch)
			} else if ch == 'x' || ch == 'X' {
				tk.typ = hexLiteralPrefix
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case binLiteralPrefix:
			if ch == '0' || ch == '1' {
				tk.typ = binLiteral
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case octLiteralPrefix:
			if ch >= '0' && ch <= '7' {
				tk.typ = octLiteral
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case hexLiteralPrefix:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				tk.typ = hexLiteral
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case binLiteral:
			if ch == '0' || ch == '1' {
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case octLiteral:
			if ch >= '0' && ch <= '7' {
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case hexLiteral:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case integer:
			if isNumber(ch) {
				t.tbuf.WriteRune(ch)
			} else if ch == '.' {
				tk.typ = float
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		case float:
			if isNumber(ch) {
				t.tbuf.WriteRune(ch)
			} else {
				t.srcUnread()
				break readLoop
			}
		default:
			panic("Not handling all possible cases")
		}
	}
	tk.row = t.row
	tk.col = t.col - t.tbuf.Len()
	tk.txt = t.tbuf.String()

	t.tokens = append(t.tokens, tk)
	t.readPos++

	t.tbuf.Reset()

	return
}
