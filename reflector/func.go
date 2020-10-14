package reflector

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spy16/slurp"
)

var (
	envType = reflect.TypeOf((*slurp.Env)(nil)).Elem()
	errType = reflect.TypeOf((*error)(nil)).Elem()
)

// Func converts the given Go func value to a slurp Invokable value. Panics
// if the given value is not of Func kind.
func Func(name string, v interface{}) slurp.Invokable {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	if rt.Kind() != reflect.Func {
		panic("not a func")
	}

	minArgs := rt.NumIn()
	if rt.IsVariadic() {
		minArgs = minArgs - 1
	}

	passScope := (minArgs > 0) && (rt.In(0) == envType)
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
		passScope:  passScope,
		returnsErr: returnsErr,
		lastOutIdx: lastOutIdx,
	}
}

type funcWrapper struct {
	rv         reflect.Value
	rt         reflect.Type
	name       string
	passScope  bool
	minArgs    int
	returnsErr bool
	lastOutIdx int
}

func (fw *funcWrapper) Invoke(env *slurp.Env, args ...slurp.Any) (slurp.Any, error) {
	argVals := reflectValues(args)
	if fw.passScope {
		argVals = append([]reflect.Value{reflect.ValueOf(env)}, argVals...)
	}

	if err := fw.checkArgCount(len(argVals)); err != nil {
		return nil, err
	}

	argVals, err := fw.convertTypes(argVals...)
	if err != nil {
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

func (fw *funcWrapper) convertTypes(args ...reflect.Value) ([]reflect.Value, error) {
	var vals []reflect.Value

	for i := 0; i < fw.rt.NumIn(); i++ {
		if fw.rt.IsVariadic() && i == fw.rt.NumIn()-1 {
			c, err := convertArgsTo(fw.rt.In(i).Elem(), args[i:]...)
			if err != nil {
				return nil, err
			}
			vals = append(vals, c...)
			break
		}

		c, err := convertArgsTo(fw.rt.In(i), args[i])
		if err != nil {
			return nil, err
		}
		vals = append(vals, c...)
	}

	return vals, nil
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

func (fw *funcWrapper) wrapReturns(vals ...reflect.Value) (slurp.Any, error) {
	if fw.rt.NumOut() == 0 {
		return slurp.Nil{}, nil
	}

	if fw.returnsErr {
		errIndex := fw.lastOutIdx + 1
		if !vals[errIndex].IsNil() {
			return nil, vals[errIndex].Interface().(error)
		}

		if fw.rt.NumOut() == 1 {
			return slurp.Nil{}, nil
		}
	}

	wrapped := slurpValues(vals[0 : fw.lastOutIdx+1])
	if len(wrapped) == 1 {
		return wrapped[0], nil
	}

	return slurp.NewList(wrapped...), nil
}

func convertArgsTo(expected reflect.Type, args ...reflect.Value) ([]reflect.Value, error) {
	var converted []reflect.Value
	for _, arg := range args {
		actual := arg.Type()
		switch {
		case isAssignable(actual, expected):
			converted = append(converted, arg)

		case actual.ConvertibleTo(expected):
			converted = append(converted, arg.Convert(expected))

		default:
			return args, fmt.Errorf(
				"value of type '%s' cannot be converted to '%s'",
				actual, expected,
			)
		}
	}

	return converted, nil
}

func isAssignable(from, to reflect.Type) bool {
	return (from == to) || from.AssignableTo(to) ||
		(to.Kind() == reflect.Interface && from.Implements(to))
}

func reflectValues(args []slurp.Any) []reflect.Value {
	var rvs []reflect.Value
	for _, arg := range args {
		rvs = append(rvs, reflect.ValueOf(arg))
	}
	return rvs
}

func slurpValues(rvs []reflect.Value) []slurp.Any {
	var vals []slurp.Any
	for _, arg := range rvs {
		vals = append(vals, arg.Interface())
	}
	return vals
}
