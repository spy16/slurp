package slurp_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/core"
)

func TestBuiltinAnalyzer_Analyze(t *testing.T) {
	t.Parallel()

	hundredFunc := fakeFn{}
	e := core.New(map[string]core.Any{
		"foo":     core.Keyword("hello"),
		"hundred": hundredFunc,
	})

	table := []struct {
		title   string
		env     core.Env
		form    core.Any
		want    core.Expr
		wantErr error
	}{
		{
			title: "SpecialForm",
			form:  core.NewList(core.Symbol("foo")),
			want:  core.ConstExpr{Const: "foo"},
		},
		{
			title: "EmptySeq",
			form:  core.NewList(),
			want:  core.ConstExpr{Const: core.NewList()},
		},
		{
			title: "SymbolForm",
			env:   e,
			form:  core.Symbol("foo"),
			want:  core.ResolveExpr{Symbol: core.Symbol("foo")},
		},
		{
			title: "Invokable",
			env:   e,
			form:  core.NewList(core.Symbol("hundred"), 1),
			want: core.InvokeExpr{
				Name:   "hundred",
				Target: core.ResolveExpr{Symbol: "hundred"},
				Args:   []core.Expr{core.ConstExpr{Const: 1}},
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			ba := &slurp.Analyzer{
				Specials: map[string]slurp.ParseSpecial{
					"foo": func(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
						return core.ConstExpr{Const: "foo"}, nil
					},
				},
			}
			got, err := ba.Analyze(tt.env, tt.form)
			if tt.wantErr != nil {
				assert(t, errors.Is(err, tt.wantErr),
					"\nwantErr=%#v\ngot=%#v", tt.wantErr, got)
				assert(t, got == nil, "want nil, got %#v", got)
			} else {
				assert(t, err == nil, "unexpected err: %#v", err)
				assert(t, reflect.DeepEqual(tt.want, got),
					"\nwant=%#v\ngot=%#v", tt.want, got)
			}
		})
	}
}

func assert(t testInstance, cond bool, msg string, args ...interface{}) {
	if !cond {
		t.Errorf(msg, args...)
	}
}

type testInstance interface {
	Errorf(msg string, args ...interface{})
}

type fakeFn struct{}

func (fakeFn) Invoke(_ ...core.Any) (core.Any, error) {
	return 100, nil
}
