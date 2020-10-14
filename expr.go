package slurp

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	_ Expr = (*ConstExpr)(nil)
	_ Expr = (*DefExpr)(nil)
	_ Expr = (*QuoteExpr)(nil)
	_ Expr = (*InvokeExpr)(nil)
	_ Expr = (*IfExpr)(nil)
	_ Expr = (*DoExpr)(nil)
)

// Expr represents an expression that can be evaluated against a context.
type Expr interface {
	// Eval should execute the expr and return the results. Eval can have
	// side-effects (e.g., DefExpr).
	Eval() (Any, error)
}

// ConstExpr returns the Const value wrapped inside when evaluated. It has
// no side-effect on the VM.
type ConstExpr struct{ Const Any }

// Eval returns the constant value unmodified.
func (ce ConstExpr) Eval() (Any, error) { return ce.Const, nil }

// QuoteExpr expression represents a quoted form and
type QuoteExpr struct{ Form Any }

// Eval returns the quoted form unmodified.
func (qe QuoteExpr) Eval() (Any, error) {
	// TODO: re-use this for syntax-quote and unquote?
	return qe.Form, nil
}

// DefExpr creates a global binding with the Name when evaluated.
type DefExpr struct {
	Env   *Env
	Name  string
	Value Expr
}

// Eval creates a symbol binding in the global (root) stack frame.
func (de DefExpr) Eval() (Any, error) {
	de.Name = strings.TrimSpace(de.Name)
	if de.Name == "" {
		return nil, fmt.Errorf("%w: '%s'", ErrInvalidBindName, de.Name)
	}

	var val Any
	var err error
	if de.Value != nil {
		val, err = de.Value.Eval()
		if err != nil {
			return nil, err
		}
	} else {
		val = Nil{}
	}

	de.Env.setGlobal(de.Name, val)
	return Symbol(de.Name), nil
}

// IfExpr represents the if-then-else form.
type IfExpr struct{ Test, Then, Else Expr }

// Eval the expression
func (ife IfExpr) Eval() (Any, error) {
	target := ife.Else
	if ife.Test != nil {
		test, err := ife.Test.Eval()
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
	return target.Eval()
}

// DoExpr represents the (do expr*) form.
type DoExpr struct{ Exprs []Expr }

// Eval the expression
func (de DoExpr) Eval() (Any, error) {
	var res Any
	var err error

	for _, expr := range de.Exprs {
		res, err = expr.Eval()
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
	Env    *Env
	Name   string
	Target Expr
	Args   []Expr
}

// Eval the expression
func (ie InvokeExpr) Eval() (Any, error) {
	val, err := ie.Target.Eval()
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
		if args[i], err = ae.Eval(); err != nil {
			return nil, err
		}
	}

	ie.Env.push(stackFrame{
		Name: ie.Name,
		Args: args,
		Vars: map[string]Any{},
	})
	defer ie.Env.pop()

	return fn.Invoke(ie.Env, args...)
}

// GoExpr evaluates an expression in a separate goroutine.
type GoExpr struct {
	Env  *Env
	Form Any
}

// Eval forks the given context to get a child context and launches goroutine
// with the child context to evaluate the form.
func (ge GoExpr) Eval() (Any, error) {
	child := ge.Env.Fork()
	go func() {
		_, _ = child.Eval(ge.Form)
	}()
	return nil, nil
}
