package dom

import (
	"html"
	"strings"
)

// ======================================================================================
// HTML ENTITIES
// ======================================================================================

// DecodeEntities decodes HTML entities in a string
func DecodeEntities(s string) string {
	return html.UnescapeString(s)
}

// EncodeEntities encodes special characters as HTML entities
func EncodeEntities(s string) string {
	return html.EscapeString(s)
}

// ======================================================================================
// NODE MANIPULATION
// ======================================================================================

// RemoveChild removes a child node
func (n *Node) RemoveChild(child *Node) bool {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.Parent = nil
			return true
		}
	}
	return false
}

// InsertBefore inserts a new child before the reference child
func (n *Node) InsertBefore(newChild, refChild *Node) bool {
	for i, c := range n.Children {
		if c == refChild {
			newChildren := make([]*Node, 0, len(n.Children)+1)
			newChildren = append(newChildren, n.Children[:i]...)
			newChildren = append(newChildren, newChild)
			newChildren = append(newChildren, n.Children[i:]...)
			n.Children = newChildren
			newChild.Parent = n
			return true
		}
	}
	return false
}

// ReplaceChild replaces an old child with a new one
func (n *Node) ReplaceChild(newChild, oldChild *Node) bool {
	for i, c := range n.Children {
		if c == oldChild {
			n.Children[i] = newChild
			newChild.Parent = n
			oldChild.Parent = nil
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the node
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}

	clone := &Node{
		Type:       n.Type,
		Tag:        n.Tag,
		Content:    n.Content,
		Display:    n.Display,
		Attributes: make(map[string]string),
	}

	// Copy attributes
	for k, v := range n.Attributes {
		clone.Attributes[k] = v
	}

	// Clone children
	for _, child := range n.Children {
		childClone := child.Clone()
		clone.AppendChild(childClone)
	}

	return clone
}

// ======================================================================================
// SERIALIZATION
// ======================================================================================

// OuterHTML returns the HTML representation of the node
func (n *Node) OuterHTML() string {
	if n == nil {
		return ""
	}

	if n.Type == NodeText {
		return EncodeEntities(n.Content)
	}

	var sb strings.Builder

	// Opening tag
	sb.WriteString("<")
	sb.WriteString(n.Tag)

	// Attributes
	for k, v := range n.Attributes {
		sb.WriteString(" ")
		sb.WriteString(k)
		sb.WriteString("=\"")
		sb.WriteString(EncodeEntities(v))
		sb.WriteString("\"")
	}

	// Check for void elements
	if isVoidElement(n.Tag) {
		sb.WriteString(" />")
		return sb.String()
	}

	sb.WriteString(">")

	// Children
	for _, child := range n.Children {
		sb.WriteString(child.OuterHTML())
	}

	// Closing tag
	sb.WriteString("</")
	sb.WriteString(n.Tag)
	sb.WriteString(">")

	return sb.String()
}

// InnerHTML returns the HTML of the node's children
func (n *Node) InnerHTML() string {
	if n == nil {
		return ""
	}

	var sb strings.Builder
	for _, child := range n.Children {
		sb.WriteString(child.OuterHTML())
	}
	return sb.String()
}

// isVoidElement returns true for HTML void elements
func isVoidElement(tag string) bool {
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true,
		"embed": true, "hr": true, "img": true, "input": true,
		"link": true, "meta": true, "source": true, "track": true,
		"wbr": true,
	}
	return voidElements[strings.ToLower(tag)]
}

// ======================================================================================
// DEBUGGING
// ======================================================================================

// DebugString returns a debug representation of the node tree
func (n *Node) DebugString() string {
	return n.debugStringIndent(0)
}

func (n *Node) debugStringIndent(indent int) string {
	if n == nil {
		return ""
	}

	prefix := strings.Repeat("  ", indent)
	var sb strings.Builder

	switch n.Type {
	case NodeElement:
		sb.WriteString(prefix)
		sb.WriteString("<")
		sb.WriteString(n.Tag)

		if id := n.GetAttr("id"); id != "" {
			sb.WriteString(" id=\"")
			sb.WriteString(id)
			sb.WriteString("\"")
		}
		if class := n.GetAttr("class"); class != "" {
			sb.WriteString(" class=\"")
			sb.WriteString(class)
			sb.WriteString("\"")
		}
		sb.WriteString(">\n")

		for _, child := range n.Children {
			sb.WriteString(child.debugStringIndent(indent + 1))
		}

	case NodeText:
		text := strings.TrimSpace(n.Content)
		if len(text) > 50 {
			text = text[:50] + "..."
		}
		if text != "" {
			sb.WriteString(prefix)
			sb.WriteString("\"")
			sb.WriteString(text)
			sb.WriteString("\"\n")
		}
	}

	return sb.String()
}
