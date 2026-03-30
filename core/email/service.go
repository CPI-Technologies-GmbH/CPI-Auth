package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/config"
	"github.com/CPI-Technologies-GmbH/CPI-Auth/core/models"
)

// Service handles email delivery with templates.
type Service struct {
	cfg       *config.Config
	templates models.EmailTemplateRepository
	logger    *zap.Logger
}

// NewService creates a new email service.
func NewService(cfg *config.Config, templates models.EmailTemplateRepository, logger *zap.Logger) *Service {
	return &Service{
		cfg:       cfg,
		templates: templates,
		logger:    logger,
	}
}

// TemplateData holds the variables available in email templates.
type TemplateData struct {
	UserName   string
	UserEmail  string
	TenantName string
	TenantLogo string
	ActionURL  string
	Code       string
	AppName    string
	ExpiresIn  string
}

// resolveLocale normalizes a locale string, falling back to "en" if empty.
func resolveLocale(locale string) string {
	if locale == "" {
		return "en"
	}
	return locale
}

// SendVerification sends an email verification message.
func (s *Service) SendVerification(ctx context.Context, tenantID uuid.UUID, email, name, actionURL, locale string) error {
	return s.sendTemplated(ctx, tenantID, "verification", resolveLocale(locale), email, TemplateData{
		UserName:  name,
		UserEmail: email,
		ActionURL: actionURL,
		AppName:   "CPI Auth",
	})
}

// SendPasswordReset sends a password reset email.
func (s *Service) SendPasswordReset(ctx context.Context, tenantID uuid.UUID, email, name, actionURL, locale string) error {
	return s.sendTemplated(ctx, tenantID, "password_reset", resolveLocale(locale), email, TemplateData{
		UserName:  name,
		UserEmail: email,
		ActionURL: actionURL,
		AppName:   "CPI Auth",
		ExpiresIn: "1 hour",
	})
}

// SendMFACode sends an MFA verification code via email.
func (s *Service) SendMFACode(ctx context.Context, tenantID uuid.UUID, email, name, code, locale string) error {
	return s.sendTemplated(ctx, tenantID, "mfa", resolveLocale(locale), email, TemplateData{
		UserName:  name,
		UserEmail: email,
		Code:      code,
		AppName:   "CPI Auth",
		ExpiresIn: "5 minutes",
	})
}

// SendWelcome sends a welcome email to a new user.
func (s *Service) SendWelcome(ctx context.Context, tenantID uuid.UUID, email, name, locale string) error {
	return s.sendTemplated(ctx, tenantID, "welcome", resolveLocale(locale), email, TemplateData{
		UserName:  name,
		UserEmail: email,
		AppName:   "CPI Auth",
	})
}

// SendInvitation sends an organization invitation email.
func (s *Service) SendInvitation(ctx context.Context, tenantID uuid.UUID, email, inviterName, orgName, actionURL, locale string) error {
	return s.sendTemplated(ctx, tenantID, "invitation", resolveLocale(locale), email, TemplateData{
		UserName:   inviterName,
		UserEmail:  email,
		ActionURL:  actionURL,
		AppName:    "CPI Auth",
		TenantName: orgName,
	})
}

func (s *Service) sendTemplated(ctx context.Context, tenantID uuid.UUID, templateType, locale, toEmail string, data TemplateData) error {
	// Try to load custom template from DB
	tmpl, err := s.templates.GetByTypeAndLocale(ctx, tenantID, templateType, locale)
	if err != nil || tmpl == nil {
		// Use default template
		tmpl = s.getDefaultTemplate(templateType)
	}

	// Render subject
	subject, err := renderTemplate(tmpl.Subject, data)
	if err != nil {
		s.logger.Error("failed to render email subject", zap.Error(err))
		subject = tmpl.Subject
	}

	// Render body
	body, err := renderTemplate(tmpl.BodyHTML, data)
	if err != nil {
		s.logger.Error("failed to render email body", zap.Error(err))
		body = tmpl.BodyHTML
	}

	return s.send(toEmail, subject, body)
}

func renderTemplate(tpl string, data TemplateData) (string, error) {
	// Replace template variables
	replacer := strings.NewReplacer(
		"{{user.name}}", data.UserName,
		"{{user.email}}", data.UserEmail,
		"{{tenant.name}}", data.TenantName,
		"{{tenant.logo}}", data.TenantLogo,
		"{{action_url}}", data.ActionURL,
		"{{code}}", data.Code,
		"{{app.name}}", data.AppName,
		"{{expires_in}}", data.ExpiresIn,
	)
	result := replacer.Replace(tpl)

	// Also support Go template syntax
	t, err := template.New("email").Parse(result)
	if err != nil {
		return result, nil // Fall back to simple replacement
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return result, nil
	}
	return buf.String(), nil
}

func (s *Service) send(to, subject, body string) error {
	smtpCfg := s.cfg.SMTP

	from := smtpCfg.From
	if smtpCfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", smtpCfg.FromName, smtpCfg.From)
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body)

	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)

	var auth smtp.Auth
	if smtpCfg.User != "" {
		auth = smtp.PlainAuth("", smtpCfg.User, smtpCfg.Password, smtpCfg.Host)
	}

	err := smtp.SendMail(addr, auth, smtpCfg.From, []string{to}, []byte(msg))
	if err != nil {
		s.logger.Error("failed to send email",
			zap.String("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("sending email: %w", err)
	}

	s.logger.Info("email sent", zap.String("to", to), zap.String("type", subject))
	return nil
}

func (s *Service) getDefaultTemplate(templateType string) *models.EmailTemplate {
	templates := map[string]*models.EmailTemplate{
		"verification": {
			Subject:  "Verify your email address",
			BodyHTML: `<html><body><h2>Hi {{user.name}},</h2><p>Please verify your email address by clicking the link below:</p><p><a href="{{action_url}}">Verify Email</a></p><p>This link will expire in 24 hours.</p><p>Thanks,<br>{{app.name}}</p></body></html>`,
		},
		"password_reset": {
			Subject:  "Reset your password",
			BodyHTML: `<html><body><h2>Hi {{user.name}},</h2><p>We received a request to reset your password. Click the link below:</p><p><a href="{{action_url}}">Reset Password</a></p><p>This link will expire in {{expires_in}}.</p><p>If you didn't request this, you can safely ignore this email.</p><p>Thanks,<br>{{app.name}}</p></body></html>`,
		},
		"mfa": {
			Subject:  "Your verification code",
			BodyHTML: `<html><body><h2>Hi {{user.name}},</h2><p>Your verification code is:</p><h1 style="font-size:32px;letter-spacing:8px;text-align:center">{{code}}</h1><p>This code expires in {{expires_in}}.</p><p>Thanks,<br>{{app.name}}</p></body></html>`,
		},
		"welcome": {
			Subject:  "Welcome to {{app.name}}!",
			BodyHTML: `<html><body><h2>Welcome {{user.name}}!</h2><p>Your account has been created successfully.</p><p>Thanks,<br>{{app.name}}</p></body></html>`,
		},
		"invitation": {
			Subject:  "You've been invited to {{tenant.name}}",
			BodyHTML: `<html><body><h2>Hi,</h2><p>{{user.name}} has invited you to join {{tenant.name}}.</p><p><a href="{{action_url}}">Accept Invitation</a></p><p>Thanks,<br>{{app.name}}</p></body></html>`,
		},
	}

	if tmpl, ok := templates[templateType]; ok {
		return tmpl
	}
	return &models.EmailTemplate{
		Subject:  "Notification from CPI Auth",
		BodyHTML: `<html><body><p>You have a new notification.</p></body></html>`,
	}
}
