package core

// Any represents any Go/slurp value.
type Any interface{}

// Env represents the environment in which forms are evaluated.
type Env interface {
	// Name returns the name of this env frame.
	Name() string

	// Parent returns the parent/outer env of this env. Returns nil
	// if this env is the root.
	Parent() Env

	// Bind creates a local binding with given name and value.
	Bind(name string, val Any) error

	// Resolve resolves the symbol in this env and return its value
	// if found. Returns ErrNotFound if name is not found in this
	// env frame.
	Resolve(name string) (Any, error)

	// Child returns a new env with given frame name and vars bound.
	// Returned env will have this env as parent/outer.
	Child(name string, vars map[string]Any) Env
}

// Analyzer implementation is responsible for performing syntax analysis
// on given form.
type Analyzer interface {
	// Analyze performs syntax analysis and macro expansions if supported
	// to produce Expr values which can be valuated against Env.
	Analyze(env Env, form Any) (Expr, error)
}

// Expr represents an expression that can be evaluated against an env.
type Expr interface {
	// Eval executes the expr against the env and returns the results.
	// It can have side-effects on env. (e.g., DefExpr).
	Eval(env Env) (Any, error)
}
