package reader

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/spy16/slurp/core"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		r        io.Reader
		fileName string
	}{
		{
			name:     "WithStringReader",
			r:        strings.NewReader(":test"),
			fileName: "<string>",
		},
		{
			name:     "WithBytesReader",
			r:        bytes.NewReader([]byte(":test")),
			fileName: "<bytes>",
		},
		{
			name:     "WihFile",
			r:        os.NewFile(0, "test.lisp"),
			fileName: "test.lisp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rd := New(tt.r)
			if rd == nil {
				t.Errorf("New() should return instance of Reader, got nil")
			} else if rd.File != tt.fileName {
				t.Errorf("File = \"%s\", want = \"%s\"", rd.File, tt.name)
			}
		})
	}
}

func TestReader_SetMacro(t *testing.T) {
	t.Run("UnsetDefaultMacro", func(t *testing.T) {
		rd := New(strings.NewReader("~hello"))
		rd.SetMacro('~', false, nil) // remove unquote operator

		want := core.Symbol("~hello")

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %+v, want = %+v", got, want)
		}
	})

	t.Run("DispatchMacro", func(t *testing.T) {
		rd := New(strings.NewReader("#$123"))
		// `#$` returns string "USD"
		rd.SetMacro('$', true, func(rd *Reader, init rune) (core.Any, error) {
			return core.String("USD"), nil
		}) // remove unquote operator

		want := core.String("USD")

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %#v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %+v, want = %+v", got, want)
		}
	})

	t.Run("CustomMacro", func(t *testing.T) {
		rd := New(strings.NewReader("~hello"))
		rd.SetMacro('~', false, func(rd *Reader, _ rune) (core.Any, error) {
			var ru []rune
			for {
				r, err := rd.NextRune()
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}

				if rd.IsTerminal(r) {
					break
				}
				ru = append(ru, r)
			}

			return core.String(ru), nil
		}) // override unquote operator

		want := core.String("hello")

		got, err := rd.One()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got = %+v, want = %+v", got, want)
		}
	})
}

