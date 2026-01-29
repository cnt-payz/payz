package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

var (
	ErrAppInvalidName    = errors.New("invalid name of app")
	ErrAppInvalidVersion = errors.New("invalid version of app")

	ErrServerInvalidTimeout = errors.New("invalid server timeout")

	ErrPostgresInvalidUser         = errors.New("invalid user of pg")
	ErrPostgresInvalidPassword     = errors.New("invalid password of pg user")
	ErrPostgresInvalidDBName       = errors.New("invalid name of pg db")
	ErrPostgresConnMaxIdleTime     = errors.New("invalid time of idle conn")
	ErrPostgresConnMaxLifetime     = errors.New("invalid life time of conn")
	ErrPostgresInvalidMaxIdleConns = errors.New("invalid max idle conns")
	ErrPostgresInvalidMaxOpenConns = errors.New("invalid max open conns")

	ErrJWTInvalidPublicKeyPath  = errors.New("invalid path to public key")
	ErrJWTInvalidPrivateKeyPath = errors.New("invalid path to private key")
	ErrJWTInvalidAccessTTL      = errors.New("invalid access ttl")

	ErrRedisInvalidSessionTTL = errors.New("invalid session ttl")
	ErrRedisInvalidUserTTL    = errors.New("invalid user ttl")

	ErrTLSInvalidServerCert = errors.New("invalid path to server cert")
	ErrTLSInvalidServerKey  = errors.New("invalid path to server key")
	ErrTLSInvalidCaCert     = errors.New("invalid path to ca cert")
	ErrTLSInvalidCaKey      = errors.New("invalid path to ca key")

	ErrInvalidPath = errors.New("invalid path to config file")
)

type app struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (a *app) validate() error {
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return ErrAppInvalidName
	}

	a.Version = strings.TrimSpace(a.Version)
	if a.Version == "" {
		return ErrAppInvalidVersion
	}

	return nil
}

