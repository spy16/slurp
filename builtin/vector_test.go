package builtin_test

import (
	"fmt"
	"testing"

	"github.com/spy16/slurp/builtin"
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
	m[builtin.EmptyVector] = struct{}{}
}

func TestSeqToVector(t *testing.T) {
	seq := builtin.NewList(builtin.Int64(0), builtin.Keyword("keyword"), builtin.String("string"))
	v, err := builtin.SeqToVector(seq)
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

	require.NotZero(t, builtin.EmptyVector,
		"zero-value empty vector is invalid (shift is missing)")

	t.Run("Seq", func(t *testing.T) {
		seq, err := builtin.EmptyVector.Seq()
		assert.NoError(t, err)
		assert.NotNil(t, seq)
	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()

		cnt, err := builtin.EmptyVector.Count()
		assert.NoError(t, err)
		assert.Zero(t, cnt, "builtin.EmptyVector has a non-zero count")
	})

	t.Run("EntryAt", func(t *testing.T) {
		t.Parallel()

		v, err := builtin.EmptyVector.EntryAt(0)
		assert.ErrorIs(t, err, builtin.ErrIndexOutOfBounds)
		assert.Nil(t, v)
	})

	t.Run("AssocOutOfBounds", func(t *testing.T) {
		t.Parallel()

		t.Run("PersistentVector", func(t *testing.T) {
			v, err := builtin.EmptyVector.Assoc(9001, builtin.Nil{})
			assert.ErrorIs(t, err, builtin.ErrIndexOutOfBounds)
			assert.Nil(t, v)
		})

		t.Run("TransientVector", func(t *testing.T) {
			v, err := builtin.EmptyVector.Transient().Assoc(9001, builtin.Nil{})
			assert.ErrorIs(t, err, builtin.ErrIndexOutOfBounds)
			assert.Nil(t, v)
		})
	})
}

