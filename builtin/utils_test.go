package builtin_test

import (
	"errors"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

func TestCompare(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		a, b    core.Any
		want    int
		wantErr error
	}{
		{
			title: "ComparableEqualValues",
			a:     builtin.Int64(10),
			b:     builtin.Int64(10),
			want:  0,
		},
		{
			title: "ComparableUnEqualValues",
			a:     builtin.Int64(10),
			b:     builtin.Int64(100),
			want:  -1,
		},
		{
			title: "EqualWithEqualityProvider",
			a:     builtin.Symbol("foo"),
			b:     builtin.Symbol("foo"),
			want:  0,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       builtin.Symbol("foo"),
			b:       builtin.Symbol("bar"),
			want:    0,
			wantErr: builtin.ErrIncomparable,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       10,
			b:       builtin.Symbol("foo"),
			want:    0,
			wantErr: builtin.ErrIncomparable,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := builtin.Compare(tt.a, tt.b)
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

	got, err = builtin.Eq(builtin.Int64(10), builtin.Int64(10))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == true, "want=true got=%t", got)

	got, err = builtin.Eq(builtin.Int64(10), builtin.Int64(100))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == false, "want=false got=%t", got)

	got, err = builtin.Eq(builtin.Int64(10), builtin.Symbol("foo"))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == false, "want=false got=%t", got)
}

func TestIsTruthy(t *testing.T) {
	assert(t, builtin.IsTruthy(true) == true, "want=true got=false")
	assert(t, builtin.IsTruthy(10) == true, "want=true got=false")
	assert(t, builtin.IsTruthy(nil) == false, "want=false got=true")
	assert(t, builtin.IsTruthy(false) == false, "want=false got=true")
}

func TestSeqString(t *testing.T) {
	seq := builtin.NewList(1, 2, builtin.Int64(3))
	want := "[1 2 3]"

	got, err := builtin.SeqString(seq, "[", "]", " ")
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == want, "want='%s'\ngot='%s'", want, got)
}
