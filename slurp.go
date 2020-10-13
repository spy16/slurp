package slurp

import (
	"sync"
)

// New returns a new root environment initialised based on given options.
func New(opts ...Option) *Env {
	env := &Env{globals: newMutexMap()}
	for _, opt := range withDefaults(opts) {
		opt(env)
	}
	return env
}

// Analyzer implementation is responsible for performing syntax analysis
// on given form.
type Analyzer interface {
	// Analyze should perform syntax checks for special forms etc. and
	// return Expr values that can be evaluated against a context.
	Analyze(env *Env, form Any) (Expr, error)
}

// ConcurrentMap is used by the Env to store variables in the global stack
// frame.
type ConcurrentMap interface {
	// Store should store the key-value pair in the map.
	Store(key string, val Any)

	// Load should return the value associated with the key if it exists.
	// Returns nil, false otherwise.
	Load(key string) (Any, bool)

	// Map should return a native Go map of all key-values in the concurrent
	// map. This can be used for iteration etc.
	Map() map[string]Any
}

// Env represents the environment in which forms are analysed and executed
// for results. Env is not safe for concurrent use.
type Env struct {
	maxDepth int
	stack    []stackFrame
	globals  ConcurrentMap
	analyzer Analyzer
}

// Eval performs syntax analysis on the given form, converts it into an
// Expr value using the configured Analyzer and evaluates it against the
// env for result.
func (env *Env) Eval(form Any) (Any, error) {
	// TODO: extract position info from 'form' and store it in Env for
	//       error annotation.
	expr, err := env.Analyze(form)
	if err != nil {
		return nil, err
	} else if expr == nil {
		return nil, nil
	}
	return expr.Eval()
}

// Resolve returns the value bound for the given symbol in the env. If the
// symbol is not found, returns 'nil'.
func (env Env) Resolve(sym string) Any {
	if len(env.stack) > 0 {
		// check inside top of the stack for local bindings.
		top := env.stack[len(env.stack)-1]
		if v, found := top.Vars[sym]; found {
			return v
		}
	}
	// return the value from global bindings if found.
	v, _ := env.globals.Load(sym)
	return v
}

// Analyze performs syntax checks for special forms etc. and returns an Expr
// value that can be evaluated against the env. Analyze might also do macro
// expansions.
func (env *Env) Analyze(form Any) (Expr, error) {
	if expr, ok := form.(Expr); ok {
		// Already an Expr, nothing to do.
		return expr, nil
	}
	return env.analyzer.Analyze(env, form)
}

// Fork creates a child context from Env and returns it. The child context
// can be used as context for an independent thread of execution.
func (env *Env) Fork() *Env {
	// TODO: we should create a new globals map and maintain a ref to the
	//       parent Env. this allows us to support thread-local bindings.
	return &Env{
		globals:  env.globals,
		analyzer: env.analyzer,
		maxDepth: env.maxDepth,
	}
}

func (env *Env) push(frame stackFrame) {
	env.stack = append(env.stack, frame)
}

func (env *Env) pop() (frame *stackFrame) {
	if len(env.stack) == 0 {
		panic("pop from empty stack")
	}
	frame, env.stack = &env.stack[len(env.stack)-1], env.stack[:len(env.stack)-1]
	return frame
}

func (env *Env) setGlobal(key string, value Any) {
	env.globals.Store(key, value)
}

type stackFrame struct {
	Name string
	Args []Any
	Vars map[string]Any
}

func newMutexMap() ConcurrentMap { return &mutexMap{} }

// mutexMap implements a simple ConcurrentMap using sync.RWMutex locks.
// Zero value is ready for use.
type mutexMap struct {
	sync.RWMutex
	vs map[string]Any
}

func (m *mutexMap) Load(name string) (v Any, ok bool) {
	m.RLock()
	defer m.RUnlock()
	v, ok = m.vs[name]
	return
}

func (m *mutexMap) Store(name string, v Any) {
	m.Lock()
	defer m.Unlock()

	if m.vs == nil {
		m.vs = map[string]Any{}
	}
	m.vs[name] = v
}

func (m *mutexMap) Map() map[string]Any {
	m.RLock()
	defer m.RUnlock()

	native := make(map[string]Any, len(m.vs))
	for k, v := range m.vs {
		native[k] = v
	}

	return native
}
