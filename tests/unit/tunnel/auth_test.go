package tunnel_test

import (
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/tunnel"
)

// Use package prefix for tunnel functions and types
var (
	NewTokenService = tunnel.NewTokenService
	ErrInvalidToken = tunnel.ErrInvalidToken
	ErrTokenExpired = tunnel.ErrTokenExpired
)

// Type aliases
type TokenInfo = tunnel.TokenInfo

func TestNewTokenService(t *testing.T) {
	logger := zerolog.Nop()
	service := NewTokenService(logger)

	assert.NotNil(t, service)
}

func TestTokenService_GenerateToken(t *testing.T) {
	logger := zerolog.Nop()
	service := NewTokenService(logger)

	token1, hash1, err := service.GenerateToken()
	require.NoError(t, err)
	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, hash1)
	assert.Len(t, token1, 64) // 32 bytes = 64 hex chars

	token2, hash2, err := service.GenerateToken()
	require.NoError(t, err)
	assert.NotEqual(t, token1, token2)
	assert.NotEqual(t, hash1, hash2)
}

func TestTokenService_HashToken(t *testing.T) {
	logger := zerolog.Nop()
	service := NewTokenService(logger)

	token := "test-token-123"
	hash1 := service.HashToken(token)
	hash2 := service.HashToken(token)

	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, token, hash1)
}

func TestTokenService_VerifyToken(t *testing.T) {
	logger := zerolog.Nop()
	service := NewTokenService(logger)

	token, hash, _ := service.GenerateToken()

	assert.True(t, service.VerifyToken(token, hash))
	assert.False(t, service.VerifyToken("wrong-token", hash))
	assert.False(t, service.VerifyToken(token, "wrong-hash"))
}

func TestTokenService_GenerateBcryptHash(t *testing.T) {
	logger := zerolog.Nop()
	service := NewTokenService(logger)

	token := "test-token"
	hash, err := service.GenerateBcryptHash(token)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, token, hash)
}

func TestTokenService_VerifyBcryptHash(t *testing.T) {
	logger := zerolog.Nop()
	service := NewTokenService(logger)

	token := "test-token"
	hash, _ := service.GenerateBcryptHash(token)

	assert.True(t, service.VerifyBcryptHash(token, hash))
	assert.False(t, service.VerifyBcryptHash("wrong-token", hash))
}

func TestTokenService_ValidateToken(t *testing.T) {
	logger := zerolog.Nop()
	service := NewTokenService(logger)

	token, hash, _ := service.GenerateToken()

	// Valid token
	info := &TokenInfo{
		TokenHash: hash,
		IsActive:  true,
	}

	err := service.ValidateToken(token, info)
	assert.NoError(t, err)

	// Inactive token
	info.IsActive = false
	err = service.ValidateToken(token, info)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)

	// Expired token
	expired := time.Now().Add(-1 * time.Hour)
	info.IsActive = true
	info.ExpiresAt = &expired
	err = service.ValidateToken(token, info)
	assert.Error(t, err)
	assert.Equal(t, ErrTokenExpired, err)
}
