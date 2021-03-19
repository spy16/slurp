package builtin

import (
	"fmt"
	"math"
	"reflect"
	"strconv"

	"github.com/spy16/slurp/core"
)

var (
	_ core.Any = Nil{}
	_ core.Any = Int64(0)
	_ core.Any = Float64(1.123123)
	_ core.Any = Bool(true)
	_ core.Any = Char('∂')
	_ core.Any = String("specimen")
	_ core.Any = Symbol("specimen")
	_ core.Any = Keyword("specimen")

	_ core.Comparable = Int64(0)
	_ core.Comparable = Float64(0)

	_ core.EqualityProvider = Nil{}
	_ core.EqualityProvider = Bool(false)
	_ core.EqualityProvider = Char('a')
	_ core.EqualityProvider = String("specimen")
	_ core.EqualityProvider = Symbol("specimen")
	_ core.EqualityProvider = Keyword("specimen")
)

// Nil represents the Value 'nil'.
type Nil struct{}

// Equals returns true IFF other is nil.
func (Nil) Equals(other core.Any) (bool, error) { return IsNil(other), nil }

func (Nil) String() string { return "nil" }

// Int64 represents a 64-bit integer Value.
type Int64 int64

// Comp performs comparison against another Int64.
func (i64 Int64) Comp(other core.Any) (int, error) {
	if n, ok := other.(Int64); ok {
		switch {
		case i64 > n:
			return 1, nil
		case i64 < n:
			return -1, nil
		default:
			return 0, nil
		}
	}

	return 0, core.ErrIncomparable
}

func (i64 Int64) String() string { return strconv.Itoa(int(i64)) }

func (i64 Int64) Format(s fmt.State, verb rune) {
	switch verb {
	case 'd':
		fmt.Fprint(s, i64.String())

	default:
		if s.Flag('#') {
			fmt.Fprintf(s, "%#v", int64(i64))
			return
		}

		fmt.Fprintf(s, "%v", int64(i64))
	}
}

// Float64 represents a 64-bit double precision floating point Value.
type Float64 float64

// Comp performs comparison against another Float64.
func (f64 Float64) Comp(other core.Any) (int, error) {
	if n, ok := other.(Float64); ok {
		switch {
		case f64 > n:
			return 1, nil
		case f64 < n:
			return -1, nil
		default:
			return 0, nil
		}
	}

	return 0, core.ErrIncomparable
}

func (f64 Float64) String() string {
	if f64 == 0 {
		return "0."
	}

	if math.Abs(float64(f64)) >= 1e16 {
		return fmt.Sprintf("%e", float64(f64))
	}

	return fmt.Sprintf("%f", float64(f64))
}

func (f64 Float64) Format(s fmt.State, verb rune) {
	switch verb {
	case 'f':
		fmt.Fprint(s, f64.String())

	default:
		if s.Flag('#') {
			fmt.Fprintf(s, "%#v", float64(f64))
			return
		}

		fmt.Fprintf(s, "%v", float64(f64))
	}
}

// Bool represents a boolean Value.
type Bool bool

// Equals returns true if 'other' is a boolean and has same logical Value.
func (b Bool) Equals(other core.Any) (bool, error) {
	val, ok := other.(Bool)
	return ok && (val == b), nil
}

func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

// Char represents a Unicode character.
type Char rune

// Equals returns true if the other Value is also a character and has same Value.
func (char Char) Equals(other core.Any) (bool, error) {
	val, isChar := other.(Char)
	return isChar && (val == char), nil
}

func (char Char) String() string { return fmt.Sprintf("\\%c", char) }

// String represents a string of characters.
type String string

// Equals returns true if 'other' is string and has same Value.
func (str String) Equals(other core.Any) (bool, error) {
	otherStr, isStr := other.(String)
	return isStr && (otherStr == str), nil
}

func (str String) String() string { return fmt.Sprintf("\"%s\"", string(str)) }

// Symbol represents a lisp symbol Value.
type Symbol string

// Equals returns true if the other Value is also a symbol and has same Value.
func (sym Symbol) Equals(other core.Any) (bool, error) {
	otherSym, isSym := other.(Symbol)
	return isSym && (sym == otherSym), nil
}

func (sym Symbol) String() string { return string(sym) }

// Keyword represents a keyword Value.
type Keyword string

// Equals returns true if the other Value is keyword and has same Value.
func (kw Keyword) Equals(other core.Any) (bool, error) {
	otherKW, isKeyword := other.(Keyword)
	return isKeyword && (otherKW == kw), nil
}

func (kw Keyword) String() string { return fmt.Sprintf(":%s", string(kw)) }

// IsNil returns true if value is native go `nil` or `Nil{}`.
func IsNil(v core.Any) bool {
	if v == nil {
		return true
	}
	_, isNilType := v.(Nil)
	return isNilType
}

// IsTruthy returns true if the value has a logical vale of `true`.
func IsTruthy(v core.Any) bool {
	if IsNil(v) {
		return false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Bool {
		return rv.Bool()
	}

	return true
}
