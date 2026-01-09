// Package dom provides HTML Document Object Model structures and parsing
package dom

import (
	"regexp"
	"strings"
)

// NodeType represents the type of DOM node
type NodeType int

const (
	NodeDocument NodeType = iota
	NodeElement
	NodeText
)

// DisplayMode represents CSS display property
type DisplayMode int

const (
	DisplayBlock DisplayMode = iota
	DisplayInline
	DisplayNone
)

// Node represents a node in the DOM tree
type Node struct {
	Type          NodeType
	Tag           string
	Content       string // Only for NodeText
	Children      []*Node
	Parent        *Node
	Display       DisplayMode
	Attributes    map[string]string
	ComputedStyle interface{} // *css.ComputedStyle (interface to avoid circular import)
}

// NewElement creates a new element node
func NewElement(tag string) *Node {
	return &Node{
		Type:       NodeElement,
		Tag:        tag,
		Display:    GetDefaultDisplay(tag),
		Children:   []*Node{},
		Attributes: make(map[string]string),
	}
}

// NewText creates a new text node
func NewText(content string) *Node {
	return &Node{Type: NodeText, Content: content, Display: DisplayInline}
}

// AppendChild adds a child node to this node
func (n *Node) AppendChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// GetAttr returns an attribute value or empty string
func (n *Node) GetAttr(name string) string {
	if n.Attributes == nil {
		return ""
	}
	return n.Attributes[name]
}

// GetDefaultDisplay returns the default display mode for a tag
func GetDefaultDisplay(tag string) DisplayMode {
	switch tag {
	case "div", "p", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "table", "tr", "br", "title",
		"section", "article", "header", "footer", "nav", "main", "ul", "ol", "li", "form",
		"blockquote", "pre", "figure", "figcaption", "aside":
		return DisplayBlock
	case "span", "b", "i", "strong", "em", "a", "td", "th", "img", "label", "small", "sub", "sup":
		return DisplayInline
	case "head", "script", "style", "svg", "path", "meta", "link":
		return DisplayNone
	default:
		return DisplayInline
	}
}

// Attribute parsing regex
var attrRegex = regexp.MustCompile(`(\w+)\s*=\s*["']([^"']*)["']`)

// ParseAttributes extracts attributes from a tag string
func ParseAttributes(tagContent string) map[string]string {
	attrs := make(map[string]string)
	matches := attrRegex.FindAllStringSubmatch(tagContent, -1)
	for _, m := range matches {
		if len(m) >= 3 {
			attrs[strings.ToLower(m[1])] = m[2]
		}
	}
	return attrs
}
