package sec

import (
	"fmt"
)

type (
	tokenErr struct {
		srcInfo
		err error
	}

	errUnexpected rune
	errFuncParam  struct {
		name string
		n    int
	}
	errIsNotFunc              string
	errFuncRetTooManyVals     string
	errFuncRetNoVals          string
	errFuncRetNotFloat64      string
	errFuncVariadicNotFloat64 string
	errLiteralHasNoDigits     int
	errInvalidDigitInLiteral  struct {
		bit int
		ch  rune
	}
)

func (t tokenErr) Unwrap() error { return t.err }

func (t tokenErr) Error() string {
	return fmt.Sprintf("%s: %s", t.srcInfo, t.err)
}

func (e errUnexpected) Error() string {
	return fmt.Sprintf("unexpected %q", rune(e))
}

func (f errFuncParam) Error() string {
	var text string
	switch f.n {
	case 1:
		text = "first"
	case 2:
		text = "second"
	case 3:
		text = "third"
	default:
		text = fmt.Sprintf("%dth", f.n)
	}

	return fmt.Sprintf("the %s parameter of function %q is not float64", text, f.name)
}

func (e errIsNotFunc) Error() string {
	return fmt.Sprintf("%q is not a function", string(e))
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
