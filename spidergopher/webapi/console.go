package webapi

import (
	"fmt"

	"github.com/dop251/goja"
)

// Console implements a subset of the Console API
type Console struct{}

func NewConsole() *Console {
	return &Console{}
}

func (c *Console) Log(call goja.FunctionCall) goja.Value {
	msg := formatArgs(call.Arguments)
	fmt.Println("[LOG]", msg)
	return goja.Undefined()
}

func (c *Console) Warn(call goja.FunctionCall) goja.Value {
	msg := formatArgs(call.Arguments)
	fmt.Println("[WARN]", msg)
	return goja.Undefined()
}

func (c *Console) Error(call goja.FunctionCall) goja.Value {
	msg := formatArgs(call.Arguments)
	fmt.Println("[ERROR]", msg)
	return goja.Undefined()
}

func formatArgs(args []goja.Value) string {
	var s string
	for i, arg := range args {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprint(arg.Export())
	}
	return s
}
