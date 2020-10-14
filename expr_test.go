package slurp

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

var unknownErr = errors.New("failed")

func TestConstExpr_Eval(t *testing.T) {
	t.Parallel()
	runExprTests(t, []exprTest{
		{
			title: "SomeValue",
			expr: func() (Expr, *Env) {
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
			expr: func() (Expr, *Env) {
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
			expr: func() (Expr, *Env) {
				return DoExpr{Exprs: nil}, nil
			},
			want: Nil{},
		},
		{
			title: "WithSingleExpr",
			expr: func() (Expr, *Env) {
				return DoExpr{Exprs: []Expr{ConstExpr{Const: 10}}}, nil
			},
			want: 10,
		},
		{
			title: "ExprEvalFail",
			expr: func() (Expr, *Env) {
				return DoExpr{Exprs: []Expr{
					fakeExpr{Err: unknownErr},
				}}, nil
			},
			wantErr: unknownErr,
		},
		{
			title: "MultipleExpr",
			expr: func() (Expr, *Env) {
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
			expr: func() (Expr, *Env) {
				return DefExpr{}, New()
			},
			wantErr: ErrInvalidBindName,
		},
		{
			title: "NilValue",
			expr: func() (Expr, *Env) {
				e := New()
				return DefExpr{Name: "foo", Env: e}, e
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got Any, err error, env *Env) {
				assert(t, env.Resolve("foo") == Nil{},
					"expecting Nil{}, got %#v", env.Resolve("foo"))
			},
		},
		{
			title: "ExprEvalErr",
			expr: func() (Expr, *Env) {
				e := New()
				return DefExpr{
					Name:  "foo",
					Value: fakeExpr{Err: unknownErr},
					Env:   e,
				}, e
			},
			wantErr: unknownErr,
		},
		{
			title: "ExprValue",
			expr: func() (Expr, *Env) {
				e := New()
				return DefExpr{
					Name:  "foo",
					Value: ConstExpr{Const: 10},
					Env:   e,
				}, e
			},
			want: Symbol("foo"),
			assert: func(t *testing.T, got Any, err error, env *Env) {
				assert(t, env.Resolve("foo") == 10,
					"expecting 10, got %#v", env.Resolve("foo"))
			},
		},
	})
}

func TestIfExpr_Eval(t *testing.T) {
	t.Parallel()

	runExprTests(t, []exprTest{
		{
			title: "EmptyIf",
			expr: func() (Expr, *Env) {
				return IfExpr{}, New()
			},
			want: Nil{},
		},
		{
			title: "WithoutThen",
			expr: func() (Expr, *Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
				}, New()
			},
			want: Nil{},
		},
		{
			title: "WithoutElse",
			expr: func() (Expr, *Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
				}, New()
			},
			want: Nil{},
		},
		{
			title: "Then",
			expr: func() (Expr, *Env) {
				return IfExpr{
					Test: ConstExpr{Const: true},
					Then: ConstExpr{Const: "hello"},
				}, New()
			},
			want: "hello",
		},
		{
			title: "TestEvalErr",
			expr: func() (Expr, *Env) {
				return IfExpr{
					Test: fakeExpr{Err: unknownErr},
				}, New()
			},
			wantErr: unknownErr,
		},
		{
			title: "Else",
			expr: func() (Expr, *Env) {
				return IfExpr{
					Test: ConstExpr{Const: false},
					Else: ConstExpr{Const: "else-case"},
				}, New()
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
			expr: func() (Expr, *Env) {
				return &InvokeExpr{
					Target: fakeExpr{Err: unknownErr},
				}, New()
			},
			wantErr: unknownErr,
		},
		{
			title: "NonInvokable",
			expr: func() (Expr, *Env) {
				e := New()
				return &InvokeExpr{
					Env:    e,
					Target: ConstExpr{Const: 10},
				}, e
			},
			wantErr: ErrNotInvokable,
		},
		{
			title: "InvokeWithArgs",
			expr: func() (Expr, *Env) {
				e := New()
				return &InvokeExpr{
					Env:  e,
					Name: "foo",
					Target: ConstExpr{Const: fakeInvokable(func(env *Env, args ...Any) (Any, error) {
						got := env.stack[len(env.stack)-1].Name
						if got != "foo" {
							return nil, fmt.Errorf("stack name expected to be \"foo\", got \"%s\"", got)
						}
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
			expr: func() (Expr, *Env) {
				return &InvokeExpr{
					Target: ConstExpr{Const: fakeInvokable(nil)},
					Args: []Expr{
						fakeExpr{Err: unknownErr},
					},
				}, New()
			},
			wantErr: unknownErr,
		},
	})
}

func runExprTests(t *testing.T, table []exprTest) {
	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			expr, env := tt.expr()
			got, err := expr.Eval()
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
	expr    func() (Expr, *Env)
	want    Any
	wantErr error
	assert  func(t *testing.T, got Any, err error, env *Env)
}

type fakeExpr struct {
	Res Any
	Err error
}

func (f fakeExpr) Eval() (Any, error) { return f.Res, f.Err }

type fakeInvokable func(env *Env, args ...Any) (Any, error)

func (f fakeInvokable) Invoke(env *Env, args ...Any) (Any, error) {
	return f(env, args...)
}
