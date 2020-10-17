package builtin

import (
	"fmt"
	"reflect"

	"github.com/spy16/slurp/core"
)

var (
	_ core.Expr = (*ConstExpr)(nil)
	_ core.Expr = (*DefExpr)(nil)
	_ core.Expr = (*QuoteExpr)(nil)
	_ core.Expr = (*InvokeExpr)(nil)
	_ core.Expr = (*IfExpr)(nil)
	_ core.Expr = (*DoExpr)(nil)
)

// ConstExpr returns the Const value wrapped inside when evaluated. It has
// no side-effect on the VM.
type ConstExpr struct{ Const core.Any }

// Eval returns the constant value unmodified.
func (ce ConstExpr) Eval(_ core.Env) (core.Any, error) { return ce.Const, nil }

// QuoteExpr expression represents a quoted form and
type QuoteExpr struct{ Form core.Any }

// Eval returns the quoted form unmodified.
//
// TODO: re-use this for syntax-quote and unquote?
func (qe QuoteExpr) Eval(_ core.Env) (core.Any, error) { return qe.Form, nil }

type DefExpr struct {
	Name  string
	Value core.Expr
}

func (de DefExpr) Eval(env core.Env) (core.Any, error) {
	var val core.Any
	var err error
	if de.Value != nil {
		val, err = de.Value.Eval(env)
		if err != nil {
			return nil, err
		}
	} else {
		val = Nil{}
	}

	if err := core.Root(env).Bind(de.Name, val); err != nil {
		return nil, err
	}
	return Symbol(de.Name), nil
}

// IfExpr represents the if-then-else form.
type IfExpr struct{ Test, Then, Else core.Expr }

// Eval the expression
func (ife IfExpr) Eval(env core.Env) (core.Any, error) {
	target := ife.Else
	if ife.Test != nil {
		test, err := ife.Test.Eval(env)
		if err != nil {
			return nil, err
		}
		if IsTruthy(test) {
			target = ife.Then
		}
	}

	if target == nil {
		return Nil{}, nil
	}
	return target.Eval(env)
}

// DoExpr represents the (do expr*) form.
type DoExpr struct{ Exprs []core.Expr }

// Eval the expression
func (de DoExpr) Eval(env core.Env) (core.Any, error) {
	var res core.Any
	var err error

	for _, expr := range de.Exprs {
		res, err = expr.Eval(env)
		if err != nil {
			return nil, err
		}
	}

	if res == nil {
		return Nil{}, nil
	}
	return res, nil
}

// InvokeExpr performs invocation of target when evaluated.
type InvokeExpr struct {
	Name   string
	Target core.Expr
	Args   []core.Expr
}

// Eval the expression
func (ie InvokeExpr) Eval(env core.Env) (core.Any, error) {
	val, err := ie.Target.Eval(env)
	if err != nil {
		return nil, err
	}

	fn, ok := val.(Invokable)
	if !ok {
		return nil, Error{
			Cause:   ErrNotInvokable,
			Message: fmt.Sprintf("value of type '%s' is not invokable", reflect.TypeOf(val)),
		}
	}

	args := make([]core.Any, len(ie.Args))
	for i, ae := range ie.Args {
		if args[i], err = ae.Eval(env); err != nil {
			return nil, err
		}
	}

	return fn.Invoke(args...)
}

// GoExpr evaluates an expression in a separate goroutine.
type GoExpr struct {
	Form core.Expr
}

// Eval forks the given context to get a child context and launches goroutine
// with the child context to evaluate the form.
func (ge GoExpr) Eval(env core.Env) (core.Any, error) {
	e := core.Root(env).Child("<go>", nil)

	go func() {
		_, _ = ge.Form.Eval(e)
	}()

	return nil, nil
}

// Invokable represents a value that can be invoked for result.
type Invokable interface {
	// Invoke is called if this value appears as the first argument of
	// invocation form (i.e., list).
	Invoke(args ...core.Any) (core.Any, error)
}
