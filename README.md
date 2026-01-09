# 🌐 GoBrowser

> **⚠️ IMPORTANT DISCLAIMER**
> 
> This is an **experimental and educational project** created to explore whether it's possible to build a web browser from scratch using **only Go** without relying on existing web rendering libraries (like WebKit, Blink, Gecko, etc.).
> 
> **This is NOT a functional browser for daily use.** It's a learning exercise about how browsers work internally: HTML parsing, CSS, layout engines, rendering, etc.
> 
> 🤖 **Developed with lots of Opus** (Claude) as pair-programming assistant.

🌍 [Leer en Español](README.es.md)

---

## 🎯 What is this?

GoBrowser is a minimalist attempt to implement the fundamental components of a web browser:

- **HTML Parser** → Converts HTML into a DOM tree
- **CSS Parser** → Parses stylesheets and computes styles
- **Layout Engine** → Positions elements on the screen
- **Renderer** → Draws pixels using Ebiten

## 📦 Architecture

The project has two main engines inspired by Firefox:

| Engine | Inspired by | Responsibility |
|--------|------------|----------------|
| **Gocko** | Gecko | HTML/CSS Rendering |
| **SpiderGopher** | SpiderMonkey | JavaScript (via [goja](https://github.com/dop251/goja)) |

```
go-browser/
├── gocko/           # 🦎 Rendering engine (HTML/CSS)
│   ├── engine.go    # Main coordinator
│   ├── box/         # CSS Box Model
│   ├── layout/      # Layout engine
│   ├── paint/       # Rendering
│   └── forms/       # Form components
├── browser/         # App shell, NavBar, events
├── css/             # CSS Parser, cascade, selectors
├── dom/             # HTML Parser, DOM nodes
├── render/          # Drawing utilities
├── fonts/           # Embedded fonts
└── demos/           # Test HTML pages
```

See [ROADMAP.md](ROADMAP.md) for development phases.

## 🚀 How to Run

```bash
# Clone
git clone https://github.com/arturoeanton/go-browser.git
cd go-browser

# Run
go run main.go

# Or load a local file
go run main.go demos/09_forms.html

# Or a URL
go run main.go https://example.com
```

## ✨ Implemented Features

| Feature | Status |
|---------|--------|
| Basic HTML Parser | ✅ |
| Inline CSS and `<style>` parser | ✅ |
| CSS Selectors (tag, class, id) | ✅ |
| Block Layout | ✅ |
| Basic Flexbox | ✅ |
| Navigation (Back/Forward/Refresh) | ✅ |
| Editable URL bar | ✅ |
| Clickable links | ✅ |
| Images (async loading) | ✅ |
| Tables | ✅ |
| Form elements | 🔨 In progress |
| Tab navigation | ✅ |
| Form submission | 📋 Planned |
| Clipboard (copy/paste) | 📋 Planned |
| JavaScript (SpiderGopher) | 📋 Planned |

## 🛠️ Dependencies

We only use minimal dependencies for graphics and fonts:

- [**ebiten/v2**](https://github.com/hajimehoshi/ebiten) - 2D Game engine for rendering
- [**golang.org/x/net/html**](https://pkg.go.dev/golang.org/x/net/html) - HTML Tokenizer

**We DO NOT use:** WebKit, Blink, Gecko, CEF, WebView, or any existing browser engine.

## 📸 Screenshots

*The browser loading example.com and interactive demos*

## 🎓 Educational Purpose

This project exists to answer questions like:
- How does HTML parsing work?
- How is CSS cascade calculated?
- What is a layout engine and how does it position elements?
- How are pixels rendered to the screen?

**Don't try to use this to browse the real web** - it's an educational toy.

## 🤝 Contributing

Contributions are welcome! This is a learning project, so any improvement or new feature is helpful.

## 📄 License

Apache 2.0

---

*Made with 💚 Go and 🤖 Opus*
