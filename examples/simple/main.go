package main

import (
	"context"
	"log"

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
}

func main() {
	env := builtin.NewEnv(globals)
	eval := slurp.New(slurp.WithEnv(env))

	r := repl.New(eval,
		repl.WithBanner("Welcome to slurp!"),
		repl.WithPrompts(">>", " |"))

	if err := r.Loop(context.Background()); err != nil {
		log.Fatal(err)
	}
}
