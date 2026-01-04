package security

import (
	"testing"
)

func TestNewAPIKeyService(t *testing.T) {
	service := NewAPIKeyService("test-secret")
	if service == nil {
		t.Fatal("NewAPIKeyService returned nil")
	}
	if service.secret != "test-secret" {
		t.Errorf("Expected secret 'test-secret', got '%s'", service.secret)
	}
}

func TestAPIKeyService_GenerateAPIKey(t *testing.T) {
	service := NewAPIKeyService("test-secret")

	key, err := service.GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey returned error: %v", err)
	}
	if key == "" {
		t.Error("Generated API key should not be empty")
	}
	if len(key) < 10 {
		t.Error("Generated API key should be reasonably long")
	}
	// Check prefix
	if len(key) < 3 || key[:3] != "ur_" {
		t.Error("API key should have 'ur_' prefix")
	}
}

func TestAPIKeyService_ValidateAPIKey(t *testing.T) {
	service := NewAPIKeyService("test-secret")

	key, err := service.GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey returned error: %v", err)
	}

	// Valid key should pass
	if !service.ValidateAPIKey(key) {
		t.Error("Valid API key should pass validation")
	}

	// Invalid key should fail
	if service.ValidateAPIKey("invalid-key") {
		t.Error("Invalid API key should fail validation")
	}

	// Empty key should fail
	if service.ValidateAPIKey("") {
		t.Error("Empty API key should fail validation")
	}
}

func TestExtractAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "Bearer token",
			header:   "Bearer ur_abc123",
			expected: "ur_abc123",
		},
		{
			name:     "Direct key",
			header:   "ur_abc123",
			expected: "ur_abc123",
		},
		{
			name:     "Empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "Multiple spaces",
			header:   "Bearer   ur_abc123",
			expected: "ur_abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractAPIKey(tt.header)
			if got != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