func TestReader_All(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		want    []core.Any
		wantErr bool
	}{
		{
			name: "ValidLiteralSample",
			src:  `123 "Hello World" 12.34 -0xF +010 true nil 0b1010 \a :hello`,
			want: []core.Any{
				core.Int64(123),
				core.String("Hello World"),
				core.Float64(12.34),
				core.Int64(-15),
				core.Int64(8),
				core.Bool(true),
				core.Nil{},
				core.Int64(10),
				core.Char('a'),
				core.Keyword("hello"),
			},
		},
		{
			name: "WithComment",
			src:  `:valid-keyword ; comment should return errSkip`,
			want: []core.Any{core.Keyword("valid-keyword")},
		},
		{
			name:    "UnterminatedString",
			src:     `:valid-keyword "unterminated string literal`,
			wantErr: true,
		},
		{
			name: "CommentFollowedByForm",
			src:  `; comment should return errSkip` + "\n" + `:valid-keyword`,
			want: []core.Any{core.Keyword("valid-keyword")},
		},
		{
			name:    "UnterminatedList",
			src:     `:valid-keyword (add 1 2`,
			wantErr: true,
		},
		{
			name:    "EOFAfterQuote",
			src:     `:valid-keyword '`,
			wantErr: true,
		},
		{
			name:    "CommentAfterQuote",
			src:     `:valid-keyword ';hello world`,
			wantErr: true,
		},
		{
			name:    "UnbalancedParenthesis",
			src:     `())`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(strings.NewReader(tt.src)).All()
			if (err != nil) != tt.wantErr {
				t.Errorf("All() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("All() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestReader_One(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name:    "Empty",
			src:     "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "QuotedEOF",
			src:     "';comment is a no-op form\n",
			wantErr: true,
		},
		{
			name:    "ListEOF",
			src:     "( 1",
			wantErr: true,
		},
		{
			name: "UnQuote",
			src:  "~(x 3)",
			want: core.NewList(
				core.Symbol("unquote"),
				core.NewList(
					core.Symbol("x"),
					core.Int64(3),
				),
			),
		},
	})
}

func TestReader_One_Number(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "NumberWithLeadingSpaces",
			src:  "    +1234",
			want: core.Int64(1234),
		},
		{
			name: "PositiveInt",
			src:  "+1245",
			want: core.Int64(1245),
		},
		{
			name: "NegativeInt",
			src:  "-234",
			want: core.Int64(-234),
		},
		{
			name: "PositiveFloat",
			src:  "+1.334",
			want: core.Float64(1.334),
		},
		{
			name: "NegativeFloat",
			src:  "-1.334",
			want: core.Float64(-1.334),
		},
		{
			name: "PositiveHex",
			src:  "0x124",
			want: core.Int64(0x124),
		},
		{
			name: "NegativeHex",
			src:  "-0x124",
			want: core.Int64(-0x124),
		},
		{
			name: "PositiveOctal",
			src:  "0123",
			want: core.Int64(0123),
		},
		{
			name: "NegativeOctal",
			src:  "-0123",
			want: core.Int64(-0123),
		},
		{
			name: "PositiveBinary",
			src:  "0b10",
			want: core.Int64(2),
		},
		{
			name: "NegativeBinary",
			src:  "-0b10",
			want: core.Int64(-2),
		},
		{
			name: "PositiveBase2Radix",
			src:  "2r10",
			want: core.Int64(2),
		},
		{
			name: "NegativeBase2Radix",
			src:  "-2r10",
			want: core.Int64(-2),
		},
		{
			name: "PositiveBase4Radix",
			src:  "4r123",
			want: core.Int64(27),
		},
		{
			name: "NegativeBase4Radix",
			src:  "-4r123",
			want: core.Int64(-27),
		},
		{
			name: "ScientificSimple",
			src:  "1e10",
			want: core.Float64(1e10),
		},
		{
			name: "ScientificNegativeExponent",
			src:  "1e-10",
			want: core.Float64(1e-10),
		},
		{
			name: "ScientificWithDecimal",
			src:  "1.5e10",
			want: core.Float64(1.5e+10),
		},
		{
			name:    "FloatStartingWith0",
			src:     "012.3",
			want:    core.Float64(012.3),
			wantErr: false,
		},
		{
			name:    "InvalidValue",
			src:     "1ABe13",
			wantErr: true,
		},
		{
			name:    "InvalidScientificFormat",
			src:     "1e13e10",
			wantErr: true,
		},
		{
			name:    "InvalidExponent",
			src:     "1e1.3",
			wantErr: true,
		},
		{
			name:    "InvalidRadixFormat",
			src:     "1r2r3",
			wantErr: true,
		},
		{
			name:    "RadixBase3WithDigit4",
			src:     "-3r1234",
			wantErr: true,
		},
		{
			name:    "RadixMissingValue",
			src:     "2r",
			wantErr: true,
		},
		{
			name:    "RadixInvalidBase",
			src:     "2ar",
			wantErr: true,
		},
		{
			name:    "RadixWithFloat",
			src:     "2.3r4",
			wantErr: true,
		},
		{
			name:    "DecimalPointInBinary",
			src:     "0b1.0101",
			wantErr: true,
		},
		{
			name:    "InvalidDigitForOctal",
			src:     "08",
			wantErr: true,
		},
		{
			name:    "IllegalNumberFormat",
			src:     "9.3.2",
			wantErr: true,
		},
	})
}

