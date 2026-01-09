// Package properties provides CSS property parsing and application
// Implements property-specific parsing as per CSS specifications
package properties

import (
	"strconv"
	"strings"

	"go-browser/gocko/css/values"
)

// =============================================================================
// PROPERTY PARSER
// Central function for parsing any CSS property value into the appropriate type
// =============================================================================

// ParseProperty parses a CSS property value and applies it to the computed style
func ParseProperty(style *values.ComputedStyle, property, value string) {
	property = strings.ToLower(strings.TrimSpace(property))
	value = strings.TrimSpace(value)

	switch property {
	// =========================
	// DIMENSIONS
	// =========================
	case "width":
		style.Width = parseLength(value)
	case "height":
		style.Height = parseLength(value)
	case "min-width":
		style.MinWidth = parseLength(value)
	case "max-width":
		style.MaxWidth = parseLength(value)
	case "min-height":
		style.MinHeight = parseLength(value)
	case "max-height":
		style.MaxHeight = parseLength(value)

	// =========================
	// MARGIN
	// =========================
	case "margin":
		t, r, b, l := parseBoxShorthand(value)
		style.MarginTop, style.MarginRight = t, r
		style.MarginBottom, style.MarginLeft = b, l
	case "margin-top":
		style.MarginTop = parseLength(value)
	case "margin-right":
		style.MarginRight = parseLength(value)
	case "margin-bottom":
		style.MarginBottom = parseLength(value)
	case "margin-left":
		style.MarginLeft = parseLength(value)

	// =========================
	// PADDING
	// =========================
	case "padding":
		t, r, b, l := parseBoxShorthand(value)
		style.PaddingTop, style.PaddingRight = t, r
		style.PaddingBottom, style.PaddingLeft = b, l
	case "padding-top":
		style.PaddingTop = parseLength(value)
	case "padding-right":
		style.PaddingRight = parseLength(value)
	case "padding-bottom":
		style.PaddingBottom = parseLength(value)
	case "padding-left":
		style.PaddingLeft = parseLength(value)

	// =========================
	// BORDER
	// =========================
	case "border":
		parseBorderShorthand(style, value)
	case "border-width":
		w := parseLength(value)
		style.BorderTopWidth = w
		style.BorderRightWidth = w
		style.BorderBottomWidth = w
		style.BorderLeftWidth = w
	case "border-top-width":
		style.BorderTopWidth = parseLength(value)
	case "border-right-width":
		style.BorderRightWidth = parseLength(value)
	case "border-bottom-width":
		style.BorderBottomWidth = parseLength(value)
	case "border-left-width":
		style.BorderLeftWidth = parseLength(value)
	case "border-color":
		c := parseColor(value)
		style.BorderTopColor = c
		style.BorderRightColor = c
		style.BorderBottomColor = c
		style.BorderLeftColor = c
	case "border-style":
		style.BorderTopStyle = value
		style.BorderRightStyle = value
		style.BorderBottomStyle = value
		style.BorderLeftStyle = value
	case "border-radius":
		r := parseLength(value)
		style.BorderTopLeftRadius = r
		style.BorderTopRightRadius = r
		style.BorderBottomRightRadius = r
		style.BorderBottomLeftRadius = r
	case "border-top-left-radius":
		style.BorderTopLeftRadius = parseLength(value)
	case "border-top-right-radius":
		style.BorderTopRightRadius = parseLength(value)
	case "border-bottom-left-radius":
		style.BorderBottomLeftRadius = parseLength(value)
	case "border-bottom-right-radius":
		style.BorderBottomRightRadius = parseLength(value)

	// =========================
	// BOX
	// =========================
	case "box-sizing":
		style.BoxSizing = value

	// =========================
	// LAYOUT
	// =========================
	case "display":
		style.Display = value
	case "position":
		style.Position = value
	case "top":
		style.Top = parseLength(value)
	case "right":
		style.Right = parseLength(value)
	case "bottom":
		style.Bottom = parseLength(value)
	case "left":
		style.Left = parseLength(value)
	case "z-index":
		if v, err := strconv.Atoi(value); err == nil {
			style.ZIndex = v
			style.ZIndexSet = true
		}

	// =========================
	// FLEXBOX
	// =========================
	case "flex":
		parseFlexShorthand(style, value)
	case "flex-direction":
		style.FlexDirection = value
	case "flex-wrap":
		style.FlexWrap = value
	case "flex-flow":
		parts := strings.Fields(value)
		for _, p := range parts {
			if p == "row" || p == "column" || p == "row-reverse" || p == "column-reverse" {
				style.FlexDirection = p
			} else if p == "wrap" || p == "nowrap" || p == "wrap-reverse" {
				style.FlexWrap = p
			}
		}
	case "justify-content":
		style.JustifyContent = value
	case "align-items":
		style.AlignItems = value
	case "align-content":
		style.AlignContent = value
	case "align-self":
		style.AlignSelf = value
	case "flex-grow":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			style.FlexGrow = v
		}
	case "flex-shrink":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			style.FlexShrink = v
		}
	case "flex-basis":
		style.FlexBasis = parseLength(value)
	case "order":
		if v, err := strconv.Atoi(value); err == nil {
			style.Order = v
		}
	case "gap":
		style.Gap = parseLength(value)

	// =========================
	// OVERFLOW
	// =========================
	case "overflow":
		style.OverflowX = value
		style.OverflowY = value
	case "overflow-x":
		style.OverflowX = value
	case "overflow-y":
		style.OverflowY = value

	// =========================
	// TYPOGRAPHY
	// =========================
	case "color":
		style.Color = parseColor(value)
	case "font-family":
		style.FontFamily = value
	case "font-size":
		style.FontSize = parseFontSize(value)
	case "font-weight":
		style.FontWeight = parseFontWeight(value)
	case "font-style":
		style.FontStyle = value
	case "line-height":
		parseLineHeight(style, value)
	case "text-align":
		style.TextAlign = value
	case "text-decoration":
		style.TextDecoration = value
	case "text-transform":
		style.TextTransform = value
	case "letter-spacing":
		style.LetterSpacing = parseLength(value)
	case "word-spacing":
		style.WordSpacing = parseLength(value)
	case "white-space":
		style.WhiteSpace = value

	// =========================
	// VISUAL
	// =========================
	case "background":
		parseBackgroundShorthand(style, value)
	case "background-color":
		style.BackgroundColor = parseColor(value)
	case "background-image":
		style.BackgroundImage = value
	case "background-size":
		style.BackgroundSize = value
	case "background-position":
		style.BackgroundPosition = value
	case "background-repeat":
		style.BackgroundRepeat = value
	case "opacity":
		if v, err := strconv.ParseFloat(value, 64); err == nil {
			style.Opacity = v
		}
	case "visibility":
		style.Visibility = value
	case "box-shadow":
		style.BoxShadow = value
	case "cursor":
		style.Cursor = value
	case "transform":
		style.Transform = value

	// =========================
	// LIST
	// =========================
	case "list-style-type":
		style.ListStyleType = value
	case "list-style-position":
		style.ListStylePosition = value
	case "list-style":
		parts := strings.Fields(value)
		for _, p := range parts {
			if p == "inside" || p == "outside" {
				style.ListStylePosition = p
			} else if p != "none" {
				style.ListStyleType = p
			}
		}

	// =========================
	// TABLE
	// =========================
	case "border-collapse":
		style.BorderCollapse = value
	case "border-spacing":
		style.BorderSpacing = parseLength(value)
	}
}

