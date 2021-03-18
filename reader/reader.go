// Package reader implements a highly customizable and generic reader, reader-macros
// for primitive runtime types and reader-macro factories for collection types like
// vector, map, set.
package reader

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

const dispatchTrigger = '#'

var (
	escapeMap = map[rune]rune{
		'"':  '"',
		'n':  '\n',
		'\\': '\\',
		't':  '\t',
		'a':  '\a',
		'f':  '\a',
		'r':  '\r',
		'b':  '\b',
		'v':  '\v',
	}

	charLiterals = map[string]rune{
		"tab":       '\t',
		"space":     ' ',
		"newline":   '\n',
		"return":    '\r',
		"backspace": '\b',
		"formfeed":  '\f',
	}
)

// New returns a lisp reader instance which can read forms from r. Returned instance
// supports only primitive data  types from value package by default. Support for
// custom forms can be added using SetMacro(). File name is inferred from the value &
// type information of 'r' OR can be set manually on the Reader instance returned.
func New(r io.Reader, opts ...Option) *Reader {
	rd := &Reader{
		File: inferFileName(r),
		rs:   bufio.NewReader(r),
		macros: map[rune]Macro{
			'"':  readString,
			';':  readComment,
			':':  readKeyword,
			'\\': readCharacter,
			'(':  readList,
			')':  UnmatchedDelimiter(),
			'[':  readVector,
			']':  UnmatchedDelimiter(),
			'\'': quoteFormReader("quote"),
			'~':  quoteFormReader("unquote"),
			'`':  quoteFormReader("syntax-quote"),
		},
		dispatch: map[rune]Macro{},
	}

	for _, option := range withDefaults(opts) {
		option(rd)
	}

	return rd
}

// Reader consumes characters from a stream and parses them into symbolic expressions
// or forms. Reader is customizable through Macro implementations which can be set as
// handler for specific trigger runes.
type Reader struct {
	File string

	rs                   io.RuneReader
	buf                  []rune
	line, col            int
	lastCol              int
	dispatching          bool
	dispatch             map[rune]Macro
	macros               map[rune]Macro
	numReader, symReader Macro
}

// All consumes characters from stream until EOF and returns a list of all the forms
// parsed. Any no-op forms (e.g., comment) will not be included in the result.
func (rd *Reader) All() ([]core.Any, error) {
	var forms []core.Any

	for {
		form, err := rd.One()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		forms = append(forms, form)
	}

	return forms, nil
}

// One consumes characters from underlying stream until a complete form is parsed and
// returns the form while ignoring the no-op forms like comments. Except EOF, all other
// errors will be wrapped with reader Error type along with the positional information
// obtained using Position().
func (rd *Reader) One() (core.Any, error) {
	for {
		form, err := rd.readOne()
		if err != nil {
			if errors.Is(err, ErrSkip) {
				continue
			}
			return nil, err
		}
		return form, nil
	}
}

// IsTerminal returns true if the rune should terminate a form. Macro trigger runes
// defined in the read table and all whitespace characters are considered terminal.
// "," is also considered a whitespace character and hence a terminal.
func (rd *Reader) IsTerminal(r rune) bool {
	if isSpace(r) {
		return true
	}

	if rd.dispatching {
		_, found := rd.dispatch[r]
		if found {
			return true
		}
	}

	_, found := rd.macros[r]
	return found
}

// SetMacro sets the given reader macro as the handler for init rune in the read table.
// Overwrites if a macro is already present. If the macro value given is nil, entry for
// the init rune will be removed from the read table. isDispatch decides if the macro is
// a dispatch macro and takes effect only after a '#' sign.
func (rd *Reader) SetMacro(init rune, isDispatch bool, macro Macro) {
	if isDispatch {
		if macro == nil {
			delete(rd.dispatch, init)
			return
		}
		rd.dispatch[init] = macro
	} else {
		if macro == nil {
			delete(rd.macros, init)
			return
		}
		rd.macros[init] = macro
	}
}

// NextRune returns next rune from the stream and advances the stream.
func (rd *Reader) NextRune() (rune, error) {
	var r rune
	if len(rd.buf) > 0 {
		r = rd.buf[0]
		rd.buf = rd.buf[1:]
	} else {
		temp, _, err := rd.rs.ReadRune()
		if err != nil {
			return -1, err
		}
		r = temp
	}

	if r == '\n' {
		rd.line++
		rd.lastCol = rd.col
		rd.col = 0
	} else {
		rd.col++
	}
	return r, nil
}

// Unread returns runes consumed from the stream back to the stream. Un-reading more
// runes than read is guaranteed to work but will cause inconsistency in  positional
// information of the Reader.
func (rd *Reader) Unread(runes ...rune) {
	newLine := false
	for _, r := range runes {
		if r == '\n' {
			newLine = true
			break
		}
	}

	if newLine {
		rd.line--
		rd.col = rd.lastCol
	} else {
		rd.col--
	}

	rd.buf = append(runes, rd.buf...)
}

// Position returns information about the stream including file name and the position
// of the reader.
func (rd Reader) Position() Position {
	file := strings.TrimSpace(rd.File)
	return Position{
		File: file,
		Ln:   rd.line + 1,
		Col:  rd.col,
	}
}

// SkipSpaces consumes and discards runes from stream repeatedly until a character that
// is not a whitespace is identified. Along with standard unicode whitespace characters,
// "," is also considered a whitespace and discarded.
func (rd *Reader) SkipSpaces() error {
	for {
		r, err := rd.NextRune()
		if err != nil {
			return err
		}

		if !isSpace(r) {
			rd.Unread(r)
			break
		}
	}

	return nil
}

