package main

import (
	"context"
	"log"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/reflector"
	"github.com/spy16/slurp/repl"
)

var globals = map[string]slurp.Any{
	"nil":       slurp.Nil{},
	"true":      slurp.Bool(true),
	"false":     slurp.Bool(false),
	"*version*": slurp.String("1.0"),

	// custom Go functions.
	"+": reflector.Func("sum", func(a ...int) int {
		sum := 0
		for _, item := range a {
			sum += item
		}
		return sum
	}),
	">": reflector.Func(">", func(a, b slurp.Int64) bool {
		return a > b
	}),
}

func main() {
	env := slurp.New(slurp.WithGlobals(globals, nil))

	r := repl.New(env,
		repl.WithBanner("Welcome to slurp!"),
		repl.WithPrompts(">>", " |"),
	)

	if err := r.Loop(context.Background()); err != nil {
		log.Fatal(err)
	}
}