func TestReader_One_String(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleString",
			src:  `"hello"`,
			want: core.String("hello"),
		},
		{
			name: "EscapeQuote",
			src:  `"double quote is \""`,
			want: core.String(`double quote is "`),
		},
		{
			name: "EscapeTab",
			src:  `"hello\tworld"`,
			want: core.String("hello\tworld"),
		},
		{
			name: "EscapeSlash",
			src:  `"hello\\world"`,
			want: core.String(`hello\world`),
		},
		{
			name:    "UnexpectedEOF",
			src:     `"double quote is`,
			wantErr: true,
		},
		{
			name:    "InvalidEscape",
			src:     `"hello \x world"`,
			wantErr: true,
		},
		{
			name:    "EscapeEOF",
			src:     `"hello\`,
			wantErr: true,
		},
	})
}

func TestReader_One_Keyword(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleASCII",
			src:  `:test`,
			want: core.Keyword("test"),
		},
		{
			name: "LeadingTrailingSpaces",
			src:  "          :test          ",
			want: core.Keyword("test"),
		},
		{
			name: "SimpleUnicode",
			src:  `:∂`,
			want: core.Keyword("∂"),
		},
		{
			name: "WithSpecialChars",
			src:  `:this-is-valid?`,
			want: core.Keyword("this-is-valid?"),
		},
		{
			name: "FollowedByMacroChar",
			src:  `:this-is-valid'hello`,
			want: core.Keyword("this-is-valid"),
		},
	})
}

func TestReader_One_Character(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "ASCIILetter",
			src:  `\a`,
			want: core.Char('a'),
		},
		{
			name: "ASCIIDigit",
			src:  `\1`,
			want: core.Char('1'),
		},
		{
			name: "Unicode",
			src:  `\∂`,
			want: core.Char('∂'),
		},
		{
			name: "Newline",
			src:  `\newline`,
			want: core.Char('\n'),
		},
		{
			name: "FormFeed",
			src:  `\formfeed`,
			want: core.Char('\f'),
		},
		{
			name: "Unicode",
			src:  `\u00AE`,
			want: core.Char('®'),
		},
		{
			name:    "InvalidUnicode",
			src:     `\uHELLO`,
			wantErr: true,
		},
		{
			name:    "OutOfRangeUnicode",
			src:     `\u-100`,
			wantErr: true,
		},
		{
			name:    "UnknownSpecial",
			src:     `\hello`,
			wantErr: true,
		},
		{
			name:    "EOF",
			src:     `\`,
			wantErr: true,
		},
	})
}

func TestReader_One_Symbol(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "SimpleASCII",
			src:  `hello`,
			want: core.Symbol("hello"),
		},
		{
			name: "Unicode",
			src:  `find-∂`,
			want: core.Symbol("find-∂"),
		},
		{
			name: "SingleChar",
			src:  `+`,
			want: core.Symbol("+"),
		},
	})
}

func TestReader_One_List(t *testing.T) {
	executeReaderTests(t, []readerTestCase{
		{
			name: "EmptyList",
			src:  `()`,
			want: core.NewList(),
		},
		{
			name: "ListWithOneEntry",
			src:  `(help)`,
			want: core.NewList(core.Symbol("help")),
		},
		{
			name: "ListWithMultipleEntry",
			src:  `(+ 0xF 3.1413)`,
			want: core.NewList(
				core.Symbol("+"),
				core.Int64(15),
				core.Float64(3.1413),
			),
		},
		{
			name: "ListWithCommaSeparator",
			src:  `(+,0xF,3.1413)`,
			want: core.NewList(
				core.Symbol("+"),
				core.Int64(15),
				core.Float64(3.1413),
			),
		},
		{
			name: "MultiLine",
			src: `(+
                      0xF
                      3.1413
					)`,
			want: core.NewList(
				core.Symbol("+"),
				core.Int64(15),
				core.Float64(3.1413),
			),
		},
		{
			name: "MultiLineWithComments",
			src: `(+         ; plus operator adds numerical values
                      0xF    ; hex representation of 15
                      3.1413 ; value of math constant pi
                  )`,
			want: core.NewList(
				core.Symbol("+"),
				core.Int64(15),
				core.Float64(3.1413),
			),
		},
		{
			name:    "UnexpectedEOF",
			src:     "(+ 1 2 ",
			wantErr: true,
		},
	})
}

type readerTestCase struct {
	name    string
	src     string
	want    core.Any
	wantErr bool
}

func executeReaderTests(t *testing.T, tests []readerTestCase) {
	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(strings.NewReader(tt.src)).One()
			if (err != nil) != tt.wantErr {
				t.Errorf("One() error = %#v, wantErr %#v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("One() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
