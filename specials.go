package slurp

import (
	"errors"
	"fmt"
	"reflect"
)

// ErrSpecialForm is returned when parsing a special form invocation
// fails due to malformed syntax.
var ErrSpecialForm = errors.New("invalid sepcial form")

func parseDoExpr(env *Env, args Seq) (Expr, error) {
	var de DoExpr
	err := ForEach(args, func(item Any) (bool, error) {
		expr, err := env.Analyze(item)
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

func parseIfExpr(env *Env, args Seq) (Expr, error) {
	count, err := args.Count()
	if err != nil {
		return nil, err
	} else if count != 2 && count != 3 {
		return nil, Error{
			Cause:   fmt.Errorf("%w: if", ErrSpecialForm),
			Message: fmt.Sprintf("requires 2 or 3 arguments, got %d", count),
		}
	}

	exprs := [3]Expr{}
	for i := 0; i < count; i++ {
		f, err := args.First()
		if err != nil {
			return nil, err
		}

		expr, err := env.Analyze(f)
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

func parseQuoteExpr(_ *Env, args Seq) (Expr, error) {
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

func parseDefExpr(env *Env, args Seq) (Expr, error) {
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

	res, err := env.Analyze(second)
	if err != nil {
		return nil, err
	}

	return &DefExpr{
		Env:   env,
		Name:  string(sym),
		Value: res,
	}, nil
}

func parseGoExpr(env *Env, args Seq) (Expr, error) {
	v, err := args.First()
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, Error{
			Cause: errors.New("go expr requires exactly one argument"),
		}
	}

	return GoExpr{Env: env, Form: v}, nil
}
