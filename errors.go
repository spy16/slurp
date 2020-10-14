package slurp

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound is returned when a binding not found.
	ErrNotFound = errors.New("not found")

	// ErrInvalidBindName is returned by DefExpr when the bind name is
	// invalid.
	ErrInvalidBindName = errors.New("invalid name for def")

	// ErrNotInvokable is returned by InvokeExpr when the target is not
	// invokable.
	ErrNotInvokable = errors.New("not invokable")

	// ErrIncomparable is returned by Any.Comp when a comparison between
	// two tpyes is undefined.  Users should generally consider the types
	// to be not equal in such cases, but not assume any ordering.
	ErrIncomparable = errors.New("incomparable types")
)

// Error is returned by all slurp operations. Cause indicates the underlying
// error type. Use errors.Is() with Cause to check for specific errors.
type Error struct {
	Message string
	Cause   error
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
