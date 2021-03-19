package builtin_test

import (
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/require"
)

func TestRootEnv(t *testing.T) {
	t.Parallel()

	t.Run("Child", func(t *testing.T) {
		t.Parallel()

		root := builtin.NewEnv()
		require.NotEmpty(t, root.Name(), "namespace must not be empty")

		child := root.Child("foo", nil)
		got := core.Root(child)
		require.NotEqual(t, child, got, "expecting root, got child")
		require.Equal(t, root.Name(), got.Name(), "wrong name for root env")
	})

	t.Run("BindResolve", func(t *testing.T) {
		t.Parallel()

		var v core.Any
		var err error

		env := builtin.NewEnv(builtin.WithNamespace("", map[string]core.Any{"foo": "bar"}))
		require.Equal(t, "<main>", env.Name())

		err = env.Scope().Bind(builtin.Symbol("v"), 1000)
		require.NoError(t, err)

		err = env.Scope().Bind(builtin.Symbol(""), 1000)
		require.ErrorIs(t, err, core.ErrInvalidName)

		v, err = env.Scope().Resolve(builtin.Symbol("foo"))
		require.NoError(t, err)
		require.Equal(t, "bar", v)

		v, err = env.Scope().Resolve(builtin.Symbol("non-existent"))
		require.ErrorIs(t, err, core.ErrNotFound)
		require.Nil(t, v)
	})

	t.Run("WithNamespace", func(t *testing.T) {
		t.Parallel()

		var v core.Any
		var err error

		root := builtin.NewEnv(builtin.WithNamespace("", map[string]core.Any{"foo": "bar"}))
		test := root.WithNamespace("test")

		t.Run("Name", func(t *testing.T) {
			// ensure root and test have the expected names
			require.Equal(t, core.Namespace("test"), test.Namespace())
			require.NotEqual(t, core.Namespace("test"), root.Namespace())
		})

		t.Run("Resolve", func(t *testing.T) {
			// ensure root can still access its bound variables.
			v, err = root.Scope().Resolve(builtin.Symbol("foo"))
			require.NoError(t, err)
			require.Equal(t, "bar", v)

			// ensure test cannot access the default namespace
			v, err = test.Scope().Resolve(builtin.Symbol("foo"))
			require.ErrorIs(t, err, core.ErrNotFound)
			require.Nil(t, v)

			// ensure fully-qualified symbol names can be resolved
			v, err = test.Scope().Resolve(builtin.Symbol(".main.foo"))
			require.NoError(t, err)
			require.Equal(t, "bar", v)
		})

	})
}
