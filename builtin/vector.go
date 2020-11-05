package builtin

import (
	"errors"
	"fmt"

	"github.com/spy16/slurp/core"
)

var (
	_ core.Vector = (*PersistentVector)(nil)
	_ core.Vector = (*transientVector)(nil)
)

const (
	bits  = 5 // number of bits needed to represent the range (0 32].
	width = 32
	mask  = width - 1 // 0x1f
)

var (
	// ErrIndexOutOfBounds is returned when a sequence's index is out of range.
	ErrIndexOutOfBounds = errors.New("index out of bounds")

	// EmptyVector is the zero-value PersistentVector
	EmptyVector = PersistentVector{
		shift: bits,
		root:  emptyNode,
		tail:  emptyNode,
	}

	emptyNode = new(node)
)

// PersistentVector is an immutable core.Vector implementation with O(1) lookup,
// insertion, appending, and deletion.
type PersistentVector struct {
	cnt, shift int
	root, tail *node
}

// NewVector builds a PersistentVector efficiently.
func NewVector(items ...core.Any) PersistentVector {
	return newTransientVector(items...).persistent()
}

// SeqToVector efficiently builds a PersistentVector from a Seq.
func SeqToVector(seq core.Seq) (PersistentVector, error) {
	vec := EmptyVector.asTransient()
	err := core.ForEach(seq, func(val core.Any) (bool, error) {
		_, _ = vec.Conj(val)
		return false, nil
	})
	return vec.persistent(), err
}

func (v PersistentVector) asTransient() *transientVector {
	return &transientVector{
		cnt:   v.cnt,
		shift: v.shift,
		root:  v.root.clone(),
		tail:  v.tail.clone(),
	}
}

// Count returns the number of elements contained in the Vector.
func (v PersistentVector) Count() (int, error) { return v.cnt, nil }

// SExpr returns a parsable s-expression for the Vector.
func (v PersistentVector) SExpr() (string, error) {
	if v.cnt == 0 {
		return "[]", nil
	}

	seq, _ := v.Seq()
	return core.SeqString(seq, "[", "]", " ")
}

func (v PersistentVector) tailoff() int {
	if v.cnt < width {
		return 0
	}

	return ((v.cnt - 1) >> bits) << bits
}

func (v PersistentVector) nodeFor(i int) (*node, error) {
	if i >= 0 && i < v.cnt {
		if i >= v.tailoff() {
			return v.tail, nil
		}

		n := v.root
		for level := v.shift; level > 0; level -= bits {
			n = n.array[(i>>level)&mask].(*node) // TODO:  unsafe.Pointer
		}

		return n, nil
	}

	return nil, ErrIndexOutOfBounds
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
	vec, err := v.assoc(i, val)
	if err != nil {
		return nil, err
	}

	return vec, nil
}

