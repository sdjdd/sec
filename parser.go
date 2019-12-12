package sec

import (
	"io"
)

type Parser struct {
	tokenReader tokenReader
	token       token // current token
}

func (p *Parser) next() {
	var err error
	p.token, err = p.tokenReader.read()
	if err != nil && err != io.EOF {
		panic(err)
	}
}

func (p *Parser) Parse(s string) (ast Expr, err error) {
	defer func() {
		switch er := recover().(type) {
		case nil:
		case secError:
			err = er
		default:
			panic(er)
		}
	}()

	p.tokenReader.load(s)
	p.next()
	ast = p.parseAddition()

	if p.token.typ != initial {
		err = ErrUnexpected{[]rune(p.token.txt)[0]}
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
		panic(p.token.wrapErr(errUnexpectedEOF))
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
				err := ErrUnexpected{[]rune(p.token.txt)[0]}
				panic(secError{p.token.SourceInfo, err})
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
			panic(p.token.wrapErr(ErrUnexpected{[]rune(p.token.txt)[0]}))
		}
		p.next() // consume ')'
		return e
	default:
		err := p.token.wrapErr(ErrUnexpected{[]rune(p.token.txt)[0]})
		panic(err)
	}
}
