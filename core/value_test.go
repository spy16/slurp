package core_test

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
			a:     fakeComparable{10},
			b:     fakeComparable{10},
			want:  0,
		},
		{
			title: "ComparableUnEqualValues",
			a:     fakeComparable{10},
			b:     fakeComparable{100},
			want:  -1,
		},
		{
			title: "EqualWithEqualityProvider",
			a:     fakeEqProvider{true},
			b:     fakeEqProvider{true},
			want:  0,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       fakeEqProvider{true},
			b:       fakeEqProvider{false},
			want:    0,
			wantErr: core.ErrIncomparable,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       fakeComparable{10},
			b:       fakeEqProvider{false},
			want:    0,
			wantErr: core.ErrIncomparable,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := core.Compare(tt.a, tt.b)
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
	t.Parallel()

	table := []struct {
		title   string
		a, b    core.Any
		want    bool
		wantErr error
	}{
		{
			title:   "Same_Eq_Providers",
			a:       fakeEqProvider{true},
			b:       fakeEqProvider{true},
			want:    true,
			wantErr: nil,
		},
		{
			title:   "Same_Comparables",
			a:       fakeComparable{10},
			b:       fakeComparable{10},
			want:    true,
			wantErr: nil,
		},
		{
			title:   "Diff_Comparables",
			a:       fakeComparable{10},
			b:       fakeComparable{100},
			want:    false,
			wantErr: nil,
		},
		{
			title:   "Eq_Provider_With_Comparable",
			a:       fakeEqProvider{true},
			b:       fakeComparable{10},
			want:    false,
			wantErr: nil,
		},
		{
			title:   "NilSeqs",
			a:       core.Seq(nil),
			b:       core.Seq(nil),
			want:    true,
			wantErr: nil,
		},
		{
			title:   "SeqEqual",
			a:       builtin.NewList(builtin.Int64(1), builtin.Symbol("foo")),
			b:       builtin.NewList(builtin.Int64(1), builtin.Symbol("foo")),
			want:    true,
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := core.Eq(tt.a, tt.b)
			if tt.wantErr != nil {
				assert(t, errors.Is(tt.wantErr, err),
					"wantErr=%#v\ngotErr=%#v", tt.wantErr, err)
			} else {
				assert(t, err == nil, "unexpected err: %#v", err)
			}
			assert(t, tt.want == got, "want=%t, got=%t", tt.want, got)
		})
	}
}

type fakeEqProvider struct {
	eq bool
}

func (fe fakeEqProvider) Equals(other core.Any) (bool, error) {
	feo, ok := other.(fakeEqProvider)
	if !ok {
		return false, nil
	}
	return feo.eq == fe.eq, nil
}

type fakeComparable struct {
	value int
}

func (fc fakeComparable) Comp(other core.Any) (int, error) {
	fco, ok := other.(fakeComparable)
	if !ok {
		return 0, core.ErrIncomparable
	}

	switch {
	case fc.value < fco.value:
		return -1, nil

	case fc.value > fco.value:
		return 1, nil

	default:
		return 0, nil
	}
}

func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
	if !cond {
		t.Errorf(msg, args...)
	}
}
