// Package css provides the complete CSS engine for Gocko
// This package re-exports types from sub-packages for convenience
package css

import (
	"go-browser/gocko/css/properties"
	"go-browser/gocko/css/values"
)

// =============================================================================
// RE-EXPORTS FOR CONVENIENCE
// =============================================================================

// Type aliases for easy access
type (
	// Length represents a CSS length value
	Length = values.Length

	// Color represents a CSS color value
	Color = values.Color

	// ComputedStyle contains all computed CSS properties
	ComputedStyle = values.ComputedStyle

	// ResolveContext contains context for resolving relative values
	ResolveContext = values.ResolveContext
)

// =============================================================================
// VALUE CONSTRUCTORS
// =============================================================================

// Px creates a pixel length
func Px(value float64) Length {
	return values.Px(value)
}

// Em creates an em length
func Em(value float64) Length {
	return values.Em(value)
}

// Rem creates a rem length
func Rem(value float64) Length {
	return values.Rem(value)
}

// Percent creates a percentage length
func Percent(value float64) Length {
	return values.Percent(value)
}

// Vw creates a viewport width percentage
func Vw(value float64) Length {
	return values.Vw(value)
}

// Vh creates a viewport height percentage
func Vh(value float64) Length {
	return values.Vh(value)
}

// Auto creates an auto length
func Auto() Length {
	return values.Auto()
}

// Zero creates a zero length
func Zero() Length {
	return values.Zero()
}

// =============================================================================
// COLOR CONSTRUCTORS
// =============================================================================

// RGB creates a color from RGB values
func RGB(r, g, b uint8) Color {
	return values.RGB(r, g, b)
}

// RGBA creates a color from RGBA values
func RGBA(r, g, b, a uint8) Color {
	return values.RGBA(r, g, b, a)
}

// Black returns black
func Black() Color {
	return values.Black()
}

// White returns white
func White() Color {
	return values.White()
}

// Transparent returns transparent
func Transparent() Color {
	return values.Transparent()
}

// =============================================================================
// PARSERS
// =============================================================================

// ParseLength parses a CSS length string
func ParseLength(s string) (Length, error) {
	return values.ParseLength(s)
}

// ParseColor parses a CSS color string
func ParseColor(s string) (Color, error) {
	return values.ParseColor(s)
}

// =============================================================================
// COMPUTED STYLE
// =============================================================================

// NewComputedStyle creates a new computed style with defaults
func NewComputedStyle() *ComputedStyle {
	return values.NewComputedStyle()
}

// DefaultContext returns a default resolution context
func DefaultContext() ResolveContext {
	return values.DefaultContext()
}

// ParseProperty parses a CSS property and applies it to the style
func ParseProperty(style *ComputedStyle, property, value string) {
	properties.ParseProperty(style, property, value)
}
