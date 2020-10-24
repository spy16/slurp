# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.2.0 - 2020-10-24

### Added

- Support for Closures.
- `builtin` package with all builtin value types.
- `Fn` type to represent multi-arity functions/macros.
- `Analyzer` supports macro expansion.

### Changed

- `Env` uses an interface model.
- `Analyzer` is decoupled frm `Env`.
- Repository is re-organized around a `core` package.

## v0.1.1 - 2020-10-14

### Added

- Minor performance improvements in reflector funcWrapper.

## v0.1.0 - 2020-10-14

### Added

- `slurp` root package with builtin types and builtin Analyzer.
- special forms: `go`, `do`, `if`, `def`, `quote`
- `reader` package with reader macros for all builtin types.
- fully customisable `repl` package.
- `reflector` to simplify Go interop.
