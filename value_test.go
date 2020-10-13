package slurp_test

import (
	"errors"
	"testing"

	"github.com/spy16/slurp"
)

func TestCons(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		first   slurp.Any
		rest    slurp.Seq
		items   []slurp.Any
		wantSz  int
		wantErr error
	}{
		{
			title:  "NilSeq",
			first:  slurp.Int64(100),
			rest:   nil,
			wantSz: 1,
		},
		{
			title:  "ZeroLenSeq",
			first:  slurp.Int64(100),
			rest:   slurp.NewList(),
			wantSz: 1,
		},
		{
			title:  "OneItemSeq",
			first:  slurp.Int64(100),
			rest:   slurp.NewList(1),
			wantSz: 2,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			seq, err := slurp.Cons(tt.first, tt.rest)
			if tt.wantErr != nil {
				assert(t, errors.Is(err, tt.wantErr),
					"wantErr=%#v\ngot=%#v", tt.wantErr, err)
				assert(t, seq == nil, "want=nil got=%#v", seq)
			} else {
				count, err := seq.Count()
				assert(t, err == nil, "unexpected err: %#v", err)
				assert(t, count == tt.wantSz, "want=%d got=%d", tt.wantSz, count)
			}
		})
	}
}

func TestNil(t *testing.T) {
	n := slurp.Nil{}
	assert(t, n.String() == "nil", `want="nil" got="%s"`, n.String())
	testSExpr(t, n, "nil")
}

func TestInt64(t *testing.T) {
	v := slurp.Int64(100)
	assert(t, v.String() == "100", `want="100" got="%s"`, v.String())
	testSExpr(t, v, "100")
	testComp(t, v, 10, 0, slurp.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, slurp.Int64(1), 1, nil)
	testComp(t, v, slurp.Int64(10000), -1, nil)
}

func TestFloat64(t *testing.T) {
	vLarge := slurp.Float64(1e19).String()
	assert(t, vLarge == "1.000000e+19", `want="1.000000e+19" got="%s"`, vLarge)

	v := slurp.Float64(100)
	assert(t, v.String() == "100.000000", `want="100.000000" got="%s"`, v.String())
	testSExpr(t, v, "100.000000")
	testComp(t, v, 10, 0, slurp.ErrIncomparable)
	testComp(t, v, v, 0, nil)
	testComp(t, v, slurp.Float64(1), 1, nil)
	testComp(t, v, slurp.Float64(10000), -1, nil)
}

func testSExpr(t *testing.T, v slurp.SExpressable, want string) {
	s, err := v.SExpr()
	assert(t, err == nil, "unexpected err: %#v", err)
	assert(t, want == s, `want="%s" got="%s"`, want, s)
}

func testComp(t *testing.T, v slurp.Comparable, other slurp.Any, want int, wantErr error) {
	got, err := v.Comp(other)
	if wantErr != nil {
		assert(t, errors.Is(err, wantErr), "wantErr=%#v\ngotErr=%#v", wantErr, err)
	} else {
		assert(t, err == nil, "unexpected err: %#v", err)
	}
	assert(t, got == want, "want=%d got=%d", want, got)
}
