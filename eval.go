package sec

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type Expr interface {
	Val(env Env) (val float64, err error)
}

type (
	unary struct {
		op   token
		expr Expr
	}

	binary struct {
		op   token
		l, r Expr
	}

	variable token

	literal token

	call struct {
		token
		args []Expr
	}
)

type AstError struct {
	srcInfo
	err error
}

func (a AstError) Error() string {
	return fmt.Sprintf("[%d, %d]: %s", a.row, a.col, a.err)
}

func (s srcInfo) errorf(format string, a ...interface{}) error {
	return AstError{s, fmt.Errorf(format, a...)}
}

func (v variable) Val(env Env) (val float64, err error) {
	var ok bool
	if val, ok = env.Vars[v.txt]; !ok {
		err = v.errorf("Undeclared variable %q", v.txt)
	}
	return
}

func (l literal) Val(_ Env) (val float64, err error) {
	switch l.typ {
	case integer, float:
		val, _ = strconv.ParseFloat(l.txt, 64)
	case binLiteral, octLiteral, hexLiteral:
		t, _ := strconv.ParseInt(l.txt, 0, 64)
		val = float64(t)
	}
	return
}

func (u unary) Val(env Env) (val float64, err error) {
	if val, err = u.expr.Val(env); err != nil {
		return
	}
	switch u.op.typ {
	case plus:
		val = +val
	case minus:
		val = -val
	}
	return
}

func (b binary) Val(env Env) (val float64, err error) {
	var left, right float64
	if left, err = b.l.Val(env); err != nil {
		return
	}
	if right, err = b.r.Val(env); err != nil {
		return
	}

	switch b.op.typ {
	case plus:
		val = left + right
	case minus:
		val = left - right
	case star:
		val = left * right
	case slash:
		val = left / right
	case doubleSlash:
		val = math.Floor(left / right)
	case percent:
		val = math.Mod(left, right)
	case doubleStar:
		val = math.Pow(left, right)
	}
	return
}

func (c call) Val(env Env) (val float64, err error) {
	fun, ok := env.Funcs[c.txt]
	if !ok {
		err = c.errorf("Undeclared function %q", c.txt)
		return
	}

	ftype := reflect.TypeOf(fun)

	argc := ftype.NumIn()
	if ftype.IsVariadic() {
		argc--
	}

	if len(c.args) < argc {
		err = c.errorf("too few arguments to call %q", c.txt)
		return
	} else if len(c.args) > argc && !ftype.IsVariadic() {
		err = c.errorf("too many arguments to call %q", c.txt)
		return
	}

	args := make([]reflect.Value, len(c.args))
	for i, arg := range c.args {
		if val, err = arg.Val(env); err != nil {
			return
		}
		args[i] = reflect.ValueOf(val)
	}

	results := reflect.ValueOf(fun).Call(args)
	val = results[0].Float()

	return
}
