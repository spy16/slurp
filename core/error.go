package core

import (
	"errors"
	"fmt"
)

// Error is returned by all slurp operations. Cause indicates the underlying
// error type. Use errors.Is() with Cause to check for specific errors.
type Error struct {
	Cause   error
	Message string
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
		return fmt.Sprintf("EvalError: %v: %s", e.Cause, e.Message)
	}

	return fmt.Sprintf("EvalError: %s", e.Message)
}

func (e Error) Format(s fmt.State, verb rune) {
	if !s.Flag('#') {
		fmt.Fprint(s, e.Error())
	}

	// TODO:  render the offending form.
	fmt.Fprint(s, e.Error())
}
