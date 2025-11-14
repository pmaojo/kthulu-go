# Kthulu CLI

El comando `kthulu-cli` incluye utilidades para el proyecto, como generadores de código y tareas de mantenimiento.

## Generadores `make:*`

Los subcomandos `make:*` crean archivos base a partir de plantillas.
Para una referencia completa consulte [docs/cli/make.md](cli/make.md).

Ejemplo rápido para crear un módulo backend:

```sh
kthulu-cli make:module user
```

Otros generadores disponibles:

```sh
kthulu-cli make:handler health
kthulu-cli make:service-test account
```

