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
			return ErrNotFunction{fname}
		}

		numIn, numOut := funcType.NumIn(), funcType.NumOut()
		switch {
		case numOut == 0:
			return ErrFuncNoReturnVal{fname}
		case numOut > 1:
			return ErrFuncReturnTooManyVal{fname}
		case funcType.Out(0).Kind() != reflect.Float64:
			return ErrReturnValNotFloat64{fname}
		}

		if funcType.IsVariadic() {
			if reflect.SliceOf(funcType.In(numIn-1)).Kind() != reflect.Float64 {
				return ErrParamNotFloat64{fname, numIn}
			}
			numIn--
		}
		for i := 0; i < numIn; i++ {
			if funcType.In(i).Kind() != reflect.Float64 {
				return ErrParamNotFloat64{fname, i + 1}
			}
		}
	}

	return nil
}
