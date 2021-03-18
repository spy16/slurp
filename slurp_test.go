package slurp_test

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/repl"
)

var testdir string = "./test"

func init() {
	flag.StringVar(&testdir, "testdir", testdir, "root test directory (default: ./test)")
}

func TestSlurp(t *testing.T) {
	t.Parallel()
	t.Helper()

	if err := filepath.Walk(testdir, visit(t)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func visit(t *testing.T) filepath.WalkFunc {
	return func(path string, info fs.FileInfo, err error) error {
		t.Helper()

		if err != nil {
			t.Logf("skipping %s:  %s", path, err)
		} else if !info.IsDir() {
			t.Run(path, run(t, path, info))
		}

		return nil
	}
}

func run(t *testing.T, path string, file fs.FileInfo) func(*testing.T) {
	return func(t *testing.T) {
		f, err := os.Open(path)
		if err != nil {
			t.Errorf("open file:  %s", err)
		}
		defer f.Close()

		r := repl.New(evaluator(t),
			repl.WithInput(repl.NewLineReader(f), nil),
			repl.WithPrinter(printer{t}))

		if err := r.Loop(context.Background()); err != nil {
			t.Error(err)
		}
	}
}

func evaluator(t *testing.T) repl.Evaluator {
	env := core.New(map[string]core.Any{
		"nil?": slurp.Func("is-nil", builtin.IsNil),
		"not": slurp.Func("not", func(any core.Any) bool {
			return builtin.IsTruthy(any) == false
		}),
		"=": slurp.Func("equal", core.Eq),
		"!=": slurp.Func("not-equal", func(a, b core.Any) (bool, error) {
			eq, err := core.Eq(a, b)
			return !eq, err
		}),
		"<": slurp.Func("less-than", func(a, b core.Any) (bool, error) {
			i, err := core.Compare(a, b)
			return i < 0, err
		}),
		">": slurp.Func("greater-than", func(a, b core.Any) (bool, error) {
			i, err := core.Compare(a, b)
			return i > 0, err
		}),
		"<=": slurp.Func("less-than-or-eq", func(a, b core.Any) (bool, error) {
			i, err := core.Compare(a, b)
			return i <= 0, err
		}),
		">=": slurp.Func("greater-than-or-eq", func(a, b core.Any) (bool, error) {
			i, err := core.Compare(a, b)
			return i >= 0, err
		}),
	})

	return slurp.New(slurp.WithEnv(env))
}

type printer struct{ t *testing.T }

func (p printer) Print(v interface{}) error {
	if err, ok := v.(error); ok {
		p.t.Errorf("runtime error: %v", err)
	} else if !builtin.IsTruthy(v) {
		p.t.Fail()
	}

	return nil
}
