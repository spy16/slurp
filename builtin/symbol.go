package builtin

import (
	"strings"

	"github.com/spy16/slurp/core"
)

var (
	_ core.Any = (*fq)(nil)
	_ core.Any = (*rel)(nil)

	_ core.Symbol = (*fq)(nil)
	_ core.Symbol = (*rel)(nil)

	_ core.EqualityProvider = (*fq)(nil)
	_ core.EqualityProvider = (*rel)(nil)
)

const SymbolSep = "."

// Symbol represents a lisp symbol Value.
func Symbol(s string) core.Symbol {
	if s = strings.TrimSpace(s); !isQualified(s) {
		return rel(s)
	}

	return fq(strings.SplitN(strings.Trim(s, SymbolSep), SymbolSep, 2))
}

func isQualified(s string) bool {
	// shortest valid fully-qualified symbol is ".x.y"
	if len(s) < 4 || s[0] != '.' {
		return false
	}

	// ".foo"
	if strings.Count(s, SymbolSep) < 2 {
		return false
	}

	return true
}

type fq []string

func (s fq) String() string                              { return "." + strings.Join(s, SymbolSep) }
func (s fq) Valid() bool                                 { return len(s) == 2 && s.Relative().Valid() }
func (s fq) Namespace() core.Namespace                   { return core.Namespace(s[0]) }
func (fq) Qualified() bool                               { return true }
func (s fq) Relative() core.Symbol                       { return rel(s[1]) }
func (s fq) WithNamespace(ns core.Namespace) core.Symbol { return fq{ns.String(), s[1]} }

func (s fq) Equals(other core.Any) (bool, error) {
	o, ok := other.(core.Symbol)
	return ok && o.Valid() && s.String() == o.String(), nil
}

type rel string

func (s rel) String() string                              { return string(s) }
func (s rel) Valid() bool                                 { return len(s) > 0 }
func (rel) Namespace() core.Namespace                     { panic("relative symbol") }
func (rel) Qualified() bool                               { return false }
func (s rel) Relative() core.Symbol                       { return s }
func (s rel) WithNamespace(ns core.Namespace) core.Symbol { return fq{ns.String(), s.String()} }

func (s rel) Equals(other core.Any) (bool, error) {
	o, ok := other.(core.Symbol)
	return ok && o.Valid() && string(s) == o.String(), nil
}
