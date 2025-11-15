// @kthulu:module:notifier
package notifier

import (
	"os"

	"go.uber.org/fx"

	"github.com/kthulu/kthulu-go/backend/core"
	"github.com/kthulu/kthulu-go/backend/internal/repository"
)

// NotifierModule provides notification services for Fx.
var NotifierModule = fx.Options(
	fx.Provide(NewNotificationProvider),
)

// NewNotificationProvider creates the appropriate notification provider based on configuration
func NewNotificationProvider(logger core.Logger) repository.NotificationProvider {
	// Check if SMTP is configured
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	// If SMTP is fully configured, use SMTP provider
	if smtpHost != "" && smtpPort != "" && smtpUsername != "" && smtpPassword != "" && smtpFrom != "" {
		logger.Info("Using SMTP notification provider")
		config := SMTPConfig{
			Host:     smtpHost,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
			From:     smtpFrom,
		}
		return NewSMTPProvider(config, logger)
	}

	// Default to console provider for development
	logger.Info("Using console notification provider")
	return NewConsoleProvider(logger)
}
