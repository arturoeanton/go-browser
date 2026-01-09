// Package values provides CSS value types for the Gocko CSS engine
// ComputedStyle represents the fully resolved style for an element
package values

import (
	"image/color"
)

// =============================================================================
// COMPUTED STYLE
// The final resolved style for an element after cascade, inheritance, and
// value computation. This is what's used for layout and painting.
// =============================================================================

// ComputedStyle contains all computed CSS property values for an element
type ComputedStyle struct {
	// =========================
	// BOX MODEL
	// =========================

	// Dimensions
	Width     Length
	Height    Length
	MinWidth  Length
	MaxWidth  Length
	MinHeight Length
	MaxHeight Length

	// Margins
	MarginTop    Length
	MarginRight  Length
	MarginBottom Length
	MarginLeft   Length

	// Padding
	PaddingTop    Length
	PaddingRight  Length
	PaddingBottom Length
	PaddingLeft   Length

	// Border widths
	BorderTopWidth    Length
	BorderRightWidth  Length
	BorderBottomWidth Length
	BorderLeftWidth   Length

	// Border colors
	BorderTopColor    Color
	BorderRightColor  Color
	BorderBottomColor Color
	BorderLeftColor   Color

	// Border styles
	BorderTopStyle    string // none, solid, dashed, dotted, etc.
	BorderRightStyle  string
	BorderBottomStyle string
	BorderLeftStyle   string

	// Border radius
	BorderTopLeftRadius     Length
	BorderTopRightRadius    Length
	BorderBottomRightRadius Length
	BorderBottomLeftRadius  Length

	// Box sizing
	BoxSizing string // content-box, border-box

	// =========================
	// LAYOUT
	// =========================

	Display  string // block, inline, flex, grid, none, etc.
	Position string // static, relative, absolute, fixed, sticky

	// Positioned offsets
	Top    Length
	Right  Length
	Bottom Length
	Left   Length

	// Flex container properties
	FlexDirection  string // row, column, row-reverse, column-reverse
	FlexWrap       string // nowrap, wrap, wrap-reverse
	JustifyContent string // flex-start, center, flex-end, space-between, space-around
	AlignItems     string // flex-start, center, flex-end, stretch, baseline
	AlignContent   string // flex-start, center, flex-end, stretch, space-between
	Gap            Length // gap between flex/grid items

	// Flex item properties
	FlexGrow   float64
	FlexShrink float64
	FlexBasis  Length
	AlignSelf  string // auto, flex-start, center, flex-end, stretch
	Order      int

	// Overflow
	OverflowX string // visible, hidden, scroll, auto
	OverflowY string

	// Z-index
	ZIndex    int
	ZIndexSet bool // true if z-index was explicitly set

	// =========================
	// TYPOGRAPHY
	// =========================

	Color          Color
	FontFamily     string  // font stack
	FontSize       float64 // computed in pixels
	FontWeight     int     // 100-900
	FontStyle      string  // normal, italic, oblique
	LineHeight     float64 // computed in pixels or ratio
	LineHeightUnit string  // "px" or "number"
	TextAlign      string  // left, center, right, justify
	TextDecoration string  // none, underline, line-through, overline
	TextTransform  string  // none, uppercase, lowercase, capitalize
	LetterSpacing  Length
	WordSpacing    Length
	WhiteSpace     string // normal, nowrap, pre, pre-wrap, pre-line

	// =========================
	// VISUAL
	// =========================

	BackgroundColor    Color
	BackgroundImage    string // url() or gradient
	BackgroundSize     string // cover, contain, or length
	BackgroundPosition string
	BackgroundRepeat   string

	Opacity    float64 // 0-1
	Visibility string  // visible, hidden, collapse

	// Shadow
	BoxShadow string // full shadow definition

	// Cursor
	Cursor string // default, pointer, text, etc.

	// Transform
	Transform string // transform functions

	// =========================
	// LIST
	// =========================
	ListStyleType     string // disc, circle, square, decimal, none, etc.
	ListStylePosition string // inside, outside

	// =========================
	// TABLE
	// =========================
	BorderCollapse string // separate, collapse
	BorderSpacing  Length
}

