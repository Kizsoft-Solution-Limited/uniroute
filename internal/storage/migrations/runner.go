package migrations

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

//go:embed sql/*.sql
var sqlFS embed.FS

const schemaMigrationsTable = `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY);`

type Runner struct {
	pool   *pgxpool.Pool
	log    zerolog.Logger
	sqlFS  fs.FS
	sqlDir string
}

func NewRunner(pool *pgxpool.Pool, log zerolog.Logger, sqlFS fs.FS, sqlDir string) *Runner {
	return &Runner{pool: pool, log: log, sqlFS: sqlFS, sqlDir: sqlDir}
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, log zerolog.Logger) error {
	r := NewRunner(pool, log, sqlFS, "sql")
	return r.Run(ctx)
}

func (r *Runner) Run(ctx context.Context) error {
	if _, err := r.pool.Exec(ctx, schemaMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(r.sqlFS, r.sqlDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		if version == "" {
			continue
		}
		applied, err := r.isApplied(ctx, version)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			r.log.Debug().Str("version", version).Msg("migration already applied")
			continue
		}

		path := r.sqlDir + "/" + name
		body, err := fs.ReadFile(r.sqlFS, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("run %s: %w", version, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record %s: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit %s: %w", version, err)
		}
		r.log.Info().Str("version", version).Msg("migration applied")
	}

	return nil
}

func (r *Runner) isApplied(ctx context.Context, version string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE version = $1`, version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}