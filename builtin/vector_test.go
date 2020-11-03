package builtin

import (
	"testing"

	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	t.Run("SExpr", func(t *testing.T) {
		t.Parallel()

		vec := NewVector(Int64(0), Keyword("keyword"), String("string"))
		testSExpr(t, vec, "[0 :keyword \"string\"]")
	})

	t.Run("Assoc", func(t *testing.T) {
		t.Parallel()

		var err error
		var val core.Any
		var v core.Vector = EmptyVector
		for i := 0; i < 4095; i++ {
			v, err = v.Assoc(i, Int64(i))
			require.NoError(t, err, "Assoc() failed")
			require.NotNil(t, v, "Assoc() returned a nil vector")

			val, err = v.EntryAt(i)
			require.NoError(t, err, "EntryAt() failed")
			require.NotNil(t, val, "EntryAt() returned a nil value")
			require.Equal(t, Int64(i), val,
				"value recovered does not match associated value")

			cnt, err := v.Count()
			require.NoError(t, err, "Count() failed")
			require.Equal(t, i+1, cnt, "Count() returned incorrect value '%d'", cnt)
		}

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

					val, err = v.EntryAt(tt.idx)
					assert.NoError(t, err)
					assert.Equal(t, Nil{}, val)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
					assert.Nil(t, v)
				}
			})
		}
	})
}
