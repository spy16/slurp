package core

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

var (
	_ Any = Nil{}
	_ Any = Int64(0)
	_ Any = Float64(1.123123)
	_ Any = Bool(true)
	_ Any = Char('∂')
	_ Any = String("specimen")
	_ Any = Symbol("specimen")
	_ Any = Keyword("specimen")

	_ Comparable = Int64(0)
	_ Comparable = Float64(0)

	_ EqualityProvider = Nil{}
	_ EqualityProvider = Bool(false)
	_ EqualityProvider = Char('a')
	_ EqualityProvider = String("specimen")
	_ EqualityProvider = Symbol("specimen")
	_ EqualityProvider = Keyword("specimen")
)

var (
	// ErrIncomparable is returned by Compare() when a comparison
	// between two types is undefined. Users should  consider the
	// types to  be not equal  in such cases, but not  assume any
	// ordering.
	ErrIncomparable = errors.New("incomparable types")
)

// Any represents any slurp value.
type Any interface{}

// Seq represents a sequence of values.
type Seq interface {
	// Count returns the number of items in the sequence.
	Count() (int, error)

	// First returns the first item in the sequence.
	First() (Any, error)

	// Next returns the tail of the sequence (i.e, sequence after
	// excluding the head). Returns nil, nil if it has no tail.
	Next() (Seq, error)

	// Conj returns a new sequence with given items conjoined.
	Conj(items ...Any) (Seq, error)
}

// Invokable represents a value that can be invoked for result.
type Invokable interface {
	// Invoke is called if this value appears as the first argument of
	// invocation form (i.e., list).
	Invoke(args ...Any) (Any, error)
}

// SExpressable forms can be rendered as s-expressions.
type SExpressable interface {
	// SExpr returns a parsable s-expression of the given value. Returns
	// error if not possible.
	SExpr() (string, error)
}

// Comparable values define a partial ordering.
type Comparable interface {
	// Comp(pare) the value to another value.  Returns:
	//
	// -1 if v < other
	//  0 if v == other
	//  1 if v > other
	//
	// If the values are not comparable, ErrIncomparable is returned.
	Comp(other Any) (int, error)
}

// EqualityProvider asserts equality between two values.
type EqualityProvider interface {
	Equals(other Any) (bool, error)
}

// Nil represents the Value 'nil'.
type Nil struct{}

// SExpr returns a valid s-expression representing Nil.
func (Nil) SExpr() (string, error) { return "nil", nil }

// Equals returns true IFF other is nil.
func (Nil) Equals(other Any) (bool, error) { return IsNil(other), nil }

func (Nil) String() string { return "nil" }

// Int64 represents a 64-bit integer Value.
type Int64 int64

// SExpr returns a valid s-expression representing Int64.
func (i64 Int64) SExpr() (string, error) { return i64.String(), nil }

// Comp performs comparison against another Int64.
func (i64 Int64) Comp(other Any) (int, error) {
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

	return 0, ErrIncomparable
}

func (i64 Int64) String() string { return strconv.Itoa(int(i64)) }

// Float64 represents a 64-bit double precision floating point Value.
type Float64 float64

// SExpr returns a valid s-expression representing Float64.
func (f64 Float64) SExpr() (string, error) { return f64.String(), nil }

// Comp performs comparison against another Float64.
func (f64 Float64) Comp(other Any) (int, error) {
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

	return 0, ErrIncomparable
}

func (f64 Float64) String() string {
	if math.Abs(float64(f64)) >= 1e16 {
		return fmt.Sprintf("%e", f64)
	}
	return fmt.Sprintf("%f", f64)
}

// Bool represents a boolean Value.
type Bool bool

// SExpr returns a valid s-expression representing Bool.
func (b Bool) SExpr() (string, error) { return b.String(), nil }

// Equals returns true if 'other' is a boolean and has same logical Value.
func (b Bool) Equals(other Any) (bool, error) {
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

// SExpr returns a valid s-expression representing Char.
func (char Char) SExpr() (string, error) {
	return fmt.Sprintf("\\%c", char), nil
}

// Equals returns true if the other Value is also a character and has same Value.
func (char Char) Equals(other Any) (bool, error) {
	val, isChar := other.(Char)
	return isChar && (val == char), nil
}

func (char Char) String() string { return fmt.Sprintf("\\%c", char) }

// String represents a string of characters.
type String string

// SExpr returns a valid s-expression representing String.
func (str String) SExpr() (string, error) { return str.String(), nil }

// Equals returns true if 'other' is string and has same Value.
func (str String) Equals(other Any) (bool, error) {
	otherStr, isStr := other.(String)
	return isStr && (otherStr == str), nil
}

func (str String) String() string { return fmt.Sprintf("\"%s\"", string(str)) }

// Symbol represents a lisp symbol Value.
type Symbol string

// SExpr returns a valid s-expression representing Symbol.
func (sym Symbol) SExpr() (string, error) { return string(sym), nil }

// Equals returns true if the other Value is also a symbol and has same Value.
func (sym Symbol) Equals(other Any) (bool, error) {
	otherSym, isSym := other.(Symbol)
	return isSym && (sym == otherSym), nil
}

func (sym Symbol) String() string { return string(sym) }

// Keyword represents a keyword Value.
type Keyword string

// SExpr returns a valid s-expression representing Keyword.
func (kw Keyword) SExpr() (string, error) { return kw.String(), nil }

// Equals returns true if the other Value is keyword and has same Value.
func (kw Keyword) Equals(other Any) (bool, error) {
	otherKW, isKeyword := other.(Keyword)
	return isKeyword && (otherKW == kw), nil
}

func (kw Keyword) String() string { return fmt.Sprintf(":%s", string(kw)) }

// Compare the value to another value. Returns:
//
//  0 if a == b (Same as b == a).
//  1 if a > b.
//  -1 if a < b.
//
// If the values are not comparable, ErrIncomparable is returned.
func Compare(a, b Any) (int, error) {
	if cmp, ok := a.(Comparable); ok {
		return cmp.Comp(b)
	}

	if ep, ok := a.(EqualityProvider); ok {
		eq, err := ep.Equals(b)
		if eq || err != nil {
			return 0, err
		}
	} else if ep, ok := b.(EqualityProvider); ok {
		eq, err := ep.Equals(a)
		if eq || err != nil {
			return 0, err
		}
	}

	return 0, ErrIncomparable
}

// Eq returns true if a == b.
func Eq(a, b Any) (bool, error) {
	c, err := Compare(a, b)
	if err != nil {
		if err == ErrIncomparable {
			return false, nil
		}
		return false, err
	}
	return c == 0, nil
}

// IsNil returns true if value is native go `nil` or `Nil{}`.
func IsNil(v Any) bool {
	if v == nil {
		return true
	}
	_, isNilType := v.(Nil)
	return isNilType
}

// IsTruthy returns true if the value has a logical vale of `true`.
func IsTruthy(v Any) bool {
	if IsNil(v) {
		return false
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Bool {
		return rv.Bool()
	}

	return true
}
