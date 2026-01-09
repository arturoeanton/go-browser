// Package browser provides the main browser application
package browser

import (
	"fmt"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"go-browser/css"
	"go-browser/dom"
	"go-browser/gocko/forms"
	"go-browser/layout"
	"go-browser/render"
	"go-browser/spidergopher"
	spiderdom "go-browser/spidergopher/dom"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Constants
const (
	WindowWidth  = 1024
	WindowHeight = 768
	NavBarHeight = 56.0
	URLBarHeight = 36.0
	Padding      = 16.0
	ContentTop   = NavBarHeight + 20
	FontSizeBody = 15
	FontSizeUI   = 14
)

// Colors
var (
	ColorBackground    = color.RGBA{255, 255, 255, 255} // White default
	ColorSurface       = color.RGBA{245, 245, 245, 255}
	ColorNavBar        = color.RGBA{28, 28, 32, 255}
	ColorURLBar        = color.RGBA{45, 45, 52, 255}
	ColorURLBarHover   = color.RGBA{55, 55, 64, 255}
	ColorURLBarFocus   = color.RGBA{70, 70, 82, 255}
	ColorText          = color.RGBA{33, 33, 33, 255} // Dark text
	ColorTextMuted     = color.RGBA{100, 100, 110, 255}
	ColorAccent        = color.RGBA{25, 118, 210, 255}
	ColorBorder        = color.RGBA{200, 200, 210, 255}
	ColorButton        = color.RGBA{230, 230, 235, 255}
	ColorButtonPrimary = color.RGBA{76, 120, 90, 255}
	ColorButtonText    = color.RGBA{255, 255, 255, 255}
	ColorCursor        = color.RGBA{25, 118, 210, 255}
	ColorLink          = color.RGBA{25, 118, 210, 255}
	ColorHR            = color.RGBA{200, 200, 210, 255}
	ColorTableRow1     = color.RGBA{250, 250, 252, 255}
	ColorTableRow2     = color.RGBA{240, 240, 245, 255}
	ColorImageBg       = color.RGBA{230, 230, 235, 255}
)

// NavBar represents the navigation bar
type NavBar struct {
	URLText     string
	IsEditing   bool
	IsHovering  bool
	CursorPos   int
	CursorBlink int
	URLBarX     float32
	URLBarY     float32
	URLBarW     float32
}

// App represents the browser application
type App struct {
	URL               string
	DOMRoot           *dom.Node
	RenderTree        *layout.RenderBox
	Stylesheets       []*css.Stylesheet
	ScrollY           float64
	IsLoading         bool
	ErrorMsg          string
	NavBar            NavBar
	History           []string             // Browser history
	HistoryPos        int                  // Current position in history
	FormState         *forms.FormState     // Form element state
	captureScreenshot bool                 // Flag to capture screenshot on next draw
	JSEngine          *spidergopher.Engine // SpiderGopher JavaScript engine
}

// NewApp creates a new browser application
func NewApp() *App {
	return &App{
		URL:        "https://example.com",
		History:    []string{},
		HistoryPos: -1,
		FormState:  forms.NewFormState(),
	}
}

// Navigate navigates to a URL and adds it to history
func (a *App) Navigate(urlStr string) {
	// Truncate forward history if we were in the middle
	if a.HistoryPos < len(a.History)-1 {
		a.History = a.History[:a.HistoryPos+1]
	}
	// Add to history
	a.History = append(a.History, urlStr)
	a.HistoryPos = len(a.History) - 1
	a.URL = urlStr
	a.LoadFromURL(urlStr)
}

// LoadContent parses and renders HTML content
func (a *App) LoadContent(rawHTML string) {
	// Parse HTML into DOM
	a.DOMRoot = dom.ParseHTML(rawHTML)

	// Extract <style> blocks
	a.Stylesheets = css.ExtractStylesheets(a.DOMRoot)

	// Fetch external stylesheets from <link rel="stylesheet">
	externalCSS := css.FetchExternalStylesheets(a.DOMRoot, render.CurrentBaseURL)
	if len(externalCSS) > 0 {
		a.Stylesheets = append(a.Stylesheets, externalCSS...)
	}

	// Apply CSS to DOM tree
	css.ApplyStylesToTree(a.DOMRoot, a.Stylesheets)

	// Build render tree with computed styles
	a.RenderTree = layout.BuildRenderTree(a.DOMRoot, WindowWidth-(Padding*2))

	// Initialize SpiderGopher and connect to DOM
	a.initJSEngine()
}

// LoadFromURL fetches and loads content from a URL
func (a *App) LoadFromURL(urlStr string) {
	// Handle file:// protocol for local files
	if strings.HasPrefix(urlStr, "file://") {
		path := strings.TrimPrefix(urlStr, "file://")
		a.LoadFromFile(path)
		return
	}

	// Handle relative paths (no protocol) as local files
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		// Check if it's a local file path
		if _, err := os.Stat(urlStr); err == nil {
			a.LoadFromFile(urlStr)
			return
		}
		// Otherwise assume https
		urlStr = "https://" + urlStr
	}

	a.IsLoading = true
	render.CurrentBaseURL = urlStr
	go func() {
		resp, err := http.Get(urlStr)
		if err != nil {
			a.ErrorMsg = err.Error()
			a.IsLoading = false
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		a.LoadContent(string(body))
		a.IsLoading = false
	}()
}

// LoadFromFile loads HTML from a local file
func (a *App) LoadFromFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		a.ErrorMsg = "File not found: " + err.Error()
		return
	}
	a.LoadContent(string(content))
}

