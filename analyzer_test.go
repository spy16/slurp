package slurp_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp"
)

type fakeFn struct{}

func (fakeFn) Invoke(_ *slurp.Env, _ ...slurp.Any) (slurp.Any, error) {
	return 100, nil
}

func TestBuiltinAnalyzer_Analyze(t *testing.T) {
	t.Parallel()

	hundredFunc := fakeFn{}
	e := slurp.New(slurp.WithGlobals(map[string]slurp.Any{
		"foo":     slurp.Keyword("hello"),
		"hundred": hundredFunc,
	}, nil))

	table := []struct {
		title   string
		env     *slurp.Env
		form    slurp.Any
		want    slurp.Expr
		wantErr error
	}{
		{
			title: "SpecialForm",
			form:  slurp.NewList(slurp.Symbol("foo")),
			want:  slurp.ConstExpr{Const: "foo"},
		},
		{
			title: "EmptySeq",
			form:  slurp.NewList(),
			want:  slurp.ConstExpr{Const: slurp.NewList()},
		},
		{
			title: "SymbolForm",
			env:   e,
			form:  slurp.Symbol("foo"),
			want:  slurp.ConstExpr{Const: slurp.Keyword("hello")},
		},
		{
			title: "Invokable",
			env:   e,
			form:  slurp.NewList(slurp.Symbol("hundred"), 1),
			want: slurp.InvokeExpr{
				Env:    e,
				Name:   "hundred",
				Target: slurp.ConstExpr{Const: hundredFunc},
				Args:   []slurp.Expr{slurp.ConstExpr{Const: 1}},
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			ba := &slurp.BuiltinAnalyzer{
				Specials: map[string]slurp.ParseSpecial{
					"foo": func(env *slurp.Env, args slurp.Seq) (slurp.Expr, error) {
						return slurp.ConstExpr{Const: "foo"}, nil
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
