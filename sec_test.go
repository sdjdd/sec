package sec

import (
	"testing"
)

func TestEvalFunction(t *testing.T) {
	calc := Calc{}
	calc.Env.Funcs = MathFuncs
	val, err := calc.Eval(`(111)`)
	if err != nil {
		t.Fatal(err)
	} else if val != 111 {
		t.Fatalf("got %f, want %d", val, 1024)
	}
}

func TestEval(t *testing.T) {
	calc := New()
	calc.Env.Funcs = MathFuncs
	calc.Env.Funcs["gtmd"] = func(name float64) float64 { return name }
	calc.BeforeEval = func(env Env, varNames []string) {
		t.Log("varNames:", varNames)
		for _, name := range varNames {
			env.Vars[name] = 1919810
		}
	}

	scripts := []string{
		"18111/2*pow(5,4)-90555*pow(5,3)+633885/2*pow(5,2)-472973*5+215504",
		"gtmd(c3p) + c3p",
		"1+1",
		"0b101010",
	}

	for _, s := range scripts {
		val, err := calc.Eval(s)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("result:", val)
	}
}

func Test_FunctionReturnVal(t *testing.T) {
	calc := New()
	funcs := []interface{}{
		func() {},
		func() int { return 0 },
		func() (float64, float64) { return 0, 0 },
	}

	for i, f := range funcs {
		calc.Env.Funcs["func"] = f
		err := calc.CheckFuncs()
		t.Log(err)
		if err == nil {
			t.Fatalf("test case %d: err should not nil", i)
		}
	}
}