// Update handles input and updates state
func (a *App) Update() error {

	_, dy := ebiten.Wheel()
	a.ScrollY += dy * 30
	if a.ScrollY > 0 {
		a.ScrollY = 0
	}

	// Update form state cursor blink
	a.FormState.CursorBlink++

	// Handle mouse clicks
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()

		// First check nav bar
		a.NavBar.HandleClick(a, mx, my)

		// Then check content area
		if my > int(NavBarHeight) && a.RenderTree != nil {
			clickX := float64(mx) - Padding
			clickY := float64(my) - ContentTop - a.ScrollY

			// First try to handle form element clicks
			if handled := a.handleFormClick(a.RenderTree, clickX, clickY); handled {
				// Form element handled the click
			} else {
				// Check for link clicks
				clickedURL := a.findClickedLink(a.RenderTree, clickX, clickY)
				if clickedURL != "" {
					if strings.HasPrefix(clickedURL, "#") {
						// Anchor link
					} else if strings.HasPrefix(clickedURL, "http") {
						a.Navigate(clickedURL)
					} else if strings.HasPrefix(clickedURL, "/") {
						if base, err := url.Parse(a.URL); err == nil {
							rel, _ := url.Parse(clickedURL)
							fullURL := base.ResolveReference(rel).String()
							a.Navigate(fullURL)
						}
					} else {
						if base, err := url.Parse(a.URL); err == nil {
							rel, _ := url.Parse(clickedURL)
							fullURL := base.ResolveReference(rel).String()
							a.Navigate(fullURL)
						}
					}
				} else {
					// Click outside form elements - clear focus
					a.FormState.ClearFocus()
				}
			}
		}
	}

	// Handle keyboard input for focused form elements
	if a.FormState.FocusedID != "" && !a.NavBar.IsEditing {
		runes := ebiten.AppendInputChars(nil)
		var keys []ebiten.Key
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			keys = append(keys, ebiten.KeyBackspace)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
			keys = append(keys, ebiten.KeyDelete)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			keys = append(keys, ebiten.KeyLeft)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			keys = append(keys, ebiten.KeyRight)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
			keys = append(keys, ebiten.KeyHome)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
			keys = append(keys, ebiten.KeyEnd)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			keys = append(keys, ebiten.KeyEnter)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
			keys = append(keys, ebiten.KeyTab)
		}

		if len(runes) > 0 || len(keys) > 0 {
			a.handleFormInput(runes, keys)
		}
	}

	// URL bar hover detection
	mx, my := ebiten.CursorPosition()
	a.NavBar.IsHovering = float32(my) >= a.NavBar.URLBarY &&
		float32(my) <= a.NavBar.URLBarY+URLBarHeight &&
		float32(mx) >= a.NavBar.URLBarX &&
		float32(mx) <= a.NavBar.URLBarX+a.NavBar.URLBarW

	// Change cursor based on hover target
	btnSize := float32(30)
	btnSpacing := float32(6)
	btnStartX := float32(12)
	btnY := float32((NavBarHeight - btnSize) / 2)

	isOverButton := float32(my) >= btnY && float32(my) <= btnY+btnSize &&
		float32(mx) >= btnStartX && float32(mx) <= btnStartX+(btnSize+btnSpacing)*3

	if isOverButton {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else if my > int(NavBarHeight) && a.RenderTree != nil {
		hoveredURL := a.findClickedLink(a.RenderTree, float64(mx)-Padding, float64(my)-ContentTop-a.ScrollY)
		if hoveredURL != "" {
			ebiten.SetCursorShape(ebiten.CursorShapePointer)
		} else {
			ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		}
	} else if a.NavBar.IsHovering {
		ebiten.SetCursorShape(ebiten.CursorShapeText)
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}

	// Handle URL bar input
	a.NavBar.HandleInput(a)

	// Reload with R key (only when not editing URL or form)
	if !a.NavBar.IsEditing && a.FormState.FocusedID == "" && inpututil.IsKeyJustPressed(ebiten.KeyR) {
		if strings.HasPrefix(a.URL, "http") {
			a.LoadFromURL(a.URL)
		}
	}
	return nil
}

