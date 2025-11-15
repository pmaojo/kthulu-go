// @kthulu:module:notifier
package notifier

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/pmaojo/kthulu-go/backend/core"
	"github.com/pmaojo/kthulu-go/backend/internal/repository"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// SMTPProvider implements NotificationProvider using SMTP
type SMTPProvider struct {
	config SMTPConfig
	logger core.Logger
}

// NewSMTPProvider creates a new SMTP notification provider
func NewSMTPProvider(config SMTPConfig, logger core.Logger) repository.NotificationProvider {
	return &SMTPProvider{
		config: config,
		logger: logger,
	}
}

// SendNotification sends a notification via SMTP
func (s *SMTPProvider) SendNotification(ctx context.Context, req repository.NotificationRequest) error {
	s.logger.Info("Sending SMTP notification",
		"type", string(req.Type),
		"to", req.To,
		"subject", req.Subject,
	)

	// Create SMTP auth
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Compose email message
	msg := s.composeMessage(s.config.From, req.To, req.Subject, req.Body)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)
	err := smtp.SendMail(addr, auth, s.config.From, []string{req.To}, []byte(msg))
	if err != nil {
		s.logger.Error("Failed to send SMTP notification",
			"to", req.To,
			"subject", req.Subject,
			"error", err,
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("SMTP notification sent successfully",
		"to", req.To,
		"subject", req.Subject,
	)

	return nil
}

// SendEmailConfirmation sends an email confirmation notification
func (s *SMTPProvider) SendEmailConfirmation(ctx context.Context, email, confirmationCode string) error {
	req := repository.NotificationRequest{
		To:      email,
		Subject: "Confirm Your Email Address",
		Body:    s.renderEmailConfirmationTemplate(confirmationCode),
		Type:    repository.NotificationTypeEmailConfirmation,
		Data: map[string]interface{}{
			"confirmationCode": confirmationCode,
		},
	}

	return s.SendNotification(ctx, req)
}

// SendPasswordReset sends a password reset notification
func (s *SMTPProvider) SendPasswordReset(ctx context.Context, email, resetCode string) error {
	req := repository.NotificationRequest{
		To:      email,
		Subject: "Reset Your Password",
		Body:    s.renderPasswordResetTemplate(resetCode),
		Type:    repository.NotificationTypePasswordReset,
		Data: map[string]interface{}{
			"resetCode": resetCode,
		},
	}

	return s.SendNotification(ctx, req)
}

// SendWelcomeEmail sends a welcome email notification
func (s *SMTPProvider) SendWelcomeEmail(ctx context.Context, email, name string) error {
	req := repository.NotificationRequest{
		To:      email,
		Subject: "Welcome!",
		Body:    s.renderWelcomeTemplate(name),
		Type:    repository.NotificationTypeWelcome,
		Data: map[string]interface{}{
			"name": name,
		},
	}

	return s.SendNotification(ctx, req)
}

// composeMessage creates a properly formatted email message
func (s *SMTPProvider) composeMessage(from, to, subject, body string) string {
	msg := fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=UTF-8\r\n"
	msg += "\r\n"
	msg += body

	return msg
}

// renderEmailConfirmationTemplate renders the email confirmation template
func (s *SMTPProvider) renderEmailConfirmationTemplate(confirmationCode string) string {
	template := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Confirm Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">Confirm Your Email Address</h2>
        <p>Thank you for registering! Please confirm your email address by using the confirmation code below:</p>
        <div style="background-color: #f8f9fa; padding: 15px; border-radius: 5px; text-align: center; margin: 20px 0;">
            <h3 style="color: #007bff; font-family: monospace; letter-spacing: 2px;">%s</h3>
        </div>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="font-size: 12px; color: #666;">This is an automated message, please do not reply.</p>
    </div>
</body>
</html>`

	return fmt.Sprintf(template, confirmationCode)
}

// renderPasswordResetTemplate renders the password reset template
func (s *SMTPProvider) renderPasswordResetTemplate(resetCode string) string {
	template := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">Reset Your Password</h2>
        <p>You requested a password reset. Use the code below to reset your password:</p>
        <div style="background-color: #f8f9fa; padding: 15px; border-radius: 5px; text-align: center; margin: 20px 0;">
            <h3 style="color: #dc3545; font-family: monospace; letter-spacing: 2px;">%s</h3>
        </div>
        <p>If you didn't request a password reset, you can safely ignore this email.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="font-size: 12px; color: #666;">This is an automated message, please do not reply.</p>
    </div>
</body>
</html>`

	return fmt.Sprintf(template, resetCode)
}

// renderWelcomeTemplate renders the welcome email template
func (s *SMTPProvider) renderWelcomeTemplate(name string) string {
	template := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome!</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h2 style="color: #2c3e50;">Welcome%s!</h2>
        <p>Thank you for joining us. We're excited to have you on board!</p>
        <p>You can now start using all the features available in your account.</p>
        <p>If you have any questions, feel free to reach out to our support team.</p>
        <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">
        <p style="font-size: 12px; color: #666;">This is an automated message, please do not reply.</p>
    </div>
</body>
</html>`

	nameStr := ""
	if name != "" {
		nameStr = " " + name
	}

	return fmt.Sprintf(template, nameStr)
}

// Ensure SMTPProvider implements NotificationProvider
var _ repository.NotificationProvider = (*SMTPProvider)(nil)
