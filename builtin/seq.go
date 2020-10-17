package builtin

import (
	"fmt"
	"strings"

	"github.com/spy16/slurp/core"
)

// Seq represents a sequence of values.
type Seq interface {
	// Count returns the number of items in the sequence.
	Count() (int, error)

	// First returns the first item in the sequence.
	First() (core.Any, error)

	// Next returns the tail of the sequence (i.e, sequence after
	// excluding the head). Returns nil, nil if it has no tail.
	Next() (Seq, error)

	// Conj returns a new sequence with given items conjoined.
	Conj(items ...core.Any) (Seq, error)
}

// NewList returns a new linked-list containing given values.
func NewList(items ...core.Any) Seq {
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

// Cons returns a new seq with `v` added as the first and `seq` as the rest.
// seq can be nil as well.
func Cons(v core.Any, seq Seq) (Seq, error) {
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

// SeqString returns a string representation for the sequence with given prefix
// suffix and separator.
func SeqString(seq Seq, begin, end, sep string) (string, error) {
	var b strings.Builder
	b.WriteString(begin)
	err := ForEach(seq, func(item core.Any) (bool, error) {
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
func ForEach(seq Seq, call func(item core.Any) (bool, error)) (err error) {
	var v core.Any
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

// LinkedList implements an immutable Seq using linked-list data structure.
type LinkedList struct {
	count int
	first core.Any
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
func (ll *LinkedList) Equals(other core.Any) (eq bool, err error) {
	o, ok := other.(*LinkedList)
	if !ok || o.count != ll.count {
		return
	}

	var s Seq = ll
	err = ForEach(o, func(any core.Any) (bool, error) {
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
func (ll *LinkedList) Conj(items ...core.Any) (res Seq, err error) {
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
func (ll *LinkedList) First() (core.Any, error) {
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
