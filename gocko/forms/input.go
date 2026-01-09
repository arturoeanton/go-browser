package forms

import (
	"image/color"

	"go-browser/dom"
	"go-browser/layout"
	"go-browser/render"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// InputHandler handles <input> elements
type InputHandler struct{}

// Render draws the input element
func (ih *InputHandler) Render(screen *ebiten.Image, box *layout.RenderBox, node *dom.Node, state *FormState) {
	inputType := node.Attributes["type"]
	if inputType == "" {
		inputType = "text"
	}

	id := GetElementID(node)
	x, y := float32(box.X), float32(box.Y)
	w, inputH := float32(200), float32(30)

	switch inputType {
	case "text", "password", "email", "search", "tel", "url":
		ih.renderTextInput(screen, x, y, w, inputH, node, id, state, inputType == "password")
	case "checkbox":
		ih.renderCheckbox(screen, x, y, node, id, state)
	case "radio":
		ih.renderRadio(screen, x, y, node, id, state)
	case "submit", "button":
		ih.renderButton(screen, x, y, node)
	case "number":
		ih.renderTextInput(screen, x, y, w, inputH, node, id, state, false)
	}
}

func (h *InputHandler) renderTextInput(screen *ebiten.Image, x, y, w, bh float32, node *dom.Node, id string, state *FormState, isPassword bool) {
	// Background
	bgColor := color.RGBA{255, 255, 255, 255}
	borderColor := color.RGBA{180, 180, 190, 255}

	if state.IsFocused(id) {
		borderColor = color.RGBA{66, 133, 244, 255} // Blue focus
	}

	// Border
	vector.DrawFilledRect(screen, x-1, y-1, w+2, bh+2, borderColor, false)
	vector.DrawFilledRect(screen, x, y, w, bh, bgColor, false)

	// Value or placeholder
	value := state.GetValue(id)
	displayValue := value
	if isPassword {
		displayValue = ""
		for range value {
			displayValue += "•"
		}
	}

	placeholder := node.Attributes["placeholder"]
	textColor := color.RGBA{33, 33, 33, 255}

	if displayValue == "" && placeholder != "" {
		displayValue = placeholder
		textColor = color.RGBA{150, 150, 160, 255}
	}

	render.DrawText(screen, displayValue, float64(x+8), float64(y+20), 14, textColor)

	// Cursor when focused
	if state.IsFocused(id) && (state.CursorBlink/30)%2 == 0 {
		textBeforeCursor := value[:state.CursorPos]
		if isPassword {
			textBeforeCursor = ""
			for i := 0; i < state.CursorPos; i++ {
				textBeforeCursor += "•"
			}
		}
		cursorX := float32(float64(x) + 8 + render.MeasureText(textBeforeCursor, 14))
		vector.DrawFilledRect(screen, cursorX, y+6, 2, bh-12, color.RGBA{66, 133, 244, 255}, false)
	}
}

func (h *InputHandler) renderCheckbox(screen *ebiten.Image, x, y float32, node *dom.Node, id string, state *FormState) {
	size := float32(18)

	// Box
	borderColor := color.RGBA{100, 100, 110, 255}
	bgColor := color.RGBA{255, 255, 255, 255}

	vector.DrawFilledRect(screen, x, y, size, size, borderColor, false)
	vector.DrawFilledRect(screen, x+1, y+1, size-2, size-2, bgColor, false)

	// Checkmark if checked
	if state.IsChecked(id) {
		checkColor := color.RGBA{66, 133, 244, 255}
		vector.DrawFilledRect(screen, x+4, y+4, size-8, size-8, checkColor, false)
	}
}

func (h *InputHandler) renderRadio(screen *ebiten.Image, x, y float32, node *dom.Node, id string, state *FormState) {
	size := float32(18)

	// Circle
	borderColor := color.RGBA{100, 100, 110, 255}
	bgColor := color.RGBA{255, 255, 255, 255}

	render.DrawRoundedRect(screen, x, y, size, size, size/2, borderColor)
	render.DrawRoundedRect(screen, x+1, y+1, size-2, size-2, (size-2)/2, bgColor)

	// Dot if selected
	if state.IsChecked(id) {
		dotColor := color.RGBA{66, 133, 244, 255}
		render.DrawRoundedRect(screen, x+5, y+5, size-10, size-10, (size-10)/2, dotColor)
	}
}

func (h *InputHandler) renderButton(screen *ebiten.Image, x, y float32, node *dom.Node) {
	value := node.Attributes["value"]
	if value == "" {
		value = "Submit"
	}

	w := float32(render.MeasureText(value, 14)) + 24
	bh := float32(32)

	// Button background
	bgColor := color.RGBA{66, 133, 244, 255}
	render.DrawRoundedRect(screen, x, y, w, bh, 4, bgColor)

	// Button text
	textColor := color.RGBA{255, 255, 255, 255}
	render.DrawText(screen, value, float64(x+12), float64(y+21), 14, textColor)
}

// HandleClick processes click on input
func (h *InputHandler) HandleClick(box *layout.RenderBox, node *dom.Node, x, y float64, state *FormState) bool {
	inputType := node.Attributes["type"]
	if inputType == "" {
		inputType = "text"
	}

	id := GetElementID(node)

	switch inputType {
	case "text", "password", "email", "search", "tel", "url", "number":
		state.SetFocus(id)
		// Initialize value if not set
		if _, ok := state.Values[id]; !ok {
			defVal := node.Attributes["value"]
			state.SetValue(id, defVal)
		}
		return true

	case "checkbox":
		state.SetChecked(id, !state.IsChecked(id))
		return true

	case "radio":
		// Uncheck all radios with same name
		name := node.Attributes["name"]
		for k := range state.CheckedState {
			// Simple approach - clear all and set this one
			if name != "" {
				state.CheckedState[k] = false
			}
		}
		state.SetChecked(id, true)
		return true

	case "submit":
		// TODO: trigger form submission
		return true
	}

	return false
}

// HandleInput processes keyboard input
func (h *InputHandler) HandleInput(node *dom.Node, runes []rune, keys []ebiten.Key, state *FormState) bool {
	id := GetElementID(node)
	if !state.IsFocused(id) {
		return false
	}

	value := state.GetValue(id)
	changed := false

	// Handle character input
	for _, r := range runes {
		if state.CursorPos <= len(value) {
			value = value[:state.CursorPos] + string(r) + value[state.CursorPos:]
			state.CursorPos++
			changed = true
		}
	}

	// Handle special keys
	for _, key := range keys {
		switch key {
		case ebiten.KeyBackspace:
			if state.CursorPos > 0 && len(value) > 0 {
				value = value[:state.CursorPos-1] + value[state.CursorPos:]
				state.CursorPos--
				changed = true
			}
		case ebiten.KeyDelete:
			if state.CursorPos < len(value) {
				value = value[:state.CursorPos] + value[state.CursorPos+1:]
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
		case ebiten.KeyHome:
			state.CursorPos = 0
		case ebiten.KeyEnd:
			state.CursorPos = len(value)
		}
	}

	if changed {
		state.SetValue(id, value)
	}

	return changed
}

// GetValue returns input value
func (h *InputHandler) GetValue(node *dom.Node, state *FormState) string {
	inputType := node.Attributes["type"]
	id := GetElementID(node)

	switch inputType {
	case "checkbox", "radio":
		if state.IsChecked(id) {
			val := node.Attributes["value"]
			if val == "" {
				val = "on"
			}
			return val
		}
		return ""
	default:
		return state.GetValue(id)
	}
}

// IsFocusable returns true for text inputs
func (h *InputHandler) IsFocusable() bool {
	return true
}
