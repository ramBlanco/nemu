# Nemu

TUI en Go para ingresar un hostname, consultar su resolución DNS y presentar un resultado asociado al nombre consultado.

## Requisitos

- Para construir el ejecutable Linux con Docker Compose: Docker Engine con `docker compose`, un shell compatible con POSIX y `awk`. No hace falta tener Go instalado en el host para este flujo.
- Para ejecutar desde el código (`go run`) o ejecutar las pruebas: Go compatible con la directiva del módulo (`1.26.3`).
- Para usar Nemu: un terminal interactivo y acceso al resolver DNS configurado en el sistema. El binario generado está destinado a Linux y debe ejecutarse en un entorno compatible con la arquitectura para la que se construyó.

La TUI usa Bubble Tea v2, Bubbles v2 y Lip Gloss v2. Las versiones quedan fijadas en `go.mod` y `go.sum`.

## Ejecutar

Desde la raíz del proyecto:

```sh
go run ./cmd/nemu
```

Escribe un hostname en el campo y pulsa **Enter** para iniciar la consulta. Junto al hostname, Nemu muestra uno de estos estados:

- **Resuelve**: la consulta devolvió una o más direcciones.
- **No resuelve**: la consulta terminó sin respuestas.
- **No concluyente**: un error impidió completar la consulta; esto no demuestra que el hostname no exista.

Cuando hay respuestas, también se muestran las direcciones. **Ctrl+C** o **Esc** cierra la TUI.

## Construir para Linux con Docker Compose

Desde la raíz del proyecto, deriva la versión de Go de `go.mod` y construye el ejecutable:

```sh
set -eu
GO_VERSION="$(awk '$1 == "go" { print $2; exit }' go.mod)"
test -n "$GO_VERSION"
mkdir -p dist
GO_VERSION="$GO_VERSION" docker compose run --build --rm build
```

NOTA: se puede actualizar/crear el archivo .env basandose en el .env_example
```
docker compose run --build --rm build
```

El build usa la versión declarada en `go.mod` y deja el binario Linux ejecutable en `dist/nemu`. Para iniciarlo en una terminal interactiva de un Linux compatible:

```sh
./dist/nemu
```

Docker Compose solo construye el binario; Nemu se ejecuta directamente en el host, no como un contenedor. Este flujo no genera builds multiplataforma.

## Hostnames guardados

Los hostnames agregados se guardan localmente y aparecen bajo **Hostnames guardados** al iniciar Nemu de nuevo. El archivo está en `nemu/hostnames.json` dentro del directorio de configuración del usuario:

- Linux: `$XDG_CONFIG_HOME/nemu/hostnames.json`, o `~/.config/nemu/hostnames.json` si `XDG_CONFIG_HOME` no está definido.
- macOS: `~/Library/Application Support/nemu/hostnames.json`.
- Windows: `%AppData%\nemu\hostnames.json`.

Solo se conservan los nombres; los resultados DNS no se guardan. Si Nemu no puede leer o guardar el archivo, muestra un error en la TUI en lugar de afirmar que el hostname fue guardado.

## Verificación

```sh
go test ./...
go vet ./...
```

Las pruebas del flujo TUI usan un resolver falso, por lo que no dependen de DNS público ni de una conexión externa.
