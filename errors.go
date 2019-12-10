package sec

import (
	"errors"
	"fmt"
)

type (
	tokenErr struct {
		SourceInfo
		err error
	}

	// get a unexpected Char
	ErrUnexpected struct {
		SourceInfo
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

	errFuncRetTooManyVals     string
	errFuncRetNoVals          string
	errFuncRetNotFloat64      string
	errFuncVariadicNotFloat64 string
	errLiteralHasNoDigits     int
	errInvalidDigitInLiteral  struct {
		bit int
		ch  rune
	}
	errUndeclaredVar string
	errUndeclaredFun string
	errTooFewArgs    string
	errTooManyArgs   string
)

var errUnexpectedEOF = errors.New("unexpected EOF")

func (t tokenErr) Unwrap() error { return t.err }

func (t tokenErr) Error() string {
	return fmt.Sprintf("%s: %s", t.SourceInfo, t.err)
}

func (e ErrUnexpected) Error() string {
	return fmt.Sprintf("%s: unexpected %q", e.SourceInfo, e.Char)
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

func (e errFuncRetTooManyVals) Error() string {
	return fmt.Sprintf("function %q must return only one value", string(e))
}

func (e errFuncRetNoVals) Error() string {
	return fmt.Sprintf("function %q must return a value", string(e))
}

func (e errFuncRetNotFloat64) Error() string {
	return fmt.Sprintf("function %q must return a float64 value", string(e))
}

func (e errFuncVariadicNotFloat64) Error() string {
	return fmt.Sprintf("variadic parameter of %q must be float64", string(e))
}

func bit2str(bit int) (str string) {
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

func (e errLiteralHasNoDigits) Error() string {
	return bit2str(int(e)) + " literal has no digits"
}

func (e errInvalidDigitInLiteral) Error() string {
	return fmt.Sprintf("invalid digit %q in %s literal", e.ch, bit2str(e.bit))
}

func (e errUndeclaredVar) Error() string {
	return fmt.Sprintf("undeclared variable %q", string(e))
}

func (e errUndeclaredFun) Error() string {
	return fmt.Sprintf("undeclared function %q", string(e))
}

func (e errTooFewArgs) Error() string {
	return fmt.Sprintf("too few arguments to call %q", string(e))
}

func (e errTooManyArgs) Error() string {
	return fmt.Sprintf("too many arguments to call %q", string(e))
}
