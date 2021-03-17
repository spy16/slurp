package repl

import (
	"bufio"
	"io"
	"strings"
)

// Input implementation is used by REPL to read user-input. See WithInput()
// REPL option to configure an Input.
type Input interface {
	Readline() (string, error)
}

// Prompter is an optional interface for Input.  If provided, the REPL
// will attempt to set a prompt during the read-phase.
type Prompter interface {
	Prompt(string)
}

// LineReader is a simple Input that uses bufio.Scanner.
type LineReader struct{ s *bufio.Scanner }

// NewLineReader that reads from r.
func NewLineReader(r io.Reader) LineReader {
	return LineReader{bufio.NewScanner(r)}
}

func (lr LineReader) Readline() (string, error) {
	if !lr.s.Scan() {
		if lr.s.Err() == nil { // scanner swallows EOF
			return lr.s.Text(), io.EOF
		}

		return "", lr.s.Err()
	}

	return lr.s.Text(), nil
}

// Prompt is an interactive Input that that can signals 'out' when it is
// ready to receive data.
type Prompt struct {
	in     Input
	out    io.Writer
	prompt *strings.Reader
}

func NewPrompt(in Input, out io.Writer) *Prompt {
	return &Prompt{
		in:     in,
		out:    out,
		prompt: &strings.Reader{},
	}
}

func (p *Prompt) Prompt(s string) { p.prompt.Reset(s) }

func (p *Prompt) Readline() (string, error) {
	if err := p.writePrompt(); err != nil {
		return "", err
	}

	return p.in.Readline()
}

func (p *Prompt) writePrompt() error {
	defer p.prompt.Seek(0, io.SeekStart)

	_, err := io.Copy(p.out, p.prompt)
	return err
}
