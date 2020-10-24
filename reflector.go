package slurp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

var errType = reflect.TypeOf((*error)(nil)).Elem()

// Value converts the given arbitrary Go value into a slurp compatible value
// type with well defined behaviours. If no known equivalent type is found,
// then the value is returned as is.
func Value(v interface{}) core.Any {
	if v == nil {
		return builtin.Nil{}
	}

	if expr, ok := v.(core.Expr); ok {
		return expr
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Func:
		return Func(fmt.Sprintf("%v", v), rv)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return builtin.Int64(rv.Int())

	case reflect.Float32, reflect.Float64:
		return builtin.Float64(rv.Float())

	case reflect.String:
		return builtin.String(rv.String())

	case reflect.Uint8:
		return builtin.Char(rv.Uint())

	case reflect.Bool:
		return builtin.Bool(rv.Bool())

	default:
		// TODO: handle array & slice as list/vector.
		return v
	}
}

// Func converts the given Go func value to a slurp Invokable value. Panics
// if the given value is not of Func kind.
func Func(name string, v interface{}) builtin.Invokable {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		panic("not a func")
	}

	minArgs := rt.NumIn()
	if rt.IsVariadic() {
		minArgs = minArgs - 1
	}

	lastOutIdx := rt.NumOut() - 1
	returnsErr := lastOutIdx >= 0 && rt.Out(lastOutIdx) == errType
	if returnsErr {
		lastOutIdx-- // ignore error value from return values
	}

	return &funcWrapper{
		rv:         rv,
		rt:         rt,
		name:       name,
		minArgs:    minArgs,
		returnsErr: returnsErr,
		lastOutIdx: lastOutIdx,
	}
}

type funcWrapper struct {
	rv         reflect.Value
	rt         reflect.Type
	name       string
	minArgs    int
	returnsErr bool
	lastOutIdx int
}

func (fw *funcWrapper) Invoke(args ...core.Any) (core.Any, error) {
	// allocate argument slice.
	argCount := len(args)
	argVals := make([]reflect.Value, argCount, argCount)

	// populate reflect.Value version of each argument.
	for i, arg := range args {
		argVals[i] = reflect.ValueOf(arg)
	}

	// verify number of args match the required function parameters.
	if err := fw.checkArgCount(len(argVals)); err != nil {
		return nil, err
	}

	if err := fw.convertTypes(argVals...); err != nil {
		return nil, err
	}

	return fw.wrapReturns(fw.rv.Call(argVals)...)
}

func (fw *funcWrapper) String() string {
	args := fw.argNames()
	if fw.rt.IsVariadic() {
		args[len(args)-1] = "..." + args[len(args)-1]
	}

	for i, arg := range args {
		args[i] = fmt.Sprintf("arg%d %s", i, arg)
	}

	return fmt.Sprintf("func %s(%v)",
		fw.name, strings.Join(args, ", "))
}

func (fw *funcWrapper) argNames() []string {
	cleanArgName := func(t reflect.Type) string {
		return strings.Replace(t.String(), "slurp.", "", -1)
	}

	var argNames []string

	i := 0
	for ; i < fw.minArgs; i++ {
		argNames = append(argNames, cleanArgName(fw.rt.In(i)))
	}

	if fw.rt.IsVariadic() {
		argNames = append(argNames, cleanArgName(fw.rt.In(i).Elem()))
	}

	return argNames
}

func (fw *funcWrapper) convertTypes(args ...reflect.Value) error {
	lastArgIdx := fw.rt.NumIn() - 1
	isVariadic := fw.rt.IsVariadic()

	for i := 0; i < fw.rt.NumIn(); i++ {
		if i == lastArgIdx && isVariadic {
			c, err := convertArgsTo(fw.rt.In(i).Elem(), args[i:]...)
			if err != nil {
				return err
			}
			copy(args[i:], c)
			break
		}

		c, err := convertArgsTo(fw.rt.In(i), args[i])
		if err != nil {
			return err
		}
		args[i] = c[0]
	}

	return nil
}

func (fw *funcWrapper) checkArgCount(count int) error {
	if count != fw.minArgs {
		if fw.rt.IsVariadic() && count < fw.minArgs {
			return fmt.Errorf(
				"call requires at-least %d argument(s), got %d",
				fw.minArgs, count,
			)
		}

		if !fw.rt.IsVariadic() {
			return fmt.Errorf(
				"call requires exactly %d argument(s), got %d",
				fw.minArgs, count,
			)
		}
	}

	return nil
}

func (fw *funcWrapper) wrapReturns(vals ...reflect.Value) (core.Any, error) {
	if fw.rt.NumOut() == 0 {
		return builtin.Nil{}, nil
	}

	if fw.returnsErr {
		errIndex := fw.lastOutIdx + 1
		if !vals[errIndex].IsNil() {
			return nil, vals[errIndex].Interface().(error)
		}

		if fw.rt.NumOut() == 1 {
			return builtin.Nil{}, nil
		}
	}

	retValCount := len(vals[0 : fw.lastOutIdx+1])
	wrapped := make([]core.Any, retValCount, retValCount)
	for i := 0; i < retValCount; i++ {
		wrapped[i] = vals[i].Interface()
	}

	if retValCount == 1 {
		return wrapped[0], nil
	}

	return builtin.NewList(wrapped...), nil
}

func convertArgsTo(expected reflect.Type, args ...reflect.Value) ([]reflect.Value, error) {
	converted := make([]reflect.Value, len(args), len(args))
	for i, arg := range args {
		actual := arg.Type()
		isAssignable := (actual == expected) ||
			actual.AssignableTo(expected) ||
			(expected.Kind() == reflect.Interface && actual.Implements(expected))
		if isAssignable {
			converted[i] = arg
		} else if actual.ConvertibleTo(expected) {
			converted[i] = arg.Convert(expected)
		} else {
			return args, fmt.Errorf(
				"value of type '%s' cannot be converted to '%s'",
				actual, expected,
			)
		}
	}

	return converted, nil
}

func reflectValues(args []core.Any) []reflect.Value {
	var rvs []reflect.Value
	for _, arg := range args {
		rvs = append(rvs, reflect.ValueOf(arg))
	}
	return rvs
}

func slurpValues(rvs []reflect.Value) []core.Any {
	var vals []core.Any
	for _, arg := range rvs {
		vals = append(vals, arg.Interface())
	}
	return vals
}
