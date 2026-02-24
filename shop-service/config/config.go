package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

var (
	ErrEmptyGRPCHost = errors.New("empty GRPC_HOST")
	ErrEmptyGRPCPort = errors.New("empty GRPC_PORT")

	ErrEmptyCACert     = errors.New("empty ca-cert")
	ErrEmptyServerCert = errors.New("empty server-cert")
	ErrEmptyServerKey  = errors.New("empty server-key")

	ErrEmptySSOHost = errors.New("empty SSO_HOST")
	ErrEmptySSOPort = errors.New("empty SSO_PORT")

	ErrEmptyDBName     = errors.New("empty DB_NAME")
	ErrEmptyDBUser     = errors.New("empty DB_USER")
	ErrEmptyDBPassword = errors.New("empty DB_PASSWORD")
	ErrEmptyDBHost     = errors.New("empty DB_HOST")
	ErrEmptyDBPort     = errors.New("empty DB_PORT")

	ErrEmptyRedisHost = errors.New("empty REDIS_HOST")
	ErrEmptyRedisPort = errors.New("empty REDIS_PORT")

	ErrEmptyLoggerOutputType = errors.New("empty logger output-type")
	ErrEmptyLoggerLevel      = errors.New("empty logger level")
)

type GRPCConfig struct {
	Host string    `yaml:"host"`
	Port string    `yaml:"port"`
	TLS  TLSConfig `yaml:"tls"`
	SSO  SSOConfig `yaml:"sso"`
}

type TLSConfig struct {
	Enable     bool   `yaml:"enable"`
	CACert     string `yaml:"ca-cert"`
	ServerCert string `yaml:"server-cert"`
	ServerKey  string `yaml:"server-key"`
}

type SSOConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type DBConfig struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	OutputType string `yaml:"output-type"`
	JSONFormat bool   `yaml:"json-format"`
}

type Config struct {
	GRPC   GRPCConfig   `yaml:"grpc"`
	DB     DBConfig     `yaml:"db"`
	Redis  RedisConfig  `yaml:"redis"`
	Logger LoggerConfig `yaml:"logger"`
}

func New(path string) (*Config, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *Config) validate() error {
	switch "" {
	case cfg.GRPC.Host:
		return ErrEmptyGRPCHost
	case cfg.GRPC.Port:
		return ErrEmptyGRPCPort
	case cfg.GRPC.TLS.CACert:
		return ErrEmptyCACert
	case cfg.GRPC.TLS.ServerCert:
		return ErrEmptyServerCert
	case cfg.GRPC.TLS.ServerKey:
		return ErrEmptyServerKey
	case cfg.GRPC.SSO.Host:
		return ErrEmptySSOHost
	case cfg.GRPC.SSO.Port:
		return ErrEmptySSOPort
	case cfg.DB.Name:
		return ErrEmptyDBName
	case cfg.DB.User:
		return ErrEmptyDBUser
	case cfg.DB.Password:
		return ErrEmptyDBPassword
	case cfg.DB.Host:
		return ErrEmptyDBHost
	case cfg.DB.Port:
		return ErrEmptyDBPort
	case cfg.Redis.Host:
		return ErrEmptyRedisHost
	case cfg.Redis.Port:
		return ErrEmptyRedisPort
	case cfg.Logger.Level:
		return ErrEmptyLoggerLevel
	case cfg.Logger.OutputType:
		return ErrEmptyLoggerOutputType
	default:
		return nil
	}
}
