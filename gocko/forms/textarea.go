package forms

import (
	"image/color"
	"strings"

	"go-browser/dom"
	"go-browser/layout"
	"go-browser/render"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TextareaHandler handles <textarea> elements
type TextareaHandler struct{}

// Render draws the textarea
func (h *TextareaHandler) Render(screen *ebiten.Image, box *layout.RenderBox, node *dom.Node, state *FormState) {
	id := GetElementID(node)
	x, y := float32(box.X), float32(box.Y)
	w, bh := float32(300), float32(80)

	// Colors
	bgColor := color.RGBA{255, 255, 255, 255}
	borderColor := color.RGBA{180, 180, 190, 255}
	if state.IsFocused(id) {
		borderColor = color.RGBA{66, 133, 244, 255}
	}

	// Draw border and background
	vector.DrawFilledRect(screen, x-1, y-1, w+2, bh+2, borderColor, false)
	vector.DrawFilledRect(screen, x, y, w, bh, bgColor, false)

	// Get value
	value := state.GetValue(id)
	lines := strings.Split(value, "\n")

	// Show placeholder if empty
	placeholder := node.Attributes["placeholder"]
	textColor := color.RGBA{33, 33, 33, 255}
	if value == "" && placeholder != "" {
		lines = []string{placeholder}
		textColor = color.RGBA{150, 150, 160, 255}
	}

	// Draw lines
	lineHeight := float32(18)
	for i, line := range lines {
		if i >= 4 { // Max 4 visible lines
			break
		}
		render.DrawText(screen, line, float64(x+8), float64(y+float32(i)*lineHeight+16), 14, textColor)
	}

	// Draw cursor when focused
	if state.IsFocused(id) && (state.CursorBlink/30)%2 == 0 {
		cursorX := float32(x + 8)
		cursorY := y + 8
		vector.DrawFilledRect(screen, cursorX, cursorY, 2, lineHeight-4, color.RGBA{66, 133, 244, 255}, false)
	}
}

// HandleClick handles textarea click
func (h *TextareaHandler) HandleClick(box *layout.RenderBox, node *dom.Node, x, y float64, state *FormState) bool {
	id := GetElementID(node)
	state.SetFocus(id)

	// Initialize value if needed
	if _, ok := state.Values[id]; !ok {
		// Get default content from children
		for _, child := range node.Children {
			if child.Tag == "" && child.Content != "" {
				state.SetValue(id, strings.TrimSpace(child.Content))
				break
			}
		}
	}
	return true
}

// HandleInput handles keyboard input
func (h *TextareaHandler) HandleInput(node *dom.Node, runes []rune, keys []ebiten.Key, state *FormState) bool {
	id := GetElementID(node)
	if !state.IsFocused(id) {
		return false
	}

	value := state.GetValue(id)
	changed := false

	// Character input
	for _, r := range runes {
		if state.CursorPos <= len(value) {
			value = value[:state.CursorPos] + string(r) + value[state.CursorPos:]
			state.CursorPos++
			changed = true
		}
	}

	// Special keys
	for _, key := range keys {
		switch key {
		case ebiten.KeyEnter:
			value = value[:state.CursorPos] + "\n" + value[state.CursorPos:]
			state.CursorPos++
			changed = true
		case ebiten.KeyBackspace:
			if state.CursorPos > 0 && len(value) > 0 {
				value = value[:state.CursorPos-1] + value[state.CursorPos:]
				state.CursorPos--
				changed = true
			}
		case ebiten.KeyLeft:
			if state.CursorPos > 0 {
				state.CursorPos--
			}
		case ebiten.KeyRight:
			if state.CursorPos < len(value) {
				state.CursorPos++
			}
		}
	}

	if changed {
		state.SetValue(id, value)
	}

	return changed
}

// GetValue returns textarea content
func (h *TextareaHandler) GetValue(node *dom.Node, state *FormState) string {
	return state.GetValue(GetElementID(node))
}

// IsFocusable returns true
func (h *TextareaHandler) IsFocusable() bool {
	return true
}
