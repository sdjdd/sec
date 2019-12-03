package sec

import (
	"reflect"
)

type (
	Vars  map[string]float64
	Funcs map[string]interface{}

	Env struct {
		Vars  Vars
		Funcs Funcs
	}
)

var (
	DefaultParser Parser
	DefaultEnv    = Env{Vars{}, Funcs{}}
)

func Parse(s string) (Expr, error) { return DefaultParser.Parse(s) }
func Eval(s string) (val float64, err error) {
	var expr Expr
	if expr, err = DefaultParser.Parse(s); err != nil {
		return
	}
	return expr.Val(DefaultEnv)
}

// Check returns a non-nil error when at least one illegal function in Funcs.
func (f Funcs) Check() error {
	for fname, fun := range f {
		funcType := reflect.TypeOf(fun)
		if funcType.Kind() != reflect.Func {
			return errIsNotFunc(fname)
		}

		numIn, numOut := funcType.NumIn(), funcType.NumOut()
		switch {
		case numOut == 0:
			return errFuncRetNoVals(fname)
		case numOut > 1:
			return errFuncRetTooManyVals(fname)
		case funcType.Out(0).Kind() != reflect.Float64:
			return errFuncRetNotFloat64(fname)
		}

		if funcType.IsVariadic() {
			numIn--
			if reflect.SliceOf(funcType.In(numIn)).Kind() != reflect.Float64 {
				return errFuncVariadicNotFloat64(fname)
			}
		}
		for i := 0; i < numIn; i++ {
			if funcType.In(i).Kind() != reflect.Float64 {
				return errFuncParam{fname, i + 1}
			}
		}
	}

	return nil
}
