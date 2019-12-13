package sec

import (
	"errors"
	"testing"
)

func TestEvalUndeclaredVar(t *testing.T) {
	const text = "balah"

	var env Env
	variable := variable(token{txt: text})

	var uerr ErrUndeclaredVar
	if _, err := variable.Val(env); errors.As(err, &uerr) {
		if uerr.Name != text {
			t.Fatal("variable name not correct")
		}
	} else {
		t.Fatal("expect ErrUndeclaredVar error")
	}
}

func TestEvalVariable(t *testing.T) {
	env := Env{
		Vars: Vars{
			"a": 1,
			"b": 2,
			"c": 3,
			"d": 4,
			"e": 5,
		},
	}
	for txt, val := range env.Vars {
		variable := variable(token{txt: txt})
		v, err := variable.Val(env)
		if err != nil {
			t.Fatal("expect no error")
		} else if v != val {
			t.Fatalf("expect %f, got %f", val, v)
		}
	}
}

func TestEvalLiteral(t *testing.T) {
	pairsGroup := map[tokenType][]struct {
		txt string
		val float64
	}{
		integer:    {{"0", 0}, {"114514", 114514}},
		float:      {{"1919.810", 1919.81}, {"0.00001", 0.00001}},
		binLiteral: {{"0b1", 1}, {"0B1111", 15}},
		octLiteral: {{"0755", 0755}, {"0o10", 8}},
		hexLiteral: {{"0xFF", 0xff}},
	}

	for typ, pairs := range pairsGroup {
		for _, pair := range pairs {
			literal := literal(token{
				typ: typ,
				txt: pair.txt,
			})
			if val, err := literal.Val(Env{}); err != nil {
				t.Fatal("expect no error")
			} else if val != pair.val {
				t.Fatalf("expect %f, got %f", pair.val, val)
			}
		}
	}
}
