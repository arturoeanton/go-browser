package css

import (
	"strings"
)

// ======================================================================================
// CSS PARSER
// ======================================================================================

// Declaration represents a single CSS property: value pair
type Declaration struct {
	Property  string
	Value     string
	Important bool
}

// Rule represents a CSS rule with selectors and declarations
type Rule struct {
	Selectors    []Selector
	Declarations []Declaration
}

// Stylesheet represents a collection of CSS rules
type Stylesheet struct {
	Rules []Rule
}

// ParseInlineStyle parses a style attribute value like "color: red; font-size: 16px;"
func ParseInlineStyle(styleAttr string) []Declaration {
	var declarations []Declaration

	// Split by semicolon
	parts := strings.Split(styleAttr, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by first colon
		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			continue
		}

		property := strings.TrimSpace(part[:colonIdx])
		value := strings.TrimSpace(part[colonIdx+1:])

		if property == "" || value == "" {
			continue
		}

		// Check for !important
		important := false
		if strings.HasSuffix(strings.ToLower(value), "!important") {
			important = true
			value = strings.TrimSpace(strings.TrimSuffix(value, "!important"))
			value = strings.TrimSpace(strings.TrimSuffix(value, "!IMPORTANT"))
		}

		declarations = append(declarations, Declaration{
			Property:  strings.ToLower(property),
			Value:     value,
			Important: important,
		})
	}

	return declarations
}

// ParseStylesheet parses a CSS stylesheet (e.g., from <style> block)
func ParseStylesheet(css string) *Stylesheet {
	stylesheet := &Stylesheet{}

	// Remove comments
	css = removeComments(css)

	// Parse rules
	pos := 0
	for pos < len(css) {
		// Skip whitespace
		for pos < len(css) && isWhitespace(css[pos]) {
			pos++
		}
		if pos >= len(css) {
			break
		}

		// Find selector (everything before {)
		braceStart := strings.Index(css[pos:], "{")
		if braceStart == -1 {
			break
		}
		braceStart += pos

		selectorText := strings.TrimSpace(css[pos:braceStart])
		if selectorText == "" {
			pos = braceStart + 1
			continue
		}

		// Find closing brace
		braceEnd := findMatchingBrace(css, braceStart)
		if braceEnd == -1 {
			break
		}

		declarationsText := css[braceStart+1 : braceEnd]

		// Parse selectors
		selectors := ParseSelectors(selectorText)

		// Parse declarations
		declarations := ParseInlineStyle(declarationsText)

		if len(selectors) > 0 && len(declarations) > 0 {
			stylesheet.Rules = append(stylesheet.Rules, Rule{
				Selectors:    selectors,
				Declarations: declarations,
			})
		}

		pos = braceEnd + 1
	}

	return stylesheet
}

