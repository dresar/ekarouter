package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid default config, got %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Port)
	}
	if cfg.TokenSaverMode != "safe" {
		t.Errorf("expected safe mode, got %s", cfg.TokenSaverMode)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("EKAROUTER_PORT", "9090")
	os.Setenv("EKAROUTER_TOKEN_SAVER_MODE", "balanced")
	defer func() {
		os.Unsetenv("EKAROUTER_PORT")
		os.Unsetenv("EKAROUTER_TOKEN_SAVER_MODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.TokenSaverMode != "balanced" {
		t.Errorf("expected balanced, got %s", cfg.TokenSaverMode)
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Port = 99999
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid port")
	}

	cfg = DefaultConfig()
	cfg.TokenSaverMode = "invalid_mode"
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid token saver mode")
	}
}
