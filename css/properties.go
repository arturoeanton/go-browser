// Package css provides CSS parsing, selector matching, and style computation
package css

import (
	"image/color"
	"strconv"
	"strings"
)

// ======================================================================================
// CSS VALUES
// ======================================================================================

// Unit represents a CSS unit type
type Unit int

const (
	UnitNone Unit = iota
	UnitPx
	UnitEm
	UnitRem
	UnitPercent
	UnitVw
	UnitVh
)

// Value represents a CSS value
type Value struct {
	Type    ValueType
	Keyword string
	Number  float64
	Unit    Unit
	Color   color.RGBA
}

// ValueType indicates what kind of value this is
type ValueType int

const (
	ValueNone ValueType = iota
	ValueKeyword
	ValueLength
	ValueColor
	ValueNumber
	ValueInherit
	ValueInitial
)

// ======================================================================================
// CSS GRADIENTS
// ======================================================================================

// GradientStop represents a color stop in a gradient
type GradientStop struct {
	Color    color.RGBA
	Position float64 // 0.0 to 1.0
}

// Gradient represents a CSS gradient (linear or radial)
type Gradient struct {
	IsLinear bool    // true for linear-gradient, false for radial
	Angle    float64 // degrees (0 = to top, 90 = to right, etc.)
	Stops    []GradientStop
}

// ======================================================================================
// COMPUTED STYLE
// ======================================================================================

// ComputedStyle holds all computed CSS properties for a node
type ComputedStyle struct {
	// Display
	Display    string // block, inline, none, flex, grid, inline-flex, inline-block
	Visibility string // visible, hidden

	// Flexbox
	FlexDirection  string  // row, row-reverse, column, column-reverse
	JustifyContent string  // flex-start, flex-end, center, space-between, space-around, space-evenly
	AlignItems     string  // flex-start, flex-end, center, stretch, baseline
	AlignContent   string  // flex-start, flex-end, center, stretch, space-between, space-around
	FlexWrap       string  // nowrap, wrap, wrap-reverse
	Gap            float64 // gap between flex/grid items
	RowGap         float64
	ColumnGap      float64
	FlexGrow       float64
	FlexShrink     float64
	FlexBasis      float64
	AlignSelf      string // auto, flex-start, flex-end, center, stretch
	Order          int

	// CSS Grid
	GridTemplateColumns string // e.g., "1fr 1fr 1fr", "repeat(3, 1fr)", "200px auto"
	GridTemplateRows    string
	GridColumn          string // e.g., "1 / 3", "span 2"
	GridRow             string
	GridColumnCount     int // Parsed number of columns

	// Colors and Backgrounds
	Color              color.RGBA
	BackgroundColor    color.RGBA
	BackgroundGradient *Gradient // For linear-gradient, radial-gradient

	// Typography
	FontSize   float64
	FontWeight int // 100-900
	FontFamily string
	TextAlign  string // left, center, right, justify
	LineHeight float64

	// Box Model (in pixels)
	Width     float64
	Height    float64
	MinWidth  float64
	MinHeight float64
	MaxWidth  float64
	MaxHeight float64

	// Margins
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	MarginLeft   float64

	// Padding
	PaddingTop    float64
	PaddingRight  float64
	PaddingBottom float64
	PaddingLeft   float64

	// Borders
	BorderTopWidth    float64
	BorderRightWidth  float64
	BorderBottomWidth float64
	BorderLeftWidth   float64
	BorderColor       color.RGBA
	BorderRadius      float64

	// Position
	Position string // static, relative, absolute, fixed
	Top      float64
	Right    float64
	Bottom   float64
	Left     float64
	ZIndex   int
}

// NewComputedStyle creates a ComputedStyle with default values
func NewComputedStyle() *ComputedStyle {
	return &ComputedStyle{
		Display:         "inline",
		Visibility:      "visible",
		Color:           color.RGBA{0, 0, 0, 255},
		BackgroundColor: color.RGBA{0, 0, 0, 0}, // transparent
		FontSize:        16,
		FontWeight:      400,
		FontFamily:      "sans-serif",
		TextAlign:       "left",
		LineHeight:      1.2,
		Position:        "static",
	}
}

