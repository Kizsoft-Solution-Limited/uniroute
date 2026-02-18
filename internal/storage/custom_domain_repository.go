package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDomainNotFound      = errors.New("domain not found")
	ErrDomainAlreadyExists = errors.New("domain already exists")
)

type CustomDomain struct {
	ID              uuid.UUID  `db:"id"`
	UserID          uuid.UUID  `db:"user_id"`
	Domain          string     `db:"domain"`
	Verified        bool       `db:"verified"`
	VerificationToken *string  `db:"verification_token"`
	DNSConfigured   bool       `db:"dns_configured"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type CustomDomainRepository struct {
	pool *pgxpool.Pool
}

func NewCustomDomainRepository(pool *pgxpool.Pool) *CustomDomainRepository {
	return &CustomDomainRepository{
		pool: pool,
	}
}

func (r *CustomDomainRepository) CreateDomain(ctx context.Context, userID uuid.UUID, domain string) (*CustomDomain, error) {
	existing, err := r.GetDomainByUserAndDomain(ctx, userID, domain)
	if err == nil && existing != nil {
		return nil, ErrDomainAlreadyExists
	}

	query := `
		INSERT INTO custom_domains (id, user_id, domain, verified, dns_configured, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, false, false, NOW(), NOW())
		RETURNING id, user_id, domain, verified, verification_token, dns_configured, created_at, updated_at
	`

	var d CustomDomain
	err = r.pool.QueryRow(ctx, query, userID, domain).Scan(
		&d.ID,
		&d.UserID,
		&d.Domain,
		&d.Verified,
		&d.VerificationToken,
		&d.DNSConfigured,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	return &d, nil
}

func (r *CustomDomainRepository) GetDomainByID(ctx context.Context, domainID uuid.UUID) (*CustomDomain, error) {
	query := `
		SELECT id, user_id, domain, verified, verification_token, dns_configured, created_at, updated_at
		FROM custom_domains
		WHERE id = $1
	`

	var d CustomDomain
	err := r.pool.QueryRow(ctx, query, domainID).Scan(
		&d.ID,
		&d.UserID,
		&d.Domain,
		&d.Verified,
		&d.VerificationToken,
		&d.DNSConfigured,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrDomainNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return &d, nil
}

func (r *CustomDomainRepository) GetDomainByUserAndDomain(ctx context.Context, userID uuid.UUID, domain string) (*CustomDomain, error) {
	query := `
		SELECT id, user_id, domain, verified, verification_token, dns_configured, created_at, updated_at
		FROM custom_domains
		WHERE user_id = $1 AND domain = $2
	`

	var d CustomDomain
	err := r.pool.QueryRow(ctx, query, userID, domain).Scan(
		&d.ID,
		&d.UserID,
		&d.Domain,
		&d.Verified,
		&d.VerificationToken,
		&d.DNSConfigured,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	return &d, nil
}

func (r *CustomDomainRepository) ListDomainsByUser(ctx context.Context, userID uuid.UUID) ([]*CustomDomain, error) {
	query := `
		SELECT id, user_id, domain, verified, verification_token, dns_configured, created_at, updated_at
		FROM custom_domains
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	var domains []*CustomDomain
	for rows.Next() {
		var d CustomDomain
		err := rows.Scan(
			&d.ID,
			&d.UserID,
			&d.Domain,
			&d.Verified,
			&d.VerificationToken,
			&d.DNSConfigured,
			&d.CreatedAt,
			&d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, &d)
	}

	return domains, nil
}

func (r *CustomDomainRepository) UpdateDomain(ctx context.Context, domainID uuid.UUID, verified *bool, dnsConfigured *bool) error {
	query := `
		UPDATE custom_domains
		SET updated_at = NOW()
	`
	args := []interface{}{domainID}
	argIndex := 2

	if verified != nil {
		query += fmt.Sprintf(", verified = $%d", argIndex)
		args = append(args, *verified)
		argIndex++
	}

	if dnsConfigured != nil {
		query += fmt.Sprintf(", dns_configured = $%d", argIndex)
		args = append(args, *dnsConfigured)
		argIndex++
	}

	query += " WHERE id = $1"

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrDomainNotFound
	}

	return nil
}

func (r *CustomDomainRepository) DeleteDomain(ctx context.Context, domainID uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM custom_domains
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.pool.Exec(ctx, query, domainID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrDomainNotFound
	}

	return nil
}
