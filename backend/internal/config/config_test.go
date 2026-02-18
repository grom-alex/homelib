package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidYAML(t *testing.T) {
	content := `
server:
  port: 9090
  host: "127.0.0.1"
database:
  host: "db.local"
  port: 5433
  user: "testuser"
  password: "testpass"
  dbname: "testdb"
  sslmode: "require"
auth:
  jwt_secret: "test-secret-key-must-be-at-least-32-chars-long"
  access_token_ttl: 10m
  refresh_token_ttl: 720h
  registration_enabled: false
library:
  inpx_path: "/data/lib.inpx"
  archives_path: "/data/archives"
import:
  batch_size: 5000
  log_every: 20000
`
	path := writeTemp(t, content)

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "127.0.0.1", cfg.Server.Host)
	assert.Equal(t, "db.local", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)
	assert.Equal(t, "test-secret-key-must-be-at-least-32-chars-long", cfg.Auth.JWTSecret)
	assert.Equal(t, 10*time.Minute, cfg.Auth.AccessTokenTTL)
	assert.Equal(t, 30*24*time.Hour, cfg.Auth.RefreshTokenTTL)
	assert.False(t, cfg.Auth.RegistrationEnabled)
	assert.Equal(t, "/data/lib.inpx", cfg.Library.INPXPath)
	assert.Equal(t, "/data/archives", cfg.Library.ArchivesPath)
	assert.Equal(t, 5000, cfg.Import.BatchSize)
	assert.Equal(t, 20000, cfg.Import.LogEvery)
}

func TestLoad_Defaults(t *testing.T) {
	content := `
database:
  host: "localhost"
  user: "app"
  password: "pw"
  dbname: "homelib"
auth:
  jwt_secret: "default-test-secret-must-be-at-least-32-chars"
`
	path := writeTemp(t, content)

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 15*time.Minute, cfg.Auth.AccessTokenTTL)
	assert.Equal(t, 30*24*time.Hour, cfg.Auth.RefreshTokenTTL)
	assert.True(t, cfg.Auth.RegistrationEnabled)
	assert.Equal(t, 3000, cfg.Import.BatchSize)
	assert.Equal(t, 10000, cfg.Import.LogEvery)
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	content := `
database:
  host: "localhost"
  user: "app"
  password: "pw"
  dbname: "homelib"
`
	path := writeTemp(t, content)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt_secret must be at least")
}

func TestLoad_ShortJWTSecret(t *testing.T) {
	content := `
database:
  host: "localhost"
  user: "app"
  password: "pw"
  dbname: "homelib"
auth:
  jwt_secret: "too-short"
`
	path := writeTemp(t, content)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "jwt_secret must be at least")
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read config file")
}

func TestLoad_InvalidYAML(t *testing.T) {
	path := writeTemp(t, "not: [valid: yaml: {{")

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse config file")
}

func TestLoad_EnvOverrides(t *testing.T) {
	content := `
database:
  host: "original-host"
  user: "original-user"
  password: "original-pass"
  dbname: "original-db"
auth:
  jwt_secret: "original-secret-key-must-be-long-enough-32"
library:
  inpx_path: "/original/inpx"
  archives_path: "/original/archives"
`
	path := writeTemp(t, content)

	t.Setenv("DB_HOST", "env-host")
	t.Setenv("DB_USER", "env-user")
	t.Setenv("DB_PASSWORD", "env-pass")
	t.Setenv("DB_NAME", "env-db")
	t.Setenv("JWT_SECRET", "env-secret-key-must-be-long-enough-too-32")
	t.Setenv("LIBRARY_PATH", "/env/archives")
	t.Setenv("INPX_PATH", "/env/inpx")

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "env-host", cfg.Database.Host)
	assert.Equal(t, "env-user", cfg.Database.User)
	assert.Equal(t, "env-pass", cfg.Database.Password)
	assert.Equal(t, "env-db", cfg.Database.DBName)
	assert.Equal(t, "env-secret-key-must-be-long-enough-too-32", cfg.Auth.JWTSecret)
	assert.Equal(t, "/env/archives", cfg.Library.ArchivesPath)
	assert.Equal(t, "/env/inpx", cfg.Library.INPXPath)
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "admin",
		Password: "secret",
		DBName:   "homelib",
		SSLMode:  "disable",
	}

	expected := "postgres://admin:secret@localhost:5432/homelib?sslmode=disable"
	assert.Equal(t, expected, cfg.DSN())
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}
