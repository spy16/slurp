package slurp

import (
	"bytes"
	"errors"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
	"github.com/spy16/slurp/reflector"
)

// New returns a new slurp interpreter session.
func New(globals map[string]core.Any) *Instance {
	buf := bytes.Buffer{}
	ins := &Instance{
		buf:      &buf,
		reader:   reader.New(&buf),
		env:      core.New(globals),
		analyzer: builtin.NewAnalyzer(),
	}

	globals["macroexpand"] = reflector.Func("macroexpand", func(form core.Any) (core.Any, error) {
		f, err := builtin.MacroExpand(ins.analyzer, ins.env, form)
		if errors.Is(err, builtin.ErrNoExpand) {
			return form, nil
		}
		return f, err
	})
	return ins
}

// Instance represents a Slurp interpreter session.
type Instance struct {
	env      core.Env
	buf      *bytes.Buffer
	reader   *reader.Reader
	analyzer core.Analyzer
}

// Eval performs syntax analysis of the given form to produce an Expr
// and evalautes the Expr for result.
func (ins *Instance) Eval(form core.Any) (core.Any, error) {
	return core.Eval(ins.env, ins.analyzer, form)
}

func (ins *Instance) EvalStr(s string) (core.Any, error) {
	if _, err := ins.buf.WriteString(s); err != nil {
		return nil, err
	}

	f, err := ins.reader.All()
	if err != nil {
		return nil, err
	}

	do, err := builtin.Cons(builtin.Symbol("do"), builtin.NewList(f...))
	if err != nil {
		return nil, err
	}

	return ins.Eval(do)
}
