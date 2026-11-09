package config

import (
	"bufio"
	"os"
	"strings"
)

func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, val)
	}
}

type Config struct {
	Addr          string
	SQLitePath    string
	JWTSecret     string
	MigrationsDir string
}

func Load() Config {
	return Config{
		Addr:          envOr("ADDR", ":8220"),
		SQLitePath:    envOr("SQLITE_PATH", "./data/agent_brain.db"),
		JWTSecret:     envOr("JWT_SECRET", "dev-agent-brain-jwt-change-me"),
		MigrationsDir: envOr("MIGRATIONS_DIR", "migrations"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
