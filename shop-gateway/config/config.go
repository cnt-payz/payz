package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

var (
	ErrEmptyHTTPHost = errors.New("empty htpp host")
	ErrEmptyHTTPPort = errors.New("empty htpp port")

	ErrEmptyTLSCACert     = errors.New("empty tls CA cert")
	ErrEmptyTLSClientCert = errors.New("empty tls client cert")
	ErrEmptyTLSClientKey  = errors.New("empty tls client key")

	ErrEmptyShopServiceHost = errors.New("empty shop-service host")
	ErrEmptyShopServicePort = errors.New("empty shop-service port")

	ErrEmptyLoggerLevel      = errors.New("empty logger level")
	ErrEmptyLoggerOutputType = errors.New("empty logger output type")
)

type HTTPConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type TLSConfig struct {
	Enable     bool   `yaml:"enable"`
	CACert     string `yaml:"ca-cert"`
	ClientCert string `yaml:"client-cert"`
	ClientKey  string `yaml:"client-key"`
}

type ShopServiceConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type GRPCConfig struct {
	TLS         TLSConfig         `yaml:"tls"`
	ShopService ShopServiceConfig `yaml:"shop-service"`
}

type LoggerConfig struct {
	Level      string `json:"level"`
	OutputType string `json:"output-type"`
	JSONFormat bool   `json:"json-format"`
}

type Config struct {
	HTTP   HTTPConfig   `yaml:"http"`
	GRPC   GRPCConfig   `yaml:"grpc"`
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
	case cfg.HTTP.Host:
		return ErrEmptyHTTPHost
	case cfg.HTTP.Port:
		return ErrEmptyHTTPPort
	case cfg.GRPC.TLS.CACert:
		return ErrEmptyTLSCACert
	case cfg.GRPC.TLS.ClientCert:
		return ErrEmptyTLSClientCert
	case cfg.GRPC.TLS.ClientKey:
		return ErrEmptyTLSClientKey
	case cfg.GRPC.ShopService.Host:
		return ErrEmptyShopServiceHost
	case cfg.GRPC.ShopService.Port:
		return ErrEmptyShopServicePort
	case cfg.Logger.Level:
		return ErrEmptyLoggerLevel
	case cfg.Logger.OutputType:
		return ErrEmptyLoggerOutputType
	default:
		return nil
	}
}
