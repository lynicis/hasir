package email

import (
	"bytes"
	"crypto/tls"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"

	"apps/api/pkg/config"
)

//go:embed templates/*.html
var templateFS embed.FS

type Service interface {
	SendInvite(to, organizationName, inviteToken string) error
}

type smtpService struct {
	config       *config.SmtpConfig
	dashboardUrl string
	templates    *template.Template
}

func NewService(cfg *config.Config) Service {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(fmt.Sprintf("failed to parse email templates: %v", err))
	}

	return &smtpService{
		config:       &cfg.Smtp,
		dashboardUrl: cfg.DashboardUrl,
		templates:    tmpl,
	}
}

type inviteTemplateData struct {
	OrganizationName string
	InviteUrl        string
}

func (s *smtpService) SendInvite(to, organizationName, inviteToken string) error {
	inviteUrl := fmt.Sprintf("%s/invite/%s", s.dashboardUrl, inviteToken)

	data := inviteTemplateData{
		OrganizationName: organizationName,
		InviteUrl:        inviteUrl,
	}

	var body bytes.Buffer
	if err := s.templates.ExecuteTemplate(&body, "invite.html", data); err != nil {
		return fmt.Errorf("failed to execute invite template: %w", err)
	}

	subject := fmt.Sprintf("You've been invited to join %s", organizationName)
	return s.sendEmail(to, subject, body.String(), true)
}

func (s *smtpService) sendEmail(to, subject, body string, isHTML bool) error {
	from := s.config.From

	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=\"utf-8\"\r\n\r\n%s",
		from, to, subject, contentType, body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, from, to, []byte(msg))
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func (s *smtpService) sendWithTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}
