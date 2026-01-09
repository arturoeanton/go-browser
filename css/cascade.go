package css

import (
	"go-browser/dom"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// ======================================================================================
// CASCADE & STYLE COMPUTATION
// ======================================================================================

// StyleEntry represents a matched rule with its specificity and order
type StyleEntry struct {
	Declarations []Declaration
	Specificity  Specificity
	Order        int
	Important    bool
}

// ComputeStyles calculates the final computed style for a DOM node
func ComputeStyles(node *dom.Node, stylesheets []*Stylesheet) *ComputedStyle {
	if node == nil || node.Type != dom.NodeElement {
		return NewComputedStyle()
	}

	// Start with defaults for the tag
	style := DefaultForTag(node.Tag)

	// Collect all matching rules
	var entries []StyleEntry
	order := 0

	// From stylesheets
	for _, stylesheet := range stylesheets {
		for _, rule := range stylesheet.Rules {
			for _, selector := range rule.Selectors {
				if selector.Matches(node) {
					for _, decl := range rule.Declarations {
						entries = append(entries, StyleEntry{
							Declarations: []Declaration{decl},
							Specificity:  selector.CalculateSpecificity(),
							Order:        order,
							Important:    decl.Important,
						})
						order++
					}
				}
			}
		}
	}

	// From inline style attribute
	inlineStyle := node.GetAttr("style")
	if inlineStyle != "" {
		declarations := ParseInlineStyle(inlineStyle)
		for _, decl := range declarations {
			entries = append(entries, StyleEntry{
				Declarations: []Declaration{decl},
				Specificity:  Specificity{Inline: 1},
				Order:        order,
				Important:    decl.Important,
			})
			order++
		}
	}

	// Sort by cascade order: important, specificity, source order
	sort.SliceStable(entries, func(i, j int) bool {
		// Important declarations win
		if entries[i].Important != entries[j].Important {
			return !entries[i].Important // non-important comes first
		}
		// Then by specificity
		cmp := entries[i].Specificity.Compare(entries[j].Specificity)
		if cmp != 0 {
			return cmp < 0 // lower specificity comes first
		}
		// Then by source order
		return entries[i].Order < entries[j].Order
	})

	// Apply in order (later declarations override earlier)
	for _, entry := range entries {
		ApplyDeclarations(style, entry.Declarations)
	}

	return style
}

// ApplyStylesToTree applies computed styles to all nodes in a DOM tree
func ApplyStylesToTree(root *dom.Node, stylesheets []*Stylesheet) {
	applyStylesRecursive(root, stylesheets)
}

func applyStylesRecursive(node *dom.Node, stylesheets []*Stylesheet) {
	if node == nil {
		return
	}

	if node.Type == dom.NodeElement {
		node.ComputedStyle = ComputeStyles(node, stylesheets)

		// Inherit from parent if available
		if node.Parent != nil && node.Parent.ComputedStyle != nil {
			if parentStyle, ok := node.Parent.ComputedStyle.(*ComputedStyle); ok {
				if childStyle, ok := node.ComputedStyle.(*ComputedStyle); ok {
					InheritFromParent(childStyle, parentStyle)
				}
			}
		}
	}

	for _, child := range node.Children {
		applyStylesRecursive(child, stylesheets)
	}
}

// InheritableProperties lists CSS properties that inherit from parent
var InheritableProperties = map[string]bool{
	"color":       true,
	"font-family": true,
	"font-size":   true,
	"font-weight": true,
	"line-height": true,
	"text-align":  true,
	"visibility":  true,
}

// InheritFromParent applies inherited properties from parent style
func InheritFromParent(child, parent *ComputedStyle) {
	if parent == nil || child == nil {
		return
	}

	// Only inherit if child hasn't set its own value
	// For now, inherit key properties
	if child.FontSize == 16 && parent.FontSize != 16 {
		child.FontSize = parent.FontSize
	}
	if child.FontWeight == 400 && parent.FontWeight != 400 {
		child.FontWeight = parent.FontWeight
	}
	// Color inherits
	child.Color = parent.Color
}

// ExtractStylesheets finds and parses all <style> blocks in a DOM tree
func ExtractStylesheets(root *dom.Node) []*Stylesheet {
	var stylesheets []*Stylesheet
	extractStylesRecursive(root, &stylesheets)
	return stylesheets
}

func extractStylesRecursive(node *dom.Node, stylesheets *[]*Stylesheet) {
	if node == nil {
		return
	}

	if node.Tag == "style" {
		// Get text content
		var cssText string
		for _, child := range node.Children {
			if child.Type == dom.NodeText {
				cssText += child.Content
			}
		}
		if cssText != "" {
			*stylesheets = append(*stylesheets, ParseStylesheet(cssText))
		}
	}

	for _, child := range node.Children {
		extractStylesRecursive(child, stylesheets)
	}
}

// ======================================================================================
// EXTERNAL CSS FETCHING
// ======================================================================================

// FetchExternalStylesheets finds <link rel="stylesheet"> tags and fetches CSS
func FetchExternalStylesheets(root *dom.Node, baseURL string) []*Stylesheet {
	// Find all link tags with rel="stylesheet"
	var cssURLs []string
	findStylesheetLinks(root, &cssURLs)

	if len(cssURLs) == 0 {
		return nil
	}

	// Resolve relative URLs and fetch in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	var stylesheets []*Stylesheet

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	for _, cssURL := range cssURLs {
		// Resolve relative URL
		fullURL := resolveURL(cssURL, baseURL)
		if fullURL == "" {
			continue
		}

		wg.Add(1)
		go func(u string) {
			defer wg.Done()

			resp, err := client.Get(u)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return
			}

			// Read CSS content
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}

			// Parse stylesheet
			stylesheet := ParseStylesheet(string(body))
			if stylesheet != nil && len(stylesheet.Rules) > 0 {
				mu.Lock()
				stylesheets = append(stylesheets, stylesheet)
				mu.Unlock()
			}
		}(fullURL)
	}

	wg.Wait()
	return stylesheets
}

// findStylesheetLinks recursively finds all <link rel="stylesheet" href="...">
func findStylesheetLinks(node *dom.Node, urls *[]string) {
	if node == nil {
		return
	}

	if node.Tag == "link" {
		rel := strings.ToLower(node.GetAttr("rel"))
		href := node.GetAttr("href")
		if rel == "stylesheet" && href != "" {
			*urls = append(*urls, href)
		}
	}

	for _, child := range node.Children {
		findStylesheetLinks(child, urls)
	}
}

// resolveURL resolves a relative URL against a base URL
func resolveURL(href, baseURL string) string {
	if href == "" {
		return ""
	}

	// Already absolute
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	// Parse base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	// Parse relative URL
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}

	// Resolve
	return base.ResolveReference(ref).String()
}
