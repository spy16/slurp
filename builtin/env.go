package builtin

import (
	"errors"
	"fmt"
	"sync"

	"github.com/spy16/slurp/core"
)

// Env is a global environment.
type Env struct {
	active core.Namespace
	scope  *globalScope

	mu *sync.RWMutex
	ns map[string]*globalScope
}

// NewEnv returns a root Env that can be used to execute forms.
func NewEnv(opt ...Option) core.Env {
	env := Env{mu: new(sync.RWMutex)}

	for _, option := range withDefault(opt) {
		option(&env)
	}

	return env
}

// Name returns the active namespace.
func (env Env) Name() string { return "<main>" }

// Namespace returns the active namespace.
func (env Env) Namespace() core.Namespace { return env.active }

// Parent is nil.
func (Env) Parent() core.Env { return nil }

// Scope returns the scope for the active namespace.
// Scope's methods are safe for concurrent access.
func (env Env) Scope() core.Scope { return parseScope(env) }

// MaybeScope returns the scope for the supplied namespace,
// it exists. Else, it returns a Scope whose methods return
// an error.
//
// It is intended for use inside of SymbolParser, in order to
// resolve relative symbols.
//
// To access the global scope without parsing, use
//     env.MaybeScope(env.Namespace())
func (env Env) MaybeScope(ns core.Namespace) core.Scope {
	name := ns.String()
	if name == env.active.String() {
		return env.scope
	}

	env.mu.RLock()
	defer env.mu.RUnlock()

	if scope, ok := env.ns[name]; ok {
		return scope
	}

	return errScope{
		Cause:   core.ErrNotFound,
		Message: name,
	}
}

func (env Env) WithNamespace(ns core.Namespace) core.Env {
	name := ns.String()
	if name == env.active.String() {
		return env
	}

	env.mu.Lock()
	defer env.mu.Unlock()

	var ok bool
	if env.scope, ok = env.ns[name]; !ok {
		env.scope = &globalScope{vars: map[string]core.Any{}}
		env.ns[name] = env.scope
	}

	// 'env' is a copy, so no need to allocate
	env.active = ns
	// env.scope.env = &env
	return env
}

// Child returns a new environment with 'env' as it's parent.
// The child environment is logically contained within the parent's
// active namespace.
//
// Note that the child's scope is NOT safe for concurrent access.
func (env Env) Child(name string, vars map[string]core.Any) core.Env {
	if vars == nil {
		vars = map[string]core.Any{}
	}

	return childEnv{
		name:   name,
		active: env.active,
		parent: env,
		scope:  vars,
	}
}

type childEnv struct {
	name   string
	active core.Namespace
	parent core.Env
	scope  localScope
}

func (env childEnv) Name() string              { return env.name }
func (env childEnv) Namespace() core.Namespace { return env.active }
func (env childEnv) Parent() core.Env          { return env.parent }
func (env childEnv) Scope() core.Scope         { return env.scope }

func (env childEnv) WithNamespace(ns core.Namespace) core.Env {
	// XXX:  what happens if we're already in the namespace?  Scope will be cleared.

	// TODO:  what is the expected behavior when calling something like this?
	//			(let []
	//				(ns foo))
	//
	return childEnv{
		name:   ns.String(),
		parent: env.parent.WithNamespace(ns),
		scope:  localScope{},
	}
}

func (env childEnv) Child(name string, vars map[string]core.Any) core.Env {
	if vars == nil {
		vars = map[string]core.Any{}
	}

	return childEnv{
		parent: env,
		scope:  vars,
	}
}

type Option func(*Env)

// WithNamespace sets the active namespace.
func WithNamespace(ns core.Namespace, vars map[string]core.Any) Option {
	if vars == nil {
		vars = map[string]core.Any{}
	}

	return func(env *Env) {
		env.active = ns
		env.scope = &globalScope{vars: vars}
		env.ns = map[string]*globalScope{ns.String(): env.scope}
	}
}

func withDefault(opt []Option) []Option {
	return append([]Option{
		WithNamespace("", nil),
	}, opt...)
}

// globalScope is safe for concurrent access.
type globalScope struct {
	/*
	 * TODO(performance):  investigate optimistic concurrency-control.
	 *
	 * Rationale is that access to the global scope is likely to be
	 * dominated by read operations, with relatively few writes.
	 * If so, it seems desirable for readers to never block, even in
	 * the presence of concurrent writers.
	 *
	 * The simplest solution would be a CAS loop on a map[string]core.Any.
	 * Although this would involve copying the map on each write, the
	 * overhead may be negligible.  A more complex solution would involve
	 * the use of immutable maps, but the constant factors are likely
	 * to be much higher.
	 */

	mu   sync.RWMutex
	vars map[string]core.Any
}

func (gs *globalScope) Bind(s core.Symbol, value core.Any) error {
	if !s.Valid() {
		return fmt.Errorf("%w: %s", core.ErrInvalidName, s)
	}

	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.vars[s.String()] = value
	return nil
}

func (gs *globalScope) Resolve(s core.Symbol) (core.Any, error) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if v, found := gs.vars[s.String()]; found {
		return v, nil
	}

	return nil, fmt.Errorf("%w: %s", core.ErrNotFound, s)
}

// localScope stores local variables.
type localScope map[string]core.Any

func (vars localScope) Bind(s core.Symbol, val core.Any) error {
	if !s.Valid() {
		return fmt.Errorf("%w: %s", core.ErrInvalidName, s)
	}

	vars[s.String()] = val
	return nil
}

func (vars localScope) Resolve(s core.Symbol) (core.Any, error) {
	if v, found := vars[s.String()]; found {
		return v, nil
	}

	return nil, fmt.Errorf("%w: %s", core.ErrNotFound, s)
}

type parseScope Env

func (p parseScope) parse(s core.Symbol) (core.Scope, core.Symbol) {
	if s.Qualified() {
		return Env(p).MaybeScope(s.Namespace()), s.Relative()
	}

	return Env(p).MaybeScope(Env(p).Namespace()), s
}

func (p parseScope) Bind(s core.Symbol, val core.Any) error {
	scope, s := p.parse(s)
	return scope.Bind(s, val)
}

func (p parseScope) Resolve(s core.Symbol) (core.Any, error) {
	scope, s := p.parse(s)
	return scope.Resolve(s)
}

// errScope wraps an error, and returns it immediately from
// its methods.
type errScope struct {
	Cause   error
	Message string
}

func (err errScope) Bind(core.Symbol, core.Any) error      { return err }
func (err errScope) Resolve(core.Symbol) (core.Any, error) { return nil, err }

func (err errScope) Error() string {
	if err.Cause == nil {
		return err.Message
	}

	return fmt.Sprintf("%s: %s", err.Cause, err.Message)
}

// Is returns true if the other error is same as the cause of this error.
func (err errScope) Is(other error) bool { return errors.Is(err.Cause, other) }
func (err errScope) Unwrap() error       { return err.Cause }
