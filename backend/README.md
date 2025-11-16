# Kthulu

Kthulu is a Dockerized monolith that combines a Go backend and a React frontend. The backend exposes modular packages such as `auth`, `user`, `organization`, `product`, and `contact`, while the frontend in `frontend/` consumes the API.

Run the full stack with Docker:

```sh
docker-compose up --build
```

For a deeper overview of the project layout and module roadmap, see [backend/docs/project_planning.md](backend/docs/project_planning.md).

## Backend and Frontend

The Go backend is organized into modules under `backend/internal/modules`:

- `auth` for authentication and token issuance
- `user` for account management
- `organization` for multi-tenant features
- `product` for catalog management
- `contact` for customer and vendor tracking
- `inventory` for stock and warehouse control
- `calendar` for scheduling and events

The React frontend lives in the `frontend` directory and consumes the API. The
application now boots from `src/main.tsx`, which configures global providers and
the router. Global state is managed with a lightweight Zustand store at
`frontend/src/state/store.ts`. The previous `App.tsx` entry point has been
removed.

## Development with Docker Compose


1. Copy `.env.example` to `.env` and adjust values to fit your local setup.
2. Key environment variables include:
   - `API_PORT` for the backend API service
   - `WEB_PORT` for the React web service
   - `DB_PORT` for PostgreSQL
3. Start all services:
   ```sh
   docker-compose up --build
   ```
4. Access the applications:
   - API: http://localhost:${API_PORT}
   - Web: http://localhost:${WEB_PORT}
   - PostgreSQL: localhost:${DB_PORT}

For production deployment options using Docker Compose, Kustomize, or Helm, see [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md).

## Local Secrets

Development secrets live in a `.env.local` file or can be retrieved from a Vault
instance. The `scripts/setup-dev.sh` script and backend configuration load this
file automatically and fall back to `VAULT_SECRET_PATH` when the `vault` CLI is
available.

```sh
cp .env.local.example .env.local
# edit .env.local with your values
```

Required keys include `POSTGRES_*`, `DATABASE_URL`, `JWT_SECRET`, and
`JWT_REFRESH_SECRET`.

## Database Migrations

1. Install `goose`:

   ```sh
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

2. Apply migrations in development:

   ```sh
   goose -dir backend/migrations postgres $DATABASE_URL up
   ```

Migrations run automatically when the backend starts if `core/migrate.go` is executed.

## Contracts

The `backend/internal/contracts` package contains compile-time tests that assert
concrete types satisfy their corresponding interfaces. The pattern helps detect
implementation drift early without introducing runtime overhead.

## Running Tests

Run the complete backend and frontend test suites:

```sh
make test
```

To run only the frontend tests:

```sh
npm test
```

## API Type Generation

Regenerate frontend TypeScript types and Zod schemas from the OpenAPI specification:

```sh
npm --prefix frontend run gen:types
```

The script uses Bash and works on Unix and Windows environments. On Windows, run the
command from Git Bash or WSL so `bash` is available.

## Overrides and Extensions

Kthulu modules expose customization points through tagged annotations.
See [docs/EXTENDING.md](docs/EXTENDING.md) for guidance on safely
extending functions with `@kthulu:wrap` or overriding implementations
with `@kthulu:shadow`.

## Kthulu CLI

`kthulu-cli` ofrece utilidades de scaffolding y tareas de mantenimiento.

Subcomandos disponibles:

- `make:module <nombre>` – crea un módulo backend en `backend/internal/modules`.
- `make:handler <nombre>` – genera un handler HTTP en `backend/internal/handlers`.
- `make:service-test <nombre>` – genera una prueba con tabla de casos y fakes para cada puerto del servicio.

Ejemplo de uso:

```sh
kthulu-cli make:module user
```

Las plantillas se encuentran en `backend/cmd/kthulu-cli/templates`. Puede modificarlas o agregar nuevas para extender el generador. Revise y formatee el código generado antes de usarlo. Para añadir escenarios extra a las pruebas de servicio, edite el slice `tests` del archivo generado y configure los fakes dentro de `deps` siguiendo el patrón descrito en [docs/cli/make.md](docs/cli/make.md).

Para una guía más extensa con ejemplos de las plantillas consulte [docs/cli/make.md](docs/cli/make.md).

### Instalación del CLI

```sh
go install github.com/pmaojo/kthulu-go/backend/cmd/kthulu-cli@latest
```

