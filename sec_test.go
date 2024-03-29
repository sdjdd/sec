package sec

import (
	"testing"
)

func TestSec(t *testing.T) {
	t.Log(Eval("114514-1919810"))
}

func TestFuncCheck(t *testing.T) {
	var env Env
	env.Funcs = make(Funcs)

	env.Funcs["f"] = "not a function"
	if _, ok := env.Funcs.Check().(ErrNotFunction); !ok {
		t.Fatal("expect errIsNotFunc error")
	}

	env.Funcs["f"] = func() {}
	if _, ok := env.Funcs.Check().(ErrFuncNoReturnVal); !ok {
		t.Fatal("expect errFuncRetNoVals error")
	}

	env.Funcs["f"] = func() (float64, float64) { return 0, 0 }
	if _, ok := env.Funcs.Check().(ErrFuncReturnTooManyVal); !ok {
		t.Fatal("expect errFuncRetTooManyVals error")
	}

	env.Funcs["f"] = func() int { return 0 }
	if _, ok := env.Funcs.Check().(ErrReturnValNotFloat64); !ok {
		t.Fatal("expect errFuncRetNotFloat64 error")
	}

	env.Funcs["f"] = func(p1 float64, p2 ...int) float64 { return 0 }
	if err, ok := env.Funcs.Check().(ErrParamNotFloat64); !ok {
		t.Fatal("expect ErrParamNotFloat64 error")
	} else if err.N != 2 {
		t.Fatal("param number not correct")
	}

	env.Funcs["f"] = func(p1 float64, p2 int) float64 { return 0 }
	if err, ok := env.Funcs.Check().(ErrParamNotFloat64); !ok {
		t.Fatal("expect errFuncParam error")
	} else if err.N != 2 {
		t.Fatal("param number not correct")
	}

	env.Funcs["f"] = func() float64 { return 0 }
	if err := env.Funcs.Check(); err != nil {
		t.Fatal("expect no error")
	}
}
