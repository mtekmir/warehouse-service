package config

import (
	"os"
	"time"
)

// Config stores congif values for the application.
type Config struct {
	Port             string
	DBURL            string
	DBMigrationsPath string
	WriteTimeout     time.Duration
	ReadTimeout      time.Duration
	IdleTimeout      time.Duration
	LogFile          *string
	Env              string
}

func getEnvOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

// Parse parses the env variables and returns a populated config struct.
func Parse() (*Config, error) {
	port := getEnvOrDefault("PORT", "8080")
	dbURL := getEnvOrDefault("DB_URL", "postgres://postgres:postgres@localhost:2345/postgres?sslmode=disable")
	DBMigrationsPath := getEnvOrDefault("DB_MIGRATIONS_PATH", "internal/postgres/migrations")
	env := getEnvOrDefault("ENV", "local")
	var logFile *string
	if lf := os.Getenv("LOG_FILE"); lf != "" {
		logFile = &lf
	}

	wTimeout, err := time.ParseDuration(getEnvOrDefault("WRITE_TIMEOUT", "15s"))
	if err != nil {
		return nil, err
	}
	rTimeout, err := time.ParseDuration(getEnvOrDefault("READ_TIMEOUT", "15s"))
	if err != nil {
		return nil, err
	}
	idleTimeout, err := time.ParseDuration(getEnvOrDefault("IDLE_TIMEOUT", "60s"))
	if err != nil {
		return nil, err
	}

	// TODO use flags if env vars are missing

	c := &Config{
		Port:             port,
		DBURL:            dbURL,
		DBMigrationsPath: DBMigrationsPath,
		WriteTimeout:     wTimeout,
		ReadTimeout:      rTimeout,
		IdleTimeout:      idleTimeout,
		LogFile:          logFile,
		Env:              env,
	}

	return c, nil
}
