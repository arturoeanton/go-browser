# HELP-dev.md - Guía de Desarrollo para Gocko/SpiderGopher

Este documento recopila insights de [thdwb](https://github.com/danfragoso/thdwb) y otros browsers en Go para ayudar a resolver problemas comunes en el desarrollo de Gocko.

---

## 📋 Arquitectura Comparativa

### THDWB (The Hotdog Web Browser)
```
ketchup  → HTML Parser / DOM Tree
mayo     → CSS Parser / Render Tree Builder
bun      → Layout Calculator
mustard  → UI Toolkit / Events / OpenGL
sauce    → Network / Cache
gg       → Drawing / Text Rendering
```

### Gocko (Nuestro Browser)
```
dom/         → HTML Parser / DOM Tree
css/         → CSS Parser / Styling
layout/      → Layout Calculator (RenderBox)
render/      → Drawing
browser/     → UI / Events (Ebiten)
spidergopher → JavaScript Engine (Goja)
```

---

## 🔧 Problemas Conocidos y Soluciones

### 1. DOM Modifications no se Reflejan en el Render

**Problema:** Cuando JavaScript modifica el DOM (e.g., `textContent = "nuevo"`), los cambios no aparecen en pantalla.

**Causa:** El render tree se construye UNA vez al cargar la página.

**Solución de thdwb:**
```go
// bun/bun.go - Re-renderiza desde cero
func RenderDocument(ctx *gg.Context, document *hotdog.Document, experimentalLayout bool) {
    renderTree := createRenderTree(html)  // Recrea el tree
    layoutNode(ctx, renderTree)           // Recalcula layout
    paintNode(ctx, renderTree)            // Repinta
}
```

**Implementación en Gocko:**
```go
// browser/app.go - Rebuild after JS changes
func (a *App) refreshRender() {
    a.RenderTree = layout.BuildRenderTree(a.DOMRoot, WindowWidth-(Padding*2))
}
```

**Cuándo llamar `refreshRender()`:**
1. Después de ejecutar scripts iniciales
2. Después de cada evento click que modifica DOM
3. Después de cualquier callback async (setTimeout, fetch)

---

### 2. Event Loop y Re-Render

**Problema:** Los cambios hechos en `setTimeout` o event handlers no se ven.

**Causa:** El EventLoop de SpiderGopher corre en un goroutine separado. Cuando modifica el DOM, nadie reconstruye el render tree.

**Solución Propuesta:**
```go
// En spidergopher/core/loop.go - Añadir callback de dirty flag
type EventLoop struct {
    // ...existing fields...
    OnDOMChange func()  // Callback cuando algo cambia
}

// En browser/app.go
engine.Loop.OnDOMChange = func() {
    a.RenderTree = layout.BuildRenderTree(a.DOMRoot, WindowWidth-(Padding*2))
}
```

---

### 3. Click Events no Funcionan

**Problema:** Los botones con `addEventListener("click", ...)` no responden.

**Causa:** 
1. Los event listeners se registran en un mapa global (`nodeEventListeners`)
2. El browser no sabe buscar en ese mapa cuando hay clicks
3. La conexión entre Gocko click → SpiderGopher callback no existe

**Solución (ya implementada parcialmente):**
```go
// browser/app.go
func (a *App) dispatchJSClickEvent(node *dom.Node) {
    spiderdom.DispatchClickEvent(node, a.JSEngine.GetVM())
    a.refreshRender()  // Re-render después del evento
}
```

**Problema Adicional:** El nodo que recibe el click en el RenderTree puede no ser el mismo objeto que el nodo donde se registró el listener.

**Solución:** Usar IDs para matching:
```go
// spidergopher/dom/jsnode.go
var nodeEventListenersByID = make(map[string]map[string][]goja.Callable)

func (n *JSNode) addEventListener(eventType string, callback goja.Callable) {
    nodeID := n.node.GetAttr("id")
    if nodeEventListenersByID[nodeID] == nil {
        nodeEventListenersByID[nodeID] = make(map[string][]goja.Callable)
    }
    nodeEventListenersByID[nodeID][eventType] = append(...)
}
```

---

### 4. textContent/innerHTML Setter

**Problema:** Asignar `element.textContent = "texto"` no funciona.

**Causa:** Goja's `DefineAccessorProperty` pasa el valor directamente al setter, NO via `call.Arguments`.

**Solución Correcta:**
```go
// ❌ INCORRECTO
obj.DefineAccessorProperty("textContent",
    getter,
    n.vm.ToValue(func(call goja.FunctionCall) goja.Value {
        value := call.Argument(0).String()  // NUNCA tiene valor!
    }),
    ...)

// ✅ CORRECTO  
obj.DefineAccessorProperty("textContent",
    getter,
    n.vm.ToValue(func(this goja.Value, value goja.Value) {
        text := value.String()  // Valor recibido correctamente
        n.setTextContent(text)
    }),
    ...)
```

---

### 5. Display: None Elements

**Insight de thdwb:** Filtrar elementos con `display: none` al crear render tree.

```go
// bun/layout.go
func createRenderTree(root *hotdog.NodeDOM) *hotdog.NodeDOM {
    if root.Style.Display == "none" {
        return nil  // No incluir en render tree
    }
    // ...continue building tree
}
```

**Para Gocko:** En `layout/builder.go`, verificar:
```go
if node.ComputedStyle != nil && node.ComputedStyle.Display == "none" {
    return nil
}
```

---

## 🎯 Checklist de Mejoras Prioritarias

### SpiderGopher (JavaScript)
- [ ] Callback para notificar cambios al DOM
- [ ] Matching de event listeners por ID, no por puntero
- [ ] Soporte para `className` setter (actualizar class attribute)
- [ ] Soporte para `style.` setters

### Gocko (Browser)
- [ ] Auto-refresh render tree cuando JS modifica DOM
- [ ] Disparar click events a todos los ancestros (bubbling)
- [ ] Implementar `requestAnimationFrame` para animaciones
- [ ] Cache de render tree (solo reconstruir nodos dirty)

### Layout
- [ ] Filtrar display:none en render tree
- [ ] Soporte para position: absolute/relative
- [ ] Mejor handling de inline elements

---

## 🔗 Referencias

- **thdwb:** https://github.com/danfragoso/thdwb
- **Goja (JS Engine):** https://github.com/dop251/goja
- **Servo (Rust browser):** https://github.com/servo/servo
- **WebKit Layout Docs:** https://webkit.org/blog/category/layout/

---

*Última actualización: 2026-01-09*
