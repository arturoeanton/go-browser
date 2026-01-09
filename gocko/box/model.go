// Package box provides the CSS Box Model implementation
package box

import (
	"go-browser/dom"
	"go-browser/gocko/forms"

	"github.com/hajimehoshi/ebiten/v2"
)

// Box represents a CSS box in the layout tree
type Box struct {
	// Reference to DOM node
	Node *dom.Node

	// Position and dimensions (content box)
	X, Y          float64
	Width, Height float64

	// Box model spacing
	MarginTop, MarginRight, MarginBottom, MarginLeft     float64
	BorderTop, BorderRight, BorderBottom, BorderLeft     float64
	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft float64

	// Children boxes
	Children []*Box

	// Display and positioning
	Display  string // block, inline, flex, grid, none
	Position string // static, relative, absolute, fixed
	Float    string // none, left, right

	// Text content (for inline boxes)
	Text     string
	FontSize float64
	IsBold   bool

	// Link info
	IsLink  bool
	LinkURL string

	// Image info
	IsImage  bool
	ImageURL string

	// Form component (if this box is a form element)
	FormComponent forms.FormComponent

	// Unique ID for this box (from DOM id or generated)
	ID string
}

// ContentBox returns the content area dimensions
func (b *Box) ContentBox() (x, y, w, h float64) {
	x = b.X + b.MarginLeft + b.BorderLeft + b.PaddingLeft
	y = b.Y + b.MarginTop + b.BorderTop + b.PaddingTop
	w = b.Width
	h = b.Height
	return
}

// BorderBox returns the border box dimensions
func (b *Box) BorderBox() (x, y, w, h float64) {
	x = b.X + b.MarginLeft
	y = b.Y + b.MarginTop
	w = b.Width + b.PaddingLeft + b.PaddingRight + b.BorderLeft + b.BorderRight
	h = b.Height + b.PaddingTop + b.PaddingBottom + b.BorderTop + b.BorderBottom
	return
}

// MarginBox returns the full margin box dimensions
func (b *Box) MarginBox() (x, y, w, h float64) {
	x = b.X
	y = b.Y
	bx, by, bw, bh := b.BorderBox()
	_ = bx
	_ = by
	w = bw + b.MarginLeft + b.MarginRight
	h = bh + b.MarginTop + b.MarginBottom
	return
}

// Contains checks if a point is inside the border box
func (b *Box) Contains(px, py float64) bool {
	x, y, w, h := b.BorderBox()
	return px >= x && px <= x+w && py >= y && py <= y+h
}

// HandleClick processes click events recursively
func (b *Box) HandleClick(x, y float64, state *forms.FormState) bool {
	// Check children first (they're on top)
	for i := len(b.Children) - 1; i >= 0; i-- {
		if b.Children[i].HandleClick(x, y, state) {
			return true
		}
	}

	// Check this box
	if b.Contains(x, y) {
		if b.FormComponent != nil {
			bx, by, _, _ := b.ContentBox()
			return b.FormComponent.HandleClick(x-bx, y-by, state)
		}
	}
	return false
}

// FindByID finds a box by its ID
func (b *Box) FindByID(id string) *Box {
	if b.ID == id {
		return b
	}
	for _, child := range b.Children {
		if found := child.FindByID(id); found != nil {
			return found
		}
	}
	return nil
}

// FindLink finds a link at the given coordinates
func (b *Box) FindLink(x, y float64) string {
	// Check children first
	for _, child := range b.Children {
		if link := child.FindLink(x, y); link != "" {
			return link
		}
	}

	// Check this box
	if b.IsLink && b.LinkURL != "" && b.Contains(x, y) {
		return b.LinkURL
	}
	return ""
}

// Render draws this box (delegates to FormComponent if present)
func (b *Box) Render(screen *ebiten.Image, state *forms.FormState) {
	if b.FormComponent != nil {
		b.FormComponent.Render(screen, b, state)
	}
}
