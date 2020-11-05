package builtin

import (
	"fmt"
	"testing"

	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const size = 2048

func TestVectorIsHashable(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r != nil {
			t.Error("PersistentVector is not hashable.")
		}
	}()

	m := make(map[core.Vector]struct{})
	m[EmptyVector] = struct{}{}
}

func TestSeqToVector(t *testing.T) {
	seq := NewList(Int64(0), Keyword("keyword"), String("string"))
	v, err := SeqToVector(seq)
	require.NoError(t, err)
	require.NotNil(t, v)

	var i int
	_ = core.ForEach(seq, func(want core.Any) (bool, error) {
		ve, err := v.EntryAt(i)
		require.NoError(t, err, "iteration %d", i)

		assert.Equal(t, want, ve)
		i++
		return false, nil
	})
}

func TestEmptyVector(t *testing.T) {
	t.Parallel()

	require.NotZero(t, EmptyVector,
		"zero-value empty vector is invalid (shift is missing)")

	t.Run("SExpr", func(t *testing.T) {
		testSExpr(t, EmptyVector, "[]")
	})

	t.Run("Seq", func(t *testing.T) {
		seq, err := EmptyVector.Seq()
		assert.NoError(t, err)
		assert.NotNil(t, seq)
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()

		cnt, err := EmptyVector.Count()
		assert.NoError(t, err)
		assert.Zero(t, cnt, "EmptyVector has a non-zero count")
	})

	t.Run("EntryAt", func(t *testing.T) {
		t.Parallel()

		v, err := EmptyVector.EntryAt(0)
		assert.EqualError(t, err, ErrIndexOutOfBounds.Error())
		assert.Nil(t, v)
	})

	t.Run("AssocOutOfBounds", func(t *testing.T) {
		t.Parallel()

		t.Run("PersistentVector", func(t *testing.T) {
			v, err := EmptyVector.Assoc(9001, Nil{})
			assert.EqualError(t, err, ErrIndexOutOfBounds.Error())
			assert.Nil(t, v)
		})

		t.Run("TransientVector", func(t *testing.T) {
			v, err := EmptyVector.asTransient().Assoc(9001, Nil{})
			assert.EqualError(t, err, ErrIndexOutOfBounds.Error())
			assert.Nil(t, v)
		})
	})
}

func TestPersistentVector(t *testing.T) {
	t.Parallel()

	as := make([]core.Any, size)
	for i := 0; i < size; i++ {
		as[i] = Int64(i)
	}

	t.Run("SExpr", func(t *testing.T) {
		t.Parallel()

		testSExpr(t, NewVector(Int64(0), Keyword("keyword"), String("string")),
			"[0 :keyword \"string\"]")
	})

	t.Run("Conj", func(t *testing.T) {
		t.Parallel()

		t.Run("Nop", func(t *testing.T) {
			v, err := EmptyVector.Conj()
			require.NoError(t, err)
			require.NotNil(t, v)
			assert.Equal(t, EmptyVector, v)
		})

		t.Run("PersistentConj", func(t *testing.T) {
			v, err := EmptyVector.Conj(Nil{})
			require.NoError(t, err)
			require.NotNil(t, v)

			cnt, err := v.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, 1, cnt, "Count() returned incorrect value '%d'", cnt)
		})

		t.Run("TransientConj", func(t *testing.T) {
			v, err := EmptyVector.Conj(as...)
			require.NoError(t, err)
			require.NotNil(t, v)

			cnt, err := v.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, size, cnt, "Count() returned incorrect value '%d'", cnt)

			for i, any := range as {
				val, err := v.EntryAt(i)
				require.NoError(t, err, "EntryAt() failed")
				require.NotNil(t, val, "EntryAt() returned a nil value")
				require.Equal(t, any, val,
					"value recovered does not match associated value")
			}
		})

	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()

		var err error
		var v core.Vector = EmptyVector
		for i, any := range as {
			v, err = v.Assoc(i, any)
			require.NoError(t, err, "Assoc() failed")
			require.NotNil(t, v, "Assoc() returned a nil vector")

			val, err := v.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, any, val,
				"value recovered does not match associated value")

			cnt, err := v.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, i+1, cnt, "Count() returned incorrect value '%d'", cnt)
		}
	})

	t.Run("Replace", func(t *testing.T) {
		t.Parallel()

		v := NewVector(as...)
		for i := range as {
			vPrime, err := v.Assoc(i, Nil{})
			assert.NoError(t, err)
			assert.NotNil(t, vPrime)

			val, err := vPrime.EntryAt(i)
			assert.NoError(t, err)
			assert.Equal(t, Nil{}, val)
		}
	})

	t.Run("Pop", func(t *testing.T) {
		t.Parallel()

		var v core.Vector = NewVector(as...)

		cnt, err := v.Count()
		require.NoError(t, err, "test precondition failed")
		require.Equal(t, size, cnt)

		for i := range as {
			v, err = v.Pop()
			require.NoError(t, err, "iteration %d", i)
			require.NotNil(t, v, "iteration %d", i)

			cnt, err = v.Count()
			require.NoError(t, err, "iteration %d", i)
			require.Equal(t, size-1-i, cnt, "iteration %d", i)
		}

		v, err = v.Pop()
		assert.EqualError(t, err, "cannot pop from empty vector")
		assert.Nil(t, v)
	})

	t.Run("Seq", func(t *testing.T) {
		seq, err := EmptyVector.Seq()
		require.NoError(t, err)
		require.NotNil(t, seq)

		seq, err = seq.Conj(as[1:]...)
		require.NoError(t, err)
		require.NotNil(t, seq)

		wants := make([]core.Any, len(as))
		copy(wants, as)
		for left, right := 0, len(wants)-1; left < right; left, right = left+1, right-1 {
			wants[left], wants[right] = wants[right], wants[left]
		}

		var i int
		err = core.ForEach(seq, func(got core.Any) (bool, error) {
			if !assert.Equal(t, wants[i], got) {
				return true, nil
			}

			i++
			return false, nil
		})
	})
}

