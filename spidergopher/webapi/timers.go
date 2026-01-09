package webapi

import (
	"sync"
	"sync/atomic"
	"time"

	"go-browser/spidergopher/core"

	"github.com/dop251/goja"
)

// Timers implements setTimeout, setInterval, clearTimeout, clearInterval
type Timers struct {
	loop     *core.EventLoop
	vm       *goja.Runtime
	timers   map[int64]*timerEntry
	timersMu sync.Mutex
	nextID   int64
}

type timerEntry struct {
	timer      *time.Timer
	ticker     *time.Ticker
	done       chan struct{}
	isInterval bool
}

func NewTimers(loop *core.EventLoop) *Timers {
	return &Timers{
		loop:   loop,
		timers: make(map[int64]*timerEntry),
	}
}

// SetVM sets the runtime (called during engine setup)
func (t *Timers) SetVM(vm *goja.Runtime) {
	t.vm = vm
}

// SetTimeout schedules a one-time callback
func (t *Timers) SetTimeout(call goja.FunctionCall) goja.Value {
	callback := call.Argument(0)
	delay := call.Argument(1).ToInteger()

	fn, ok := goja.AssertFunction(callback)
	if !ok {
		return t.vm.ToValue(0)
	}

	id := atomic.AddInt64(&t.nextID, 1)

	timer := time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		t.loop.Schedule(func() {
			fn(goja.Undefined())
		})
		t.removeTimer(id)
	})

	t.timersMu.Lock()
	t.timers[id] = &timerEntry{timer: timer, isInterval: false}
	t.timersMu.Unlock()

	return t.vm.ToValue(id)
}

// ClearTimeout cancels a setTimeout
func (t *Timers) ClearTimeout(call goja.FunctionCall) goja.Value {
	id := call.Argument(0).ToInteger()
	t.cancelTimer(id)
	return goja.Undefined()
}

// SetInterval schedules a repeating callback
func (t *Timers) SetInterval(call goja.FunctionCall) goja.Value {
	callback := call.Argument(0)
	delay := call.Argument(1).ToInteger()

	fn, ok := goja.AssertFunction(callback)
	if !ok {
		return t.vm.ToValue(0)
	}

	if delay < 1 {
		delay = 1
	}

	id := atomic.AddInt64(&t.nextID, 1)
	done := make(chan struct{})

	ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)

	go func() {
		for {
			select {
			case <-ticker.C:
				t.loop.Schedule(func() {
					fn(goja.Undefined())
				})
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	t.timersMu.Lock()
	t.timers[id] = &timerEntry{ticker: ticker, done: done, isInterval: true}
	t.timersMu.Unlock()

	return t.vm.ToValue(id)
}

// ClearInterval cancels a setInterval
func (t *Timers) ClearInterval(call goja.FunctionCall) goja.Value {
	id := call.Argument(0).ToInteger()
	t.cancelTimer(id)
	return goja.Undefined()
}

func (t *Timers) cancelTimer(id int64) {
	t.timersMu.Lock()
	defer t.timersMu.Unlock()

	entry, ok := t.timers[id]
	if !ok {
		return
	}

	if entry.isInterval {
		if entry.done != nil {
			close(entry.done)
		}
	} else {
		if entry.timer != nil {
			entry.timer.Stop()
		}
	}

	delete(t.timers, id)
}

func (t *Timers) removeTimer(id int64) {
	t.timersMu.Lock()
	defer t.timersMu.Unlock()
	delete(t.timers, id)
}