func (v PersistentVector) assoc(i int, val core.Any) (PersistentVector, error) {
	if i >= 0 && i < v.cnt {
		if i >= v.tailoff() {
			newTail := v.tail.clone()
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

func (v PersistentVector) doAssoc(level int, n *node, i int, val core.Any) *node {
	ret := n
	if level == 0 {
		ret.array[i&mask] = val
	} else {
		subidx := (i >> level) & mask
		ret.array[subidx] = v.doAssoc(level-bits, n.array[subidx].(*node), i, val) // TODO: unsafe.Pointer
	}

	return ret
}

// Conj conjoins a value to the vector, appending it to the tail.
func (v PersistentVector) Conj(vs ...core.Any) (core.Vector, error) { return v.Cons(vs...) }

// Cons appends a value to the Vector.
func (v PersistentVector) Cons(vs ...core.Any) (core.Vector, error) {
	switch len(vs) {
	case 0:
		return v, nil

	case 1:
		return v.cons(vs[0]), nil

	default:
		head, vs := vs[0], vs[1:]
		t := v.cons(head).asTransient()
		for _, val := range vs {
			_ = t.cons(val)
		}
		return t.persistent(), nil

	}
}

func (v PersistentVector) cons(val core.Any) PersistentVector {
	// room in tail?
	if v.cnt-v.tailoff() < 32 {
		newTail := v.tail.clone()
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
	newRoot := &node{}
	tailNode := v.tail.clone()
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

func newPath(level int, n *node) *node {
	if level == 0 {
		return n
	}

	return newNode(newPath(level-bits, n))
}

func (v PersistentVector) pushTail(level int, parent, tailNode *node) *node {
	//if parent is leaf, insert node,
	// else does it map to an existing child? -> nodeToInsert = pushNode one more level
	// else alloc new path
	//return  nodeToInsert placed in copy of parent

	subidx := ((v.cnt - 1) >> level) & mask
	ret := parent.clone()

	var nodeToInsert *node

	if level == bits {
		nodeToInsert = tailNode
	} else {
		if child := parent.array[subidx]; child != nil {
			nodeToInsert = v.pushTail(level-bits, child.(*node), tailNode) // TODO: unsafe.Pointer
		} else {
			nodeToInsert = newPath(level-bits, tailNode)
		}
	}

	ret.array[subidx] = nodeToInsert
	return ret
}

// Pop returns a copy of the Vector without its last element.
func (v PersistentVector) Pop() (core.Vector, error) {
	if v.cnt == 0 {
		return nil, errors.New("cannot pop from empty vector")
	}

	if v.cnt == 1 {
		return EmptyVector, nil
	}

	// len(tail) > 1 ?
	if v.cnt-v.tailoff() > 1 {
		newTail := &node{len: v.tail.len - 1}
		copy(newTail.array[:newTail.len], v.tail.array[:])

		return PersistentVector{
			cnt:   v.cnt - 1,
			shift: v.shift,
			root:  v.root,
			tail:  newTail,
		}, nil
	}

	newTail, err := v.nodeFor(v.cnt - 2)
	if err != nil {
		// TODO: we *should* be able to remove this error check.
		//	     If this panic is triggered and the vector is a correct state (i.e.
		// 		 lthibault was wrong and the error check CANNOT be removed), just return
		//	     the error.
		panic(fmt.Errorf("unreachable: %w", err))
	}

	newRoot := v.popTail(v.shift, v.root)
	newShift := v.shift
	if newRoot == nil {
		newRoot = emptyNode
	}
	if v.shift > bits && newRoot.array[1] == nil {
		if newRoot.array[0] == nil {
			newRoot = emptyNode
		}
		newShift -= bits
	}

	return PersistentVector{
		cnt:   v.cnt - 1,
		shift: newShift,
		root:  newRoot,
		tail:  newTail,
	}, nil
}

func (v PersistentVector) popTail(level int, n *node) *node {
	subidx := ((v.cnt - 2) >> level) & mask
	if level > bits {
		newChild := v.popTail(level-bits, n.array[subidx].(*node)) // TODO: unsafe.Pointer
		if newChild == nil && subidx == 0 {
			return nil
		}

		ret := n.clone()
		ret.array[subidx] = newChild
		// ret.len++
		return ret
	} else if subidx == 0 {
		return nil
	}

	ret := n.clone()
	ret.array[subidx] = node{}
	return ret
}

// Seq returns a sequence representation of the underlying Vector.
// Note that the resulting Seq type has Vector semantics for Conj().
func (v PersistentVector) Seq() (core.Seq, error) { return newChunkedSeq(v, 0, 0), nil }

type node struct {
	len   int
	array [width]interface{}
}

func newNode(vs ...interface{}) *node {
	n := &node{len: len(vs)}
	for i, v := range vs {
		n.array[i] = v
	}
	return n
}

func (n node) clone() *node {
	nn := &node{len: n.len}
	for i, v := range n.array {
		nn.array[i] = v
	}
	return nn
}

type chunkedSeq struct {
	vec       PersistentVector
	node      *node
	i, offset int
}

func newChunkedSeq(v PersistentVector, i, offset int) chunkedSeq {
	n, err := v.nodeFor(i)
	if err != nil {
		n = &node{}
	}

	return chunkedSeq{
		vec:    v,
		node:   n,
		i:      i,
		offset: offset,
	}
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
		return newChunkedSeq(cs.vec, cs.i+cs.node.len, 0), nil
	}

	return nil, nil
}

func (cs chunkedSeq) Conj(items ...core.Any) (_ core.Seq, err error) {
	i, _ := cs.vec.Count()

	// TODO(performance):  transient vector if len(items) > 1
	for _, v := range items {
		cs.vec, _ = cs.vec.assoc(i, v)
	}

	return newChunkedSeq(cs.vec, cs.i, cs.offset), nil
}

// TransientVector is used to efficiently build a PersistentVector using the Conj method.
// It minimizes memory copying.
type transientVector PersistentVector

func newTransientVector(items ...core.Any) *transientVector {
	vec := EmptyVector.asTransient()
	for _, val := range items {
		_, _ = vec.Conj(val)
	}
	return vec
}

// N.B.:  transientVector must not be modified after call to persistent()
func (t transientVector) persistent() PersistentVector { return PersistentVector(t) }

func (t transientVector) tailoff() int { return PersistentVector(t).tailoff() }

func (t transientVector) SExpr() (string, error) { return PersistentVector(t).SExpr() }

func (t transientVector) Count() (int, error) { return t.cnt, nil }

func (t transientVector) Seq() (core.Seq, error) { return PersistentVector(t).Seq() }

func (t *transientVector) Conj(vs ...core.Any) (core.Vector, error) { return t.Cons(vs...) }

func (t *transientVector) Cons(vs ...core.Any) (core.Vector, error) {
	for _, val := range vs {
		t.cons(val)
	}

	return t, nil
}

func (t *transientVector) cons(val core.Any) *transientVector {
	// room in tail?
	if t.cnt-t.tailoff() < 32 {
		t.tail.array[t.cnt&mask] = val
		t.tail.len++
		t.cnt++
		return t
	}

	// full tail; push into trie
	newRoot := &node{}
	tailNode := t.tail.clone()
	t.tail = newNode(val)
	newShift := t.shift

	// overflow root?
	if (t.cnt >> bits) > (1 << t.shift) {
		newRoot.len += 2
		newRoot.array[0] = t.root
		newRoot.array[1] = newPath(t.shift, tailNode)
		newShift += 5
	} else {
		newRoot = t.pushTail(t.shift, t.root, tailNode)
	}

	t.root = newRoot
	t.shift = newShift
	t.cnt++
	return t
}

func (t *transientVector) pushTail(level int, parent, tailNode *node) *node {
	//if parent is leaf, insert node,
	// else does it map to an existing child? -> nodeToInsert = pushNode one more level
	// else alloc new path
	//return  nodeToInsert placed in parent

	subidx := ((t.cnt - 1) >> level) & mask
	ret := parent // mutable; don't clone
	var nodeToInsert *node
	if level == bits {
		nodeToInsert = tailNode
	} else {
		if child := parent.array[subidx]; child != nil {
			nodeToInsert = t.pushTail(level-bits, child.(*node), tailNode) // TODO: unsafe.Pointer
		} else {
			nodeToInsert = newPath(level-bits, tailNode)
		}
	}

	ret.array[subidx] = nodeToInsert
	return ret
}

func (t *transientVector) nodeFor(i int) (*node, error) {
	return (*PersistentVector)(t).nodeFor(i)
}

func (t *transientVector) Assoc(i int, val core.Any) (core.Vector, error) {
	return t.assoc(i, val)
}

func (t *transientVector) assoc(i int, val core.Any) (*transientVector, error) {
	if i >= 0 && i < t.cnt {
		if i >= t.tailoff() {
			t.tail.array[i&mask] = val
			return t, nil
		}

		t.root = t.doAssoc(t.shift, t.root, i, val)
		return t, nil
	}

	if i == t.cnt {
		return t.cons(val), nil
	}

	return nil, ErrIndexOutOfBounds
}

func (t *transientVector) doAssoc(level int, n *node, i int, val core.Any) *node {
	ret := n
	if level == 0 {
		ret.array[i&mask] = val
	} else {
		subidx := (i >> level) & mask
		ret.array[subidx] = t.doAssoc(level-5, n.array[subidx].(*node), i, val)
	}

	return ret
}

func (t transientVector) EntryAt(i int) (core.Any, error) {
	return PersistentVector(t).EntryAt(i)
}

func (t *transientVector) Pop() (core.Vector, error) {
	if t.cnt == 0 {
		return nil, errors.New("cannot pop from empty vector")
	}

	if t.cnt == 1 { // TODO:  is this block necessary?
		t.cnt = 0
		t.tail.len = 0
		return t, nil
	}

	// pop from tail?
	if t.cnt&mask > 0 {
		t.cnt--
		t.tail.len--
		return t, nil
	}

	newTail, err := t.nodeFor(t.cnt - 2)
	if err != nil {
		// TODO: we *should* be able to remove this error check.
		//	     If this panic is triggered and the vector is a correct state (i.e.
		// 		 lthibault was wrong and the error check CANNOT be removed), just return
		//	     the error.
		panic(fmt.Errorf("unreachable: %w", err))
	}

	newRoot := t.popTail(t.shift, t.root)
	newShift := t.shift

	if newRoot == nil {
		newRoot = &node{}
	}
	if t.shift > 5 && newRoot.array[1] == nil {
		if newRoot.array[0] == nil {
			newRoot = &node{}
		} else {
			newRoot = newRoot.array[0].(*node) // TODO:  unsafe.Pointer
		}
		newShift -= bits
	}

	t.cnt--
	t.shift = newShift
	t.root = newRoot
	t.tail = newTail
	return t, nil
}

func (t *transientVector) popTail(level int, n *node) *node {
	subidx := (t.cnt - 2>>level) & mask
	if level > bits {
		newChild := t.popTail(level-bits, n.array[subidx].(*node))
		if newChild == nil && subidx == 0 {
			return nil
		}

		ret := n
		ret.array[subidx] = newChild
		return ret
	} else if subidx == 0 {
		return nil
	}

	ret := n
	ret.array[subidx] = nil
	return ret
}

// VectorBuilder is used to efficiently build a PersistentVector using the Cons method.
// It minimizes memory copying. The zero value is ready to use.  Do not copy a
// VectorBuilder after first use.
type VectorBuilder struct {
	persisted bool
	vec       *transientVector
}

// Cons constructs a vector by append the items sequentially to the tail of the vector.
func (v *VectorBuilder) Cons(item ...core.Any) {
	if v.persisted == true {
		panic("vector already persisted")
	}

	if v.vec == nil {
		v.vec = newTransientVector()
	}

	_, _ = v.vec.Cons(item...)
}

// Vector returns the constructed vector.  VectorBuilder must not be used after a call
// to Vector().
func (v *VectorBuilder) Vector() PersistentVector {
	v.persisted = true
	if v.vec == nil {
		return EmptyVector
	}

	return v.vec.persistent()
}
