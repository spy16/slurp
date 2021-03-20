package builtin

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spy16/slurp/core"
)

var (
	_ core.Invokable        = (*Fn)(nil)
	_ core.EqualityProvider = (*Fn)(nil)
)

// Fn represents a multi-arity function definition. Fn implements
// Invokable.
type Fn struct {
	Env   core.Env
	Name  string
	Doc   string
	Funcs []Func
	Macro bool
}

// Invoke selects and executes a func defined in the Fn and returns
// the result of execution.
func (fn Fn) Invoke(args ...core.Any) (core.Any, error) {
	f, err := fn.selectFunc(args)
	if err != nil {
		return nil, err
	}

	env := fn.Env.Child(fn.Name, nil)
	for i, p := range f.Params {
		if err := env.Bind(p, args[i]); err != nil {
			return nil, err
		}
	}

	return f.Body.Eval(env)
}

// Equals returns true if 'v' is also a MultiFn and all methods are
// equivalent.
func (fn Fn) Equals(v core.Any) (bool, error) {
	other, ok := v.(Fn)
	if !ok {
		return false, nil
	}

	sameHeader := (fn.Name == other.Name) &&
		(len(fn.Funcs) == len(other.Funcs))
	if !sameHeader {
		return false, nil
	}

	for i, fn1 := range fn.Funcs {
		fn2 := other.Funcs[i]
		eq, err := fn1.compare(fn2)
		if err != nil || !eq {
			return eq, err
		}
	}

	return true, nil
}

func (fn Fn) String() string {
	buf := strings.Builder{}
	buf.WriteString("(fn ")
	if fn.Name != "" {
		buf.WriteString(fn.Name + " ")
	}

	if fn.Doc != "" {
		buf.WriteString(fmt.Sprintf("\n  \"%s\"", fn.Doc))
	}

	buf.WriteString(")")

	return buf.String()
}

func (fn Fn) selectFunc(args []core.Any) (Func, error) {
	for _, f := range fn.Funcs {
		if f.matchArity(args) {
			return f, nil
		}
	}

	return Func{}, fmt.Errorf(
		"%w (%d) to '%s'", core.ErrArity, len(args), fn.Name)
}

// Func represents a method of specific arity in Fn.
type Func struct {
	Body     core.Expr
	Params   []string
	Variadic bool
}

func (f Func) matchArity(args []core.Any) bool {
	argc := len(args)
	if f.Variadic {
		return argc >= len(f.Params)-1
	}
	return argc == len(f.Params)
}

// func (f Func) minArity() int {
// 	if len(f.Params) > 0 && f.Variadic {
// 		return len(f.Params) - 1
// 	}
// 	return len(f.Params)
// }

func (f *Func) compare(other Func) (bool, error) {
	if f.Variadic != other.Variadic ||
		!reflect.DeepEqual(f.Params, other.Params) {
		return false, nil
	}

	bodyEq, err := core.Eq(f.Body, other.Body)
	if err != nil {
		return false, err
	}

	return bodyEq, nil
}