func removeComments(css string) string {
	result := strings.Builder{}
	i := 0
	for i < len(css) {
		if i+1 < len(css) && css[i] == '/' && css[i+1] == '*' {
			// Skip until */
			end := strings.Index(css[i+2:], "*/")
			if end == -1 {
				break
			}
			i = i + 2 + end + 2
		} else {
			result.WriteByte(css[i])
			i++
		}
	}
	return result.String()
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func findMatchingBrace(css string, start int) int {
	depth := 1
	for i := start + 1; i < len(css); i++ {
		if css[i] == '{' {
			depth++
		} else if css[i] == '}' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// ApplyDeclarations applies CSS declarations to a ComputedStyle
func ApplyDeclarations(style *ComputedStyle, declarations []Declaration) {
	for _, decl := range declarations {
		ApplyProperty(style, decl.Property, decl.Value)
	}
}

// ApplyProperty applies a single CSS property to a ComputedStyle
func ApplyProperty(style *ComputedStyle, property, value string) {
	value = strings.TrimSpace(value)

	switch property {
	// Display
	case "display":
		style.Display = value
	case "visibility":
		style.Visibility = value

	// Colors
	case "color":
		if c, ok := ParseColor(value); ok {
			style.Color = c
		}
	case "background-color", "background":
		// Check for gradient first
		if strings.Contains(value, "gradient") {
			if g, ok := ParseGradient(value); ok {
				style.BackgroundGradient = g
			}
		} else if c, ok := ParseColor(value); ok {
			style.BackgroundColor = c
		}

	// Typography
	case "font-size":
		if l, unit, ok := ParseLength(value); ok {
			if unit == UnitPx {
				style.FontSize = l
			} else if unit == UnitEm || unit == UnitRem {
				style.FontSize = l * 16 // base font size
			}
		}
	case "font-weight":
		switch value {
		case "normal":
			style.FontWeight = 400
		case "bold":
			style.FontWeight = 700
		case "lighter":
			style.FontWeight = 300
		case "bolder":
			style.FontWeight = 800
		default:
			if w, _, ok := ParseLength(value); ok {
				style.FontWeight = int(w)
			}
		}
	case "font-family":
		style.FontFamily = value
	case "text-align":
		style.TextAlign = value
	case "line-height":
		if l, unit, ok := ParseLength(value); ok {
			if unit == UnitPx {
				style.LineHeight = l / style.FontSize
			} else {
				style.LineHeight = l
			}
		}

	// Box Model - Width/Height
	case "width":
		if l, unit, ok := ParseLength(value); ok && unit == UnitPx {
			style.Width = l
		}
	case "height":
		if l, unit, ok := ParseLength(value); ok && unit == UnitPx {
			style.Height = l
		}

	// Margins
	case "margin":
		applyBoxShorthand(value, func(top, right, bottom, left float64) {
			style.MarginTop = top
			style.MarginRight = right
			style.MarginBottom = bottom
			style.MarginLeft = left
		})
	case "margin-top":
		if l, _, ok := ParseLength(value); ok {
			style.MarginTop = l
		}
	case "margin-right":
		if l, _, ok := ParseLength(value); ok {
			style.MarginRight = l
		}
	case "margin-bottom":
		if l, _, ok := ParseLength(value); ok {
			style.MarginBottom = l
		}
	case "margin-left":
		if l, _, ok := ParseLength(value); ok {
			style.MarginLeft = l
		}

	// Padding
	case "padding":
		applyBoxShorthand(value, func(top, right, bottom, left float64) {
			style.PaddingTop = top
			style.PaddingRight = right
			style.PaddingBottom = bottom
			style.PaddingLeft = left
		})
	case "padding-top":
		if l, _, ok := ParseLength(value); ok {
			style.PaddingTop = l
		}
	case "padding-right":
		if l, _, ok := ParseLength(value); ok {
			style.PaddingRight = l
		}
	case "padding-bottom":
		if l, _, ok := ParseLength(value); ok {
			style.PaddingBottom = l
		}
	case "padding-left":
		if l, _, ok := ParseLength(value); ok {
			style.PaddingLeft = l
		}

	// Border
	case "border-radius":
		if l, _, ok := ParseLength(value); ok {
			style.BorderRadius = l
		}
	case "border-width":
		if l, _, ok := ParseLength(value); ok {
			style.BorderTopWidth = l
			style.BorderRightWidth = l
			style.BorderBottomWidth = l
			style.BorderLeftWidth = l
		}
	case "border-color":
		if c, ok := ParseColor(value); ok {
			style.BorderColor = c
		}

	// Position
	case "position":
		style.Position = value
	case "top":
		if l, _, ok := ParseLength(value); ok {
			style.Top = l
		}
	case "right":
		if l, _, ok := ParseLength(value); ok {
			style.Right = l
		}
	case "bottom":
		if l, _, ok := ParseLength(value); ok {
			style.Bottom = l
		}
	case "left":
		if l, _, ok := ParseLength(value); ok {
			style.Left = l
		}
	case "z-index":
		if l, _, ok := ParseLength(value); ok {
			style.ZIndex = int(l)
		}

	// Flexbox properties
	case "flex-direction":
		style.FlexDirection = value
	case "justify-content":
		style.JustifyContent = value
	case "align-items":
		style.AlignItems = value
	case "align-content":
		style.AlignContent = value
	case "align-self":
		style.AlignSelf = value
	case "flex-wrap":
		style.FlexWrap = value
	case "gap":
		if l, _, ok := ParseLength(value); ok {
			style.Gap = l
			style.RowGap = l
			style.ColumnGap = l
		}
	case "row-gap":
		if l, _, ok := ParseLength(value); ok {
			style.RowGap = l
		}
	case "column-gap":
		if l, _, ok := ParseLength(value); ok {
			style.ColumnGap = l
		}
	case "flex-grow":
		if l, _, ok := ParseLength(value); ok {
			style.FlexGrow = l
		}
	case "flex-shrink":
		if l, _, ok := ParseLength(value); ok {
			style.FlexShrink = l
		}
	case "flex-basis":
		if l, _, ok := ParseLength(value); ok {
			style.FlexBasis = l
		}
	case "order":
		if l, _, ok := ParseLength(value); ok {
			style.Order = int(l)
		}
	case "flex":
		// Shorthand: flex: grow shrink basis OR flex: grow
		parts := strings.Fields(value)
		if len(parts) >= 1 {
			if l, _, ok := ParseLength(parts[0]); ok {
				style.FlexGrow = l
			}
		}
		if len(parts) >= 2 {
			if l, _, ok := ParseLength(parts[1]); ok {
				style.FlexShrink = l
			}
		}
		if len(parts) >= 3 {
			if l, _, ok := ParseLength(parts[2]); ok {
				style.FlexBasis = l
			}
		}

	// CSS Grid properties
	case "grid-template-columns":
		style.GridTemplateColumns = value
		// Count columns from the value
		style.GridColumnCount = countGridColumns(value)
	case "grid-template-rows":
		style.GridTemplateRows = value
	case "grid-column":
		style.GridColumn = value
	case "grid-row":
		style.GridRow = value
	}
}

// countGridColumns counts the number of columns from a grid-template-columns value
func countGridColumns(value string) int {
	// Handle repeat(n, ...) syntax
	if strings.HasPrefix(value, "repeat(") {
		// Extract the number from repeat(n, ...)
		inner := strings.TrimPrefix(value, "repeat(")
		if idx := strings.Index(inner, ","); idx > 0 {
			numStr := strings.TrimSpace(inner[:idx])
			if n, _, ok := ParseLength(numStr); ok {
				return int(n)
			}
		}
		return 3 // default for repeat
	}

	// Count space-separated values
	parts := strings.Fields(value)
	return len(parts)
}

// applyBoxShorthand handles margin/padding shorthand (1, 2, 3, or 4 values)
func applyBoxShorthand(value string, apply func(top, right, bottom, left float64)) {
	parts := strings.Fields(value)
	values := make([]float64, 0, 4)

	for _, p := range parts {
		if l, _, ok := ParseLength(p); ok {
			values = append(values, l)
		}
	}

	switch len(values) {
	case 1:
		apply(values[0], values[0], values[0], values[0])
	case 2:
		apply(values[0], values[1], values[0], values[1])
	case 3:
		apply(values[0], values[1], values[2], values[1])
	case 4:
		apply(values[0], values[1], values[2], values[3])
	}
}
