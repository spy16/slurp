package slurp_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp"
)

func TestCompare(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		a, b    slurp.Any
		want    int
		wantErr error
	}{
		{
			title: "ComparableEqualValues",
			a:     slurp.Int64(10),
			b:     slurp.Int64(10),
			want:  0,
		},
		{
			title: "ComparableUnEqualValues",
			a:     slurp.Int64(10),
			b:     slurp.Int64(100),
			want:  -1,
		},
		{
			title: "EqualWithEqualityProvider",
			a:     slurp.Symbol("foo"),
			b:     slurp.Symbol("foo"),
			want:  0,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       slurp.Symbol("foo"),
			b:       slurp.Symbol("bar"),
			want:    0,
			wantErr: slurp.ErrIncomparable,
		},
		{
			title:   "UnEqualWithEqualityProvider",
			a:       10,
			b:       slurp.Symbol("foo"),
			want:    0,
			wantErr: slurp.ErrIncomparable,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := slurp.Compare(tt.a, tt.b)
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

	got, err = slurp.Eq(slurp.Int64(10), slurp.Int64(10))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == true, "want=true got=%t", got)

	got, err = slurp.Eq(slurp.Int64(10), slurp.Int64(100))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == false, "want=false got=%t", got)

	got, err = slurp.Eq(slurp.Int64(10), slurp.Symbol("foo"))
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == false, "want=false got=%t", got)
}

func TestIsTruthy(t *testing.T) {
	assert(t, slurp.IsTruthy(true) == true, "want=true got=false")
	assert(t, slurp.IsTruthy(10) == true, "want=true got=false")
	assert(t, slurp.IsTruthy(nil) == false, "want=false got=true")
	assert(t, slurp.IsTruthy(false) == false, "want=false got=true")
}

func TestSeqString(t *testing.T) {
	seq := slurp.NewList(1, 2, slurp.Int64(3))
	want := "[1 2 3]"

	got, err := slurp.SeqString(seq, "[", "]", " ")
	assert(t, err == nil, "unexpected error: %#v", err)
	assert(t, got == want, "want='%s'\ngot='%s'", want, got)
}

func TestEvalAll(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		env     *slurp.Env
		vals    []slurp.Any
		want    []slurp.Any
		wantErr bool
	}{
		{
			title: "EmptyList",
			env:   slurp.New(),
			vals:  nil,
			want:  []slurp.Any{},
		},
		{
			title:   "EvalFails",
			env:     slurp.New(),
			vals:    []slurp.Any{slurp.Symbol("foo")},
			wantErr: true,
		},
		{
			title: "Success",
			env:   slurp.New(),
			vals:  []slurp.Any{slurp.String("foo"), slurp.Keyword("hello")},
			want:  []slurp.Any{slurp.String("foo"), slurp.Keyword("hello")},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := slurp.EvalAll(tt.env, tt.vals)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvalAll() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EvalAll() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
	if !cond {
		t.Errorf(msg, args...)
	}
}
