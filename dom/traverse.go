package dom

import "strings"

// ======================================================================================
// DOM TREE TRAVERSAL
// ======================================================================================

// GetElementById finds an element by ID attribute
func (n *Node) GetElementById(id string) *Node {
	if n == nil {
		return nil
	}
	if n.Type == NodeElement && n.GetAttr("id") == id {
		return n
	}
	for _, child := range n.Children {
		if found := child.GetElementById(id); found != nil {
			return found
		}
	}
	return nil
}

// GetElementsByTagName finds all elements with the given tag name
func (n *Node) GetElementsByTagName(tag string) []*Node {
	var results []*Node
	n.getElementsByTagNameRecursive(strings.ToLower(tag), &results)
	return results
}

func (n *Node) getElementsByTagNameRecursive(tag string, results *[]*Node) {
	if n == nil {
		return
	}
	if n.Type == NodeElement && strings.ToLower(n.Tag) == tag {
		*results = append(*results, n)
	}
	for _, child := range n.Children {
		child.getElementsByTagNameRecursive(tag, results)
	}
}

// GetElementsByClassName finds all elements with the given class name
func (n *Node) GetElementsByClassName(className string) []*Node {
	var results []*Node
	n.getElementsByClassNameRecursive(className, &results)
	return results
}

func (n *Node) getElementsByClassNameRecursive(className string, results *[]*Node) {
	if n == nil {
		return
	}
	if n.Type == NodeElement && n.HasClass(className) {
		*results = append(*results, n)
	}
	for _, child := range n.Children {
		child.getElementsByClassNameRecursive(className, results)
	}
}

// HasClass checks if the node has a specific class
func (n *Node) HasClass(className string) bool {
	classAttr := n.GetAttr("class")
	if classAttr == "" {
		return false
	}
	classes := strings.Fields(classAttr)
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}

// GetClasses returns all classes on the node
func (n *Node) GetClasses() []string {
	classAttr := n.GetAttr("class")
	if classAttr == "" {
		return nil
	}
	return strings.Fields(classAttr)
}

// ======================================================================================
// TEXT CONTENT
// ======================================================================================

// TextContent returns the text content of the node and its descendants
func (n *Node) TextContent() string {
	if n == nil {
		return ""
	}
	if n.Type == NodeText {
		return n.Content
	}
	var sb strings.Builder
	for _, child := range n.Children {
		sb.WriteString(child.TextContent())
		sb.WriteString(" ")
	}
	return strings.TrimSpace(sb.String())
}

// InnerText returns visible text content (excludes hidden elements)
func (n *Node) InnerText() string {
	if n == nil {
		return ""
	}
	if n.Type == NodeText {
		return n.Content
	}
	if n.Display == DisplayNone {
		return ""
	}
	var sb strings.Builder
	for _, child := range n.Children {
		text := child.InnerText()
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}
	return strings.TrimSpace(sb.String())
}

// ======================================================================================
// TREE NAVIGATION
// ======================================================================================

// FirstChild returns the first child node
func (n *Node) FirstChild() *Node {
	if n == nil || len(n.Children) == 0 {
		return nil
	}
	return n.Children[0]
}

// LastChild returns the last child node
func (n *Node) LastChild() *Node {
	if n == nil || len(n.Children) == 0 {
		return nil
	}
	return n.Children[len(n.Children)-1]
}

// NextSibling returns the next sibling in the parent's children
func (n *Node) NextSibling() *Node {
	if n == nil || n.Parent == nil {
		return nil
	}
	siblings := n.Parent.Children
	for i, child := range siblings {
		if child == n && i+1 < len(siblings) {
			return siblings[i+1]
		}
	}
	return nil
}

// PreviousSibling returns the previous sibling in the parent's children
func (n *Node) PreviousSibling() *Node {
	if n == nil || n.Parent == nil {
		return nil
	}
	siblings := n.Parent.Children
	for i, child := range siblings {
		if child == n && i > 0 {
			return siblings[i-1]
		}
	}
	return nil
}

// ChildElementCount returns the number of child elements (not text nodes)
func (n *Node) ChildElementCount() int {
	count := 0
	for _, child := range n.Children {
		if child.Type == NodeElement {
			count++
		}
	}
	return count
}

// FirstElementChild returns the first child that is an element
func (n *Node) FirstElementChild() *Node {
	for _, child := range n.Children {
		if child.Type == NodeElement {
			return child
		}
	}
	return nil
}

// LastElementChild returns the last child that is an element
func (n *Node) LastElementChild() *Node {
	for i := len(n.Children) - 1; i >= 0; i-- {
		if n.Children[i].Type == NodeElement {
			return n.Children[i]
		}
	}
	return nil
}

// ======================================================================================
// ANCESTORS
// ======================================================================================

// Ancestors returns all ancestor nodes from parent to root
func (n *Node) Ancestors() []*Node {
	var ancestors []*Node
	current := n.Parent
	for current != nil {
		ancestors = append(ancestors, current)
		current = current.Parent
	}
	return ancestors
}

// ClosestAncestor finds the nearest ancestor matching a tag
func (n *Node) ClosestAncestor(tag string) *Node {
	tag = strings.ToLower(tag)
	current := n.Parent
	for current != nil {
		if strings.ToLower(current.Tag) == tag {
			return current
		}
		current = current.Parent
	}
	return nil
}

// Contains checks if node contains the given descendant
func (n *Node) Contains(descendant *Node) bool {
	if n == nil || descendant == nil {
		return false
	}
	for _, child := range n.Children {
		if child == descendant {
			return true
		}
		if child.Contains(descendant) {
			return true
		}
	}
	return false
}

// Depth returns the depth of the node in the tree (root = 0)
func (n *Node) Depth() int {
	depth := 0
	current := n.Parent
	for current != nil {
		depth++
		current = current.Parent
	}
	return depth
}
