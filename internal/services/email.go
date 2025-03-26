package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/smtp"
	"path/filepath"
	"time"

	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
)

// EmailService handles email delivery
type EmailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewEmailService creates a new email service
func NewEmailService(smtpHost string, smtpPort int, smtpUsername, smtpPassword, fromEmail, fromName string) *EmailService {
	return &EmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		fromName:     fromName,
	}
}

// SendReport sends a report via email
func (s *EmailService) SendReport(recipients []string, report *Report, schedule *models.ReportSchedule) error {
	// Create email template
	tmpl, err := template.New("report").Parse(reportEmailTemplate)
	if err != nil {
		return fmt.Errorf("error parsing email template: %v", err)
	}

	// Prepare email data
	data := struct {
		ReportName    string
		ReportType    string
		GeneratedAt   time.Time
		TenantName    string
		ReportSummary string
	}{
		ReportName:    schedule.Name,
		ReportType:    schedule.Type,
		GeneratedAt:   report.CreatedAt,
		TenantName:    report.Options.TenantID, // You might want to fetch actual tenant name
		ReportSummary: generateReportSummary(report),
	}

	// Generate email body
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("error generating email body: %v", err)
	}

	// Create multipart message
	message := &bytes.Buffer{}
	writer := multipart.NewWriter(message)

	// Add headers
	fmt.Fprintf(message, "From: %s <%s>\r\n", s.fromName, s.fromEmail)
	fmt.Fprintf(message, "To: %s\r\n", recipients[0])
	if len(recipients) > 1 {
		fmt.Fprintf(message, "Cc: %s\r\n", recipients[1:])
	}
	fmt.Fprintf(message, "Subject: %s Report: %s\r\n", schedule.Type, schedule.Name)
	fmt.Fprintf(message, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(message, "Content-Type: multipart/mixed; boundary=%s\r\n", writer.Boundary())

	// Add text part
	textPart, err := writer.CreatePart(nil)
	if err != nil {
		return fmt.Errorf("error creating text part: %v", err)
	}
	fmt.Fprintf(textPart, "Content-Type: text/html; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(textPart, body.String())

	// Add attachment
	attachmentPart, err := writer.CreatePart(nil)
	if err != nil {
		return fmt.Errorf("error creating attachment part: %v", err)
	}

	// Convert report data to appropriate format
	var attachmentData []byte
	var filename string
	switch schedule.Format {
	case "json":
		attachmentData, err = json.MarshalIndent(report.Data, "", "  ")
		filename = fmt.Sprintf("%s_%s.json", schedule.Type, time.Now().Format("20060102"))
	case "csv":
		attachmentData = []byte(convertToCSV(report.Data))
		filename = fmt.Sprintf("%s_%s.csv", schedule.Type, time.Now().Format("20060102"))
	case "pdf":
		// Implement PDF generation
		attachmentData, err = generatePDF(report.Data)
		filename = fmt.Sprintf("%s_%s.pdf", schedule.Type, time.Now().Format("20060102"))
	default:
		return fmt.Errorf("unsupported report format: %s", schedule.Format)
	}

	if err != nil {
		return fmt.Errorf("error preparing attachment: %v", err)
	}

	// Add attachment headers
	fmt.Fprintf(attachmentPart, "Content-Type: application/octet-stream\r\n")
	fmt.Fprintf(attachmentPart, "Content-Disposition: attachment; filename=%s\r\n", filename)
	fmt.Fprintf(attachmentPart, "Content-Transfer-Encoding: base64\r\n\r\n")

	// Add attachment content
	encoder := base64.NewEncoder(base64.StdEncoding, attachmentPart)
	encoder.Write(attachmentData)
	encoder.Close()

	writer.Close()

	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%d", s.smtpHost, smtpPort)
	if err := smtp.SendMail(addr, auth, s.fromEmail, recipients, message.Bytes()); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

// generateReportSummary creates a summary of the report for the email body
func generateReportSummary(report *Report) string {
	switch report.Type {
	case "user_activity":
		data := report.Data.(map[string]interface{})
		summary := data["summary"].(map[string]interface{})
		return fmt.Sprintf("Total Events: %d\nTotal Metrics: %d\nTotal Errors: %d",
			len(data["events"].([]interface{})),
			len(data["metrics"].([]interface{})),
			len(data["errors"].([]interface{})))
	case "location":
		data := report.Data.(map[string]interface{})
		summary := data["summary"].(map[string]interface{})
		return fmt.Sprintf("Total Locations: %d\nTotal Distance: %.2f km\nAverage Speed: %.2f km/h",
			summary["total_points"].(int),
			summary["total_distance"].(float64),
			summary["average_speed"].(float64))
	case "system_health":
		data := report.Data.(map[string]interface{})
		summary := data["summary"].(map[string]interface{})
		return fmt.Sprintf("Total Metrics: %d\nTotal Errors: %d",
			len(data["metrics"].([]interface{})),
			len(data["errors"].([]interface{})))
	default:
		return "Report summary not available"
	}
}

// convertToCSV converts report data to CSV format
func convertToCSV(data interface{}) string {
	// Implement CSV conversion logic
	// This is a placeholder implementation
	return "CSV conversion not implemented"
}

// generatePDF generates a PDF from report data
func generatePDF(data interface{}) ([]byte, error) {
	// Implement PDF generation logic
	// This is a placeholder implementation
	return nil, fmt.Errorf("PDF generation not implemented")
}

// Email template for reports
const reportEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .header { background-color: #f8f9fa; padding: 20px; }
        .content { padding: 20px; }
        .footer { background-color: #f8f9fa; padding: 20px; text-align: center; }
    </style>
</head>
<body>
    <div class="header">
        <h2>{{.ReportName}}</h2>
        <p>Generated for {{.TenantName}} on {{.GeneratedAt.Format "January 2, 2006 15:04:05"}}</p>
    </div>
    
    <div class="content">
        <h3>Report Summary</h3>
        <pre>{{.ReportSummary}}</pre>
        
        <p>Please find the detailed report attached to this email.</p>
    </div>
    
    <div class="footer">
        <p>This is an automated report from the Chat System.</p>
    </div>
</body>
</html>
` 