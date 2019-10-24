package sec

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type parserPanic string
type evalPanic string

type parseErr struct {
	process string
	message string
}

type parser struct {
	lex  lexer
	vars []string
}

type expr interface {
	val(Env) float64
}

type unary struct {
	op   string
	expr expr
}

type binary struct {
	operator string
	l, r     expr
}

type variable string

type literal token

type call struct {
	name string
	args []expr
}

func perr(process, layout string, a ...interface{}) parseErr {
	return parseErr{
		process: process,
		message: fmt.Sprintf(layout, a...),
	}
}

func (e parseErr) Error() string { return e.message }

func (v variable) val(env Env) float64 {
	val, ok := env.Vars[string(v)]
	if !ok {
		panic(evalPanic(fmt.Sprintf("undeclared variable %q", v)))
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

func (u unary) val(env Env) float64 {
	switch u.op {
	case "+":
		return +u.expr.val(env)
	case "-":
		return -u.expr.val(env)
	}

	panic(evalPanic(fmt.Sprintf("unsupported unary operator %q", u.op)))
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

	panic(fmt.Sprintf("the operator %q is not implemented", b.operator))
}

func (c call) val(env Env) float64 {
	rawFunc, ok := env.Funcs[c.name]
	if !ok {
		panic(evalPanic(fmt.Sprintf("undeclared function %q", c.name)))
	}

	ft := reflect.TypeOf(rawFunc)
	ni, no := ft.NumIn(), ft.NumOut()

	if ft.Kind() != reflect.Func {
		panic(evalPanic(fmt.Sprintf("%q is not a function", c.name)))
	} else if no == 0 {
		panic(evalPanic("function must return a value"))
	} else if no > 1 {
		panic(evalPanic("function can return only one value"))
	} else if ft.Out(0).Kind() != reflect.Float64 {
		panic(evalPanic("function must return a float64 value"))
	}

	for i := 0; i < ni; i++ {
		if ft.In(i).Kind() != reflect.Float64 {
			if ft.IsVariadic() && i == ni-1 {
				break
			}
			panic(evalPanic(fmt.Sprintf("argument %d of %q is not float64",
				i+1, c.name)))
		}
	}

	minArgs := ni
	if ft.IsVariadic() {
		minArgs--
	}
	if len(c.args) < minArgs {
		panic(evalPanic(fmt.Sprintf("too few arguments to call %q", c.name)))
	} else if len(c.args) > minArgs && !ft.IsVariadic() {
		panic(evalPanic(fmt.Sprintf("too many arguments to call %q", c.name)))
	}

	args := make([]reflect.Value, len(c.args))
	for i, arg := range c.args {
		args[i] = reflect.ValueOf(arg.val(env))
	}

	return reflect.ValueOf(rawFunc).Call(args)[0].Float()
}

func (p *parser) parse(script string) (exp expr, err error) {
	err = p.lex.tokenize(script)
	if err != nil {
		return
	}

	p.vars = p.vars[:0]

	defer func() {
		switch t := recover().(type) {
		case nil: // do nothing
		case parserPanic:
			err = errors.New(string(t))
		case parseErr:
			err = fmt.Errorf("process: %s, message: %s", t.process, t.message)
		default:
			panic(t)
		}
	}()

	if !p.lex.next() {
		err = errors.New("empty input")
		return
	}

	exp = p.parseAdditive()

	if exp == nil || p.lex.next() {
		panic(perr("parse", "unexpected %q at %d", p.lex.token.txt, p.lex.token.col))
	}
	return
}

// Additive = Multiplicative ('+' Multiplicative)*
func (p *parser) parseAdditive() (e expr) {
	if e = p.parseMultiplicative(); e == nil {
		return
	}

	for p.lex.next() {
		op := p.lex.token
		if !op.eq("+", "-") || !p.lex.next() {
			p.lex.unread(1)
			break
		}
		right := p.parseMultiplicative()
		if right == nil {
			p.lex.unread(2)
			break
		}
		e = binary{op.txt, e, right}
	}

	return
}

// Multiplicative = Unary ('*' Unary)*
func (p *parser) parseMultiplicative() (e expr) {
	if e = p.parseUnary(); e == nil {
		return
	}

	for p.lex.next() {
		op := p.lex.token
		if !op.eq("*", "/", "%") || !p.lex.next() {
			p.lex.unread(1)
			break
		}
		right := p.parseUnary()
		if right == nil {
			p.lex.unread(2)
			break
		}
		e = binary{op.txt, e, right}
	}

	return
}

// Unary = '+' Unary
//       | Primary
func (p *parser) parseUnary() (e expr) {
	if p.lex.token.typ == operator {
		op := p.lex.token
		if !p.lex.next() {
			p.lex.unread(1)
			return nil
		}
		return unary{op.txt, p.parseUnary()}
	}

	return p.parsePrimary()
}

// Primary = identifier
//         | number
//         | identifier '(' Additive ')'
//         | '(' Additive ')'
func (p *parser) parsePrimary() (e expr) {
	token := p.lex.token
	switch token.typ {
	case identifier:
		// identifier + '(' :function call
		if p.lex.next() {
			if p.lex.token.typ == lBracket {
				var args []expr
				for {
					if !p.lex.next() {
						panic(perr("primary", "unexpected EOF, want ')'"))
					}
					arg := p.parseAdditive()
					if arg == nil {
						break
					}
					args = append(args, arg)
					if !p.lex.next() {
						panic(perr("primary", "unexpected EOF, want ')'"))
					}
					if p.lex.token.typ != comma {
						break
					}
				}
				if p.lex.token.typ != rBracket {
					panic(perr("primary", "unexpected %q, want ')'", p.lex.token))
				}
				return call{token.txt, args}
			}
			p.lex.unread(1)
		}
		p.vars = append(p.vars, token.txt)
		return variable(token.txt)
	case integer, float, binLiteral, octLiteral, hexLiteral:
		return literal(token)
	case lBracket: // '('
		if !p.lex.next() {
			panic(parserPanic("unexpected EOF"))
		}
		addStmt := p.parseAdditive()
		if !p.lex.next() || p.lex.token.typ != rBracket {
			msg := fmt.Sprintf("expect ')' at %d", p.lex.token.col)
			panic(parserPanic(msg))
		}
		return addStmt
	}

	return nil
}
