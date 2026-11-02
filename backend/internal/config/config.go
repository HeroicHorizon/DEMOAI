package config

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
)

// Config holds all runtime configuration, read from environment variables.
type Config struct {
	Port        string
	MongoURI    string
	MongoDB     string
	CORSOrigin  string
	MaxBodySize int64
	LogLevel    slog.Level
}

// Load reads configuration using getenv. MONGODB_URI, MONGODB_DB and
// CORS_ALLOWED_ORIGIN are required and have no defaults.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:        orDefault(getenv("PORT"), "8080"),
		MongoURI:    getenv("MONGODB_URI"),
		MongoDB:     getenv("MONGODB_DB"),
		CORSOrigin:  getenv("CORS_ALLOWED_ORIGIN"),
		MaxBodySize: 1 << 20,
	}
	var missing []string
	for name, val := range map[string]string{
		"MONGODB_URI":         cfg.MongoURI,
		"MONGODB_DB":          cfg.MongoDB,
		"CORS_ALLOWED_ORIGIN": cfg.CORSOrigin,
	} {
		if strings.TrimSpace(val) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	if v := getenv("MAX_BODY_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("MAX_BODY_BYTES must be a positive integer")
		}
		cfg.MaxBodySize = n
	}
	if err := cfg.LogLevel.UnmarshalText([]byte(orDefault(getenv("LOG_LEVEL"), "info"))); err != nil {
		return Config{}, fmt.Errorf("invalid LOG_LEVEL")
	}
	return cfg, nil
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
