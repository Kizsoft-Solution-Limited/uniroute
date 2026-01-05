package tunnel_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
)

var (
	NewDomainManager = tunnel.NewDomainManager
	NewSubdomainPool = tunnel.NewSubdomainPool
	NewDNSValidator  = tunnel.NewDNSValidator
)

func TestNewDomainManager(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewDomainManager("uniroute.dev", logger)

	assert.NotNil(t, manager)
	// Note: baseDomain and subdomainPool are unexported, so we can't test them directly
	// We test the exported behavior instead
}

func TestDomainManager_GetPublicURL(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewDomainManager("uniroute.dev", logger)

	tests := []struct {
		name      string
		subdomain string
		port      int
		https     bool
		expected  string
	}{
		{"HTTP with domain", "abc123", 8080, false, "http://abc123.uniroute.dev"},
		{"HTTPS with domain", "abc123", 8080, true, "https://abc123.uniroute.dev"},
		{"HTTP localhost", "abc123", 8080, false, "http://abc123.localhost:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "HTTP localhost" {
				// Test without base domain
				managerNoDomain := NewDomainManager("", logger)
				result := managerNoDomain.GetPublicURL(tt.subdomain, tt.port, tt.https)
				assert.Equal(t, tt.expected, result)
			} else {
				result := manager.GetPublicURL(tt.subdomain, tt.port, tt.https)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSubdomainPool_Allocate(t *testing.T) {
	logger := zerolog.Nop()
	pool := NewSubdomainPool(logger)

	subdomain1 := pool.Allocate()
	subdomain2 := pool.Allocate()

	assert.NotEmpty(t, subdomain1)
	assert.NotEmpty(t, subdomain2)
	assert.NotEqual(t, subdomain1, subdomain2)
	assert.Len(t, subdomain1, 8) // 8 character subdomain
}

func TestSubdomainPool_Release(t *testing.T) {
	logger := zerolog.Nop()
	pool := NewSubdomainPool(logger)

	subdomain := pool.Allocate()
	assert.NotEmpty(t, subdomain)

	pool.Release(subdomain)

	// Allocate again - should be able to reuse (though unlikely to get same one)
	subdomain2 := pool.Allocate()
	assert.NotEmpty(t, subdomain2)
}

func TestDNSValidator_ValidateTXTRecord(t *testing.T) {
	logger := zerolog.Nop()
	validator := NewDNSValidator(logger)

	// Test with a real domain (this will make actual DNS lookup)
	ctx := context.Background()

	// This test requires internet connection and may fail in some environments
	// Using a well-known domain for testing
	valid, err := validator.ValidateTXTRecord(ctx, "_dmarc.google.com", "")

	// We don't assert on the result as it depends on DNS
	// Just check that the function doesn't panic
	assert.NoError(t, err)
	_ = valid
}

func TestDomainManager_ValidateCustomDomain(t *testing.T) {
	logger := zerolog.Nop()
	manager := NewDomainManager("uniroute.dev", logger)

	ctx := context.Background()

	// Test with a real domain
	err := manager.ValidateCustomDomain(ctx, "google.com")

	// Should succeed for a valid domain
	assert.NoError(t, err)

	// Test with invalid domain
	err = manager.ValidateCustomDomain(ctx, "")
	assert.Error(t, err)

	err = manager.ValidateCustomDomain(ctx, "not-a-domain")
	assert.Error(t, err)
}
