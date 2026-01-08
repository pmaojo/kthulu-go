# Proyecto **kthulu‑original** – Plan de entrega completo

> **Objetivo:** disponer de un monolito de referencia con todos los módulos esenciales de backend empresarial (ERP‑lite) + frontend Vite, 100 % dockerizado, que servirá de fuente para el scaffolder `kthulu`.

---

## 1 · Estructura de carpetas

```bash
/kthulu-original/
├── backend/
│   ├── cmd/
│   │   └── service/           # main wiring (Router + ModuleSet) – binario final
│   │       └── main.go        # arranca fx, carga config y monta los módulos dinámicos
│   ├── internal/
│   │   ├── adapters/
│   │   │   ├── http/          # controllers REST, middlewares y módulos Fx
│   │   │   │   └── modules/   # módulos auto‑registrables (auth, user, etc.)
│   │   │   ├── cli/           # tooling del generador y validaciones
│   │   │   └── mcp/           # server MCP para exponer comandos
│   │   ├── domain/            # entidades y value objects puros DDD
│   │   │   ├── common/        # utilidades puras de dominio (constantes, fechas)
│   │   │   └── repository/    # puertos (interfaces) consumidos por usecase
│   │   ├── usecase/           # casos de uso (coordinan puertos)
│   │   └── infrastructure/
│   │       ├── db/            # adaptadores gorm ↔ postgres/sqlite
│   │       ├── observability/ # logging, tracing, métricas compartidas
│   │       └── config/        # lectura/normalización de config
│   ├── core/                  # config, logger, jwt util, shared errors
│   └── migrations/            # goose / sql‑migrate
├── frontend/
│   ├── src/                   # (estructura detallada más abajo)
│   ├── index.html
│   ├── vite.config.ts
│   └── tsconfig.json
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## 2 · Módulos – **Fase 1 (MVP)**

| Módulo       | Rutas REST / Funcionalidad mínima                                                                         |
| ------------ | --------------------------------------------------------------------------------------------------------- |
| **auth**     | `POST /auth/register`, `POST /auth/login`, `GET /auth/confirm`, `POST /auth/refresh`, `POST /auth/logout` |
| **user**     | `GET /users/me`, `PATCH /users/me`                                                                        |
| **access**   | `roles`, `permissions`, middleware RBAC (`X‑Role‑Scope`)                                                  |
| **notifier** | Producer de eventos email → consola (mock SMTP)                                                           |
| **core**     | JWT, configuración, migraciones automáticas, logger                                                       |

> Los módulos HTTP viven en `internal/adapters/http/modules` y se empaquetan como `fx.Option`. Cada uno registra sus rutas mediante un `RouteRegistry` con `fx.Invoke` y puede habilitarse o deshabilitarse mediante la variable `MODULES`.

---

## 3 · Módulos **ERP‑lite** (Fase 2, opcionales via CLI)

| Módulo        | Descripción breve                                    | Dependencias | Estado |
| ------------- | ---------------------------------------------------- | ------------ | ------ |
| **org**       | Multi‑tenant: organizaciones, invitaciones, dominios | auth, user   | ✅ Implementado |
| **contacts**  | Personas / empresas externas, embudo de clientes     | org          | ✅ Implementado |
| **invoices**  | Facturas, impuestos, pagos, PDF output (GoFPDF)      | contacts     | ✅ Implementado |
| **products**  | Catálogo, precios, variantes, tax class              | org          | ✅ Implementado |
| **inventory** | Stock, movimientos, almacenes                        | products     | ✅ Implementado |
| **calendar**  | Citas, eventos, disponibilidad                       | user, org    | ✅ Implementado |

Estos módulos se activarán con `kthulu add module <nombre>` y se copiarán desde este original.

---

## 4 · Diseño técnico – Backend

| Layer           | Tech / librería                            |
| --------------- | ------------------------------------------ |
| **Router**      | `github.com/go-chi/chi/v5`                 |
| **DI / wiring** | `go.uber.org/fx` + ModuleSet dinámico      |
| **ORM**         | `gorm.io/gorm` + `gorm.io/driver/postgres` |
| **Validation**  | `github.com/go-playground/validator/v10`   |
| **Auth tokens** | `github.com/golang-jwt/jwt/v5`             |
| **Env**         | `github.com/joho/godotenv`                 |
| **Migrations**  | `github.com/pressly/goose`                 |
| **Logging**     | `go.uber.org/zap`                          |
| **Job queue**   | `github.com/hibiken/asynq`                 |
| **Testing**     | `stretchr/testify` + `httptest`            |

### Sistema de módulos dinámicos con Fx

El contenedor de dependencias es `go.uber.org/fx`. Los módulos HTTP residen en `internal/adapters/http/modules` y se definen como `fx.Option`. En `cmd/service/main.go` se construye un `Registry` con todos los módulos disponibles y un `ModuleSetBuilder` compone la lista final según la variable de entorno `MODULES` (vacía = todos). `ModuleSet.Build` genera las opciones de Fx que inyectan casos de uso, adaptadores y registran rutas en un `RouteRegistry`, permitiendo habilitar o deshabilitar funcionalidades sin recurrir a globals ni recompilar.

### Esquema DB MVP

* `users` (id, email, password\_hash, confirmed\_at, role\_id)
* `roles` (id, name, description)
* `refresh_tokens` (id, user\_id, token, expires\_at)

### 4.1 Tipado compartido **Go ↔ TypeScript**

> **Objetivo:** evitar des‑sincronización de modelos entre backend y frontend y aprovechar al máximo el sistema de tipos de TS.

| Paso | Herramienta / Acción                                                                                                | Resultado                                                                             |
| ---- | ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| 1.   | Definir DTOs en Go como structs con tags `json` y validaciones `validator`                                          | **Single‑source‑of‑truth** ← dominio de datos vive en backend.                        |
| 2.   | Generar especificación **OpenAPI 3.1** automática (`github.com/getkin/kin-openapi/openapi3filter` o `oapi-codegen`) | Contrato versionado en `api/openapi.yaml`.                                            |
| 3.   | Ejecutar **oapi-codegen** (`--generate "typescript-fetch"`) o **go2ts** en CI                                       | Se produce `src/types/kthulu-api.ts` con interfaces TypeScript 100 % en línea con Go. |
| 4.   | En frontend, reexportar los modelos desde `types/` y envolver con Zod (`z.infer<typeof User>`)                      | Validación runtime + type‑safety compile‑time.                                        |
| 5.   | Hook TanStack Query usa tipos generados (`useQuery<User[]>`)                                                        | Autocompletado y chequeo estricto.                                                    |

**Ventajas**

* Cero duplicación manual de interfaces.
* Cambios en structs → PR rompe build si no se regenera contrato.
* Alineación total con reglas de validación backend.

**Tooling rápido**

```bash
# generar openapi a partir de rutas + structs
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
     --generate types,chi-server -o internal/adapters/http/auto_gen.go api/openapi.yaml

