package slurp

import (
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

// Option values can be used with New() to customise slurp instance
// during initialisation.
type Option func(ins *Instance)

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
			a = builtin.NewAnalyzer()
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
