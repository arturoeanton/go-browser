# SpiderGopher Roadmap

Motor JavaScript para Go-Browser, inspirado en SpiderMonkey de Firefox.

## Fase 1: Fundamentos ✅
- [x] Integración de Goja Runtime
- [x] Event Loop básico (macrotasks)
- [x] Console API (log, warn, error)
- [x] Timers: setTimeout
- [x] Mock de document.getElementById

## Fase 2: DOM Bridge ✅
- [x] Conectar con el DOM real de Gocko
- [x] Exponer nodos como objetos JS
- [x] querySelector / querySelectorAll
- [x] createElement / createTextNode
- [ ] innerHTML / textContent (set)
- [ ] appendChild / removeChild

## Fase 3: Eventos ✅
- [x] addEventListener completo
- [x] dispatchEvent
- [ ] Event bubbling/capturing
- [ ] Eventos de mouse/teclado

## Fase 4: Async Avanzado
- [ ] Promises / Microtasks
- [x] setInterval / clearInterval
- [x] setTimeout / clearTimeout
- [ ] requestAnimationFrame

## Fase 5: Network ✅
- [x] fetch API
- [ ] XMLHttpRequest (opcional)

## Fase 6: Storage ✅
- [x] localStorage (SQLite file)
- [x] sessionStorage (in-memory)

## Fase 7: Debugging
- [ ] Source maps
- [ ] Stack traces mejorados
- [ ] DevTools básico

---

**Estado actual:** Fases 1,2,3,5,6 completas. Timers completos. SpiderGopher conectado al DOM de Gocko.
