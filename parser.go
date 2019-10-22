package main

type parser struct {
	lex lexer
}

type expr interface {
	val() float64
}

type binary struct {
	operator string
	l, r     expr
}

func (b binary) val() float64 {
	switch b.operator {
	case "+":
		return b.l.val() + b.r.val()
	case "-":
		return b.l.val() - b.r.val()
	default:
		panic("unknown operator " + b.operator)
	}
}

func (p *parser) parse(script string) (float64, error) {
	err := p.lex.tokenize(script)
	if err != nil {
		return 0, err
	}

	if !p.lex.next() {
		panic("empty input")
	}

	expr := p.parseBinary()
	return expr.val(), nil
}

func (p *parser) parseBinary() expr {
	var left expr
	left = p.lex.peek()

	for p.lex.next() {
		op := p.lex.peek()
		if !p.lex.next() {
			panic("expect operator")
		}
		right := p.parseBinary()
		left = binary{
			operator: op.txt,
			l:        left,
			r:        right,
		}
	}
	return left
}
