package dom

import (
	"sync"

	"github.com/dop251/goja"
)

// Event represents a DOM Event
type Event struct {
	Type             string
	Target           interface{}
	CurrentTarget    interface{}
	Bubbles          bool
	Cancelable       bool
	DefaultPrevented bool
	TimeStamp        int64
}

// NewEvent creates a new Event
func NewEvent(eventType string) *Event {
	return &Event{
		Type:       eventType,
		Bubbles:    false,
		Cancelable: false,
	}
}

// ToJSObject converts Event to a JS-compatible map
func (e *Event) ToJSObject() map[string]interface{} {
	return map[string]interface{}{
		"type":             e.Type,
		"target":           e.Target,
		"currentTarget":    e.CurrentTarget,
		"bubbles":          e.Bubbles,
		"cancelable":       e.Cancelable,
		"defaultPrevented": e.DefaultPrevented,
		"preventDefault": func() {
			if e.Cancelable {
				e.DefaultPrevented = true
			}
		},
		"stopPropagation": func() {
			// For now, just a stub
		},
	}
}

// EventListener wraps a JS function
type EventListener struct {
	Callback goja.Value
	Once     bool
	Capture  bool
}

// EventTarget simulates the DOM EventTarget interface.
type EventTarget struct {
	listeners map[string][]*EventListener
	mu        sync.RWMutex
}

func NewEventTarget() *EventTarget {
	return &EventTarget{
		listeners: make(map[string][]*EventListener),
	}
}

// AddEventListener adds a listener for the specified event type.
func (et *EventTarget) AddEventListener(eventType string, callback goja.Value, options ...interface{}) {
	et.mu.Lock()
	defer et.mu.Unlock()

	listener := &EventListener{
		Callback: callback,
		Once:     false,
		Capture:  false,
	}

	// Parse options if provided
	if len(options) > 0 {
		if opts, ok := options[0].(map[string]interface{}); ok {
			if once, ok := opts["once"].(bool); ok {
				listener.Once = once
			}
			if capture, ok := opts["capture"].(bool); ok {
				listener.Capture = capture
			}
		}
	}

	// Check for duplicates (same callback)
	for _, l := range et.listeners[eventType] {
		if l.Callback.SameAs(callback) {
			return // Already registered
		}
	}

	et.listeners[eventType] = append(et.listeners[eventType], listener)
}

// RemoveEventListener removes a listener.
func (et *EventTarget) RemoveEventListener(eventType string, callback goja.Value) {
	et.mu.Lock()
	defer et.mu.Unlock()

	if list, ok := et.listeners[eventType]; ok {
		for i, l := range list {
			if l.Callback.SameAs(callback) {
				et.listeners[eventType] = append(list[:i], list[i+1:]...)
				return
			}
		}
	}
}

// DispatchEvent dispatches an event to listeners.
// Returns false if event.preventDefault() was called.
func (et *EventTarget) DispatchEvent(vm *goja.Runtime, event *Event) bool {
	event.Target = et
	event.CurrentTarget = et

	et.mu.RLock()
	// Copy listeners to avoid locking issues during execution
	var toCall []*EventListener
	if list, ok := et.listeners[event.Type]; ok {
		toCall = append([]*EventListener(nil), list...)
	}
	et.mu.RUnlock()

	// Track listeners to remove (for "once" listeners)
	var toRemove []goja.Value

	for _, listener := range toCall {
		if fn, ok := goja.AssertFunction(listener.Callback); ok {
			eventObj := vm.ToValue(event.ToJSObject())
			fn(goja.Undefined(), eventObj)

			if listener.Once {
				toRemove = append(toRemove, listener.Callback)
			}
		}
	}

	// Remove "once" listeners
	for _, cb := range toRemove {
		et.RemoveEventListener(event.Type, cb)
	}

	return !event.DefaultPrevented
}

// HasEventListeners returns true if there are listeners for the event type
func (et *EventTarget) HasEventListeners(eventType string) bool {
	et.mu.RLock()
	defer et.mu.RUnlock()
	return len(et.listeners[eventType]) > 0
}
