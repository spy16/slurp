package builtin

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
			want:  DoExpr{},
		},
		{
			title: "SomeArgs",
			env:   core.New(nil),
			args:  NewList(1, 2),
			want: DoExpr{
				Exprs: []core.Expr{
					ConstExpr{Const: 1},
					ConstExpr{Const: 2},
				},
			},
		},
		{
			title:   "AnalyzeFail",
			env:     core.New(nil),
			args:    NewList(1, NewList(Symbol("def"))),
			want:    nil,
			wantErr: ErrSpecialForm,
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
			wantErr: ErrSpecialForm,
		},
		{
			title:   "SomeArgs",
			args:    NewList(1, 2),
			want:    nil,
			wantErr: ErrSpecialForm,
		},
		{
			title: "Valid",
			args:  NewList(Symbol("foo"), 100),
			want: &DefExpr{
				Name:  "foo",
				Value: ConstExpr{Const: 100},
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
	args    Seq
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
