package builtin

import (
	"fmt"
	"strings"
	"sync"

	"github.com/spy16/slurp/core"
)

const rootEnv = "<main>"

// NewEnv returns a root Env that can be used to execute forms.
func NewEnv(globals map[string]core.Any) *Env {
	if globals == nil {
		globals = map[string]core.Any{}
	}
	return &Env{
		parent: nil,
		name:   rootEnv,
		vars:   globals,
	}
}

// Env implements Env using a Go native map and RWMutex.
type Env struct {
	parent core.Env
	name   string
	mu     sync.RWMutex
	vars   map[string]core.Any
}

func (env *Env) Name() string     { return env.name }
func (env *Env) Parent() core.Env { return env.parent }

func (env *Env) Child(name string, vars map[string]core.Any) core.Env {
	if vars == nil {
		vars = map[string]core.Any{}
	}
	return &Env{
		name:   name,
		parent: env,
		vars:   vars,
	}
}

func (env *Env) Bind(name string, val core.Any) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: %s", core.ErrInvalidName, name)
	}

	if env.parent == nil {
		// only root env is shared between threads. so make sure
		// concurrent accesses are synchronized.
		env.mu.Lock()
		defer env.mu.Unlock()
	}

	env.vars[name] = val
	return nil
}

func (env *Env) Resolve(name string) (core.Any, error) {
	if env.parent == nil {
		// only root env is shared between threads. so make sure
		// concurrent accesses are synchronized.
		env.mu.RLock()
		defer env.mu.RUnlock()
	}

	v, found := env.vars[name]
	if !found {
		return nil, fmt.Errorf("%w: %s", core.ErrNotFound, name)
	}
	return v, nil
}
