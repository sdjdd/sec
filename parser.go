package sec

import (
	"io"
)

type parserErr error

type Parser struct {
	tokenReader tokenReader
	token       token // current token
}

func (p *Parser) next() {
	var err error
	p.token, err = p.tokenReader.read()
	if err != nil && err != io.EOF {
		panic(parserErr(err))
	}
}

func (p *Parser) Parse(s string) (ast Expr, err error) {
	defer func() {
		switch er := recover().(type) {
		case nil:
		case parserErr:
			err = er
		default:
			panic(er)
		}
	}()

	p.tokenReader.load(s)
	p.next()
	ast = p.parseAddition()

	if p.token.typ != initial {
		err = p.token.errorf("Unexpected %q", p.token.txt)
	}

	return
}

// Addition  = Multiplicative ('+' Multiplicative)*
func (p *Parser) parseAddition() Expr {
	left := p.parseMultiplication()
	for {
		switch p.token.typ {
		case plus, minus:
			op := p.token
			p.next() // consume operator
			right := p.parseMultiplication()
			left = binary{op, left, right}
		default:
			return left
		}
	}
}

// Multiplication = Exponentiation ('*' Exponentiation)*
func (p *Parser) parseMultiplication() Expr {
	left := p.parseExponentiation()
	for {
		switch p.token.typ {
		case star, slash, percent, doubleSlash:
			op := p.token
			p.next() // consume operator
			right := p.parseExponentiation()
			left = binary{op, left, right}
		default:
			return left
		}
	}
}

// Exponentiation = Unary ('**' Unary)*
func (p *Parser) parseExponentiation() Expr {
	left := p.parseUnary()

	for {
		if p.token.typ != doubleStar {
			break
		}
		op := p.token
		p.next() // consume operator
		right := p.parseUnary()
		left = binary{op, left, right}
	}

	return left
}

// Unary = '+' Unary
//       | Primary
func (p *Parser) parseUnary() Expr {
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
func (p *Parser) parsePrimary() Expr {
	switch p.token.typ {
	case initial:
		panic(parserErr(p.token.errorf("unexpected EOF")))
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
				args = append(args, p.parseAddition())
				if p.token.typ != comma {
					break
				}
				p.next() // consume ','
			}
			if p.token.typ != rBracket {
				var text string
				if p.token.typ == initial {
					text = "EOF"
				} else {
					text = "'" + string(p.token.txt[0]) + "'"
				}
				panic(parserErr(p.token.errorf("want ')', got %s", text)))
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
		e := p.parseAddition()
		if p.token.typ != rBracket {
			var text string
			if p.token.typ == initial {
				text = "EOF"
			} else {
				text = "'" + string(p.token.txt[0]) + "'"
			}
			p.next()
			panic(parserErr(p.token.errorf("want ')', got %s", text)))
		}
		p.next() // consume ')'
		return e
	default:
		panic(parserErr(p.token.errorf("unexpected %q", p.token.txt[0])))
	}
}
