package builtin

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp/core"
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
				return DoExpr{Exprs: nil}, nil
			},
			want: Nil{},
		},
		{
			title: "WithSingleExpr",
			expr: func() (core.Expr, core.Env) {
				return DoExpr{Exprs: []core.Expr{ConstExpr{Const: 10}}}, nil
			},
			want: 10,
		},
		{
			title: "ExprEvalFail",
			expr: func() (core.Expr, core.Env) {
				return DoExpr{Exprs: []core.Expr{
					fakeExpr{Err: errUnknown},
				}}, nil
			},
			wantErr: errUnknown,
		},
		{
			title: "MultipleExpr",
			expr: func() (core.Expr, core.Env) {
				return DoExpr{Exprs: []core.Expr{
					ConstExpr{Const: 10},
					ConstExpr{Const: "foo"},
				}}, nil
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
				return DefExpr{}, core.New(nil)
			},
			wantErr: core.ErrInvalidBindName,
		},
		{
			title: "NilValue",
			expr: func() (core.Expr, core.Env) {
				return DefExpr{Name: "foo"}, core.New(nil)
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got core.Any, err error, env core.Env) {
				v, err := env.Resolve("foo")
				assert(t, err == nil, "unexpected error: %#v", err)
				assert(t, v == Nil{}, "expecting Nil{}, got %#v", v)
			},
		},
		{
			title: "ExprEvalErr",
			expr: func() (core.Expr, core.Env) {
				return DefExpr{
					Name:  "foo",
					Value: fakeExpr{Err: errUnknown},
				}, core.New(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "ExprValue",
			expr: func() (core.Expr, core.Env) {
				return DefExpr{
					Name:  "foo",
					Value: ConstExpr{Const: 10},
				}, core.New(nil)
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got core.Any, err error, env core.Env) {
				v, err := env.Resolve("foo")
				assert(t, err == nil, "unexpected error: %#v", err)
				assert(t, v == 10, "expecting 10, got %#v", v)
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
				return IfExpr{}, core.New(nil)
			},
			want: Nil{},
		},
		{
			title: "WithoutThen",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
				}, core.New(nil)
			},
			want: Nil{},
		},
		{
			title: "WithoutElse",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
				}, core.New(nil)
			},
			want: Nil{},
		},
		{
			title: "Then",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
					Then: ConstExpr{Const: "hello"},
				}, core.New(nil)
			},
			want: "hello",
		},
		{
			title: "TestEvalErr",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: fakeExpr{Err: errUnknown},
				}, core.New(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "Else",
			expr: func() (core.Expr, core.Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
					Else: ConstExpr{Const: "else-case"},
				}, core.New(nil)
			},
			want: "else-case",
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
				}, core.New(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "NonInvokable",
			expr: func() (core.Expr, core.Env) {
				return &InvokeExpr{
					Target: ConstExpr{Const: 10},
				}, core.New(nil)
			},
			wantErr: ErrNotInvokable,
		},
		{
			title: "InvokeWithArgs",
			expr: func() (core.Expr, core.Env) {
				e := core.New(nil)
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
				}, core.New(nil)
			},
			wantErr: errUnknown,
		},
	})
}

func runExprTests(t *testing.T, table []exprTest) {
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			expr, env := tt.expr()
			got, err := expr.Eval(env)
			if tt.wantErr != nil {
				assert(t, errors.Is(err, tt.wantErr),
					"wantErr=%#v\ngotErr=%#v", tt.wantErr, got)
				assert(t, got == nil, "expected nil, got %#v", got)
			} else {
				assert(t, err == nil, "unexpected err: %#v", err)
				assert(t, reflect.DeepEqual(tt.want, got),
					"want=%#v\n%#v", tt.want, got)
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

func (f fakeInvokable) Invoke(args ...core.Any) (core.Any, error) {
	return f(args...)
}

func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
	if !cond {
		t.Errorf(msg, args...)
	}
}
