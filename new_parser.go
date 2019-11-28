package sec

type parser2 struct {
	tokenReader *tokenReader

	// current token
	token token
}

type binary2 struct {
	op          int
	left, right expr
}
