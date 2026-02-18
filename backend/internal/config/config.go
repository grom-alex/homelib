package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Library  LibraryConfig  `yaml:"library"`
	Import   ImportConfig   `yaml:"import"`
	Reader   ReaderConfig   `yaml:"reader"`
}

type ReaderConfig struct {
	CachePath   string        `yaml:"cache_path"`
	CacheTTLRaw string        `yaml:"cache_ttl"`
	CacheTTL    time.Duration `yaml:"-"`
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
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:     d.DBName,
		RawQuery: fmt.Sprintf("sslmode=%s", url.QueryEscape(d.SSLMode)),
	}
	return u.String()
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
		Reader: ReaderConfig{
			CachePath: "./cache/books",
			CacheTTL:  30 * 24 * time.Hour,
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// Parse cache_ttl (supports "d" suffix for days, e.g. "30d")
	if cfg.Reader.CacheTTLRaw != "" {
		d, err := parseDuration(cfg.Reader.CacheTTLRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid reader.cache_ttl %q: %w", cfg.Reader.CacheTTLRaw, err)
		}
		cfg.Reader.CacheTTL = d
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
	if v := os.Getenv("READER_CACHE_PATH"); v != "" {
		cfg.Reader.CachePath = v
	}
	if v := os.Getenv("READER_CACHE_TTL"); v != "" {
		if d, err := parseDuration(v); err == nil {
			cfg.Reader.CacheTTL = d
		}
	}
}

// parseDuration extends time.ParseDuration with support for "d" (days) suffix.
// Examples: "30d" → 30*24h, "720h", "0".
func parseDuration(s string) (time.Duration, error) {
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	if strings.HasSuffix(s, "d") {
		days, err := strconv.ParseFloat(strings.TrimSuffix(s, "d"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(days * float64(24*time.Hour)), nil
	}
	return 0, fmt.Errorf("invalid duration: %s", s)
}
