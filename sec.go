package main

import "errors"

type Calc struct {
	parser *parser
	Env    Env
}

func New() Calc {
	return Calc{
		parser: new(parser),
	}
}

func (e Calc) Eval(s string) (val float64, err error) {
	expr, err := e.parser.parse(s)
	if err != nil {
		return
	}

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

	return expr.val(e.Env), nil
}
