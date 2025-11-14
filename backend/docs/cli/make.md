# Comandos `make:*`

El CLI `kthulu-cli` ofrece generadores de código que crean archivos base a partir de plantillas localizadas en `backend/cmd/kthulu-cli/templates`.

## `make:module`

Genera un nuevo módulo backend dentro de `backend/internal/modules`.

**Uso**

```sh
kthulu-cli make:module user
```

No acepta flags adicionales.

**Plantilla**

```go
package modules

import "go.uber.org/fx"

// UserModule is a Kthulu module scaffold.
var UserModule = fx.Options(
)
```

**Salida**

`backend/internal/modules/user.go`

## `make:handler`

Crea un handler HTTP en `backend/internal/handlers`.

**Uso**

```sh
kthulu-cli make:handler health
```

No acepta flags adicionales.

**Plantilla**

```go
package handlers

import "net/http"

// Health handles HTTP requests.
func Health(w http.ResponseWriter, r *http.Request) {
}
```

**Salida**

`backend/internal/handlers/health.go`

## `make:service-test`

Genera una prueba básica para un servicio.

**Uso**

```sh
kthulu-cli make:service-test inventory
```

No acepta flags adicionales.

**Plantilla**

```go
package inventory

import "testing"

// TestInventory exercises the Inventory service.
func TestInventory(t *testing.T) {
}
```

**Salida**

`backend/internal/inventory/service_test.go`

## Extender las plantillas

Las plantillas se encuentran en `backend/cmd/kthulu-cli/templates`. Puede modificarlas o añadir nuevas para ajustar el código generado. Para registrar un nuevo comando, siga el patrón de los archivos `cmd/make_*.go` y añádalo a `root.go`.

## Buenas prácticas

- Revise y formatee el código generado antes de usarlo.
- Añada la lógica de dominio manualmente; las plantillas solo proveen un punto de partida.
- Ejecute las pruebas (`make test`) después de generar código para validar que el proyecto sigue compilando.
