package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
	"github.com/spy16/slurp/reflector"
	"github.com/spy16/slurp/repl"
)

var globals = map[string]core.Any{
	"nil":     builtin.Nil{},
	"true":    builtin.Bool(true),
	"false":   builtin.Bool(false),
	"version": builtin.String("1.0"),

	// custom Go functions.
	"=": reflector.Func("=", builtin.Eq),
	"+": reflector.Func("sum", func(a ...int) int {
		sum := 0
		for _, item := range a {
			sum += item
		}
		return sum
	}),
	">": reflector.Func(">", func(a, b builtin.Int64) bool {
		return a > b
	}),
}

func main() {
	env := slurp.New()
	if err := env.Bind(globals); err != nil {
		fmt.Printf("bind failed: %+v\n", err)
		os.Exit(1)
	}

	r := repl.New(env,
		repl.WithBanner("Welcome to slurp!"),
		repl.WithPrompts(">>", " |"))

	if err := r.Loop(context.Background()); err != nil {
		log.Fatal(err)
	}
}
