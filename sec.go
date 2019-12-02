package sec

import (
	"errors"
	"fmt"
	"reflect"
)

type Calc struct {
	parser parser

	Env Env
}

type Env struct {
	Vars  Vars
	Funcs Funcs
}

type (
	Vars  map[string]float64
	Funcs map[string]interface{}
)

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

	if err = c.CheckFuncs(); err != nil {
		return
	}

	val, err = expr.Val(c.Env)
	return
}

// CheckFuncs returns a non-nil error when at least one illegal function in
// Calc.Env.Funcs.
func (c Calc) CheckFuncs() error {
	for fname, function := range c.Env.Funcs {
		funcType := reflect.TypeOf(function)
		numIn, numOut := funcType.NumIn(), funcType.NumOut()

		if funcType.Kind() != reflect.Func {
			return fmt.Errorf("%q is not a function", fname)
		} else if numOut == 0 {
			return errors.New("function must return a value")
		} else if numOut > 1 {
			return errors.New("function must return only one value")
		} else if funcType.Out(0).Kind() != reflect.Float64 {
			return errors.New("function must return a float64 value")
		}

		for i := 0; i < numIn; i++ {
			if funcType.In(i).Kind() != reflect.Float64 {
				if funcType.IsVariadic() && i == numIn-1 {
					break
				}
				return fmt.Errorf("parameter %d of %q is not float64", i+1, fname)
			}
		}
	}

	return nil
}