func TestPersistentVector(t *testing.T) {
	t.Parallel()

	as := make([]core.Any, size)
	for i := 0; i < size; i++ {
		as[i] = builtin.Int64(i)
	}

	t.Run("Conj", func(t *testing.T) {
		t.Parallel()

		t.Run("Nop", func(t *testing.T) {
			v, err := builtin.EmptyVector.Conj()
			require.NoError(t, err)
			require.NotNil(t, v)
			assert.Equal(t, builtin.EmptyVector, v)
		})

		t.Run("PersistentConj", func(t *testing.T) {
			v, err := builtin.EmptyVector.Conj(builtin.Nil{})
			require.NoError(t, err)
			require.NotNil(t, v)

			cnt, err := v.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, 1, cnt, "Count() returned incorrect value '%d'", cnt)
		})

		t.Run("TransientConj", func(t *testing.T) {
			v, err := builtin.EmptyVector.Conj(as...)
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
		var v core.Vector = builtin.EmptyVector
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

		v := builtin.NewVector(as...)
		for i := range as {
			vPrime, err := v.Assoc(i, builtin.Nil{})
			assert.NoError(t, err)
			assert.NotNil(t, vPrime)

			val, err := vPrime.EntryAt(i)
			assert.NoError(t, err)
			assert.Equal(t, builtin.Nil{}, val)
		}
	})

	t.Run("Pop", func(t *testing.T) {
		t.Parallel()

		var v core.Vector = builtin.NewVector(as...)

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
		seq, err := builtin.EmptyVector.Seq()
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
		require.NoError(t, err)
	})
}

func TestTransientVector(t *testing.T) {
	t.Parallel()

	as := make([]core.Any, size)
	for i := 0; i < size; i++ {
		as[i] = builtin.Int64(i)
	}

	t.Run("NewTransientVector", func(t *testing.T) {
		t.Parallel()

		v := builtin.NewVector(as...).Transient()
		assert.NotNil(t, v)

		cnt, err := v.Count()
		assert.NoError(t, err)
		assert.Equal(t, size, cnt)

	})

	t.Run("Count", func(t *testing.T) {
		t.Parallel()

		v := builtin.NewVector(as...).Transient()

		cnt, err := v.Count()
		assert.NoError(t, err)
		assert.Equal(t, size, cnt)
	})

	t.Run("Conj", func(t *testing.T) {
		t.Parallel()

		v, err := builtin.EmptyVector.Transient().Conj(as...)
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

		v := builtin.EmptyVector.Transient()

		for i := 0; i < size; i++ {
			vPrime, err := v.Assoc(i, builtin.Int64(i))
			require.NoError(t, err, "Assoc() failed")
			require.NotNil(t, vPrime, "Assoc() returned a nil vector")

			val, err := v.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, builtin.Int64(i), val,
				"value recovered does not match associated value")

			val, err = vPrime.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, builtin.Int64(i), val,
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

		var v core.Vector = builtin.NewVector(as...).Transient()
		for i := range as {
			vPrime, err := v.Assoc(i, builtin.Nil{})
			assert.NoError(t, err)
			assert.NotNil(t, vPrime)

			val, err := v.EntryAt(i)
			assert.NoError(t, err)
			assert.Equal(t, builtin.Nil{}, val)

			val, err = vPrime.EntryAt(i)
			assert.NoError(t, err)
			assert.Equal(t, builtin.Nil{}, val)
		}
	})

	t.Run("Pop", func(t *testing.T) {
		t.Parallel()

		var v core.Vector = builtin.NewVector(as...).Transient()

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
		v := builtin.NewVector(as...).Transient()
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

		v := builtin.EmptyVector.Transient()
		v.Conj(builtin.Nil{})

		assert.NotEqual(t, builtin.EmptyVector, v,
			"derived transient mutated builtin.EmptyVector")

		assert.Equal(t, builtin.EmptyVector, builtin.EmptyVector.Transient().Persistent(),
			"persistent() ∘ Transient() ∘ persistent() ∘ builtin.EmptyVector != builtin.EmptyVector")
	})
}

func BenchmarkVector(b *testing.B) {
	for name, runner := range map[string]func(*testing.B){
		"PersistentVector_NoTransient":   runBenchmarks(b, new(persistentUnoptimized)),
		"PersistentVector_WithTransient": runBenchmarks(b, new(persistentOptimized)),
		"TransientVector_NoBatch":        runBenchmarks(b, new(transientUnbatched)),
		"TransientVector_Batched":        runBenchmarks(b, new(transientBatched)),
	} {
		b.Run(name, runner)
	}
}

func runBenchmarks(b *testing.B, s benchSuite) func(*testing.B) {
	return func(b *testing.B) {
		for name, runner := range map[string]func(*testing.B){
			"Conj": s.BenchmarkConj,
		} {
			b.Run(name, func(b *testing.B) {
				s.Setup(b)
				defer s.Teardown()

				b.ReportAllocs()
				b.ResetTimer()

				runner(b)
			})
		}
	}
}

type benchSuite interface {
	Setup(*testing.B)
	Teardown()
	BenchmarkConj(*testing.B)
}

type persistentUnoptimized struct{ vec core.Vector }

func (suite *persistentUnoptimized) Setup(b *testing.B) { suite.vec = builtin.EmptyVector }

func (suite *persistentUnoptimized) Teardown() {}

func (suite *persistentUnoptimized) BenchmarkConj(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// call Conj() one item at a time to avoid triggering transient optimization.
		suite.vec, _ = suite.vec.Conj(builtin.Int64(i))
	}
}

type persistentOptimized struct {
	vec   core.Vector
	items []core.Any
}

func (suite *persistentOptimized) Setup(b *testing.B) {
	suite.vec = builtin.EmptyVector

	suite.items = make([]core.Any, b.N)
	for i := 0; i < b.N; i++ {
		suite.items[i] = builtin.Int64(i)
	}
}

func (suite *persistentOptimized) Teardown() {}

func (suite *persistentOptimized) BenchmarkConj(b *testing.B) {
	suite.vec, _ = suite.vec.Conj(suite.items...)
}

type transientUnbatched struct{ vec core.Vector }

func (suite *transientUnbatched) Setup(b *testing.B) {
	suite.vec = builtin.EmptyVector.Transient()
}

func (suite *transientUnbatched) Teardown() {}

func (suite *transientUnbatched) BenchmarkConj(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// call Conj() one item at a time to avoid triggering transient optimization.
		suite.vec, _ = suite.vec.Conj(builtin.Int64(i))
	}
}

type transientBatched struct {
	vec   core.Vector
	items []core.Any
}

func (suite *transientBatched) Setup(b *testing.B) {
	suite.vec = builtin.EmptyVector.Transient()

	suite.items = make([]core.Any, b.N)
	for i := 0; i < b.N; i++ {
		suite.items[i] = builtin.Int64(i)
	}
}

func (suite *transientBatched) Teardown() {}

func (suite *transientBatched) BenchmarkConj(b *testing.B) {
	suite.vec, _ = suite.vec.Conj(suite.items...)
}
