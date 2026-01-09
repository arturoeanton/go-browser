package dom

import (
	"fmt"
	realdom "go-browser/dom"

	"github.com/dop251/goja"
)

// JSNode wraps a real DOM Node for JavaScript access
type JSNode struct {
	node *realdom.Node
	vm   *goja.Runtime
}

// NewJSNode creates a JS-accessible wrapper around a real DOM node
func NewJSNode(node *realdom.Node, vm *goja.Runtime) *JSNode {
	if node == nil {
		return nil
	}
	return &JSNode{node: node, vm: vm}
}

// ToJSObject creates a JS object representing this node
func (n *JSNode) ToJSObject() *goja.Object {
	if n == nil || n.node == nil {
		return nil
	}

	obj := n.vm.NewObject()

	// Basic properties (safe - no recursion)
	obj.Set("tagName", n.node.Tag)
	obj.Set("nodeName", n.node.Tag)
	obj.Set("id", n.node.GetAttr("id"))
	obj.Set("className", n.node.GetAttr("class"))

	// nodeType: 1 for Element, 3 for Text
	nodeType := 1
	if n.node.Type == realdom.NodeText {
		nodeType = 3
	}
	obj.Set("nodeType", nodeType)

	// textContent as accessor property (get/set)
	obj.DefineAccessorProperty("textContent",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return n.vm.ToValue(n.getTextContent())
		}),
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			fmt.Printf("[textContent setter] Called with %d args\n", len(call.Arguments))
			if len(call.Arguments) > 0 {
				text := call.Argument(0).String()
				fmt.Printf("[textContent setter] Setting: %s\n", text)
				n.setTextContent(text)
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	// innerHTML as accessor property (get/set)
	obj.DefineAccessorProperty("innerHTML",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return n.vm.ToValue(n.getInnerHTML())
		}),
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				html := call.Argument(0).String()
				n.setInnerHTML(html)
			}
			return goja.Undefined()
		}),
		goja.FLAG_FALSE, goja.FLAG_TRUE)

	// Attributes as a map
	obj.Set("attributes", n.node.Attributes)

	// getAttribute method
	obj.Set("getAttribute", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}
		name := call.Argument(0).String()
		val := n.node.GetAttr(name)
		if val == "" {
			return goja.Null()
		}
		return n.vm.ToValue(val)
	})

	// setAttribute method
	obj.Set("setAttribute", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		name := call.Argument(0).String()
		value := call.Argument(1).String()
		if n.node.Attributes == nil {
			n.node.Attributes = make(map[string]string)
		}
		n.node.Attributes[name] = value
		return goja.Undefined()
	})

	// LAZY GETTERS for navigation properties (prevents recursion!)
	// These are functions that return the value when called
	obj.DefineAccessorProperty("parentNode",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getParentNode() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("parentElement",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getParentNode() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("childNodes",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getChildNodes() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("children",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getChildElements() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("firstChild",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getFirstChild() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("lastChild",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getLastChild() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("nextSibling",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getNextSibling() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("previousSibling",
		n.vm.ToValue(func(call goja.FunctionCall) goja.Value { return n.getPreviousSibling() }),
		goja.Undefined(), goja.FLAG_FALSE, goja.FLAG_TRUE)

	// appendChild method - adds a child node
	obj.Set("appendChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		// Extract the node from the JS object
		childObj := call.Argument(0).ToObject(n.vm)
		if childObj == nil {
			return goja.Undefined()
		}
		// For now, we can't get the real node back from JS object easily
		// This is a simplified implementation
		return call.Argument(0) // Return the child as per spec
	})

	// removeChild method - removes a child node
	obj.Set("removeChild", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		// Simplified - would need node mapping to properly implement
		return call.Argument(0)
	})

	// addEventListener method - crucial for interactivity!
	obj.Set("addEventListener", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}
		eventType := call.Argument(0).String()
		callback := call.Argument(1)

		if fn, ok := goja.AssertFunction(callback); ok {
			// Store the listener
			n.addEventListener(eventType, fn)
		}
		return goja.Undefined()
	})

	// click method - programmatic click
	obj.Set("click", func(call goja.FunctionCall) goja.Value {
		n.dispatchEvent("click")
		return goja.Undefined()
	})

	// querySelector method (searches within this node)
	obj.Set("querySelector", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}
		selector := call.Argument(0).String()
		found := n.querySelector(n.node, selector)
		if found == nil {
			return goja.Null()
		}
		return NewJSNode(found, n.vm).ToJSObject()
	})

	// querySelectorAll method
	obj.Set("querySelectorAll", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return n.vm.NewArray()
		}
		selector := call.Argument(0).String()
		nodes := n.querySelectorAll(n.node, selector)

		arr := n.vm.NewArray()
		for i, node := range nodes {
			jsNode := NewJSNode(node, n.vm).ToJSObject()
			arr.Set(intToString(i), jsNode)
		}
		arr.Set("length", len(nodes))
		return arr
	})

	return obj
}

