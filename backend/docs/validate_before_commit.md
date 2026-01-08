📚 Stack oficial Kthulu + reglas “de hierro”
(si rompes una, la CI te lo tira en la cara)

1 · Backend Go
Capa	Librería obligatoria	Versión mínima	Normas duras
HTTP Router	Chi v5	^5.0.0	- ❌ No Gin, no Echo.
- Todos los middlewares en internal/adapters/http/middleware.
DI / Wiring	Uber Fx	^1.25	- Proveedor por módulo (module.go).
- Prohibido init() global.
ORM	Gorm v1.25	>=1.25	- Nada de raw SQL en usecase.
- Funciones DB sólo en infrastructure/db.
Migraciones	Goose	^3.15	- Un archivo SQL por cambio.
- Convención YYYYMMDOptimalMM_<name>.sql.
Validación	validator/v10	>=10.18	- Validaciones en constructores de entidad.
Logging	Zap	^1.26	- Usa logger.Sugar().Infow sólo en adapters.
- Entidades nunca hacen logging.
Config	godotenv + core/config.go	—	- .env es la única fuente local.
- Variables en mayúsculas snake.
Tokens	golang-jwt/jwt/v5	>=5.2	- Sólo HS256 y RS256 permitidos.
Observabilidad	OpenTelemetry (otel)	SDK 1.28+	- Tracing en cada handler.
- Export a OTLP si OTEL_EXPORTER_OTLP_ENDPOINT set.
Job Queue       Asynq   ^0.27   - Usar Asynq; nada de goroutines infinitas.

2 · Frontend (Vite + React)
Área	Herramienta obligatoria	Versión	Normas duras
Bundler	Vite (create-vite)	^5	- Alias @/ a src.
UI Library	React 18	18.2+	- Stricto Mode ON.
Tipado	TypeScript	5.4+	- noImplicitAny y strict = true.
CSS Utility	Tailwind CSS	^3.4	- No CSS-in-JS salvo twMerge.
Estado global	Zustand	^4.5	- Un slice por bounded-context.
- Prohibido Redux.
Data Fetching	TanStack Query v5	>=5.0	- fetcher central axiosClient.
- Mutaciones type-safe.
HTTP Client	Axios	^1.7	- Interceptor refresh-token pre-instalado.
Testing	Vitest + React-Testing-Library	^1.5	- Cobertura mínima 70 %.
Lint / Format	ESLint (Airbnb-TS) + Prettier	—	- Error on warning.
- Husky pre-commit (lint-staged).

3 · Infraestructura & Dev EX
Herramienta	Regla
Docker	Multi-stage build para backend; frontend sirve con vite-preview.
Makefile	Objetivos estándar: dev, test, gen-types, openapi, lint, ci.
Git Hooks (Husky)	pre-commit = ESLint + Prettier + go vet ./....
pre-push = make test.
CI (GitHub Actions)	Jobs: go-lint, go-test, ts-lint, vite-test, openapi-drift.

4 · Reglas de arquitectura
Capas inmutables

scss
Copiar
Editar
adapters  →  usecase  →  repository(interface)  →  infra(db)
Dependencias sólo hacia la derecha.

go mod graph revisado en CI para romper si hay import cruzado.

Envuelve, no modifiques

Extiende en /app/wrap/.

Sombra total sólo en /app/shadow/ con tag //go:build shadow.

OpenAPI fuente de contrato

Cambiar struct → hay que ejecutar make openapi gen-types.

PR sin diff YAML/TS = ✗.

Nomenclatura

Go packages snake_case (no “models”).

TS files camelCase.file.ts.

Entidades suffijo sin “Entity” (ej. User).

Use-case files verb_noun.go (create_invoice.go).

Sin “magia”

Cero reflexión salvo validator y otel.

Cero global vars (usar Fx).

5 · Checklist de revisión (pull-request)

✅	Punto
Interfaces nuevas tienen tests de contrato.	
go test ./... y npm run test verdes.	
Cobertura backend > 80 %, frontend > 70 %.	
ESLint/Prettier - sin warnings.	
make openapi gen-types ejecutado y comiteado.	
No se añadió librería no aprobada (lista arriba).	
Nueva migración = archivo timestamp + down section.	

6 · Avisos rápidos
¿Necesitas websockets? → módulo realtime (Action Cable-like) pendiente, no metas gorilla/websocket adhoc.

¿Storage estático? → espera módulo files (S3/MinIO); no uses SDK directo.

¿Job queue? → usar Asynq; no DIY con goroutines infinitas.

Cumpliendo estas normas, el código se mantiene homogéneo, actual y testeable, evitando la tentación de “meto esto rápido y ya“.







