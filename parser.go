package sec

import (
	"io"
)

type parseErr error

type parser struct {
	tokenReader *tokenReader

	// current token
	token token
}

func (p *parser) next() {
	var err error
	p.token, err = p.tokenReader.read()
	if err != nil && err != io.EOF {
		panic(parseErr(err))
	}
}

func (p *parser) parse(src string) (ast Expr, err error) {
	// lazy load
	if p.tokenReader == nil {
		p.tokenReader = new(tokenReader)
	}

	defer func() {
		switch er := recover().(type) {
		case nil:
		case parseErr:
			err = er
		default:
			panic(er)
		}
	}()

	p.tokenReader.load(src)
	p.next()
	ast = p.parseAdditive()

	if p.token.typ != initial {
		err = p.token.errorf("Unexpected %q", p.token.txt)
	}

	return
}

// Additive = Multiplicative ('+' Multiplicative)*
func (p *parser) parseAdditive() Expr {
	left := p.parseMultiplicative()

	for {
		if p.token.typ != plus && p.token.typ != minus {
			break
		}
		op := p.token.txt
		p.next() // consume operator
		right := p.parseMultiplicative()
		left = binary{op, left, right}
	}

	return left
}

// Multiplicative = Unary ('*' Unary)*
func (p *parser) parseMultiplicative() Expr {
	left := p.parseUnary()

	for {
		if p.token.typ != star && p.token.typ != slash {
			break
		}
		op := p.token.txt
		p.next() // consume operator
		right := p.parseUnary()
		left = binary{op, left, right}
	}

	return left
}

// Unary = '+' Unary
//       | Primary
func (p *parser) parseUnary() Expr {
	if p.token.typ == plus || p.token.typ == minus {
		op := p.token
		p.next() // consume operator
		return unary{op, p.parseUnary()}
	}
	return p.parsePrimary()
}

// Primary = identifier
//         | number
//         | identifier '(' Additive ')'
//         | '(' Additive ')'
func (p *parser) parsePrimary() Expr {
	switch p.token.typ {
	case initial:
		panic(parseErr(p.token.errorf("unexpected EOF")))
	case identifier:
		id := p.token
		p.next() // consume identifier
		if p.token.typ != lBracket {
			return variable(id)
		}
		p.next() // consume '('
		var args []Expr
		if p.token.typ != rBracket {
			for {
				args = append(args, p.parseAdditive())
				if p.token.typ != comma {
					break
				}
				p.next() // consume ','
			}
			if p.token.typ != rBracket {
				panic(parseErr(p.token.errorf("want ')', got %q", p.token.txt[0])))
			}
		}
		p.next() // consume ')'
		return call{id, args}
	case integer, float, binLiteral, octLiteral, hexLiteral:
		token := p.token
		p.next()
		return literal(token)
	case lBracket:
		p.next() // consume '('
		e := p.parseAdditive()
		if p.token.typ != rBracket {
			panic(parseErr(p.token.errorf("want ')', got %q", p.token.txt)))
		}
		p.next() // consume ')'
		return e
	default:
		panic(parseErr(p.token.errorf("unexpected %q", p.token.txt)))
	}
}
