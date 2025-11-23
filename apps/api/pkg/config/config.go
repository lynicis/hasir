package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/knadh/koanf/parsers/json"
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

type Config struct {
	Server            ServerConfig   `koanf:"server"`
	OtelTraceEndpoint string         `koanf:"otelTraceEndpoint"`
	PostgresConfig    PostgresConfig `koanf:"postgresql"`
	JwtSecret         []byte         `koanf:"jwtSecret"`
	DashboardUrl      string         `koanf:"dashboardUrl"`
	RootUser          RootUserConfig `koanf:"rootUser"`
}

type ConfigReader interface {
	Read() *Config
}

var FileStrategies = map[string]ConfigReader{
	"json": &JsonConfig{},
}

// TODO: Only support json for now
func NewConfigReader() ConfigReader {
	return FileStrategies["json"]
}

func getCwd() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "../..")
}

type JsonConfig struct{}

func (c *JsonConfig) Read() *Config {
	koanfInstance := koanf.New(".")
	rootDir := getCwd()
	configPath := file.Provider(filepath.Join(rootDir, "config/config.json"))
	if err := koanfInstance.Load(configPath, json.Parser()); err != nil {
		panic(fmt.Sprintf("error occurred while reading config: %s", err))
	}

	var config Config
	if err := koanfInstance.Unmarshal("", &config); err != nil {
		panic(fmt.Sprintf("error occurred while unmarshalling config: %s", err))
	}

	return &config
}
