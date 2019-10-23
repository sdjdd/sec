package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

type parserPanic string

type parser struct {
	lex lexer
}

type expr interface {
	val(Env) float64
}

type binary struct {
	operator string
	l, r     expr
}

type Env map[string]float64

type variable string

type literal token

func (v variable) val(env Env) float64 {
	val, ok := env[string(v)]
	if !ok {
		panic(parserPanic(fmt.Sprintf("undeclared variable %q", v)))
	}
	return val
}

func (l literal) val(_ Env) (v float64) {
	switch l.typ {
	case integer, float:
		v, _ = strconv.ParseFloat(l.txt, 64)
	case binLiteral, octLiteral, hexLiteral:
		t, _ := strconv.ParseInt(l.txt, 0, 64)
		v = float64(t)
	default:
		panic(fmt.Sprintf("unsupported token %s", token(l)))
	}
	return
}

func (b binary) val(env Env) float64 {
	l, r := b.l.val(env), b.r.val(env)
	switch b.operator {
	case "+":
		return l + r
	case "-":
		return l - r
	case "*":
		return l * r
	case "/":
		return l / r
	case "%":
		return math.Mod(l, r)
	}

	panic(fmt.Sprintf("The operator %q is not implemented", b.operator))
}

func (p *parser) parse(script string) (exp expr, err error) {
	err = p.lex.tokenize(script)
	if err != nil {
		return
	}
	fmt.Println(p.lex.tokens)

	defer func() {
		switch t := recover().(type) {
		case nil: // do nothing
		case lexerPanic:
			err = errors.New(string(t))
		case parserPanic:
			err = errors.New(string(t))
		default:
			panic(t)
		}
	}()

	if !p.lex.next() {
		err = errors.New("empty input")
		return
	}

	exp = p.parseAdditive()

	fmt.Printf("%+v\n", p.lex)
	if p.lex.next() {
		msg := fmt.Sprintf("unexpected %q at %d", p.lex.peek().txt, p.lex.col)
		panic(parserPanic(msg))
	}
	return
}

// Addition = Multiplication ('+' Multiplication)*
func (p *parser) parseAdditive() (e expr) {
	left := p.parseMultiplicative()
	if left == nil {
		panic("expect multiplicative statement")
	}

	e = left
	for p.lex.next() {
		op := p.lex.peek()
		if !op.eq("+", "-") {
			p.lex.unread(1)
			break
		}
		if !p.lex.next() {
			panic(parserPanic("unexpected EOF"))
		}
		right := p.parseMultiplicative()
		if right == nil {
			p.lex.unread(2)
			break
		}
		e = binary{op.txt, e, right}
	}

	return e
}

// Multiplication = Primary ('*' Primary)*
func (p *parser) parseMultiplicative() (e expr) {
	left := p.parsePrimary()
	if left == nil {
		panic("expect primary statement")
	}

	e = left
	for p.lex.next() {
		op := p.lex.peek()
		if !op.eq("*", "/", "%") || !p.lex.next() {
			p.lex.unread(1)
			break
		}
		right := p.parsePrimary()
		if right == nil {
			panic("expect primary statement")
		}
		e = binary{op.txt, e, right}
	}

	return e
}

func (p *parser) parsePrimary() (e expr) {
	tk := p.lex.peek()
	switch tk.typ {
	case identifier:
		return variable(tk.txt)
	case integer, float, binLiteral, octLiteral, hexLiteral:
		return literal(tk)
	case lBracket: // '('
		if !p.lex.next() {
			panic(parserPanic("unexpected EOF"))
		}
		addStmt := p.parseAdditive()
		if !p.lex.next() || p.lex.peek().typ != rBracket {
			msg := fmt.Sprintf("expect ')' at %d", p.lex.col+1)
			panic(parserPanic(msg))
		}
		return addStmt
	}

	msg := fmt.Sprintf("unexpected %q at %d", tk.txt, p.lex.col)
	panic(parserPanic(msg))
}
