package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/smtp"
	"time"

	"saas-chat-system/internal/models"
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
func (s *EmailService) SendReport(recipients []string, schedule *models.ReportSchedule, report *Report) error {
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
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
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
		
		eventCount := 0
		if events, ok := data["events"].([]interface{}); ok {
			eventCount = len(events)
		}
		
		metricCount := 0
		if metrics, ok := data["metrics"].([]interface{}); ok {
			metricCount = len(metrics)
		}
		
		errorCount := 0
		if errors, ok := data["errors"].([]interface{}); ok {
			errorCount = len(errors)
		}
		
		return fmt.Sprintf("Total Events: %d\nTotal Metrics: %d\nTotal Errors: %d\nSummary: %v",
			eventCount, metricCount, errorCount, summary)
			
	case "location":
		data := report.Data.(map[string]interface{})
		locationSummary := data["summary"].(map[string]interface{})
		
		totalPoints := 0
		if pointsVal, ok := locationSummary["total_points"]; ok {
			if points, ok := pointsVal.(int); ok {
				totalPoints = points
			}
		}
		
		totalDistance := 0.0
		if distVal, ok := locationSummary["total_distance"]; ok {
			if dist, ok := distVal.(float64); ok {
				totalDistance = dist
			}
		}
		
		averageSpeed := 0.0
		if speedVal, ok := locationSummary["average_speed"]; ok {
			if speed, ok := speedVal.(float64); ok {
				averageSpeed = speed
			}
		}
		
		return fmt.Sprintf("Total Locations: %d\nTotal Distance: %.2f km\nAverage Speed: %.2f km/h",
			totalPoints, totalDistance, averageSpeed)
			
	case "system_health":
		data := report.Data.(map[string]interface{})
		healthSummary := data["summary"].(map[string]interface{})
		
		metricCount := 0
		if metrics, ok := data["metrics"].([]interface{}); ok {
			metricCount = len(metrics)
		}
		
		errorCount := 0
		if errors, ok := data["errors"].([]interface{}); ok {
			errorCount = len(errors)
		}
		
		return fmt.Sprintf("Total Metrics: %d\nTotal Errors: %d\nSummary: %v",
			metricCount, errorCount, healthSummary)
			
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

// SendTicketEmail sends an email notification for ticket-related events
func (s *EmailService) SendTicketEmail(ctx context.Context, notification *models.TicketNotification) error {
	// Create email template
	tmpl, err := template.New("ticket").Parse(ticketEmailTemplate)
	if err != nil {
		return fmt.Errorf("error parsing email template: %v", err)
	}

	// Get user email from user ID
	recipientEmail, err := s.getUserEmail(ctx, notification.UserID)
	if err != nil {
		return fmt.Errorf("error getting user email: %v", err)
	}

	// Generate email body
	var body bytes.Buffer
	if err := tmpl.Execute(&body, notification); err != nil {
		return fmt.Errorf("error generating email body: %v", err)
	}

	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	
	// Construct email
	message := &bytes.Buffer{}
	fmt.Fprintf(message, "From: %s <%s>\r\n", s.fromName, s.fromEmail)
	fmt.Fprintf(message, "To: %s\r\n", recipientEmail)
	fmt.Fprintf(message, "Subject: %s: %s\r\n", notification.Type, notification.Title)
	fmt.Fprintf(message, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(message, "Content-Type: text/html; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(message, body.String())

	if err := smtp.SendMail(addr, auth, s.fromEmail, []string{recipientEmail}, message.Bytes()); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

// SendForumEmail sends an email notification for forum-related events
func (s *EmailService) SendForumEmail(ctx context.Context, notification *models.ForumNotification) error {
	// Create email template
	tmpl, err := template.New("forum").Parse(forumEmailTemplate)
	if err != nil {
		return fmt.Errorf("error parsing email template: %v", err)
	}

	// Get topic details for the template
	topic, err := s.getForumTopic(ctx, notification.TopicID)
	if err != nil {
		return fmt.Errorf("error getting forum topic: %v", err)
	}

	// Get user email from user ID
	recipientEmail, err := s.getUserEmail(ctx, notification.UserID)
	if err != nil {
		return fmt.Errorf("error getting user email: %v", err)
	}

	// Prepare template data
	templateData := struct {
		*models.ForumNotification
		TopicTitle string
		ForumName  string
		Content    string
		URL        string
		AuthorName string
		CreatedAt  time.Time
		UnsubscribeURL string
	}{
		ForumNotification: notification,
		TopicTitle: topic.Title,
		ForumName: "Forum", // Get this from category
		Content: topic.Content,
		URL: fmt.Sprintf("/forums/topics/%s", notification.TopicID),
		AuthorName: "Author", // Get this from user
		CreatedAt: topic.CreatedAt,
		UnsubscribeURL: fmt.Sprintf("/forums/subscriptions/unsubscribe?topic=%s&user=%s", notification.TopicID, notification.UserID),
	}

	// Generate email body
	var body bytes.Buffer
	if err := tmpl.Execute(&body, templateData); err != nil {
		return fmt.Errorf("error generating email body: %v", err)
	}

	// Send email
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)
	
	// Construct email
	message := &bytes.Buffer{}
	fmt.Fprintf(message, "From: %s <%s>\r\n", s.fromName, s.fromEmail)
	fmt.Fprintf(message, "To: %s\r\n", recipientEmail)
	fmt.Fprintf(message, "Subject: %s: %s\r\n", notification.Type, templateData.TopicTitle)
	fmt.Fprintf(message, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(message, "Content-Type: text/html; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(message, body.String())

	if err := smtp.SendMail(addr, auth, s.fromEmail, []string{recipientEmail}, message.Bytes()); err != nil {
		return fmt.Errorf("error sending email: %v", err)
	}

	return nil
}

// Helper functions for email operations
func (s *EmailService) getUserEmail(ctx context.Context, userID string) (string, error) {
	// Implement this to fetch user email from database
	// For now, return a placeholder
	return "user@example.com", nil
}

// getForumTopic retrieves a forum topic by ID
func (s *EmailService) getForumTopic(ctx context.Context, topicID string) (*models.ForumTopic, error) {
	// Implement this to fetch forum topic details
	// For now, return a placeholder
	return &models.ForumTopic{
		Title:      "Sample Topic",
		Content:    "Sample content",
		CreatedAt:  time.Now(),
	}, nil
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

// Email template for ticket notifications
const ticketEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .header { background-color: #f8f9fa; padding: 20px; }
        .content { padding: 20px; }
        .footer { background-color: #f8f9fa; padding: 20px; text-align: center; }
        .button { background-color: #4CAF50; border: none; color: white; padding: 15px 32px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; margin: 4px 2px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="header">
        <h2>Ticket: {{.Title}}</h2>
    </div>
    
    <div class="content">
        <p>Hello {{.UserName}},</p>
        <p>{{.Message}}</p>
        
        <p>Ticket Details:</p>
        <ul>
            <li><strong>Ticket ID:</strong> {{.ID}}</li>
            <li><strong>Status:</strong> {{.Status}}</li>
            <li><strong>Priority:</strong> {{.Priority}}</li>
            <li><strong>Created:</strong> {{.CreatedAt}}</li>
            <li><strong>Updated:</strong> {{.UpdatedAt}}</li>
        </ul>
        
        <p><a href="{{.URL}}" class="button">View Ticket</a></p>
    </div>
    
    <div class="footer">
        <p>This is an automated notification from the Support System.</p>
    </div>
</body>
</html>
`

// Email template for forum notifications
const forumEmailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .header { background-color: #f8f9fa; padding: 20px; }
        .content { padding: 20px; }
        .footer { background-color: #f8f9fa; padding: 20px; text-align: center; }
        .button { background-color: #4CAF50; border: none; color: white; padding: 15px 32px; text-align: center; text-decoration: none; display: inline-block; font-size: 16px; margin: 4px 2px; cursor: pointer; }
    </style>
</head>
<body>
    <div class="header">
        <h2>{{.Type}}: {{.Title}}</h2>
    </div>
    
    <div class="content">
        <p>Hello {{.UserName}},</p>
        <p>{{.Message}}</p>
        
        <p>Topic Details:</p>
        <ul>
            <li><strong>Forum:</strong> {{.ForumName}}</li>
            <li><strong>Topic:</strong> {{.Title}}</li>
            <li><strong>Posted by:</strong> {{.AuthorName}}</li>
            <li><strong>Posted at:</strong> {{.CreatedAt}}</li>
        </ul>
        
        <div style="margin: 20px; padding: 15px; background-color: #f5f5f5; border-left: 4px solid #ccc;">
            {{.Content}}
        </div>
        
        <p><a href="{{.URL}}" class="button">View Topic</a></p>
    </div>
    
    <div class="footer">
        <p>This is an automated notification from the Forum System.</p>
        <p>You are receiving this email because you are subscribed to this topic or forum.</p>
        <p><a href="{{.UnsubscribeURL}}">Unsubscribe</a></p>
    </div>
</body>
</html>
`
