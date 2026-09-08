package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultTLD string `yaml:"default_tld"`
	AutoHTTPS  bool   `yaml:"auto_https"`
	DaemonPort int    `yaml:"daemon_port"`
	HTTPSPort  int    `yaml:"https_port"`
	DNSPort    int    `yaml:"dns_port"`
	LogLevel   string `yaml:"log_level"`
	LogFile    string `yaml:"log_file"`
}

func Default() *Config {
	return &Config{
		DefaultTLD: "test",
		AutoHTTPS:  false,
		DaemonPort: 80,
		HTTPSPort:  443,
		DNSPort:    15353,
		LogLevel:   "info",
		LogFile:    filepath.Join(Dir(), "hatch.log"),
	}
}

func (c *Config) CertsDir() string {
	return filepath.Join(Dir(), "certs")
}

func Dir() string {
	if v := os.Getenv("HATCH_DIR"); v != "" {
		return v
	}
	switch runtime.GOOS {
	case "windows":
		if pd := os.Getenv("PROGRAMDATA"); pd != "" {
			return filepath.Join(pd, "hatch")
		}
		return `C:\ProgramData\hatch`
	default:
		return "/usr/local/var/hatch"
	}
}

func DefaultPath() string {
	return filepath.Join(Dir(), "config.yaml")
}

func (c *Config) SocketPath() string {
	return filepath.Join(Dir(), "hatch.sock")
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func Save(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
