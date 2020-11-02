package builtin

import (
	"errors"

	"github.com/spy16/slurp/core"
)

// var _ core.Vector = (*Vector)(nil)

const (
	bits  = 5 // number of bits needed to represent the range (0 32].
	width = 32
	mask  = width - 1 // 0x1f
)

var (
	// ErrIndexOutOfBounds is returned when a sequence's index is out of range.
	ErrIndexOutOfBounds = errors.New("index out of bounds")

	// EmptyVector is the zero-value PersistentVector
	EmptyVector = PersistentVector{shift: bits}
)

// PersistentVector is an immutable core.Vector implementation with O(1) lookup,
// insertion, appending, and deletion.
type PersistentVector struct {
	cnt, shift int
	root, tail node
}

// Count returns the number of elements contained in the Vector.
func (v PersistentVector) Count() (int, error) { return v.cnt, nil }

// SExpr returns a parsable s-expression for the Vector.
func (v PersistentVector) SExpr() (string, error) {
	if v.cnt == 0 {
		return "[]", nil
	}

	seq, err := v.Seq()
	if err != nil {
		return "", err
	}

	return core.SeqString(seq, "[", "]", " ")
}

func (v PersistentVector) tailoff() int {
	if v.cnt < width {
		return 0
	}

	return ((v.cnt - 1) >> bits) << bits
}

func (v PersistentVector) nodeFor(i int) (node, error) {
	if i >= 0 && i < v.cnt {
		if i >= v.tailoff() {
			return v.tail, nil
		}

		n := v.root
		for level := v.shift; level > 0; level -= bits {
			n = n.array[(i>>level)&mask].(node) // TODO:  unsafe.Pointer
		}

		return n, nil
	}

	return node{}, ErrIndexOutOfBounds
}

// EntryAt i returns the ith entry in the Vector
func (v PersistentVector) EntryAt(i int) (core.Any, error) {
	n, err := v.nodeFor(i)
	if err != nil {
		return nil, err
	}

	return n.array[i&mask], nil
}

// Assoc takes a value and "associates" it to the Vector,
// assigning it to the index i.
func (v PersistentVector) Assoc(i int, val core.Any) (core.Vector, error) {
	vv, err := v.assoc(i, val)
	if err != nil {
		return nil, err
	}

	return vv, nil
}

func (v PersistentVector) assoc(i int, val core.Any) (PersistentVector, error) {
	if i >= 0 && i < v.cnt {
		if i >= v.tailoff() {
			newTail := v.tail.copy()
			newTail.array[i&mask] = val
			return PersistentVector{
				cnt:   v.cnt,
				shift: v.shift,
				root:  v.root,
				tail:  newTail,
			}, nil
		}

		return PersistentVector{
			cnt:   v.cnt,
			shift: v.shift,
			root:  v.doAssoc(v.shift, v.root, i, val),
			tail:  v.tail,
		}, nil
	}

	if i == v.cnt {
		return v.cons(val), nil
	}

	return PersistentVector{}, ErrIndexOutOfBounds
}

func (v PersistentVector) doAssoc(level int, n node, i int, val core.Any) node {
	ret := n
	if level == 0 {
		ret.array[i&mask] = val
	} else {
		subidx := (i >> level) & mask
		ret.array[subidx] = v.doAssoc(level-bits, n.array[subidx].(node), i, val) // TODO: unsafe.Pointer
	}

	return ret
}

// Cons appends a value to the Vector.
func (v PersistentVector) Cons(val core.Any) core.Vector { return v.cons(val) }

func (v PersistentVector) cons(val core.Any) PersistentVector {
	// room in tail?
	if v.cnt-v.tailoff() < 32 {
		newTail := v.tail.copy()
		newTail.len++
		newTail.array[v.tail.len] = val

		return PersistentVector{
			cnt:   v.cnt + 1,
			shift: v.shift,
			root:  v.root,
			tail:  newTail,
		}
	}

	// full tail; push into trie
	var newRoot node
	tailNode := v.tail.copy()
	newShift := v.shift

	// overflow root?
	if (v.cnt >> bits) > (1 << v.shift) {
		newRoot.len += 2
		newRoot.array[0] = v.root
		newRoot.array[1] = newPath(v.shift, tailNode)
		newShift += bits
	} else {
		newRoot = v.pushTail(v.shift, v.root, tailNode)
	}

	return PersistentVector{
		cnt:   v.cnt + 1,
		shift: newShift,
		root:  newRoot,
		tail:  newNode(val),
	}
}

func newPath(level int, n node) node {
	if level == 0 {
		return n
	}

	return newNode(newPath(level-bits, n))
}

func (v PersistentVector) pushTail(level int, parent, tailNode node) node {
	//if parent is leaf, insert node,
	// else does it map to an existing child? -> nodeToInsert = pushNode one more level
	// else alloc new path
	//return  nodeToInsert placed in copy of parent

	subidx := ((v.cnt - 1) >> level) & mask
	ret := parent.copy()

	var nodeToInsert node

	if level == bits {
		nodeToInsert = tailNode
	} else {
		if child := parent.array[subidx]; child != nil {
			nodeToInsert = v.pushTail(level-bits, child.(node), tailNode) // TODO: unsafe.Pointer
		} else {
			nodeToInsert = newPath(level-bits, tailNode)
		}
	}

	ret.array[subidx] = nodeToInsert
	return ret
}

// Pop returns a copy of the Vector without its last element.
func (v PersistentVector) Pop() (core.Vector, error) {
	panic("PersistentVector.Pop() NOT IMPLEMENTED")
}

// Seq returns a sequence representation of the underlying Vector.
// Note that the resulting Seq type has Vector semantics for Conj().
func (v PersistentVector) Seq() (core.Seq, error) { return newChunkedSeq(v, 0, 0) }

type node struct {
	len   int
	array [width]interface{}
}

func newNode(vs ...interface{}) (n node) {
	n.len = len(vs)
	for i, v := range vs {
		n.array[i] = v
	}
	return
}

func (n node) copy() (nn node) {
	nn.len = n.len
	for i, v := range n.array {
		nn.array[i] = v
	}
	return nn
}

type chunkedSeq struct {
	vec       PersistentVector
	node      node
	i, offset int
}

func newChunkedSeq(v PersistentVector, i, offset int) (seq chunkedSeq, err error) {
	seq.vec = v
	seq.i = i
	seq.offset = offset
	seq.node, err = v.nodeFor(i)
	return
}

func (cs chunkedSeq) Count() (int, error) { return cs.vec.cnt - (cs.i + cs.offset), nil }

func (cs chunkedSeq) First() (core.Any, error) { return cs.node.array[cs.offset], nil }

func (cs chunkedSeq) Next() (core.Seq, error) {
	if cs.offset+1 < cs.node.len {
		return chunkedSeq{
			vec:    cs.vec,
			node:   cs.node,
			i:      cs.i,
			offset: cs.offset + 1,
		}, nil
	}

	return cs.chunkedNext()
}

func (cs chunkedSeq) chunkedNext() (core.Seq, error) {
	if cs.i+cs.node.len < cs.vec.cnt {
		return newChunkedSeq(cs.vec, cs.i+cs.node.len, 0)
	}

	return nil, nil
}

func (cs chunkedSeq) Conj(items ...core.Any) (_ core.Seq, err error) {
	i, _ := cs.vec.Count()

	// TODO(performance):  transient vector if len(items) > 1
	for _, v := range items {
		if cs.vec, err = cs.vec.assoc(i, v); err != nil {
			return
		}
	}

	return newChunkedSeq(cs.vec, cs.i, cs.offset)
}