// NewComputedStyle creates a new computed style with default values
func NewComputedStyle() *ComputedStyle {
	return &ComputedStyle{
		// Box Model defaults
		Width:     Auto(),
		Height:    Auto(),
		MinWidth:  Zero(),
		MaxWidth:  Length{Unit: UnitNone},
		MinHeight: Zero(),
		MaxHeight: Length{Unit: UnitNone},
		BoxSizing: "content-box",

		// Layout defaults
		Display:        "inline",
		Position:       "static",
		FlexDirection:  "row",
		FlexWrap:       "nowrap",
		JustifyContent: "flex-start",
		AlignItems:     "stretch",
		AlignContent:   "stretch",
		FlexGrow:       0,
		FlexShrink:     1,
		FlexBasis:      Auto(),
		AlignSelf:      "auto",
		OverflowX:      "visible",
		OverflowY:      "visible",

		// Typography defaults
		Color:          Black(),
		FontFamily:     "sans-serif",
		FontSize:       16,
		FontWeight:     400,
		FontStyle:      "normal",
		LineHeight:     1.2,
		LineHeightUnit: "number",
		TextAlign:      "start",
		TextDecoration: "none",
		TextTransform:  "none",
		WhiteSpace:     "normal",

		// Visual defaults
		BackgroundColor: Transparent(),
		Opacity:         1.0,
		Visibility:      "visible",
		Cursor:          "auto",

		// List defaults
		ListStyleType:     "disc",
		ListStylePosition: "outside",

		// Table defaults
		BorderCollapse: "separate",
	}
}

// Clone creates a deep copy of the computed style
func (cs *ComputedStyle) Clone() *ComputedStyle {
	clone := *cs
	return &clone
}

// GetColor returns the text color as Go's color.RGBA
func (cs *ComputedStyle) GetColor() color.RGBA {
	return cs.Color.ToRGBA()
}

// GetBackgroundColor returns the background color as Go's color.RGBA
func (cs *ComputedStyle) GetBackgroundColor() color.RGBA {
	return cs.BackgroundColor.ToRGBA()
}

// IsBlock returns true if this is a block-level element
func (cs *ComputedStyle) IsBlock() bool {
	switch cs.Display {
	case "block", "flex", "grid", "table", "list-item":
		return true
	}
	return false
}

// IsInline returns true if this is an inline element
func (cs *ComputedStyle) IsInline() bool {
	return cs.Display == "inline" || cs.Display == "inline-block"
}

// IsFlex returns true if this is a flex container
func (cs *ComputedStyle) IsFlex() bool {
	return cs.Display == "flex" || cs.Display == "inline-flex"
}

// IsHidden returns true if element should not be rendered
func (cs *ComputedStyle) IsHidden() bool {
	return cs.Display == "none" || cs.Visibility == "hidden"
}

// IsPositioned returns true if element is not statically positioned
func (cs *ComputedStyle) IsPositioned() bool {
	return cs.Position != "static"
}

// GetMargin returns all margins as a 4-value array (top, right, bottom, left)
func (cs *ComputedStyle) GetMargin() [4]Length {
	return [4]Length{cs.MarginTop, cs.MarginRight, cs.MarginBottom, cs.MarginLeft}
}

// GetPadding returns all padding as a 4-value array
func (cs *ComputedStyle) GetPadding() [4]Length {
	return [4]Length{cs.PaddingTop, cs.PaddingRight, cs.PaddingBottom, cs.PaddingLeft}
}

// GetBorderWidth returns all border widths as a 4-value array
func (cs *ComputedStyle) GetBorderWidth() [4]Length {
	return [4]Length{cs.BorderTopWidth, cs.BorderRightWidth, cs.BorderBottomWidth, cs.BorderLeftWidth}
}

// ResolveWidth resolves the width using the given context
func (cs *ComputedStyle) ResolveWidth(ctx ResolveContext) float64 {
	if cs.Width.IsAuto() {
		return ctx.ParentWidth
	}
	return cs.Width.Resolve(ctx)
}

// ResolveHeight resolves the height using the given context
func (cs *ComputedStyle) ResolveHeight(ctx ResolveContext) float64 {
	if cs.Height.IsAuto() {
		return -1 // Auto height should be computed from content
	}
	return cs.Height.ResolveHeight(ctx)
}
