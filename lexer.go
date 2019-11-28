package sec

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type lexer struct {
	tokens []token
	token  token        // current token
	buf    bytes.Buffer // reading buffer
	col    int          // current token's column
	index  int          // read index
}

type lexerPanic string

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

	operator

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

func (t token) eq(texts ...string) bool {
	for _, txt := range texts {
		if t.txt == txt {
			return true
		}
	}
	return false
}

func (t token) is(types ...int) bool {
	for _, typ := range types {
		if t.typ == typ {
			return true
		}
	}
	return false
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

	//lineInfo := fmt.Sprintf("[%d,%d]", t.row, t.col)
	return fmt.Sprintf("<%-14s %q>", typ, t.txt)
}

func (l *lexer) init(ch rune) {
	l.flush()
	if isAlpha(ch) {
		l.token.typ = identifier
	} else if isNumber(ch) {
		if ch == '0' {
			l.token.typ = zero
		} else {
			l.token.typ = integer
		}
	} else if isBlank(ch) {
		l.col++
		return // ignore blank
	} else {
		switch ch {
		case '+', '-', '*', '/', '%':
			l.token.typ = operator
		case '(':
			l.token.typ = lBracket
		case ')':
			l.token.typ = rBracket
		case ',':
			l.token.typ = comma
		default:
			msg := fmt.Sprintf("invalid character %q at %d", ch, l.col)
			panic(lexerPanic(msg))
		}
	}
	l.token.col = l.col
	l.write(ch)
}

func (l *lexer) unread(i int) {
	for ; i > 0 && l.index > 0; i-- {
		l.index--
	}
	l.token = l.tokens[l.index]
}

func (l *lexer) write(ch rune) {
	l.buf.WriteRune(ch)
	l.col++
}

func (l *lexer) flush() {
	if l.buf.Len() == 0 {
		return
	}

	defer func() {
		l.token.txt = l.buf.String()
		l.tokens = append(l.tokens, l.token)
		l.token = token{}
		l.buf.Reset()
	}()

	switch l.token.typ {
	case binLiteral:
		if l.buf.Len() == 2 {
			msg := fmt.Sprintf("invalid binary literal %q at %d",
				l.buf.String(), l.col-1)
			panic(lexerPanic(msg))
		}
	case octLiteral:
		if l.buf.Len() == 2 {
			str := l.buf.String()
			if str == "0o" || str == "0O" {
				msg := fmt.Sprintf("invalid octal literal %q at %d",
					l.buf.String(), l.col-1)
				panic(lexerPanic(msg))
			}
		}
	case hexLiteral:
		if l.buf.Len() == 2 {
			msg := fmt.Sprintf("invalid hexdecimal literal %q at %d",
				l.buf.String(), l.col-1)
			panic(lexerPanic(msg))
		}
	case zero:
		l.token.typ = integer
	}
}

func (l *lexer) tokenize(script string) (err error) {
	l.tokens = l.tokens[:0]
	l.index = -1
	l.col = 1

	defer func() {
		switch t := recover().(type) {
		case nil: // no panic
		case lexerPanic:
			err = errors.New(string(t))
		default:
			panic(t)
		}
		l.col = 0
	}()

	expr := strings.NewReader(script)
	for {
		ch, _, err := expr.ReadRune()
		if err != nil {
			break
		}

		if l.buf.Len() == 0 {
			l.init(ch)
			continue
		}
		switch l.token.typ {
		case lBracket:
			fallthrough
		case rBracket:
			fallthrough
		case operator:
			fallthrough
		case comma:
			l.init(ch)
		case identifier:
			if isAlpha(ch) || isNumber(ch) {
				l.write(ch)
			} else {
				l.init(ch)
			}
		case zero:
			if isNumber(ch) || ch == 'o' || ch == 'O' {
				l.write(ch)
				l.token.typ = octLiteral
			} else if ch == 'b' || ch == 'B' {
				l.write(ch)
				l.token.typ = binLiteral
			} else if ch == 'x' || ch == 'X' {
				l.write(ch)
				l.token.typ = hexLiteral
			} else {
				l.init(ch)
			}
		case integer:
			if isNumber(ch) {
				l.write(ch)
			} else if ch == '.' {
				l.write(ch)
				l.token.typ = float
			} else {
				l.init(ch)
			}
		case float:
			if isNumber(ch) {
				l.write(ch)
			} else {
				l.init(ch)
			}
		case binLiteral:
			if ch == '0' || ch == '1' {
				l.write(ch)
			} else {
				l.init(ch)
			}
		case octLiteral:
			if ch >= '0' && ch <= '7' {
				l.write(ch)
			} else {
				l.init(ch)
			}
		case hexLiteral:
			if isNumber(ch) || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F' {
				l.write(ch)
			} else {
				l.init(ch)
			}
		}
	}
	l.flush()

	return
}

func (l *lexer) next() bool {
	if l.index < len(l.tokens)-1 {
		l.index++
		l.token = l.tokens[l.index]
		return true
	}
	return false
}
