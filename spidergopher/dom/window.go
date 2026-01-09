package dom

// Window represents the global window object.
type Window struct {
	*EventTarget
	Document *Document
}

func NewWindow() *Window {
	return &Window{
		EventTarget: NewEventTarget(),
		Document:    NewDocument(),
	}
}

// Document represents the document object
type Document struct {
	*EventTarget
}

func NewDocument() *Document {
	return &Document{
		EventTarget: NewEventTarget(),
	}
}

// GetElementById returns a mock element (to be replaced with real DOM later)
func (d *Document) GetElementById(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":       id,
		"tagName":  "DIV",
		"nodeName": "DIV",
	}
}
