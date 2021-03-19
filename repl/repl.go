// Package repl provides facilities to build an interactive REPL using slurp.
package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reader"
)

// Evaluator implementation is responsible for executing givenn forms.
type Evaluator interface {
	CurrentNS() string
	Eval(forms ...core.Any) (core.Any, error)
}

// REPL implements a read-eval-print loop for a generic Runtime.
type REPL struct {
	eval        Evaluator
	input       Input
	output      Printer
	mapInputErr ErrMapper
	factory     ReaderFactory

	banner string

	prompter    Prompter
	prompt      string
	multiPrompt string
}

// New returns a new instance of REPL with given slurp Env. Option values
// can be used to configure REPL input, output etc.
func New(eval Evaluator, opts ...Option) *REPL {
	repl := &REPL{eval: eval}
	for _, option := range withDefaults(opts) {
		option(repl)
	}

	repl.prompter, _ = repl.input.(Prompter)

	return repl
}

// Loop starts the read-eval-print loop. Loop runs until context is cancelled
// or input stream returns an irrecoverable error (See WithInput()).
func (repl *REPL) Loop(ctx context.Context) error {
	repl.printBanner()
	repl.setPrompt(false)

	for ctx.Err() == nil {
		err := repl.readEvalPrint()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}
	}

	return ctx.Err()
}

// readEval reads one form from the input, evaluates it and prints the result.
func (repl *REPL) readEvalPrint() error {
	forms, err := repl.read()
	if err != nil {
		switch err.(type) {
		case reader.Error:
			_ = repl.output.Print(err)
		default:
			return err
		}
	}

	if len(forms) == 0 {
		return nil
	}

	res, err := evalAll(repl.eval, forms)
	if err != nil {
		return repl.output.Print(err)
	}
	if len(res) == 0 {
		return repl.output.Print(nil)
	}

	return repl.output.Print(res[len(res)-1])
}

func (repl *REPL) read() ([]core.Any, error) {
	var src string
	lineNo := 1

	for {
		repl.setPrompt(lineNo > 1)

		line, err := repl.input.Readline()
		if err = repl.mapInputErr(err); err != nil {
			return nil, err
		}

		src += line + "\n"

		if strings.TrimSpace(src) == "" {
			return nil, nil
		}

		rd := repl.factory.NewReader(strings.NewReader(src))
		rd.File = "<REPL>"

		form, err := rd.All()
		if err != nil {
			if errors.Is(err, reader.ErrEOF) {
				lineNo++
				continue
			}

			return nil, err
		}

		return form, nil
	}
}

func (repl *REPL) setPrompt(multiline bool) {
	if repl.prompter == nil || repl.prompt == "" {
		return
	}

	nsPrefix := repl.eval.CurrentNS()
	prompt := repl.prompt

	if multiline {
		nsPrefix = strings.Repeat(" ", len(nsPrefix)+1)
		prompt = repl.multiPrompt
	}

	repl.prompter.Prompt(fmt.Sprintf("%s%s ", nsPrefix, prompt))
}

func (repl *REPL) printBanner() {
	if repl.banner != "" {
		fmt.Println(repl.banner)
	}
}

func evalAll(eval Evaluator, vals []core.Any) ([]core.Any, error) {
	res := make([]core.Any, 0, len(vals))
	for _, form := range vals {
		form, err := eval.Eval(form)
		if err != nil {
			return nil, err
		}
		res = append(res, form)
	}
	return res, nil
}
