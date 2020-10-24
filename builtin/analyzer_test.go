package builtin_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

func TestBuiltinAnalyzer_Analyze(t *testing.T) {
	t.Parallel()

	hundredFunc := fakeFn{}
	e := builtin.New(map[string]core.Any{
		"foo":     builtin.Keyword("hello"),
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
			form:  builtin.NewList(builtin.Symbol("foo")),
			want:  builtin.ConstExpr{Const: "foo"},
		},
		{
			title: "EmptySeq",
			form:  builtin.NewList(),
			want:  builtin.ConstExpr{Const: builtin.NewList()},
		},
		{
			title: "SymbolForm",
			env:   e,
			form:  builtin.Symbol("foo"),
			want:  builtin.ResolveExpr{Symbol: builtin.Symbol("foo")},
		},
		{
			title: "Invokable",
			env:   e,
			form:  builtin.NewList(builtin.Symbol("hundred"), 1),
			want: builtin.InvokeExpr{
				Name:   "hundred",
				Target: builtin.ResolveExpr{Symbol: "hundred"},
				Args:   []core.Expr{builtin.ConstExpr{Const: 1}},
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			ba := &builtin.Analyzer{
				Specials: map[string]builtin.ParseSpecial{
					"foo": func(a core.Analyzer, env core.Env, args builtin.Seq) (core.Expr, error) {
						return builtin.ConstExpr{Const: "foo"}, nil
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
