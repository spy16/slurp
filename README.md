# Slurp


[![GoDoc](https://godoc.org/github.com/spy16/slurp?status.svg)](https://godoc.org/github.com/spy16/slurp)
[![Go Report Card](https://goreportcard.com/badge/github.com/spy16/slurp)](https://goreportcard.com/report/github.com/spy16/slurp)
[![Code Change](https://github.com/spy16/slurp/actions/workflows/code_change.yml/badge.svg)](https://github.com/spy16/slurp/actions/workflows/code_change.yml)

Slurp is a highly customisable, embeddable LISP toolkit for `Go` applications.

- [Slurp](#slurp)
  - [Why Slurp](#why-slurp)
  - [Features](#features)
  - [Usage](#usage)
  - [Documentation](#documentation)

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
* Easy to extend. See [Wiki](https://github.com/spy16/slurp/wiki/Customizing-Syntax).
* Tiny & powerful REPL package.
* Zero dependencies (outside of tests).

## Usage

Slurp requires Go 1.14 or higher.  It can be installed using `go get`:

```bash
go get -u github.com/spy16/slurp
```

What can you use it for?

1. Embedded script engine to provide dynamic behavior without requiring re-compilation of your application ([example](./examples/simple/main.go)).
2. Business rule engine exposing specific, composable rules ([example](./examples/rule-engine/main.go)).
3. To build DSLs.
4. To build your own LISP dialect ([example](https://github.com/wetware/ww)).

Refer [./examples](./examples) for more usage examples.

## Documentation

In addition to the GoDocs, we maintain in-depth tutorials on the GitHub wiki.  The following
pages are good starting points:

1. [Getting Started](https://github.com/spy16/slurp/wiki/Getting-Started)
2. [Customizing Syntax](https://github.com/spy16/slurp/wiki/Customizing-Syntax#Custom-Parsing)
3. [Customizing Evaluation](https://github.com/spy16/slurp/wiki/Customizing-Syntax#Custom-Evaluation)
