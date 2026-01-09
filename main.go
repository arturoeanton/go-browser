package main

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go-browser/browser"
	"go-browser/render"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed fonts/Inter-Regular.ttf
var fontData []byte

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		log.Fatal("Error loading font:", err)
	}
	render.SetFontSource(src)
}

func main() {
	ebiten.SetWindowSize(browser.WindowWidth, browser.WindowHeight)
	ebiten.SetWindowTitle("GoBrowser")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	app := browser.NewApp()

	// Load initial URL or default
	if len(os.Args) > 1 {
		url := os.Args[1]

		// If it's not a URL (no protocol), treat as file path
		if !strings.HasPrefix(url, "http://") &&
			!strings.HasPrefix(url, "https://") &&
			!strings.HasPrefix(url, "file://") {
			// Convert to absolute path for file:// protocol
			absPath, err := filepath.Abs(url)
			if err == nil {
				url = "file://" + absPath
			}
		}

		app.URL = url
		app.LoadFromURL(url)
	} else {
		app.LoadFromURL("https://example.com")
	}

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
