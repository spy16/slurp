package slurp

import (
	"bytes"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// New returns a new slurp interpreter session.
func New(globals map[string]core.Any) *Instance {
	buf := bytes.Buffer{}
	return &Instance{
		buf:      &buf,
		reader:   reader.New(&buf),
		env:      core.New(globals),
		analyzer: builtin.NewAnalyzer(),
	}
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
