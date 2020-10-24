package builtin

import (
	"errors"
	"reflect"
	"testing"

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
			env:      New(nil),
			analyzer: nil,
			form:     100,
			want:     100,
		},
		{
			title:    "WithCustomAnalyzer",
			env:      New(nil),
			analyzer: fakeAnalyzer{Res: "foo"},
			form:     100,
			want:     "foo",
		},
		{
			title:    "WithAnalyzerError",
			env:      New(nil),
			analyzer: fakeAnalyzer{Err: errUnknown},
			form:     100,
			want:     nil,
			wantErr:  errUnknown,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			got, err := Eval(tt.env, tt.analyzer, tt.form)
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

func TestRoot(t *testing.T) {
	env := New(nil)
	child := env.Child("foo", nil)
	got := Root(child)
	assert(t, !reflect.DeepEqual(child, got), "expecting root, got child")
	assert(t, reflect.DeepEqual(env, got), "returned env is not root")
	assert(t, env.Name() == got.Name(), "want='%s', got='%s'", env.Name(), got.Name())
}

func Test_mapEnv_Bind_Resolve(t *testing.T) {
	var v core.Any
	var err error

	env := New(map[string]core.Any{"foo": "bar"})

	err = env.Bind("v", 1000)
	assert(t, err == nil, "unexpected err: %+v", err)

	err = env.Bind("", 1000)
	assert(t, errors.Is(err, core.ErrInvalidName), "want ErrInvalidName, got '%+v'", err)

	v, err = env.Resolve("foo")
	assert(t, err == nil, "unexpected err: %+v", err)
	assert(t, v == "bar", "want=%+v\ngot=%+v", "bar", v)

	v, err = env.Resolve("non-existent")
	assert(t, errors.Is(err, core.ErrNotFound), "want ErrNotFound, got '%+v'", err)
	assert(t, v == nil, "want=nil, got=%+v", v)
}

type fakeAnalyzer struct {
	Res core.Any
	Err error
}

func (fa fakeAnalyzer) Analyze(env core.Env, form core.Any) (core.Expr, error) {
	return ConstExpr{Const: fa.Res}, fa.Err
}
