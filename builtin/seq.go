package builtin

import (
	"github.com/spy16/slurp/core"
)

var (
	_ core.Any = (*LinkedList)(nil)
	_ core.Seq = (*LinkedList)(nil)
)

// Cons returns a new seq with `v` added as the first and `seq` as the rest.
// seq can be nil as well.
func Cons(v core.Any, seq core.Seq) (core.Seq, error) {
	newSeq := &LinkedList{first: v, rest: seq, count: 1}

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
func NewList(items ...core.Any) core.Seq {
	if len(items) == 0 {
		return core.Seq((*LinkedList)(nil))
	}

	var err error
	lst := core.Seq(&LinkedList{})
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
	first core.Any
	rest  core.Seq
}

// Conj returns a new list with all the items added at the head of the list.
func (ll *LinkedList) Conj(items ...core.Any) (res core.Seq, err error) {
	res = ll
	if ll == nil {
		res = &LinkedList{}
	}

	for _, item := range items {
		if res, err = Cons(item, res); err != nil {
			break
		}
	}

	return res, err
}

// First returns the head or first item of the list.
func (ll *LinkedList) First() (core.Any, error) {
	if ll == nil {
		return nil, nil
	}
	return ll.first, nil
}

// Next returns the tail of the list.
func (ll *LinkedList) Next() (core.Seq, error) {
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
