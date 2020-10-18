package core

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	_ Expr = (*DoExpr)(nil)
	_ Expr = (*IfExpr)(nil)
	_ Expr = (*DefExpr)(nil)
	_ Expr = (*QuoteExpr)(nil)
	_ Expr = (*ConstExpr)(nil)
	_ Expr = (*InvokeExpr)(nil)
	_ Expr = (*ResolveExpr)(nil)
)

// ConstExpr returns the Const value wrapped inside when evaluated. It has
// no side-effect on the VM.
type ConstExpr struct{ Const Any }

// Eval returns the constant value unmodified.
func (ce ConstExpr) Eval(_ Env) (Any, error) { return ce.Const, nil }

// QuoteExpr expression represents a quoted form and
type QuoteExpr struct{ Form Any }

// Eval returns the quoted form unmodified.
func (qe QuoteExpr) Eval(_ Env) (Any, error) { return qe.Form, nil }

// DefExpr represents the (def name value) binding form.
type DefExpr struct {
	Name  string
	Value Expr
}

// Eval creates the binding with the name and value in Root env.
func (de DefExpr) Eval(env Env) (Any, error) {
	var val Any
	var err error
	if de.Value != nil {
		val, err = de.Value.Eval(env)
		if err != nil {
			return nil, err
		}
	} else {
		val = Nil{}
	}

	if err := Root(env).Bind(de.Name, val); err != nil {
		return nil, err
	}
	return Symbol(de.Name), nil
}

// IfExpr represents the if-then-else form.
type IfExpr struct{ Test, Then, Else Expr }

// Eval evaluates the then or else expr based on truthiness of the test
// expr result.
func (ife IfExpr) Eval(env Env) (Any, error) {
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
type DoExpr struct{ Exprs []Expr }

// Eval evaluates each expr in the do form in the order and returns the
// result of the last eval.
func (de DoExpr) Eval(env Env) (Any, error) {
	var res Any
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

// ResolveExpr resolves a symbol from the given environment.
type ResolveExpr struct{ Symbol Symbol }

// Eval resolves the symbol in the given environment or its parent env
// and returns the result. Returns ErrNotFound if the symbol was not
// found in the entire heirarchy.
func (re ResolveExpr) Eval(env Env) (Any, error) {
	var v Any
	var err error
	for env != nil {
		v, err = env.Resolve(string(re.Symbol))
		if errors.Is(err, ErrNotFound) {
			// not found in the current frame. check parent.
			env = env.Parent()
			continue
		}

		// found the symbol or there was some unexpected error.
		break

	}
	return v, err
}

// GoExpr evaluates an expression in a separate goroutine.
type GoExpr struct{ Form Expr }

// Eval forks the given env to get a child env and launches goroutine
// with the child env to evaluate the form.
func (ge GoExpr) Eval(env Env) (Any, error) {
	// TODO: verify this.
	e := env.Child("<go>", nil)

	go func() {
		_, _ = ge.Form.Eval(e)
	}()

	return nil, nil
}

// InvokeExpr performs invocation of target when evaluated.
type InvokeExpr struct {
	Name   string
	Target Expr
	Args   []Expr
}

// Eval evaluates the target expr and invokes the result if it is an
// Invokable  Returns error otherwise.
func (ie InvokeExpr) Eval(env Env) (Any, error) {
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

	args := make([]Any, len(ie.Args))
	for i, ae := range ie.Args {
		if args[i], err = ae.Eval(env); err != nil {
			return nil, err
		}
	}

	return fn.Invoke(args...)
}
