package builtin_test

import (
	"errors"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinAnalyzer_Analyze(t *testing.T) {
	t.Parallel()

	hundredFunc := fakeFn{}
	e := builtin.NewEnv(builtin.WithNamespace("", map[string]core.Any{
		"foo":     builtin.Keyword("hello"),
		"hundred": hundredFunc,
	}))

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
				Target: builtin.ResolveExpr{Symbol: builtin.Symbol("hundred")},
				Args:   []core.Expr{builtin.ConstExpr{Const: 1}},
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			ba := &builtin.Analyzer{
				Specials: map[string]builtin.ParseSpecial{
					"foo": func(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
						return builtin.ConstExpr{Const: "foo"}, nil
					},
				},
			}

			got, err := ba.Analyze(tt.env, tt.form)
			if tt.wantErr != nil {
				if assert.Error(t, err) {
					assert.True(t, errors.Is(err, tt.wantErr),
						"error is not '%s'", tt.wantErr)
				}
			} else if assert.NoError(t, err) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestBultinAnalyzer_Analyze_Vector(t *testing.T) {
	t.Parallel()

	vec := builtin.NewVector(builtin.Symbol("foo"))

	var ba builtin.Analyzer
	expr, err := ba.Analyze(builtin.NewEnv(), vec)
	require.NoError(t, err)
	require.IsType(t, builtin.VectorExpr{}, expr)
	assert.Equal(t, vec, expr.(builtin.VectorExpr).Vector)
	assert.Equal(t, ba, expr.(builtin.VectorExpr).Analyzer)
}

type fakeFn struct{}

func (fakeFn) Invoke(_ ...core.Any) (core.Any, error) { return 100, nil }
