package slurp

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

// ErrParseSpecial is returned when parsing a special form invocation
// fails due to malformed syntax.
var ErrParseSpecial = errors.New("invalid sepcial form")

func parseDo(a core.Analyzer, env core.Env, args builtin.Seq) (core.Expr, error) {
	var de builtin.DoExpr
	err := builtin.ForEach(args, func(item core.Any) (bool, error) {
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

func parseIf(a core.Analyzer, env core.Env, args builtin.Seq) (core.Expr, error) {
	count, err := args.Count()
	if err != nil {
		return nil, err
	} else if count != 2 && count != 3 {
		return nil, core.Error{
			Cause:   fmt.Errorf("%w: if", ErrParseSpecial),
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

	return builtin.IfExpr{
		Test: exprs[0],
		Then: exprs[1],
		Else: exprs[2],
	}, nil
}

func parseQuote(a core.Analyzer, _ core.Env, args builtin.Seq) (core.Expr, error) {
	if count, err := args.Count(); err != nil {
		return nil, err
	} else if count != 1 {
		return nil, core.Error{
			Cause:   fmt.Errorf("%w: quote", ErrParseSpecial),
			Message: fmt.Sprintf("requires exactly 1 argument, got %d", count),
		}
	}

	first, err := args.First()
	if err != nil {
		return nil, err
	}

	return builtin.QuoteExpr{Form: first}, nil
}

func parseDef(a core.Analyzer, env core.Env, args builtin.Seq) (core.Expr, error) {
	e := core.Error{Cause: fmt.Errorf("%w: def", ErrParseSpecial)}

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

	sym, ok := first.(builtin.Symbol)
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

	return builtin.DefExpr{
		Name:  string(sym),
		Value: res,
	}, nil
}

func parseGo(a core.Analyzer, env core.Env, args builtin.Seq) (core.Expr, error) {
	count, err := args.Count()
	if err != nil {
		return nil, err
	}

	v, err := args.First()
	if err != nil {
		return nil, err
	}

	if v == nil {
		return nil, core.Error{
			Cause:   fmt.Errorf("%w: go", ErrParseSpecial),
			Message: fmt.Sprintf("requires exactly 1 argument, got %d", count),
		}
	}

	e, err := a.Analyze(env, v)
	if err != nil {
		return nil, err
	}

	return builtin.GoExpr{Form: e}, nil
}

// parseFn parses (fn name? doc? (<params>*) <body>*) special form and
// returns an Fn definition.
func parseFn(a core.Analyzer, env core.Env, argSeq builtin.Seq) (core.Expr, error) {
	fn, err := parseFnDef(a, env, argSeq)
	if err != nil {
		return nil, err
	}
	return builtin.ConstExpr{Const: *fn}, nil
}

// parseMacro parses (macro name? doc? (<params>*) <body>*) special form and
// returns an Fn definition.
func parseMacro(a core.Analyzer, env core.Env, argSeq builtin.Seq) (core.Expr, error) {
	fn, err := parseFnDef(a, env, argSeq)
	if err != nil {
		return nil, err
	}
	fn.Macro = true
	return builtin.ConstExpr{Const: *fn}, nil
}

func parseFnDef(a core.Analyzer, env core.Env, argSeq builtin.Seq) (*builtin.Fn, error) {
	fn := builtin.Fn{}

	cnt, err := argSeq.Count()
	if err != nil {
		return nil, err
	} else if cnt < 1 {
		return nil, fmt.Errorf("%w: got %d, want at-least 1", builtin.ErrArity, cnt)
	}

	args, err := toSlice(argSeq)
	if err != nil {
		return nil, err
	}

	i := 0
	if sym, ok := args[i].(builtin.Symbol); ok {
		fn.Name = strings.TrimSpace(sym.String())
		i++
	}

	if str, ok := args[i].(builtin.String); ok {
		fn.Doc = string(str)
		i++
	}

	// TODO: add support for multi-arity parsing.

	fnArgs, ok := args[i].(builtin.Seq)
	if !ok {
		return nil, fmt.Errorf(
			"expecting a list of symbols, got '%s'", reflect.TypeOf(args[i]))
	}
	i++

	f := builtin.Func{}
	fnEnv := env.Child(fn.Name, nil)
	err = builtin.ForEach(fnArgs, func(item core.Any) (bool, error) {
		sym, ok := item.(builtin.Symbol)
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
	bodyExprs, err := builtin.Cons(builtin.Symbol("do"), builtin.NewList(args[i:]...))
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

	return &fn, nil
}

func toSlice(seq builtin.Seq) ([]core.Any, error) {
	var sl []core.Any
	err := builtin.ForEach(seq, func(item core.Any) (bool, error) {
		sl = append(sl, item)
		return false, nil
	})
	return sl, err
}
