# Sistema de Etiquetado Kthulu - Especificación Completa

El sistema de etiquetas `@kthulu:*` es el corazón de la arquitectura modular y CLI-deconstructible del framework Kthulu. Proporciona metadatos ricos que permiten análisis automático, generación inteligente y funcionalidades avanzadas.

## 🏷️ **Taxonomía de Etiquetas**

### **Etiquetas Básicas (Implementadas)**

#### `@kthulu:core`
- **Propósito**: Marca archivos esenciales del framework
- **Uso**: Infraestructura, configuración, logging, database
- **CLI**: Siempre incluido en proyectos generados
- **Ejemplo**: `backend/core/config.go`, `backend/core/db.go`

#### `@kthulu:module:<name>`
- **Propósito**: Marca archivos específicos de un módulo
- **Uso**: Funcionalidad de negocio modular
- **CLI**: Incluido solo si el módulo es seleccionado
- **Ejemplos**: 
  - `@kthulu:module:auth` - Autenticación
  - `@kthulu:module:user` - Gestión de usuarios
  - `@kthulu:module:invoices` - Facturación
  - `@kthulu:module:verifactu` - Cumplimiento fiscal español

#### `@kthulu:generated`
- **Propósito**: Marca archivos auto-generados
- **Uso**: OpenAPI specs, tipos TypeScript, migraciones
- **CLI**: Regenerado automáticamente
- **Ejemplo**: `api/openapi.yaml`, `frontend/src/types/kthulu-api.ts`

---

## 🚀 **Etiquetas Avanzadas (Propuestas)**

### **Etiquetas de Extensibilidad**

#### `@kthulu:wrap`
- **Propósito**: Marca puntos de extensión seguros
- **Uso**: Funciones/clases que pueden ser extendidas sin romper funcionalidad
- **CLI**: Genera hooks de extensión automáticamente
- **Beneficio**: Permite customización sin fork del código

```go
// @kthulu:wrap
// @kthulu:module:auth
func (a *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
    // Implementación base que puede ser extendida
}
```

#### `@kthulu:shadow`
- **Propósito**: Marca override completo (peligroso)
- **Uso**: Reemplazo total de funcionalidad
- **CLI**: Genera advertencias y backups
- **Beneficio**: Máxima flexibilidad con advertencias de seguridad

```go
// @kthulu:shadow
// @kthulu:module:auth
// WARNING: Shadowing this function replaces core authentication logic
func (a *AuthUseCase) ValidateToken(token string) error {
    // Implementación que puede ser completamente reemplazada
}
```

### **Etiquetas de Observabilidad**

#### `@kthulu:observable`
- **Propósito**: Marca componentes que requieren métricas/tracing
- **Uso**: Handlers críticos, operaciones de negocio importantes
- **CLI**: Genera instrumentación automática
- **Beneficio**: Observabilidad enterprise sin código manual

```go
// @kthulu:observable:metrics,tracing,logging
// @kthulu:module:invoices
func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
    // Automáticamente instrumentado con métricas, tracing y logging
}
```

#### `@kthulu:metrics:<type>`
- **Propósito**: Especifica tipo de métricas a generar
- **Tipos**: `counter`, `histogram`, `gauge`, `summary`
- **CLI**: Genera código de métricas Prometheus
- **Beneficio**: Métricas de negocio automáticas

```go
// @kthulu:metrics:counter,histogram
// @kthulu:module:auth
func (a *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
    // Genera: login_attempts_total (counter), login_duration_seconds (histogram)
}
```

### **Etiquetas de Arquitectura**

#### `@kthulu:microservice`
- **Propósito**: Marca módulos candidatos a microservicio
- **Uso**: Módulos con bajo acoplamiento
- **CLI**: Genera configuración de microservicio
- **Beneficio**: Migración gradual a microservicios

```go
// @kthulu:microservice:standalone
// @kthulu:module:invoices
package invoices

// Este módulo puede ejecutarse como microservicio independiente
```

