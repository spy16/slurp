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

	// ErrNumberFormat is returned when a reader macro encounters a illegally
	// formatted numerical form.
	ErrNumberFormat = errors.New("invalid number format")
)

// Error is returned by the error when reading from a stream fails due to
// some issue. Use errors.Is() with Cause to check for specific underlying
// errors.
type Error struct {
	Cause      error
	Begin, End Position
}

// Is returns true if the other error is same as the cause of this error.
func (e Error) Is(other error) bool { return errors.Is(e.Cause, other) }

// Unwrap returns the underlying cause of the error.
func (e Error) Unwrap() error { return e.Cause }

func (e Error) Error() string { return fmt.Sprintf("ReaderError: %v", e.Cause) }

func (e Error) Format(s fmt.State, verb rune) {
	if s.Flag('#') {
		/*
		* File "<REPL>" line 1, column 2
		* ReaderError:  unmatched delimiter ']'
		 */
		fmt.Fprintf(s, "File \"%s\" line %d, column %d\n",
			e.Begin.File,
			e.Begin.Ln,
			e.Begin.Col)
	}

	fmt.Fprint(s, e.Error())
}

type unmatchedDelimiterError rune

func (r unmatchedDelimiterError) Error() string {
	return fmt.Sprintf("unmatched delimiter '%c'", r)
}
