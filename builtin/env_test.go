package builtin_test

import (
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/require"
)

func TestRoot(t *testing.T) {
	root := builtin.NewEnv(nil)
	child := root.Child("foo", nil)
	got := core.Root(child)
	require.NotEqual(t, child, got, "expecting root, got child")
	require.Equal(t, root, got, "returned env is not root")
	require.Equal(t, root.Name(), got.Name(), "wrong name for root env")
}

func Test_Env_Bind_Resolve(t *testing.T) {
	var v core.Any
	var err error

	env := builtin.NewEnv(map[string]core.Any{"foo": "bar"})

	err = env.Bind("v", 1000)
	require.NoError(t, err)

	err = env.Bind("", 1000)
	require.ErrorIs(t, err, core.ErrInvalidName)

	v, err = env.Resolve("foo")
	require.NoError(t, err)
	require.Equal(t, "bar", v)

	v, err = env.Resolve("non-existent")
	require.ErrorIs(t, err, core.ErrNotFound)
	require.Nil(t, v)
}
