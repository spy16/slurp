package builtin

import (
	"errors"
	"fmt"

	"github.com/spy16/slurp/core"
)

// ErrNoExpand is returned during macro-expansion to indicate the form
// was not expanded.
var ErrNoExpand = errors.New("no macro expansion")

// Analyzer parses builtin value forms and returns Expr that can
// be evaluated against Env.
type Analyzer struct {
	Specials map[string]ParseSpecial
}

// ParseSpecial validates a special form invocation, parse the form and
// returns an expression that can be evaluated for result.
type ParseSpecial func(analyzer core.Analyzer, env core.Env, args core.Seq) (core.Expr, error)

// Analyze performs syntactic analysis of given form and returns an Expr
// that can be evaluated for result against an Env.
func (ba Analyzer) Analyze(env core.Env, form core.Any) (core.Expr, error) {
	if IsNil(form) {
		return ConstExpr{Const: Nil{}}, nil
	}

	exp, err := macroExpand(ba, env, form)
	if err != nil && !errors.Is(err, ErrNoExpand) {
		return nil, err
	}

	switch f := exp.(type) {
	case Symbol:
		return ResolveExpr{Symbol: f}, nil

	case core.Vector:
		return VectorExpr{
			Vector:   f,
			Analyzer: ba,
		}, nil

	case core.Seq:
		cnt, err := f.Count()
		if err != nil {
			return nil, err
		} else if cnt == 0 {
			break
		}

		return ba.analyzeSeq(env, f)
	}

	return ConstExpr{Const: exp}, nil
}

func (ba Analyzer) analyzeSeq(env core.Env, seq core.Seq) (core.Expr, error) {
	//	Get the call target.  This is the first item in the sequence.
	first, err := seq.First()
	if err != nil {
		return nil, err
	}

	// The call target may be a special form.  In this case, we need to get the
	// corresponding parser function, which will take care of parsing/analyzing
	// the tail.
	if sym, ok := first.(Symbol); ok {
		if parse, found := ba.Specials[string(sym)]; found {
			next, err := seq.Next()
			if err != nil {
				return nil, err
			}
			return parse(ba, env, next)
		}
	}

	// Call target is not a special form and must be a Invokable. Analyze
	// the arguments and create an InvokeExpr.
	ie := InvokeExpr{Name: fmt.Sprintf("%v", first)}
	err = core.ForEach(seq, func(item core.Any) (done bool, err error) {
		if ie.Target == nil {
			ie.Target, err = ba.Analyze(env, first)
			return
		}

		var arg core.Expr
		if arg, err = ba.Analyze(env, item); err == nil {
			ie.Args = append(ie.Args, arg)
		}
		return
	})
	return ie, err
}

func macroExpand(a core.Analyzer, env core.Env, form core.Any) (core.Any, error) {
	res, err := macroExpand1(env, form)
	if errors.Is(err, ErrNoExpand) {
		return res, nil
	} else if err != nil {
		return nil, err
	}
	return a.Analyze(env, res)
}

func macroExpand1(env core.Env, form core.Any) (core.Any, error) {
	lst, ok := form.(core.Seq)
	if !ok {
		return form, ErrNoExpand
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
			return form, ErrNoExpand
		}
		target = v
	}

	fn, ok := target.(Fn)
	if !ok || !fn.Macro {
		return form, ErrNoExpand
	}

	sl, err := core.ToSlice(lst)
	if err != nil {
		return nil, err
	}

	return fn.Invoke(sl[1:]...)
}
