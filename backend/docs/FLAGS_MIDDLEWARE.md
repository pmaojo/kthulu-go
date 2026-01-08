# Flags Middleware

El módulo `flags` permite propagar *feature flags* a través de cabeceras HTTP.

## Configuración

Defina un archivo `config/headers.yml` donde cada clave es el nombre de la cabecera
que habilita un flag y el valor es el nombre interno del flag:

```yaml
X-Experimental: experimental
X-Beta: beta
```

## Uso

El middleware `FlagsMiddleware` lee estas cabeceras y las almacena en el contexto de la
petición. Puede recuperarse un flag dentro de un handler mediante `middleware.GetFlag`:

```go
value, ok := middleware.GetFlag(r.Context(), "experimental")
```

Para obtener todos los flags presentes:

```go
flags := middleware.GetAllFlags(r.Context())
```

Si no existe configuración o cabeceras, el middleware no modifica la petición.
