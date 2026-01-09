// Package layout provides the layout engine for positioning elements
package layout

import (
	"image/color"
	"strings"

	"go-browser/css"
	"go-browser/dom"

	"github.com/hajimehoshi/ebiten/v2"
)

// Constants for layout
const (
	FontSizeBody = 15
	FontSizeH1   = 28
	FontSizeH2   = 22
)

// RenderBox represents a positioned element for rendering
type RenderBox struct {
	Node     *dom.Node
	X, Y     float64
	W, H     float64
	Children []*RenderBox
	Text     string
	FontSize float64
	IsH1     bool
	IsH2     bool
	IsBold   bool
	IsLink   bool
	IsButton bool
	LinkURL  string
	RowIndex int // For table striping
	// Image support
	IsImage  bool
	ImageURL string
	// CSS computed colors
	TextColor *color.RGBA
	BgColor   *color.RGBA
	// Text alignment
	TextAlign string // left, center, right
	// Positioning
	Position string // static, relative, absolute, fixed
	IsFixed  bool   // true if position: fixed
}

// Default spacing for block elements (margin in pixels)
var ElementSpacing = map[string]float64{
	"p":          16,
	"div":        8,
	"h1":         24,
	"h2":         20,
	"h3":         18,
	"h4":         16,
	"h5":         14,
	"h6":         12,
	"section":    16,
	"article":    16,
	"header":     16,
	"footer":     16,
	"nav":        12,
	"main":       16,
	"aside":      16,
	"ul":         16,
	"ol":         16,
	"li":         8,
	"blockquote": 20,
	"pre":        16,
	"form":       16,
	"table":      16,
	"tr":         4,
	"figure":     16,
	"figcaption": 8,
	"hr":         16,
}

// Inline elements that flow horizontally
var InlineElements = map[string]bool{
	"a": true, "abbr": true, "b": true, "bdo": true, "br": true,
	"cite": true, "code": true, "dfn": true, "em": true, "i": true,
	"kbd": true, "label": true, "q": true, "samp": true, "small": true,
	"span": true, "strong": true, "sub": true, "sup": true,
	"time": true, "var": true, "mark": true,
}

// LayoutContext holds the current layout state
type LayoutContext struct {
	CursorX, CursorY float64
	MaxW             float64
	LineHeight       float64
	RowCounter       int
}

// BuildRenderTree creates a render tree from DOM nodes
func BuildRenderTree(node *dom.Node, width float64) *RenderBox {
	box := &RenderBox{Node: node, W: width}
	ctx := &LayoutContext{CursorX: 0, CursorY: 0, MaxW: width, LineHeight: 24}
	layoutRecursive(node, box, ctx)
	box.H = ctx.CursorY + ctx.LineHeight
	return box
}

