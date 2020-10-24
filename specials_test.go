package slurp

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

func Test_parseDo(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title: "NilArgs",
			env:   builtin.New(nil),
			args:  nil,
			want:  builtin.DoExpr{},
		},
		{
			title: "SomeArgs",
			env:   builtin.New(nil),
			args:  builtin.NewList(1, 2),
			want: builtin.DoExpr{
				Exprs: []core.Expr{
					builtin.ConstExpr{Const: 1},
					builtin.ConstExpr{Const: 2},
				},
			},
		},
		{
			title:   "AnalyzeFail",
			env:     builtin.New(nil),
			args:    builtin.NewList(1, builtin.NewList(builtin.Symbol("def"))),
			want:    nil,
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
			want:    nil,
			wantErr: ErrParseSpecial,
		},
		{
			title: "Valid",
			args:  builtin.NewList(builtin.Symbol("foo"), 100),
			want: builtin.DefExpr{
				Name:  "foo",
				Value: builtin.ConstExpr{Const: 100},
			},
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.env = builtin.New(nil)
			runSpecialTest(t, tt, parseDef)
		})
	}
}

type specialTest struct {
	title   string
	env     core.Env
	args    builtin.Seq
	want    core.Expr
	wantErr error
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
		assert(t, errors.Is(err, tt.wantErr),
			"wantErr=%#v\ngotErr=%#v", tt.wantErr, err)
		assert(t, got == nil, "expecting nil, got %#v", got)
	} else {
		assert(t, err == nil, "unexpected err: %#v", err)
		assert(t, reflect.DeepEqual(tt.want, got),
			"want=%#v\ngot=%#v", tt.want, got)
	}
}
