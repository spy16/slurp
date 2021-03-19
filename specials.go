package slurp

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

// ErrParseSpecial is returned when parsing a special form invocation
// fails due to malformed syntax.
var ErrParseSpecial = errors.New("invalid special form")

// parseDo parses the (do <expr>*) form and returns a DoExpr.
func parseDo(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
	var de builtin.DoExpr
	err := core.ForEach(args, func(item core.Any) (bool, error) {
		expr, err := a.Analyze(env, item)
		if err != nil {
			return true, err
		}
		de = append(de, expr)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return de, nil
}

// parseIf parses the (if <test> <then> <else>?) form and returns IfExpr.
func parseIf(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
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

func parseQuote(a core.Analyzer, _ core.Env, args core.Seq) (core.Expr, error) {
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

func parseDef(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
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

	sym, ok := first.(core.Symbol)
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
		Name:  sym,
		Value: res,
	}, nil
}

func parseLet(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {

	e := core.Error{Cause: fmt.Errorf("%w: def", ErrParseSpecial)}

	if args == nil {
		return nil, e.With("requires exactly 2 args, got 0")
	}

	cnt, err := args.Count()
	if err != nil {
		return nil, err
	}
	if cnt == 0 {
		return nil, e.With("requires exactly 2 args, got 0")
	}

	var let builtin.LetExpr

	// analyze bindings
	bs, err := args.First()
	if err != nil {
		return nil, err
	}

	switch s := bs.(type) {
	case core.Seq:
		if err = parseLetBindings(a, env, s, &let); err != nil {
			return nil, err
		}

	case core.Seqable:
		seq, err := s.Seq()
		if err != nil {
			return nil, err
		}
		if err = parseLetBindings(a, env, seq, &let); err != nil {
			return nil, err
		}

	default:
		return nil, core.Error{
			Cause:   fmt.Errorf("%w: let", ErrParseSpecial),
			Message: fmt.Sprintf("%s is not a sequence type", reflect.TypeOf(bs)),
		}
	}

	// analyze expressions
	if args, err = args.Next(); err != nil {
		return nil, err
	}

	err = core.ForEach(args, func(item core.Any) (bool, error) {
		expr, err := a.Analyze(env, item)
		if err == nil {
			let.Exprs = append(let.Exprs, expr)
		}
		return false, err
	})

	return let, err
}

func parseLetBindings(a core.Analyzer, env core.Env, seq core.Seq, le *builtin.LetExpr) error {
	return core.ForEach(seq, func(item core.Any) (bool, error) {
		// symbol?
		if len(le.Names)%2 == 0 {
			s, ok := item.(core.Symbol)
			if !ok {
				return false, core.Error{
					Cause:   fmt.Errorf("%w: let", ErrParseSpecial),
					Message: fmt.Sprintf("expected symbol, got %s", reflect.TypeOf(item)),
				}
			}

			if s.Qualified() {
				return false, core.Error{
					Cause:   fmt.Errorf("%s: let", ErrParseSpecial),
					Message: fmt.Sprintf("cannot bind fully-qualified symbol '%s' to local scope", s),
				}
			}

			le.Names = append(le.Names, s.String())
			return false, nil
		}

		// it's a value
		expr, err := a.Analyze(env, item)
		if err == nil {
			le.Values = append(le.Values, expr)
		}

		return false, err
	})
}

func parseGo(a core.Analyzer, env core.Env, args core.Seq) (core.Expr, error) {
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
func parseFn(a core.Analyzer, env core.Env, argSeq core.Seq) (core.Expr, error) {
	fn, err := parseFnDef(a, env, argSeq)
	if err != nil {
		return nil, err
	}
	return builtin.ConstExpr{Const: *fn}, nil
}

// parseMacro parses (macro name? doc? (<params>*) <body>*) special form and
// returns an Fn definition.
func parseMacro(a core.Analyzer, env core.Env, argSeq core.Seq) (core.Expr, error) {
	fn, err := parseFnDef(a, env, argSeq)
	if err != nil {
		return nil, err
	}
	fn.Macro = true
	return builtin.ConstExpr{Const: *fn}, nil
}

func parseFnDef(a core.Analyzer, env core.Env, argSeq core.Seq) (*builtin.Fn, error) {
	if argSeq == nil {
		return nil, errors.New("nil argument sequence")
	}

	fn := builtin.Fn{}

	cnt, err := argSeq.Count()
	if err != nil {
		return nil, err
	} else if cnt < 1 {
		return nil, fmt.Errorf("%w: got %d, want at-least 1", core.ErrArity, cnt)
	}

	args, err := core.ToSlice(argSeq)
	if err != nil {
		return nil, err
	}

	i := 0
	if sym, ok := args[i].(core.Symbol); ok {
		fn.Name = sym.String()
		i++
	}

	if str, ok := args[i].(builtin.String); ok {
		fn.Doc = string(str)
		i++
	}

	// TODO: add support for multi-arity parsing.

	fnArgs, ok := args[i].(core.Seq)
	if !ok {
		return nil, fmt.Errorf(
			"expecting a list of symbols, got '%s'", reflect.TypeOf(args[i]))
	}
	i++

	f := builtin.Func{}
	fnEnv := env.Child(fn.Name, nil)
	argSet := map[string]struct{}{}
	err = core.ForEach(fnArgs, func(item core.Any) (bool, error) {
		sym, ok := item.(core.Symbol)
		if !ok {
			return true, fmt.Errorf(
				"expecting parameter to be a symbol, got '%s'",
				reflect.TypeOf(item))
		}

		if sym.Qualified() {
			return false, core.Error{
				Cause:   ErrParseSpecial,
				Message: fmt.Sprintf("cannot bind fully-qualified symbol '%s' to local scope", sym),
			}
		}

		symName := sym.String()
		if _, found := argSet[symName]; found {
			return true, fmt.Errorf("duplicate arg name '%s'", sym)
		}
		argSet[symName] = struct{}{}
		f.Params = append(f.Params, sym)

		if err := fnEnv.Scope().Bind(sym, nil); err != nil {
			return false, err
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

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

// parseNS
func parseNS(a core.Analyzer, env core.Env, argSeq core.Seq) (core.Expr, error) {
	panic("NOT IMPLEMENTED")
}
