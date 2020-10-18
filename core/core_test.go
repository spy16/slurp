package core

import (
	"errors"
	"testing"
)

var errUnknown = errors.New("failed")

func assert(t *testing.T, cond bool, msg string, args ...interface{}) {
	if !cond {
		t.Errorf(msg, args...)
	}
}

type fakeAnalyzer struct {
	Res Any
	Err error
}

func (fa fakeAnalyzer) Analyze(env Env, form Any) (Expr, error) {
	return ConstExpr{Const: fa.Res}, fa.Err
}

type fakeExpr struct {
	Res Any
	Err error
}

func (f fakeExpr) Eval(_ Env) (Any, error) { return f.Res, f.Err }

type fakeInvokable func(args ...Any) (Any, error)

func (f fakeInvokable) Invoke(args ...Any) (Any, error) {
	return f(args...)
}
