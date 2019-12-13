package sec

import (
	"errors"
	"fmt"
)

type (
	secError struct {
		pos
		err error
	}

	// get a unexpected Char
	ErrUnexpected struct {
		Char rune
	}

	// function's Nth parameter is not float64
	ErrParamNotFloat64 struct {
		Name string // functhon name
		N    int    // Nth parameter
	}

	// Name's value in Env.Funcs in not a legal function
	ErrNotFunction struct {
		Name string
	}

	// function return too many values
	ErrFuncReturnTooManyVal struct {
		Name string // function name
	}

	// function return no value
	ErrFuncNoReturnVal struct {
		Name string // function name
	}

	ErrReturnValNotFloat64 struct {
		Name string
	}

	ErrLiteralNoDigit struct {
		Base int
	}

	ErrInvalidDigitInLiteral struct {
		Base  int
		Digit rune
	}

	ErrUndeclaredVar struct {
		Name string
	}

	ErrUndeclaredFunc struct {
		Name string
	}

	ErrTooFewArgsToCall struct {
		Name string
	}

	ErrTooManyArgsToCall struct {
		Name string
	}
)

var errUnexpectedEOF = errors.New("unexpected EOF")

func (t secError) Unwrap() error { return t.err }

func (t secError) Error() string {
	return fmt.Sprintf("%s: %s", t.pos, t.err)
}

func (e ErrUnexpected) Error() string {
	return fmt.Sprintf("unexpected %q", e.Char)
}

func (f ErrParamNotFloat64) Error() string {
	var text string
	switch f.N {
	case 1:
		text = "first"
	case 2:
		text = "second"
	case 3:
		text = "third"
	default:
		text = fmt.Sprintf("%dth", f.N)
	}

	return fmt.Sprintf("the %s parameter of function %q is not float64", text, f.Name)
}

func (e ErrNotFunction) Error() string {
	return fmt.Sprintf("%q is not a function", e.Name)
}

func (e ErrFuncReturnTooManyVal) Error() string {
	return fmt.Sprintf("function %q must return only one value", e.Name)
}

func (e ErrFuncNoReturnVal) Error() string {
	return fmt.Sprintf("function %q must return a value", e.Name)
}

func (e ErrReturnValNotFloat64) Error() string {
	return fmt.Sprintf("function %q must return a float64 value", e.Name)
}

func baseToStr(bit int) (str string) {
	switch bit {
	case 2:
		str = "binary"
	case 8:
		str = "octal"
	case 16:
		str = "hexadecimal"
	default:
		str = "unknown"
	}
	return
}

func (e ErrLiteralNoDigit) Error() string {
	return baseToStr(e.Base) + " literal has no digits"
}

func (e ErrInvalidDigitInLiteral) Error() string {
	return fmt.Sprintf("invalid digit %q in %s literal", e.Digit, baseToStr(e.Base))
}

func (e ErrUndeclaredVar) Error() string {
	return fmt.Sprintf("undeclared variable %q", e.Name)
}

func (e ErrUndeclaredFunc) Error() string {
	return fmt.Sprintf("undeclared function %q", e.Name)
}

func (e ErrTooFewArgsToCall) Error() string {
	return fmt.Sprintf("too few arguments to call %q", e.Name)
}

func (e ErrTooManyArgsToCall) Error() string {
	return fmt.Sprintf("too many arguments to call %q", e.Name)
}