#### `@kthulu:dependency:<modules>`
- **Propósito**: Declara dependencias explícitas entre módulos
- **Uso**: Resolución automática de dependencias
- **CLI**: Valida y resuelve dependencias automáticamente
- **Beneficio**: Previene configuraciones inválidas

```go
// @kthulu:dependency:auth,user,organization
// @kthulu:module:invoices
package invoices

// Requiere módulos: auth, user, organization
```

### **Etiquetas de Generación**

#### `@kthulu:cli:generator`
- **Propósito**: Marca templates para generación de código
- **Uso**: Plantillas reutilizables para nuevos módulos
- **CLI**: Usado por `kthulu generate module <name>`
- **Beneficio**: Scaffolding consistente de nuevos módulos

```go
// @kthulu:cli:generator:crud
// @kthulu:template:entity
type {{.EntityName}} struct {
    ID        uint      `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

#### `@kthulu:cli:config`
- **Propósito**: Marca configuraciones que requieren input del usuario
- **Uso**: Variables de entorno, configuraciones específicas
- **CLI**: Genera prompts interactivos
- **Beneficio**: Configuración guiada

```go
// @kthulu:cli:config:required,prompt="Database URL"
// @kthulu:core
type DatabaseConfig struct {
    URL string `env:"DATABASE_URL"`
}
```

### **Etiquetas de Calidad**

#### `@kthulu:deprecated`
- **Propósito**: Marca código obsoleto
- **Uso**: Funciones/módulos que serán removidos
- **CLI**: Genera advertencias y alternativas
- **Beneficio**: Migración gradual y comunicación clara

```go
// @kthulu:deprecated:v2.0,alternative="NewAuthService"
// @kthulu:module:auth
func (a *AuthUseCase) OldLogin() {
    // Será removido en v2.0, usar NewAuthService.Login()
}
```

#### `@kthulu:experimental`
- **Propósito**: Marca características experimentales
- **Uso**: Funcionalidad en desarrollo o beta
- **CLI**: Genera advertencias y flags de activación
- **Beneficio**: Innovación controlada

```go
// @kthulu:experimental:v1.5,flag="ENABLE_EXPERIMENTAL_AUTH"
// @kthulu:module:auth
func (a *AuthUseCase) BiometricLogin() {
    // Característica experimental, requiere flag de activación
}
```

### **Etiquetas de Seguridad**

#### `@kthulu:security:<level>`
- **Propósito**: Marca nivel de seguridad requerido
- **Niveles**: `public`, `authenticated`, `admin`, `system`
- **CLI**: Genera middleware de seguridad automático
- **Beneficio**: Seguridad por defecto

```go
// @kthulu:security:admin
// @kthulu:module:user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    // Requiere permisos de administrador
}
```

#### `@kthulu:audit`
- **Propósito**: Marca operaciones que requieren auditoría
- **Uso**: Operaciones críticas de negocio
- **CLI**: Genera logging de auditoría automático
- **Beneficio**: Compliance y trazabilidad

```go
// @kthulu:audit:financial
// @kthulu:module:invoices
func (u *InvoiceUseCase) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) {
    // Operación auditada automáticamente
}
```

---

## 🔧 **Implementación del Sistema de Etiquetas**

### **Parser de Etiquetas**

```go
// @kthulu:core
package tags

type Tag struct {
    Type       string            // "core", "module", "observable", etc.
    Value      string            // Valor principal (nombre del módulo, etc.)
    Attributes map[string]string // Atributos adicionales
    File       string            // Archivo donde se encuentra
    Line       int               // Línea del archivo
}

type TagParser struct {
    tags []Tag
}

func (p *TagParser) ParseFile(filename string) ([]Tag, error) {
    // Implementación del parser
}

func (p *TagParser) FilterByType(tagType string) []Tag {
    // Filtrar por tipo de etiqueta
}
```

### **Analizador de Dependencias**

