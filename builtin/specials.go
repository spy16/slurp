package builtin

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/spy16/slurp/core"
)

// ErrSpecialForm is returned when parsing a special form invocation
// fails due to malformed syntax.
var ErrSpecialForm = errors.New("invalid sepcial form")

func parseDo(a core.Analyzer, env core.Env, args Seq) (core.Expr, error) {
	var de DoExpr
	err := ForEach(args, func(item core.Any) (bool, error) {
		expr, err := a.Analyze(env, item)
		if err != nil {
			return true, err
		}
		de.Exprs = append(de.Exprs, expr)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return de, nil
}

func parseIf(a core.Analyzer, env core.Env, args Seq) (core.Expr, error) {
	count, err := args.Count()
	if err != nil {
		return nil, err
	} else if count != 2 && count != 3 {
		return nil, Error{
			Cause:   fmt.Errorf("%w: if", ErrSpecialForm),
			Message: fmt.Sprintf("requires 2 or 3 arguments, got %d", count),
		}
	}

	exprs := [3]core.Expr{}
	for i := 0; i < count; i++ {
		f, err := args.First()
		if err != nil {
			return nil, err
		}

		expr, err := a.Analyze(env, f)
		if err != nil {
			return nil, err
		}
		exprs[i] = expr

		args, err = args.Next()
		if err != nil {
			return nil, err
		}
	}

	return &IfExpr{
		Test: exprs[0],
		Then: exprs[1],
		Else: exprs[2],
	}, nil
}

func parseQuote(a core.Analyzer, _ core.Env, args Seq) (core.Expr, error) {
	if count, err := args.Count(); err != nil {
		return nil, err
	} else if count != 1 {
		return nil, Error{
			Cause:   errors.New("invalid quote form"),
			Message: fmt.Sprintf("requires exactly 1 argument, got %d", count),
		}
	}

	first, err := args.First()
	if err != nil {
		return nil, err
	}

	return QuoteExpr{
		Form: first,
	}, nil
}

func parseDef(a core.Analyzer, env core.Env, args Seq) (core.Expr, error) {
	e := Error{Cause: fmt.Errorf("%w: def", ErrSpecialForm)}

	if args == nil {
		return nil, e.With("requires exactly 2 args, got 0")
	}

	if count, err := args.Count(); err != nil {
		return nil, err
	} else if count != 2 {
		return nil, e.With(fmt.Sprintf(
			"requires exactly 2 arguments, got %d", count))
	}

	first, err := args.First()
	if err != nil {
		return nil, err
	}

	sym, ok := first.(Symbol)
	if !ok {
		return nil, e.With(fmt.Sprintf(
			"first arg must be symbol, not '%s'", reflect.TypeOf(first)))
	}

	rest, err := args.Next()
	if err != nil {
		return nil, err
	}

	second, err := rest.First()
	if err != nil {
		return nil, err
	}

	res, err := a.Analyze(env, second)
	if err != nil {
		return nil, err
	}

	return &DefExpr{
		Name:  string(sym),
		Value: res,
	}, nil
}

func parseGo(a core.Analyzer, env core.Env, args Seq) (core.Expr, error) {
	v, err := args.First()
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, Error{
			Cause: errors.New("go expr requires exactly one argument"),
		}
	}

	e, err := a.Analyze(env, v)
	if err != nil {
		return nil, err
	}

	return GoExpr{Form: e}, nil
}

func parseFn(a core.Analyzer, env core.Env, argSeq Seq) (core.Expr, error) {
	fn := Fn{}

	cnt, err := argSeq.Count()
	if err != nil {
		return nil, err
	} else if cnt < 2 {
		return nil, fmt.Errorf("%w: got %d, want at-least 2", ErrArity, cnt)
	}

	args, err := toSlice(argSeq)
	if err != nil {
		return nil, err
	}

	i := 0
	if sym, ok := args[i].(Symbol); ok {
		fn.Name = sym.String()
		i++
	}

	if str, ok := args[i].(String); ok {
		fn.Doc = string(str)
		i++
	}

	fnArgs, ok := args[i].(*LinkedList)
	if !ok {
		return nil, fmt.Errorf(
			"expecting a list of symbols, got '%s'", reflect.TypeOf(args[i]))
	}
	i++

	f := Func{}
	fnEnv := env.Child(fn.Name, nil)
	err = ForEach(fnArgs, func(item core.Any) (bool, error) {
		sym, ok := item.(Symbol)
		if !ok {
			return true, fmt.Errorf(
				"expecting parameter to be a symbol, got '%s'",
				reflect.TypeOf(item))
		}
		f.Params = append(f.Params, string(sym))

		if err := fnEnv.Bind(string(sym), nil); err != nil {
			return false, err
		}

		return false, nil
	})

	// wrap body in (do <expr>*) and analyze.
	bodyExprs, err := Cons(Symbol("do"), NewList(args[i:]...))
	if err != nil {
		return nil, err
	}

	body, err := a.Analyze(fnEnv, bodyExprs)
	if err != nil {
		return nil, err
	}
	f.Body = body

	fn.Env = fnEnv
	fn.Funcs = append(fn.Funcs, f)
	return &ConstExpr{Const: fn}, nil
}

func toSlice(seq Seq) ([]core.Any, error) {
	var sl []core.Any
	err := ForEach(seq, func(item core.Any) (bool, error) {
		sl = append(sl, item)
		return false, nil
	})
	return sl, err
}
