package builtin

import (
	"errors"

	"github.com/spy16/slurp/core"
)

var ErrNoExpand = errors.New("no macro expansion")

func MacroExpand(a core.Analyzer, env core.Env, form core.Any) (core.Any, error) {
	lst, ok := form.(Seq)
	if !ok {
		return nil, ErrNoExpand
	}

	first, err := lst.First()
	if err != nil {
		return nil, err
	}

	var target core.Any
	if sym, ok := first.(Symbol); ok {
		v, err := ResolveExpr{Symbol: sym}.Eval(env)
		if err != nil {
			return nil, ErrNoExpand
		}
		target = v
	}

	fn, ok := target.(Fn)
	if !ok || !fn.Macro {
		return nil, ErrNoExpand
	}

	sl, err := toSlice(lst)
	if err != nil {
		return nil, err
	}

	return fn.Invoke(sl[1:]...)
}
