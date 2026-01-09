// Package layout provides the layout engine for Gocko
package layout

import (
	"strings"

	"go-browser/css"
	"go-browser/dom"
	"go-browser/gocko/box"
	"go-browser/gocko/forms"
)

// Constants
const (
	FontSizeDefault = 15.0
	LineHeight      = 1.4
)

// LayoutContext holds the current layout state
type LayoutContext struct {
	CursorX, CursorY float64
	MaxWidth         float64
	LineHeight       float64
	RowCounter       int
}

// BuildLayoutTree creates a layout tree from DOM nodes
func BuildLayoutTree(root *dom.Node, width float64, styles []*css.Stylesheet) *box.Box {
	ctx := &LayoutContext{
		CursorX:    0,
		CursorY:    0,
		MaxWidth:   width,
		LineHeight: FontSizeDefault * LineHeight,
	}

	rootBox := &box.Box{
		Node:    root,
		Width:   width,
		Display: "block",
	}

	layoutChildren(root, rootBox, ctx)
	rootBox.Height = ctx.CursorY + ctx.LineHeight

	return rootBox
}

func layoutChildren(node *dom.Node, container *box.Box, ctx *LayoutContext) {
	for _, child := range node.Children {
		childBox := layoutNode(child, ctx)
		if childBox != nil {
			container.Children = append(container.Children, childBox)
		}
	}
}

func layoutNode(node *dom.Node, ctx *LayoutContext) *box.Box {
	// Skip invisible elements
	if node.Display == dom.DisplayNone {
		return nil
	}

	// Check computed style for display:none
	if node.ComputedStyle != nil {
		if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
			if cs.Display == "none" {
				return nil
			}
		}
	}

	switch node.Type {
	case dom.NodeText:
		return layoutText(node, ctx)
	case dom.NodeElement:
		return layoutElement(node, ctx)
	}

	return nil
}

func layoutText(node *dom.Node, ctx *LayoutContext) *box.Box {
	text := strings.TrimSpace(node.Content)
	if text == "" {
		return nil
	}

	fontSize := FontSizeDefault
	var isBold bool
	var isLink bool
	var linkURL string

	// Get styles from parent
	if node.Parent != nil {
		if node.Parent.ComputedStyle != nil {
			if cs, ok := node.Parent.ComputedStyle.(*css.ComputedStyle); ok {
				if cs.FontSize > 0 {
					fontSize = cs.FontSize
				}
				if cs.FontWeight >= 600 {
					isBold = true
				}
			}
		}

		// Check for link
		if node.Parent.Tag == "a" {
			isLink = true
			linkURL = node.Parent.GetAttr("href")
		}
	}

	lineH := fontSize * LineHeight
	charW := fontSize * 0.55

	// Word wrap
	words := strings.Fields(text)
	var lines []string
	currentLine := ""
	currentWidth := ctx.CursorX

	for _, word := range words {
		wordWidth := float64(len(word)+1) * charW
		if currentWidth+wordWidth > ctx.MaxWidth && currentLine != "" {
			lines = append(lines, strings.TrimSpace(currentLine))
			currentLine = word + " "
			currentWidth = wordWidth
		} else {
			currentLine += word + " "
			currentWidth += wordWidth
		}
	}
	if currentLine != "" {
		lines = append(lines, strings.TrimSpace(currentLine))
	}

	// Create box for first line
	if len(lines) == 0 {
		return nil
	}

	textBox := &box.Box{
		Node:     node,
		Text:     lines[0],
		X:        ctx.CursorX,
		Y:        ctx.CursorY,
		Width:    float64(len(lines[0])) * charW,
		Height:   lineH,
		FontSize: fontSize,
		IsBold:   isBold,
		IsLink:   isLink,
		LinkURL:  linkURL,
		Display:  "inline",
	}

	ctx.CursorX += textBox.Width

	// Handle overflow lines
	for i := 1; i < len(lines); i++ {
		ctx.CursorX = 0
		ctx.CursorY += lineH
		childBox := &box.Box{
			Node:     node,
			Text:     lines[i],
			X:        0,
			Y:        ctx.CursorY,
			Width:    float64(len(lines[i])) * charW,
			Height:   lineH,
			FontSize: fontSize,
			IsBold:   isBold,
			IsLink:   isLink,
			LinkURL:  linkURL,
			Display:  "inline",
		}
		textBox.Children = append(textBox.Children, childBox)
		ctx.CursorX = childBox.Width
	}

	// Move to next line for block parents
	if node.Parent != nil {
		tag := node.Parent.Tag
		if tag == "p" || tag == "div" || tag == "h1" || tag == "h2" || tag == "h3" || tag == "li" {
			ctx.CursorY += lineH
			ctx.CursorX = 0
		}
	}

	return textBox
}

