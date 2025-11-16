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

Crea un adaptador HTTP hexagonal en `backend/internal/handlers` con constructor para Fx,
registro explícito de rutas y DTOs independientes del dominio.

**Uso**

```sh
kthulu-cli make:handler health
```

No acepta flags adicionales.

**Plantilla**

```go
package handlers

import (
        "encoding/json"
        "errors"
        "net/http"

        "github.com/go-chi/chi/v5"
)

type HealthHandler struct {
        svc HealthPort
}

func NewHealth(svc HealthPort) *HealthHandler {
        return &HealthHandler{svc: svc}
}

func (h *HealthHandler) RegisterRoutes(router chi.Router) {
        router.Method(http.MethodPost, "/health", http.HandlerFunc(h.Handle))
}

type HealthRequest struct {
        Payload string `json:"payload"`
}

type HealthResponse struct {
        Result string `json:"result"`
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
        req, _ := decodeHealthRequest(r)
        resp, _ := h.svc.HandleHealth(r.Context(), req)
        encodeHealthResponse(w, resp)
}
```

La plantilla incluye `decodeHealthRequest`/`encodeHealthResponse` para encapsular JSON y
un archivo `_handler_test.go` que usa `httptest` con un mock para comprobar que el handler
delegue en el puerto inyectado.

**Salida**

- `backend/internal/handlers/health.go`
- `backend/internal/handlers/health_handler_test.go`

## `make:service-test`

Genera una prueba con tabla de casos y dobles de prueba para cada puerto del servicio.

**Uso**

```sh
kthulu-cli make:service-test inventory
```

No acepta flags adicionales.

**Plantilla**

```go
package inventory

import (
    "context"
    "errors"
    "testing"
)

type testDeps struct {
    primaryPort *fakePrimaryPort
}

func TestInventory(t *testing.T) {
    tests := []struct {
        name    string
        deps    func(t *testing.T) testDeps
        args    args
        want    any
        wantErr error
    }{
        { /* caso exitoso */ },
        { /* caso de error */ },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            subject := newTestInventory(t, tt.deps(t))
            got, err := subject(tt.args.ctx, tt.args.input)
            // assertErr + assertResult helpers
        })
    }
}
```

La plantilla completa incluye `fakePrimaryPort` (replíquelo para cada dependencia real), los helpers `assertError`, `assertResult` y `newTest<Service>` que fallan si no se reemplazan por la construcción real del servicio.

**¿Cómo extender la tabla?**

1. Añada una nueva entrada al slice `tests` con un `name` descriptivo.
2. Configure los fakes dentro de la función `deps` para simular el comportamiento esperado del puerto (por ejemplo, retornos alternativos, errores, validaciones sobre la entrada).
3. Ajuste `args`, `want` y `wantErr` según el escenario.
4. Si el servicio utiliza puertos adicionales, duplique el patrón de `fakePrimaryPort` para cada uno y expóngalos desde `testDeps`.

De esta forma puede cubrir casos límite, regresiones y fallos de integraciones externas sin duplicar código de preparación.

**Salida**

`backend/internal/inventory/service_test.go`

## Extender las plantillas

Las plantillas se encuentran en `backend/cmd/kthulu-cli/templates`. Puede modificarlas o añadir nuevas para ajustar el código generado. Para registrar un nuevo comando, siga el patrón de los archivos `cmd/make_*.go` y añádalo a `root.go`.

## Buenas prácticas

- Revise y formatee el código generado antes de usarlo.
- Añada la lógica de dominio manualmente; las plantillas solo proveen un punto de partida.
- Ejecute las pruebas (`make test`) después de generar código para validar que el proyecto sigue compilando.
