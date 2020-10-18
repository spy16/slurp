package slurp

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp/core"
)

func Test_parseDo(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title: "NilArgs",
			env:   core.New(nil),
			args:  nil,
			want:  core.DoExpr{},
		},
		{
			title: "SomeArgs",
			env:   core.New(nil),
			args:  core.NewList(1, 2),
			want: core.DoExpr{
				Exprs: []core.Expr{
					core.ConstExpr{Const: 1},
					core.ConstExpr{Const: 2},
				},
			},
		},
		{
			title:   "AnalyzeFail",
			env:     core.New(nil),
			args:    core.NewList(1, core.NewList(core.Symbol("def"))),
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
			args:    core.NewList(1, 2),
			want:    nil,
			wantErr: ErrParseSpecial,
		},
		{
			title: "Valid",
			args:  core.NewList(core.Symbol("foo"), 100),
			want: core.DefExpr{
				Name:  "foo",
				Value: core.ConstExpr{Const: 100},
			},
			wantErr: nil,
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
}

func runSpecialTest(t *testing.T, tt specialTest, parse ParseSpecial) {
	a := NewAnalyzer()
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
