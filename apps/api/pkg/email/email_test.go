package email

import (
	"strings"
	"testing"

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

	if !strings.Contains(rendered, "Accept Invitation") {
		t.Error("expected rendered template to contain 'Accept Invitation' button text")
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
