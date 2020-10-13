package slurp

import (
	"fmt"
	"reflect"
	"strings"
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
	Comp(other Any) (int, error)
}

// EqualityProvider asserts equality between two values.
type EqualityProvider interface {
	Equals(other Any) (bool, error)
}

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

// EvalAll evaluates each value in the list against the given env and returns
// a list of resultant values.
func EvalAll(env *Env, vals []Any) ([]Any, error) {
	res := make([]Any, 0, len(vals))
	for _, form := range vals {
		form, err := env.Eval(form)
		if err != nil {
			return nil, err
		}
		res = append(res, form)
	}
	return res, nil
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

// SeqString returns a string representation for the sequence with given prefix
// suffix and separator.
func SeqString(seq Seq, begin, end, sep string) (string, error) {
	var b strings.Builder
	b.WriteString(begin)
	err := ForEach(seq, func(item Any) (bool, error) {
		if sxpr, ok := item.(SExpressable); ok {
			s, err := sxpr.SExpr()
			if err != nil {
				return false, err
			}
			b.WriteString(s)

		} else {
			b.WriteString(fmt.Sprintf("%v", item))
		}

		b.WriteString(sep)
		return false, nil
	})

	if err != nil {
		return "", err
	}

	return strings.TrimRight(b.String(), sep) + end, err
}

// ForEach reads from the sequence and calls the given function for each item.
// Function can return true to stop the iteration.
func ForEach(seq Seq, call func(item Any) (bool, error)) (err error) {
	var v Any
	var done bool
	for seq != nil {
		if v, err = seq.First(); err != nil || v == nil {
			break
		}

		if done, err = call(v); err != nil || done {
			break
		}

		if seq, err = seq.Next(); err != nil {
			break
		}
	}

	return
}
