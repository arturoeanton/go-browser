package dom

import (
	"fmt"
	realdom "go-browser/dom"

	"github.com/dop251/goja"
)

// DOMBridge connects SpiderGopher's JS runtime with the real DOM tree
type DOMBridge struct {
	root *realdom.Node
	vm   *goja.Runtime
}

// NewDOMBridge creates a new bridge to a real DOM tree
func NewDOMBridge(root *realdom.Node, vm *goja.Runtime) *DOMBridge {
	return &DOMBridge{root: root, vm: vm}
}

// SetRoot updates the DOM root (called when page loads)
func (b *DOMBridge) SetRoot(root *realdom.Node) {
	b.root = root
}

// GetDocumentObject returns a JS document object connected to the real DOM
func (b *DOMBridge) GetDocumentObject() *goja.Object {
	obj := b.vm.NewObject()

	// getElementById
	obj.Set("getElementById", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}
		id := call.Argument(0).String()
		node := b.findById(b.root, id)
		if node == nil {
			fmt.Printf("[DOMBridge] getElementById('%s') = null (not found)\n", id)
			return goja.Null()
		}
		fmt.Printf("[DOMBridge] getElementById('%s') = <%s> ptr=%p\n", id, node.Tag, node)
		return NewJSNode(node, b.vm).ToJSObject()
	})

	// getElementsByClassName
	obj.Set("getElementsByClassName", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return b.vm.NewArray()
		}
		className := call.Argument(0).String()
		nodes := b.findByClassName(b.root, className)
		return b.nodesToArray(nodes)
	})

	// getElementsByTagName
	obj.Set("getElementsByTagName", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return b.vm.NewArray()
		}
		tagName := call.Argument(0).String()
		nodes := b.findByTagName(b.root, tagName)
		return b.nodesToArray(nodes)
	})

	// querySelector
	obj.Set("querySelector", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}
		selector := call.Argument(0).String()
		jsNode := NewJSNode(b.root, b.vm)
		node := jsNode.querySelector(b.root, selector)
		if node == nil {
			return goja.Null()
		}
		return NewJSNode(node, b.vm).ToJSObject()
	})

	// querySelectorAll
	obj.Set("querySelectorAll", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return b.vm.NewArray()
		}
		selector := call.Argument(0).String()
		jsNode := NewJSNode(b.root, b.vm)
		nodes := jsNode.querySelectorAll(b.root, selector)
		return b.nodesToArray(nodes)
	})

	// createElement
	obj.Set("createElement", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}
		tagName := call.Argument(0).String()
		newNode := realdom.NewElement(tagName)
		return NewJSNode(newNode, b.vm).ToJSObject()
	})

	// createTextNode
	obj.Set("createTextNode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Null()
		}
		text := call.Argument(0).String()
		newNode := realdom.NewText(text)
		return NewJSNode(newNode, b.vm).ToJSObject()
	})

	// documentElement (root html element)
	obj.Set("documentElement", func() goja.Value {
		html := b.findByTagName(b.root, "html")
		if len(html) > 0 {
			return NewJSNode(html[0], b.vm).ToJSObject()
		}
		return goja.Null()
	}())

	// body
	obj.Set("body", func() goja.Value {
		body := b.findByTagName(b.root, "body")
		if len(body) > 0 {
			return NewJSNode(body[0], b.vm).ToJSObject()
		}
		return goja.Null()
	}())

	// head
	obj.Set("head", func() goja.Value {
		head := b.findByTagName(b.root, "head")
		if len(head) > 0 {
			return NewJSNode(head[0], b.vm).ToJSObject()
		}
		return goja.Null()
	}())

	return obj
}

func (b *DOMBridge) findById(node *realdom.Node, id string) *realdom.Node {
	if node == nil {
		return nil
	}
	if node.GetAttr("id") == id {
		return node
	}
	for _, child := range node.Children {
		if found := b.findById(child, id); found != nil {
			return found
		}
	}
	return nil
}

func (b *DOMBridge) findByClassName(node *realdom.Node, className string) []*realdom.Node {
	var results []*realdom.Node
	b.collectByClass(node, className, &results)
	return results
}

func (b *DOMBridge) collectByClass(node *realdom.Node, className string, results *[]*realdom.Node) {
	if node == nil {
		return
	}
	if containsClass(node.GetAttr("class"), className) {
		*results = append(*results, node)
	}
	for _, child := range node.Children {
		b.collectByClass(child, className, results)
	}
}

func (b *DOMBridge) findByTagName(node *realdom.Node, tagName string) []*realdom.Node {
	var results []*realdom.Node
	b.collectByTag(node, tagName, &results)
	return results
}

func (b *DOMBridge) collectByTag(node *realdom.Node, tagName string, results *[]*realdom.Node) {
	if node == nil {
		return
	}
	if node.Tag == tagName {
		*results = append(*results, node)
	}
	for _, child := range node.Children {
		b.collectByTag(child, tagName, results)
	}
}

func (b *DOMBridge) nodesToArray(nodes []*realdom.Node) goja.Value {
	arr := b.vm.NewArray()
	for i, node := range nodes {
		jsNode := NewJSNode(node, b.vm).ToJSObject()
		arr.Set(intToString(i), jsNode)
	}
	arr.Set("length", len(nodes))
	return arr
}
