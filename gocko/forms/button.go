package forms

import (
	"image/color"

	"go-browser/dom"
	"go-browser/layout"
	"go-browser/render"

	"github.com/hajimehoshi/ebiten/v2"
)

// ButtonHandler handles <button> elements
type ButtonHandler struct{}

// Render draws the button
func (h *ButtonHandler) Render(screen *ebiten.Image, box *layout.RenderBox, node *dom.Node, state *FormState) {
	text := getButtonText(node)

	x, y := float32(box.X), float32(box.Y)
	w := float32(render.MeasureText(text, 14)) + 40
	bh := float32(36)

	// Button type
	btnType := node.Attributes["type"]

	// Colors based on type
	bgColor := color.RGBA{66, 133, 244, 255} // Primary blue
	if btnType == "button" {
		bgColor = color.RGBA{100, 100, 110, 255} // Gray for non-submit
	}
	if btnType == "reset" {
		bgColor = color.RGBA{108, 117, 125, 255} // Gray for reset
	}

	// More rounded corners (radius 10)
	render.DrawRoundedRect(screen, x, y, w, bh, 10, bgColor)

	// Text centered properly
	textColor := color.RGBA{255, 255, 255, 255}
	textX := float64(x) + float64(w)/2
	textY := float64(y) + float64(bh)/2 + 4 // Adjusted for better vertical centering
	render.DrawTextCentered(screen, text, textX, textY, 14, textColor)
}

func getButtonText(node *dom.Node) string {
	// Get text from children
	for _, child := range node.Children {
		if child.Tag == "" && child.Content != "" {
			return child.Content
		}
	}
	// Fallback
	if node.Attributes["value"] != "" {
		return node.Attributes["value"]
	}
	return "Button"
}

// HandleClick handles button click
func (h *ButtonHandler) HandleClick(box *layout.RenderBox, node *dom.Node, x, y float64, state *FormState) bool {
	btnType := node.Attributes["type"]
	if btnType == "" {
		btnType = "submit"
	}

	if btnType == "submit" {
		// TODO: trigger form submission
	}

	return true
}

// HandleInput - buttons don't handle input
func (h *ButtonHandler) HandleInput(node *dom.Node, runes []rune, keys []ebiten.Key, state *FormState) bool {
	return false
}

// GetValue returns button value
func (h *ButtonHandler) GetValue(node *dom.Node, state *FormState) string {
	return node.Attributes["value"]
}

// IsFocusable - buttons can be focused
func (h *ButtonHandler) IsFocusable() bool {
	return true
}