func (n *JSNode) getTextContent() string {
	return collectText(n.node)
}

func collectText(node *realdom.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == realdom.NodeText {
		return node.Content
	}
	var text string
	for _, child := range node.Children {
		text += collectText(child)
	}
	return text
}

func (n *JSNode) getInnerHTML() string {
	// Simplified: just return text content for now
	return n.getTextContent()
}

// setTextContent replaces all children with a single text node
func (n *JSNode) setTextContent(text string) {
	nodeID := n.node.GetAttr("id")
	fmt.Printf("[setTextContent] Setting '%s' on #%s (tag=%s, ptr=%p)\n", text, nodeID, n.node.Tag, n.node)
	// Clear all children
	n.node.Children = nil
	// Add new text node
	textNode := realdom.NewText(text)
	n.node.AppendChild(textNode)
	fmt.Printf("[setTextContent] Node now has %d children\n", len(n.node.Children))
}

// setInnerHTML parses HTML and replaces children (simplified - just sets text for now)
func (n *JSNode) setInnerHTML(html string) {
	// For now, treat innerHTML like textContent
	// A full implementation would parse the HTML
	n.setTextContent(html)
}

// Event listener storage - CHANGED from pointer to ID-based for reliable matching
// Key is element ID, value is map of event types to callbacks
var nodeEventListenersByID = make(map[string]map[string][]goja.Callable)

// addEventListener registers an event listener on this node
func (n *JSNode) addEventListener(eventType string, callback goja.Callable) {
	// Use element ID for storage - if no ID, use a generated key
	nodeID := n.getNodeKey()

	if nodeEventListenersByID[nodeID] == nil {
		nodeEventListenersByID[nodeID] = make(map[string][]goja.Callable)
	}
	nodeEventListenersByID[nodeID][eventType] = append(nodeEventListenersByID[nodeID][eventType], callback)

	// Debug log
	fmt.Printf("[SpiderGopher] addEventListener: %s on #%s\n", eventType, nodeID)
}

// getNodeKey returns a unique key for this node (ID or generated)
func (n *JSNode) getNodeKey() string {
	if n.node == nil {
		return ""
	}
	// Prefer ID attribute
	if id := n.node.GetAttr("id"); id != "" {
		return id
	}
	// Fallback to pointer address (less reliable but better than nothing)
	return fmt.Sprintf("node_%p", n.node)
}

// dispatchEvent triggers all listeners for an event type
func (n *JSNode) dispatchEvent(eventType string) {
	nodeID := n.getNodeKey()
	listeners := nodeEventListenersByID[nodeID]
	if listeners == nil {
		return
	}
	callbacks := listeners[eventType]
	for _, cb := range callbacks {
		// Create a simple event object
		eventObj := n.vm.NewObject()
		eventObj.Set("type", eventType)
		eventObj.Set("target", n.ToJSObject())
		cb(goja.Undefined(), eventObj)
	}
}

// GetNodeListeners returns listeners for a node by ID
func GetNodeListeners(node *realdom.Node, eventType string) []goja.Callable {
	if node == nil {
		return nil
	}
	// Get ID from node
	nodeID := node.GetAttr("id")
	if nodeID == "" {
		nodeID = fmt.Sprintf("node_%p", node)
	}

	listeners := nodeEventListenersByID[nodeID]
	if listeners == nil {
		return nil
	}
	return listeners[eventType]
}

