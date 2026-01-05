package testutil

import (
	"testing"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/email"
	"github.com/rs/zerolog"
)

// CreateMockEmailService creates a mock email service
func CreateMockEmailService(t *testing.T) *email.EmailService {
	logger := zerolog.Nop()
	return email.NewEmailService(logger)
}

// MockEmailService is a mock implementation of email service
// This can be extended to track sent emails for testing
type MockEmailService struct {
	SentEmails []SentEmail
}

type SentEmail struct {
	To      string
	Subject string
	Body    string
}

// SendEmail mocks sending an email
func (m *MockEmailService) SendEmail(to, subject, body string) error {
	m.SentEmails = append(m.SentEmails, SentEmail{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	return nil
}

// GetSentEmails returns all sent emails
func (m *MockEmailService) GetSentEmails() []SentEmail {
	return m.SentEmails
}

// ClearSentEmails clears the sent emails list
func (m *MockEmailService) ClearSentEmails() {
	m.SentEmails = []SentEmail{}
}
