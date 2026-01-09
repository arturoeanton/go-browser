// Package paint provides rendering functions for Gocko
package paint

import (
	"image/color"

	"go-browser/gocko/box"
	"go-browser/gocko/forms"
	"go-browser/render"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Colors
var (
	ColorText      = color.RGBA{33, 33, 33, 255}
	ColorLink      = color.RGBA{25, 118, 210, 255}
	ColorBorder    = color.RGBA{200, 200, 210, 255}
	ColorHR        = color.RGBA{180, 180, 190, 255}
	ColorImageBg   = color.RGBA{230, 230, 235, 255}
	ColorTextMuted = color.RGBA{100, 100, 110, 255}
)

// PaintTree renders the entire layout tree
func PaintTree(screen *ebiten.Image, root *box.Box, offsetX, offsetY float64, state *forms.FormState) {
	paintBox(screen, root, offsetX, offsetY, state)
}

func paintBox(screen *ebiten.Image, b *box.Box, offsetX, offsetY float64, state *forms.FormState) {
	if b == nil {
		return
	}

	x := b.X + offsetX
	y := b.Y + offsetY

	// Paint form component if present
	if b.FormComponent != nil {
		b.FormComponent.Render(screen, b, state)
		// Paint children
		for _, child := range b.Children {
			paintBox(screen, child, offsetX, offsetY, state)
		}
		return
	}

	// Paint HR
	if b.Node != nil && b.Node.Tag == "hr" {
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(b.Width), float32(b.Height), ColorHR, false)
	}

	// Paint image
	if b.IsImage && b.ImageURL != "" {
		paintImage(screen, b, x, y)
	}

	// Paint text
	if b.Text != "" {
		paintText(screen, b, x, y)
	}

	// Paint children
	for _, child := range b.Children {
		paintBox(screen, child, offsetX, offsetY, state)
	}
}

func paintText(screen *ebiten.Image, b *box.Box, x, y float64) {
	textColor := ColorText
	fontSize := b.FontSize
	if fontSize == 0 {
		fontSize = 15
	}

	// Links are blue
	if b.IsLink {
		textColor = ColorLink
	}

	render.DrawText(screen, b.Text, x, y+fontSize, fontSize, textColor)
}

func paintImage(screen *ebiten.Image, b *box.Box, x, y float64) {
	imgX := float32(x)
	imgY := float32(y)
	imgW := float32(b.Width)
	imgH := float32(b.Height)

	img, loaded, failed := render.Cache.Get(b.ImageURL)

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
		render.LoadImageAsync(b.ImageURL, render.CurrentBaseURL)
	}
}
