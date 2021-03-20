package repl

import (
	"fmt"
	"io"
	"os"
)

// Printer can print arbitrary values to output.
type Printer interface {
	Print(interface{}) error
}

// BasicPrinter prints the value using fmt.Println.
// It applies no special formatting.
type BasicPrinter struct{ Out, Err io.Writer }

// Print v.
func (p *BasicPrinter) Print(v interface{}) error {
	if p.Out == nil {
		p.Out = os.Stdout
		p.Err = os.Stderr
	}

	if err, ok := v.(error); ok {
		_, err := fmt.Fprintln(p.Err, err)
		return err
	}

	_, err := fmt.Fprintln(p.Out, v)
	return err
}

// Renderer pretty-prints the value.
type Renderer struct{ Out, Err io.Writer }

// Print prints val to w.
func (r *Renderer) Print(val interface{}) (err error) {
	if r.Out == nil {
		r.Out = os.Stdout
		r.Err = os.Stderr
	}

	switch val.(type) {
	case error:
		_, err = fmt.Fprintf(r.Err, "%#s\n", val)
	default:
		// Values should be represented as an s-expression
		// to avoid ambiguity between similar values (e.g. symbol vs string)
		_, err = fmt.Fprintf(r.Out, "%s\n", val)
	}

	return
}
