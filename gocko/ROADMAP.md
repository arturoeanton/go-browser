# Gocko Engine - Roadmap

> A professional CSS rendering engine inspired by Gecko (Firefox) and WebKit (Safari)

---

## Current Status: Phase 1 ✅

### Completed
- ✅ Form handlers migrated to `gocko/forms/`
- ✅ CSS Value System (`gocko/css/values/`)
  - Length: 14 units (px, em, rem, %, vw, vh, pt, cm, mm, in, ch, ex, vmin, vmax)
  - Color: 140+ named colors, hex, rgb(), rgba(), hsl(), hsla()
  - ComputedStyle: 40+ CSS properties with proper defaults
- ✅ Property Parser (`gocko/css/properties/`)
  - All shorthand expansions (margin, padding, border, flex, background)
- ✅ CSS Package Export (`gocko/css/css.go`)
  - Convenient re-exports of all types and functions

---

## Roadmap

### v0.2 - Box Model & Layout ✅ (Completed)
- [x] Box model calculations (content-box, border-box)
- [x] Complete flexbox implementation
  - flex-direction, flex-wrap, flex-flow
  - justify-content, align-items, align-content
  - flex-grow, flex-shrink, flex-basis
  - gap, order
- [ ] Positioned elements (absolute, fixed, sticky)
- [ ] Overflow handling (scroll, auto, hidden)

### v0.3 - Typography (Q1 2026)
- [ ] Font loading system (local + web fonts)
- [ ] font-weight rendering (regular, bold, light)
- [ ] font-style (italic, oblique)
- [ ] Complete text properties
  - text-align, text-decoration, text-transform
  - letter-spacing, word-spacing
  - white-space handling

### v0.4 - Visual Effects (Q2 2026)
- [ ] border-radius (all corners)
- [ ] box-shadow (multiple shadows)
- [ ] opacity and visibility
- [ ] Basic transforms (translate, rotate, scale)
- [ ] Background images and gradients

### v0.5 - Grid Layout (Q2 2026)
- [ ] CSS Grid basic implementation
- [ ] grid-template-columns/rows
- [ ] grid-gap
- [ ] grid-area, grid-column, grid-row

### v1.0 - Production Ready (Q3 2026)
- [ ] Full CSS 2.1 compliance
- [ ] CSS 3 core modules
- [ ] Performance optimization
- [ ] Memory efficiency
- [ ] Test suite with visual regression

---

## Architecture

```
gocko/
├── css/                    # CSS Engine
│   ├── values/             # Value types (Length, Color, ComputedStyle)
│   ├── properties/         # Property parsers
│   ├── cascade/            # Cascade algorithm (future)
│   └── parser/             # CSS tokenizer (future)
├── layout/                 # Layout algorithms
├── paint/                  # Rendering pipeline
├── box/                    # Box model
└── forms/                  # Form element handlers
```

---

## Design Principles

1. **Modularity** - Each component is self-contained
2. **Performance** - Avoid allocations in hot paths
3. **Correctness** - Follow W3C specifications
4. **Testability** - Every module is unit testable
5. **Documentation** - Inline docs with spec references
