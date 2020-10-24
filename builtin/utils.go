package builtin

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spy16/slurp/core"
)

const rootEnv = "<main>"

// New returns a root Env that can be used to execute forms.
func New(globals map[string]core.Any) core.Env {
	if globals == nil {
		globals = map[string]core.Any{}
	}
	return &mapEnv{
		parent: nil,
		name:   rootEnv,
		vars:   globals,
	}
}

// Eval performs syntax analysis of the given form to produce an Expr
// and evalautes the Expr against the given Env.
func Eval(env core.Env, analyzer core.Analyzer, form core.Any) (core.Any, error) {
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
func Root(env core.Env) core.Env {
	for env.Parent() != nil {
		env = env.Parent()
	}
	return env
}

// mapEnv implements Env using a Go native map and RWMutex.
type mapEnv struct {
	parent core.Env
	name   string
	mu     sync.RWMutex
	vars   map[string]core.Any
}

func (me *mapEnv) Name() string     { return me.name }
func (me *mapEnv) Parent() core.Env { return me.parent }

func (me *mapEnv) Child(name string, vars map[string]core.Any) core.Env {
	if vars == nil {
		vars = map[string]core.Any{}
	}
	return &mapEnv{
		name:   name,
		parent: me,
		vars:   vars,
	}
}

func (me *mapEnv) Bind(name string, val core.Any) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: %s", core.ErrInvalidName, name)
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

func (me *mapEnv) Resolve(name string) (core.Any, error) {
	if me.parent == nil {
		// only root env is shared between threads. so make sure
		// concurrent accesses are synchronized.
		me.mu.RLock()
		defer me.mu.RUnlock()
	}

	v, found := me.vars[name]
	if !found {
		return nil, fmt.Errorf("%w: %s", core.ErrNotFound, name)
	}
	return v, nil
}

type constAnalyzer struct{}

func (constAnalyzer) Analyze(_ core.Env, form core.Any) (core.Expr, error) {
	return constExpr{Const: form}, nil
}

type constExpr struct{ Const core.Any }

func (ce constExpr) Eval(env core.Env) (core.Any, error) { return ce.Const, nil }