// =============================================================================
// HELPER PARSERS
// =============================================================================

func parseLength(s string) values.Length {
	l, _ := values.ParseLength(s)
	return l
}

func parseColor(s string) values.Color {
	c, _ := values.ParseColor(s)
	return c
}

// parseBoxShorthand parses margin/padding shorthand (1-4 values)
func parseBoxShorthand(value string) (top, right, bottom, left values.Length) {
	parts := strings.Fields(value)
	switch len(parts) {
	case 1:
		v := parseLength(parts[0])
		return v, v, v, v
	case 2:
		tb := parseLength(parts[0])
		lr := parseLength(parts[1])
		return tb, lr, tb, lr
	case 3:
		t := parseLength(parts[0])
		lr := parseLength(parts[1])
		b := parseLength(parts[2])
		return t, lr, b, lr
	case 4:
		return parseLength(parts[0]), parseLength(parts[1]),
			parseLength(parts[2]), parseLength(parts[3])
	}
	return values.Zero(), values.Zero(), values.Zero(), values.Zero()
}

// parseBorderShorthand parses "border: 1px solid red"
func parseBorderShorthand(style *values.ComputedStyle, value string) {
	parts := strings.Fields(value)
	for _, p := range parts {
		// Try as width
		if l, err := values.ParseLength(p); err == nil && !l.IsZero() {
			style.BorderTopWidth = l
			style.BorderRightWidth = l
			style.BorderBottomWidth = l
			style.BorderLeftWidth = l
			continue
		}
		// Try as style
		if isBorderStyle(p) {
			style.BorderTopStyle = p
			style.BorderRightStyle = p
			style.BorderBottomStyle = p
			style.BorderLeftStyle = p
			continue
		}
		// Try as color
		if c, err := values.ParseColor(p); err == nil {
			style.BorderTopColor = c
			style.BorderRightColor = c
			style.BorderBottomColor = c
			style.BorderLeftColor = c
		}
	}
}

