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

	calc := New()
	calc.Env = Env{
		"c3p": 114514,
	}
	fmt.Println(calc.Eval("0-c3p"))
}

func showTokens(lex *lexer, script string) {
	err := lex.tokenize(script)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(lex.tokens)
	}
}
