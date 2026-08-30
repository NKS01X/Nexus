package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	command := "up"
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/aegis?sslmode=disable"
	}
	if len(os.Args) > 2 {
		dsn = os.Args[2]
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Migrations directory not found: %s\n", migrationsDir)
		os.Exit(1)
	}

	switch command {
	case "up":
		if err := migrateUp(db, migrationsDir); err != nil {
			fmt.Fprintf(os.Stderr, "Migration up failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		if err := migrateDown(db, migrationsDir); err != nil {
			fmt.Fprintf(os.Stderr, "Migration down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations rolled back successfully")
	case "status":
		if err := migrateStatus(db, migrationsDir); err != nil {
			fmt.Fprintf(os.Stderr, "Migration status failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func migrateUp(db *sql.DB, dir string) error {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	files, err := fs.ReadDir(os.DirFS(dir), ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	applied := make(map[string]bool)
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
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
			fmt.Printf("Skipping already applied: %s\n", version)
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		fmt.Printf("Applying: %s\n", version)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", f, err)
		}

		if _, err := db.Exec(`INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}
	}

	return nil
}

func migrateDown(db *sql.DB, dir string) error {
	files, err := fs.ReadDir(os.DirFS(dir), ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var downFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".down.sql") {
			downFiles = append(downFiles, f.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(downFiles)))

	for _, f := range downFiles {
		version := strings.TrimSuffix(f, ".down.sql")

		// Check if this version was applied
		var exists bool
		err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if !exists {
			fmt.Printf("Skipping not applied: %s\n", version)
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		fmt.Printf("Rolling back: %s\n", version)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("execute rollback %s: %w", f, err)
		}

		if _, err := db.Exec(`DELETE FROM schema_migrations WHERE version = $1`, version); err != nil {
			return fmt.Errorf("remove migration record %s: %w", version, err)
		}
	}

	return nil
}

func migrateStatus(db *sql.DB, dir string) error {
	files, err := fs.ReadDir(os.DirFS(dir), ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	applied := make(map[string]bool)
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
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

	var upFiles []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".up.sql") {
			upFiles = append(upFiles, strings.TrimSuffix(f.Name(), ".up.sql"))
		}
	}
	sort.Strings(upFiles)

	fmt.Println("Migration Status:")
	for _, v := range upFiles {
		status := "PENDING"
		if applied[v] {
			status = "APPLIED"
		}
		fmt.Printf("  %s: %s\n", v, status)
	}

	return nil
}