// DefaultForTag returns default styles for HTML tags
func DefaultForTag(tag string) *ComputedStyle {
	style := NewComputedStyle()

	switch tag {
	case "div", "section", "article", "header", "footer", "nav", "main",
		"ul", "ol", "li", "form", "table", "tr", "blockquote", "pre":
		style.Display = "block"
	case "h1":
		style.Display = "block"
		style.FontSize = 32
		style.FontWeight = 700
		style.MarginTop = 21
		style.MarginBottom = 21
	case "h2":
		style.Display = "block"
		style.FontSize = 24
		style.FontWeight = 700
		style.MarginTop = 19
		style.MarginBottom = 19
	case "h3":
		style.Display = "block"
		style.FontSize = 18
		style.FontWeight = 700
		style.MarginTop = 18
		style.MarginBottom = 18
	case "h4", "h5", "h6":
		style.Display = "block"
		style.FontSize = 16
		style.FontWeight = 700
		style.MarginTop = 16
		style.MarginBottom = 16
	case "p":
		style.Display = "block"
		style.MarginTop = 16
		style.MarginBottom = 16
	case "a":
		style.Color = color.RGBA{0, 0, 238, 255} // blue
	case "b", "strong":
		style.FontWeight = 700
	case "i", "em":
		// italic would be handled separately
	case "button":
		style.Display = "inline-block"
		style.PaddingTop = 8
		style.PaddingBottom = 8
		style.PaddingLeft = 16
		style.PaddingRight = 16
		style.BackgroundColor = color.RGBA{240, 240, 240, 255}
		style.BorderRadius = 4
	}

	return style
}

// ======================================================================================
// COLOR PARSING
// ======================================================================================

// Named colors (subset of CSS named colors)
var namedColors = map[string]color.RGBA{
	// Basic colors
	"black":       {0, 0, 0, 255},
	"white":       {255, 255, 255, 255},
	"red":         {255, 0, 0, 255},
	"green":       {0, 128, 0, 255},
	"blue":        {0, 0, 255, 255},
	"yellow":      {255, 255, 0, 255},
	"cyan":        {0, 255, 255, 255},
	"magenta":     {255, 0, 255, 255},
	"gray":        {128, 128, 128, 255},
	"grey":        {128, 128, 128, 255},
	"transparent": {0, 0, 0, 0},

	// Extended colors
	"orange":     {255, 165, 0, 255},
	"purple":     {128, 0, 128, 255},
	"pink":       {255, 192, 203, 255},
	"brown":      {165, 42, 42, 255},
	"navy":       {0, 0, 128, 255},
	"teal":       {0, 128, 128, 255},
	"olive":      {128, 128, 0, 255},
	"maroon":     {128, 0, 0, 255},
	"lime":       {0, 255, 0, 255},
	"aqua":       {0, 255, 255, 255},
	"silver":     {192, 192, 192, 255},
	"fuchsia":    {255, 0, 255, 255},
	"lightgray":  {211, 211, 211, 255},
	"lightgrey":  {211, 211, 211, 255},
	"darkgray":   {169, 169, 169, 255},
	"darkgrey":   {169, 169, 169, 255},
	"whitesmoke": {245, 245, 245, 255},
	"inherit":    {0, 0, 0, 0}, // special handling
}

// ParseColor parses a CSS color value
func ParseColor(value string) (color.RGBA, bool) {
	value = strings.ToLower(strings.TrimSpace(value))

	// Named color
	if c, ok := namedColors[value]; ok {
		return c, true
	}

	// Hex color #RGB or #RRGGBB
	if strings.HasPrefix(value, "#") {
		hex := value[1:]
		if len(hex) == 3 {
			// #RGB -> #RRGGBB
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) == 6 {
			r, _ := strconv.ParseUint(hex[0:2], 16, 8)
			g, _ := strconv.ParseUint(hex[2:4], 16, 8)
			b, _ := strconv.ParseUint(hex[4:6], 16, 8)
			return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, true
		}
		if len(hex) == 8 {
			r, _ := strconv.ParseUint(hex[0:2], 16, 8)
			g, _ := strconv.ParseUint(hex[2:4], 16, 8)
			b, _ := strconv.ParseUint(hex[4:6], 16, 8)
			a, _ := strconv.ParseUint(hex[6:8], 16, 8)
			return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, true
		}
	}

	// rgb(r, g, b) or rgba(r, g, b, a)
	if strings.HasPrefix(value, "rgb") {
		start := strings.Index(value, "(")
		end := strings.LastIndex(value, ")")
		if start != -1 && end > start {
			parts := strings.Split(value[start+1:end], ",")
			if len(parts) >= 3 {
				r, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
				g, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
				b, _ := strconv.Atoi(strings.TrimSpace(parts[2]))
				a := 255
				if len(parts) >= 4 {
					af, _ := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
					if af <= 1 {
						a = int(af * 255)
					} else {
						a = int(af)
					}
				}
				return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, true
			}
		}
	}

	return color.RGBA{}, false
}

