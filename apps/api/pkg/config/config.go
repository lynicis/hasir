package config

import (
	stdjson "encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type PostgresConfig struct {
	ConnectionString string `koanf:"connectionString" sops:"secret"`
	Host             string `koanf:"host"`
	Port             string `koanf:"port"`
	Username         string `koanf:"username"`
	Password         string `koanf:"password" sops:"secret"`
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
	SshHost   string `koanf:"sshHost"`
	Ip        string `koanf:"ip"`
	Port      string `koanf:"port"`
}

func (srvc *ServerConfig) GetServerAddress() string {
	if srvc.Ip != "" {
		return fmt.Sprintf("%s:%s", srvc.Ip, srvc.Port)
	}

	return fmt.Sprintf(":%s", srvc.Port)
}

type OtelConfig struct {
	Enabled       bool   `koanf:"enabled"`
	TraceEndpoint string `koanf:"traceEndpoint"`
}

type SmtpConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password" sops:"secret"`
	From     string `koanf:"from"`
	UseTLS   bool   `koanf:"useTLS"`
}

type SshConfig struct {
	Enabled     bool   `koanf:"enabled"`
	Port        string `koanf:"port"`
	HostKeyPath string `koanf:"hostKeyPath"`
}

type SdkGenerationConfig struct {
	WorkerCount      int               `koanf:"workerCount"`
	PollInterval     string            `koanf:"pollInterval"`
	OutputPath       string            `koanf:"outputPath"`
	ModuleBasePath   string            `koanf:"moduleBasePath"`
	BufToken         string            `koanf:"bufToken" sops:"secret"`
	BufBinaryPath    string            `koanf:"bufBinaryPath"`
	CustomBsrModules map[string]string `koanf:"customBsrModules"`
}

func (sdk SdkGenerationConfig) GetModuleBasePath() string {
	if sdk.ModuleBasePath != "" {
		return sdk.ModuleBasePath
	}

	return "localhost"
}

func (sdk SdkGenerationConfig) GetOutputPath() string {
	if sdk.OutputPath != "" {
		return sdk.OutputPath
	}

	return "./sdk"
}

func (sdk SdkGenerationConfig) GetBufBinaryPath() string {
	if sdk.BufBinaryPath != "" {
		return sdk.BufBinaryPath
	}
	return "buf"
}

func (sdk SdkGenerationConfig) GetCustomBsrModules() map[string]string {
	if sdk.CustomBsrModules != nil {
		return sdk.CustomBsrModules
	}
	return nil
}

type Config struct {
	Server         ServerConfig        `koanf:"server"`
	Otel           OtelConfig          `koanf:"otel"`
	PostgresConfig PostgresConfig      `koanf:"postgresql"`
	Smtp           SmtpConfig          `koanf:"smtp"`
	Ssh            SshConfig           `koanf:"ssh"`
	SdkGeneration  SdkGenerationConfig `koanf:"sdkGeneration"`
	JwtSecret      []byte              `koanf:"jwtSecret" sops:"secret"`
	DashboardUrl   string              `koanf:"dashboardUrl"`
}

func (c *Config) String() string {
	m := redactStructToMap(reflect.ValueOf(*c))
	b, err := stdjson.Marshal(m)
	if err != nil {
		return fmt.Sprintf("error marshalling config: %v", err)
	}
	return string(b)
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(redactStructToAttrs(reflect.ValueOf(c))...)
}

func redactStructToAttrs(v reflect.Value) []slog.Attr {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	var attrs []slog.Attr

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		name := fieldType.Tag.Get("koanf")
		if name == "" {
			name = fieldType.Name
		}
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}

		if fieldType.Tag.Get("sops") == "secret" {
			attrs = append(attrs, slog.String(name, "[REDACTED]"))
			continue
		}

		switch fieldVal.Kind() {
		case reflect.Struct:
			nested := redactStructToAttrs(fieldVal)
			attrs = append(attrs, slog.Attr{
				Key:   name,
				Value: slog.GroupValue(nested...),
			})
		case reflect.Ptr:
			if fieldVal.IsNil() {
				attrs = append(attrs, slog.Any(name, nil))
			} else {
				nested := redactStructToAttrs(fieldVal.Elem())
				attrs = append(attrs, slog.Attr{
					Key:   name,
					Value: slog.GroupValue(nested...),
				})
			}
		default:
			attrs = append(attrs, slog.Any(name, fieldVal.Interface()))
		}
	}

	return attrs
}

func redactStructToMap(v reflect.Value) map[string]interface{} {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	res := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		name := fieldType.Tag.Get("koanf")
		if name == "" {
			name = fieldType.Name
		}
		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}

		if fieldType.Tag.Get("sops") == "secret" {
			res[name] = "[REDACTED]"
			continue
		}

		switch fieldVal.Kind() {
		case reflect.Struct:
			res[name] = redactStructToMap(fieldVal)
		case reflect.Ptr:
			if fieldVal.IsNil() {
				res[name] = nil
			} else {
				res[name] = redactStructToMap(fieldVal.Elem())
			}
		case reflect.Slice:
			if fieldVal.Type().Elem().Kind() == reflect.Uint8 {
				if fieldType.Tag.Get("sops") == "secret" {
					res[name] = "[REDACTED]"
				} else {
					res[name] = string(fieldVal.Bytes())
				}
			} else {
				res[name] = fieldVal.Interface()
			}
		default:
			res[name] = fieldVal.Interface()
		}
	}

	return res
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
