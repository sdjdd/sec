package sec

import "testing"

func Test_Tokenize(t *testing.T) {
	tk := func(typ int, txt string) token {
		return token{
			typ: typ,
			txt: txt,
		}
	}
	tests := []struct {
		s      string
		tokens []token
	}{
		{
			"1+22*3/44.4 % (50 + f(var1, var2, 6))",
			[]token{
				tk(integer, "1"),
				tk(operator, "+"),
				tk(integer, "22"),
				tk(operator, "*"),
				tk(integer, "3"),
				tk(operator, "/"),
				tk(float, "44.4"),
				tk(operator, "%"),
				tk(lBracket, "("),
				tk(integer, "50"),
				tk(operator, "+"),
				tk(identifier, "f"),
				tk(lBracket, "("),
				tk(identifier, "var1"),
				tk(comma, ","),
				tk(identifier, "var2"),
				tk(comma, ","),
				tk(integer, "6"),
				tk(rBracket, ")"),
				tk(rBracket, ")"),
			},
		},
	}
	lex := lexer{}
	for _, test := range tests {
		lex.tokenize(test.s)
		// t.Log(lex.tokens)
		// t.Log(test.tokens)
		if len(lex.tokens) != len(test.tokens) {
			t.Fatal("token number is not correct")
		}
		for i, token := range test.tokens {
			if token.typ != lex.tokens[i].typ || token.txt != lex.tokens[i].txt {
				t.Fatal("token  is not correct")
			}
		}
	}
}
