// Package layout provides the layout engine for Gocko
// Box Model implementation follows CSS Box Model Module Level 3
// https://www.w3.org/TR/css-box-3/
package layout

import (
	"go-browser/gocko/css/values"
)

// =============================================================================
// BOX MODEL
// Implements CSS Box Model with margin, border, padding, content
// =============================================================================

// BoxDimensions holds all box model dimensions
type BoxDimensions struct {
	// Content box (innermost)
	ContentWidth  float64
	ContentHeight float64

	// Padding (between content and border)
	PaddingTop    float64
	PaddingRight  float64
	PaddingBottom float64
	PaddingLeft   float64

	// Border (between padding and margin)
	BorderTop    float64
	BorderRight  float64
	BorderBottom float64
	BorderLeft   float64

	// Margin (outermost)
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	MarginLeft   float64

	// Position
	X, Y float64
}

// NewBoxDimensions creates empty box dimensions
func NewBoxDimensions() *BoxDimensions {
	return &BoxDimensions{}
}

// PaddingWidth returns total horizontal padding
func (b *BoxDimensions) PaddingWidth() float64 {
	return b.PaddingLeft + b.PaddingRight
}

// PaddingHeight returns total vertical padding
func (b *BoxDimensions) PaddingHeight() float64 {
	return b.PaddingTop + b.PaddingBottom
}

// BorderWidth returns total horizontal border
func (b *BoxDimensions) BorderWidth() float64 {
	return b.BorderLeft + b.BorderRight
}

// BorderHeight returns total vertical border
func (b *BoxDimensions) BorderHeight() float64 {
	return b.BorderTop + b.BorderBottom
}

// MarginWidth returns total horizontal margin
func (b *BoxDimensions) MarginWidth() float64 {
	return b.MarginLeft + b.MarginRight
}

// MarginHeight returns total vertical margin
func (b *BoxDimensions) MarginHeight() float64 {
	return b.MarginTop + b.MarginBottom
}

// PaddingBoxWidth returns width including padding (content + padding)
func (b *BoxDimensions) PaddingBoxWidth() float64 {
	return b.ContentWidth + b.PaddingWidth()
}

// PaddingBoxHeight returns height including padding
func (b *BoxDimensions) PaddingBoxHeight() float64 {
	return b.ContentHeight + b.PaddingHeight()
}

// BorderBoxWidth returns width including border (content + padding + border)
func (b *BoxDimensions) BorderBoxWidth() float64 {
	return b.ContentWidth + b.PaddingWidth() + b.BorderWidth()
}

// BorderBoxHeight returns height including border
func (b *BoxDimensions) BorderBoxHeight() float64 {
	return b.ContentHeight + b.PaddingHeight() + b.BorderHeight()
}

// MarginBoxWidth returns total width (content + padding + border + margin)
func (b *BoxDimensions) MarginBoxWidth() float64 {
	return b.BorderBoxWidth() + b.MarginWidth()
}

// MarginBoxHeight returns total height
func (b *BoxDimensions) MarginBoxHeight() float64 {
	return b.BorderBoxHeight() + b.MarginHeight()
}

// ContentRect returns the content box rectangle (x, y, width, height)
func (b *BoxDimensions) ContentRect() (float64, float64, float64, float64) {
	return b.X + b.MarginLeft + b.BorderLeft + b.PaddingLeft,
		b.Y + b.MarginTop + b.BorderTop + b.PaddingTop,
		b.ContentWidth,
		b.ContentHeight
}

// PaddingRect returns the padding box rectangle
func (b *BoxDimensions) PaddingRect() (float64, float64, float64, float64) {
	return b.X + b.MarginLeft + b.BorderLeft,
		b.Y + b.MarginTop + b.BorderTop,
		b.PaddingBoxWidth(),
		b.PaddingBoxHeight()
}

// BorderRect returns the border box rectangle
func (b *BoxDimensions) BorderRect() (float64, float64, float64, float64) {
	return b.X + b.MarginLeft,
		b.Y + b.MarginTop,
		b.BorderBoxWidth(),
		b.BorderBoxHeight()
}

// MarginRect returns the margin box rectangle
func (b *BoxDimensions) MarginRect() (float64, float64, float64, float64) {
	return b.X, b.Y, b.MarginBoxWidth(), b.MarginBoxHeight()
}

// =============================================================================
// BOX MODEL CALCULATIONS
// =============================================================================