// findClickedLink recursively finds a link at the given coordinates
func (a *App) findClickedLink(box *layout.RenderBox, x, y float64) string {
	if box == nil {
		return ""
	}

	// Check if click is within this box and it's a link
	if box.IsLink && box.LinkURL != "" {
		if x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H {
			return box.LinkURL
		}
	}

	// Check children
	for _, child := range box.Children {
		if url := a.findClickedLink(child, x, y); url != "" {
			return url
		}
	}

	return ""
}

// handleFormClick recursively finds and handles form element clicks
func (a *App) handleFormClick(box *layout.RenderBox, x, y float64) bool {
	if box == nil {
		return false
	}

	// Check if click is within this box
	if box.Node != nil && x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+box.H {
		// Dispatch click event to SpiderGopher listeners
		a.dispatchJSClickEvent(box.Node)
	}

	// Check if this is a form element
	if box.Node != nil && forms.IsInteractive(box.Node.Tag) {
		// For select elements, expand hit area when dropdown is open
		hitH := box.H
		if box.Node.Tag == "select" {
			id := forms.GetElementID(box.Node)
			if a.FormState.SelectOpen == id {
				// Expand hit area to include dropdown options
				options := a.countSelectOptions(box.Node)
				hitH += float64(options)*28 + 10 // 28 = option height
			}
		}

		if x >= box.X && x <= box.X+box.W && y >= box.Y && y <= box.Y+hitH {
			if handler := forms.GetHandler(box.Node.Tag); handler != nil {
				return handler.HandleClick(box, box.Node, x, y, a.FormState)
			}
		}
	}

	// Check children
	for _, child := range box.Children {
		if a.handleFormClick(child, x, y) {
			return true
		}
	}

	return false
}

// countSelectOptions counts the number of options in a select
func (a *App) countSelectOptions(node *dom.Node) int {
	count := 0
	for _, child := range node.Children {
		if child.Tag == "option" {
			count++
		}
	}
	return count
}

// handleFormInput sends keyboard input to the focused form element
func (a *App) handleFormInput(runes []rune, keys []ebiten.Key) {
	if a.FormState.FocusedID == "" {
		return
	}

	// Handle Tab key for navigation
	for _, key := range keys {
		if key == ebiten.KeyTab {
			a.focusNextElement()
			return
		}
	}

	// Find the focused element and its handler
	focusedNode := a.findNodeByID(a.DOMRoot, a.FormState.FocusedID)
	if focusedNode == nil {
		return
	}

	if handler := forms.GetHandler(focusedNode.Tag); handler != nil {
		handler.HandleInput(focusedNode, runes, keys, a.FormState)
	}
}

// focusNextElement moves focus to the next focusable form element
func (a *App) focusNextElement() {
	// Collect all focusable elements
	var focusableIDs []string
	a.collectFocusableIDs(a.DOMRoot, &focusableIDs)

	if len(focusableIDs) == 0 {
		return
	}

	// Find current position
	currentIdx := -1
	for i, id := range focusableIDs {
		if id == a.FormState.FocusedID {
			currentIdx = i
			break
		}
	}

	// Move to next element (wrap around)
	nextIdx := (currentIdx + 1) % len(focusableIDs)
	a.FormState.FocusedID = focusableIDs[nextIdx]
	a.FormState.CursorPos = len(a.FormState.Values[a.FormState.FocusedID])
}

