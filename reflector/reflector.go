package reflector

import (
	"fmt"
	"reflect"

	"github.com/spy16/slurp"
)

// Value converts the given arbitrary Go value into a slurp compatible value
// type with well defined behaviours. If no known equivalent type is found,
// then the value is returned as is.
func Value(v interface{}) slurp.Any {
	if v == nil {
		return slurp.Nil{}
	}

	if expr, ok := v.(slurp.Expr); ok {
		return expr
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Func:
		return Func(fmt.Sprintf("%v", v), rv)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return slurp.Int64(rv.Int())

	case reflect.Float32, reflect.Float64:
		return slurp.Float64(rv.Float())

	case reflect.String:
		return slurp.String(rv.String())

	case reflect.Uint8:
		return slurp.Char(rv.Uint())

	case reflect.Bool:
		return slurp.Bool(rv.Bool())

	default:
		// TODO: handle array & slice as list/vector.
		return v
	}
}
