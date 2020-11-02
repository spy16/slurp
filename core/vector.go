package core

// Vector is an ordered collection providing fast random access.
type Vector interface {
	Seqable
	SExpressable

	// Count returns the number of elements contained in the Vector.
	Count() (int, error)

	// Assoc takes a value and "associates" it to the Vector,
	// assigning it to the index i.
	Assoc(i int, val Any) (Vector, error)

	// EntryAt i returns the ith element in the Vector.
	EntryAt(i int) (Any, error)

	// Pop returns a copy of the Vector without its last element.
	Pop() (Vector, error)
}
