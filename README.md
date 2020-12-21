# Slurp


[![GoDoc](https://godoc.org/github.com/spy16/slurp?status.svg)](https://godoc.org/github.com/spy16/slurp) [![Go Report Card](https://goreportcard.com/badge/github.com/spy16/slurp)](https://goreportcard.com/report/github.com/spy16/slurp) ![Go](https://github.com/spy16/slurp/workflows/Go/badge.svg?branch=master)

Slurp is a highly customisable, embeddable LISP toolkit for `Go` applications.

- [Slurp](#slurp)
  - [Why Slurp](#why-slurp)
  - [Features](#features)
  - [Usage](#usage)
  - [Extending](#extending)
    - [Reader](#reader)
    - [Evaluation](#evaluation)

## Why Slurp

![I've just received word that the Emperor has dissolved the MIT computer science program permanently.](https://imgs.xkcd.com/comics/lisp_cycles.png)

Slurp is for developers who want to design and embed an interpreted language inside of a Go program.

Slurp provides composable building-blocks that make it easy to design a custom lisp, even if you've never written an interpreter before.  Since Slurp is written in pure Go, your new language can be embedded into any Go program by importing it — just like any other library.

> **NOTE:**  Slurp is _NOT_ an implementation of a particular LISP dialect.
> 
> It provides pieces that can be used to build a LISP dialect or can be used as a scripting layer.

Slurp is designed around three core values:

1. **Simplicity:**  The library is small and has few moving parts.
2. **Batteries Included:**  There are no external dependencies and little configuration required.
3. **Go Interoperability:**  Slurp can call Go code and fully supports Go's concurrency features.

We hope that you will find Slurp to be powerful, useful and fun to use.  We look forward to seeing what you build with it!


## Features

* Highly customizable, safe and powerful reader/parser through
  a read table (Inspired by [Clojure](https://github.com/clojure/clojure/blob/master/src/jvm/clojure/lang/LispReader.java)) (See [Reader](#reader))
* Immutable datatypes including: nil, bool, string, int & float,
  character, keyword, symbol, list, vector & map.
* Multiple number formats supported: decimal, octal, hexadecimal,
  radix and scientific notations.
* Full unicode support. Symbols, keywords etc. can include unicode
  characters (Example: `find-δ`, `π` etc.) and `🧠`, `🏃` etc. (yes,
  smileys too).
* Character Literals with support for:
  1. simple literals  (e.g., `\a` for `a`)
  2. special literals (e.g., `\newline`, `\tab` etc.)
  3. unicode literals (e.g., `\u00A5` for `¥` etc.)
* Full interoperability with Go:  call native Go functions/libraries, and manipulate native Go datatypes from your language.
* Support for macros.
* Easy to extend. See [Extending](#extending).
* Tiny & powerful REPL package.
* Zero dependencies (outside of tests).

## Usage

Slurp requires Go 1.14 or higher.  It can be installed using `go get`:

```bash
go get -u github.com/spy16/slurp
```

First-time users should visit the ["Getting Started" Tutorial](https://github.com/spy16/slurp/wiki/Getting-Started)

What can you use it for?

1. Embedded script engine to provide dynamic behavior without requiring re-compilation of your application ([example](./examples/simple/main.go)).
2. Business rule engine exposing specific, composable rules ([example](./examples/rule-engine/main.go)).
3. To build DSLs.
4. To build your own LISP dialect ([example](https://github.com/wetware/ww)).

Refer [./examples](./examples) for more usage examples.

## Extending

### Reader

slurp reader is inspired by Clojure reader and uses a _read table_. Reader can be extended
to add new syntactical features by adding _reader macros_ to the _read table_. _Reader Macros_
are implementations of `reader.Macro` function type. All syntax that reader can read are
implemented using _Reader Macros_. Use `SetMacro()` method of `reader.Reader` to override or
add a custom reader or dispatch macro.

Reader returned by `reader.New(...)`, is configured to support following forms:

* Numbers:
  * Integers use `int64` Go representation and can be specified using decimal, binary
    hexadecimal or radix notations. (e.g., 123, -123, 0b101011, 0xAF, 2r10100, 8r126 etc.)
  * Floating point numbers use `float64` Go representation and can be specified using
    decimal notation or scientific notation. (e.g.: 3.1412, -1.234, 1e-5, 2e3, 1.5e3 etc.)
  * You can override number reader using `WithNumReader()`.
* Characters: Characters use `rune` or `uint8` Go representation and can be written in 3 ways:
  * Simple: `\a`, `\λ`, `\β` etc.
  * Special: `\newline`, `\tab` etc.
  * Unicode: `\u1267`
* Boolean: `true` or `false` are converted to `Bool` type.
* Nil: `nil` is represented as a zero-allocation empty struct in Go.
* Keywords: Keywords represent symbolic data and start with `:`. (e.g., `:foo`)
* Symbols: Symbols can be used to name a value and can contain any Unicode symbol.
* Lists: Lists are zero or more forms contained within parentheses. (e.g., `(1 2 3)`, `(1 :hello ())`).

### Evaluation

Slurp uses implementation of `core.Env` as the environment for evaluating
forms. A form is first analyzed using a `core.Analyzer` to produce `core.Expr`
that can be evaluated against an Env. Custom implementations for all of
these can be provided to optimise performance, customise evaluation rules
or support additional features.

A builtin Analyzer is provided which supports pure lisp forms with support
for macros as well.
