package builtin_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNil(t *testing.T) {
	n := builtin.Nil{}
	assert.Equal(t, "nil", n.String())
	assertEq(t, builtin.Nil{}, n)
}

func TestInt64(t *testing.T) {
	v := builtin.Int64(100)
	assert.Equal(t, "100", v.String())
	assertEq(t, builtin.Int64(100), v)
	testComp(t, v, 10, 0, core.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, builtin.Int64(1), 1, nil)
	testComp(t, v, builtin.Int64(10000), -1, nil)
}

func TestFloat64(t *testing.T) {
	assert.Equal(t, "1.000000e+19", builtin.Float64(1e19).String())

	v := builtin.Float64(100)
	assert.Equal(t, "100.000000", v.String())
	assertEq(t, builtin.Float64(100), v)
	testComp(t, v, 10, 0, core.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, builtin.Float64(1), 1, nil)
	testComp(t, v, builtin.Float64(10000), -1, nil)
}

func TestIsTruthy(t *testing.T) {
	assert.True(t, builtin.IsTruthy(true))
	assert.True(t, builtin.IsTruthy(10))
	assert.False(t, builtin.IsTruthy(nil))
	assert.False(t, builtin.IsTruthy(false))
}

func TestFormat(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		any          core.Any
		pretty, sxpr string
	}{
		{
			any:    builtin.Nil{},
			pretty: "nil",
			sxpr:   "nil",
		},
		{
			any:    builtin.Int64(0),
			pretty: "0",
			sxpr:   "0",
		},
		{
			any:    builtin.Float64(0),
			pretty: "0.",
			sxpr:   "0.",
		},
		{
			any:    builtin.Float64(1e16),
			pretty: "1e16",
			sxpr:   "1e16",
		},
		{
			any:    builtin.Float64(-.2),
			pretty: "-.2",
			sxpr:   "-.2",
		},
		{
			any:    builtin.Bool(false),
			pretty: "false",
			sxpr:   "false",
		},
		{
			any:    builtin.Bool(true),
			pretty: "true",
			sxpr:   "true",
		},
		{
			any:    builtin.Char('🧠'),
			pretty: "\\🧠",
			sxpr:   "\\🧠",
		},
		{
			any:    builtin.String("foo"),
			pretty: "\"foo\"",
			sxpr:   "\"foo\"",
		},
		{
			any:    builtin.Symbol("foo"),
			pretty: "foo",
			sxpr:   "foo",
		},
	} {
		t.Run(reflect.TypeOf(tt.any).String(), func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.pretty, fmt.Sprintf("%#s", tt.any),
				"invalid pretty-print output")

			assert.Equal(t, tt.sxpr, fmt.Sprintf("%s", tt.any),
				"invalid symbolic expression")
		})
	}
}

func assertEq(t *testing.T, want, got core.Any, msgAndArgs ...interface{}) {
	ok, err := core.Eq(want, got)
	require.NoError(t, err, msgAndArgs...)
	require.True(t, ok, msgAndArgs...)
}

func testComp(t *testing.T, v core.Comparable, other core.Any, want int, wantErr error) {
	got, err := v.Comp(other)
	if wantErr != nil {
		assert.True(t, errors.Is(err, wantErr), "wantErr=%#v\ngotErr=%#v", wantErr, err)
	} else {
		assert.NoError(t, err)
	}

	assert.Equal(t, want, got)
}