type server struct {
	GRPC struct {
		Protocol string `yaml:"protocol"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		TLS      struct {
			Enable     bool   `yaml:"enable"`
			ServerCert string `yaml:"server-cert"`
			ServerKey  string `yaml:"server-key"`
			CaCert     string `yaml:"ca-cert"`
			CaKey      string `yaml:"ca-key"`
		} `yaml:"tls"`
	} `yaml:"grpc"`
	Timeout time.Duration `yaml:"timeout"`
}

func (s *server) validate() error {
	if s.Timeout < 100*time.Millisecond {
		return ErrServerInvalidTimeout
	}

	s.GRPC.TLS.ServerCert = strings.TrimSpace(s.GRPC.TLS.ServerCert)
	if s.GRPC.TLS.ServerCert == "" {
		return ErrTLSInvalidServerCert
	}

	s.GRPC.TLS.ServerKey = strings.TrimSpace(s.GRPC.TLS.ServerKey)
	if s.GRPC.TLS.ServerKey == "" {
		return ErrTLSInvalidServerKey
	}

	s.GRPC.TLS.CaCert = strings.TrimSpace(s.GRPC.TLS.CaCert)
	if s.GRPC.TLS.CaCert == "" {
		return ErrTLSInvalidCaCert
	}

	s.GRPC.TLS.CaKey = strings.TrimSpace(s.GRPC.TLS.CaKey)
	if s.GRPC.TLS.CaKey == "" {
		return ErrTLSInvalidCaKey
	}

	return nil
}

func (s *server) applyDefaults() {
	s.GRPC.Host = strings.TrimSpace(s.GRPC.Host)
	if s.GRPC.Host == "" {
		s.GRPC.Host = "localhost"
	}

	if s.GRPC.Port == 0 {
		s.GRPC.Port = 50050
	}

	s.GRPC.Protocol = strings.TrimSpace(s.GRPC.Protocol)
	if s.GRPC.Protocol == "" {
		s.GRPC.Protocol = "tcp"
	}
}

type postgres struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         string        `yaml:"sslmode"`
	ConnMaxIdleTime time.Duration `yaml:"conn-max-idle-time"`
	ConnMaxLifetime time.Duration `yaml:"conn-max-lifetime"`
	MaxIdleConns    int           `yaml:"max-idle-conns"`
	MaxOpenConns    int           `yaml:"max-open-conns"`
}

func (pg *postgres) applyDefaults() {
	pg.Host = strings.TrimSpace(pg.Host)
	if pg.Host == "" {
		pg.Host = "localhost"
	}

	if pg.Port == 0 {
		pg.Port = 5432
	}

	pg.SSLMode = strings.TrimSpace(pg.SSLMode)
	if pg.SSLMode == "" {
		pg.SSLMode = "disable"
	}
}

func (pg *postgres) validate() error {
	pg.User = strings.TrimSpace(pg.User)
	if pg.User == "" {
		return ErrPostgresInvalidUser
	}

	pg.Password = strings.TrimSpace(pg.Password)
	if pg.Password == "" {
		return ErrPostgresInvalidPassword
	}

	pg.DBName = strings.TrimSpace(pg.DBName)
	if pg.DBName == "" {
		return ErrPostgresInvalidDBName
	}

	if pg.ConnMaxIdleTime < time.Second {
		return ErrPostgresConnMaxIdleTime
	}

	if pg.ConnMaxLifetime < time.Second {
		return ErrPostgresConnMaxLifetime
	}

	if pg.MaxIdleConns <= 0 {
		return ErrPostgresInvalidMaxIdleConns
	}

	if pg.MaxOpenConns <= 0 {
		return ErrPostgresInvalidMaxOpenConns
	}

	return nil
}

type jwt struct {
	PrivateKeyPath string        `yaml:"private-key-path"`
	PublicKeyPath  string        `yaml:"public-key-path"`
	AccessTTL      time.Duration `yaml:"access-ttl"`
}

func (j *jwt) validate() error {
	j.PrivateKeyPath = strings.TrimSpace(j.PrivateKeyPath)
	if j.PrivateKeyPath == "" {
		return ErrJWTInvalidPrivateKeyPath
	}

	j.PublicKeyPath = strings.TrimSpace(j.PublicKeyPath)
	if j.PublicKeyPath == "" {
		return ErrJWTInvalidPublicKeyPath
	}

	if j.AccessTTL < time.Minute {
		return ErrJWTInvalidAccessTTL
	}

	return nil
}

type redis struct {
	Host       string        `yaml:"host"`
	Port       int           `yaml:"port"`
	DB         int           `yaml:"db"`
	User       string        `yaml:"user"`
	Password   string        `yaml:"password"`
	SessionTTL time.Duration `yaml:"session-ttl"`
	UserTTL    time.Duration `yaml:"user-ttl"`
}

func (r *redis) applyDefaults() {
	r.Host = strings.TrimSpace(r.Host)
	if r.Host == "" {
		r.Host = "localhost"
	}

	if r.Port == 0 {
		r.Port = 6379
	}
}

func (r *redis) validate() error {
	if r.SessionTTL < 100*time.Millisecond {
		return ErrRedisInvalidSessionTTL
	}
	if r.UserTTL < 100*time.Millisecond {
		return ErrRedisInvalidUserTTL
	}

	return nil
}

type Config struct {
	Env     string `yaml:"env"`
	App     app    `yaml:"app"`
	Server  server `yaml:"server"`
	Secrets struct {
		Postgres postgres `yaml:"postgres"`
		JWT      jwt      `yaml:"jwt"`
		Redis    redis    `yaml:"redis"`
	} `yaml:"secrets"`
}

func New(path string) (*Config, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrInvalidPath
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	cfg.Env = strings.TrimSpace(cfg.Env)
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustNew(path string) *Config {
	cfg, err := New(path)
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return cfg
}

func (c *Config) validate() error {
	c.Server.applyDefaults()
	c.Secrets.Postgres.applyDefaults()
	c.Secrets.Redis.applyDefaults()

	if err := c.App.validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Secrets.Postgres.validate(); err != nil {
		return fmt.Errorf("invalid postgres: %w", err)
	}
	if err := c.Secrets.JWT.validate(); err != nil {
		return fmt.Errorf("invalid jwt: %w", err)
	}
	if err := c.Secrets.Redis.validate(); err != nil {
		return fmt.Errorf("invalid redis: %w", err)
	}

	return nil
}
