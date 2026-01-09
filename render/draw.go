// Package render provides drawing and rendering functions
package render

import (
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// FontSource holds the loaded font
var FontSource *text.GoTextFaceSource

// SetFontSource sets the font source for text rendering
func SetFontSource(src *text.GoTextFaceSource) {
	FontSource = src
}

// DrawRoundedRect draws a filled rectangle
func DrawRoundedRect(screen *ebiten.Image, x, y, w, h, radius float32, clr color.Color) {
	vector.DrawFilledRect(screen, x, y, w, h, clr, false)
}

// DrawText draws text at the specified position
func DrawText(screen *ebiten.Image, txt string, x, y float64, size float64, clr color.Color) {
	if FontSource == nil {
		return
	}
	face := &text.GoTextFace{
		Source: FontSource,
		Size:   size,
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, txt, face, op)
}

// DrawTextCentered draws text centered at the specified position
func DrawTextCentered(screen *ebiten.Image, txt string, x, y float64, size float64, clr color.Color) {
	if FontSource == nil {
		return
	}
	face := &text.GoTextFace{
		Source: FontSource,
		Size:   size,
	}
	// Measure text width for centering
	w, _ := text.Measure(txt, face, 0)
	op := &text.DrawOptions{}
	op.GeoM.Translate(x-w/2, y)
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, txt, face, op)
}

// MeasureText returns the width of text at a given font size
func MeasureText(txt string, size float64) float64 {
	if FontSource == nil {
		return float64(len(txt)) * size * 0.6 // Fallback
	}
	face := &text.GoTextFace{
		Source: FontSource,
		Size:   size,
	}
	w, _ := text.Measure(txt, face, 0)
	return w
}

// ======================================================================================
// IMAGE CACHE
// ======================================================================================

// ImageCache stores loaded images
type ImageCache struct {
	images  map[string]*ebiten.Image
	loading map[string]bool
	failed  map[string]bool
	mutex   sync.RWMutex
}

// Cache is the global image cache
var Cache = &ImageCache{
	images:  make(map[string]*ebiten.Image),
	loading: make(map[string]bool),
	failed:  make(map[string]bool),
}

// Get returns a cached image and its loading/failed status
func (c *ImageCache) Get(imgURL string) (*ebiten.Image, bool, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if img, ok := c.images[imgURL]; ok {
		return img, true, false
	}
	if c.failed[imgURL] {
		return nil, false, true
	}
	return nil, c.loading[imgURL], false
}

// StartLoading marks an image as loading
func (c *ImageCache) StartLoading(imgURL string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.loading[imgURL] || c.images[imgURL] != nil || c.failed[imgURL] {
		return false
	}
	c.loading[imgURL] = true
	return true
}

// SetImage stores a loaded image in the cache
func (c *ImageCache) SetImage(imgURL string, img *ebiten.Image) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.images[imgURL] = img
	delete(c.loading, imgURL)
}

// SetFailed marks an image as failed to load
func (c *ImageCache) SetFailed(imgURL string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.failed[imgURL] = true
	delete(c.loading, imgURL)
}

// CurrentBaseURL tracks the current page URL for relative image resolution
var CurrentBaseURL string

// LoadImageAsync loads an image asynchronously
func LoadImageAsync(imgURL string, baseURL string) {
	if !Cache.StartLoading(imgURL) {
		return
	}

	go func() {
		// Resolve relative URLs
		fullURL := imgURL
		if !strings.HasPrefix(imgURL, "http") && baseURL != "" {
			if base, err := url.Parse(baseURL); err == nil {
				if ref, err := url.Parse(imgURL); err == nil {
					fullURL = base.ResolveReference(ref).String()
				}
			}
		}

		resp, err := http.Get(fullURL)
		if err != nil {
			Cache.SetFailed(imgURL)
			return
		}
		defer resp.Body.Close()

		img, _, err := image.Decode(resp.Body)
		if err != nil {
			Cache.SetFailed(imgURL)
			return
		}

		ebitenImg := ebiten.NewImageFromImage(img)
		Cache.SetImage(imgURL, ebitenImg)
	}()
}

// ======================================================================================
// GRADIENT RENDERING
// ======================================================================================

// GradientStop for rendering
type GradientStop struct {
	R, G, B, A float64
	Position   float64
}

// DrawLinearGradient draws a linear gradient on the screen
func DrawLinearGradient(screen *ebiten.Image, x, y, w, h float32, angle float64, stops []GradientStop) {
	if len(stops) < 2 {
		return
	}

	// Create a temporary image for the gradient
	gradImg := ebiten.NewImage(int(w), int(h))

	// For simplicity, we'll render horizontal or vertical gradients
	// Convert angle to radians and determine direction
	// 0deg = to top, 90deg = to right, 180deg = to bottom, 270deg = to left

	for py := 0; py < int(h); py++ {
		for px := 0; px < int(w); px++ {
			// Calculate position along gradient axis (0.0 to 1.0)
			var t float64
			switch {
			case angle == 0 || angle == 360:
				// To top
				t = 1.0 - float64(py)/float64(h)
			case angle == 90:
				// To right
				t = float64(px) / float64(w)
			case angle == 180:
				// To bottom
				t = float64(py) / float64(h)
			case angle == 270:
				// To left
				t = 1.0 - float64(px)/float64(w)
			case angle == 135:
				// To bottom-right (diagonal)
				t = (float64(px)/float64(w) + float64(py)/float64(h)) / 2.0
			case angle == 315:
				// To top-left (diagonal)
				t = 1.0 - (float64(px)/float64(w)+float64(py)/float64(h))/2.0
			default:
				// Default: treat as to bottom
				t = float64(py) / float64(h)
			}

			// Interpolate color
			c := interpolateColor(stops, t)
			gradImg.Set(px, py, c)
		}
	}

	// Draw the gradient image onto the screen
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(gradImg, op)
}

// interpolateColor finds the right color for position t (0.0 to 1.0)
func interpolateColor(stops []GradientStop, t float64) color.RGBA {
	if t <= stops[0].Position {
		return color.RGBA{
			R: uint8(stops[0].R),
			G: uint8(stops[0].G),
			B: uint8(stops[0].B),
			A: uint8(stops[0].A),
		}
	}
	if t >= stops[len(stops)-1].Position {
		last := stops[len(stops)-1]
		return color.RGBA{
			R: uint8(last.R),
			G: uint8(last.G),
			B: uint8(last.B),
			A: uint8(last.A),
		}
	}

	// Find the two stops we're between
	for i := 0; i < len(stops)-1; i++ {
		if t >= stops[i].Position && t <= stops[i+1].Position {
			// Interpolate between stops[i] and stops[i+1]
			range_ := stops[i+1].Position - stops[i].Position
			if range_ == 0 {
				range_ = 0.001
			}
			localT := (t - stops[i].Position) / range_

			return color.RGBA{
				R: uint8(stops[i].R + (stops[i+1].R-stops[i].R)*localT),
				G: uint8(stops[i].G + (stops[i+1].G-stops[i].G)*localT),
				B: uint8(stops[i].B + (stops[i+1].B-stops[i].B)*localT),
				A: uint8(stops[i].A + (stops[i+1].A-stops[i].A)*localT),
			}
		}
	}

	return color.RGBA{0, 0, 0, 255}
}
