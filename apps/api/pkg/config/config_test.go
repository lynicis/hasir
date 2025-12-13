package config

import (
	"os"
	"path/filepath"
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

		envVars := map[string]string{
			"HASIR_SERVER_PUBLICURL":            "https://api.example.com",
			"HASIR_SERVER_IP":                   "0.0.0.0",
			"HASIR_SERVER_PORT":                 "3000",
			"HASIR_OTEL_ENABLED":                "true",
			"HASIR_OTEL_TRACEENDPOINT":          "localhost:4317",
			"HASIR_POSTGRESQL_CONNECTIONSTRING": "postgres://user:pass@localhost:5432/db",
			"HASIR_POSTGRESQL_HOST":             "localhost",
			"HASIR_POSTGRESQL_PORT":             "5432",
			"HASIR_POSTGRESQL_USERNAME":         "dbuser",
			"HASIR_POSTGRESQL_PASSWORD":         "dbpass",
			"HASIR_POSTGRESQL_DATABASE":         "testdb",
			"HASIR_SDKGENERATION_OUTPUTPATH":    "/var/sdk",
			"HASIR_SDKGENERATION_POLLINTERVAL":  "5s",
			"HASIR_SDKGENERATION_WORKERCOUNT":   "10",
			"HASIR_DASHBOARDURL":                "https://dashboard.example.com",
			"HASIR_SMTP_HOST":                   "smtp.example.com",
			"HASIR_SMTP_PORT":                   "587",
			"HASIR_SMTP_USERNAME":               "smtpuser",
			"HASIR_SMTP_PASSWORD":               "smtppass",
			"HASIR_SMTP_FROM":                   "no-reply@example.com",
			"HASIR_SMTP_USETLS":                 "true",
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
		assert.True(t, config.Otel.Enabled)
		assert.Equal(t, "localhost:4317", config.Otel.TraceEndpoint)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", config.PostgresConfig.ConnectionString)
		assert.Equal(t, "localhost", config.PostgresConfig.Host)
		assert.Equal(t, "5432", config.PostgresConfig.Port)
		assert.Equal(t, "dbuser", config.PostgresConfig.Username)
		assert.Equal(t, "dbpass", config.PostgresConfig.Password)
		assert.Equal(t, "testdb", config.PostgresConfig.Database)
		assert.Equal(t, "https://dashboard.example.com", config.DashboardUrl)
		assert.Equal(t, "/var/sdk", config.SdkGeneration.OutputPath)
		assert.Equal(t, "5s", config.SdkGeneration.PollInterval)
		assert.Equal(t, 10, config.SdkGeneration.WorkerCount)
		assert.Equal(t, "smtp.example.com", config.Smtp.Host)
		assert.Equal(t, 587, config.Smtp.Port)
		assert.Equal(t, "smtpuser", config.Smtp.Username)
		assert.Equal(t, "smtppass", config.Smtp.Password)
		assert.Equal(t, "no-reply@example.com", config.Smtp.From)
		assert.True(t, config.Smtp.UseTLS)
	})

	t.Run("returns empty config when no env vars set", func(t *testing.T) {

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
	t.Run("reads config from json file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `{
			"server": {
				"publicUrl": "https://api.example.com",
				"ip": "0.0.0.0",
				"port": "3000"
			},
			"otel": {
				"enabled": true,
				"traceEndpoint": "localhost:4317"
			},
			"postgresql": {
				"connectionString": "postgres://user:pass@localhost:5432/db",
				"host": "localhost",
				"port": "5432",
				"username": "dbuser",
				"password": "dbpass",
				"database": "testdb"
			},
			"sdkGeneration": {
				"workerCount": 10,
				"pollInterval": "5s",
				"outputPath": "/var/sdk"
			},
			"smtp": {
				"host": "smtp.example.com",
				"port": 587,
				"username": "smtpuser",
				"password": "smtppass",
				"from": "no-reply@example.com",
				"useTLS": true
			},
			"dashboardUrl": "https://dashboard.example.com"
		}`
		configPath := filepath.Join(tmpDir, "config.json")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		reader := &JsonConfig{ConfigPath: configPath}
		config := reader.Read()

		assert.Equal(t, "https://api.example.com", config.Server.PublicUrl)
		assert.Equal(t, "0.0.0.0", config.Server.Ip)
		assert.Equal(t, "3000", config.Server.Port)
		assert.True(t, config.Otel.Enabled)
		assert.Equal(t, "localhost:4317", config.Otel.TraceEndpoint)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db", config.PostgresConfig.ConnectionString)
		assert.Equal(t, "localhost", config.PostgresConfig.Host)
		assert.Equal(t, "5432", config.PostgresConfig.Port)
		assert.Equal(t, "dbuser", config.PostgresConfig.Username)
		assert.Equal(t, "dbpass", config.PostgresConfig.Password)
		assert.Equal(t, "testdb", config.PostgresConfig.Database)
		assert.Equal(t, "https://dashboard.example.com", config.DashboardUrl)
		assert.Equal(t, 10, config.SdkGeneration.WorkerCount)
		assert.Equal(t, "5s", config.SdkGeneration.PollInterval)
		assert.Equal(t, "/var/sdk", config.SdkGeneration.OutputPath)
		assert.Equal(t, "smtp.example.com", config.Smtp.Host)
		assert.Equal(t, 587, config.Smtp.Port)
		assert.Equal(t, "smtpuser", config.Smtp.Username)
		assert.Equal(t, "smtppass", config.Smtp.Password)
		assert.Equal(t, "no-reply@example.com", config.Smtp.From)
		assert.True(t, config.Smtp.UseTLS)
	})

	t.Run("panics when config file does not exist", func(t *testing.T) {
		reader := &JsonConfig{ConfigPath: "/nonexistent/path/config.json"}

		assert.Panics(t, func() {
			reader.Read()
		}, "expected panic when config file does not exist")
	})
}

func TestGetCwd(t *testing.T) {
	cwd := getCwd()
	require.NotEmpty(t, cwd)
}

func TestSdkGenerationConfig(t *testing.T) {
	t.Run("env config reads SDK generation settings", func(t *testing.T) {
		envVars := map[string]string{
			"HASIR_SDKGENERATION_WORKERCOUNT":  "5",
			"HASIR_SDKGENERATION_POLLINTERVAL": "10s",
			"HASIR_SDKGENERATION_OUTPUTPATH":   "/custom/sdk/path",
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

		assert.Equal(t, 5, config.SdkGeneration.WorkerCount)
		assert.Equal(t, "10s", config.SdkGeneration.PollInterval)
		assert.Equal(t, "/custom/sdk/path", config.SdkGeneration.OutputPath)
	})

	t.Run("json config reads SDK generation settings", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `{
			"sdkGeneration": {
				"workerCount": 10,
				"pollInterval": "5s",
				"outputPath": "/var/sdk"
			}
		}`
		configPath := filepath.Join(tmpDir, "config.json")
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		require.NoError(t, err)

		reader := &JsonConfig{ConfigPath: configPath}
		config := reader.Read()

		assert.Equal(t, 10, config.SdkGeneration.WorkerCount)
		assert.Equal(t, "5s", config.SdkGeneration.PollInterval)
		assert.Equal(t, "/var/sdk", config.SdkGeneration.OutputPath)
	})

	t.Run("defaults when SDK generation not configured", func(t *testing.T) {
		reader := &EnvConfig{}
		config := reader.Read()

		assert.Equal(t, 0, config.SdkGeneration.WorkerCount)
		assert.Empty(t, config.SdkGeneration.PollInterval)
		assert.Empty(t, config.SdkGeneration.OutputPath)
	})
}
