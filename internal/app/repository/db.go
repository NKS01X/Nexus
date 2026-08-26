package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB wraps sql.DB with helper methods.
type DB struct {
	*sql.DB
}

// NewDB creates a new database connection pool.
func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}

	return &DB{db}, nil
}

// Tx executes fn within a transaction.
func (db *DB) Tx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("BeginTx: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Commit: %w", err)
	}
	return nil
}

// QueryRowCtx is a convenience wrapper for QueryRowContext.
func (db *DB) QueryRowCtx(ctx context.Context, query string, args ...any) *sql.Row {
	return db.QueryRowContext(ctx, query, args...)
}

// QueryCtx is a convenience wrapper for QueryContext.
func (db *DB) QueryCtx(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

// ExecCtx is a convenience wrapper for ExecContext.
func (db *DB) ExecCtx(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.ExecContext(ctx, query, args...)
}

// RunMigrations runs database migrations from the migrations directory.
func RunMigrations(db *DB) error {

	_, err := db.ExecCtx(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {

		migrationsDir = filepath.Join("..", "migrations")
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {

			migrationsDir = filepath.Join("..", "..", "..", "migrations")
			if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
				return fmt.Errorf("migrations directory not found")
			}
		}
	}

	files, err := fs.ReadDir(os.DirFS(migrationsDir), ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	applied := make(map[string]bool)
	rows, err := db.QueryCtx(context.Background(), `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}

	// Sort migration files
	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		version := strings.TrimSuffix(f, ".up.sql")
		if applied[version] {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		if _, err := db.ExecCtx(context.Background(), string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", f, err)
		}

		if _, err := db.ExecCtx(context.Background(), `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}
	}

	return nil
}
