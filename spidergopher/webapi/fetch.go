package webapi

import (
	"io"
	"net/http"

	"go-browser/spidergopher/core"

	"github.com/dop251/goja"
)

// FetchAPI provides the fetch function
type FetchAPI struct {
	loop *core.EventLoop
	vm   *goja.Runtime
}

// NewFetchAPI creates a new FetchAPI
func NewFetchAPI(loop *core.EventLoop, vm *goja.Runtime) *FetchAPI {
	return &FetchAPI{loop: loop, vm: vm}
}

// Fetch implements the fetch() function
// Returns a Promise-like object
func (f *FetchAPI) Fetch(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	url := call.Argument(0).String()

	// Create a promise-like object
	promiseObj := f.vm.NewObject()

	var thenCallback goja.Callable
	var catchCallback goja.Callable

	promiseObj.Set("then", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
				thenCallback = fn
			}
		}
		return promiseObj // Allow chaining
	})

	promiseObj.Set("catch", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
				catchCallback = fn
			}
		}
		return promiseObj
	})

	// Make the HTTP request asynchronously
	go func() {
		resp, err := http.Get(url)

		// Schedule the callback on the event loop
		f.loop.Schedule(func() {
			if err != nil {
				if catchCallback != nil {
					catchCallback(goja.Undefined(), f.vm.ToValue(err.Error()))
				}
				return
			}
			defer resp.Body.Close()

			// Create Response object
			responseObj := f.createResponse(resp)

			if thenCallback != nil {
				thenCallback(goja.Undefined(), responseObj)
			}
		})
	}()

	return promiseObj
}

// createResponse creates a JS Response object
func (f *FetchAPI) createResponse(resp *http.Response) goja.Value {
	responseObj := f.vm.NewObject()

	responseObj.Set("ok", resp.StatusCode >= 200 && resp.StatusCode < 300)
	responseObj.Set("status", resp.StatusCode)
	responseObj.Set("statusText", resp.Status)

	// Read body once and store it
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)

	// text() returns a promise-like that resolves with the body as text
	responseObj.Set("text", func(call goja.FunctionCall) goja.Value {
		textPromise := f.vm.NewObject()
		var thenCb goja.Callable

		textPromise.Set("then", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
					thenCb = fn
				}
			}
			// Immediately resolve since we have the data
			if thenCb != nil {
				f.loop.Schedule(func() {
					thenCb(goja.Undefined(), f.vm.ToValue(bodyStr))
				})
			}
			return textPromise
		})

		return textPromise
	})

	// json() returns a promise-like that resolves with parsed JSON
	responseObj.Set("json", func(call goja.FunctionCall) goja.Value {
		jsonPromise := f.vm.NewObject()
		var thenCb goja.Callable

		jsonPromise.Set("then", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				if fn, ok := goja.AssertFunction(call.Argument(0)); ok {
					thenCb = fn
				}
			}
			if thenCb != nil {
				f.loop.Schedule(func() {
					// Parse JSON using Goja's JSON.parse
					jsonParse := f.vm.Get("JSON").ToObject(f.vm).Get("parse")
					if parseFn, ok := goja.AssertFunction(jsonParse); ok {
						result, err := parseFn(goja.Undefined(), f.vm.ToValue(bodyStr))
						if err != nil {
							// Return the raw string on parse error
							thenCb(goja.Undefined(), f.vm.ToValue(bodyStr))
						} else {
							thenCb(goja.Undefined(), result)
						}
					}
				})
			}
			return jsonPromise
		})

		return jsonPromise
	})

	return responseObj
}
