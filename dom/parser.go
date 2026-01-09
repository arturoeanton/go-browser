package dom

import "strings"

// Tokenizer splits HTML into tokens
type Tokenizer struct {
	Raw string
	Pos int
}

// NewTokenizer creates a new HTML tokenizer
func NewTokenizer(html string) *Tokenizer {
	return &Tokenizer{Raw: html, Pos: 0}
}

// HasMore returns true if there are more tokens
func (t *Tokenizer) HasMore() bool { return t.Pos < len(t.Raw) }

// NextToken returns the next token from the HTML
func (t *Tokenizer) NextToken() (tagName string, fullTag string, isTag bool, isClose bool) {
	start := t.Pos

	if t.Pos < len(t.Raw) && t.Raw[t.Pos] == '<' {
		isTag = true
		t.Pos++
		if t.Pos < len(t.Raw) && t.Raw[t.Pos] == '/' {
			isClose = true
			t.Pos++
		}

		// Skip DOCTYPE and other ! declarations
		if t.Pos < len(t.Raw) && t.Raw[t.Pos] == '!' {
			// Check for comment <!--
			if t.Pos+2 < len(t.Raw) && t.Raw[t.Pos:t.Pos+2] == "!-" {
				// Find end of comment -->
				for t.Pos < len(t.Raw) {
					if t.Pos+2 < len(t.Raw) && t.Raw[t.Pos:t.Pos+3] == "-->" {
						t.Pos += 3
						break
					}
					t.Pos++
				}
			} else {
				// Skip to end of declaration (like DOCTYPE)
				for t.Pos < len(t.Raw) && t.Raw[t.Pos] != '>' {
					t.Pos++
				}
				if t.Pos < len(t.Raw) {
					t.Pos++
				}
			}
			return "", "", false, false
		}

		for t.Pos < len(t.Raw) && t.Raw[t.Pos] != '>' {
			t.Pos++
		}
		offset := 1
		if isClose {
			offset = 2
		}
		if t.Pos > start+offset {
			fullTag = t.Raw[start+offset : t.Pos]
		}
		if t.Pos < len(t.Raw) {
			t.Pos++
		}

		// Extract just the tag name (first word)
		parts := strings.Fields(fullTag)
		if len(parts) > 0 {
			tagName = strings.TrimSuffix(parts[0], "/") // Handle self-closing like <br/>
		}
		return
	}

	isTag = false
	for t.Pos < len(t.Raw) && t.Raw[t.Pos] != '<' {
		t.Pos++
	}
	tagName = t.Raw[start:t.Pos]
	return
}

// ParseHTML parses HTML string into a DOM tree
func ParseHTML(html string) *Node {
	root := NewElement("root")
	current := root
	tokenizer := NewTokenizer(html)

	// Tags to skip entirely (including their content) - NOTE: script NOT included, we need to extract it
	skipTags := map[string]bool{"svg": true, "noscript": true, "template": true}

	// Tags where we need to preserve raw content (script, style)
	rawContentTags := map[string]bool{"script": true, "style": true}

	// Void elements that never have children
	voidElements := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}

	// Inline elements (for better layout)
	inlineElements := map[string]bool{
		"a": true, "abbr": true, "b": true, "bdo": true, "big": true, "br": true,
		"cite": true, "code": true, "dfn": true, "em": true, "i": true, "img": true,
		"input": true, "kbd": true, "label": true, "map": true, "object": true,
		"q": true, "samp": true, "script": true, "select": true, "small": true,
		"span": true, "strong": true, "sub": true, "sup": true, "textarea": true,
		"tt": true, "var": true, "button": true,
	}
	_ = inlineElements // Will be used for display mode detection

	for tokenizer.HasMore() {
		token, fullTag, isTag, isClose := tokenizer.NextToken()

		if isTag {
			tagName := strings.ToLower(strings.TrimSpace(token))

			// Skip empty tag names
			if tagName == "" {
				continue
			}

			// Check for self-closing tag syntax like <br/> or <img .../>
			isSelfClosing := strings.HasSuffix(fullTag, "/")

			if isClose {
				// Closing tag - walk up to find matching opening tag
				found := false
				for p := current; p != nil && p.Parent != nil; p = p.Parent {
					if p.Tag == tagName {
						current = p.Parent
						found = true
						break
					}
				}
				// Even if not found, continue parsing (malformed HTML tolerance)
				_ = found
			} else if rawContentTags[tagName] {
				// For script/style, preserve raw content as a child text node
				newNode := NewElement(tagName)
				newNode.Attributes = ParseAttributes(fullTag)
				current.AppendChild(newNode)

				// Find the closing tag and capture everything in between
				closeTag := "</" + tagName
				startPos := tokenizer.Pos
				for tokenizer.Pos < len(tokenizer.Raw) {
					// Look for closing tag
					if tokenizer.Pos+len(closeTag) <= len(tokenizer.Raw) {
						if strings.ToLower(tokenizer.Raw[tokenizer.Pos:tokenizer.Pos+len(closeTag)]) == closeTag {
							// Found closing tag, extract content
							content := tokenizer.Raw[startPos:tokenizer.Pos]
							if strings.TrimSpace(content) != "" {
								newNode.AppendChild(NewText(content))
							}
							// Skip past the closing tag
							for tokenizer.Pos < len(tokenizer.Raw) && tokenizer.Raw[tokenizer.Pos] != '>' {
								tokenizer.Pos++
							}
							if tokenizer.Pos < len(tokenizer.Raw) {
								tokenizer.Pos++
							}
							break
						}
					}
					tokenizer.Pos++
				}
			} else if skipTags[tagName] {
				// Skip entire content of these tags
				closeTag := "</" + tagName
				depth := 1
				for tokenizer.HasMore() && depth > 0 {
					_, _, nextIsTag, nextIsClose := tokenizer.NextToken()
					if nextIsTag {
						remaining := tokenizer.Raw[max(0, tokenizer.Pos-len(closeTag)-5):tokenizer.Pos]
						if nextIsClose && strings.Contains(strings.ToLower(remaining), closeTag) {
							depth--
						}
					}
				}
			} else if voidElements[tagName] || isSelfClosing {
				// Void/self-closing element - add but don't descend
				newNode := NewElement(tagName)
				newNode.Attributes = ParseAttributes(fullTag)
				current.AppendChild(newNode)
			} else {
				// Regular opening tag
				// Handle implicit closing for p, li elements
				if (tagName == "p" || tagName == "li") && current.Tag == tagName {
					current = current.Parent
				}

				newNode := NewElement(tagName)
				newNode.Attributes = ParseAttributes(fullTag)
				current.AppendChild(newNode)
				current = newNode
			}
		} else {
			// Text node - skip if only whitespace
			text := strings.TrimSpace(token)
			if len(text) > 0 {
				current.AppendChild(NewText(text))
			}
		}
	}
	return root
}
