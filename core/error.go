package core

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned by Env when a binding is not found
	// for a given symbol/name.
	ErrNotFound = errors.New("not found")

	// ErrInvalidName is returned by Env when the bind name is
	// invalid.
	ErrInvalidName = errors.New("invalid bind name")

	// ErrNotInvokable is returned by InvokeExpr when the target is
	// not invokable.
	ErrNotInvokable = errors.New("not invokable")
)

// Error is returned by all slurp operations. Cause indicates the underlying
// error type. Use errors.Is() with Cause to check for specific errors.
type Error struct {
	Message    string
	Cause      error
	Form       string
	Begin, End Position
}

// With returns a clone of the error with message set to given value.
func (e Error) With(msg string) Error {
	return Error{
		Cause:   e.Cause,
		Message: msg,
	}
}

// Is returns true if the other error is same as the cause of this error.
func (e Error) Is(other error) bool { return errors.Is(e.Cause, other) }

// Unwrap returns the underlying cause of the error.
func (e Error) Unwrap() error { return e.Cause }

func (e Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%v: %s", e.Cause, e.Message)
	}
	return e.Message
}

// Position represents the positional information about a value read
// by reader.
type Position struct {
	File string
	Ln   int
	Col  int
}

func (p Position) String() string {
	if p.File == "" {
		p.File = "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Ln, p.Col)
}