func layoutElement(node *dom.Node, ctx *LayoutContext) *box.Box {
	tag := strings.ToLower(node.Tag)

	// Get element ID
	id := node.GetAttr("id")
	if id == "" {
		id = node.GetAttr("name")
	}
	if id == "" {
		id = tag
	}

	elemBox := &box.Box{
		Node:    node,
		ID:      id,
		X:       ctx.CursorX,
		Y:       ctx.CursorY,
		Display: getDisplayType(node),
	}

	// Apply computed styles
	applyStyles(elemBox, node)

	// Check if this is a form element
	if formComp := createFormComponent(node, id); formComp != nil {
		elemBox.FormComponent = formComp
		elemBox.Width = 200 // Default form element width
		elemBox.Height = 32 // Default form element height
		ctx.CursorY += elemBox.Height + 8
		ctx.CursorX = 0
		return elemBox
	}

	// Handle special elements
	switch tag {
	case "br":
		ctx.CursorX = 0
		ctx.CursorY += ctx.LineHeight
		return nil
	case "hr":
		elemBox.X = 0
		elemBox.Y = ctx.CursorY + 8
		elemBox.Width = ctx.MaxWidth
		elemBox.Height = 2
		ctx.CursorY += 20
		ctx.CursorX = 0
		return elemBox
	case "img":
		src := node.GetAttr("src")
		if src != "" {
			elemBox.IsImage = true
			elemBox.ImageURL = src
			elemBox.Width = 200
			elemBox.Height = 150
			ctx.CursorY += elemBox.Height + 10
			ctx.CursorX = 0
			return elemBox
		}
		return nil
	}

	// Block elements start on new line
	if isBlockElement(tag) {
		if ctx.CursorX > 0 {
			ctx.CursorX = 0
			ctx.CursorY += ctx.LineHeight
		}
		// Apply spacing
		ctx.CursorY += getElementSpacing(tag)
		elemBox.X = 0
		elemBox.Y = ctx.CursorY
	}

	// Layout children
	startY := ctx.CursorY
	layoutChildren(node, elemBox, ctx)
	elemBox.Width = ctx.MaxWidth
	elemBox.Height = ctx.CursorY - startY

	// Block element spacing after
	if isBlockElement(tag) {
		ctx.CursorY += getElementSpacing(tag) / 2
		ctx.CursorX = 0
	}

	return elemBox
}

func createFormComponent(node *dom.Node, id string) forms.FormComponent {
	// Form components are handled by forms.GetHandler() in browser/app.go
	// This layout engine doesn't need to create components directly
	return nil
}

func applyStyles(b *box.Box, node *dom.Node) {
	if node.ComputedStyle == nil {
		return
	}
	cs, ok := node.ComputedStyle.(*css.ComputedStyle)
	if !ok {
		return
	}

	b.MarginTop = cs.MarginTop
	b.MarginRight = cs.MarginRight
	b.MarginBottom = cs.MarginBottom
	b.MarginLeft = cs.MarginLeft

	b.PaddingTop = cs.PaddingTop
	b.PaddingRight = cs.PaddingRight
	b.PaddingBottom = cs.PaddingBottom
	b.PaddingLeft = cs.PaddingLeft

	if cs.Position != "" {
		b.Position = cs.Position
	}
	if cs.Display != "" {
		b.Display = cs.Display
	}
}

func getDisplayType(node *dom.Node) string {
	if node.ComputedStyle != nil {
		if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
			if cs.Display != "" {
				return cs.Display
			}
		}
	}
	if isBlockElement(node.Tag) {
		return "block"
	}
	return "inline"
}

func isBlockElement(tag string) bool {
	blocks := map[string]bool{
		"div": true, "p": true, "h1": true, "h2": true, "h3": true,
		"h4": true, "h5": true, "h6": true, "section": true, "article": true,
		"header": true, "footer": true, "nav": true, "main": true, "aside": true,
		"ul": true, "ol": true, "li": true, "form": true, "fieldset": true,
		"table": true, "tr": true, "pre": true, "blockquote": true,
	}
	return blocks[tag]
}

func getElementSpacing(tag string) float64 {
	spacing := map[string]float64{
		"p": 16, "div": 8, "h1": 24, "h2": 20, "h3": 18,
		"section": 16, "article": 16, "ul": 12, "ol": 12, "li": 8,
		"form": 16, "table": 16, "hr": 16, "fieldset": 16,
	}
	if s, ok := spacing[tag]; ok {
		return s
	}
	return 0
}