// ParseLength parses a CSS length value (e.g., "16px", "1.5em")
func ParseLength(value string) (float64, Unit, bool) {
	value = strings.ToLower(strings.TrimSpace(value))

	if value == "0" {
		return 0, UnitPx, true
	}

	// Try different units
	units := map[string]Unit{
		"px":  UnitPx,
		"em":  UnitEm,
		"rem": UnitRem,
		"%":   UnitPercent,
		"vw":  UnitVw,
		"vh":  UnitVh,
	}

	for suffix, unit := range units {
		if strings.HasSuffix(value, suffix) {
			numStr := strings.TrimSuffix(value, suffix)
			num, err := strconv.ParseFloat(numStr, 64)
			if err == nil {
				return num, unit, true
			}
		}
	}

	// Try bare number (treated as px)
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num, UnitPx, true
	}

	return 0, UnitNone, false
}

// ParseGradient parses a CSS gradient value (linear-gradient, radial-gradient)
func ParseGradient(value string) (*Gradient, bool) {
	value = strings.TrimSpace(value)

	// Check for linear-gradient
	if strings.HasPrefix(value, "linear-gradient(") {
		inner := strings.TrimPrefix(value, "linear-gradient(")
		inner = strings.TrimSuffix(inner, ")")
		return parseLinearGradient(inner)
	}

	return nil, false
}

// parseLinearGradient parses the inner content of linear-gradient()
func parseLinearGradient(inner string) (*Gradient, bool) {
	gradient := &Gradient{
		IsLinear: true,
		Angle:    180, // default: to bottom
		Stops:    []GradientStop{},
	}

	// Split by comma, but respect nested parentheses
	parts := splitGradientParts(inner)
	if len(parts) == 0 {
		return nil, false
	}

	startIdx := 0

	// Check if first part is a direction/angle
	firstPart := strings.TrimSpace(parts[0])
	if strings.HasPrefix(firstPart, "to ") {
		// Direction keywords
		direction := strings.TrimPrefix(firstPart, "to ")
		switch direction {
		case "top":
			gradient.Angle = 0
		case "right":
			gradient.Angle = 90
		case "bottom":
			gradient.Angle = 180
		case "left":
			gradient.Angle = 270
		case "top right", "right top":
			gradient.Angle = 45
		case "bottom right", "right bottom":
			gradient.Angle = 135
		case "bottom left", "left bottom":
			gradient.Angle = 225
		case "top left", "left top":
			gradient.Angle = 315
		}
		startIdx = 1
	} else if strings.HasSuffix(firstPart, "deg") {
		// Angle in degrees
		angleStr := strings.TrimSuffix(firstPart, "deg")
		if angle, err := strconv.ParseFloat(angleStr, 64); err == nil {
			gradient.Angle = angle
		}
		startIdx = 1
	}

	// Parse color stops
	numStops := len(parts) - startIdx
	for i := startIdx; i < len(parts); i++ {
		stop := parseColorStop(strings.TrimSpace(parts[i]), i-startIdx, numStops)
		if stop != nil {
			gradient.Stops = append(gradient.Stops, *stop)
		}
	}

	if len(gradient.Stops) < 2 {
		return nil, false
	}

	return gradient, true
}

// parseColorStop parses a gradient color stop
func parseColorStop(value string, index, total int) *GradientStop {
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return nil
	}

	color, ok := ParseColor(parts[0])
	if !ok {
		return nil
	}

	position := float64(index) / float64(total-1)
	if len(parts) > 1 {
		// Has explicit position
		posStr := parts[1]
		if strings.HasSuffix(posStr, "%") {
			if p, err := strconv.ParseFloat(strings.TrimSuffix(posStr, "%"), 64); err == nil {
				position = p / 100
			}
		}
	}

	return &GradientStop{
		Color:    color,
		Position: position,
	}
}

// splitGradientParts splits gradient parts respecting nested parentheses
func splitGradientParts(s string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, c := range s {
		if c == '(' {
			depth++
			current.WriteRune(c)
		} else if c == ')' {
			depth--
			current.WriteRune(c)
		} else if c == ',' && depth == 0 {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
