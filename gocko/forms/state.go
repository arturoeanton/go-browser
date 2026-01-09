// Package forms provides HTML form components for Gocko
package forms

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// FormComponent defines the interface for all form elements
type FormComponent interface {
	// Render draws the component
	Render(screen *ebiten.Image, box interface{}, state *FormState)

	// HandleClick processes click events, returns true if handled
	HandleClick(localX, localY float64, state *FormState) bool

	// HandleInput processes keyboard input when focused
	HandleInput(runes []rune, keys []ebiten.Key, state *FormState) bool

	// GetValue returns the current value for form submission
	GetValue(state *FormState) string

	// SetValue sets the component value
	SetValue(value string, state *FormState)

	// Validate returns (isValid, errorMessage)
	Validate(state *FormState) (bool, string)

	// IsFocusable returns true if element can receive focus
	IsFocusable() bool

	// GetID returns the element ID
	GetID() string
}

// FormState tracks the state of all form elements
type FormState struct {
	// Values maps element ID to current value
	Values map[string]string

	// CheckedState tracks checkbox/radio states
	CheckedState map[string]bool

	// Focus state
	FocusedID string
	CursorPos int

	// Text selection
	SelectionStart int
	SelectionEnd   int

	// Cursor animation
	CursorBlink int

	// Dropdown state
	SelectOpen string

	// Validation errors
	ValidationErrors map[string]string

	// Files for file inputs
	Files map[string][]FileInfo
}

// FileInfo represents an uploaded file
type FileInfo struct {
	Name string
	Size int64
	Data []byte
}

// NewFormState creates a new form state
func NewFormState() *FormState {
	return &FormState{
		Values:           make(map[string]string),
		CheckedState:     make(map[string]bool),
		ValidationErrors: make(map[string]string),
		Files:            make(map[string][]FileInfo),
	}
}

// SetValue sets a value in form state
func (fs *FormState) SetValue(id, value string) {
	fs.Values[id] = value
}

// GetValue gets a value from form state
func (fs *FormState) GetValue(id string) string {
	return fs.Values[id]
}

// IsChecked returns true if a checkbox/radio is checked
func (fs *FormState) IsChecked(id string) bool {
	return fs.CheckedState[id]
}

// SetChecked sets the checked state
func (fs *FormState) SetChecked(id string, checked bool) {
	fs.CheckedState[id] = checked
}

// IsFocused returns true if element has focus
func (fs *FormState) IsFocused(id string) bool {
	return fs.FocusedID == id
}

// SetFocus sets focus to an element
func (fs *FormState) SetFocus(id string) {
	fs.FocusedID = id
	fs.CursorPos = len(fs.Values[id])
	fs.SelectionStart = 0
	fs.SelectionEnd = 0
}

// ClearFocus removes focus
func (fs *FormState) ClearFocus() {
	fs.FocusedID = ""
}

// HasSelection returns true if there's a text selection
func (fs *FormState) HasSelection() bool {
	return fs.SelectionStart != fs.SelectionEnd
}

// GetSelectedText returns the selected text for the focused element
func (fs *FormState) GetSelectedText() string {
	if fs.FocusedID == "" || !fs.HasSelection() {
		return ""
	}
	value := fs.Values[fs.FocusedID]
	start, end := fs.SelectionStart, fs.SelectionEnd
	if start > end {
		start, end = end, start
	}
	if end > len(value) {
		end = len(value)
	}
	if start < 0 {
		start = 0
	}
	return value[start:end]
}

// SelectAll selects all text in the focused element
func (fs *FormState) SelectAll() {
	if fs.FocusedID == "" {
		return
	}
	fs.SelectionStart = 0
	fs.SelectionEnd = len(fs.Values[fs.FocusedID])
}
