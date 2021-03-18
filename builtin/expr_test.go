package builtin

import (
	"errors"
	"testing"

	"github.com/spy16/slurp/core"
	"github.com/stretchr/testify/assert"
)

var errUnknown = errors.New("failed")

func TestConstExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "SomeValue",
			expr: func() (core.Expr, core.Env) {
				return ConstExpr{Const: 10}, nil
			},
			want: 10,
		},
	})
}

func TestQuoteExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "SomeValue",
			expr: func() (core.Expr, core.Env) {
				return QuoteExpr{Form: 10}, nil
			},
			want: 10,
		},
	})
}

func TestDoExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "EmptyDo",
			expr: func() (core.Expr, core.Env) {
				return DoExpr(nil), nil
			},
			want: Nil{},
		},
		{
			title: "WithSingleExpr",
			expr: func() (core.Expr, core.Env) {
				return DoExpr{ConstExpr{Const: 10}}, nil
			},
			want: 10,
		},
		{
			title: "ExprEvalFail",
			expr: func() (core.Expr, core.Env) {
				return DoExpr{
					fakeExpr{Err: errUnknown},
				}, nil
			},
			wantErr: errUnknown,
		},
		{
			title: "MultipleExpr",
			expr: func() (core.Expr, core.Env) {
				return DoExpr{
					ConstExpr{Const: 10},
					ConstExpr{Const: "foo"},
				}, nil
			},
			want: "foo",
		},
	})
}

func TestDefExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "NoName",
			expr: func() (core.Expr, core.Env) {
				return DefExpr{}, NewEnv(nil)
			},
			wantErr: core.ErrInvalidName,
		},
		{
			title: "NilValue",
			expr: func() (core.Expr, core.Env) {
				return DefExpr{Name: "foo"}, NewEnv(nil)
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got core.Any, _ error, env core.Env) {
				v, err := env.Resolve("foo")
				assert.NoError(t, err)
				assert.Equal(t, Nil{}, v)
			},
		},
		{
			title: "ExprEvalErr",
			expr: func() (core.Expr, core.Env) {
				return DefExpr{
					Name:  "foo",
					Value: fakeExpr{Err: errUnknown},
				}, NewEnv(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "ExprValue",
			expr: func() (core.Expr, core.Env) {
				return DefExpr{
					Name:  "foo",
					Value: ConstExpr{Const: 10},
				}, NewEnv(nil)
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got core.Any, _ error, env core.Env) {
				v, err := env.Resolve("foo")
				assert.NoError(t, err)
				assert.Equal(t, 10, v)
			},
		},
	})
}

func TestIfExpr_Eval(t *testing.T) {
	t.Parallel()

	runExprTests(t, []exprTest{
		{
			title: "EmptyIf",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{}, NewEnv(nil)
			},
			want: Nil{},
		},
		{
			title: "WithoutThen",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
				}, NewEnv(nil)
			},
			want: Nil{},
		},
		{
			title: "WithoutElse",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
				}, NewEnv(nil)
			},
			want: Nil{},
		},
		{
			title: "Then",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
					Then: ConstExpr{Const: "hello"},
				}, NewEnv(nil)
			},
			want: "hello",
		},
		{
			title: "TestEvalErr",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: fakeExpr{Err: errUnknown},
				}, NewEnv(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "Else",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
					Else: ConstExpr{Const: "else-case"},
				}, NewEnv(nil)
			},
			want: "else-case",
		},
	})
}

func TestGoExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "WithError",
			expr: func() (core.Expr, core.Env) {
				return &GoExpr{
					Form: fakeExpr{Err: errUnknown},
				}, NewEnv(nil)
			},
			wantErr: nil,
		},
		{
			title: "WithSuccess",
			expr: func() (core.Expr, core.Env) {
				return &GoExpr{
					Form: fakeExpr{Res: 100},
				}, NewEnv(nil)
			},
			wantErr: nil,
		},
	})
}

func TestInvokeExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "TargetEvalErr",
			expr: func() (core.Expr, core.Env) {
				return &InvokeExpr{
					Target: fakeExpr{Err: errUnknown},
				}, NewEnv(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "NonInvokable",
			expr: func() (core.Expr, core.Env) {
				return &InvokeExpr{
					Target: ConstExpr{Const: 10},
				}, NewEnv(nil)
			},
			wantErr: core.ErrNotInvokable,
		},
		{
			title: "InvokeWithArgs",
			expr: func() (core.Expr, core.Env) {
				e := NewEnv(nil)
				return &InvokeExpr{
					Name: "foo",
					Target: ConstExpr{Const: fakeInvokable(func(args ...core.Any) (core.Any, error) {
						return args[0], nil
					})},
					Args: []core.Expr{
						ConstExpr{Const: 10},
					},
				}, e
			},
			want: 10,
		},
		{
			title: "ArgEvalErr",
			expr: func() (core.Expr, core.Env) {
				return &InvokeExpr{
					Target: ConstExpr{Const: fakeInvokable(nil)},
					Args: []core.Expr{
						fakeExpr{Err: errUnknown},
					},
				}, NewEnv(nil)
			},
			wantErr: errUnknown,
		},
	})
}

func TestVectorExpr_Eval(t *testing.T) {
	t.Run("ConstMembers", func(t *testing.T) {
		vec := NewVector(Keyword("foo"))

		any, err := (VectorExpr{
			Analyzer: &Analyzer{},
			Vector:   vec,
		}).Eval(nil)

		assert.NoError(t, err)
		assert.NotNil(t, any)
		assert.IsType(t, PersistentVector{}, any)
		assert.Equal(t, NewVector(Keyword("foo")), any)

		any, err = any.(core.Vector).EntryAt(0)
		assert.NoError(t, err)
		assert.Equal(t, Keyword("foo"), any)
	})

	t.Run("SymbolMember", func(t *testing.T) {
		env := NewEnv(map[string]core.Any{"foo": Keyword("foo")})
		vec := NewVector(Symbol("foo"))

		any, err := (VectorExpr{
			Analyzer: &Analyzer{},
			Vector:   vec,
		}).Eval(env)

		assert.NoError(t, err)
		assert.NotNil(t, any)
		assert.IsType(t, PersistentVector{}, any)
		assert.Equal(t, NewVector(Keyword("foo")), any)

		any, err = any.(core.Vector).EntryAt(0)
		assert.NoError(t, err)
		assert.Equal(t, Keyword("foo"), any)
	})

	t.Run("UnboundSymbolMember", func(t *testing.T) {
		vec := NewVector(Symbol("foo"), Symbol("bar"), Symbol("baz"))

		any, err := (VectorExpr{
			Analyzer: &Analyzer{},
			Vector:   vec,
		}).Eval(NewEnv(nil))

		assert.Error(t, err)
		assert.Nil(t, any)
	})
}

func runExprTests(t *testing.T, table []exprTest) {
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			expr, env := tt.expr()
			got, err := expr.Eval(env)
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, tt.wantErr),
					"wantErr=%#v\ngotErr=%#v", tt.wantErr, got)
				assert.Nil(t, got,
					"failed call to Eval returned non-nil value")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			if tt.assert != nil {
				tt.assert(t, got, err, env)
			}
		})
	}
}

type exprTest struct {
	title   string
	expr    func() (core.Expr, core.Env)
	want    core.Any
	wantErr error
	assert  func(t *testing.T, got core.Any, err error, env core.Env)
}

type fakeExpr struct {
	Res core.Any
	Err error
}

func (f fakeExpr) Eval(_ core.Env) (core.Any, error) { return f.Res, f.Err }

type fakeInvokable func(args ...core.Any) (core.Any, error)

func (f fakeInvokable) Invoke(args ...core.Any) (core.Any, error) { return f(args...) }
