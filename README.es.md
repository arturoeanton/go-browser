# 🌐 GoBrowser

> **⚠️ DISCLAIMER IMPORTANTE**
> 
> Este es un **proyecto experimental y educativo** creado para explorar si es posible construir un navegador web desde cero usando **únicamente Go** y sin depender de librerías de rendering web existentes (como WebKit, Blink, Gecko, etc.).
> 
> **No es un navegador funcional para uso diario.** Es un ejercicio de aprendizaje sobre cómo funcionan los navegadores internamente: parsing HTML, CSS, layout engine, rendering, etc.
> 
> 🤖 **Desarrollado con mucho Opus** (Claude) como pair-programming assistant.

🌍 [Read in English](README.md)

---

## 🎯 ¿Qué es esto?

GoBrowser es un intento minimalista de implementar los componentes fundamentales de un navegador web:

- **Parser HTML** → Convierte HTML en un árbol DOM
- **Parser CSS** → Parsea stylesheets y calcula estilos
- **Layout Engine** → Posiciona elementos en la pantalla
- **Renderer** → Dibuja píxeles usando Ebiten

## 📦 Arquitectura

El proyecto tiene dos motores principales inspirados en Firefox:

| Motor | Inspirado en | Responsabilidad |
|-------|-------------|-----------------|
| **Gocko** | Gecko | HTML/CSS Rendering |
| **SpiderGopher** | SpiderMonkey | JavaScript (via [goja](https://github.com/dop251/goja)) |

```
go-browser/
├── gocko/           # 🦎 Motor de rendering (HTML/CSS)
│   ├── engine.go    # Coordinador principal
│   ├── box/         # CSS Box Model
│   ├── layout/      # Layout engine
│   ├── paint/       # Rendering
│   └── forms/       # Componentes de formularios
├── browser/         # App shell, NavBar, eventos
├── css/             # Parser CSS, cascade, selectores
├── dom/             # Parser HTML, nodos DOM
├── render/          # Utilidades de dibujo
├── fonts/           # Fuentes embebidas
└── demos/           # Páginas HTML de prueba
```

Ver [ROADMAP.md](ROADMAP.md) para las fases de desarrollo.

## 🚀 Cómo Ejecutar

```bash
# Clonar
git clone https://github.com/arturoeanton/go-browser.git
cd go-browser

# Ejecutar
go run main.go

# O cargar un archivo local
go run main.go demos/09_forms.html

# O una URL
go run main.go https://example.com
```

## ✨ Características Implementadas

| Feature | Estado |
|---------|--------|
| Parser HTML básico | ✅ |
| Parser CSS inline y `<style>` | ✅ |
| Selectores CSS (tag, class, id) | ✅ |
| Layout de bloques | ✅ |
| Flexbox básico | ✅ |
| Navegación (Back/Forward/Refresh) | ✅ |
| Barra de URL editable | ✅ |
| Links clickeables | ✅ |
| Imágenes (async loading) | ✅ |
| Tablas | ✅ |
| Elementos de formulario | 🔨 En progreso |
| Navegación con Tab | ✅ |
| Form submission | 📋 Planeado |
| Clipboard (copy/paste) | 📋 Planeado |
| JavaScript (SpiderGopher) | 📋 Planeado |

## 🛠️ Dependencias

Solo usamos dependencias mínimas para gráficos y fuentes:

- [**ebiten/v2**](https://github.com/hajimehoshi/ebiten) - Game engine 2D para rendering
- [**golang.org/x/net/html**](https://pkg.go.dev/golang.org/x/net/html) - Tokenizer HTML

**No usamos:** WebKit, Blink, Gecko, CEF, WebView, ni ningún engine de browser existente.

## 📸 Screenshots

*El navegador cargando example.com y demos interactivos*

## 🎓 Propósito Educativo

Este proyecto existe para responder preguntas como:
- ¿Cómo funciona el parsing HTML?
- ¿Cómo se calcula la cascada CSS?
- ¿Qué es un layout engine y cómo posiciona elementos?
- ¿Cómo se renderizan píxeles en pantalla?

**No intentes usar esto para navegar la web real** - es un juguete educativo.

## 🤝 Contribuciones

¡Las contribuciones son bienvenidas! Este es un proyecto para aprender, así que cualquier mejora o nueva feature es útil.

## 📄 Licencia

Apache 2.0

---

*Hecho con 💚 Go y 🤖 Opus*
