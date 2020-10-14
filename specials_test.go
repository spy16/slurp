package slurp

import (
	"errors"
	"reflect"
	"testing"
)

func Test_parseDo(t *testing.T) {
	t.Parallel()

	table := []specialTest{
		{
			title: "NilArgs",
			env:   New(),
			args:  nil,
			want:  DoExpr{},
		},
		{
			title: "SomeArgs",
			env:   New(),
			args:  NewList(1, 2),
			want: DoExpr{
				Exprs: []Expr{
					ConstExpr{Const: 1},
					ConstExpr{Const: 2},
				},
			},
		},
		{
			title:   "AnalyzeFail",
			env:     New(),
			args:    NewList(1, NewList(Symbol("def"))),
			want:    nil,
			wantErr: ErrSpecialForm,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			runSpecialTest(t, tt, parseDoExpr)
		})
	}
}

func Test_parseDef(t *testing.T) {
	t.Parallel()

	e := New()

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
				Env:   e,
				Name:  "foo",
				Value: ConstExpr{Const: 100},
			},
			wantErr: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			tt.env = e
			runSpecialTest(t, tt, parseDefExpr)
		})
	}
}

type specialTest struct {
	title   string
	env     *Env
	args    Seq
	want    Expr
	wantErr error
}

func runSpecialTest(t *testing.T, tt specialTest, parse ParseSpecial) {
	got, err := parse(tt.env, tt.args)
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
