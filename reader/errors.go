package reader

import (
	"errors"
	"fmt"
)

var (
	// ErrSkip can be returned by reader macro to indicate a no-op form which
	// should be discarded (e.g., Comments).
	ErrSkip = errors.New("skip expr")

	// ErrEOF is returned by reader when stream ends prematurely to indicate
	// that more data is needed to complete the current form.
	ErrEOF = errors.New("unexpected EOF while parsing")

	// ErrUnmatchedDelimiter is returned when a reader macro encounters a closing
	// container- delimiter without a corresponding opening delimiter (e.g. ']'
	// but no '[').
	ErrUnmatchedDelimiter = errors.New("unmatched delimiter")

	// ErrNumberFormat is returned when a reader macro encounters a illegally
	// formatted numerical form.
	ErrNumberFormat = errors.New("invalid number format")
)

// Error is returned by the error when reading from a stream fails due to
// some issue. Use errors.Is() with Cause to check for specific underlying
// errors.
type Error struct {
	Form       string
	Cause      error
	Rune       rune
	Begin, End Position
}

// Is returns true if the other error is same as the cause of this error.
func (e Error) Is(other error) bool { return errors.Is(e.Cause, other) }

// Unwrap returns the underlying cause of the error.
func (e Error) Unwrap() error { return e.Cause }

func (e Error) Error() string {
	cause := e.Cause
	if errors.Is(cause, ErrUnmatchedDelimiter) {
		cause = fmt.Errorf("unmatched delimiter '%c'", e.Rune)
	}

	return fmt.Sprintf("error while reading: %v", cause)
}