```go
// @kthulu:core
package analyzer

type DependencyAnalyzer struct {
    parser *TagParser
}

func (a *DependencyAnalyzer) ResolveDependencies(modules []string) ([]string, error) {
    // Resolver dependencias automáticamente
}

func (a *DependencyAnalyzer) ValidateConfiguration(config ModuleConfig) error {
    // Validar configuración de módulos
}
```

### **Generador de Código**

```go
// @kthulu:core
package generator

type CodeGenerator struct {
    analyzer *DependencyAnalyzer
    templates map[string]Template
}

func (g *CodeGenerator) GenerateObservability(tags []Tag) error {
    // Generar código de métricas y tracing
}

func (g *CodeGenerator) GenerateSecurity(tags []Tag) error {
    // Generar middleware de seguridad
}
```

---

## 📊 **Casos de Uso Avanzados**

### **1. Generación Inteligente de Microservicios**

```bash
# CLI detecta módulos marcados como @kthulu:microservice
kthulu extract microservice --module=invoices

# Genera:
# - Dockerfile independiente
# - docker-compose para el microservicio
# - Cliente gRPC/REST
# - Configuración de service mesh
```

### **2. Instrumentación Automática**

```bash
# CLI genera observabilidad basada en @kthulu:observable
kthulu generate observability

# Genera:
# - Métricas Prometheus
# - Trazas OpenTelemetry
# - Dashboards Grafana
# - Alertas automáticas
```

### **3. Auditoría de Seguridad**

```bash
# CLI analiza etiquetas de seguridad
kthulu audit security

# Reporta:
# - Endpoints sin autenticación
# - Operaciones sin auditoría
# - Configuraciones inseguras
# - Recomendaciones de mejora
```

### **4. Migración Asistida**

```bash
# CLI detecta código deprecated
kthulu migrate --from=v1.0 --to=v2.0

# Genera:
# - Plan de migración
# - Scripts de actualización
# - Tests de compatibilidad
# - Documentación de cambios
```

---

## 🎯 **Beneficios del Sistema Avanzado**

### **✅ Para Desarrolladores**
- **Scaffolding Inteligente**: Generación de código basada en patrones
- **Observabilidad Automática**: Métricas y tracing sin código manual
- **Seguridad por Defecto**: Middleware generado automáticamente
- **Migración Asistida**: Actualizaciones guiadas y seguras

### **✅ Para Arquitectos**
- **Análisis de Dependencias**: Visualización de acoplamiento
- **Extracción de Microservicios**: Identificación automática de candidatos
- **Compliance**: Auditoría y cumplimiento automatizado
- **Documentación Viva**: Metadatos siempre actualizados

### **✅ Para DevOps**
- **Instrumentación Consistente**: Observabilidad estandarizada
- **Deployment Inteligente**: Configuraciones optimizadas
- **Monitoreo Automático**: Alertas basadas en patrones de negocio
- **Escalabilidad**: Identificación de cuellos de botella

---

## 🚀 **Roadmap de Implementación**

### **Fase 1: Parser Básico** (1-2 semanas)
- Implementar parser de etiquetas existentes
- Crear analizador de dependencias
- Validar configuraciones de módulos

### **Fase 2: Etiquetas de Observabilidad** (2-3 semanas)
- Implementar `@kthulu:observable`
- Generar métricas Prometheus automáticas
- Crear instrumentación de tracing

### **Fase 3: Etiquetas de Extensibilidad** (2-3 semanas)
- Implementar `@kthulu:wrap` y `@kthulu:shadow`
- Crear sistema de hooks de extensión
- Generar advertencias de seguridad

### **Fase 4: Etiquetas Avanzadas** (3-4 semanas)
- Implementar etiquetas de microservicios
- Crear generadores de código
- Añadir análisis de seguridad

---

**Este sistema de etiquetado convierte a Kthulu en el framework más inteligente y automatizado del mercado, proporcionando capacidades que ningún otro scaffolder tiene.** 🎯

¿Te gustaría que implemente alguna de estas etiquetas específicas o prefieres que comience con el parser básico?