// collectFocusableIDs collects IDs of all focusable form elements
func (a *App) collectFocusableIDs(node *dom.Node, ids *[]string) {
	if node == nil {
		return
	}

	// Check if this is a focusable form element
	if forms.IsInteractive(node.Tag) {
		if handler := forms.GetHandler(node.Tag); handler != nil && handler.IsFocusable() {
			id := forms.GetElementID(node)
			*ids = append(*ids, id)
		}
	}

	// Check children
	for _, child := range node.Children {
		a.collectFocusableIDs(child, ids)
	}
}

// findNodeByID finds a DOM node by its id or name attribute
func (a *App) findNodeByID(node *dom.Node, id string) *dom.Node {
	if node == nil {
		return nil
	}

	// Check this node
	nodeID := node.GetAttr("id")
	if nodeID == "" {
		nodeID = node.GetAttr("name")
	}
	if nodeID == "" {
		nodeID = node.Tag
	}
	if nodeID == id {
		return node
	}

	// Check children
	for _, child := range node.Children {
		if found := a.findNodeByID(child, id); found != nil {
			return found
		}
	}

	return nil
}

// Draw renders the browser window
func (a *App) Draw(screen *ebiten.Image) {
	// Get page background from body/html computed style
	pageBackground := ColorBackground
	if a.DOMRoot != nil {
		pageBackground = a.getPageBackground()
	}

	// Check for gradient background on body
	if gradient := a.getBodyGradient(); gradient != nil && len(gradient.Stops) >= 2 {
		// Convert CSS gradient stops to render stops
		stops := make([]render.GradientStop, len(gradient.Stops))
		for i, s := range gradient.Stops {
			stops[i] = render.GradientStop{
				R:        float64(s.Color.R),
				G:        float64(s.Color.G),
				B:        float64(s.Color.B),
				A:        float64(s.Color.A),
				Position: s.Position,
			}
		}
		render.DrawLinearGradient(screen, 0, float32(NavBarHeight), float32(WindowWidth), float32(WindowHeight-NavBarHeight), gradient.Angle, stops)
	} else {
		screen.Fill(pageBackground)
	}

	// Draw content area
	if a.IsLoading {
		render.DrawText(screen, "Loading...", Padding, ContentTop+30, FontSizeBody, ColorTextMuted)
	} else if a.ErrorMsg != "" {
		render.DrawText(screen, "Error: "+a.ErrorMsg, Padding, ContentTop+30, FontSizeBody, color.RGBA{255, 100, 100, 255})
	} else if a.RenderTree != nil {
		a.renderNode(screen, a.RenderTree, Padding, ContentTop+a.ScrollY)

		// Render select dropdown overlay (on top of everything)
		if a.FormState.SelectOpen != "" {
			a.renderSelectOverlay(screen, a.RenderTree, Padding, ContentTop+a.ScrollY)
		}
	}

	// Draw nav bar on top
	a.NavBar.Draw(screen, a)

	// Capture screenshot if requested
	if a.captureScreenshot {
		a.saveScreenshot(screen)
		a.captureScreenshot = false
	}
}

// saveScreenshot saves the current screen to a PNG file
func (a *App) saveScreenshot(screen *ebiten.Image) {
	filename := fmt.Sprintf("screenshot_%s.png", time.Now().Format("20060102_150405"))
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating screenshot:", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, screen); err != nil {
		fmt.Println("Error encoding screenshot:", err)
		return
	}
	fmt.Println("Screenshot saved:", filename)
}

// renderSelectOverlay finds and renders the open select dropdown on top of other content
func (a *App) renderSelectOverlay(screen *ebiten.Image, box *layout.RenderBox, offsetX, offsetY float64) {
	if box == nil || box.Node == nil {
		return
	}

	// Check if this is the open select
	if box.Node.Tag == "select" {
		id := forms.GetElementID(box.Node)
		if a.FormState.SelectOpen == id {
			// Render the dropdown portion
			absY := box.Y + offsetY
			if handler, ok := forms.GetHandler("select").(*forms.SelectHandler); ok {
				tempBox := &layout.RenderBox{
					Node: box.Node,
					X:    box.X + offsetX,
					Y:    absY,
					W:    box.W,
					H:    box.H,
				}
				handler.RenderDropdownOnly(screen, tempBox, box.Node, a.FormState)
			}
		}
	}

	// Check children
	for _, child := range box.Children {
		a.renderSelectOverlay(screen, child, offsetX, offsetY)
	}
}

