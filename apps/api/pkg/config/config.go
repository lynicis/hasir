package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type PostgresConfig struct {
	ConnectionString string `koanf:"connectionString"`
	Host             string `koanf:"host"`
	Port             string `koanf:"port"`
	Username         string `koanf:"username"`
	Password         string `koanf:"password"`
	Database         string `koanf:"database"`
}

func (pgc *PostgresConfig) GetPostgresDsn() string {
	if pgc.ConnectionString != "" {
		return pgc.ConnectionString
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgc.Host,
		pgc.Port,
		pgc.Username,
		pgc.Password,
		pgc.Database,
	)
}

type ServerConfig struct {
	PublicUrl string `koanf:"publicUrl"`
	Ip        string `koanf:"ip"`
	Port      string `koanf:"port"`
}

func (srvc *ServerConfig) GetServerAddress() string {
	if srvc.Ip != "" {
		return fmt.Sprintf("%s:%s", srvc.Ip, srvc.Port)
	}

	return fmt.Sprintf(":%s", srvc.Port)
}

type RootUserConfig struct {
	Username     string `koanf:"username"`
	TempPassword string `koanf:"tempPassword"`
}

type OtelConfig struct {
	Enabled       bool   `koanf:"enabled"`
	TraceEndpoint string `koanf:"traceEndpoint"`
}

type Config struct {
	Server         ServerConfig   `koanf:"server"`
	Otel           OtelConfig     `koanf:"otel"`
	PostgresConfig PostgresConfig `koanf:"postgresql"`
	JwtSecret      []byte         `koanf:"jwtSecret"`
	DashboardUrl   string         `koanf:"dashboardUrl"`
	RootUser       RootUserConfig `koanf:"rootUser"`
}

type ConfigReader interface {
	Read() *Config
}

func NewConfigReader() ConfigReader {
	mode := os.Getenv("MODE")
	if mode == "development" {
		return &JsonConfig{}
	}
	return &EnvConfig{}
}

func getCwd() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "../..")
}

type JsonConfig struct {
	ConfigPath string
}

func (c *JsonConfig) Read() *Config {
	koanfInstance := koanf.New(".")

	configFilePath := c.ConfigPath
	if configFilePath == "" {
		rootDir := getCwd()
		configFilePath = filepath.Join(rootDir, "config.json")
	}

	configPath := file.Provider(configFilePath)
	if err := koanfInstance.Load(configPath, json.Parser()); err != nil {
		panic(fmt.Sprintf("error occurred while reading config: %s", err))
	}

	var config Config
	if err := koanfInstance.Unmarshal("", &config); err != nil {
		panic(fmt.Sprintf("error occurred while unmarshalling config: %s", err))
	}

	return &config
}

type EnvConfig struct{}

func (c *EnvConfig) Read() *Config {
	koanfInstance := koanf.New(".")

	err := koanfInstance.Load(env.Provider("HASIR_", ".", func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, "HASIR_")),
			"_",
			".",
		)
	}), nil)
	if err != nil {
		panic(fmt.Sprintf("error occurred while reading env config: %s", err))
	}

	var config Config
	if err := koanfInstance.Unmarshal("", &config); err != nil {
		panic(fmt.Sprintf("error occurred while unmarshalling config: %s", err))
	}

	return &config
}
