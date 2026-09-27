package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config собирается из переменных окружения.
type Config struct {
	Env       string
	LogLevel  slog.Level
	AdminHost string

	APIAddr     string
	DatabaseURL string
	Dialect     string // sqlite | postgres
	WebOrigin   string

	AdminEmail    string
	AdminPassword string

	TGBotToken   string
	TGNotifyChat string
}

func Load() (Config, error) {
	c := Config{
		Env:       env("APP_ENV", "dev"),
		LogLevel:  level(env("LOG_LEVEL", "info")),
		AdminHost: env("ADMIN_HOST", "panel.wolfandwings.ru"),

		APIAddr:     env("API_ADDR", "127.0.0.1:8080"),
		DatabaseURL: env("DATABASE_URL", "sqlite://./data/wolf.db"),
		WebOrigin:   env("WEB_ORIGIN", "http://localhost:3000"),

		AdminEmail:    env("ADMIN_EMAIL", "admin@wolfandwings.ru"),
		AdminPassword: env("ADMIN_PASSWORD", ""),

		TGBotToken:   env("TG_BOT_TOKEN", ""),
		TGNotifyChat: env("TG_NOTIFY_CHAT", ""),
	}

	dialect, err := parseDialect(c.DatabaseURL)
	if err != nil {
		return c, err
	}
	c.Dialect = dialect

	if c.Env == "prod" && c.AdminPassword == "" {
		return c, fmt.Errorf("ADMIN_PASSWORD required in prod")
	}
	return c, nil
}

func parseDialect(dsn string) (string, error) {
	switch {
	case strings.HasPrefix(dsn, "sqlite://"), strings.HasPrefix(dsn, "sqlite3://"):
		return "sqlite", nil
	case strings.HasPrefix(dsn, "postgres://"), strings.HasPrefix(dsn, "postgresql://"):
		return "postgres", nil
	}
	return "", fmt.Errorf("DATABASE_URL: unsupported scheme (want sqlite:// or postgres://): %s", dsn)
}

// SQLitePath достаёт путь к файлу из sqlite://./data/wolf.db
func SQLitePath(dsn string) string {
	for _, p := range []string{"sqlite3://", "sqlite://"} {
		if strings.HasPrefix(dsn, p) {
			return strings.TrimPrefix(dsn, p)
		}
	}
	return dsn
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func level(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
