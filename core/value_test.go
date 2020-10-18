package core

import (
	"errors"
	"testing"
)

func TestCompare(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		a, b    Any
		want    int
		wantErr error
	}{
		{
			title: "ComparableEqualValues",
			a:     Int64(10),
			b:     Int64(10),
			want:  0,
		},
		{
			title: "ComparableUnEqualValues",
			a:     Int64(10),
			b:     Int64(100),
			want:  -1,
		},
		{
			title: "EqualWithEqualityProvider",
			a:     Symbol("foo"),
			b:     Symbol("foo"),
			want:  0,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       Symbol("foo"),
			b:       Symbol("bar"),
			want:    0,
			wantErr: ErrIncomparable,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       10,
			b:       Symbol("foo"),
			want:    0,
			wantErr: ErrIncomparable,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := Compare(tt.a, tt.b)
			if tt.wantErr != nil {
				assert(t, errors.Is(err, tt.wantErr),
					"wantErr=%#v\ngot=%#v", tt.wantErr, err)
				assert(t, got == 0, "want=0 got=%d", got)
			} else {
				assert(t, err == nil, "unexpected error: %#v", err)
				assert(t, tt.want == got, "want=%d got=%d", tt.want, got)
			}
		})
	}
}

func TestEq(t *testing.T) {
	var got bool
	var err error

	got, err = Eq(Int64(10), Int64(10))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == true, "want=true got=%t", got)

	got, err = Eq(Int64(10), Int64(100))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == false, "want=false got=%t", got)

	got, err = Eq(Int64(10), Symbol("foo"))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == false, "want=false got=%t", got)
}

func TestNil(t *testing.T) {
	n := Nil{}
	assert(t, n.String() == "nil", `want="nil" got="%s"`, n.String())
	testSExpr(t, n, "nil")
}

func TestInt64(t *testing.T) {
	v := Int64(100)
	assert(t, v.String() == "100", `want="100" got="%s"`, v.String())
	testSExpr(t, v, "100")
	testComp(t, v, 10, 0, ErrIncomparable)
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
	testComp(t, v, 10, 0, ErrIncomparable)
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

func testSExpr(t *testing.T, v SExpressable, want string) {
	s, err := v.SExpr()
	assert(t, err == nil, "unexpected err: %#v", err)
	assert(t, want == s, `want="%s" got="%s"`, want, s)
}

func testComp(t *testing.T, v Comparable, other Any, want int, wantErr error) {
	got, err := v.Comp(other)
	if wantErr != nil {
		assert(t, errors.Is(err, wantErr), "wantErr=%#v\ngotErr=%#v", wantErr, err)
	} else {
		assert(t, err == nil, "unexpected err: %#v", err)
	}
	assert(t, got == want, "want=%d got=%d", want, got)
}
