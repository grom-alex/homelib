package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Library  LibraryConfig  `yaml:"library"`
	Import   ImportConfig   `yaml:"import"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

type AuthConfig struct {
	JWTSecret           string        `yaml:"jwt_secret"`
	AccessTokenTTL      time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL     time.Duration `yaml:"refresh_token_ttl"`
	RegistrationEnabled bool          `yaml:"registration_enabled"`
	CookieSecure        bool          `yaml:"cookie_secure"`
}

type LibraryConfig struct {
	INPXPath     string `yaml:"inpx_path"`
	ArchivesPath string `yaml:"archives_path"`
}

type ImportConfig struct {
	BatchSize int `yaml:"batch_size"`
	LogEvery  int `yaml:"log_every"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Port:    5432,
			SSLMode: "disable",
		},
		Auth: AuthConfig{
			AccessTokenTTL:      15 * time.Minute,
			RefreshTokenTTL:     30 * 24 * time.Hour,
			RegistrationEnabled: true,
		},
		Import: ImportConfig{
			BatchSize: 3000,
			LogEvery:  10000,
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	applyEnvOverrides(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

const minJWTSecretLength = 32

func (c *Config) Validate() error {
	if len(c.Auth.JWTSecret) < minJWTSecretLength {
		return fmt.Errorf("jwt_secret must be at least %d characters long", minJWTSecretLength)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.DBName = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("LIBRARY_PATH"); v != "" {
		cfg.Library.ArchivesPath = v
	}
	if v := os.Getenv("INPX_PATH"); v != "" {
		cfg.Library.INPXPath = v
	}
}
