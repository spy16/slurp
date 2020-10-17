package reflector

import (
	"fmt"
	"reflect"

	"github.com/spy16/slurp/builtin"
	"github.com/spy16/slurp/core"
)

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
