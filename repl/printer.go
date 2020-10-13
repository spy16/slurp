package repl

import (
	"fmt"
	"io"
)

// Printer can print arbitrary values to output.
type Printer interface {
	Fprintln(w io.Writer, val interface{}) error
}

// BasicPrinter prints the value using fmt.Println.  It applies no special formatting.
type BasicPrinter struct{}

// Fprintln prints val to w.
func (p BasicPrinter) Fprintln(w io.Writer, val interface{}) error {
	_, err := fmt.Fprintln(w, val)
	return err
}

// Renderer pretty-prints the value.  It checks if the value implements any of the
// following interfaces (in decreasing order of preference) before printing the default
// Go value.
//
//  1. fmt.Formatter
//  2. fmt.Stringer
//  3. SExpr
type Renderer struct{}

// SExpr can render a parseable s-expression
type SExpr interface {
	// SExpr returns a string representation of the value, suitable for parsing with
	// reader.Reader.
	SExpr() (string, error)
}

// Fprintln prints val to w.
func (r Renderer) Fprintln(w io.Writer, val interface{}) (err error) {
	switch v := val.(type) {
	case fmt.Formatter:
		_, err = fmt.Fprintf(w, "%+v\n", v)
	case fmt.Stringer:
		_, err = fmt.Fprintf(w, "%s\n", v)
	case SExpr:
		var s string
		if s, err = v.SExpr(); err == nil {
			_, err = fmt.Fprintln(w, s)
		}
	default:
		_, err = fmt.Fprintf(w, "%v\n", val)
	}

	return
}