// DispatchClickEvent dispatches a click event to all listeners on a node
// This is called from the browser when a user clicks on an element
func DispatchClickEvent(node *realdom.Node, vm *goja.Runtime) {
	if node == nil || vm == nil {
		return
	}

	// Get ID from node
	nodeID := node.GetAttr("id")
	if nodeID == "" {
		nodeID = fmt.Sprintf("node_%p", node)
	}

	fmt.Printf("[SpiderGopher] DispatchClickEvent: looking for listeners on #%s\n", nodeID)

	callbacks := GetNodeListeners(node, "click")
	if len(callbacks) == 0 {
		fmt.Printf("[SpiderGopher] No click listeners found for #%s\n", nodeID)
		return
	}

	fmt.Printf("[SpiderGopher] Found %d click listeners for #%s\n", len(callbacks), nodeID)

	// Create event object
	eventObj := vm.NewObject()
	eventObj.Set("type", "click")
	eventObj.Set("bubbles", true)
	eventObj.Set("cancelable", true)

	// Call all callbacks
	for _, cb := range callbacks {
		cb(goja.Undefined(), eventObj)
	}
}

func (n *JSNode) getParentNode() goja.Value {
	if n.node.Parent == nil {
		return goja.Null()
	}
	return NewJSNode(n.node.Parent, n.vm).ToJSObject()
}

func (n *JSNode) getChildNodes() goja.Value {
	arr := n.vm.NewArray()
	for i, child := range n.node.Children {
		jsNode := NewJSNode(child, n.vm).ToJSObject()
		arr.Set(intToString(i), jsNode)
	}
	arr.Set("length", len(n.node.Children))
	return arr
}

func (n *JSNode) getChildElements() goja.Value {
	arr := n.vm.NewArray()
	idx := 0
	for _, child := range n.node.Children {
		if child.Type == realdom.NodeElement {
			jsNode := NewJSNode(child, n.vm).ToJSObject()
			arr.Set(intToString(idx), jsNode)
			idx++
		}
	}
	arr.Set("length", idx)
	return arr
}

func (n *JSNode) getFirstChild() goja.Value {
	if len(n.node.Children) == 0 {
		return goja.Null()
	}
	return NewJSNode(n.node.Children[0], n.vm).ToJSObject()
}

func (n *JSNode) getLastChild() goja.Value {
	if len(n.node.Children) == 0 {
		return goja.Null()
	}
	return NewJSNode(n.node.Children[len(n.node.Children)-1], n.vm).ToJSObject()
}

func (n *JSNode) getNextSibling() goja.Value {
	if n.node.Parent == nil {
		return goja.Null()
	}
	siblings := n.node.Parent.Children
	for i, sib := range siblings {
		if sib == n.node && i+1 < len(siblings) {
			return NewJSNode(siblings[i+1], n.vm).ToJSObject()
		}
	}
	return goja.Null()
}

func (n *JSNode) getPreviousSibling() goja.Value {
	if n.node.Parent == nil {
		return goja.Null()
	}
	siblings := n.node.Parent.Children
	for i, sib := range siblings {
		if sib == n.node && i > 0 {
			return NewJSNode(siblings[i-1], n.vm).ToJSObject()
		}
	}
	return goja.Null()
}

// querySelector - simple selector matching
func (n *JSNode) querySelector(node *realdom.Node, selector string) *realdom.Node {
	if node == nil {
		return nil
	}

	if n.matchesSelector(node, selector) {
		return node
	}

	for _, child := range node.Children {
		if found := n.querySelector(child, selector); found != nil {
			return found
		}
	}
	return nil
}

// querySelectorAll - simple selector matching
func (n *JSNode) querySelectorAll(node *realdom.Node, selector string) []*realdom.Node {
	var results []*realdom.Node
	n.collectMatching(node, selector, &results)
	return results
}

func (n *JSNode) collectMatching(node *realdom.Node, selector string, results *[]*realdom.Node) {
	if node == nil {
		return
	}
	if n.matchesSelector(node, selector) {
		*results = append(*results, node)
	}
	for _, child := range node.Children {
		n.collectMatching(child, selector, results)
	}
}

// matchesSelector - simple selector matching (tag, .class, #id)
func (n *JSNode) matchesSelector(node *realdom.Node, selector string) bool {
	if node.Type != realdom.NodeElement {
		return false
	}

	if len(selector) == 0 {
		return false
	}

	switch selector[0] {
	case '#': // ID selector
		return node.GetAttr("id") == selector[1:]
	case '.': // Class selector
		class := node.GetAttr("class")
		return containsClass(class, selector[1:])
	default: // Tag selector
		return node.Tag == selector
	}
}

func containsClass(classList, className string) bool {
	classes := splitClasses(classList)
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}

func splitClasses(s string) []string {
	var result []string
	current := ""
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func intToString(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
