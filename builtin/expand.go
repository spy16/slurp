package builtin

import (
	"errors"
	"fmt"

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
	sym, ok := first.(Symbol)
	if ok {
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

	res, err := fn.Invoke(sl[1:]...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func quoteExpand(a core.Analyzer, env core.Env, form core.Any) (core.Any, error) {
	switch v := form.(type) {
	case Seq:
		s, err := quoteSeq(a, env, v)
		if err != nil {
			return nil, err
		}
		return Cons(Symbol("list"), NewList(s...))
	}

	return NewList(Symbol("quote"), form), nil
}

func quoteSeq(a core.Analyzer, env core.Env, seq Seq) ([]core.Any, error) {
	var items []core.Any
	err := ForEach(seq, func(item core.Any) (bool, error) {
		q, err := recursiveQuote(a, env, item)
		if err != nil {
			return true, err
		}
		items = append(items, q)
		return false, nil
	})
	return items, err
}

func recursiveQuote(a core.Analyzer, env core.Env, form core.Any) (core.Any, error) {
	switch v := form.(type) {
	case Seq:
		first, err := v.First()
		if err != nil {
			return nil, err
		}

		if sym, ok := first.(Symbol); ok && sym == "unquote" {
			rest, err := v.Next()
			if err != nil {
				return nil, err
			}
			count, err := rest.Count()
			if err != nil {
				return nil, err
			} else if count != 1 {
				return nil, fmt.Errorf(
					"%w: unquote needs exactly 1 arg, got %d", ErrParseSpecial, count)
			}
			return rest.First()
		}

		return quoteSeq(a, env, v)
	}

	return Cons(Symbol("quote"), NewList(form))
}
