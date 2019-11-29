package sec

import (
	"testing"
)

func TestParse(t *testing.T) {
	var psr parser

	ast, err := psr.parse("%")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ast)
	t.Log(ast.Val(Env{
		Vars: Vars{
			"c3p": 100,
			"A_1": 999,
		},
		Funcs: Funcs{
			"gtmd": func(m float64) float64 {
				return m + 100
			},
		},
	}))
}