func isBorderStyle(s string) bool {
	styles := []string{"none", "solid", "dashed", "dotted", "double", "groove", "ridge", "inset", "outset"}
	for _, style := range styles {
		if s == style {
			return true
		}
	}
	return false
}

// parseFlexShorthand parses "flex: 1 1 auto" or "flex: 1"
func parseFlexShorthand(style *values.ComputedStyle, value string) {
	if value == "none" {
		style.FlexGrow = 0
		style.FlexShrink = 0
		style.FlexBasis = values.Auto()
		return
	}
	if value == "auto" {
		style.FlexGrow = 1
		style.FlexShrink = 1
		style.FlexBasis = values.Auto()
		return
	}

	parts := strings.Fields(value)
	if len(parts) >= 1 {
		if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
			style.FlexGrow = v
		}
	}
	if len(parts) >= 2 {
		if v, err := strconv.ParseFloat(parts[1], 64); err == nil {
			style.FlexShrink = v
		}
	}
	if len(parts) >= 3 {
		style.FlexBasis = parseLength(parts[2])
	}
}

// parseFontSize parses font size with keyword support
func parseFontSize(value string) float64 {
	// Named sizes
	sizes := map[string]float64{
		"xx-small": 9, "x-small": 10, "small": 13, "medium": 16,
		"large": 18, "x-large": 24, "xx-large": 32, "xxx-large": 48,
		"smaller": 13, "larger": 19,
	}
	if size, ok := sizes[value]; ok {
		return size
	}

	// Parse as length
	l := parseLength(value)
	if l.Unit == values.UnitPercent {
		return l.Value / 100 * 16 // 16px base
	}
	return l.Resolve(values.DefaultContext())
}

// parseFontWeight parses font weight (number or keyword)
func parseFontWeight(value string) int {
	weights := map[string]int{
		"normal": 400, "bold": 700, "lighter": 300, "bolder": 600,
	}
	if w, ok := weights[value]; ok {
		return w
	}
	if w, err := strconv.Atoi(value); err == nil {
		return w
	}
	return 400
}

// parseLineHeight handles line-height values
func parseLineHeight(style *values.ComputedStyle, value string) {
	if value == "normal" {
		style.LineHeight = 1.2
		style.LineHeightUnit = "number"
		return
	}

	// Try as number (unitless)
	if v, err := strconv.ParseFloat(value, 64); err == nil {
		style.LineHeight = v
		style.LineHeightUnit = "number"
		return
	}

	// Parse as length
	l := parseLength(value)
	style.LineHeight = l.Resolve(values.DefaultContext())
	style.LineHeightUnit = "px"
}

// parseBackgroundShorthand handles background shorthand
func parseBackgroundShorthand(style *values.ComputedStyle, value string) {
	// Simple case: just a color
	if c, err := values.ParseColor(value); err == nil {
		style.BackgroundColor = c
		return
	}

	// Handle url() or gradient
	if strings.HasPrefix(value, "url(") || strings.Contains(value, "gradient") {
		style.BackgroundImage = value
		return
	}

	// Try to extract color from mixed value
	parts := strings.Fields(value)
	for _, p := range parts {
		if c, err := values.ParseColor(p); err == nil {
			style.BackgroundColor = c
		}
	}
}
