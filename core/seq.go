package core

import (
	"fmt"
	"strings"
)

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

// ToSlice converts the given sequence into a slice.
func ToSlice(seq Seq) ([]Any, error) {
	var sl []Any
	err := ForEach(seq, func(item Any) (bool, error) {
		sl = append(sl, item)
		return false, nil
	})
	return sl, err
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

	return err
}

// SeqString returns a string representation for the sequence with given prefix
// suffix and separator.
func SeqString(seq Seq, begin, end, sep string) (string, error) {
	var b strings.Builder
	b.WriteString(begin)
	err := ForEach(seq, func(item Any) (bool, error) {
		if se, ok := item.(SExpressable); ok {
			s, err := se.SExpr()
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
