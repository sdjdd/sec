package sec

import (
	"testing"
)

func TestSec(t *testing.T) {
	t.Log(Eval("3**2"))
}

func TestFuncCheck(t *testing.T) {
	var env Env
	env.Funcs = make(Funcs)

	env.Funcs["f"] = "not a function"
	if _, ok := env.Funcs.Check().(errIsNotFunc); !ok {
		t.Fatal("expect errIsNotFunc error")
	}

	env.Funcs["f"] = func() {}
	if _, ok := env.Funcs.Check().(errFuncRetNoVals); !ok {
		t.Fatal("expect errFuncRetNoVals error")
	}

	env.Funcs["f"] = func() (float64, float64) { return 0, 0 }
	if _, ok := env.Funcs.Check().(errFuncRetTooManyVals); !ok {
		t.Fatal("expect errFuncRetTooManyVals error")
	}

	env.Funcs["f"] = func() int { return 0 }
	if _, ok := env.Funcs.Check().(errFuncRetNotFloat64); !ok {
		t.Fatal("expect errFuncRetNotFloat64 error")
	}

	env.Funcs["f"] = func(p1 float64, p2 ...int) float64 { return 0 }
	if _, ok := env.Funcs.Check().(errFuncVariadicNotFloat64); !ok {
		t.Fatal("expect errFuncVariadicNotFloat64 error")
	}

	env.Funcs["f"] = func(p1 float64, p2 int) float64 { return 0 }
	if err, ok := env.Funcs.Check().(errFuncParam); !ok {
		t.Fatal("expect errFuncParam error")
	} else if err.n != 2 {
		t.Fatal("param number not correct")
	}

	env.Funcs["f"] = func() float64 { return 0 }
	if err := env.Funcs.Check(); err != nil {
		t.Fatal("expect no error")
	}
}
