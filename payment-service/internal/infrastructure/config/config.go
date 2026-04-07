package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

var (
	ErrAppInvalidName    = errors.New("invalid name of service")
	ErrAppInvalidVersion = errors.New("invalid version of service")

	ErrServerInvalidTimeout    = errors.New("invalid too little timeout")
	ErrServerInvalidServerCert = errors.New("invalid path to server cert")
	ErrServerInvalidServerKey  = errors.New("invalid path to server key")
	ErrServerInvalidCaCert     = errors.New("invalid path to ca cert")

	ErrPostgresInvalidUser     = errors.New("invalid postgres user")
	ErrPostgresInvalidPassword = errors.New("invalid postgres password")
	ErrPostgresInvalidDBName   = errors.New("invalid postgres db name")
	ErrPostgresInvalidLifetime = errors.New("invalid max lifetime")
	ErrPostgresInvalidIdletime = errors.New("invalid max idletime")
	ErrPostgresInvalidMaxIdle  = errors.New("invalid max idle conns")
	ErrPostgresInvalidMaxOpen  = errors.New("invalid max open conns")
	ErrPostgresInvalidTime     = errors.New("too little time ticker")

	ErrIdempotencyTooLittleTTL = errors.New("ttl is too little")

	ErrTimecfgTooLittleWindow = errors.New("too little window")

	ErrJWTPublicKeyPath = errors.New("invalid path to public key")

	ErrKafkaInvalidHost  = errors.New("invalid host")
	ErrKafkaInvalidPort  = errors.New("invalid port")
	ErrKafkaInvalidTopic = errors.New("invalid topic")

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
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Protocol string `yaml:"protocol"`
		TLS      struct {
			Enable     bool   `yaml:"enable"`
			ServerCert string `yaml:"server-cert"`
			ServerKey  string `yaml:"server-key"`
			CaCert     string `yaml:"ca-cert"`
		} `yaml:"tls"`
	} `yaml:"grpc"`
	Timeout time.Duration `yaml:"timeout"`
}

func (s *server) applyDefaults() {
	s.GRPC.Host = strings.TrimSpace(s.GRPC.Host)
	if s.GRPC.Host == "" {
		s.GRPC.Host = "localhost"
	}

	if s.GRPC.Port == 0 {
		s.GRPC.Port = 50052
	}

	s.GRPC.Protocol = strings.TrimSpace(s.GRPC.Protocol)
	if s.GRPC.Protocol == "" {
		s.GRPC.Protocol = "tcp"
	}
}

func (s *server) validate() error {
	if s.Timeout < 100*time.Millisecond {
		return ErrServerInvalidTimeout
	}

	s.GRPC.TLS.ServerCert = strings.TrimSpace(s.GRPC.TLS.ServerCert)
	if s.GRPC.TLS.ServerCert == "" {
		return ErrServerInvalidServerCert
	}

	s.GRPC.TLS.ServerKey = strings.TrimSpace(s.GRPC.TLS.ServerKey)
	if s.GRPC.TLS.ServerKey == "" {
		return ErrServerInvalidServerKey
	}

	s.GRPC.TLS.CaCert = strings.TrimSpace(s.GRPC.TLS.CaCert)
	if s.GRPC.TLS.CaCert == "" {
		return ErrServerInvalidCaCert
	}

	return nil
}

type postgres struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	SSLMode      string        `yaml:"sslmode"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	DBName       string        `yaml:"dbname"`
	WorkerTicker time.Duration `yaml:"worker-ticker"`
	Conn         struct {
		MaxLifetime time.Duration `yaml:"max-lifetime"`
		MaxIdletime time.Duration `yaml:"max-idletime"`
		MaxIdle     int           `yaml:"max-idle"`
		MaxOpen     int           `yaml:"max-open"`
	} `yaml:"conn"`
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

	if pg.Conn.MaxLifetime < time.Minute {
		return ErrPostgresInvalidLifetime
	}

	if pg.Conn.MaxIdletime < time.Minute {
		return ErrPostgresInvalidMaxIdle
	}

	if pg.Conn.MaxIdle <= 0 {
		return ErrPostgresInvalidMaxIdle
	}

	if pg.Conn.MaxOpen <= 0 {
		return ErrPostgresInvalidMaxOpen
	}

	if pg.WorkerTicker < 100*time.Millisecond {
		return ErrPostgresInvalidTime
	}

	return nil
}

type redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DB       int    `yaml:"db"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
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

type idempotency struct {
	TTL time.Duration `yaml:"ttl"`
}

func (i *idempotency) validate() error {
	if i.TTL < 100*time.Millisecond {
		return ErrIdempotencyTooLittleTTL
	}

	return nil
}

type jwt struct {
	PublicKeyPath string `yaml:"public-key-path"`
}

func (j *jwt) validate() error {
	j.PublicKeyPath = strings.TrimSpace(j.PublicKeyPath)
	if j.PublicKeyPath == "" {
		return ErrJWTPublicKeyPath
	}

	return nil
}

type timecfg struct {
	Window time.Duration `yaml:"window"`
}

func (t *timecfg) validate() error {
	if t.Window < time.Second {
		return ErrTimecfgTooLittleWindow
	}

	return nil
}

type kafka struct {
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Topic string `yaml:"topic"`
}

func (k *kafka) validate() error {
	k.Host = strings.TrimSpace(k.Host)
	if k.Host == "" {
		return ErrKafkaInvalidHost
	}

	if k.Port == 0 {
		return ErrKafkaInvalidPort
	}

	k.Topic = strings.TrimSpace(k.Topic)
	if k.Topic == "" {
		return ErrKafkaInvalidTopic
	}

	return nil
}

type Config struct {
	Env     string `yaml:"env"`
	App     app    `yaml:"app"`
	Server  server `yaml:"server"`
	Service struct {
		Private            string        `yaml:"private"`
		TransactionTimeout time.Duration `yaml:"transaction-timeout"`
		Timecfg            timecfg       `yaml:"time"`
		Idempotency        idempotency   `yaml:"idempotency"`
	} `yaml:"service"`
	Secrets struct {
		Postgres postgres `yaml:"postgres"`
		Redis    redis    `yaml:"redis"`
		JWT      jwt      `yaml:"jwt"`
		Kafka    kafka    `yaml:"kafka"`
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
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.Env = strings.TrimSpace(cfg.Env)
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cfg.Server.applyDefaults()
	cfg.Secrets.Postgres.applyDefaults()
	cfg.Secrets.Redis.applyDefaults()

	return &cfg, nil
}

func (c *Config) validate() error {
	if err := c.App.validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Secrets.Postgres.validate(); err != nil {
		return fmt.Errorf("invalid postgres: %w", err)
	}
	if err := c.Service.Idempotency.validate(); err != nil {
		return fmt.Errorf("invalid idempotency: %w", err)
	}
	if err := c.Secrets.JWT.validate(); err != nil {
		return fmt.Errorf("invalid jwt: %w", err)
	}
	if err := c.Service.Timecfg.validate(); err != nil {
		return fmt.Errorf("invalid time: %w", err)
	}
	if err := c.Secrets.Kafka.validate(); err != nil {
		return fmt.Errorf("invalid kafka: %w", err)
	}

	return nil
}
