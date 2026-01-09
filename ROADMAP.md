# 🗺️ ROADMAP - GoBrowser

The project has two main engines inspired by Firefox:

| Engine | Inspired by | Responsibility | Technology |
|--------|------------|----------------|------------|
| **Gocko** | Gecko | HTML/CSS Rendering | Pure Go |
| **SpiderGopher** | SpiderMonkey | JavaScript Engine | Go + [goja](https://github.com/dop251/goja) |

---

## 🦎 Gocko (HTML/CSS Engine)

Aspiration: **100% compatible with Gecko**

### Phase 1: Core Foundation ✅
- [x] Basic HTML Parser
- [x] CSS Parser (inline, `<style>`, external)
- [x] DOM Tree
- [x] Basic Box Model
- [x] Initial Layout Engine

### Phase 2: Complete Forms 🔨 (In Progress)
- [x] `<input type="text/password/email">`
- [x] `<input type="checkbox/radio">`
- [x] `<select>` dropdown with overlay
- [x] `<textarea>` multiline
- [x] `<button>` types
- [x] Tab navigation between elements
- [ ] `<input type="number">` with functional spinners
- [ ] `<input type="date/time">` pickers
- [ ] `<input type="range">` interactive slider
- [ ] `<input type="color">` color picker
- [ ] `<input type="file">` file dialog
- [ ] Form submission (GET/POST)
- [ ] Form validation (required, pattern, min/max)

### Phase 3: Advanced Layout
- [ ] Complete Flexbox (align, justify, wrap)
- [ ] Complete CSS Grid
- [ ] Position (relative, absolute, fixed, sticky)
- [ ] Float and clear
- [ ] Overflow and scroll
- [ ] Z-index stacking

### Phase 4: Advanced CSS
- [ ] Media queries
- [ ] CSS Variables (custom properties)
- [ ] Basic transitions
- [ ] Animations (@keyframes)
- [ ] Transforms (translate, rotate, scale)
- [ ] Pseudo-classes (:hover, :focus, :active)
- [ ] Pseudo-elements (::before, ::after)

### Phase 5: Advanced Rendering
- [ ] Web fonts (@font-face)
- [ ] SVG rendering
- [ ] Canvas 2D
- [ ] Gradients (linear, radial)
- [ ] Box shadows
- [ ] Complete border radius
- [ ] Filters (blur, grayscale, etc.)

---

## 🕷️ SpiderGopher (JavaScript Engine)

Aspiration: **100% compatible with SpiderMonkey**

Implemented on [goja](https://github.com/dop251/goja) - an ES5.1+ interpreter written in pure Go.

### Phase 1: Core Integration
- [ ] Integrate goja with browser
- [ ] Execute `<script>` inline
- [ ] Load `<script src="...">` external
- [ ] Console (log, warn, error)

### Phase 2: Basic DOM API
- [ ] `document.getElementById()`
- [ ] `document.querySelector/querySelectorAll()`
- [ ] `document.createElement()`
- [ ] `element.appendChild/removeChild()`
- [ ] `element.innerHTML/textContent`
- [ ] `element.setAttribute/getAttribute()`
- [ ] `element.classList` (add, remove, toggle)
- [ ] `element.style` property access

### Phase 3: Events
- [ ] `addEventListener/removeEventListener`
- [ ] Click events
- [ ] Keyboard events (keydown, keyup)
- [ ] Focus events (focus, blur)
- [ ] Input events (input, change)
- [ ] Submit events
- [ ] Event bubbling and capturing

### Phase 4: Timers & Async
- [ ] `setTimeout/clearTimeout`
- [ ] `setInterval/clearInterval`
- [ ] `requestAnimationFrame`
- [ ] Promises (goja supports ES6)
- [ ] Basic `fetch()` API

### Phase 5: Storage & APIs
- [ ] `localStorage/sessionStorage`
- [ ] `location` object
- [ ] `history` API (pushState, popState)
- [ ] `navigator` object
- [ ] Clipboard API

### Phase 6: Advanced (Long-term)
- [ ] XMLHttpRequest
- [ ] FormData
- [ ] WebSocket
- [ ] Worker (basic Web Workers)

---

## 📊 Target Compatibility

| Feature | Gecko | Gocko Target | SpiderMonkey | SpiderGopher Target |
|---------|-------|--------------|--------------|---------------------|
| HTML5 | 100% | 80% | - | - |
| CSS3 | 100% | 60% | - | - |
| ES6+ | - | - | 100% | 70% (via goja) |
| DOM Level 3 | 100% | 50% | 100% | 50% |
| CSSOM | 100% | 30% | 100% | 20% |

---

## 🚀 Milestones

### v0.1.0 - Forms MVP ✅
- Interactive forms working
- Basic navigation
- CSS inline/style tags

### v0.2.0 - Basic JavaScript (Next)
- SpiderGopher integrated
- Basic DOM manipulation
- Event listeners

### v0.3.0 - Complete Layout
- Functional Flexbox
- Basic Grid
- Positioning

### v0.4.0 - Interactivity
- Animations/Transitions
- Fetch API
- localStorage

### v1.0.0 - Usable Browser
- Can render simple web pages
- Functional JavaScript
- Forms with submission

---

## 🛠️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                    GoBrowser                         │
├─────────────────────────────────────────────────────┤
│  ┌─────────────────┐     ┌─────────────────────┐   │
│  │     Gocko       │     │   SpiderGopher      │   │
│  │   (HTML/CSS)    │◄───►│   (JavaScript)      │   │
│  │                 │     │   [goja runtime]    │   │
│  └────────┬────────┘     └──────────┬──────────┘   │
│           │                         │               │
│           ▼                         ▼               │
│  ┌─────────────────────────────────────────────┐   │
│  │              Shared DOM Tree                 │   │
│  └─────────────────────────────────────────────┘   │
│           │                                         │
│           ▼                                         │
│  ┌─────────────────────────────────────────────┐   │
│  │         Ebiten (Graphics/Input)              │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

---

*Experimental project with 💚 Go and 🤖 Opus*