// ComputeBoxDimensions calculates all box dimensions from a computed style
func ComputeBoxDimensions(style *values.ComputedStyle, containingWidth, containingHeight float64) *BoxDimensions {
	box := NewBoxDimensions()

	// Build resolve context
	ctx := values.ResolveContext{
		FontSize:       style.FontSize,
		RootFontSize:   16,
		ParentWidth:    containingWidth,
		ParentHeight:   containingHeight,
		ViewportWidth:  1024,
		ViewportHeight: 768,
		CharWidth:      style.FontSize * 0.55,
		XHeight:        style.FontSize * 0.5,
	}

	// Resolve margins
	box.MarginTop = style.MarginTop.Resolve(ctx)
	box.MarginRight = style.MarginRight.Resolve(ctx)
	box.MarginBottom = style.MarginBottom.Resolve(ctx)
	box.MarginLeft = style.MarginLeft.Resolve(ctx)

	// Resolve padding
	box.PaddingTop = style.PaddingTop.Resolve(ctx)
	box.PaddingRight = style.PaddingRight.Resolve(ctx)
	box.PaddingBottom = style.PaddingBottom.Resolve(ctx)
	box.PaddingLeft = style.PaddingLeft.Resolve(ctx)

	// Resolve border widths
	box.BorderTop = style.BorderTopWidth.Resolve(ctx)
	box.BorderRight = style.BorderRightWidth.Resolve(ctx)
	box.BorderBottom = style.BorderBottomWidth.Resolve(ctx)
	box.BorderLeft = style.BorderLeftWidth.Resolve(ctx)

	// Calculate content width
	if style.Width.IsAuto() {
		// Auto width: fill available space
		availableWidth := containingWidth - box.MarginWidth() - box.BorderWidth() - box.PaddingWidth()
		box.ContentWidth = availableWidth
	} else {
		resolvedWidth := style.Width.Resolve(ctx)
		if style.BoxSizing == "border-box" {
			// Width includes padding and border
			box.ContentWidth = resolvedWidth - box.PaddingWidth() - box.BorderWidth()
			if box.ContentWidth < 0 {
				box.ContentWidth = 0
			}
		} else {
			// content-box (default): width is content only
			box.ContentWidth = resolvedWidth
		}
	}

	// Apply min/max width
	if style.MinWidth.Unit != values.UnitNone {
		minW := style.MinWidth.Resolve(ctx)
		if box.ContentWidth < minW {
			box.ContentWidth = minW
		}
	}
	if style.MaxWidth.Unit != values.UnitNone {
		maxW := style.MaxWidth.Resolve(ctx)
		if box.ContentWidth > maxW {
			box.ContentWidth = maxW
		}
	}

	// Calculate content height
	if style.Height.IsAuto() {
		// Auto height: determined by content (will be set later)
		box.ContentHeight = 0
	} else {
		resolvedHeight := style.Height.ResolveHeight(ctx)
		if style.BoxSizing == "border-box" {
			box.ContentHeight = resolvedHeight - box.PaddingHeight() - box.BorderHeight()
			if box.ContentHeight < 0 {
				box.ContentHeight = 0
			}
		} else {
			box.ContentHeight = resolvedHeight
		}
	}

	// Apply min/max height
	if style.MinHeight.Unit != values.UnitNone {
		minH := style.MinHeight.ResolveHeight(ctx)
		if box.ContentHeight < minH {
			box.ContentHeight = minH
		}
	}
	if style.MaxHeight.Unit != values.UnitNone {
		maxH := style.MaxHeight.ResolveHeight(ctx)
		if box.ContentHeight > maxH {
			box.ContentHeight = maxH
		}
	}

	return box
}

// SetContentHeight sets the content height (after content is laid out)
func (b *BoxDimensions) SetContentHeight(height float64) {
	b.ContentHeight = height
}

// SetPosition sets the box position
func (b *BoxDimensions) SetPosition(x, y float64) {
	b.X = x
	b.Y = y
}

// =============================================================================
// MARGIN COLLAPSING
// CSS margin collapsing (simplified implementation)
// =============================================================================

// CollapseMargins returns the collapsed margin between two adjacent boxes
func CollapseMargins(margin1, margin2 float64) float64 {
	// Both positive: take the larger
	if margin1 >= 0 && margin2 >= 0 {
		if margin1 > margin2 {
			return margin1
		}
		return margin2
	}

	// Both negative: take the more negative (smaller)
	if margin1 < 0 && margin2 < 0 {
		if margin1 < margin2 {
			return margin1
		}
		return margin2
	}

	// One positive, one negative: add them
	return margin1 + margin2
}
