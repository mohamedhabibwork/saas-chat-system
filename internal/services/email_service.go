package services

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"path/filepath"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
	templatesDir string
}

func NewEmailService(config map[string]string) *EmailService {
	return &EmailService{
		smtpHost:     config["smtp_host"],
		smtpPort:     config["smtp_port"],
		smtpUsername: config["smtp_username"],
		smtpPassword: config["smtp_password"],
		fromEmail:    config["from_email"],
		fromName:     config["from_name"],
		templatesDir: config["templates_dir"],
	}
}

type EmailData struct {
	Username    string
	Email       string
	ResetToken  string
	PlanName    string
	Usage       int64
	Limit       int64
	ExpiryDate  string
	ChannelName string
	BotName     string
}

func (s *EmailService) SendEmail(to, subject, templateName string, data EmailData) error {
	// Read template file
	templatePath := filepath.Join(s.templatesDir, templateName+".html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	// Execute template
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	// Set up email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build email message
	var message bytes.Buffer
	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	message.WriteString("\r\n")
	message.Write(body.Bytes())

	// Connect to SMTP server
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	// Send email
	if err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, message.Bytes()); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (s *EmailService) SendWelcomeEmail(to, username string) error {
	data := EmailData{
		Username: username,
		Email:    to,
	}

	return s.SendEmail(to, "Welcome to Our Platform", "welcome", data)
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to, username, resetToken string) error {
	data := EmailData{
		Username:   username,
		Email:      to,
		ResetToken: resetToken,
	}

	return s.SendEmail(to, "Password Reset Request", "password_reset", data)
}

// SendSubscriptionStartEmail sends an email when a subscription starts
func (s *EmailService) SendSubscriptionStartEmail(to, username, planName string) error {
	data := EmailData{
		Username: username,
		Email:    to,
		PlanName: planName,
	}

	return s.SendEmail(to, "Subscription Started", "subscription_start", data)
}

// SendStorageWarningEmail sends an email when storage usage is high
func (s *EmailService) SendStorageWarningEmail(to, username string, usage, limit int64) error {
	data := EmailData{
		Username: username,
		Email:    to,
		Usage:    usage,
		Limit:    limit,
	}

	return s.SendEmail(to, "Storage Usage Warning", "storage_warning", data)
}

// SendSubscriptionExpiryEmail sends an email when a subscription is about to expire
func (s *EmailService) SendSubscriptionExpiryEmail(to, username, planName, expiryDate string) error {
	data := EmailData{
		Username:   username,
		Email:      to,
		PlanName:   planName,
		ExpiryDate: expiryDate,
	}

	return s.SendEmail(to, "Subscription Expiring Soon", "subscription_expiry", data)
}

// SendChannelInviteEmail sends an email when a user is invited to a channel
func (s *EmailService) SendChannelInviteEmail(to, username, channelName string) error {
	data := EmailData{
		Username:    username,
		Email:       to,
		ChannelName: channelName,
	}

	return s.SendEmail(to, "Channel Invitation", "channel_invite", data)
}

// SendBotCreatedEmail sends an email when a bot is created
func (s *EmailService) SendBotCreatedEmail(to, username, botName string) error {
	data := EmailData{
		Username: username,
		Email:    to,
		BotName:  botName,
	}

	return s.SendEmail(to, "Bot Created", "bot_created", data)
} 