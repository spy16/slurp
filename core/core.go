// Package core provides the core execution primitives in slurp. This
// includes the Env, Analyzer and core functions built on them.
package core

import "errors"

var (
	// ErrNotFound is returned by Env when a binding is not found
	// for a given symbol/name.
	ErrNotFound = errors.New("not found")

	// ErrInvalidBindName is returned by Env when the bind name is
	// invalid.
	ErrInvalidBindName = errors.New("invalid bind name")
)

// Any represents any slurp value.
type Any interface{}

// Analyzer implementation is responsible for performing syntax analysis
// on given form.
type Analyzer interface {
	// Analyze should perform syntax checks for special forms etc. and
	// return Expr values that can be evaluated against a context.
	Analyze(env Env, form Any) (Expr, error)
}

// Expr represents an expression that can be evaluated against an env.
type Expr interface {
	// Eval executes the expr against the env and return the results.
	// Eval can have side-effects (e.g., DefExpr).
	Eval(env Env) (Any, error)
}

// Env represents the environment in which forms are evaluated.
type Env interface {
	// Name should return the name of the frame.
	Name() string

	// Parent should return the parent/outer env of this env.
	Parent() Env

	// Bind should create a local binding with given name and value.
	Bind(name string, val Any) error

	// Resolve should resolve the symbol and return its value. Lookup
	// should be done in locals first (this env) and then done in root
	// env.
	Resolve(name string) (Any, error)

	// Child should return a new env with given frame name and vars
	// bound. Returned env should have this env as parent/outer.
	Child(name string, vars map[string]Any) Env
}

// Eval performs syntax analysis of the given form to produce an Expr
// and evalautes the Expr against the given Env.
func Eval(env Env, analyzer Analyzer, form Any) (Any, error) {
	expr, err := analyzer.Analyze(env, form)
	if err != nil {
		return nil, err
	}
	return expr.Eval(env)
}

// Root traverses the env hierarchy until it reaches the root env
// and returns it.
func Root(env Env) Env {
	for env.Parent() != nil {
		env = env.Parent()
	}
	return env
}