func TestTransientVector(t *testing.T) {
	t.Parallel()

	as := make([]core.Any, size)
	for i := 0; i < size; i++ {
		as[i] = Int64(i)
	}

	t.Run("NewTransientVector", func(t *testing.T) {
		t.Parallel()

		v := newTransientVector(as...)
		assert.NotNil(t, v)

		cnt, err := v.Count()
		assert.NoError(t, err)
		assert.Equal(t, size, cnt)

	})

	t.Run("SExpr", func(t *testing.T) {
		t.Parallel()

		vec := newTransientVector(Int64(0), Keyword("keyword"), String("string"))
		testSExpr(t, vec, "[0 :keyword \"string\"]")
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()

		v := newTransientVector(as...)

		cnt, err := v.Count()
		assert.NoError(t, err)
		assert.Equal(t, size, cnt)
	})

	t.Run("Conj", func(t *testing.T) {
		t.Parallel()

		v, err := newTransientVector().Conj(as...)
		require.NoError(t, err)
		require.NotNil(t, v)

		cnt, err := v.Count()
		require.NoError(t, err, "Count() failed")
		require.Equal(t, size, cnt, "Count() returned incorrect value '%d'", cnt)

		for i, any := range as {
			val, err := v.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, any, val,
				"value recovered does not match associated value")
		}
	})

	t.Run("Append", func(t *testing.T) {
		t.Parallel()

		v := EmptyVector.asTransient()

		for i := 0; i < size; i++ {
			vPrime, err := v.Assoc(i, Int64(i))
			require.NoError(t, err, "Assoc() failed")
			require.NotNil(t, vPrime, "Assoc() returned a nil vector")

			val, err := v.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, Int64(i), val,
				"value recovered does not match associated value")

			val, err = vPrime.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, Int64(i), val,
				"value recovered does not match associated value")

			cnt, err := v.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, i+1, cnt, "Count() returned incorrect value '%d'", cnt)

			cnt, err = vPrime.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, i+1, cnt, "Count() returned incorrect value '%d'", cnt)
		}
	})

	t.Run("Replace", func(t *testing.T) {
		t.Parallel()

		var v core.Vector = newTransientVector(as...)
		for i := range as {
			vPrime, err := v.Assoc(i, Nil{})
			assert.NoError(t, err)
			assert.NotNil(t, vPrime)

			val, err := v.EntryAt(i)
			assert.NoError(t, err)
			assert.Equal(t, Nil{}, val)

			val, err = vPrime.EntryAt(i)
			assert.NoError(t, err)
			assert.Equal(t, Nil{}, val)
		}
	})

	t.Run("Pop", func(t *testing.T) {
		t.Parallel()

		var v core.Vector = newTransientVector(as...)

		cnt, err := v.Count()
		require.NoError(t, err, "test precondition failed")
		require.Equal(t, size, cnt)

		for i := range as {
			vPrime, err := v.Pop()
			require.NoError(t, err, "iteration %d", i)
			require.NotNil(t, vPrime, "iteration %d", i)

			cnt, err = v.Count()
			require.NoError(t, err, "iteration %d", i)
			require.Equal(t, size-1-i, cnt, "iteration %d", i)

			cnt, err = vPrime.Count()
			require.NoError(t, err, "iteration %d", i)
			require.Equal(t, size-1-i, cnt, "iteration %d", i)
		}

		v, err = v.Pop()
		assert.EqualError(t, err, "cannot pop from empty vector")
		assert.Nil(t, v)
	})

	t.Run("Seq", func(t *testing.T) {
		v := newTransientVector(as...)
		seq, err := v.Seq()
		require.NoError(t, err)

		var i int
		err = core.ForEach(seq, func(want core.Any) (bool, error) {
			got, err := v.EntryAt(i)
			if err != nil {
				return true, fmt.Errorf("%w (%d)", err, i)
			}

			if !assert.Equal(t, want, got) {
				return true, nil
			}

			i++
			return false, nil
		})

		assert.NoError(t, err)
	})

	t.Run("Invariants", func(t *testing.T) {
		t.Parallel()

		v := EmptyVector.asTransient()
		v.Conj(Nil{})

		assert.NotEqual(t, EmptyVector, v,
			"derived transient mutated EmptyVector")

		assert.Equal(t, EmptyVector, EmptyVector.asTransient().persistent(),
			"persistent() ∘ asTransient() ∘ persistent() ∘ EmptyVector != EmptyVector")
	})
}

func TestVectorBuilder(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		var b VectorBuilder
		assert.Equal(t, EmptyVector, b.Vector())
	})

	t.Run("NonEmpty", func(t *testing.T) {
		var b VectorBuilder
		for i := 0; i < size; i++ {
			b.Cons(Int64(i))
		}

		v := b.Vector()
		assert.NotZero(t, v)

		n, err := b.Vector().Count()
		assert.NoError(t, err)
		assert.Equal(t, size, n)
		assert.Panics(t, func() { b.Cons(Nil{}) })
	})

}
