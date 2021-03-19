package core

type Symbol interface {
	// String returns the symbol's resolution path.
	String() string

	// Namespace returns a valid namespace for a fully-qualified
	// symbol.  If the symbol is not fully qualified, Namespace
	// panics.
	Namespace() Namespace

	// Valid returns true if the symbol is valid.
	Valid() bool

	// Qualified returns true if the symbol's resolution
	// path is fully-qualified.
	Qualified() bool

	// Relative returns a relative symbol.  If Qualified is
	// true, the implementation should return an equivalent
	// value.
	Relative() Symbol

	// WithNamespace returns a fully-qualified symbol with the
	// supplied namespace.  If the symbol is fully-qualified,
	// implementations should replace it with ns.
	WithNamespace(ns Namespace) Symbol
}
