package builtin

import (
	"errors"
	"testing"

	"github.com/spy16/slurp/core"
)

func TestNil(t *testing.T) {
	n := Nil{}
	assert(t, n.String() == "nil", `want="nil" got="%s"`, n.String())
	testSExpr(t, n, "nil")
}

func TestInt64(t *testing.T) {
	v := Int64(100)
	assert(t, v.String() == "100", `want="100" got="%s"`, v.String())
	testSExpr(t, v, "100")
	testComp(t, v, 10, 0, core.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, Int64(1), 1, nil)
	testComp(t, v, Int64(10000), -1, nil)
}

func TestFloat64(t *testing.T) {
	vLarge := Float64(1e19).String()
	assert(t, vLarge == "1.000000e+19", `want="1.000000e+19" got="%s"`, vLarge)

	v := Float64(100)
	assert(t, v.String() == "100.000000", `want="100.000000" got="%s"`, v.String())
	testSExpr(t, v, "100.000000")
	testComp(t, v, 10, 0, core.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, Float64(1), 1, nil)
	testComp(t, v, Float64(10000), -1, nil)
}

func TestIsTruthy(t *testing.T) {
	assert(t, IsTruthy(true) == true, "want=true got=false")
	assert(t, IsTruthy(10) == true, "want=true got=false")
	assert(t, IsTruthy(nil) == false, "want=false got=true")
	assert(t, IsTruthy(false) == false, "want=false got=true")
}

func testSExpr(t *testing.T, v core.SExpressable, want string) {
	s, err := v.SExpr()
	assert(t, err == nil, "unexpected err: %#v", err)
	assert(t, want == s, `want="%s" got="%s"`, want, s)
}

func testComp(t *testing.T, v core.Comparable, other core.Any, want int, wantErr error) {
	got, err := v.Comp(other)
	if wantErr != nil {
		assert(t, errors.Is(err, wantErr), "wantErr=%#v\ngotErr=%#v", wantErr, err)
	} else {
		assert(t, err == nil, "unexpected err: %#v", err)
	}
	assert(t, got == want, "want=%d got=%d", want, got)
}

func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
	if !cond {
		t.Errorf(msg, args...)
	}
}
