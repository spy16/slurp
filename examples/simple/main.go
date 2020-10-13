package main

import (
	"context"
	"log"

	"github.com/spy16/slurp"
	"github.com/spy16/slurp/repl"
)

var globals = map[string]slurp.Any{
	"nil":       slurp.Nil{},
	"true":      slurp.Bool(true),
	"false":     slurp.Bool(false),
	"*version*": slurp.String("1.0"),
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
