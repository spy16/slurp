package slurp

import (
	"bytes"

	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// New returns a new slurp interpreter session.
func New(opts ...Option) *Instance {
	buf := bytes.Buffer{}
	ins := &Instance{
		buf:    &buf,
		reader: reader.New(&buf),
	}

	for _, opt := range withDefaults(opts) {
		opt(ins)
	}

	return ins
}

// Option values can be used with New() to customise slurp instance
// during initialisation.
type Option func(ins *Instance)

// Instance represents a Slurp interpreter session.
type Instance struct {
	env      core.Env
	buf      *bytes.Buffer
	reader   *reader.Reader
	analyzer core.Analyzer
}

// Eval performs syntax analysis of the given form to produce an Expr
// and evalautes the Expr for result.
func (ins *Instance) Eval(form core.Any) (core.Any, error) {
	return core.Eval(ins.env, ins.analyzer, form)
}

// EvalStr reads forms from the given straing and evalautes it for
// result.
func (ins *Instance) EvalStr(s string) (core.Any, error) {
	if _, err := ins.buf.WriteString(s); err != nil {
		return nil, err
	}

	f, err := ins.reader.All()
	if err != nil {
		return nil, err
	}

	do, err := core.Cons(core.Symbol("do"), core.NewList(f...))
	if err != nil {
		return nil, err
	}

	return ins.Eval(do)
}

// Bind can be used to set global bindings that will be available while
// executing forms.
func (ins *Instance) Bind(vals map[string]core.Any) error {
	for k, v := range vals {
		if err := ins.env.Bind(k, v); err != nil {
			return err
		}
	}
	return nil
}

// WithEnv sets the environment to be used by the slurp instance. If
// env is nil, the default map-env will be used.
func WithEnv(env core.Env) Option {
	return func(ins *Instance) {
		if env == nil {
			env = core.New(nil)
		}
		ins.env = env
	}
}

// WithAnalyzer sets the analyzer to be used by the slurp instance for
// syntax analysis and macro expansions. If nil, uses builtin analyzer
// with standard special forms.
func WithAnalyzer(a core.Analyzer) Option {
	return func(ins *Instance) {
		if a == nil {
			a = NewAnalyzer()
		}
		ins.analyzer = a
	}
}

func withDefaults(opts []Option) []Option {
	return append([]Option{
		WithAnalyzer(nil),
		WithEnv(nil),
	}, opts...)
}