// getPageBackground extracts background color from body or html element
func (a *App) getPageBackground() color.RGBA {
	// Look for body first, then html
	for _, child := range a.DOMRoot.Children {
		bg := a.findBackgroundColor(child)
		if bg.A > 0 {
			return bg
		}
	}
	return ColorBackground
}

// getBodyGradient returns the gradient from body element if present
func (a *App) getBodyGradient() *css.Gradient {
	if a.DOMRoot == nil {
		return nil
	}
	return a.findBodyGradient(a.DOMRoot)
}

func (a *App) findBodyGradient(node *dom.Node) *css.Gradient {
	if node == nil {
		return nil
	}

	if node.Tag == "body" {
		if node.ComputedStyle != nil {
			if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
				return cs.BackgroundGradient
			}
		}
	}

	// Search children
	for _, child := range node.Children {
		if g := a.findBodyGradient(child); g != nil {
			return g
		}
	}
	return nil
}

func (a *App) findBackgroundColor(node *dom.Node) color.RGBA {
	if node == nil {
		return color.RGBA{}
	}

	if node.Tag == "body" || node.Tag == "html" {
		if node.ComputedStyle != nil {
			if cs, ok := node.ComputedStyle.(*css.ComputedStyle); ok {
				if cs.BackgroundColor.A > 0 {
					return cs.BackgroundColor
				}
			}
		}
	}

	// Search children
	for _, child := range node.Children {
		bg := a.findBackgroundColor(child)
		if bg.A > 0 {
			return bg
		}
	}

	return color.RGBA{}
}

// Layout returns the window size
func (a *App) Layout(w, h int) (int, int) {
	return WindowWidth, WindowHeight
}

