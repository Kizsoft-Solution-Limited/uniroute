package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/email"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type EmailTestHandler struct {
	emailService *email.EmailService
	logger       zerolog.Logger
}

func NewEmailTestHandler(emailService *email.EmailService, logger zerolog.Logger) *EmailTestHandler {
	return &EmailTestHandler{
		emailService: emailService,
		logger:       logger,
	}
}

type TestEmailRequest struct {
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject,omitempty"`
	Message string `json:"message,omitempty"`
}

func (h *EmailTestHandler) HandleTestEmail(c *gin.Context) {
	var req TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	userEmail := "admin"
	if emailValue, exists := c.Get("user_email"); exists {
		if email, ok := emailValue.(string); ok {
			userEmail = email
		}
	}

	subject := req.Subject
	if subject == "" {
		subject = "UniRoute SMTP Test Email"
	}

	message := req.Message
	if message == "" {
		message = `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%); color: white; padding: 30px; border-radius: 8px 8px 0 0; text-align: center; }
				.content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; border-top: none; }
				.success { background: #d1fae5; border-left: 4px solid #10b981; padding: 15px; margin: 20px 0; border-radius: 4px; }
				.footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e5e7eb; font-size: 12px; color: #6b7280; text-align: center; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1 style="margin: 0; font-size: 28px;">SMTP Test Successful</h1>
				</div>
				<div class="content">
					<div class="success">
						<strong>Congratulations!</strong><br>
						Your SMTP configuration is working correctly. This is a test email from UniRoute.
					</div>
					<p>If you received this email, it means:</p>
					<ul>
						<li>SMTP server connection is working</li>
						<li>Authentication credentials are valid</li>
						<li>Email service is properly configured</li>
					</ul>
					<p><strong>Test Details:</strong></p>
					<ul>
						<li>Sent by: ` + userEmail + `</li>
						<li>Timestamp: ` + time.Now().Format(time.RFC1123) + `</li>
					</ul>
					<p>You can now use the email service for:</p>
					<ul>
						<li>Email verification</li>
						<li>Password reset emails</li>
						<li>Welcome emails</li>
					</ul>
				</div>
				<div class="footer">
					<p>© ` + fmt.Sprintf("%d", time.Now().Year()) + ` UniRoute. All rights reserved.</p>
					<p>This is an automated test email.</p>
				</div>
			</div>
		</body>
		</html>
		`
	}

	h.logger.Info().
		Str("to", req.To).
		Str("from", userEmail).
		Msg("Sending test email")

	err := h.emailService.SendEmail(req.To, subject, message)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("to", req.To).
			Msg("Failed to send test email")

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send test email",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("to", req.To).
		Msg("Test email sent successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent successfully",
		"to":      req.To,
		"note":    "Please check your inbox (and spam folder) for the test email",
	})
}

func (h *EmailTestHandler) HandleGetEmailConfig(c *gin.Context) {
	config := gin.H{
		"configured": h.emailService != nil,
		"status":      "not configured",
	}

	if h.emailService != nil {
		smtpConfig := h.emailService.GetConfig()
		config["status"] = "configured"
		config["smtp"] = smtpConfig
		config["note"] = "SMTP is configured. Use POST /admin/email/test to verify it's working."
		config["troubleshooting"] = gin.H{
			"check_mailtrap": "Verify your Mailtrap inbox at https://mailtrap.io",
			"check_credentials": "Ensure SMTP_USERNAME and SMTP_PASSWORD match your Mailtrap credentials",
			"check_port": "Mailtrap typically uses port 587 (STARTTLS) or 2525; use 465 for SSL",
			"check_host": "Mailtrap sandbox host: sandbox.smtp.mailtrap.io",
			"check_encryption": "For port 465 set SMTP_ENCRYPTION=ssl or SMTP_SECURE=true",
		}
	} else {
		config["note"] = "SMTP is not configured. Set SMTP_HOST, SMTP_PORT, SMTP_USERNAME, and SMTP_PASSWORD environment variables."
		config["required_env_vars"] = []string{
			"SMTP_HOST (e.g., sandbox.smtp.mailtrap.io)",
			"SMTP_PORT (e.g., 587 or 2525)",
			"SMTP_USERNAME (your Mailtrap username)",
			"SMTP_PASSWORD (your Mailtrap password)",
			"SMTP_FROM (e.g., noreply@uniroute.co)",
			"SMTP_ENCRYPTION (optional: ssl or tls for port 465; or SMTP_SECURE=true)",
		}
	}

	c.JSON(http.StatusOK, config)
}
