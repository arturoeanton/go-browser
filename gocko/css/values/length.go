// Package values provides CSS value types for the Gocko CSS engine
// This implements W3C CSS Values and Units Level 4 specification
package values

import (
	"fmt"
	"strconv"
	"strings"
)

// =============================================================================
// LENGTH VALUES
// CSS Lengths: https://www.w3.org/TR/css-values-4/#lengths
// =============================================================================

// LengthUnit represents CSS length units
type LengthUnit int

const (
	UnitPx      LengthUnit = iota // Pixels (absolute)
	UnitEm                        // Relative to font-size
	UnitRem                       // Relative to root font-size
	UnitPercent                   // Percentage of parent
	UnitVw                        // 1% of viewport width
	UnitVh                        // 1% of viewport height
	UnitVmin                      // min(vw, vh)
	UnitVmax                      // max(vw, vh)
	UnitCh                        // Width of '0' character
	UnitEx                        // x-height of font
	UnitPt                        // Points (1pt = 1/72 inch)
	UnitCm                        // Centimeters
	UnitMm                        // Millimeters
	UnitIn                        // Inches
	UnitAuto                      // auto keyword
	UnitNone                      // none/unset
)

// Length represents a CSS length value with unit
type Length struct {
	Value float64
	Unit  LengthUnit
}

// Zero creates a zero length
func Zero() Length {
	return Length{Value: 0, Unit: UnitPx}
}

// Auto creates an auto length
func Auto() Length {
	return Length{Value: 0, Unit: UnitAuto}
}

// Px creates a pixel length
func Px(value float64) Length {
	return Length{Value: value, Unit: UnitPx}
}

// Em creates an em length
func Em(value float64) Length {
	return Length{Value: value, Unit: UnitEm}
}

// Rem creates a rem length
func Rem(value float64) Length {
	return Length{Value: value, Unit: UnitRem}
}

// Percent creates a percentage length
func Percent(value float64) Length {
	return Length{Value: value, Unit: UnitPercent}
}

// Vw creates a viewport width percentage
func Vw(value float64) Length {
	return Length{Value: value, Unit: UnitVw}
}

// Vh creates a viewport height percentage
func Vh(value float64) Length {
	return Length{Value: value, Unit: UnitVh}
}

// IsAuto returns true if this is an auto value
func (l Length) IsAuto() bool {
	return l.Unit == UnitAuto
}

// IsZero returns true if this is a zero length
func (l Length) IsZero() bool {
	return l.Value == 0 && l.Unit != UnitAuto
}

// ResolveContext contains the context needed to resolve relative lengths
type ResolveContext struct {
	FontSize       float64 // Current element font-size in px
	RootFontSize   float64 // :root font-size in px (default 16)
	ParentWidth    float64 // Parent width for percentage calculations
	ParentHeight   float64 // Parent height for percentage calculations
	ViewportWidth  float64 // Viewport width in px
	ViewportHeight float64 // Viewport height in px
	CharWidth      float64 // Width of '0' character (for ch unit)
	XHeight        float64 // x-height of font (for ex unit)
}

// DefaultContext returns a default resolution context
func DefaultContext() ResolveContext {
	return ResolveContext{
		FontSize:       16,
		RootFontSize:   16,
		ParentWidth:    1024,
		ParentHeight:   768,
		ViewportWidth:  1024,
		ViewportHeight: 768,
		CharWidth:      8,
		XHeight:        8,
	}
}

// Resolve converts any length to pixels using the given context
func (l Length) Resolve(ctx ResolveContext) float64 {
	switch l.Unit {
	case UnitPx:
		return l.Value
	case UnitEm:
		return l.Value * ctx.FontSize
	case UnitRem:
		return l.Value * ctx.RootFontSize
	case UnitPercent:
		return l.Value / 100 * ctx.ParentWidth
	case UnitVw:
		return l.Value / 100 * ctx.ViewportWidth
	case UnitVh:
		return l.Value / 100 * ctx.ViewportHeight
	case UnitVmin:
		min := ctx.ViewportWidth
		if ctx.ViewportHeight < min {
			min = ctx.ViewportHeight
		}
		return l.Value / 100 * min
	case UnitVmax:
		max := ctx.ViewportWidth
		if ctx.ViewportHeight > max {
			max = ctx.ViewportHeight
		}
		return l.Value / 100 * max
	case UnitCh:
		return l.Value * ctx.CharWidth
	case UnitEx:
		return l.Value * ctx.XHeight
	case UnitPt:
		return l.Value * 96 / 72 // 96 DPI standard
	case UnitCm:
		return l.Value * 96 / 2.54
	case UnitMm:
		return l.Value * 96 / 25.4
	case UnitIn:
		return l.Value * 96
	case UnitAuto, UnitNone:
		return 0
	}
	return l.Value
}

// ResolveHeight resolves a length using parent height (for height percentages)
func (l Length) ResolveHeight(ctx ResolveContext) float64 {
	if l.Unit == UnitPercent {
		return l.Value / 100 * ctx.ParentHeight
	}
	return l.Resolve(ctx)
}

// String returns a CSS representation of the length
func (l Length) String() string {
	if l.Unit == UnitAuto {
		return "auto"
	}
	if l.Unit == UnitNone {
		return "none"
	}
	units := []string{"px", "em", "rem", "%", "vw", "vh", "vmin", "vmax", "ch", "ex", "pt", "cm", "mm", "in"}
	if int(l.Unit) < len(units) {
		return fmt.Sprintf("%g%s", l.Value, units[l.Unit])
	}
	return fmt.Sprintf("%gpx", l.Value)
}

// ParseLength parses a CSS length string
func ParseLength(s string) (Length, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	if s == "" || s == "0" {
		return Zero(), nil
	}
	if s == "auto" {
		return Auto(), nil
	}
	if s == "none" {
		return Length{Unit: UnitNone}, nil
	}

	// Unit mappings
	units := map[string]LengthUnit{
		"px":   UnitPx,
		"em":   UnitEm,
		"rem":  UnitRem,
		"%":    UnitPercent,
		"vw":   UnitVw,
		"vh":   UnitVh,
		"vmin": UnitVmin,
		"vmax": UnitVmax,
		"ch":   UnitCh,
		"ex":   UnitEx,
		"pt":   UnitPt,
		"cm":   UnitCm,
		"mm":   UnitMm,
		"in":   UnitIn,
	}

	// Try to find and parse unit
	for suffix, unit := range units {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			value, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return Zero(), fmt.Errorf("invalid length value: %s", s)
			}
			return Length{Value: value, Unit: unit}, nil
		}
	}

	// Try parsing as number (implicit px)
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Zero(), fmt.Errorf("invalid length: %s", s)
	}
	return Px(value), nil
}
