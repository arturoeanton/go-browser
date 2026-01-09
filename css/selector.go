package css

import (
	"strconv"
	"strings"

	"go-browser/dom"
)

// ======================================================================================
// CSS SELECTORS
// ======================================================================================

// SelectorType indicates what kind of selector this is
type SelectorType int

const (
	SelectorUniversal  SelectorType = iota // *
	SelectorElement                        // div
	SelectorClass                          // .class
	SelectorID                             // #id
	SelectorAttribute                      // [attr]
	SelectorDescendant                     // div p
	SelectorChild                          // div > p
	SelectorAdjacent                       // div + p
	SelectorSibling                        // div ~ p
)

// Selector represents a CSS selector
type Selector struct {
	Type        SelectorType
	Element     string     // tag name for element selector
	Class       string     // class name for class selector
	ID          string     // id for id selector
	Attr        string     // attribute name
	AttrVal     string     // attribute value
	Parts       []Selector // for compound selectors
	PseudoClass string     // :first-child, :last-child, etc.
	IsChild     bool       // true if this is a child combinator (>)
}

// Specificity represents CSS specificity (a, b, c, d)
// a = inline styles, b = IDs, c = classes/attrs/pseudo-classes, d = elements/pseudo-elements
type Specificity struct {
	Inline   int
	IDs      int
	Classes  int
	Elements int
}

// Compare compares two specificities, returns 1 if s > other, -1 if s < other, 0 if equal
func (s Specificity) Compare(other Specificity) int {
	if s.Inline != other.Inline {
		if s.Inline > other.Inline {
			return 1
		}
		return -1
	}
	if s.IDs != other.IDs {
		if s.IDs > other.IDs {
			return 1
		}
		return -1
	}
	if s.Classes != other.Classes {
		if s.Classes > other.Classes {
			return 1
		}
		return -1
	}
	if s.Elements != other.Elements {
		if s.Elements > other.Elements {
			return 1
		}
		return -1
	}
	return 0
}

// ParseSelectors parses a selector list (comma-separated)
func ParseSelectors(text string) []Selector {
	var selectors []Selector
	parts := strings.Split(text, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			if sel := ParseSelector(part); sel.Type != SelectorUniversal || sel.Element == "*" {
				selectors = append(selectors, sel)
			}
		}
	}
	return selectors
}

// ParseSelector parses a single selector
func ParseSelector(text string) Selector {
	text = strings.TrimSpace(text)

	// Check for child combinator (>)
	if strings.Contains(text, ">") {
		parts := strings.Split(text, ">")
		var selectorParts []Selector
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				selectorParts = append(selectorParts, parseSimpleSelector(p))
			}
		}
		if len(selectorParts) > 0 {
			return Selector{
				Type:    SelectorChild,
				Parts:   selectorParts,
				IsChild: true,
			}
		}
	}

	// Check for combinators (space, +, ~)
	if strings.Contains(text, " ") && !strings.HasPrefix(text, ".") && !strings.HasPrefix(text, "#") {
		parts := strings.Fields(text)
		if len(parts) > 1 {
			var selectorParts []Selector
			for _, p := range parts {
				// Skip combinator tokens
				if p == "+" || p == "~" {
					continue
				}
				selectorParts = append(selectorParts, parseSimpleSelector(p))
			}
			if len(selectorParts) > 0 {
				return Selector{
					Type:  SelectorDescendant,
					Parts: selectorParts,
				}
			}
		}
	}

	return parseSimpleSelector(text)
}

// parseSimpleSelector parses a simple selector (element, class, id, or universal)
func parseSimpleSelector(text string) Selector {
	text = strings.TrimSpace(text)

	// Check for pseudo-class
	pseudoClass := ""
	if idx := strings.Index(text, ":"); idx != -1 {
		pseudoClass = text[idx+1:]
		text = text[:idx]
	}

	if text == "*" || text == "" {
		return Selector{Type: SelectorUniversal, Element: "*", PseudoClass: pseudoClass}
	}

	// ID selector
	if strings.HasPrefix(text, "#") {
		return Selector{Type: SelectorID, ID: text[1:], PseudoClass: pseudoClass}
	}

	// Class selector
	if strings.HasPrefix(text, ".") {
		return Selector{Type: SelectorClass, Class: text[1:], PseudoClass: pseudoClass}
	}

	// Element with class: div.class
	if idx := strings.Index(text, "."); idx > 0 {
		return Selector{
			Type:        SelectorElement,
			Element:     text[:idx],
			Class:       text[idx+1:],
			PseudoClass: pseudoClass,
		}
	}

	// Element with ID: div#id
	if idx := strings.Index(text, "#"); idx > 0 {
		return Selector{
			Type:        SelectorElement,
			Element:     text[:idx],
			ID:          text[idx+1:],
			PseudoClass: pseudoClass,
		}
	}

	// Attribute selector [attr] or [attr=value]
	if strings.HasPrefix(text, "[") && strings.HasSuffix(text, "]") {
		content := text[1 : len(text)-1]
		if eqIdx := strings.Index(content, "="); eqIdx != -1 {
			attr := content[:eqIdx]
			val := strings.Trim(content[eqIdx+1:], "\"'")
			return Selector{Type: SelectorAttribute, Attr: attr, AttrVal: val, PseudoClass: pseudoClass}
		}
		return Selector{Type: SelectorAttribute, Attr: content, PseudoClass: pseudoClass}
	}

	// Simple element selector
	return Selector{Type: SelectorElement, Element: strings.ToLower(text), PseudoClass: pseudoClass}
}

