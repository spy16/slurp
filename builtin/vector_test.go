package builtin

import (
	"testing"

	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const size = 4096

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

func TestEmptyVector(t *testing.T) {
	t.Parallel()

	require.NotZero(t, EmptyVector,
		"zero-value empty vector is invalid (shift is missing)")

	t.Run("SExpr", func(t *testing.T) {
		testSExpr(t, EmptyVector, "[]")
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
}

func TestPersistentVector(t *testing.T) {
	t.Parallel()

	var v core.Vector = EmptyVector

	t.Run("SExpr", func(t *testing.T) {
		testSExpr(t, NewVector(Int64(0), Keyword("keyword"), String("string")),
			"[0 :keyword \"string\"]")
	})

	t.Run("Append", func(t *testing.T) {
		// N.B.:  do not run in parallel.  `v` must be constructed by prior tests.
		var err error
		for i := 0; i < size; i++ {
			v, err = v.Assoc(i, Int64(i))
			require.NoError(t, err, "Assoc() failed")
			require.NotNil(t, v, "Assoc() returned a nil vector")

			val, err := v.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, Int64(i), val,
				"value recovered does not match associated value")

			cnt, err := v.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, i+1, cnt, "Count() returned incorrect value '%d'", cnt)
		}
	})

	t.Run("Replace", func(t *testing.T) {
		// N.B.:  do not run in parallel.  `v` must be constructed by prior tests.

		for _, tt := range []struct {
			desc    string
			idx     int
			wantErr error
		}{
			{
				idx:  0,
				desc: "first",
			},
			{
				idx:  1024,
				desc: "branch-2",
			},
			{
				idx:  4094,
				desc: "tail",
			},
			{
				idx:  size,
				desc: "cons",
			},
			{
				idx:     1000000,
				desc:    "out of bounds",
				wantErr: ErrIndexOutOfBounds,
			},
		} {
			t.Run(tt.desc, func(t *testing.T) {
				t.Parallel()

				// shadow v & err in order to make this thread-safe
				v, err := v.Assoc(tt.idx, Nil{})
				if tt.wantErr == nil {
					assert.NoError(t, err)

					val, err := v.EntryAt(tt.idx)
					assert.NoError(t, err)
					assert.Equal(t, Nil{}, val)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
					assert.Nil(t, v)
				}
			})
		}
	})

	t.Run("Pop", func(t *testing.T) {
		// N.B.:  do not run in parallel.  `v` must be constructed by prior tests.

		t.Skip("Pop() NOT IMPLEMENTED.  Skipping...")

		cnt, err := v.Count()
		require.NoError(t, err, "test precondition failed")
		require.Equal(t, size, cnt)

		for i := size - 1; i >= 0; i-- {
			v, err = v.Pop()
			require.NoError(t, err)
			require.NotNil(t, v)

			cnt, err = v.Count()
			require.NoError(t, err)
			require.Equal(t, i, cnt)
		}
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

	var b VectorBuilder
	for i := 0; i < size; i++ {
		b.Conj(Int64(i))
	}

	v := b.Vector()
	assert.NotZero(t, v)

	n, err := b.Vector().Count()
	assert.NoError(t, err)
	assert.Equal(t, size, n)
	assert.Panics(t, func() { b.Conj(Nil{}) })
}