# generar TS
npx oapi-codegen --config oapi-ts.config.yaml api/openapi.yaml
```

---

## 5 · Frontend Vite (React + TS) — Estructura, linters y flujo DevEx

### 5.1 Tooling base

| Herramienta        | Motivo                                                       |
| ------------------ | ------------------------------------------------------------ |
| **Vite**           | Bundler moderno, HMR ultra‑rápido, DX similar a CRA.         |
| **React 18**       | Librería de UI declarativa.                                  |
| **TypeScript**     | Tipado fuerte, IntelliSense rico, captura errores en diseño. |
| **Tailwind CSS**   | Utilidades atómicas, velocidad de prototipado.               |
| **Zustand**        | Estado global minimalista por bounded‑context.               |
| **TanStack Query** | Caché y sincronización de datos remotos.                     |
| **Axios**          | Cliente HTTP con interceptores y cancelación.                |
| **Vitest + RTL**   | Testing unificado, rapidísimo bajo Vite.                     |

### 5.2 Árbol de carpetas sugerido

```bash
/frontend/
├── src/
│   ├── assets/            # imágenes, SVGs, fuentes
│   ├── components/
│   │   ├── context/       # React.Context (un archivo por contexto)
│   │   ├── hooks/         # Custom hooks reutilizables
│   │   ├── layouts/       # Reutiliza UI‑shells (AdminLayout, PublicLayout)
│   │   ├── routes/        # Definición central de rutas + navegación
│   │   ├── shared/        # Botones, modales, tablas genéricas
│   │   └── views/         # Páginas. <domain>/<Feature>/{Public|Private}
│   ├── config/            # axiosClient, queryClient, tailwindTheme, etc.
│   ├── data/              # mocks + fixtures para dev/test
│   ├── services/          # llamadas a API por recurso (user.service.ts)
│   ├── styles/            # Tailwind base + variables CSS + reset
│   ├── types/             # modelos / DTO compartidos con backend
│   ├── utils/             # helpers sin lógica de dominio
│   └── validations/       # esquemas Zod / Yup por recurso (user.dto.ts)
├── index.html
├── vite.config.ts
└── tsconfig.json
```

*Las carpetas pueden evolucionar, pero este baseline cubre el 95 % de necesidades.*

### 5.3 Flujo de autenticación

1. `POST /auth/login` → devuelve `accessToken` + `refreshToken`.
2. **axios interceptor** inyecta `Authorization` header y renueva con `/auth/refresh` al 401.
3. `useAuth()` expone `{user, login(), logout(), refresh()}`.
4. `<RequireAuth>` protege rutas privadas; redirige a `/login` si no hay sesión.

### 5.4 Linters, formato y ganchos Git

```bash
npm i -D eslint @typescript-eslint/{eslint-plugin,parser} eslint-plugin-jsx-a11y
npm i -D husky lint-staged prettier
```

* **ESLint + Airbnb TS** para reglas de calidad.
* **Prettier** alineado con ESLint para formato.
* **Husky (pre‑commit)** ejecuta `lint-staged` ⇒ corre ESLint/Prettier sólo sobre los archivos cambiados.

```json
// .lintstagedrc
{
  "*.{ts,tsx}": "eslint --fix"
}
```

### 5.5 VS Code workspace

`/.vscode/settings.json`

```json
{
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "editor.tabSize": 2,
  "editor.detectIndentation": false
}
```

### 5.6 Buenas prácticas clave

| Tema           | Recomendación                                        |
| -------------- | ---------------------------------------------------- |
| Rutas          | Enum `ELinks` centralizado – sin hardcode strings.   |
| App root       | Sólo providers de alto nivel + `<AppRoutes />`.      |
| Composición    | Componentes puros → hooks → servicios → API.         |
| Lógica dominio | **Nunca** en componentes UI; vive en hooks/services. |
| Import paths   | Usar path alias `@/` via Vite `resolve.alias`.       |

---

## 6 · Docker Compose (dev) · Docker Compose (dev)

```yaml
auth_db: &db
  image: postgres:15
  environment:
    POSTGRES_PASSWORD: kthulu
  volumes:
    - db-data:/var/lib/postgresql/data

