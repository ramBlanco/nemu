# 🤖 Guía de Desarrollo para el Agente (Go & TUI Expert)

Este documento define los estándares de código, la arquitectura y las mejores prácticas obligatorias para el desarrollo de este proyecto en **Go (Golang)** y su interfaz de usuario en la terminal (**TUI** basada en el ecosistema Charm: Bubble Tea, Bubbles, Lip Gloss).

---

## 1. Principios Generales de Go (Idiomatic Go)

- **Simplicidad sobre Inteligencia:** Escribe código claro, legible y directo. Prefiere la legibilidad frente a la optimización prematura.
- **Manejo de Errores:** 
  - Nunca ignores un error (`_`). 
  - Envuelve los errores usando `%w` para preservar la traza: `fmt.Errorf("falló al resolver DNS: %w", err)`.
  - Comprueba los errores inmediatamente después de que ocurran.
- **Concurrencia Segura:**
  - Pasa siempre `context.Context` como el primer parámetro en funciones que realicen operaciones de red, I/O o procesos pesados.
  - Evita fugas de gorrutinas (*goroutine leaks*): asegúrate de que toda gorrutina tenga una condición clara de salida o reciba un canal de cancelación.
- **Estructura del Proyecto (Standard Go Project Layout):**
  - `cmd/`: Puntos de entrada de la aplicación (`main.go`).
  - `internal/`: Código privado de la aplicación (lógica de negocio, TUI, integración con herramientas externas).
  - `pkg/`: Librerías reutilizables públicamente (si aplica).

---

## 2. Arquitectura y Buenas Prácticas para TUI (Bubble Tea)

Dado que la interfaz utiliza el framework **Bubble Tea** (arquitectura The Elm Architecture - TEA), sigue estrictamente estas reglas:

- **Modelo Inmutable:** El estado (`Model`) debe actualizarse de forma pura en la función `Update`. Nunca modifiques el estado globalmente ni de forma concurrente fuera del ciclo de actualización.
- **Comandos no bloqueantes (`tea.Cmd`):**
  - Las operaciones pesadas (como ejecutar `subfinder`, consultas DNS, o peticiones HTTP) **nunca** deben bloquear el hilo principal de la UI. Deben ejecutarse dentro de un `tea.Cmd` que devuelva un mensaje (`tea.Msg`) al finalizar.
- **Manejo de Redimensionamiento:**
  - Maneja siempre el evento `tea.WindowSizeMsg` en el método `Update` para adaptar tablas, inputs y contenedores dinámicamente al tamaño de la terminal del usuario.
- **Separación de Intereses:**
  - Mantén la lógica de negocio y las llamadas a herramientas CLI separadas de la vista (`View`). La vista debe ser puramente una función determinista del estado actual.
- **Estilos con Lip Gloss:**
  - Centraliza la definición de estilos (colores, fuentes, bordes) para mantener consistencia visual en toda la terminal.

---

## 3. Integración con Herramientas Externas (CLI Wrappers)

Como la herramienta orquestará software open source existente (`subfinder`, `dnsx`, etc.):
- **Seguridad en Subprocesos:** Usa `os/exec` con contextos para poder cancelar o aplicar tiempos límite (*timeouts*) a la ejecución de herramientas externas.
- **Parsing Estructurado:** Siempre que sea posible, solicita la salida de las herramientas en formato JSON (`-json`) para parsearlas de forma robusta en estructuras de Go mediante `encoding/json` en lugar de hacer *regex* sobre texto plano.

---

## 4. Pruebas y Calidad de Código

- **Tests Unitarios:** Escribe pruebas para la lógica de negocio y parseo usando la librería estándar `testing`. Utiliza pruebas orientadas a tablas (*table-driven tests*).
- **Llinting y Formato:**
  - Todo el código debe pasar obligatoriamente por `gofmt` / `go fmt`.
  - El código debe estar libre de advertencias al ejecutar `go vet` y herramientas de análisis estático como `golangci-lint`.