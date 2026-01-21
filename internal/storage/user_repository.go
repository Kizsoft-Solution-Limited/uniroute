package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
)

// UserRepository handles user database operations
type UserRepository struct {
	client *PostgresClient
	logger zerolog.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(client *PostgresClient, logger zerolog.Logger) *UserRepository {
	return &UserRepository{
		client: client,
		logger: logger,
	}
}

// CreateUser creates a new user
// SECURITY: This function ALWAYS sets roles to ['user'] - it cannot be overridden
// This prevents privilege escalation attacks during registration
func (r *UserRepository) CreateUser(ctx context.Context, email, password, name string) (*User, error) {
	// Check if user already exists
	existing, err := r.GetUserByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// SECURITY: Insert user with roles ALWAYS set to ['user'] - hardcoded, cannot be changed
	// The ARRAY['user']::TEXT[] is hardcoded in SQL to prevent any role injection
	// Even if someone tries to pass roles as a parameter, it will be ignored
	query := `
		INSERT INTO users (id, email, name, password_hash, email_verified, roles, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, false, ARRAY['user']::TEXT[], NOW(), NOW())
		RETURNING id, email, name, password_hash, email_verified, roles, created_at, updated_at
	`

	var user User
	err = r.client.pool.QueryRow(ctx, query, email, name, string(hashedPassword)).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.Roles,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	// Use COALESCE to handle NULL values (for existing users before migration)
	// Treat NULL as false (not verified), and default roles to ['user'] if NULL
	query := `
		SELECT id, email, name, password_hash, COALESCE(email_verified, false) as email_verified, 
		       COALESCE(roles, ARRAY['user']::TEXT[]) as roles, routing_strategy, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.client.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.Roles,
		&user.RoutingStrategy,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	// Use COALESCE to handle NULL values (for existing users before migration)
	// Treat NULL as false (not verified), and default roles to ['user'] if NULL
	query := `
		SELECT id, email, name, password_hash, COALESCE(email_verified, false) as email_verified, 
		       COALESCE(roles, ARRAY['user']::TEXT[]) as roles, routing_strategy, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.client.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.Roles,
		&user.RoutingStrategy,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// VerifyPassword verifies a password against a hash
func (r *UserRepository) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err = r.client.pool.Exec(ctx, query, string(hashedPassword), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// CreatePasswordResetToken creates a password reset token
func (r *UserRepository) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
	`

	_, err := r.client.pool.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

// GetPasswordResetToken retrieves a password reset token
func (r *UserRepository) GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token = $1 AND used = false AND expires_at > NOW()
	`

	var resetToken PasswordResetToken
	err := r.client.pool.QueryRow(ctx, query, token).Scan(
		&resetToken.ID,
		&resetToken.UserID,
		&resetToken.Token,
		&resetToken.ExpiresAt,
		&resetToken.Used,
		&resetToken.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return &resetToken, nil
}

// MarkPasswordResetTokenAsUsed marks a password reset token as used
func (r *UserRepository) MarkPasswordResetTokenAsUsed(ctx context.Context, token string) error {
	query := `
		UPDATE password_reset_tokens
		SET used = true
		WHERE token = $1
	`

	_, err := r.client.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// CreateEmailVerificationToken creates an email verification token
func (r *UserRepository) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	query := `
		INSERT INTO email_verification_tokens (id, user_id, token, expires_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
	`

	_, err := r.client.pool.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create email verification token: %w", err)
	}

	return nil
}

// GetEmailVerificationToken retrieves an email verification token
func (r *UserRepository) GetEmailVerificationToken(ctx context.Context, token string) (*EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used, created_at
		FROM email_verification_tokens
		WHERE token = $1 AND used = false AND expires_at > NOW()
	`

	var verificationToken EmailVerificationToken
	err := r.client.pool.QueryRow(ctx, query, token).Scan(
		&verificationToken.ID,
		&verificationToken.UserID,
		&verificationToken.Token,
		&verificationToken.ExpiresAt,
		&verificationToken.Used,
		&verificationToken.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, fmt.Errorf("failed to get email verification token: %w", err)
	}

	return &verificationToken, nil
}

// MarkEmailVerificationTokenAsUsed marks an email verification token as used
func (r *UserRepository) MarkEmailVerificationTokenAsUsed(ctx context.Context, token string) error {
	query := `
		UPDATE email_verification_tokens
		SET used = true
		WHERE token = $1
	`

	_, err := r.client.pool.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// VerifyEmail marks a user's email as verified
func (r *UserRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified = true, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.client.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	return nil
}

// UpdateUserProfile updates a user's profile (name only - roles are NOT updated)
func (r *UserRepository) UpdateUserProfile(ctx context.Context, userID uuid.UUID, name string) error {
	query := `
		UPDATE users
		SET name = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.client.pool.Exec(ctx, query, name, userID)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	// Check if user exists
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateUserRoles updates a user's roles (admin only operation)
func (r *UserRepository) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roles []string) error {
	// Validate roles
	for _, role := range roles {
		if role != "user" && role != "admin" {
			return fmt.Errorf("invalid role: %s (must be 'user' or 'admin')", role)
		}
	}

	query := `
		UPDATE users
		SET roles = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.client.pool.Exec(ctx, query, roles, userID)
	if err != nil {
		return fmt.Errorf("failed to update user roles: %w", err)
	}

	// Check if user exists
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetUserRoutingStrategy retrieves a user's routing strategy preference
// Returns empty string if user has no preference (should use default)
func (r *UserRepository) GetUserRoutingStrategy(ctx context.Context, userID uuid.UUID) (string, error) {
	query := `
		SELECT routing_strategy
		FROM users
		WHERE id = $1
	`

	var strategy *string
	err := r.client.pool.QueryRow(ctx, query, userID).Scan(&strategy)
	if err == pgx.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user routing strategy: %w", err)
	}

	if strategy == nil {
		return "", nil // User has no preference, use default
	}
	return *strategy, nil
}

// SetUserRoutingStrategy sets a user's routing strategy preference
// Pass nil to clear the preference (use default)
func (r *UserRepository) SetUserRoutingStrategy(ctx context.Context, userID uuid.UUID, strategy *string) error {
	// Validate strategy if provided
	if strategy != nil {
		validStrategies := map[string]bool{
			"model":    true,
			"cost":     true,
			"latency":  true,
			"balanced": true,
			"custom":   true,
		}
		if !validStrategies[*strategy] {
			return fmt.Errorf("invalid strategy: %s (must be one of: model, cost, latency, balanced, custom)", *strategy)
		}
	}

	query := `
		UPDATE users
		SET routing_strategy = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.client.pool.Exec(ctx, query, strategy, userID)
	if err != nil {
		return fmt.Errorf("failed to set user routing strategy: %w", err)
	}

	// Check if user exists
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsers retrieves all users (admin only)
func (r *UserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, email, name, password_hash, COALESCE(email_verified, false) as email_verified, 
		       COALESCE(roles, ARRAY['user']::TEXT[]) as roles, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.client.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.PasswordHash,
			&user.EmailVerified,
			&user.Roles,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// CountUsers returns the total number of users
func (r *UserRepository) CountUsers(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.client.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// UpdateUserEmailVerified updates the email verified status for a user
func (r *UserRepository) UpdateUserEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	query := `
		UPDATE users
		SET email_verified = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.client.pool.Exec(ctx, query, verified, userID)
	if err != nil {
		return fmt.Errorf("failed to update email verified status: %w", err)
	}

	return nil
}
