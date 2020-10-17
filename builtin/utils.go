package builtin

import (
	"reflect"

	"github.com/spy16/slurp/core"
)

// SExpressable forms can be rendered as s-expressions.
type SExpressable interface {
	SExpr() (string, error)
}

// Comparable values define a partial ordering.
type Comparable interface {
	// Comp(pare) the value to another value.  Returns:
	//
	// * 0 if v == other
	// * 1 if v > other
	// * -1 if v < other
	//
	// If the values are not comparable, ErrIncomparableTypes is returned.
	Comp(other core.Any) (int, error)
}

// EqualityProvider asserts equality between two values.
type EqualityProvider interface {
	Equals(other core.Any) (bool, error)
}

// Compare the value to another value. Returns:
//
//  0 if a == b (Same as b == a).
//  1 if a > b.
//  -1 if a < b.
//
// If the values are not comparable, ErrIncomparable is returned.
func Compare(a, b core.Any) (int, error) {
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
func Eq(a, b core.Any) (bool, error) {
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
