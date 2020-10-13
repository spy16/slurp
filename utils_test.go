package slurp_test

import (
	"reflect"
	"testing"

	"github.com/spy16/slurp"
)

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
