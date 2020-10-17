package core

import (
	"fmt"
	"strings"
	"sync"
)

const rootEnv = "<main>"

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

type mapEnv struct {
	parent Env
	name   string
	mu     sync.RWMutex
	vars   map[string]Any
}

// Name should return the name of the frame.
func (me *mapEnv) Name() string { return me.name }

// Parent should return the parent/outer env of this env.
func (me *mapEnv) Parent() Env { return me.parent }

// Child should return a new env with given frame name and vars
// bound. Returned env should have this env as parent/outer.
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

// Bind should create a global binding with given name and value.
func (me *mapEnv) Bind(name string, val Any) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("%w: %s", ErrInvalidBindName, name)
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

// Resolve should resolve the symbol and return its value. Lookup
// should be done in locals first (this env) and then done in root
// env.
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
