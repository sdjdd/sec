package sec

import (
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	var psr Parser
	script := "09"
	ast, err := psr.Parse(script)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ast.Val(Env{
		Vars: Vars{
			"c3p": 100,
			"A_1": 999,
		},
		Funcs: Funcs{
			"pow": math.Pow,
		},
	}))
}
