package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresConfig_GetPostgresDsn(t *testing.T) {
	t.Run("returns connection string when provided", func(t *testing.T) {
		pgc := &PostgresConfig{
			ConnectionString: "postgres://user:pass@localhost:5432/db",
			Host:             "ignored",
			Port:             "ignored",
			Username:         "ignored",
			Password:         "ignored",
			Database:         "ignored",
		}

		dsn := pgc.GetPostgresDsn()
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", dsn)
	})

	t.Run("builds DSN from individual fields when connection string is empty", func(t *testing.T) {
		pgc := &PostgresConfig{
			Host:     "localhost",
			Port:     "5432",
			Username: "testuser",
			Password: "testpass",
			Database: "testdb",
		}

		dsn := pgc.GetPostgresDsn()
		expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
		assert.Equal(t, expected, dsn)
	})
}

func TestServerConfig_GetServerAddress(t *testing.T) {
	t.Run("returns IP:Port when IP is provided", func(t *testing.T) {
		srvc := &ServerConfig{
			Ip:   "192.168.1.1",
			Port: "8080",
		}

		addr := srvc.GetServerAddress()
		assert.Equal(t, "192.168.1.1:8080", addr)
	})

	t.Run("returns :Port when IP is empty", func(t *testing.T) {
		srvc := &ServerConfig{
			Port: "8080",
		}

		addr := srvc.GetServerAddress()
		assert.Equal(t, ":8080", addr)
	})
}

func TestNewConfigReader(t *testing.T) {
	t.Run("returns EnvConfig when MODE is not set", func(t *testing.T) {
		_ = os.Unsetenv("MODE")

		reader := NewConfigReader()
		_, ok := reader.(*EnvConfig)
		assert.True(t, ok, "expected EnvConfig when MODE is not set")
	})

	t.Run("returns EnvConfig when MODE is production", func(t *testing.T) {
		_ = os.Setenv("MODE", "production")
		defer func() {
			_ = os.Unsetenv("MODE")
		}()

		reader := NewConfigReader()
		_, ok := reader.(*EnvConfig)
		assert.True(t, ok, "expected EnvConfig when MODE is production")
	})

	t.Run("returns JsonConfig when MODE is development", func(t *testing.T) {
		_ = os.Setenv("MODE", "development")
		defer func() {
			_ = os.Unsetenv("MODE")
		}()

		reader := NewConfigReader()
		_, ok := reader.(*JsonConfig)
		assert.True(t, ok, "expected JsonConfig when MODE is development")
	})
}

func TestEnvConfig_Read(t *testing.T) {
	t.Run("reads config from environment variables", func(t *testing.T) {
		// Set environment variables
		envVars := map[string]string{
			"HASIR_SERVER_PUBLICURL":            "https://api.example.com",
			"HASIR_SERVER_IP":                   "0.0.0.0",
			"HASIR_SERVER_PORT":                 "3000",
			"HASIR_OTELTRACEENDPOINT":           "http://otel:4317",
			"HASIR_POSTGRESQL_CONNECTIONSTRING": "postgres://user:pass@localhost:5432/db",
			"HASIR_POSTGRESQL_HOST":             "localhost",
			"HASIR_POSTGRESQL_PORT":             "5432",
			"HASIR_POSTGRESQL_USERNAME":         "dbuser",
			"HASIR_POSTGRESQL_PASSWORD":         "dbpass",
			"HASIR_POSTGRESQL_DATABASE":         "testdb",
			"HASIR_DASHBOARDURL":                "https://dashboard.example.com",
			"HASIR_ROOTUSER_USERNAME":           "admin",
			"HASIR_ROOTUSER_TEMPPASSWORD":       "temppass123",
		}

		for k, v := range envVars {
			_ = os.Setenv(k, v)
		}
		defer func() {
			for k := range envVars {
				_ = os.Unsetenv(k)
			}
		}()

		reader := &EnvConfig{}
		config := reader.Read()

		assert.Equal(t, "https://api.example.com", config.Server.PublicUrl)
		assert.Equal(t, "0.0.0.0", config.Server.Ip)
		assert.Equal(t, "3000", config.Server.Port)
		assert.Equal(t, "http://otel:4317", config.OtelTraceEndpoint)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", config.PostgresConfig.ConnectionString)
		assert.Equal(t, "localhost", config.PostgresConfig.Host)
		assert.Equal(t, "5432", config.PostgresConfig.Port)
		assert.Equal(t, "dbuser", config.PostgresConfig.Username)
		assert.Equal(t, "dbpass", config.PostgresConfig.Password)
		assert.Equal(t, "testdb", config.PostgresConfig.Database)
		assert.Equal(t, "https://dashboard.example.com", config.DashboardUrl)
		assert.Equal(t, "admin", config.RootUser.Username)
		assert.Equal(t, "temppass123", config.RootUser.TempPassword)
	})

	t.Run("returns empty config when no env vars set", func(t *testing.T) {
		// Clear any existing HASIR_ env vars
		for _, env := range os.Environ() {
			if len(env) > 6 && env[:6] == "HASIR_" {
				key := env[:len(env)-len(env[len(env):])]
				_ = os.Unsetenv(key)
			}
		}

		reader := &EnvConfig{}
		config := reader.Read()

		assert.Empty(t, config.Server.PublicUrl)
		assert.Empty(t, config.Server.Port)
	})
}

func TestJsonConfig_Read(t *testing.T) {
	t.Run("panics when config file does not exist", func(t *testing.T) {
		reader := &JsonConfig{}

		assert.Panics(t, func() {
			reader.Read()
		}, "expected panic when config file does not exist")
	})
}

func TestConfigReader_Interface(t *testing.T) {
	t.Run("EnvConfig implements ConfigReader", func(t *testing.T) {
		var _ ConfigReader = &EnvConfig{}
	})

	t.Run("JsonConfig implements ConfigReader", func(t *testing.T) {
		var _ ConfigReader = &JsonConfig{}
	})
}

func TestGetCwd(t *testing.T) {
	cwd := getCwd()
	require.NotEmpty(t, cwd)
}
