package reflector

import (
	"reflect"
	"testing"

	"github.com/spy16/slurp/core"
)

var (
	int64Type = reflect.TypeOf(int64(0))
	int8Type  = reflect.TypeOf(int8(0))
	int8Vals  []reflect.Value
)

func init() {
	int8Vals = make([]reflect.Value, 1000)
	for i := 0; i < len(int8Vals); i++ {
		int8Vals[i] = reflect.ValueOf(int8(0))
	}
}

func Benchmark_funcWrapper_Invoke(b *testing.B) {
	fw := Func("foo", func(a, b int) int { return a + b })

	var res int
	for i := 0; i < b.N; i++ {
		ret, _ := fw.Invoke(1, 2)
		res = ret.(int)
	}
	b.Logf("final result: %d", res)
}

func Benchmark_funcWrapper_Invoke_Variadic(b *testing.B) {
	fw := Func("foo", func(args ...int32) int32 { return args[0] + args[len(args)-1] })

	args := make([]core.Any, 1000)
	for i := 0; i < len(args); i++ {
		args[i] = 1
	}

	var res int32
	for i := 0; i < b.N; i++ {
		ret, _ := fw.Invoke(args...)
		res = ret.(int32)
	}
	b.Logf("final result: %d", res)
}

func Benchmark_convertArgsTo_Convertible(b *testing.B) {
	var retVals []reflect.Value

	var ret []reflect.Value
	for i := 0; i < b.N; i++ {
		ret, _ = convertArgsTo(int64Type, int8Vals...)
		retVals = ret
	}
	dummyPrint(b, retVals)
}

func Benchmark_convertArgsTo_Assign(b *testing.B) {
	var retVals []reflect.Value

	var ret []reflect.Value
	for i := 0; i < b.N; i++ {
		ret, _ = convertArgsTo(int8Type, int8Vals...)
		retVals = ret
	}
	dummyPrint(b, retVals)
}

func dummyPrint(b *testing.B, retVals []reflect.Value) {
	b.Logf("return values of size=%d", len(retVals))
}
