package reader

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

// Macro implementations can be plugged into the Reader to extend, override
// or customize behavior of the reader.
type Macro func(rd *Reader, init rune) (core.Any, error)

// // TODO(enhancement):  implement slurp.Set
// // SetReader implements the reader macro for reading set from source.
// func SetReader(setEnd rune, factory func() slurp.Set) Macro {
// 	return func(rd *Reader, _ rune) (core.Any, error) {
// 		forms, err := rd.Container(setEnd, "Set")
// 		if err != nil {
// 			return nil, err
// 		}
// 		return factory().Conj(forms...), nil
// 	}
// }

// // TODO(enhancement) implement slurp.Map
// // MapReader returns a reader macro for reading map values from source. factory
// // is used to construct the map and `Assoc` is called for every pair read.
// func MapReader(mapEnd rune, factory func() slurp.Map) Macro {
// 	return func(rd *Reader, _ rune) (core.Any, error) {
// 		forms, err := rd.Container(mapEnd, "Map")
// 		if err != nil {
// 			return nil, err
// 		}

// 		if len(forms)%2 != 0 {
// 			return nil, errors.New("expecting even number of forms within {}")
// 		}

// 		m := factory()
// 		for i := 0; i < len(forms); i += 2 {
// 			if m.HasKey(forms[i]) {
// 				return nil, fmt.Errorf("duplicate key: %v", forms[i])
// 			}

// 			m, err = m.Assoc(forms[i], forms[i+1])
// 			if err != nil {
// 				return nil, err
// 			}
// 		}

// 		return m, nil
// 	}
// }

// UnmatchedDelimiter implements a reader macro that can be used to capture
// unmatched delimiters such as closing parenthesis etc.
func UnmatchedDelimiter() Macro {
	return func(rd *Reader, initRune rune) (core.Any, error) {
		return nil, rd.annotateErr(unmatchedDelimiterError(initRune), rd.Position())
	}
}

func symbolReader(symTable map[string]core.Any) Macro {
	return func(rd *Reader, init rune) (core.Any, error) {
		beginPos := rd.Position()

		s, err := rd.Token(init)
		if err != nil {
			return nil, rd.annotateErr(err, beginPos)
		}

		if predefVal, found := symTable[s]; found {
			return predefVal, nil
		}

		return builtin.Symbol(s), nil
	}
}

func readNumber(rd *Reader, init rune) (core.Any, error) {
	beginPos := rd.Position()

	numStr, err := rd.Token(init)
	if err != nil {
		return nil, err
	}

	decimalPoint := strings.ContainsRune(numStr, '.')
	isRadix := strings.ContainsRune(numStr, 'r')
	isScientific := strings.ContainsRune(numStr, 'e')

	switch {
	case isRadix && (decimalPoint || isScientific):
		return nil, rd.annotateErr(ErrNumberFormat, beginPos)

	case isScientific:
		v, err := parseScientific(numStr)
		if err != nil {
			return nil, rd.annotateErr(err, beginPos)
		}
		return v, nil

	case decimalPoint:
		v, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return nil, rd.annotateErr(ErrNumberFormat, beginPos)
		}
		return builtin.Float64(v), nil

	case isRadix:
		v, err := parseRadix(numStr)
		if err != nil {
			return nil, rd.annotateErr(err, beginPos)
		}
		return v, nil

	default:
		v, err := strconv.ParseInt(numStr, 0, 64)
		if err != nil {
			return nil, rd.annotateErr(ErrNumberFormat, beginPos)
		}

		return builtin.Int64(v), nil
	}
}

func readString(rd *Reader, init rune) (core.Any, error) {
	beginPos := rd.Position()

	var b strings.Builder
	for {
		r, err := rd.NextRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = ErrEOF
			}
			return nil, rd.annotateErr(err, beginPos)
		}

		if r == '\\' {
			r2, err := rd.NextRune()
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = ErrEOF
				}

				return nil, rd.annotateErr(err, beginPos)
			}

			// TODO: Support for Unicode escape \uNN format.

			escaped, err := getEscape(r2)
			if err != nil {
				return nil, err
			}
			r = escaped
		} else if r == '"' {
			break
		}

		b.WriteRune(r)
	}

	return builtin.String(b.String()), nil
}

func readComment(rd *Reader, _ rune) (core.Any, error) {
	for {
		r, err := rd.NextRune()
		if err != nil {
			return nil, err
		}

		if r == '\n' {
			break
		}
	}

	return nil, ErrSkip
}

func readKeyword(rd *Reader, init rune) (core.Any, error) {
	beginPos := rd.Position()

	token, err := rd.Token(-1)
	if err != nil {
		return nil, rd.annotateErr(err, beginPos)
	}

	return builtin.Keyword(token), nil
}

func readCharacter(rd *Reader, _ rune) (core.Any, error) {
	beginPos := rd.Position()

	r, err := rd.NextRune()
	if err != nil {
		return nil, rd.annotateErr(err, beginPos)
	}

	token, err := rd.Token(r)
	if err != nil {
		return nil, err
	}
	runes := []rune(token)

	if len(runes) == 1 {
		return builtin.Char(runes[0]), nil
	}

	v, found := charLiterals[token]
	if found {
		return builtin.Char(v), nil
	}

	if token[0] == 'u' {
		return readUnicodeChar(token[1:], 16)
	}

	return nil, fmt.Errorf("unsupported character: '\\%s'", token)
}

func readList(rd *Reader, _ rune) (core.Any, error) {
	const listEnd = ')'

	beginPos := rd.Position()

	forms := make([]core.Any, 0, 32) // pre-allocate to improve performance on small lists
	if err := rd.Container(listEnd, "list", func(val core.Any) error {
		forms = append(forms, val)
		return nil
	}); err != nil {
		return nil, rd.annotateErr(err, beginPos)
	}

	return builtin.NewList(forms...), nil
}

func quoteFormReader(expandFunc string) Macro {
	return func(rd *Reader, _ rune) (core.Any, error) {
		expr, err := rd.One()
		if err != nil {
			if err == io.EOF {
				return nil, Error{
					// Form:  expandFunc,
					Cause: ErrEOF,
				}
			} else if err == ErrSkip {
				return nil, Error{
					// Form:  expandFunc,
					Cause: errors.New("cannot quote a no-op form"),
				}
			}
			return nil, err
		}

		return builtin.NewList(builtin.Symbol(expandFunc), expr), nil
	}
}

// readVector implements the reader macro for reading vector from source.
func readVector(rd *Reader, _ rune) (core.Any, error) {
	const vecEnd = ']'

	v := builtin.EmptyVector.Transient()
	err := rd.Container(vecEnd, "Vector", func(val core.Any) error {
		v.Cons(val)
		return nil
	})

	return v.Persistent(), err
}
