package eval

import (
	"math"
	"reflect"

	"github.com/tobgu/qframe/errors"
	"github.com/tobgu/qframe/function"
	"github.com/tobgu/qframe/internal/strings"
	"github.com/tobgu/qframe/types"
)

type functionsByArgCount struct {
	singleArgs map[string]interface{}
	doubleArgs map[string]interface{}
}

type functionsByArgType map[types.FunctionType]functionsByArgCount

// ArgCount is the number of arguments passed to a function to be evaluated.
type ArgCount byte

const (
	ArgCountOne ArgCount = iota
	ArgCountTwo
)

// String returns a string representation of the ArgCount
func (c ArgCount) String() string {
	switch c {
	case ArgCountOne:
		return "Single argument"
	case ArgCountTwo:
		return "Double argument"
	default:
		return "Unknown argument count"
	}
}

// Context describes the context in which an expression is executed.
// It maps function names to actual functions.
type Context struct {
	functions functionsByArgType
}

// NewDefaultCtx creates a default context containing a base set of functions.
// It can be used as is or enhanced with other/more functions. See the source code
// for the current set of functions.
func NewDefaultCtx() *Context {
	// TODO: More functions
	return &Context{
		functionsByArgType{
			types.FunctionTypeFloat: functionsByArgCount{
				singleArgs: map[string]interface{}{
					"abs": math.Abs,
					"str": function.StrF,
					"int": function.IntF,
				},
				doubleArgs: map[string]interface{}{
					"+": function.PlusF,
					"-": function.MinusF,
					"*": function.MulF,
					"/": function.DivF,
				},
			},
			types.FunctionTypeInt: functionsByArgCount{
				singleArgs: map[string]interface{}{
					"abs":   function.AbsI,
					"str":   function.StrI,
					"bool":  function.BoolI,
					"float": function.FloatI,
				},
				doubleArgs: map[string]interface{}{
					"+": function.PlusI,
					"-": function.MinusI,
					"*": function.MulI,
					"/": function.DivI,
				},
			},
			types.FunctionTypeBool: functionsByArgCount{
				singleArgs: map[string]interface{}{
					"!":   function.NotB,
					"str": function.StrB,
					"int": function.IntB,
				},
				doubleArgs: map[string]interface{}{
					"&":    function.AndB,
					"|":    function.OrB,
					"!=":   function.XorB,
					"nand": function.NandB,
				},
			},
			types.FunctionTypeString: functionsByArgCount{
				singleArgs: map[string]interface{}{
					"upper": function.UpperS,
					"lower": function.LowerS,
					"str":   function.StrS,
					"len":   function.LenS,
				},
				doubleArgs: map[string]interface{}{
					"+": function.ConcatS,
				},
			},
		},
	}
}

// GetFunc returns a reference to a function matching the given function type, argument count and name.
// If no matching function is found in the context the second return value is set to false.
func (ctx *Context) GetFunc(typ types.FunctionType, ac ArgCount, name string) (interface{}, bool) {
	var fn interface{}
	var ok bool
	if ac == ArgCountOne {
		fn, ok = ctx.functions[typ].singleArgs[name]
	} else {
		fn, ok = ctx.functions[typ].doubleArgs[name]
	}

	return fn, ok
}

func (ctx *Context) setFunc(typ types.FunctionType, ac ArgCount, name string, fn interface{}) {
	if ac == ArgCountOne {
		ctx.functions[typ].singleArgs[name] = fn
	} else {
		ctx.functions[typ].doubleArgs[name] = fn
	}
}

// SetFunc inserts a function into the context under the given name.
func (ctx *Context) SetFunc(name string, fn interface{}) error {
	if err := strings.CheckName(name); err != nil {
		return errors.Propagate("SetFunc", err)
	}

	// TODO: Check function name validity (eg must not start with $, more?)
	// Since there's such a flexibility in the function types that can be
	// used and there is no static typing to support it this function
	// acts as the gate keeper for adding new functions.
	var ac ArgCount
	var typ types.FunctionType
	switch fn.(type) {
	// Int
	case func(int, int) int:
		ac, typ = ArgCountTwo, types.FunctionTypeInt
	case func(int) int, func(int) bool, func(int) float64, func(int) *string:
		ac, typ = ArgCountOne, types.FunctionTypeInt

	// Float
	case func(float64, float64) float64:
		ac, typ = ArgCountTwo, types.FunctionTypeFloat
	case func(float64) float64, func(float64) int, func(float64) bool, func(float64) *string:
		ac, typ = ArgCountOne, types.FunctionTypeFloat

	// Bool
	case func(bool, bool) bool:
		ac, typ = ArgCountTwo, types.FunctionTypeBool
	case func(bool) bool, func(bool) int, func(bool) float64, func(bool) *string:
		ac, typ = ArgCountOne, types.FunctionTypeBool

	// String
	case func(*string, *string) *string:
		ac, typ = ArgCountTwo, types.FunctionTypeString
	case func(*string) *string, func(*string) int, func(*string) float64, func(*string) bool:
		ac, typ = ArgCountOne, types.FunctionTypeString

	default:
		return errors.New("SetFunc", "invalid function type for function \"%s\": %v", name, reflect.TypeOf(fn))
	}

	ctx.setFunc(typ, ac, name, fn)
	return nil
}
