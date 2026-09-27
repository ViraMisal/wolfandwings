package platform

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // postgres driver (database/sql)
	_ "modernc.org/sqlite"             // pure-go sqlite driver

	"github.com/wolfandwings/api/internal/config"
)

//go:embed migrations/sqlite/*.sql
var sqliteMigrations embed.FS

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

// Open открывает БД по DATABASE_URL и прогоняет миграции.
func Open(ctx context.Context, cfg config.Config, log *slog.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error
	switch cfg.Dialect {
	case "sqlite":
		path := config.SQLitePath(cfg.DatabaseURL)
		if dir := filepath.Dir(path); dir != "" {
			_ = os.MkdirAll(dir, 0o750)
		}
		// прагмы — синтаксис modernc.org/sqlite (не как у mattn/go-sqlite3)
		dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
		db, err = sql.Open("sqlite", dsn)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		db.SetMaxOpenConns(1) // sqlite: пишем последовательно
	case "postgres":
		db, err = sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		db.SetMaxOpenConns(20)
	default:
		return nil, fmt.Errorf("unknown dialect: %s", cfg.Dialect)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(ctx, db, cfg.Dialect, log); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func migrationFS(dialect string) (fs.FS, string) {
	switch dialect {
	case "sqlite":
		return sqliteMigrations, "migrations/sqlite"
	case "postgres":
		return postgresMigrations, "migrations/postgres"
	}
	return nil, ""
}

func migrate(ctx context.Context, db *sql.DB, dialect string, log *slog.Logger) error {
	fsys, dir := migrationFS(dialect)
	if fsys == nil {
		return fmt.Errorf("no migrations for dialect %s", dialect)
	}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	ph := func(n int) string {
		if dialect == "postgres" {
			return fmt.Sprintf("$%d", n)
		}
		return "?"
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL
	)`); err != nil {
		return err
	}
	applied := map[string]bool{}
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v string
		_ = rows.Scan(&v)
		applied[v] = true
	}
	rows.Close()

	for _, name := range files {
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			continue
		}
		body, err := fs.ReadFile(fsys, dir+"/"+name)
		if err != nil {
			return err
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, applied_at) VALUES (`+ph(1)+`, `+ph(2)+`)`,
			version, nowISO()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s (record): %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		log.Info("migration applied", "version", version, "dialect", dialect)
	}
	return nil
}

func nowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