// Matches checks if a selector matches a DOM node
func (s Selector) Matches(node *dom.Node) bool {
	if node == nil || node.Type != dom.NodeElement {
		return false
	}

	// Check pseudo-class first if present
	if s.PseudoClass != "" && !matchesPseudoClass(node, s.PseudoClass) {
		return false
	}

	switch s.Type {
	case SelectorUniversal:
		return true

	case SelectorElement:
		matches := s.Element == "" || strings.EqualFold(node.Tag, s.Element)
		if matches && s.Class != "" {
			matches = nodeHasClass(node, s.Class)
		}
		if matches && s.ID != "" {
			matches = node.GetAttr("id") == s.ID
		}
		return matches

	case SelectorClass:
		return nodeHasClass(node, s.Class)

	case SelectorID:
		return node.GetAttr("id") == s.ID

	case SelectorAttribute:
		attrVal := node.GetAttr(s.Attr)
		if s.AttrVal != "" {
			return attrVal == s.AttrVal
		}
		return attrVal != ""

	case SelectorChild:
		// Child combinator: each part must be a direct child of the previous
		if len(s.Parts) == 0 {
			return false
		}
		// Last part must match current node
		if !s.Parts[len(s.Parts)-1].Matches(node) {
			return false
		}
		// Each previous part must match direct parent
		current := node.Parent
		for i := len(s.Parts) - 2; i >= 0; i-- {
			if current == nil || !s.Parts[i].Matches(current) {
				return false
			}
			current = current.Parent
		}
		return true

	case SelectorDescendant:
		// All parts must match in ancestor chain
		if len(s.Parts) == 0 {
			return false
		}
		// Last part must match current node
		if !s.Parts[len(s.Parts)-1].Matches(node) {
			return false
		}
		// Other parts must match ancestors
		current := node.Parent
		for i := len(s.Parts) - 2; i >= 0; i-- {
			found := false
			for current != nil {
				if s.Parts[i].Matches(current) {
					found = true
					current = current.Parent
					break
				}
				current = current.Parent
			}
			if !found {
				return false
			}
		}
		return true
	}

	return false
}

// matchesPseudoClass checks if a node matches a pseudo-class
func matchesPseudoClass(node *dom.Node, pseudoClass string) bool {
	if node.Parent == nil {
		return false
	}

	switch pseudoClass {
	case "first-child":
		// Check if this is the first element child
		for _, child := range node.Parent.Children {
			if child.Type == dom.NodeElement {
				return child == node
			}
		}
		return false
	case "last-child":
		// Check if this is the last element child
		for i := len(node.Parent.Children) - 1; i >= 0; i-- {
			child := node.Parent.Children[i]
			if child.Type == dom.NodeElement {
				return child == node
			}
		}
		return false
	case "only-child":
		count := 0
		for _, child := range node.Parent.Children {
			if child.Type == dom.NodeElement {
				count++
				if count > 1 {
					return false
				}
			}
		}
		return count == 1
	}

	// Handle :nth-child(n), :even, :odd
	if strings.HasPrefix(pseudoClass, "nth-child") {
		// Get element index among siblings
		idx := getElementIndex(node)

		// Parse the argument: nth-child(2), nth-child(odd), nth-child(even), nth-child(2n+1)
		inner := strings.TrimPrefix(pseudoClass, "nth-child(")
		inner = strings.TrimSuffix(inner, ")")

		switch inner {
		case "odd":
			return idx%2 == 1
		case "even":
			return idx%2 == 0
		default:
			// Try parsing as a number
			if n, err := strconv.Atoi(inner); err == nil {
				return idx == n
			}
		}
		return false
	}

	if pseudoClass == "even" {
		return getElementIndex(node)%2 == 0
	}
	if pseudoClass == "odd" {
		return getElementIndex(node)%2 == 1
	}

	return true // Unknown pseudo-classes pass through
}

// getElementIndex returns 1-based index of node among element siblings
func getElementIndex(node *dom.Node) int {
	if node.Parent == nil {
		return 1
	}
	idx := 0
	for _, child := range node.Parent.Children {
		if child.Type == dom.NodeElement {
			idx++
			if child == node {
				return idx
			}
		}
	}
	return 0
}

// nodeHasClass checks if a node has a specific class
func nodeHasClass(node *dom.Node, className string) bool {
	classAttr := node.GetAttr("class")
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

// CalculateSpecificity returns the specificity of a selector
func (s Selector) CalculateSpecificity() Specificity {
	spec := Specificity{}

	switch s.Type {
	case SelectorID:
		spec.IDs = 1
	case SelectorClass, SelectorAttribute:
		spec.Classes = 1
	case SelectorElement:
		if s.Element != "" && s.Element != "*" {
			spec.Elements = 1
		}
		if s.Class != "" {
			spec.Classes = 1
		}
		if s.ID != "" {
			spec.IDs = 1
		}
	case SelectorDescendant:
		for _, part := range s.Parts {
			partSpec := part.CalculateSpecificity()
			spec.IDs += partSpec.IDs
			spec.Classes += partSpec.Classes
			spec.Elements += partSpec.Elements
		}
	}

	return spec
}
