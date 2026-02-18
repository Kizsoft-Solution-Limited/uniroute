package tunnel

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type TokenService struct {
	logger zerolog.Logger
}

func NewTokenService(logger zerolog.Logger) *TokenService {
	return &TokenService{
		logger: logger,
	}
}

func (ts *TokenService) GenerateToken() (string, string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate token: %w", err)
	}

	token := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(token))
	hashStr := hex.EncodeToString(hash[:])

	return token, hashStr, nil
}

// HashToken hashes a token for storage
func (ts *TokenService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// VerifyToken verifies a token against a hash
func (ts *TokenService) VerifyToken(token, hash string) bool {
	computedHash := ts.HashToken(token)
	return computedHash == hash
}

func (ts *TokenService) GenerateBcryptHash(token string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash token: %w", err)
	}
	return string(hash), nil
}

// VerifyBcryptHash verifies a token against a bcrypt hash
func (ts *TokenService) VerifyBcryptHash(token, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)) == nil
}

type TokenInfo struct {
	TokenHash  string
	Name       string
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	LastUsedAt *time.Time
	IsActive   bool
}

func (ts *TokenService) ValidateToken(token string, tokenInfo *TokenInfo) error {
	if tokenInfo == nil {
		return ErrInvalidToken
	}

	if !tokenInfo.IsActive {
		return ErrInvalidToken
	}

	if tokenInfo.ExpiresAt != nil && time.Now().After(*tokenInfo.ExpiresAt) {
		return ErrTokenExpired
	}

	if !ts.VerifyToken(token, tokenInfo.TokenHash) {
		return ErrInvalidToken
	}

	return nil
}
