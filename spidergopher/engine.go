package spidergopher

import (
	realdom "go-browser/dom"
	"go-browser/spidergopher/core"
	"go-browser/spidergopher/dom"
	"go-browser/spidergopher/webapi"

	"github.com/dop251/goja"
)

// Engine is the main entry point for the SpiderGopher JS environment.
type Engine struct {
	Loop      *core.EventLoop
	Window    *dom.Window
	vm        *goja.Runtime
	domBridge *dom.DOMBridge
}

// NewEngine creates a new SpiderGopher engine.
func NewEngine() *Engine {
	vm := goja.New()
	loop := core.NewEventLoop(vm)
	window := dom.NewWindow()

	engine := &Engine{
		Loop:   loop,
		Window: window,
		vm:     vm,
	}

	engine.setupGlobalEnv()
	return engine
}

// SetDOM connects the engine to a real DOM tree
func (e *Engine) SetDOM(root *realdom.Node) {
	e.domBridge = dom.NewDOMBridge(root, e.vm)
	// Update the document object in JS
	e.vm.Set("document", e.domBridge.GetDocumentObject())
}

// Start begins the event loop.
func (e *Engine) Start() {
	e.Loop.Start()
}

// Stop halts the event loop.
func (e *Engine) Stop() {
	e.Loop.Stop()
}

// Run executes a script synchronously.
// For the initial page script, this runs directly since we're on the main thread.
// For async callbacks, the EventLoop handles scheduling.
func (e *Engine) Run(script string) (goja.Value, error) {
	return e.vm.RunString(script)
}

// GetVM returns the Goja runtime for external use
func (e *Engine) GetVM() *goja.Runtime {
	return e.vm
}

func (e *Engine) setupGlobalEnv() {
	// Register global objects

	// Console
	console := webapi.NewConsole()
	consoleObj := e.vm.NewObject()
	consoleObj.Set("log", console.Log)
	consoleObj.Set("warn", console.Warn)
	consoleObj.Set("error", console.Error)
	e.vm.Set("console", consoleObj)

	// Timers
	timers := webapi.NewTimers(e.Loop)
	timers.SetVM(e.vm)
	e.vm.Set("setTimeout", timers.SetTimeout)
	e.vm.Set("clearTimeout", timers.ClearTimeout)
	e.vm.Set("setInterval", timers.SetInterval)
	e.vm.Set("clearInterval", timers.ClearInterval)

	// Document - explicitly set methods with lowercase names for JS compatibility
	doc := e.Window.Document
	vm := e.vm
	documentObj := e.vm.NewObject()
	documentObj.Set("getElementById", func(id string) map[string]interface{} {
		return doc.GetElementById(id)
	})
	documentObj.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventType := call.Argument(0).String()
		callback := call.Argument(1)
		doc.AddEventListener(eventType, callback)
		return goja.Undefined()
	})
	documentObj.Set("removeEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventType := call.Argument(0).String()
		callback := call.Argument(1)
		doc.RemoveEventListener(eventType, callback)
		return goja.Undefined()
	})
	documentObj.Set("dispatchEvent", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue(false)
		}
		// Accept event object with 'type' property
		eventArg := call.Argument(0).Export()
		if eventMap, ok := eventArg.(map[string]interface{}); ok {
			if eventType, ok := eventMap["type"].(string); ok {
				event := dom.NewEvent(eventType)
				result := doc.DispatchEvent(vm, event)
				return vm.ToValue(result)
			}
		}
		return vm.ToValue(false)
	})
	e.vm.Set("document", documentObj)

	// Window with event support
	windowObj := e.vm.NewObject()
	windowObj.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventType := call.Argument(0).String()
		callback := call.Argument(1)
		e.Window.AddEventListener(eventType, callback)
		return goja.Undefined()
	})
	windowObj.Set("removeEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventType := call.Argument(0).String()
		callback := call.Argument(1)
		e.Window.RemoveEventListener(eventType, callback)
		return goja.Undefined()
	})
	windowObj.Set("dispatchEvent", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue(false)
		}
		eventArg := call.Argument(0).Export()
		if eventMap, ok := eventArg.(map[string]interface{}); ok {
			if eventType, ok := eventMap["type"].(string); ok {
				event := dom.NewEvent(eventType)
				result := e.Window.DispatchEvent(vm, event)
				return vm.ToValue(result)
			}
		}
		return vm.ToValue(false)
	})
	windowObj.Set("document", documentObj)
	e.vm.Set("window", windowObj)

	// Self-reference
	e.vm.Set("self", windowObj)

	// Global addEventListener (browsers expose this at global scope)
	e.vm.Set("addEventListener", windowObj.Get("addEventListener"))
	e.vm.Set("removeEventListener", windowObj.Get("removeEventListener"))
	e.vm.Set("dispatchEvent", windowObj.Get("dispatchEvent"))

	// Fetch API
	fetchAPI := webapi.NewFetchAPI(e.Loop, e.vm)
	e.vm.Set("fetch", fetchAPI.Fetch)

	// NOTE: Storage APIs are disabled for now due to SQLite init blocking the event loop.
	// They will be initialized lazily when first accessed.
	// TODO: Implement lazy initialization within the event loop context
}

// Wait blocks until the loop stops (if ever) or just for a duration?
// Usually main thread waits.
