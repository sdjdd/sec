package sec

import (
	"errors"
)

type Calc struct {
	parser parser

	Env Env

	BeforeEval func(env Env, varNames []string)
}

type Env struct {
	Vars  Vars
	Funcs Funcs
}

type Vars map[string]float64
type Funcs map[string]interface{}

func New() (calc Calc) {
	calc.Env.Vars = make(Vars)
	calc.Env.Funcs = make(Funcs)
	return
}

func (c Calc) Eval(s string) (val float64, err error) {
	expr, err := c.parser.parse(s)
	if err != nil {
		return
	}

	defer func() {
		switch t := recover().(type) {
		case nil: // do nothing
		case evalPanic:
			err = errors.New(string(t))
		default:
			panic(t)
		}
	}()

	if c.BeforeEval != nil {
		varNames := make([]string, 0, len(c.parser.vars))
		for name := range c.parser.vars {
			varNames = append(varNames, name)
		}
		c.BeforeEval(c.Env, varNames)
	}

	return expr.val(c.Env), nil
}
