package slurp

import (
	"fmt"
	"math"
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
	_ Any = (*LinkedList)(nil)

	_ Comparable = Int64(0)
	_ Comparable = Float64(0)

	_ Seq = (*LinkedList)(nil)
)

// Any represents any slurp value.
type Any interface{}

// Invokable represents a value that can be invoked for result.
type Invokable interface {
	Any

	// Invoke is called if this value appears as the first argument of
	// invocation form (i.e., list).
	Invoke(env *Env, args ...Any) (Any, error)
}

// Seq represents a sequence of values.
type Seq interface {
	Any

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

// Cons returns a new seq with `v` added as the first and `seq` as the rest.
// seq can be nil as well.
func Cons(v Any, seq Seq) (Seq, error) {
	newSeq := &LinkedList{
		first: v,
		rest:  seq,
		count: 1,
	}

	if seq != nil {
		cnt, err := seq.Count()
		if err != nil {
			return nil, err
		}
		newSeq.count = cnt + 1
	}

	return newSeq, nil
}

// NewList returns a new linked-list containing given values.
func NewList(items ...Any) Seq {
	if len(items) == 0 {
		return Seq((*LinkedList)(nil))
	}

	var err error
	lst := Seq(&LinkedList{})
	for i := len(items) - 1; i >= 0; i-- {
		if lst, err = Cons(items[i], lst); err != nil {
			panic(err)
		}
	}

	return lst
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

// LinkedList implements an immutable Seq using linked-list data structure.
type LinkedList struct {
	count int
	first Any
	rest  Seq
}

// SExpr returns a valid s-expression for LinkedList.
func (ll *LinkedList) SExpr() (string, error) {
	if ll == nil {
		return "()", nil
	}

	return SeqString(ll, "(", ")", " ")
}

// Equals returns true if the other value is a LinkedList and contains the same
// values.
func (ll *LinkedList) Equals(other Any) (eq bool, err error) {
	o, ok := other.(*LinkedList)
	if !ok || o.count != ll.count {
		return
	}

	var s Seq = ll
	err = ForEach(o, func(any Any) (bool, error) {
		v, _ := s.First()

		veq, ok := v.(EqualityProvider)
		if !ok {
			return false, nil
		}

		if eq, err = veq.Equals(any); err != nil || !eq {
			return true, err
		}

		s, _ = s.Next()
		return false, nil
	})

	return
}

// Conj returns a new list with all the items added at the head of the list.
func (ll *LinkedList) Conj(items ...Any) (res Seq, err error) {
	if ll == nil {
		res = &LinkedList{}
	} else {
		res = ll
	}

	for _, item := range items {
		if res, err = Cons(item, res); err != nil {
			break
		}
	}

	return
}

// First returns the head or first item of the list.
func (ll *LinkedList) First() (Any, error) {
	if ll == nil {
		return nil, nil
	}
	return ll.first, nil
}

// Next returns the tail of the list.
func (ll *LinkedList) Next() (Seq, error) {
	if ll == nil {
		return nil, nil
	}
	return ll.rest, nil
}

// Count returns the number of the list.
func (ll *LinkedList) Count() (int, error) {
	if ll == nil {
		return 0, nil
	}

	return ll.count, nil
}
