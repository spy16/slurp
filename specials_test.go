package slurp

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

func Test_parseFn(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title:   "NilArgs",
			env:     core.New(nil),
			args:    nil,
			wantErr: errors.New("nil argument sequence"),
		},
		{
			title:   "EmptyArgSeq",
			env:     core.New(nil),
			args:    builtin.NewList(),
			wantErr: core.ErrArity,
		},
		{
			title: "Arity0_Fn",
			env:   core.New(nil),
			args:  builtin.NewList(builtin.NewList()),
			assert: func(t *testing.T, got core.Expr, err error) {
				ce, ok := got.(builtin.ConstExpr)
				assert(t, ok, "expected ConstExpr, got %#v", got)
				assert(t, ce.Const != nil, "expected Const to be not nil")

				fn, ok := ce.Const.(builtin.Fn)
				assert(t, ok, "expected Const to be Fn, got %#v", ce.Const)
				assert(t, fn.Name == "", "unexpected name: %s", fn.Name)
				assert(t, fn.Doc == "", "unexpected doc: %s", fn.Doc)
				assert(t, fn.Env != nil, "Env not set on Fn")
				assert(t, len(fn.Funcs) == 1, "expected only one method, got %d", len(fn.Funcs))
			},
		},
		{
			title: "Arity0_Fn_WithNameDoc",
			env:   core.New(nil),
			args: builtin.NewList(
				builtin.Symbol("foo"),
				builtin.String("hello"),
				builtin.NewList(),
			),
			assert: func(t *testing.T, got core.Expr, err error) {
				ce, ok := got.(builtin.ConstExpr)
				assert(t, ok, "expected ConstExpr, got %#v", got)
				assert(t, ce.Const != nil, "expected Const to be not nil")

				fn, ok := ce.Const.(builtin.Fn)
				assert(t, ok, "expected Const to be Fn, got %#v", ce.Const)
				assert(t, fn.Name == "foo", "unexpected name: %s", fn.Name)
				assert(t, fn.Doc == "hello", "unexpected doc: %s", fn.Doc)
				assert(t, fn.Env != nil, "Env not set on Fn")
				assert(t, len(fn.Funcs) == 1, "expected only one method, got %d", len(fn.Funcs))
			},
		},
		{
			title: "Arity0_Fn_WithArgs",
			env:   core.New(nil),
			args: builtin.NewList(
				builtin.Symbol("foo"),
				builtin.NewList(builtin.Symbol("a"), builtin.Symbol("b")),
				builtin.NewList(builtin.Symbol("do"), 1, 2),
			),
			assert: func(t *testing.T, got core.Expr, err error) {
				ce, ok := got.(builtin.ConstExpr)
				assert(t, ok, "expected ConstExpr, got %#v", got)
				assert(t, ce.Const != nil, "expected Const to be not nil")

				fn, ok := ce.Const.(builtin.Fn)
				assert(t, ok, "expected Const to be Fn, got %#v", ce.Const)
				assert(t, fn.Name == "foo", "unexpected name: %s", fn.Name)
				assert(t, fn.Doc == "", "unexpected doc: %s", fn.Doc)
				assert(t, fn.Env != nil, "Env not set on Fn")
				assert(t, len(fn.Funcs) == 1, "expected only one method, got %d", len(fn.Funcs))
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.env = core.New(nil)
			runSpecialTest(t, tt, parseFn)
		})
	}
}

func Test_parseDo(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title: "NilArgs",
			env:   core.New(nil),
			args:  nil,
			assert: func(t *testing.T, got core.Expr, err error) {
				want := builtin.DoExpr{}
				assert(t, reflect.DeepEqual(want, got), "want=%#v\ngot=%#v", want, got)
			},
		},
		{
			title: "SomeArgs",
			env:   core.New(nil),
			args:  builtin.NewList(1, 2),
			assert: func(t *testing.T, got core.Expr, err error) {
				want := builtin.DoExpr{
					Exprs: []core.Expr{
						builtin.ConstExpr{Const: 1},
						builtin.ConstExpr{Const: 2},
					},
				}
				assert(t, reflect.DeepEqual(want, got), "want=%#v\ngot=%#v", want, got)
			},
		},
		{
			title:   "AnalyzeFail",
			env:     core.New(nil),
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
				assert(t, reflect.DeepEqual(want, got), "want=%#v\ngot=%#v", want, got)
			},
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.env = core.New(nil)
			runSpecialTest(t, tt, parseDef)
		})
	}
}

type specialTest struct {
	title   string
	env     core.Env
	args    core.Seq
	want    core.Expr
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
			"macro": parseMacro,
			"quote": parseQuote,
		},
	}
	got, err := parse(a, tt.env, tt.args)
	if tt.wantErr != nil {
		assert(t, sameErr(err, tt.wantErr),
			"wantErr=%#v\ngotErr=%#v", tt.wantErr, err)
		assert(t, got == nil, "expecting nil, got %#v", got)
	} else {
		assert(t, err == nil, "unexpected err: %#v", err)
	}
	if tt.assert != nil {
		tt.assert(t, got, err)
	}
}

func sameErr(e1, e2 error) bool {
	return e1 == e2 ||
		errors.Is(e1, e2) ||
		errors.Is(e2, e1) ||
		reflect.DeepEqual(e1, e2)
}
