package main

import (
	"errors"
	"fmt"
)

type Calc struct {
	parser parser
	Env    Env
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

	fmt.Println(expr)

	return expr.val(e.Env), nil
}
