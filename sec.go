package sec

import (
	"errors"
	"math"
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

var MathFuncs = Funcs{
	"abs":         math.Abs,
	"acos":        math.Acos,
	"acosh":       math.Acosh,
	"asin":        math.Asin,
	"asinh":       math.Asinh,
	"atan":        math.Atan,
	"atan2":       math.Atan2,
	"atanh":       math.Atanh,
	"cbrt":        math.Cbrt,
	"ceil":        math.Ceil,
	"cos":         math.Cos,
	"cosh":        math.Cosh,
	"dim":         math.Dim,
	"floor":       math.Floor,
	"log":         math.Log,
	"log10":       math.Log10,
	"log1p":       math.Log1p,
	"log2":        math.Log2,
	"logb":        math.Logb,
	"max":         math.Max,
	"min":         math.Min,
	"mod":         math.Mod,
	"pow":         math.Pow,
	"remainder":   math.Remainder,
	"round":       math.Round,
	"roundToEven": math.RoundToEven,
	"sin":         math.Sin,
	"sinh":        math.Sinh,
	"sqrt":        math.Sqrt,
	"tan":         math.Tan,
	"tanh":        math.Tanh,
	"trunc":       math.Trunc,
}

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
