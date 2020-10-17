package builtin

import (
	"fmt"

	"github.com/spy16/slurp/core"
)

var _ Invokable = (*Fn)(nil)

type Fn struct {
	Env    core.Env
	Name   string
	Params []string
	Body   core.Expr
}

func (fn Fn) Invoke(args ...core.Any) (core.Any, error) {
	vars := map[string]core.Any{}
	if len(fn.Params) != len(args) {
		return nil, fmt.Errorf("requires exactly %d args, got %d",
			len(fn.Params), len(args))
	}

	for i, p := range fn.Params {
		vars[p] = args[i]
	}
	e := fn.Env.Child(fn.Name, vars)

	return fn.Body.Eval(e)
}
