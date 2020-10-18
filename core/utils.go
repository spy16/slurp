package core

import (
	"fmt"
	"strings"
	"sync"
)

const rootEnv = "<main>"

var (
	_ Env      = (*mapEnv)(nil)
	_ Expr     = (*constExpr)(nil)
	_ Analyzer = (*constAnalyzer)(nil)
)

// Eval performs syntax analysis of the given form to produce an Expr
// and evalautes the Expr against the given Env.
func Eval(env Env, analyzer Analyzer, form Any) (Any, error) {
	if analyzer == nil {
		analyzer = constAnalyzer{}
	}

	expr, err := analyzer.Analyze(env, form)
	if err != nil || expr == nil {
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

// New returns a root Env that can be used to execute forms.
func New(vars map[string]Any) Env {
	if vars == nil {
		vars = map[string]Any{}
	}
	return &mapEnv{
		parent: nil,
		name:   rootEnv,
		vars:   vars,
	}
}

// mapEnv implements Env using a Go native map and RWMutex.
type mapEnv struct {
	parent Env
	name   string
	mu     sync.RWMutex
	vars   map[string]Any
}

func (me *mapEnv) Name() string { return me.name }

func (me *mapEnv) Parent() Env { return me.parent }

func (me *mapEnv) Child(name string, vars map[string]Any) Env {
	if vars == nil {
		vars = map[string]Any{}
	}
	return &mapEnv{
		name:   name,
		parent: me,
		vars:   vars,
	}
}

func (me *mapEnv) Bind(name string, val Any) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: %s", ErrInvalidName, name)
	}

	if me.parent == nil {
		// only root env is shared between threads. so make sure
		// concurrent accesses are synchronized.
		me.mu.Lock()
		defer me.mu.Unlock()
	}

	me.vars[name] = val
	return nil
}

func (me *mapEnv) Resolve(name string) (Any, error) {
	if me.parent == nil {
		// only root env is shared between threads. so make sure
		// concurrent accesses are synchronized.
		me.mu.RLock()
		defer me.mu.RUnlock()
	}

	v, found := me.vars[name]
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return v, nil
}

type constAnalyzer struct{}

func (constAnalyzer) Analyze(env Env, form Any) (Expr, error) {
	return constExpr{Const: form}, nil
}

type constExpr struct{ Const Any }

func (ce constExpr) Eval(_ Env) (Any, error) { return ce.Const, nil }