// Token reads one token from the reader and returns. If init is not -1, it is included
// as first character in the token.
func (rd *Reader) Token(init rune) (string, error) {
	var b strings.Builder
	if init != -1 {
		b.WriteRune(init)
	}

	for {
		r, err := rd.NextRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return b.String(), err
		}

		if rd.IsTerminal(r) {
			rd.Unread(r)
			break
		}

		b.WriteRune(r)
	}

	return b.String(), nil
}

// Container reads multiple forms until 'end' rune is reached. Should be used to read
// collection types like List etc. formType is only used to annotate errors.
func (rd Reader) Container(end rune, formType string, f func(core.Any) error) error {
	for {
		if err := rd.SkipSpaces(); err != nil {
			if err == io.EOF {
				return Error{Cause: ErrEOF}
			}
			return err
		}

		r, err := rd.NextRune()
		if err != nil {
			if err == io.EOF {
				return Error{Cause: ErrEOF}
			}
			return err
		}

		if r == end {
			break
		}
		rd.Unread(r)

		expr, err := rd.readOne()
		if err != nil {
			if err == ErrSkip {
				continue
			}
			return err
		}

		// TODO(performance):  verify `f` is inlined by the compiler
		if err = f(expr); err != nil {
			return err
		}
	}

	return nil
}

// readOne is same as One() but always returns un-annotated errors.
func (rd *Reader) readOne() (core.Any, error) {
	if err := rd.SkipSpaces(); err != nil {
		return nil, err
	}

	r, err := rd.NextRune()
	if err != nil {
		return nil, err
	}

	if unicode.IsNumber(r) {
		return rd.numReader(rd, r)
	} else if r == '+' || r == '-' {
		r2, err := rd.NextRune()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if err != io.EOF {
			rd.Unread(r2)
			if unicode.IsNumber(r2) {
				return rd.numReader(rd, r)
			}
		}
	}

	macro, found := rd.macros[r]
	if found {
		return macro(rd, r)
	}

	if r == dispatchTrigger {
		f, err := rd.execDispatch()
		if f != nil || err != nil {
			return f, err
		}
	}

	return rd.symReader(rd, r)
}

func (rd *Reader) execDispatch() (core.Any, error) {
	r2, err := rd.NextRune()
	if err != nil {
		// ignore the error and let readOne handle it.
		return nil, nil
	}

	dispatchMacro, found := rd.dispatch[r2]
	if !found {
		rd.Unread(r2)
		return nil, nil
	}

	rd.dispatching = true
	defer func() {
		rd.dispatching = false
	}()

	form, err := dispatchMacro(rd, r2)
	if err != nil {
		return nil, err
	}

	return form, nil
}

func (rd *Reader) annotateErr(err error, beginPos Position /*, form string */) error {
	if err == io.EOF || err == ErrSkip {
		return err
	}

	readErr, ok := err.(Error)
	if !ok {
		readErr.Cause = err
	}

	// readErr.Form = form
	readErr.Begin = beginPos
	readErr.End = rd.Position()
	return readErr
}

func readUnicodeChar(token string, base int) (builtin.Char, error) {
	num, err := strconv.ParseInt(token, base, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid unicode character: '\\%s'", token)
	}

	if num < 0 || num >= unicode.MaxRune {
		return -1, fmt.Errorf("invalid unicode character: '\\%s'", token)
	}

	return builtin.Char(num), nil
}

func parseRadix(numStr string) (builtin.Int64, error) {
	parts := strings.Split(numStr, "r")
	if len(parts) != 2 {
		return 0, fmt.Errorf("%w (radix notation): '%s'", ErrNumberFormat, numStr)
	}

	base, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w (radix notation): '%s'", ErrNumberFormat, numStr)
	}

	repr := parts[1]
	if base < 0 {
		base = -1 * base
		repr = "-" + repr
	}

	v, err := strconv.ParseInt(repr, int(base), 64)
	if err != nil {
		return 0, fmt.Errorf("%w (radix notation): '%s'", ErrNumberFormat, numStr)
	}

	return builtin.Int64(v), nil
}

func parseScientific(numStr string) (builtin.Float64, error) {
	parts := strings.Split(numStr, "e")
	if len(parts) != 2 {
		return 0, fmt.Errorf("%w (scientific notation): '%s'", ErrNumberFormat, numStr)
	}

	base, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("%w (scientific notation): '%s'", ErrNumberFormat, numStr)
	}

	pow, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w (scientific notation): '%s'", ErrNumberFormat, numStr)
	}

	return builtin.Float64(base * math.Pow(10, float64(pow))), nil
}

func getEscape(r rune) (rune, error) {
	escaped, found := escapeMap[r]
	if !found {
		return -1, fmt.Errorf("illegal escape sequence '\\%c'", r)
	}

	return escaped, nil
}

func isSpace(r rune) bool {
	return unicode.IsSpace(r) || r == ','
}

func inferFileName(rs io.Reader) string {
	switch r := rs.(type) {
	case *os.File:
		return r.Name()

	case *strings.Reader:
		return "<string>"

	case *bytes.Reader:
		return "<bytes>"

	case net.Conn:
		return fmt.Sprintf("<con:%s>", r.LocalAddr())

	default:
		return fmt.Sprintf("<%s>", reflect.TypeOf(rs))
	}
}

// Position represents the positional information about a value read
// by reader.
type Position struct {
	File string
	Ln   int
	Col  int
}

func (p Position) String() string {
	if p.File == "" {
		p.File = "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", p.File, p.Ln, p.Col)
}
