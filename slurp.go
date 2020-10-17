package slurp

import (
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

// New returns a new slurp interpreter session.
func New(globals map[string]core.Any) *Instance {
	return &Instance{
		env:      core.New(globals),
		analyzer: builtin.NewAnalyzer(),
	}
}

// Instance represents a Slurp interpreter session.
type Instance struct {
	env      core.Env
	analyzer core.Analyzer
}

// Eval performs syntax analysis of the given form to produce an Expr
// and evalautes the Expr for result.
func (ins *Instance) Eval(form core.Any) (core.Any, error) {
	return core.Eval(ins.env, ins.analyzer, form)
}