func layoutRecursive(node *dom.Node, container *RenderBox, ctx *LayoutContext) {
	if node.Tag == "title" && len(node.Children) > 0 && node.Children[0].Type == dom.NodeText {
		ebiten.SetWindowTitle("GoBrowser: " + node.Children[0].Content)
		return
	}
	if node.Display == dom.DisplayNone {
		return
	}

	// Skip elements that are handled by their parent (like option inside select)
	switch node.Tag {
	case "option", "optgroup":
		return // These are rendered by the SelectHandler
	}

	// Check if computed style has display:none
	if node.ComputedStyle != nil {
		if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
			if cs.Display == "none" {
				return
			}
		}
	}

	// Apply margin-top from CSS for block elements
	marginTop := 0.0
	marginBottom := 0.0
	paddingLeft := 0.0
	paddingTop := 0.0
	position := "static"

	if node.ComputedStyle != nil {
		if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
			marginTop = cs.MarginTop
			marginBottom = cs.MarginBottom
			paddingLeft = cs.PaddingLeft
			paddingTop = cs.PaddingTop
			if cs.Position != "" {
				position = cs.Position
			}
		}
	}

	// Store position in container for later use by render
	container.Position = position
	container.IsFixed = position == "fixed"

	// Apply margin-top
	if marginTop > 0 {
		ctx.CursorY += marginTop
	}

	// Determine if this is a block or inline element
	isInline := InlineElements[node.Tag]
	defaultSpacing := ElementSpacing[node.Tag]
	isBlockElement := !isInline && defaultSpacing > 0

	// Also check for display from CSS
	isInlineBlock := false
	if node.ComputedStyle != nil {
		if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
			if cs.Display == "inline-block" {
				isInlineBlock = true
				isBlockElement = false
				isInline = true
			} else if cs.Display == "block" {
				isBlockElement = true
				isInline = false
			} else if cs.Display == "inline" {
				isInline = true
				isBlockElement = false
			}
		}
	}
	_ = isInlineBlock // Will be used for inline-block specific layout

	// Block elements always start on new line with proper spacing
	if isBlockElement {
		// Always reset X to 0 for block elements
		ctx.CursorX = 0
		// Add default spacing if no CSS margin
		if marginTop == 0 && defaultSpacing > 0 {
			ctx.CursorY += defaultSpacing
		}
	}

	// Apply padding to starting position
	ctx.CursorX += paddingLeft
	ctx.CursorY += paddingTop

	// Apply width/max-width constraints
	originalMaxW := ctx.MaxW
	if node.ComputedStyle != nil {
		if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
			// Apply max-width if set
			if cs.MaxWidth > 0 && cs.MaxWidth < ctx.MaxW {
				ctx.MaxW = cs.MaxWidth
			}
			// Apply explicit width if set
			if cs.Width > 0 {
				ctx.MaxW = cs.Width
			}
		}
	}
	_ = originalMaxW // Will be used to restore later if needed

	// Track row for table striping
	if node.Tag == "tr" {
		ctx.RowCounter++
	}

	if node.Tag == "td" {
		ctx.MaxW = 180
	}

	startX := ctx.CursorX
	_ = ctx.CursorY

	// Render text nodes
	if node.Type == dom.NodeText {
		fontSize := float64(FontSizeBody)
		lineH := 24.0
		isH1 := false
		isH2 := false
		isBold := false
		isLink := false
		isButton := false
		linkURL := ""
		textAlign := "left"
		var textColor *color.RGBA
		var bgColor *color.RGBA

		if node.Parent != nil {
			// Try to get computed styles from parent
			if node.Parent.ComputedStyle != nil {
				if cs, ok := node.Parent.ComputedStyle.(*css.ComputedStyle); ok {
					if cs.FontSize > 0 {
						fontSize = cs.FontSize
						lineH = fontSize * 1.4
					}
					if cs.FontWeight >= 600 {
						isBold = true
					}
					// Extract color
					if cs.Color.A > 0 {
						textColor = &cs.Color
					}
					if cs.BackgroundColor.A > 0 {
						bgColor = &cs.BackgroundColor
					}
					// Extract text-align
					if cs.TextAlign != "" {
						textAlign = cs.TextAlign
					}
				}
			}

			switch node.Parent.Tag {
			case "h1":
				if fontSize == FontSizeBody {
					fontSize = FontSizeH1
					lineH = 36
				}
				isH1 = true
			case "h2":
				if fontSize == FontSizeBody {
					fontSize = FontSizeH2
					lineH = 30
				}
				isH2 = true
			case "b", "strong":
				isBold = true
			case "a":
				isLink = true
				linkURL = node.Parent.GetAttr("href")
				// Check if it looks like a button
				class := node.Parent.GetAttr("class")
				if strings.Contains(class, "btn") || strings.Contains(class, "button") ||
					strings.Contains(class, "cta") || strings.Contains(class, "nav-cta") {
					isButton = true
				}
			case "button":
				isButton = true
			}
		}

		words := strings.Fields(node.Content)
		line := ""
		charW := fontSize * 0.55

		for _, w := range words {
			wLen := float64(len(w)+1) * charW
			if ctx.CursorX+wLen > ctx.MaxW {
				childBox := &RenderBox{
					Text: line, X: startX, Y: ctx.CursorY,
					W: ctx.CursorX - startX, H: lineH,
					FontSize: fontSize, IsH1: isH1, IsH2: isH2, IsBold: isBold,
					IsLink: isLink, IsButton: isButton, LinkURL: linkURL,
					TextColor: textColor, BgColor: bgColor, TextAlign: textAlign,
				}
				container.Children = append(container.Children, childBox)

				ctx.CursorX = 0
				ctx.CursorY += lineH
				startX = 0
				line = w + " "
				ctx.CursorX = wLen
			} else {
				line += w + " "
				ctx.CursorX += wLen
			}
		}

		if len(line) > 0 {
			childBox := &RenderBox{
				Text: line, X: startX, Y: ctx.CursorY,
				W: ctx.CursorX - startX, H: lineH,
				FontSize: fontSize, IsH1: isH1, IsH2: isH2, IsBold: isBold,
				IsLink: isLink, IsButton: isButton, LinkURL: linkURL,
				TextColor: textColor, BgColor: bgColor, TextAlign: textAlign,
			}
			container.Children = append(container.Children, childBox)

			// After text, if parent is block element, move to next line
			if node.Parent != nil {
				parentTag := node.Parent.Tag
				isParentBlock := parentTag == "p" || parentTag == "div" || parentTag == "h1" ||
					parentTag == "h2" || parentTag == "h3" || parentTag == "li" ||
					parentTag == "section" || parentTag == "article"
				if isParentBlock {
					ctx.CursorY += lineH
					ctx.CursorX = 0
				}
			}
		}
	} else if node.Tag == "hr" {
		ctx.CursorY += 12
		childBox := &RenderBox{Node: node, X: 0, Y: ctx.CursorY, W: ctx.MaxW, H: 2}
		container.Children = append(container.Children, childBox)
		ctx.CursorY += 16
		ctx.CursorX = 0
	} else if node.Tag == "br" {
		ctx.CursorX = 0
		ctx.CursorY += ctx.LineHeight
	} else if node.Tag == "img" {
		// Handle image tags
		src := node.GetAttr("src")
		if src != "" {
			imgW := 200.0 // Default width
			imgH := 150.0 // Default height

			// New line for images
			if ctx.CursorX > 0 {
				ctx.CursorX = 0
				ctx.CursorY += ctx.LineHeight
			}

			childBox := &RenderBox{
				Node:     node,
				X:        ctx.CursorX,
				Y:        ctx.CursorY,
				W:        imgW,
				H:        imgH,
				IsImage:  true,
				ImageURL: src,
			}
			container.Children = append(container.Children, childBox)
			ctx.CursorY += imgH + 10
		}
	} else if node.Tag == "input" || node.Tag == "select" || node.Tag == "textarea" {
		// Handle form input elements - give them proper size and spacing
		inputType := node.GetAttr("type")
		if inputType == "" {
			inputType = "text"
		}

		// Determine size based on type
		var inputW, inputH float64
		switch inputType {
		case "checkbox", "radio":
			inputW = 18
			inputH = 18
		case "submit", "button":
			inputW = 100
			inputH = 32
		default:
			inputW = 200
			inputH = 30
		}

		if node.Tag == "textarea" {
			inputW = 300
			inputH = 80
		}
		if node.Tag == "select" {
			inputW = 200
			inputH = 32
		}

		// For non-checkbox/radio, start on new line if there's content
		if inputType != "checkbox" && inputType != "radio" {
			if ctx.CursorX > 0 {
				ctx.CursorX = 0
				ctx.CursorY += ctx.LineHeight
			}
		}

		childBox := &RenderBox{
			Node: node,
			X:    ctx.CursorX,
			Y:    ctx.CursorY,
			W:    inputW,
			H:    inputH,
		}
		container.Children = append(container.Children, childBox)

		// Move cursor based on input type
		if inputType == "checkbox" || inputType == "radio" {
			// Checkboxes and radios stay inline, just move X
			ctx.CursorX += inputW + 8
		} else {
			// Other inputs are block, move to next line
			ctx.CursorY += inputH + 12
			ctx.CursorX = 0
		}
	} else if node.Tag == "button" {
		// Handle button elements - always start on new line with extra spacing
		if ctx.CursorX > 0 {
			ctx.CursorX = 0
			ctx.CursorY += ctx.LineHeight + 10 // Extra space after inline elements
		} else {
			// Even if starting from X=0, add spacing to separate from content above
			ctx.CursorY += 15
		}

		childBox := &RenderBox{
			Node: node,
			X:    ctx.CursorX,
			Y:    ctx.CursorY,
			W:    100,
			H:    36,
		}
		container.Children = append(container.Children, childBox)
		ctx.CursorY += 48
		ctx.CursorX = 0
	} else {
		// Check if this is a flex or grid container
		isFlex := false
		isGrid := false
		var flexGap float64 = 0
		flexDirection := "row"
		gridColumns := 0

		if node.ComputedStyle != nil {
			if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
				if cs.Display == "flex" || cs.Display == "inline-flex" {
					isFlex = true
					flexGap = cs.Gap
					if cs.FlexDirection != "" {
						flexDirection = cs.FlexDirection
					}
				} else if cs.Display == "grid" {
					isGrid = true
					flexGap = cs.Gap
					gridColumns = cs.GridColumnCount
					if gridColumns == 0 {
						gridColumns = 1 // Default to 1 column
					}
				}
			}
		}

		if isGrid {
			// CSS Grid layout
			startY := ctx.CursorY
			colWidth := (ctx.MaxW - flexGap*float64(gridColumns-1)) / float64(gridColumns)
			currentCol := 0
			currentRowY := startY
			maxRowH := 0.0

			for _, child := range node.Children {
				childBox := &RenderBox{Node: child}

				// Calculate position in grid
				childX := float64(currentCol) * (colWidth + flexGap)

				// Create a temporary context for child layout
				childCtx := &LayoutContext{
					CursorX:    0,
					CursorY:    0,
					MaxW:       colWidth,
					LineHeight: ctx.LineHeight,
				}

				layoutRecursive(child, childBox, childCtx)

				childBox.X = childX
				childBox.Y = currentRowY
				childBox.W = colWidth
				childBox.H = childCtx.CursorY + childCtx.LineHeight

				if childBox.H > maxRowH {
					maxRowH = childBox.H
				}

				container.Children = append(container.Children, childBox)

				// Move to next column or row
				currentCol++
				if currentCol >= gridColumns {
					currentCol = 0
					currentRowY += maxRowH + flexGap
					maxRowH = 0
				}
			}

			// Update cursor Y
			if currentCol > 0 {
				ctx.CursorY = currentRowY + maxRowH
			} else {
				ctx.CursorY = currentRowY
			}
		} else if isFlex && flexDirection == "row" {
			// Horizontal flex layout
			childX := ctx.CursorX
			startY := ctx.CursorY
			maxChildH := 0.0

			for i, child := range node.Children {
				childBox := &RenderBox{Node: child}

				// Create a temporary context for child layout
				childCtx := &LayoutContext{
					CursorX:    0,
					CursorY:    0,
					MaxW:       ctx.MaxW / float64(len(node.Children)), // Distribute width
					LineHeight: ctx.LineHeight,
				}

				layoutRecursive(child, childBox, childCtx)

				childBox.X = childX
				childBox.Y = startY
				childBox.W = childCtx.CursorX
				if childBox.W < 50 {
					childBox.W = 50 // Minimum width
				}
				childBox.H = childCtx.CursorY + childCtx.LineHeight
				if childBox.H > maxChildH {
					maxChildH = childBox.H
				}

				container.Children = append(container.Children, childBox)

				childX += childBox.W
				if i < len(node.Children)-1 {
					childX += flexGap
				}
			}

			ctx.CursorY = startY + maxChildH
		} else if isFlex && (flexDirection == "column" || flexDirection == "column-reverse") {
			// Vertical flex layout (similar to normal flow but with gap)
			for i, child := range node.Children {
				childBox := &RenderBox{Node: child}
				childYStart := ctx.CursorY

				layoutRecursive(child, childBox, ctx)

				childBox.X = 0
				childBox.Y = childYStart
				childBox.W = ctx.MaxW
				childBox.H = ctx.CursorY - childYStart
				container.Children = append(container.Children, childBox)

				if i < len(node.Children)-1 {
					ctx.CursorY += flexGap
				}
			}
		} else {
			// Normal block flow layout
			for _, child := range node.Children {
				childBox := &RenderBox{Node: child}
				childYStart := ctx.CursorY

				layoutRecursive(child, childBox, ctx)

				childBox.X = 0
				childBox.Y = childYStart
				childBox.W = ctx.MaxW
				childBox.H = ctx.CursorY - childYStart
				childBox.RowIndex = ctx.RowCounter

				if node.Tag == "tr" && child.Tag == "td" {
					ctx.CursorY = childYStart
					ctx.CursorX += 190
					childBox.X = ctx.CursorX - 190
				} else {
					container.Children = append(container.Children, childBox)
				}
			}

			if node.Tag == "tr" {
				ctx.CursorX = 0
				ctx.CursorY += ctx.LineHeight * 1.6
			}
		}
	}

	// Post-margins - apply margin-bottom from CSS or fallback defaults
	if marginBottom > 0 {
		ctx.CursorY += marginBottom
	} else {
		// Fallback margins for common elements
		if node.Tag == "p" {
			ctx.CursorY += 12
		}
		if node.Tag == "h1" {
			ctx.CursorY += 16
		}
		if node.Tag == "h2" {
			ctx.CursorY += 12
		}
	}
}
