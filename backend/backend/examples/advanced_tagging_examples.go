// @kthulu:core
// Este archivo contiene ejemplos de uso del sistema de etiquetado avanzado de Kthulu
package examples

import (
	"context"
	"net/http"
	"time"
)

// Ejemplo de función con múltiples etiquetas avanzadas
// @kthulu:observable:metrics,tracing,logging
// @kthulu:security:authenticated
// @kthulu:audit:financial
// @kthulu:module:invoices
func CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	// Esta función será automáticamente instrumentada con:
	// - Métricas: invoice_creation_total (counter), invoice_creation_duration (histogram)
	// - Tracing: Span automático con tags de negocio
	// - Logging: Structured logs con correlation ID
	// - Security: Middleware de autenticación automático
	// - Audit: Log de auditoría para operaciones financieras

	return &Invoice{}, nil
}

// Ejemplo de punto de extensión seguro
// @kthulu:wrap
// @kthulu:module:auth
// @kthulu:cli:config:prompt="Custom authentication provider"
func AuthenticateUser(ctx context.Context, credentials Credentials) (*User, error) {
	// Esta función puede ser extendida de forma segura
	// El CLI generará hooks de extensión automáticamente
	// Los usuarios pueden añadir lógica personalizada sin romper la funcionalidad base

	return &User{}, nil
}

// Ejemplo de override peligroso con advertencias
// @kthulu:shadow
// @kthulu:module:auth
// @kthulu:security:system
// WARNING: Shadowing this function replaces core security logic
func ValidateJWTToken(token string) (*Claims, error) {
	// Esta función puede ser completamente reemplazada
	// El CLI generará advertencias y backups automáticamente
	// Solo usuarios avanzados deberían usar esta funcionalidad

	return &Claims{}, nil
}

// Ejemplo de candidato a microservicio
// @kthulu:microservice:standalone
// @kthulu:dependency:auth,user
// @kthulu:module:notifications
// @kthulu:observable:metrics
type NotificationService struct {
	// Este módulo puede ejecutarse como microservicio independiente
	// Dependencias: auth, user (serán incluidas automáticamente)
	// El CLI puede generar configuración de microservicio
}

// @kthulu:metrics:counter,histogram
// @kthulu:module:notifications
func (n *NotificationService) SendEmail(ctx context.Context, email Email) error {
	// Métricas generadas automáticamente:
	// - notification_emails_sent_total (counter)
	// - notification_email_duration_seconds (histogram)

	return nil
}

// Ejemplo de handler HTTP con instrumentación completa
// @kthulu:observable:metrics,tracing,logging
// @kthulu:security:admin
// @kthulu:audit:user_management
// @kthulu:module:user
func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// Instrumentación automática:
	// - HTTP metrics (requests, duration, status codes)
	// - Distributed tracing con contexto de request
	// - Structured logging con request ID
	// - Security middleware para permisos de admin
	// - Audit logging para operaciones de gestión de usuarios
}

// Ejemplo de código deprecated con migración
// @kthulu:deprecated:v2.0,alternative="NewUserService.CreateUser"
// @kthulu:module:user
func CreateUserOld(name string, email string) error {
	// Esta función será removida en v2.0
	// El CLI generará advertencias y sugerirá la alternativa
	// Incluirá scripts de migración automática

	return nil
}

// Ejemplo de característica experimental
// @kthulu:experimental:v1.5,flag="ENABLE_BIOMETRIC_AUTH"
// @kthulu:module:auth
// @kthulu:security:experimental
func BiometricAuthentication(ctx context.Context, biometricData []byte) (*User, error) {
	// Característica experimental que requiere flag de activación
	// El CLI generará configuración condicional
	// Solo se incluye si el flag está habilitado

	return &User{}, nil
}

// Ejemplo de template para generación de código
// @kthulu:cli:generator:crud
// @kthulu:template:entity
type EntityTemplate struct {
	// Esta estructura sirve como template para generar nuevas entidades
	// El CLI puede usar esto para `kthulu generate entity <name>`
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// @kthulu:cli:generator:crud
// @kthulu:template:usecase
type UseCaseTemplate struct {
	// Template para generar use cases CRUD
	// Incluye patrones estándar de validación, logging, etc.
}

// Ejemplo de configuración que requiere input del usuario
// @kthulu:cli:config:required,prompt="SMTP Server Host",env="SMTP_HOST"
// @kthulu:core
type SMTPConfig struct {
	Host     string `env:"SMTP_HOST"`
	Port     int    `env:"SMTP_PORT" default:"587"`
	Username string `env:"SMTP_USERNAME"`
	Password string `env:"SMTP_PASSWORD"`
}

// Ejemplo de análisis de dependencias automático
// @kthulu:dependency:auth,user,organization,product
// @kthulu:module:invoices
type InvoiceModule struct {
	// El analizador de dependencias detectará automáticamente:
	// - auth: Para autenticación de usuarios
	// - user: Para asociar facturas con usuarios
	// - organization: Para multi-tenancy
	// - product: Para items de factura
	//
	// El CLI validará que estos módulos estén incluidos
	// y los añadirá automáticamente si no están presentes
}

// Tipos de ejemplo para las funciones
type CreateInvoiceRequest struct{}
type Invoice struct{}
type Credentials struct{}
type User struct{}
type Claims struct{}
type Email struct{}
