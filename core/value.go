package core

import "errors"

var (
	// ErrNotInvokable is returned by InvokeExpr when the target is not
	// invokable.
	ErrNotInvokable = errors.New("not invokable")

	// ErrArity is returned when an Invokable is invoked with wrong number
	// of arguments.
	ErrArity = errors.New("wrong number of arguments")

	// ErrIncomparable is returned by Compare() when a comparison between
	// two types is undefined. Users should  consider the types to  be not
	// equal  in such cases, but not  assume any ordering.
	ErrIncomparable = errors.New("incomparable types")
)

// Any represents any Go/slurp value.
type Any interface{}

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
	if a == nil && b == nil {
		return true, nil
	}
	aSeq, aOk := a.(Seq)
	bSeq, bOk := b.(Seq)
	if aOk && bOk {
		return seqEq(aSeq, bSeq)
	}

	c, err := Compare(a, b)
	if err != nil {
		if err == ErrIncomparable {
			return false, nil
		}
		return false, err
	}
	return c == 0, nil
}
