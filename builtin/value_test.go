package builtin

import (
	"errors"
	"testing"

	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
)

func TestNil(t *testing.T) {
	n := Nil{}
	assert.Equal(t, "nil", n.String())
	testSExpr(t, n, "nil")
}

func TestInt64(t *testing.T) {
	v := Int64(100)
	assert.Equal(t, "100", v.String())
	testSExpr(t, v, "100")
	testComp(t, v, 10, 0, core.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, Int64(1), 1, nil)
	testComp(t, v, Int64(10000), -1, nil)
}

func TestFloat64(t *testing.T) {
	assert.Equal(t, "1.000000e+19", Float64(1e19).String())

	v := Float64(100)
	assert.Equal(t, "100.000000", v.String())
	testSExpr(t, v, "100.000000")
	testComp(t, v, 10, 0, core.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, Float64(1), 1, nil)
	testComp(t, v, Float64(10000), -1, nil)
}

func TestIsTruthy(t *testing.T) {
	assert.True(t, IsTruthy(true))
	assert.True(t, IsTruthy(10))
	assert.False(t, IsTruthy(nil))
	assert.False(t, IsTruthy(false))
}

func testSExpr(t *testing.T, v core.SExpressable, want string) {
	s, err := v.SExpr()
	assert.NoError(t, err)
	assert.Equal(t, want, s)
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

// func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
// 	if !cond {
// 		t.Errorf(msg, args...)
// 	}
// }
