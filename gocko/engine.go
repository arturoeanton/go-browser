// Package gocko provides a web rendering engine written in Go
// Inspired by Gecko (Firefox's engine) but built from scratch
package gocko

import (
	"go-browser/css"
	"go-browser/dom"
	"go-browser/gocko/box"
	"go-browser/gocko/forms"
	"go-browser/gocko/layout"
	"go-browser/gocko/paint"

	"github.com/hajimehoshi/ebiten/v2"
)

// Engine is the main Gocko rendering engine
type Engine struct {
	// DOM tree
	DOM *dom.Node

	// Parsed stylesheets
	Styles []*css.Stylesheet

	// Layout tree (box model)
	LayoutTree *box.Box

	// Form state manager
	FormState *forms.FormState

	// Viewport dimensions
	ViewportWidth  float64
	ViewportHeight float64
}

// New creates a new Gocko engine instance
func New() *Engine {
	return &Engine{
		FormState: forms.NewFormState(),
	}
}

// SetDocument sets the DOM and styles for rendering
func (e *Engine) SetDocument(dom *dom.Node, styles []*css.Stylesheet) {
	e.DOM = dom
	e.Styles = styles
	e.LayoutTree = nil // Invalidate layout
}

// Layout computes the layout tree from DOM
func (e *Engine) Layout() {
	if e.DOM == nil {
		return
	}
	e.LayoutTree = layout.BuildLayoutTree(e.DOM, e.ViewportWidth, e.Styles)
}

// Paint renders the layout tree to the screen
func (e *Engine) Paint(screen *ebiten.Image, offsetX, offsetY float64) {
	if e.LayoutTree == nil {
		return
	}
	paint.PaintTree(screen, e.LayoutTree, offsetX, offsetY, e.FormState)
}

// HandleClick processes a click event at the given coordinates
func (e *Engine) HandleClick(x, y float64) bool {
	if e.LayoutTree == nil {
		return false
	}
	return e.LayoutTree.HandleClick(x, y, e.FormState)
}

// HandleInput processes keyboard input
func (e *Engine) HandleInput(runes []rune, keys []ebiten.Key) bool {
	if e.FormState.FocusedID == "" {
		return false
	}

	// Find focused element and delegate input
	focusedBox := e.LayoutTree.FindByID(e.FormState.FocusedID)
	if focusedBox != nil && focusedBox.FormComponent != nil {
		return focusedBox.FormComponent.HandleInput(runes, keys, e.FormState)
	}
	return false
}

// Update should be called each frame
func (e *Engine) Update() {
	// Increment cursor blink counter
	e.FormState.CursorBlink++
}

// FindLink finds a link at the given coordinates
func (e *Engine) FindLink(x, y float64) string {
	if e.LayoutTree == nil {
		return ""
	}
	return e.LayoutTree.FindLink(x, y)
}
