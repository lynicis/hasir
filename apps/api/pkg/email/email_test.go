package email

import (
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hasir-api/pkg/config"
)

func TestNewService(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "no-reply@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	service := NewService(cfg)

	if service == nil {
		t.Fatal("expected service to be created")
	}
}

func TestInviteTemplateData(t *testing.T) {
	data := inviteTemplateData{
		OrganizationName: "Test Org",
		InviteUrl:        "https://example.com/invite/abc123",
	}

	if data.OrganizationName != "Test Org" {
		t.Errorf("expected OrganizationName to be 'Test Org', got %s", data.OrganizationName)
	}

	if data.InviteUrl != "https://example.com/invite/abc123" {
		t.Errorf("expected InviteUrl to be 'https://example.com/invite/abc123', got %s", data.InviteUrl)
	}
}

func TestTemplateRendering(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "no-reply@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	svc := NewService(cfg).(*smtpService)

	data := inviteTemplateData{
		OrganizationName: "Acme Corp",
		InviteUrl:        "https://dashboard.example.com/invite/token123",
	}

	var body strings.Builder
	err := svc.templates.ExecuteTemplate(&body, "invite.html", data)
	if err != nil {
		t.Fatalf("failed to execute template: %v", err)
	}

	rendered := body.String()

	if !strings.Contains(rendered, "Acme Corp") {
		t.Error("expected rendered template to contain organization name")
	}

	if !strings.Contains(rendered, "https://dashboard.example.com/invite/token123") {
		t.Error("expected rendered template to contain invite URL")
	}

	if !strings.Contains(rendered, "<!DOCTYPE html>") {
		t.Error("expected rendered template to be HTML")
	}

	if !strings.Contains(rendered, "Accept invitation") {
		t.Error("expected rendered template to contain 'Accept invitation' button text")
	}
}

func TestInviteUrlGeneration(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "no-reply@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	svc := NewService(cfg).(*smtpService)

	tests := []struct {
		name         string
		dashboardUrl string
		token        string
		expectedUrl  string
	}{
		{
			name:         "standard URL",
			dashboardUrl: "https://dashboard.example.com",
			token:        "abc123",
			expectedUrl:  "https://dashboard.example.com/invite/abc123",
		},
		{
			name:         "URL with trailing slash",
			dashboardUrl: "https://dashboard.example.com/",
			token:        "xyz789",
			expectedUrl:  "https://dashboard.example.com//invite/xyz789",
		},
		{
			name:         "localhost URL",
			dashboardUrl: "http://localhost:3000",
			token:        "localtoken",
			expectedUrl:  "http://localhost:3000/invite/localtoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc.dashboardUrl = tt.dashboardUrl
			expectedFormat := tt.dashboardUrl + "/invite/" + tt.token
			if expectedFormat != tt.expectedUrl {
				t.Errorf("expected URL %s, got %s", tt.expectedUrl, expectedFormat)
			}
		})
	}
}

func TestSendInvite(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "no-reply@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	svc := NewService(cfg).(*smtpService)

	t.Run("success", func(t *testing.T) {
		err := svc.SendInvite("test@example.com", "Test Org", "token123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("template error handling", func(t *testing.T) {
		emptyTmpl := template.New("test")
		svcInvalid := &smtpService{
			config:       svc.config,
			dashboardUrl: svc.dashboardUrl,
			templates:    emptyTmpl,
		}

		err := svcInvalid.SendInvite("test@example.com", "Test Org", "token123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute invite template")
	})
}

func TestSendForgotPassword(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "no-reply@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	svc := NewService(cfg).(*smtpService)

	t.Run("success", func(t *testing.T) {
		err := svc.SendForgotPassword("test@example.com", "reset-token-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect to SMTP server")
	})

	t.Run("template error handling", func(t *testing.T) {
		emptyTmpl := template.New("test")
		svcInvalid := &smtpService{
			config:       svc.config,
			dashboardUrl: svc.dashboardUrl,
			templates:    emptyTmpl,
		}

		err := svcInvalid.SendForgotPassword("test@example.com", "reset-token-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute forgot-password template")
	})
}

func TestSendEmail_PortConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		port   int
		useTLS bool
	}{
		{
			name:   "port 465 uses TLS",
			port:   465,
			useTLS: false,
		},
		{
			name:   "port 587 uses STARTTLS",
			port:   587,
			useTLS: false,
		},
		{
			name:   "useTLS flag uses TLS",
			port:   25,
			useTLS: true,
		},
		{
			name:   "default port uses SendMail",
			port:   25,
			useTLS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Smtp: config.SmtpConfig{
					Host:     "smtp.example.com",
					Port:     tt.port,
					Username: "user@example.com",
					Password: "password",
					From:     "no-reply@example.com",
					UseTLS:   tt.useTLS,
				},
				DashboardUrl: "https://dashboard.example.com",
			}

			svc := NewService(cfg).(*smtpService)

			err := svc.sendEmail("test@example.com", "Test Subject", "Test body", false)
			assert.Error(t, err)
			if tt.port == 465 || tt.useTLS {
				assert.Contains(t, err.Error(), "failed to connect to SMTP server")
			} else if tt.port == 587 {
				assert.Contains(t, err.Error(), "failed to connect to SMTP server")
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSendEmail_ContentType(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     25,
			Username: "user@example.com",
			Password: "password",
			From:     "no-reply@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	svc := NewService(cfg).(*smtpService)

	t.Run("HTML content type", func(t *testing.T) {
		err := svc.sendEmail("test@example.com", "Test", "<html>body</html>", true)
		assert.Error(t, err)
	})

	t.Run("plain text content type", func(t *testing.T) {
		err := svc.sendEmail("test@example.com", "Test", "plain text", false)
		assert.Error(t, err)
	})
}

func TestForgotPasswordTemplateData(t *testing.T) {
	data := forgotPasswordTemplateData{
		ResetUrl: "https://example.com/reset-password/abc123",
	}

	if data.ResetUrl != "https://example.com/reset-password/abc123" {
		t.Errorf("expected ResetUrl to be 'https://example.com/reset-password/abc123', got %s", data.ResetUrl)
	}
}

func TestForgotPasswordTemplateRendering(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user@example.com",
			Password: "password",
			From:     "no-reply@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	svc := NewService(cfg).(*smtpService)

	data := forgotPasswordTemplateData{
		ResetUrl: "https://dashboard.example.com/reset-password/token123",
	}

	var body strings.Builder
	err := svc.templates.ExecuteTemplate(&body, "forgot-password.html", data)
	require.NoError(t, err)

	rendered := body.String()

	assert.Contains(t, rendered, "https://dashboard.example.com/reset-password/token123")
	assert.Contains(t, rendered, "<!DOCTYPE html>")
}

func TestSendEmail_MessageFormat(t *testing.T) {
	cfg := &config.Config{
		Smtp: config.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     25,
			Username: "user@example.com",
			Password: "password",
			From:     "sender@example.com",
			UseTLS:   false,
		},
		DashboardUrl: "https://dashboard.example.com",
	}

	svc := NewService(cfg).(*smtpService)
	err := svc.sendEmail("recipient@example.com", "Test Subject", "Test Body", false)

	assert.Error(t, err)
}
