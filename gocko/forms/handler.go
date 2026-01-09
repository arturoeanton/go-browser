// Package forms provides HTML form element handlers for Gocko engine
package forms

import (
	"fmt"

	"go-browser/dom"
	"go-browser/layout"

	"github.com/hajimehoshi/ebiten/v2"
)

// TagHandler defines how an interactive element behaves
// This is the primary interface used by browser/app.go
type TagHandler interface {
	// Render draws the element on screen
	Render(screen *ebiten.Image, box *layout.RenderBox, node *dom.Node, state *FormState)

	// HandleClick processes click events, returns true if handled
	HandleClick(box *layout.RenderBox, node *dom.Node, x, y float64, state *FormState) bool

	// HandleInput processes keyboard input when focused
	HandleInput(node *dom.Node, runes []rune, keys []ebiten.Key, state *FormState) bool

	// GetValue returns the current value for form submission
	GetValue(node *dom.Node, state *FormState) string

	// IsFocusable returns true if element can receive focus
	IsFocusable() bool
}

// =============================================================================
// HANDLER REGISTRY
// =============================================================================

// Registry holds all tag handlers
var Registry = map[string]TagHandler{}

// RegisterHandler registers a handler for a tag
func RegisterHandler(tagName string, handler TagHandler) {
	Registry[tagName] = handler
}

// GetHandler returns the handler for a tag
func GetHandler(tagName string) TagHandler {
	return Registry[tagName]
}

// IsInteractive returns true if the tag has an interactive handler
func IsInteractive(tagName string) bool {
	_, ok := Registry[tagName]
	return ok
}

// init registers all form element handlers automatically when package is imported
func init() {
	RegisterHandler("input", &InputHandler{})
	RegisterHandler("button", &ButtonHandler{})
	RegisterHandler("select", &SelectHandler{})
	RegisterHandler("textarea", &TextareaHandler{})
}

// =============================================================================
// ELEMENT ID UTILITIES
// =============================================================================

// elementCounter tracks unique IDs for elements without id/name
var elementCounter = make(map[*dom.Node]string)
var idCounter int

// GetElementID returns a unique ID for the element
func GetElementID(node *dom.Node) string {
	if id := node.Attributes["id"]; id != "" {
		return id
	}
	if name := node.Attributes["name"]; name != "" {
		return name
	}
	// Check if we already assigned an ID to this node
	if cachedID, ok := elementCounter[node]; ok {
		return cachedID
	}
	// Generate a unique ID based on tag and counter
	idCounter++
	newID := fmt.Sprintf("%s_%d", node.Tag, idCounter)
	elementCounter[node] = newID
	return newID
}

// GetValueByID is an alias for GetValue for compatibility
func (fs *FormState) GetValueByID(id string) string {
	return fs.GetValue(id)
}
