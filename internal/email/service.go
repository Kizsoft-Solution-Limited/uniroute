package email

import (
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type EmailService struct {
	logger zerolog.Logger
}

func (s *EmailService) GetConfig() map[string]interface{} {
	host := os.Getenv("SMTP_HOST")
	port := getEnvAsInt("SMTP_PORT", 0)
	username := os.Getenv("SMTP_USERNAME")
	from := os.Getenv("SMTP_FROM")
	passwordSet := os.Getenv("SMTP_PASSWORD") != ""

	return map[string]interface{}{
		"host":       host,
		"port":       port,
		"username":   username,
		"from":       from,
		"configured": host != "" && username != "" && passwordSet,
	}
}

func NewEmailService(logger zerolog.Logger) *EmailService {
	return &EmailService{
		logger: logger,
	}
}

func (s *EmailService) getSMTPConfig() (host string, port int, username, password, from string, configured bool) {
	host = strings.TrimSpace(os.Getenv("SMTP_HOST"))
	portStr := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	username = strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	password = strings.TrimSpace(os.Getenv("SMTP_PASSWORD"))
	from = strings.TrimSpace(os.Getenv("SMTP_FROM"))

	if portStr == "" {
		port = 587
	} else {
		if p, err := strconv.Atoi(portStr); err == nil {
			port = p
		} else {
			port = 587
		}
	}

	if from == "" {
		from = "noreply@uniroute.co"
	}

	configured = host != "" && username != "" && password != ""
	return
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	host, port, username, password, from, configured := s.getSMTPConfig()

	if !configured {
		s.logger.Warn().Msg("SMTP not configured, skipping email send - check environment variables")
		return fmt.Errorf("SMTP not configured: set SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD environment variables")
	}

	auth := smtp.PlainAuth("", username, password, host)
	message := s.buildMessage(from, to, subject, body)
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	if err != nil {
		s.logger.Error().Err(err).Str("to", to).Str("host", host).Int("port", port).Msg("Failed to send email")
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (s *EmailService) buildMessage(from, to, subject, body string) string {
	headers := map[string]string{
		"From":         from,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	return message
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (s *EmailService) buildEmailTemplate(title, greeting, mainContent, buttonText, buttonURL, footerText string) string {
	year := fmt.Sprintf("%d", time.Now().Year())
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en" xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta name="x-apple-disable-message-reformatting">
	<title>%s</title>
	<!--[if mso]>
	<style type="text/css">
		table {border-collapse:collapse;border-spacing:0;margin:0;}
		div, td {padding:0;}
		div {margin:0 !important;}
	</style>
	<![endif]-->
</head>
<body style="margin: 0; padding: 0; background-color: #f4f4f4; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
	<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin: 0; padding: 0; width: 100%%; background-color: #f4f4f4;">
		<tr>
			<td align="center" style="padding: 40px 20px;">
				<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="600" style="max-width: 600px; width: 100%%; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
					<!-- Header -->
					<tr>
						<td style="background: linear-gradient(135deg, #3b82f6 0%%, #8b5cf6 100%%); padding: 40px 30px; text-align: center;">
							<h1 style="margin: 0; font-size: 28px; font-weight: 700; color: #ffffff; line-height: 1.2;">%s</h1>
						</td>
					</tr>
					<!-- Content -->
					<tr>
						<td style="padding: 40px 30px;">
							<div style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #1f2937;">%s</div>
							%s
						</td>
					</tr>
					<!-- Footer -->
					<tr>
						<td style="padding: 30px; background-color: #f9fafb; border-top: 1px solid #e5e7eb;">
							<p style="margin: 0; font-size: 12px; line-height: 1.5; color: #9ca3af; text-align: center;">©&nbsp;%s&nbsp;%s&nbsp;UniRoute. All rights reserved.</p>
						</td>
					</tr>
				</table>
			</td>
		</tr>
	</table>
</body>
</html>
`, title, title, greeting, mainContent, year, footerText)
}

func (s *EmailService) buildButton(text, url string) string {
	return fmt.Sprintf(`
<table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 24px 0;">
	<tr>
		<td align="center" style="padding: 0;">
			<table role="presentation" cellspacing="0" cellpadding="0" border="0">
				<tr>
					<td align="center" style="background: linear-gradient(135deg, #3b82f6 0%%, #8b5cf6 100%%); border-radius: 6px;">
						<a href="%s" style="display: inline-block; padding: 14px 32px; font-size: 16px; font-weight: 600; color: #ffffff; text-decoration: none; border-radius: 6px; line-height: 1.5;">%s</a>
					</td>
				</tr>
			</table>
		</td>
	</tr>
</table>
`, url, text)
}

func (s *EmailService) buildLinkFallback(text, url string) string {
	return fmt.Sprintf(`
<p style="margin: 16px 0 0 0; font-size: 14px; line-height: 1.6; color: #6b7280; word-break: break-all;">
	If the button doesn't work, copy and paste this link into your browser:<br>
	<a href="%s" style="color: #3b82f6; text-decoration: underline;">%s</a>
</p>
`, url, url)
}

func (s *EmailService) SendVerificationEmail(to, name, token, frontendURL string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	subject := "Verify your UniRoute email address"

	greeting := fmt.Sprintf("Hi %s,<br><br>Thank you for signing up for UniRoute! Please verify your email address to complete your registration and start using our platform.", name)

	footerNote := "<p style=\"margin: 24px 0 0 0; font-size: 14px; line-height: 1.6; color: #6b7280;\">This verification link will expire in 24 hours. If you didn't create an account with UniRoute, you can safely ignore this email.</p>"

	mainContent := s.buildButton("Verify Email Address", verificationURL) + s.buildLinkFallback("Verify Email Address", verificationURL) + footerNote

	footerText := "You're receiving this email because you signed up for a UniRoute account."

	body := s.buildEmailTemplate(
		subject,
		greeting,
		mainContent,
		"",
		"",
		footerText,
	)

	return s.SendEmail(to, subject, body)
}

func (s *EmailService) SendPasswordResetEmail(to, name, token, frontendURL string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimSuffix(frontendURL, "/"), token)

	subject := "Reset your UniRoute password"

	greeting := fmt.Sprintf("Hi %s,<br><br>We received a request to reset your password. Click the button below to create a new password:", name)

	footerNote := "<p style=\"margin: 24px 0 0 0; font-size: 14px; line-height: 1.6; color: #6b7280;\">This password reset link will expire in 1 hour. If you didn't request a password reset, you can safely ignore this email. Your password will remain unchanged.</p>"

	mainContent := s.buildButton("Reset Password", resetURL) + s.buildLinkFallback("Reset Password", resetURL) + footerNote

	footerText := "You're receiving this email because a password reset was requested for your UniRoute account."

	body := s.buildEmailTemplate(
		subject,
		greeting,
		mainContent,
		"",
		"",
		footerText,
	)

	return s.SendEmail(to, subject, body)
}

func (s *EmailService) SendSeedAdminPasswordEmail(to, name, password string) error {
	subject := "UniRoute: Your admin account password"

	greeting := fmt.Sprintf("Hi %s,<br><br>Your UniRoute seed admin account has been created. Use the credentials below to sign in:", name)

	mainContent := fmt.Sprintf(`
<p style="margin: 0 0 16px 0; font-size: 16px; line-height: 1.6; color: #1f2937;"><strong>Email:</strong> %s</p>
<p style="margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #1f2937;"><strong>Password:</strong> %s</p>
<p style="margin: 0; font-size: 14px; line-height: 1.6; color: #6b7280;">Save this password securely. You can change it after logging in.</p>
`, to, password)

	footerText := "You're receiving this email because seed admin was enabled at startup for this UniRoute instance."

	body := s.buildEmailTemplate(subject, greeting, mainContent, "", "", footerText)

	return s.SendEmail(to, subject, body)
}

func (s *EmailService) buildFeatureBox(icon, title, description string) string {
	return fmt.Sprintf(`
<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin: 12px 0; background-color: #f9fafb; border-left: 4px solid #3b82f6; border-radius: 4px;">
	<tr>
		<td style="padding: 16px 20px;">
			<p style="margin: 0 0 4px 0; font-size: 16px; font-weight: 600; color: #1f2937; line-height: 1.4;">%s %s</p>
			<p style="margin: 0; font-size: 14px; line-height: 1.5; color: #6b7280;">%s</p>
		</td>
	</tr>
</table>
`, icon, title, description)
}

func (s *EmailService) SendWelcomeEmail(to, name, dashboardURL string) error {
	subject := "Welcome to UniRoute!"

	greeting := fmt.Sprintf("Hi %s,<br><br>Congratulations! Your email has been verified and your UniRoute account is now fully activated.", name)

	features := s.buildFeatureBox("", "Get Started", "Access your dashboard and start configuring your routes.") +
		s.buildFeatureBox("", "API Keys", "Generate API keys to authenticate your requests.") +
		s.buildFeatureBox("", "Monitor Traffic", "Track your API usage and monitor performance.")

	mainContent := fmt.Sprintf(`
<p style="margin: 0 0 24px 0; font-size: 16px; line-height: 1.6; color: #1f2937;">You're all set to start using UniRoute to manage your AI model routing and traffic. Here's what you can do next:</p>
%s
%s
<p style="margin: 24px 0 16px 0; font-size: 16px; line-height: 1.6; color: #1f2937;">If you have any questions or need help getting started, don't hesitate to reach out to our support team.</p>
<p style="margin: 0; font-size: 16px; line-height: 1.6; color: #1f2937;">Welcome aboard!<br><strong style="color: #1f2937;">The UniRoute Team</strong></p>
`, features, s.buildButton("Go to Dashboard", dashboardURL))

	footerText := "You're receiving this email because you verified your email address on UniRoute."

	body := s.buildEmailTemplate(
		subject,
		greeting,
		mainContent,
		"",
		"",
		footerText,
	)

	return s.SendEmail(to, subject, body)
}
