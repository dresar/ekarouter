package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Host                string
	Port                int
	DatabasePath        string
	SecretKey           string
	AdminUser           string
	AdminPassword       string
	AllowLocalProviders bool
	TokenSaverMode      string
	MaxRequestBodyBytes int64
	CORSOrigins         []string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	LogRetentionDays    int
}

func DefaultConfig() *Config {
	return &Config{
		Host:                "0.0.0.0",
		Port:                8080,
		DatabasePath:        "data/ekarouter.db",
		SecretKey:           generateDefaultSecretKey(),
		AdminUser:           "admin",
		AdminPassword:       "admin12345",
		AllowLocalProviders: false,
		TokenSaverMode:      "safe",
		MaxRequestBodyBytes: 10 * 1024 * 1024,
		CORSOrigins:         []string{"*"},
		ReadTimeout:         30 * time.Second,
		WriteTimeout:        120 * time.Second,
		IdleTimeout:         60 * time.Second,
		LogRetentionDays:    30,
	}
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	if v := os.Getenv("EKAROUTER_HOST"); v != "" {
		cfg.Host = v
	}
	if v := os.Getenv("EKAROUTER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 && p < 65536 {
			cfg.Port = p
		}
	}
	if v := os.Getenv("EKAROUTER_DB_PATH"); v != "" {
		cfg.DatabasePath = v
	}
	if v := os.Getenv("EKAROUTER_SECRET_KEY"); v != "" {
		cfg.SecretKey = v
	} else {
		cfg.SecretKey = getOrPersistSecretKey(cfg.DatabasePath)
	}
	if v := os.Getenv("EKAROUTER_ADMIN_USER"); v != "" {
		cfg.AdminUser = v
	}
	if v := os.Getenv("EKAROUTER_ADMIN_PASSWORD"); v != "" {
		cfg.AdminPassword = v
	}
	if v := os.Getenv("EKAROUTER_ALLOW_LOCAL_PROVIDERS"); v != "" {
		cfg.AllowLocalProviders = v == "true" || v == "1"
	}
	if v := os.Getenv("EKAROUTER_TOKEN_SAVER_MODE"); v != "" {
		cfg.TokenSaverMode = strings.ToLower(v)
	}
	if v := os.Getenv("EKAROUTER_CORS_ORIGINS"); v != "" {
		parts := strings.Split(v, ",")
		var origins []string
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
		if len(origins) > 0 {
			cfg.CORSOrigins = origins
		}
	}
	if v := os.Getenv("EKAROUTER_MAX_BODY_BYTES"); v != "" {
		if b, err := strconv.ParseInt(v, 10, 64); err == nil && b > 0 {
			cfg.MaxRequestBodyBytes = b
		}
	}

	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return errors.New("invalid port")
	}
	if strings.TrimSpace(c.DatabasePath) == "" {
		return errors.New("database path required")
	}
	if len(c.SecretKey) < 16 {
		return errors.New("secret key must be at least 16 characters")
	}
	if c.TokenSaverMode != "off" && c.TokenSaverMode != "safe" && c.TokenSaverMode != "balanced" {
		return errors.New("invalid token saver mode; must be off, safe, or balanced")
	}
	return nil
}

func generateDefaultSecretKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getOrPersistSecretKey(dbPath string) string {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}
	keyPath := filepath.Join(dir, ".secret_key")
	if data, err := os.ReadFile(keyPath); err == nil {
		k := strings.TrimSpace(string(data))
		if len(k) >= 16 {
			return k
		}
	}
	key := generateDefaultSecretKey()
	_ = os.WriteFile(keyPath, []byte(key), 0600)
	return key
}
