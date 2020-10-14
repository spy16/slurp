package slurp

import (
	"errors"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	e := New()
	assert(t, e != nil, "expected env to be not nil")
}

func TestEnv_Eval(t *testing.T) {
	t.Parallel()

	table := []struct {
		title   string
		form    Any
		want    Any
		wantErr error
	}{
		{
			title: "ConstValue",
			form:  100,
			want:  100,
		},
		{
			title: "NilValue",
			form:  nil,
			want:  Nil{},
		},
		{
			title: "AlreadyExpr",
			form:  &ConstExpr{Const: 10},
			want:  10,
		},
		{
			title:   "InvokeInt",
			form:    NewList(10),
			wantErr: ErrNotInvokable,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			e := New()
			got, err := e.Eval(tt.form)
			if tt.wantErr != nil {
				assert(t, errors.Is(err, tt.wantErr),
					"wantErr=%#v\ngotErr=%#v", tt.wantErr, err)
				assert(t, got == nil, "wanted result to be nil, got=%#v", got)
			} else {
				assert(t, err == nil, "unexpected error: %#v", err)
				assert(t, reflect.DeepEqual(tt.want, got),
					"want=%#v\ngot=%#v", tt.want, got)
			}
		})
	}
}

func TestEnv_Resolve(t *testing.T) {
	e := New()
	e.globals.Store("foo", true)

	// resolves to global 'foo'.
	got := e.Resolve("foo")
	assert(t, got == true, "want=true got=%#v", got)

	// when inside a frame, local binding will shadow global 'foo'.
	e.push(stackFrame{
		Name: "test",
		Vars: map[string]Any{
			"foo": 10,
		},
	})
	got = e.Resolve("foo")
	assert(t, got == 10, "want=10 got=%#v", got)

	// after popping, old value of 'foo' should be restored.
	e.pop()
	got = e.Resolve("foo")
	assert(t, got == true, "want=true got=%#v", got)
}

func Benchmark_Eval_QuoteForm(b *testing.B) {
	env := New()
	quoteForm := QuoteExpr{Form: NewList(1, 2, 3)}

	for i := 0; i < b.N; i++ {
		_, err := env.Eval(quoteForm)
		if err != nil {
			b.Fatalf("eval failed: %v", err)
		}
	}
}

func Benchmark_Eval_Invocation(b *testing.B) {
	env := New()
	invocation := InvokeExpr{
		Env:  env,
		Name: "test",
		Target: ConstExpr{Const: fakeInvokable(func(env *Env, args ...Any) (Any, error) {
			return 10, nil
		})},
	}

	var finalRes Any
	for i := 0; i < b.N; i++ {
		res, err := env.Eval(invocation)
		if err != nil {
			b.Fatalf("eval failed: %v", err)
		}
		finalRes = res
	}
	assert(b, finalRes == 10, "want=10, got=%#v", finalRes)
}

func assert(t testInstance, cond bool, msg string, args ...interface{}) {
	if !cond {
		t.Errorf(msg, args...)
	}
}

type testInstance interface {
	Errorf(msg string, args ...interface{})
}
