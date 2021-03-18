// Package main builds on `simple` to demonstrate a naive implementation of Clojure's conj.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/repl"
)

var globals = map[string]core.Any{
	"nil":     builtin.Nil{},
	"true":    builtin.Bool(true),
	"false":   builtin.Bool(false),
	"version": builtin.String("1.0"),

	// custom Go functions.
	"=": slurp.Func("=", core.Eq),
	"+": slurp.Func("sum", func(a ...int) int {
		sum := 0
		for _, item := range a {
			sum += item
		}
		return sum
	}),
	">": slurp.Func(">", func(a, b builtin.Int64) bool {
		return a > b
	}),
	"conj": slurp.Func("conj", conj),
}

func conj(vs ...core.Any) (core.Any, error) {
	if len(vs) == 0 {
		return nil, errors.New("conj requires at least 1 argument")
	}

	cntr, vs := vs[0], vs[1:]

	rval := reflect.ValueOf(cntr)

	conjVal := rval.MethodByName("Conj")
	if conjVal.IsZero() {
		return nil, fmt.Errorf("type '%s' has no method Conj", rval.Type())
	}

	args := make([]reflect.Value, len(vs))
	for i, val := range vs {
		args[i] = reflect.ValueOf(val)
	}

	args = conjVal.Call(args)
	if args[1].IsNil() {
		return args[0].Interface(), nil
	}

	return nil, args[1].Interface().(error)
}

func main() {
	env := core.New(globals)
	eval := slurp.New(slurp.WithEnv(env))

	r := repl.New(eval,
		repl.WithBanner("Welcome to slurp!\nTry typing '(conj [] 1)'."),
		repl.WithPrompts(">>", " |"))

	if err := r.Loop(context.Background()); err != nil {
		log.Fatal(err)
	}
}
