package builtin

import "github.com/spy16/slurp/core"

var (
	_ core.Any = (*Vector)(nil)
	_ core.Seq = (*Vector)(nil)
)

// Vector is an ordered sequence providing fast random-access.
type Vector []core.Any

// SExpr returns a valid s-expression for LinkedList.
func (v Vector) SExpr() (string, error) {
	if v == nil {
		return "[]", nil
	}
	return core.SeqString(v, "[", "]", " ")
}

// Conj returns a new vector with all the items added at the tail of the vector.
func (v Vector) Conj(items ...core.Any) (core.Seq, error) {
	vv := make([]core.Any, len(v)+len(items))
	copy(vv, v)
	copy(vv[len(v):], items)
	return Vector(vv), nil
}

// First returns the head or first item of the vector.
func (v Vector) First() (core.Any, error) { return v[0], nil }

// Next returns the tail of the vector.
func (v Vector) Next() (core.Seq, error) { return v[1:], nil }

// Count returns the number of the vector.
func (v Vector) Count() (int, error) { return len(v), nil }
