// Package slurp provides an Evaluator that composes builtin implementations of
// Env, Analyzer and Reader to produce a minimal LISP dialect.
package slurp

import (
	"bytes"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// Evaluator represents a Slurp Evaluator session.
type Evaluator struct {
	env      core.Env
	buf      *bytes.Buffer
	reader   *reader.Reader
	analyzer core.Analyzer
	ns       core.NamespaceProvider
}

// New slurp evaluator.
func New(opt ...Option) *Evaluator {
	var buf bytes.Buffer
	eval := &Evaluator{
		buf:    &buf,
		reader: reader.New(&buf),
	}

	for _, opt := range withDefaults(opt) {
		opt(eval)
	}

	return eval
}

// Namespace returns the current namespace.  It returns an empty string
// if the environment does not support namespacing.
func (eval Evaluator) Namespace() string { return eval.ns.Namespace() }

// Eval performs syntax analysis for each of the given forms and evaluates
// the resulting Exprs for a result.  If more than one form is supplied,
// it returns the result of the last Expr, or any error encountered along
// the way.
func (eval Evaluator) Eval(forms ...core.Any) (res core.Any, err error) {
	for _, form := range forms {
		if res, err = core.Eval(eval.env, eval.analyzer, form); err != nil {
			break
		}
	}

	return
}

// EvalStr reads forms from the given string, evaluates them, and returns
// the final form's value (or any error encountered along the way).
func (eval Evaluator) EvalStr(s string) (core.Any, error) {
	if _, err := eval.buf.WriteString(s); err != nil {
		return nil, err
	}

	fs, err := eval.reader.All()
	if err != nil {
		return nil, err
	}

	return eval.Eval(fs...)
}

// Option values can be used with New() to customise slurp instance
// during initialisation.
type Option func(eval *Evaluator)

// WithEnv sets the environment to be used by the slurp instance. If
// env is nil, the default map-env will be used.
func WithEnv(env core.Env) Option {
	if env == nil {
		env = builtin.NewEnv(nil)
	}

	ns, ok := env.(core.NamespaceProvider)
	if !ok {
		ns = nopNS{}
	}

	return func(eval *Evaluator) {
		eval.env = env
		eval.ns = ns
	}
}

// WithAnalyzer sets the analyzer to be used by the slurp instance for
// syntax analysis and macro expansions. If nil, uses builtin analyzer
// with standard special forms.
func WithAnalyzer(a core.Analyzer) Option {
	return func(eval *Evaluator) {
		if a == nil {
			a = &builtin.Analyzer{
				Specials: map[string]builtin.ParseSpecial{
					"go":    parseGo,
					"do":    parseDo,
					"if":    parseIf,
					"fn":    parseFn,
					"def":   parseDef,
					"let":   parseLet,
					"macro": parseMacro,
					"quote": parseQuote,
				},
			}
		}
		eval.analyzer = a
	}
}

func withDefaults(opts []Option) []Option {
	return append([]Option{
		WithAnalyzer(nil),
		WithEnv(nil),
	}, opts...)
}

type nopNS struct{}

func (nopNS) Namespace() string { return "" }