services:
  api:
    build: ./backend
    env_file: .env
    ports: ["8080:8080"]
    depends_on: [db]

  web:
    build: ./frontend
    ports: ["5173:5173"]
    depends_on: [api]

  db: *db

volumes:
  db-data:
```

---

## 7 · Entregables Fase 1

| ID   | Descripción                                                           |
| ---- | --------------------------------------------------------------------- |
| E‑01 | Repo `kthulu-original` con estructura final                           |
| E‑02 | `backend/main.go` + Fx wiring de `auth`, `user`, `access`, `notifier` |
| E‑03 | Módulo **auth** con login/registro/confirmación + tests               |
| E‑04 | Módulo **user** con `/me` y actualización de perfil                   |
| E‑05 | Módulo **notifier** stub (log) + contrato interface                   |
| E‑06 | Frontend Vite con Login + Register + Profile                          |
| E‑07 | `docker-compose up` corre la pila completa                            |

---

## 8 · Roadmap de tasks (Sprint 0 – MVP)

| Task ID | Historia / Acción                                     | Responsable | Depends     |
| ------- | ----------------------------------------------------- | ----------- | ----------- |
| T‑001   | Scaffold carpetas base + go.mod                       | backend     | —           |
| T‑002   | Implementar core/ (config, logger, DB, Fx provider)   | backend     | T‑001       |
| T‑003   | Crear entidad & repo `User`, migración SQL            | backend     | T‑002       |
| T‑004   | Implementar módulo **auth** (service, handler, tests) | backend     | T‑003       |
| T‑005   | Implementar módulo **user** (service, handler)        | backend     | T‑004       |
| T‑006   | Middleware RBAC básico en **access**                  | backend     | T‑004       |
| T‑007   | Notifier stub + inyección en auth (sendConfirmEmail)  | backend     | T‑004       |
| T‑008   | Setup Vite + Tailwind + Ruta `/login`                 | frontend    | —           |
| T‑009   | Hook `useAuth` + llamada a `/auth/login`              | frontend    | T‑008       |
| T‑010   | Página `/register` + flujo confirmación               | frontend    | T‑009       |
| T‑011   | Dockerfile backend + scripts wait‑for‑db              | devops      | T‑002       |
| T‑012   | Dockerfile frontend + nginx static (prod)             | devops      | T‑008       |
| T‑013   | docker-compose.yml orquestado                         | devops      | T‑011 T‑012 |
| T‑014   | Implementar módulo **inventory** (service, handler, tests) | backend | T‑005 |
| T‑015   | Implementar módulo **calendar** (service, handler, tests)  | backend | T‑014 |

---

## 9 · Buenas prácticas DDD · SOLID · Hexagonal por plataforma

### 9.1 Backend Go

| Capa / carpeta               | Responsabilidad única                                            | Reglas                                                        |
| ---------------------------- | ---------------------------------------------------------------- | ------------------------------------------------------------- |
| `internal/domain`            | Entidades ricas, Value Objects, invariantes en constructores     | Sin dependencias externas; sólo stdlib y paquetes de dominio. |
| `internal/domain/repository` | Puertos (interfaces) invocados por los casos de uso               | Definir contratos; jamás tocar Gorm aquí.                     |
| `internal/usecase`           | Coordinadores de aplicación. Orquestan repos, services, policies | Un use case = archivo. Nada de lógica de infraestructura.     |
| `internal/adapters/http`     | REST handlers, middlewares, módulos Fx                           | Sólo transformar DTO ↔ dominio; inyectar usecases/repos.      |
| `internal/adapters/cli|mcp`  | CLI tooling, analizador de dependencias, servidor MCP            | Nada de acceso directo a DB; sólo casos de uso/core.          |
| `internal/infrastructure/*`  | Adaptadores externos (db, observability, storage, queues, config) | Dependen de dominio/repos. Nunca importar adapters.           |
| Validación                   | `validator/v10` en constructores de entidades                    | Errores `domain.ErrInvalidX`.                                 |
| DI / Fx                      | Módulos proporcionan sólo dependencias necesarias                | No variables globales.                                        |

### 9.2 Frontend Vite + React

| Área                 | Buenas prácticas hexagonales / SOLID | Detalles                                |
| -------------------- | ------------------------------------ | --------------------------------------- |
| **Domain models**    | Interfaces TS, Zod schemas           | Sin dep. de UI; serializables JSON.     |
| **Servicios API**    | Hooks TanStack Query                 | Separados; sin estado global implícito. |
| **Adaptadores HTTP** | `axiosClient.ts` centralizado        | Interceptors globales; retry.           |
| **Presentación**     | Componentes puros                    | Props explícitas; sin lógica negocio.   |
| **Estado global**    | Zustand slices por bounded‑context   | Evitar cross‑slice coupling.            |
| **Pruebas**          | Vitest + RTL                         | Cobertura de componentes críticos.      |

El estado compartido se gestiona con [Zustand](https://github.com/pmndrs/zustand).
Un store base vive en `frontend/src/state/store.ts` y puede ampliarse por
bounded‑context mediante slices.

#### Ejemplo TanStack Query

```tsx
// Hook reutilizable
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import healthService from '@/services/health.service';

export const useHealth = () =>
  useQuery({
    queryKey: ['health'],
    queryFn: () => healthService.check().then((r) => r.data),
  });

export const useUpdateHealth = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: healthService.update,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['health'] }),
  });
};
```

Notas:

* `QueryClientProvider` se monta una sola vez en `main.tsx` con un `QueryClient` compartido.
* `useQuery` gestiona caché, errores y estados de carga de forma automática.
* `useMutation` simplifica las operaciones de escritura e invalida las consultas relacionadas.

### 9.3 Principios transversales

* **DIP** – dominios nunca importan infraestructura ni UI.
* **SRP** – un archivo, una responsabilidad clara.
* **Open/Closed** – extiende via interfaces; no modifiques core generado.
* **Testing first** – el scaffolder crea `*_test.go` y spec Vitest.

### 9.4 Lecciones heredadas de **Symfony** (para elevar calidad)

| Concepto Symfony           | Cómo lo adoptamos en Go / Kthulu                                | Beneficio                                                 |
| -------------------------- | --------------------------------------------------------------- | --------------------------------------------------------- |
| **Service Container**      | `fx` modules → providers declarativos, scopes, lifecycle hooks  | DI explícita sin singletons; fácil test‑mock              |
| **Bundles (Modularidad)**  | `internal/adapters/http/modules/<name>` → cada uno auto‑registrable | Aislamos dominio + infra; se puede distribuir como plugin |
| **Event Dispatcher**       | `core/events` (mini bus + pub/sub via channels)                 | Desacopla side‑effects (ej. enviar mail, logging)         |
| **Messenger / Bus**        | `usecase` + goroutines workers / Asynq                          | Comandos asíncronos, retries, backoff                     |
| **Console Commands**       | `cmd/kthulu-cli` (cobra)                                        | Tareas de mantenimiento, migrations, cron                 |
| **Kernel HTTP middleware** | `chi` middlewares en `adapters/http/middleware`                 | Stack claro: logging, recovery, auth guard                |
| **Validation Annotations** | Struct tags + `validator/v10`                                   | Reglas declarativas cerca de DTO                          |
| **Param Converter**        | Bind JSON → DTO en handler + automapping DTO → Entity factories | Manejo coherente de entrada/salida                        |

> **Takeaway:** copiamos la *filosofía* (container, eventos, bundles, console) y la implementamos con herramientas idiomáticas de Go para mantener alto rendimiento y simplicidad.

### Generación y validación de grafo de dependencias

El comando `kthulu-cli plan` puede producir un grafo de validación y comprobar reglas de acoplamiento entre módulos.

```bash
# Generar grafo en formato DOT
kthulu-cli plan --graph --format=dot
# Otras opciones de salida
kthulu-cli plan --graph --format=json
kthulu-cli plan --graph --format=yaml

# Validar el grafo y finalizar con error si hay violaciones
kthulu-cli plan --validate
```

El grafo se escribe en `/tmp/kthulu.graph.<fmt>` y se construye mediante `BuildValidationGraph`. La opción `--validate` invoca `ValidateGraph` y devolverá un error si se detectan dependencias no permitidas.

---

> **Este documento es la fuente única de verdad** para `kthulu-original`. Cualquier cambio deberá reflejarse aquí para mantener coherencia entre planificación y código.
