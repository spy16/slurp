package slurp

import (
	"errors"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseFn(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title:   "NilArgs",
			env:     builtin.NewEnv(nil),
			args:    nil,
			wantErr: errors.New("nil argument sequence"),
		},
		{
			title:   "EmptyArgSeq",
			env:     builtin.NewEnv(nil),
			args:    builtin.NewList(),
			wantErr: core.ErrArity,
		},
		{
			title: "Arity0_Fn",
			env:   builtin.NewEnv(nil),
			args:  builtin.NewList(builtin.NewList()),
			assert: func(t *testing.T, got core.Expr, err error) {
				require.IsType(t, builtin.ConstExpr{}, got)
				ce := got.(builtin.ConstExpr)

				require.IsType(t, builtin.Fn{}, ce.Const)
				fn := ce.Const.(builtin.Fn)

				require.Empty(t, fn.Name, "unexpected name: %s", fn.Name)
				require.Empty(t, fn.Doc, "unexpected doc: %s", fn.Doc)
				require.NotNil(t, fn.Env, "Env not set on Fn")
				require.Len(t, fn.Funcs, 1, "expected only one method")
			},
		},
		{
			title: "Arity0_Fn_WithNameDoc",
			env:   builtin.NewEnv(nil),
			args: builtin.NewList(
				builtin.Symbol("foo"),
				builtin.String("hello"),
				builtin.NewList(),
			),
			assert: func(t *testing.T, got core.Expr, err error) {
				require.IsType(t, builtin.ConstExpr{}, got)
				ce := got.(builtin.ConstExpr)

				require.IsType(t, builtin.Fn{}, ce.Const)
				fn := ce.Const.(builtin.Fn)

				require.Equal(t, "foo", fn.Name, "unexpected name: %s", fn.Name)
				require.Equal(t, "hello", fn.Doc, "unexpected doc: %s", fn.Doc)
				require.NotNil(t, fn.Env, "Env not set on Fn")
				require.Len(t, fn.Funcs, 1, "expected only one method")
			},
		},
		{
			title: "Arity0_Fn_WithArgs",
			env:   builtin.NewEnv(nil),
			args: builtin.NewList(
				builtin.Symbol("foo"),
				builtin.NewList(builtin.Symbol("a"), builtin.Symbol("b")),
				builtin.NewList(builtin.Symbol("do"), 1, 2),
			),
			assert: func(t *testing.T, got core.Expr, err error) {
				require.IsType(t, builtin.ConstExpr{}, got)
				ce := got.(builtin.ConstExpr)

				require.IsType(t, builtin.Fn{}, ce.Const)
				fn := ce.Const.(builtin.Fn)

				require.Equal(t, "foo", fn.Name, "unexpected name: %s", fn.Name)
				require.Empty(t, fn.Doc, "unexpected doc: %s", fn.Doc)
				require.NotNil(t, fn.Env, "Env not set on Fn")
				require.Len(t, fn.Funcs, 1, "expected only one method")
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.env = builtin.NewEnv(nil)
			runSpecialTest(t, tt, parseFn)
		})
	}
}

func Test_parseDo(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title: "NilArgs",
			env:   builtin.NewEnv(nil),
			args:  nil,
			assert: func(t *testing.T, got core.Expr, err error) {
				assert.Equal(t, got, builtin.DoExpr(nil))
			},
		},
		{
			title: "SomeArgs",
			env:   builtin.NewEnv(nil),
			args:  builtin.NewList(1, 2),
			assert: func(t *testing.T, got core.Expr, err error) {
				want := builtin.DoExpr{
					builtin.ConstExpr{Const: 1},
					builtin.ConstExpr{Const: 2},
				}
				assert.Equal(t, got, want)
			},
		},
		{
			title:   "AnalyzeFail",
			env:     builtin.NewEnv(nil),
			args:    builtin.NewList(1, builtin.NewList(builtin.Symbol("def"))),
			wantErr: ErrParseSpecial,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			runSpecialTest(t, tt, parseDo)
		})
	}
}

func Test_parseDef(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title:   "NilArgs",
			args:    nil,
			wantErr: ErrParseSpecial,
		},
		{
			title:   "SomeArgs",
			args:    builtin.NewList(1, 2),
			wantErr: ErrParseSpecial,
		},
		{
			title: "Valid",
			args:  builtin.NewList(builtin.Symbol("foo"), 100),
			assert: func(t *testing.T, got core.Expr, err error) {
				want := builtin.DefExpr{
					Name:  "foo",
					Value: builtin.ConstExpr{Const: 100},
				}
				assert.Equal(t, want, got)
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.env = builtin.NewEnv(nil)
			runSpecialTest(t, tt, parseDef)
		})
	}
}

func Test_parseLet(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title:   "NilArgs",
			wantErr: ErrParseSpecial,
		},
		{
			title:   "EmptyArgs",
			args:    builtin.NewList(),
			wantErr: ErrParseSpecial,
		},
		{
			title: "ValidList",
			args: builtin.NewList(
				builtin.NewList(builtin.Symbol("x"), builtin.Int64(42)),
				builtin.Symbol("x"), // expr; return value of 'x'
			),
			assert: func(t *testing.T, got core.Expr, err error) {
				want := builtin.LetExpr{
					Names:  []string{"x"},
					Values: []core.Expr{builtin.ConstExpr{Const: builtin.Int64(42)}},
					Exprs:  builtin.DoExpr{builtin.ResolveExpr{Symbol: "x"}},
				}

				assert.Equal(t, want, got)
			},
		},
		{
			title: "ValidVector",
			args: builtin.NewList(
				builtin.NewVector(builtin.Symbol("x"), builtin.Int64(42)),
				builtin.Symbol("x"), // expr; return value of 'x'
			),
			assert: func(t *testing.T, got core.Expr, err error) {
				want := builtin.LetExpr{
					Names:  []string{"x"},
					Values: []core.Expr{builtin.ConstExpr{Const: builtin.Int64(42)}},
					Exprs:  builtin.DoExpr{builtin.ResolveExpr{Symbol: "x"}},
				}

				assert.Equal(t, want, got)
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.env = builtin.NewEnv(nil)
			runSpecialTest(t, tt, parseLet)
		})
	}
}

type specialTest struct {
	title   string
	env     core.Env
	args    core.Seq
	wantErr error
	assert  func(t *testing.T, got core.Expr, err error)
}

func runSpecialTest(t *testing.T, tt specialTest, parse builtin.ParseSpecial) {
	a := &builtin.Analyzer{
		Specials: map[string]builtin.ParseSpecial{
			"go":    parseGo,
			"do":    parseDo,
			"if":    parseIf,
			"fn":    parseFn,
			"def":   parseDef,
			"let":   parseLet,
			"macro": parseMacro,
			"quote": parseQuote,
		},
	}
	got, err := parse(a, tt.env, tt.args)
	if tt.wantErr != nil {
		// require.ErrorIs does not capture different errors with same value.
		if err.Error() != tt.wantErr.Error() {
			require.ErrorIs(t, err, tt.wantErr)
		}
	} else {
		require.NoError(t, err)
	}
	if tt.assert != nil {
		tt.assert(t, got, err)
	}
}