func (a *App) renderNode(screen *ebiten.Image, box *layout.RenderBox, offsetX, offsetY float64) {
	// Handle position:fixed - ignore scroll offset
	if box.IsFixed {
		offsetY = ContentTop // Fixed elements stay at top, ignore scroll
	}

	absY := box.Y + offsetY

	// Draw CSS background-color for any element with computed style
	// Skip form elements - they have their own handlers
	if box.Node != nil && box.Node.ComputedStyle != nil {
		tag := box.Node.Tag
		isFormElement := tag == "input" || tag == "button" || tag == "select" || tag == "textarea"
		if !isFormElement {
			if cs, ok := box.Node.ComputedStyle.(*css.ComputedStyle); ok {
				if cs.BackgroundColor.A > 0 && tag != "body" && tag != "html" {
					vector.DrawFilledRect(screen,
						float32(box.X+offsetX), float32(absY),
						float32(box.W), float32(box.H),
						cs.BackgroundColor, false)
				}
			}
		}
	}

	// Draw backgrounds based on node type (table, hr, etc)
	if box.Node != nil {
		switch box.Node.Tag {
		case "table":
			render.DrawRoundedRect(screen,
				float32(offsetX), float32(absY),
				float32(box.W), float32(box.H+20),
				8, ColorSurface)
		case "td":
			vector.DrawFilledRect(screen,
				float32(box.X+offsetX), float32(absY),
				1, float32(box.H),
				ColorBorder, false)
		case "tr":
			rowColor := ColorTableRow1
			if box.RowIndex%2 == 0 {
				rowColor = ColorTableRow2
			}
			vector.DrawFilledRect(screen,
				float32(offsetX), float32(absY),
				float32(box.W), float32(box.H+8),
				rowColor, false)
		case "hr":
			vector.DrawFilledRect(screen,
				float32(offsetX), float32(absY),
				float32(box.W), 2,
				ColorHR, false)
		case "input", "button", "select", "textarea":
			// Render form elements using tag handlers
			if handler := forms.GetHandler(box.Node.Tag); handler != nil {
				// Create a temporary box with absolute position for rendering
				tempBox := &layout.RenderBox{
					Node: box.Node,
					X:    box.X + offsetX,
					Y:    absY,
					W:    box.W,
					H:    box.H,
				}
				handler.Render(screen, tempBox, box.Node, a.FormState)
				// Don't render children of form elements - handler does that
				return
			}
		}
	}

	// Draw text
	if box.Text != "" {
		textColor := ColorText
		fontSize := box.FontSize
		if fontSize == 0 {
			fontSize = FontSizeBody
		}

		// Use CSS computed color if available
		if box.TextColor != nil {
			textColor = *box.TextColor
		} else {
			if box.IsH1 {
				textColor = ColorAccent
			}
			if box.IsLink && !box.IsButton {
				textColor = ColorLink
			}
		}

		// Draw background if set from CSS
		if box.BgColor != nil && box.BgColor.A > 0 {
			vector.DrawFilledRect(screen,
				float32(box.X+offsetX), float32(absY),
				float32(box.W), float32(box.H),
				*box.BgColor, false)
		}

		// Draw button background
		if box.IsButton {
			btnPadX := float32(12)
			btnPadY := float32(6)
			btnW := float32(box.W) + btnPadX*2
			btnH := float32(fontSize) + btnPadY*2
			btnX := float32(box.X+offsetX) - btnPadX
			btnY := float32(absY) - btnPadY + 4

			render.DrawRoundedRect(screen, btnX, btnY, btnW, btnH, 6, ColorButtonPrimary)
			textColor = ColorButtonText
		}

		if absY > NavBarHeight-30 && absY < WindowHeight+30 {
			// Calculate text X position based on text-align
			textX := box.X + offsetX

			if box.TextAlign == "center" {
				// Estimate text width and center it
				textWidth := float64(len(box.Text)) * fontSize * 0.55
				textX = offsetX + (WindowWidth-Padding*2-textWidth)/2
			} else if box.TextAlign == "right" {
				textWidth := float64(len(box.Text)) * fontSize * 0.55
				textX = offsetX + WindowWidth - Padding*2 - textWidth
			}

			render.DrawText(screen, box.Text, textX, absY+fontSize, fontSize, textColor)
		}
	}

	// Draw images
	if box.IsImage && box.ImageURL != "" {
		imgX := float32(box.X + offsetX)
		imgY := float32(absY)
		imgW := float32(box.W)
		imgH := float32(box.H)

		img, loaded, failed := render.Cache.Get(box.ImageURL)

		if loaded && img != nil {
			bounds := img.Bounds()
			scaleX := float64(imgW) / float64(bounds.Dx())
			scaleY := float64(imgH) / float64(bounds.Dy())
			scale := scaleX
			if scaleY < scaleX {
				scale = scaleY
			}

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(float64(imgX), float64(imgY))
			screen.DrawImage(img, op)
		} else if failed {
			vector.DrawFilledRect(screen, imgX, imgY, imgW, imgH, ColorImageBg, false)
			render.DrawTextCentered(screen, "✕", float64(imgX+imgW/2), float64(imgY+imgH/2+8), 24, color.RGBA{255, 80, 80, 255})
		} else {
			vector.DrawFilledRect(screen, imgX, imgY, imgW, imgH, ColorImageBg, false)
			render.DrawTextCentered(screen, "◌", float64(imgX+imgW/2), float64(imgY+imgH/2+8), 24, ColorTextMuted)
			render.LoadImageAsync(box.ImageURL, render.CurrentBaseURL)
		}
	}

	// Render children
	for _, child := range box.Children {
		a.renderNode(screen, child, offsetX, offsetY)
	}
}

// HandleClick handles click on URL bar
func (n *NavBar) HandleClick(app *App, mx, my int) {
	if float32(mx) >= n.URLBarX && float32(mx) <= n.URLBarX+n.URLBarW &&
		float32(my) >= n.URLBarY && float32(my) <= n.URLBarY+URLBarHeight {
		if !n.IsEditing {
			n.IsEditing = true
			n.URLText = app.URL
			n.CursorPos = len(n.URLText)
		}
	} else if float32(my) < NavBarHeight {
		// Button positions matching Draw function
		btnSize := float32(30)
		btnSpacing := float32(6)
		startX := float32(12)
		btnY := float32((NavBarHeight - btnSize) / 2)

		// Back button
		if float32(mx) >= startX && float32(mx) <= startX+btnSize &&
			float32(my) >= btnY && float32(my) <= btnY+btnSize {
			if app.HistoryPos > 0 {
				app.HistoryPos--
				app.URL = app.History[app.HistoryPos]
				app.LoadFromURL(app.URL)
			}
			return
		}

		// Forward button
		fwdX := startX + btnSize + btnSpacing
		if float32(mx) >= fwdX && float32(mx) <= fwdX+btnSize &&
			float32(my) >= btnY && float32(my) <= btnY+btnSize {
			if app.HistoryPos < len(app.History)-1 {
				app.HistoryPos++
				app.URL = app.History[app.HistoryPos]
				app.LoadFromURL(app.URL)
			}
			return
		}

		// Refresh button
		refreshX := fwdX + btnSize + btnSpacing
		if float32(mx) >= refreshX && float32(mx) <= refreshX+btnSize &&
			float32(my) >= btnY && float32(my) <= btnY+btnSize {
			app.LoadFromURL(app.URL)
			return
		}

		// Capture button
		captureX := refreshX + btnSize + btnSpacing
		if float32(mx) >= captureX && float32(mx) <= captureX+btnSize &&
			float32(my) >= btnY && float32(my) <= btnY+btnSize {
			app.captureScreenshot = true
			return
		}
	} else {
		n.IsEditing = false
	}
}

