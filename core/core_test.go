package core_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

var errUnknown = errors.New("failed")

func TestEval(t *testing.T) {
	t.Parallel()

	table := []struct {
		title    string
		form     core.Any
		env      core.Env
		analyzer core.Analyzer
		want     core.Any
		wantErr  error
	}{
		{
			title:    "WithNilAnalyzer",
			env:      builtin.NewEnv(nil),
			analyzer: nil,
			form:     100,
			want:     100,
		},
		{
			title:    "WithCustomAnalyzer",
			env:      builtin.NewEnv(nil),
			analyzer: fakeAnalyzer{Res: "foo"},
			form:     100,
			want:     "foo",
		},
		{
			title:    "WithAnalyzerError",
			env:      builtin.NewEnv(nil),
			analyzer: fakeAnalyzer{Err: errUnknown},
			form:     100,
			want:     nil,
			wantErr:  errUnknown,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := core.Eval(tt.env, tt.analyzer, tt.form)
			if tt.wantErr != nil {
				assert(t, errors.Is(err, tt.wantErr),
					"\nwantErr=%#v\ngot=%#v", tt.wantErr, got)
				assert(t, got == nil, "want nil, got %#v", got)
			} else {
				assert(t, err == nil, "unexpected err: %#v", err)
				assert(t, reflect.DeepEqual(tt.want, got),
					"\nwant=%#v\ngot=%#v", tt.want, got)
			}
		})
	}
}

type fakeAnalyzer struct {
	Res core.Any
	Err error
}

func (fa fakeAnalyzer) Analyze(env core.Env, form core.Any) (core.Expr, error) {
	return builtin.ConstExpr{Const: fa.Res}, fa.Err
}
