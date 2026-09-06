package config

import (
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.DefaultTLD != "test" {
		t.Errorf("DefaultTLD = %q, want %q", cfg.DefaultTLD, "test")
	}
	if cfg.DaemonPort != 8443 {
		t.Errorf("DaemonPort = %d, want %d", cfg.DaemonPort, 8443)
	}
	if cfg.DNSPort != 15353 {
		t.Errorf("DNSPort = %d, want %d", cfg.DNSPort, 15353)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := Default()
	original.DefaultTLD = "local"
	original.DaemonPort = 9999

	if err := Save(original, path); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.DefaultTLD != "local" {
		t.Errorf("loaded DefaultTLD = %q, want %q", loaded.DefaultTLD, "local")
	}
	if loaded.DaemonPort != 9999 {
		t.Errorf("loaded DaemonPort = %d, want %d", loaded.DaemonPort, 9999)
	}
}

func TestLoadNonExistentReturnsDefault(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.DefaultTLD != "test" {
		t.Errorf("DefaultTLD = %q, want %q", cfg.DefaultTLD, "test")
	}
}

func TestSocketPath(t *testing.T) {
	cfg := Default()
	sock := cfg.SocketPath()
	if sock == "" {
		t.Error("SocketPath() returned empty string")
	}
}
