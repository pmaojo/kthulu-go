# Kthulu CLI

El comando `kthulu-cli` incluye utilidades para el proyecto, como generadores de código y tareas de mantenimiento.

## Generadores `make:*`

Los subcomandos `make:*` crean archivos base a partir de plantillas.
Para una referencia completa consulte [docs/cli/make.md](cli/make.md).

Ejemplo rápido para crear un módulo backend:

```sh
kthulu-cli make:module user
```

## Instalación

Instale el CLI desde el módulo Go oficial:

```sh
go install github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli@latest
```

Otros generadores disponibles:

```sh
kthulu-cli make:handler health
kthulu-cli make:service-test account
```

El generador `make:handler` ahora crea un struct que recibe su puerto por constructor,
expone `RegisterRoutes(chi.Router)` para mantener el ruteo fuera de la lógica y añade un
`_handler_test.go` basado en `httptest` para demostrar la delegación al servicio inyectado.

