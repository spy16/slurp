package core

import (
	"fmt"
	"strings"
)

var (
	_ Any = (*LinkedList)(nil)
	_ Seq = (*LinkedList)(nil)
)

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

func seqEq(s1, s2 Seq) (bool, error) {
	if sEq, ok := s1.(EqualityProvider); ok {
		return sEq.Equals(s2)
	} else if sEq, ok := s2.(EqualityProvider); ok {
		return sEq.Equals(s1)
	}

	if s1 == nil && s2 == nil {
		return true, nil
	} else if (s1 == nil && s2 != nil) ||
		(s1 != nil && s2 == nil) {
		return false, nil
	}

	c1, err := s1.Count()
	if err != nil {
		return false, err
	}

	c2, err := s2.Count()
	if err != nil {
		return false, err
	}

	if c1 != c2 {
		return false, nil
	}

	bothEqual := true
	for i := 0; i < c1; i++ {
		v1, err := s1.First()
		if err != nil {
			return false, err
		}

		v2, err := s2.First()
		if err != nil {
			return false, err
		}

		eq, err := Eq(v1, v2)
		if err != nil {
			return false, err
		}
		bothEqual = bothEqual && eq
	}

	return bothEqual, nil
}
