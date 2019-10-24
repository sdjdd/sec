package main

import (
	"fmt"
)

func main() {
	//lex := new(lexer)
	// showTokens(lex, "1+05")
	// showTokens(lex, "0xabc*55/2")
	// showTokens(lex, "(1+2)* 3.142	/ 100%2")
	//showTokens(lex, "pow(2,  10.14)")

	calc := Calc{}
	calc.Env = Env{
		Vars: map[string]float64{
			"yjspi": 114514,
			"c3p":   250,
		},
		Funcs: map[string]interface{}{
			"gtmdc3p": func(b ...float64) (sum float64) {
				for _, n := range b {
					sum += n
				}
				return
			},
			"wyy": func(a, b float64) float64 {
				return a + b
			},
		},
	}

	fmt.Println(calc.Eval("wyy(c3p,gtmdc3p())"))

	calc.Env.Funcs = MathFuncs
	fmt.Println(calc.Eval("round(2.6)"))
}

func showTokens(lex *lexer, script string) {
	err := lex.tokenize(script)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(lex.tokens)
	}
}
