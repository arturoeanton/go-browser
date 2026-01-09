package core

import (
	"sync"

	"github.com/dop251/goja"
)

// Job represents a unit of work in the event loop
type Job func()

// EventLoop manages the execution of JavaScript tasks.
type EventLoop struct {
	jobQueue   chan Job
	stopSignal chan struct{}
	running    bool
	mu         sync.Mutex
	vm         *goja.Runtime
}

// NewEventLoop creates a new EventLoop attached to a Goja runtime.
func NewEventLoop(vm *goja.Runtime) *EventLoop {
	return &EventLoop{
		jobQueue:   make(chan Job, 100),
		stopSignal: make(chan struct{}),
		vm:         vm,
	}
}

func (el *EventLoop) Start() {
	el.mu.Lock()
	if el.running {
		el.mu.Unlock()
		return
	}
	el.running = true
	el.mu.Unlock()

	go el.runLoop()
}

func (el *EventLoop) Stop() {
	el.mu.Lock()
	defer el.mu.Unlock()
	if !el.running {
		return
	}
	close(el.stopSignal)
	el.running = false
}

func (el *EventLoop) Schedule(job Job) {
	el.mu.Lock()
	defer el.mu.Unlock()
	if !el.running {
		return
	}

	select {
	case el.jobQueue <- job:
	default:
		go func() {
			el.jobQueue <- job
		}()
	}
}

func (el *EventLoop) runLoop() {
	for {
		select {
		case job := <-el.jobQueue:
			el.safeRun(job)
		case <-el.stopSignal:
			return
		}
	}
}

func (el *EventLoop) safeRun(job Job) {
	defer func() {
		if r := recover(); r != nil {
			// In production, log this via a hooked logger
		}
	}()
	job()
}

// RunOnLoop is a helper to execute code on the loop synchronously
func (el *EventLoop) RunOnLoop(fn func(*goja.Runtime)) {
	// If not running, just execute directly
	el.mu.Lock()
	if !el.running {
		el.mu.Unlock()
		fn(el.vm)
		return
	}
	el.mu.Unlock()

	done := make(chan struct{})

	// Use a goroutine to avoid blocking if the queue is full
	go func() {
		el.jobQueue <- func() {
			fn(el.vm)
			close(done)
		}
	}()

	<-done
}
