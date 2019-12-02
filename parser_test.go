package sec

import (
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	var psr parser
	script := `
pow(
	c3p,
	2,
)
`
	ast, err := psr.parse(script)
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
