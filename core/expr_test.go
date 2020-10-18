package core

import (
	"errors"
	"reflect"
	"testing"
)

func TestConstExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "SomeValue",
			expr: func() (Expr, Env) {
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
			expr: func() (Expr, Env) {
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
			expr: func() (Expr, Env) {
				return DoExpr{Exprs: nil}, nil
			},
			want: Nil{},
		},
		{
			title: "WithSingleExpr",
			expr: func() (Expr, Env) {
				return DoExpr{Exprs: []Expr{ConstExpr{Const: 10}}}, nil
			},
			want: 10,
		},
		{
			title: "ExprEvalFail",
			expr: func() (Expr, Env) {
				return DoExpr{Exprs: []Expr{
					fakeExpr{Err: errUnknown},
				}}, nil
			},
			wantErr: errUnknown,
		},
		{
			title: "MultipleExpr",
			expr: func() (Expr, Env) {
				return DoExpr{Exprs: []Expr{
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
			expr: func() (Expr, Env) {
				return DefExpr{}, New(nil)
			},
			wantErr: ErrInvalidName,
		},
		{
			title: "NilValue",
			expr: func() (Expr, Env) {
				return DefExpr{Name: "foo"}, New(nil)
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got Any, err error, env Env) {
				v, err := env.Resolve("foo")
				assert(t, err == nil, "unexpected error: %#v", err)
				assert(t, v == Nil{}, "expecting Nil{}, got %#v", v)
			},
		},
		{
			title: "ExprEvalErr",
			expr: func() (Expr, Env) {
				return DefExpr{
					Name:  "foo",
					Value: fakeExpr{Err: errUnknown},
				}, New(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "ExprValue",
			expr: func() (Expr, Env) {
				return DefExpr{
					Name:  "foo",
					Value: ConstExpr{Const: 10},
				}, New(nil)
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got Any, err error, env Env) {
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
			expr: func() (Expr, Env) {
				return IfExpr{}, New(nil)
			},
			want: Nil{},
		},
		{
			title: "WithoutThen",
			expr: func() (Expr, Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
				}, New(nil)
			},
			want: Nil{},
		},
		{
			title: "WithoutElse",
			expr: func() (Expr, Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
				}, New(nil)
			},
			want: Nil{},
		},
		{
			title: "Then",
			expr: func() (Expr, Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
					Then: ConstExpr{Const: "hello"},
				}, New(nil)
			},
			want: "hello",
		},
		{
			title: "TestEvalErr",
			expr: func() (Expr, Env) {
				return IfExpr{
					Test: fakeExpr{Err: errUnknown},
				}, New(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "Else",
			expr: func() (Expr, Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
					Else: ConstExpr{Const: "else-case"},
				}, New(nil)
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
			expr: func() (Expr, Env) {
				return &GoExpr{
					Form: fakeExpr{Err: errUnknown},
				}, New(nil)
			},
			wantErr: nil,
		},
		{
			title: "WithSuccess",
			expr: func() (Expr, Env) {
				return &GoExpr{
					Form: fakeExpr{Res: 100},
				}, New(nil)
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
			expr: func() (Expr, Env) {
				return &InvokeExpr{
					Target: fakeExpr{Err: errUnknown},
				}, New(nil)
			},
			wantErr: errUnknown,
		},
		{
			title: "NonInvokable",
			expr: func() (Expr, Env) {
				return &InvokeExpr{
					Target: ConstExpr{Const: 10},
				}, New(nil)
			},
			wantErr: ErrNotInvokable,
		},
		{
			title: "InvokeWithArgs",
			expr: func() (Expr, Env) {
				e := New(nil)
				return &InvokeExpr{
					Name: "foo",
					Target: ConstExpr{Const: fakeInvokable(func(args ...Any) (Any, error) {
						return args[0], nil
					})},
					Args: []Expr{
						ConstExpr{Const: 10},
					},
				}, e
			},
			want: 10,
		},
		{
			title: "ArgEvalErr",
			expr: func() (Expr, Env) {
				return &InvokeExpr{
					Target: ConstExpr{Const: fakeInvokable(nil)},
					Args: []Expr{
						fakeExpr{Err: errUnknown},
					},
				}, New(nil)
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
	expr    func() (Expr, Env)
	want    Any
	wantErr error
	assert  func(t *testing.T, got Any, err error, env Env)
}
