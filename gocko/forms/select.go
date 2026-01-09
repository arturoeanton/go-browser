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

// SelectHandler handles <select> elements
type SelectHandler struct{}

// Render draws the select dropdown
func (h *SelectHandler) Render(screen *ebiten.Image, box *layout.RenderBox, node *dom.Node, state *FormState) {
	id := GetElementID(node)
	x, y := float32(box.X), float32(box.Y)
	w, bh := float32(200), float32(32)

	// Get current value
	currentValue := state.GetValue(id)
	currentText := currentValue

	// Find display text from options
	for _, child := range node.Children {
		if strings.ToLower(child.Tag) == "option" {
			optVal := child.Attributes["value"]
			if optVal == "" {
				optVal = getOptionText(child)
			}
			if optVal == currentValue || (currentValue == "" && child.Attributes["selected"] != "") {
				currentText = getOptionText(child)
				if currentValue == "" {
					state.SetValue(id, optVal)
				}
				break
			}
		}
	}

	if currentText == "" {
		currentText = "Select..."
	}

	// Background
	bgColor := color.RGBA{255, 255, 255, 255}
	borderColor := color.RGBA{180, 180, 190, 255}

	if state.SelectOpen == id {
		borderColor = color.RGBA{66, 133, 244, 255}
	}

	vector.DrawFilledRect(screen, x-1, y-1, w+2, bh+2, borderColor, false)
	vector.DrawFilledRect(screen, x, y, w, bh, bgColor, false)

	// Current value text
	textColor := color.RGBA{33, 33, 33, 255}
	render.DrawText(screen, currentText, float64(x+10), float64(y+21), 14, textColor)

	// Dropdown arrow
	render.DrawText(screen, "▼", float64(x+w-22), float64(y+21), 12, color.RGBA{100, 100, 110, 255})

	// Draw dropdown if open
	if state.SelectOpen == id {
		h.renderDropdown(screen, x, y+bh, w, node, id, state)
	}
}

func (h *SelectHandler) renderDropdown(screen *ebiten.Image, x, y, w float32, node *dom.Node, id string, state *FormState) {
	options := getOptions(node)
	optH := float32(28)
	dropH := float32(len(options)) * optH

	// Dropdown background
	bgColor := color.RGBA{255, 255, 255, 255}
	borderColor := color.RGBA{180, 180, 190, 255}

	vector.DrawFilledRect(screen, x-1, y, w+2, dropH+2, borderColor, false)
	vector.DrawFilledRect(screen, x, y, w, dropH, bgColor, false)

	// Options
	currentY := y
	for _, opt := range options {
		textColor := color.RGBA{33, 33, 33, 255}

		// Highlight selected
		if opt.value == state.GetValue(id) {
			vector.DrawFilledRect(screen, x, currentY, w, optH, color.RGBA{230, 240, 255, 255}, false)
		}

		render.DrawText(screen, opt.text, float64(x+10), float64(currentY+19), 14, textColor)
		currentY += optH
	}
}

// RenderDropdownOnly renders only the dropdown portion (for overlay rendering)
func (h *SelectHandler) RenderDropdownOnly(screen *ebiten.Image, box *layout.RenderBox, node *dom.Node, state *FormState) {
	id := GetElementID(node)
	x, y := float32(box.X), float32(box.Y)
	w, bh := float32(200), float32(32)

	h.renderDropdown(screen, x, y+bh, w, node, id, state)
}

type selectOption struct {
	value string
	text  string
}

func getOptions(node *dom.Node) []selectOption {
	var options []selectOption
	for _, child := range node.Children {
		if strings.ToLower(child.Tag) == "option" {
			val := child.Attributes["value"]
			text := getOptionText(child)
			if val == "" {
				val = text
			}
			options = append(options, selectOption{value: val, text: text})
		}
	}
	return options
}

func getOptionText(node *dom.Node) string {
	for _, child := range node.Children {
		if child.Tag == "" && child.Content != "" {
			return strings.TrimSpace(child.Content)
		}
	}
	return ""
}

// HandleClick handles select click
func (h *SelectHandler) HandleClick(box *layout.RenderBox, node *dom.Node, x, y float64, state *FormState) bool {
	id := GetElementID(node)
	boxY := box.Y
	bh := float32(32)

	// Check if clicking on main select or dropdown
	if state.SelectOpen == id {
		// Clicking on dropdown - find which option
		optH := float64(28)
		dropY := boxY + float64(bh)

		if y >= dropY {
			// Calculate which option was clicked
			// Adjust for the offset between click position and option boxes
			relY := y - dropY - 5 // Small adjustment for better targeting
			if relY < 0 {
				relY = 0
			}
			optIdx := int(relY / optH)
			options := getOptions(node)
			if optIdx >= 0 && optIdx < len(options) {
				state.SetValue(id, options[optIdx].value)
			}
		}
		state.SelectOpen = ""
	} else {
		// Toggle dropdown
		state.SelectOpen = id
	}

	return true
}

// HandleInput - select handles arrow keys
func (h *SelectHandler) HandleInput(node *dom.Node, runes []rune, keys []ebiten.Key, state *FormState) bool {
	id := GetElementID(node)
	if state.SelectOpen != id {
		return false
	}

	options := getOptions(node)
	currentVal := state.GetValue(id)
	currentIdx := 0

	for i, opt := range options {
		if opt.value == currentVal {
			currentIdx = i
			break
		}
	}

	for _, key := range keys {
		switch key {
		case ebiten.KeyDown:
			if currentIdx < len(options)-1 {
				state.SetValue(id, options[currentIdx+1].value)
			}
		case ebiten.KeyUp:
			if currentIdx > 0 {
				state.SetValue(id, options[currentIdx-1].value)
			}
		case ebiten.KeyEnter:
			state.SelectOpen = ""
		case ebiten.KeyEscape:
			state.SelectOpen = ""
		}
	}

	return true
}

// GetValue returns selected value
func (h *SelectHandler) GetValue(node *dom.Node, state *FormState) string {
	return state.GetValue(GetElementID(node))
}

// IsFocusable returns true
func (h *SelectHandler) IsFocusable() bool {
	return true
}