// HandleInput handles keyboard input for URL bar
func (n *NavBar) HandleInput(app *App) {
	if !n.IsEditing {
		return
	}

	n.CursorBlink++

	runes := ebiten.AppendInputChars(nil)
	for _, r := range runes {
		n.URLText = n.URLText[:n.CursorPos] + string(r) + n.URLText[n.CursorPos:]
		n.CursorPos++
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if n.CursorPos > 0 {
			_, size := utf8.DecodeLastRuneInString(n.URLText[:n.CursorPos])
			n.URLText = n.URLText[:n.CursorPos-size] + n.URLText[n.CursorPos:]
			n.CursorPos -= size
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && n.CursorPos > 0 {
		_, size := utf8.DecodeLastRuneInString(n.URLText[:n.CursorPos])
		n.CursorPos -= size
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) && n.CursorPos < len(n.URLText) {
		_, size := utf8.DecodeRuneInString(n.URLText[n.CursorPos:])
		n.CursorPos += size
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		n.IsEditing = false
		url := strings.TrimSpace(n.URLText)
		if url != "" {
			if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
				url = "https://" + url
			}
			app.URL = url
			app.LoadFromURL(url)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		n.IsEditing = false
		n.URLText = app.URL
	}
}

// Draw renders the navigation bar with high contrast Safari-style design
func (n *NavBar) Draw(screen *ebiten.Image, app *App) {
	// Draw navbar background - dark gray
	navBg := color.RGBA{38, 38, 42, 255}
	vector.DrawFilledRect(screen, 0, 0, WindowWidth, NavBarHeight, navBg, false)

	// Navigation buttons - pill style with clear visibility
	btnSize := float32(30)
	btnY := float32((NavBarHeight - btnSize) / 2)
	btnSpacing := float32(6)
	startX := float32(12)

	// Brighter button background for visibility
	btnColor := color.RGBA{75, 75, 85, 255}
	btnTextColor := color.RGBA{220, 220, 225, 255}

	// Back button - centered text
	render.DrawRoundedRect(screen, startX, btnY, btnSize, btnSize, 6, btnColor)
	btnCenterX := float64(startX) + float64(btnSize)/2
	btnCenterY := float64(btnY) + float64(btnSize)/2 + 2
	render.DrawTextCentered(screen, "<", btnCenterX, btnCenterY, 18, btnTextColor)

	// Forward button - centered text
	startX += btnSize + btnSpacing
	render.DrawRoundedRect(screen, startX, btnY, btnSize, btnSize, 6, btnColor)
	btnCenterX = float64(startX) + float64(btnSize)/2
	render.DrawTextCentered(screen, ">", btnCenterX, btnCenterY, 18, btnTextColor)

	// Refresh button - centered text
	startX += btnSize + btnSpacing
	render.DrawRoundedRect(screen, startX, btnY, btnSize, btnSize, 6, btnColor)
	refreshIcon := "R"
	if app.IsLoading {
		refreshIcon = "..."
	}
	btnCenterX = float64(startX) + float64(btnSize)/2
	render.DrawTextCentered(screen, refreshIcon, btnCenterX, btnCenterY, 14, btnTextColor)

	// Capture/Screenshot button
	startX += btnSize + btnSpacing
	captureColor := btnColor
	if app.captureScreenshot {
		captureColor = color.RGBA{100, 150, 200, 255} // Highlight when capturing
	}
	render.DrawRoundedRect(screen, startX, btnY, btnSize, btnSize, 6, captureColor)
	btnCenterX = float64(startX) + float64(btnSize)/2
	render.DrawTextCentered(screen, "C", btnCenterX, btnCenterY, 14, btnTextColor)

	// URL Bar - lighter background for contrast
	urlBarMargin := float32(12)
	n.URLBarX = startX + btnSize + urlBarMargin
	n.URLBarW = float32(WindowWidth) - n.URLBarX - 12
	n.URLBarY = float32((NavBarHeight - URLBarHeight) / 2)

	// Much lighter URL bar for text readability
	urlBarColor := color.RGBA{240, 240, 245, 255} // Light gray/white
	if n.IsEditing {
		urlBarColor = color.RGBA{255, 255, 255, 255} // Pure white when editing
	} else if n.IsHovering {
		urlBarColor = color.RGBA{248, 248, 252, 255}
	}

	// Draw pill-shaped URL bar with rounded corners
	render.DrawRoundedRect(screen, n.URLBarX, n.URLBarY, n.URLBarW, URLBarHeight, 8, urlBarColor)

	// URL text - dark text on light background
	displayURL := app.URL
	if n.IsEditing {
		displayURL = n.URLText
	}

	// Truncate URL for display
	maxChars := int((n.URLBarW - 30) / 8)
	if len(displayURL) > maxChars && maxChars > 0 {
		displayURL = displayURL[:maxChars] + "…"
	}

	// Dark text for contrast - vertically centered in URL bar
	urlTextColor := color.RGBA{50, 50, 55, 255}
	textY := float64(n.URLBarY) + float64(URLBarHeight)/2 - 2 // Align with cursor
	render.DrawText(screen, displayURL, float64(n.URLBarX+12), textY, FontSizeUI, urlTextColor)

	// Blinking cursor when editing
	if n.IsEditing && (n.CursorBlink/30)%2 == 0 {
		// Measure actual text width up to cursor position for accurate placement
		textBeforeCursor := n.URLText[:n.CursorPos]
		cursorOffset := render.MeasureText(textBeforeCursor, FontSizeUI)

		maxOffset := float64(n.URLBarW - 24)
		if cursorOffset > maxOffset {
			cursorOffset = maxOffset
		}
		cursorX := float32(float64(n.URLBarX) + 12 + cursorOffset)
		cursorY := n.URLBarY + 10
		cursorH := float32(URLBarHeight - 20)
		cursorColor := color.RGBA{30, 120, 210, 255} // Blue cursor
		vector.DrawFilledRect(screen, cursorX, cursorY, 2, cursorH, cursorColor, false)
	}
}

// initJSEngine initializes SpiderGopher and executes <script> tags
func (a *App) initJSEngine() {
	if a.DOMRoot == nil {
		return
	}

	// Create new engine for each page load
	a.JSEngine = spidergopher.NewEngine()

	// Connect to the real DOM
	a.JSEngine.SetDOM(a.DOMRoot)

	// Start the event loop for async operations (setTimeout, fetch, etc.)
	a.JSEngine.Start()

	// Extract and execute all <script> tags
	scripts := extractScripts(a.DOMRoot)
	fmt.Printf("[initJSEngine] Found %d script(s) to execute\n", len(scripts))
	for i, script := range scripts {
		if script != "" {
			fmt.Printf("[initJSEngine] Executing script #%d (%d chars)\n", i+1, len(script))
			_, err := a.JSEngine.Run(script)
			if err != nil {
				fmt.Printf("[JS Error] %v\n", err)
			}
		}
	}

	// IMPORTANT: Rebuild render tree AFTER JS execution
	// This ensures DOM modifications made by JS are visible
	a.RenderTree = layout.BuildRenderTree(a.DOMRoot, WindowWidth-(Padding*2))
}

// extractScripts extracts script content from <script> tags in the DOM
func extractScripts(node *dom.Node) []string {
	var scripts []string
	if node == nil {
		return scripts
	}

	if node.Tag == "script" {
		// Get text content from script tag
		for _, child := range node.Children {
			if child.Type == dom.NodeText && child.Content != "" {
				scripts = append(scripts, child.Content)
			}
		}
	}

	// Recursively search children
	for _, child := range node.Children {
		scripts = append(scripts, extractScripts(child)...)
	}

	return scripts
}

// dispatchJSClickEvent fires click event listeners registered via JavaScript
func (a *App) dispatchJSClickEvent(node *dom.Node) {
	if a.JSEngine == nil || node == nil {
		return
	}

	// Import the JSNode package to access event listeners
	// Get listeners for this specific node
	spiderdom.DispatchClickEvent(node, a.JSEngine.GetVM())

	// Rebuild render tree to reflect any DOM changes made by the handler
	a.RenderTree = layout.BuildRenderTree(a.DOMRoot, WindowWidth-(Padding*2))
}
