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
		"yjspi": 114514,
		"c3p":   250,
	}
	fmt.Println(calc.Eval(`yjspi%100*`))
}

func showTokens(lex *lexer, script string) {
	err := lex.tokenize(script)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(lex.tokens)
	}
}
