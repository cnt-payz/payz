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
	// App's
	ErrAppInvalidName    = errors.New("invalid name of service")
	ErrAppInvalidVersion = errors.New("invalid version of service")

	// Server's
	ErrServerInvalidReadTimeout  = errors.New("invalid read timeout")
	ErrServerInvalidWriteTimeout = errors.New("invalid write timeout")
	ErrServerInvalidIdleTimeout  = errors.New("invalid idle timeout")

	// Service's
	ErrServiceInvalidHost = errors.New("invalid host of service")
	ErrServiceInvalidPort = errors.New("invalid port of service")

	// Rate limiter's
	ErrRLInvalidWindow = errors.New("invalid window")
	ErrRLInvalidLimit  = errors.New("too little limit")

	// Domain's
	ErrInvalidPath = errors.New("invalid path to config file")
)

// Just a description of current service
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

// Settings of server
type server struct {
	HTTP struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"http"`
	ReadTimeout  time.Duration `yaml:"read-timeout"`
	WriteTimeout time.Duration `yaml:"write-timeout"`
	IdleTimeout  time.Duration `yaml:"idle-timeout"`
}

func (s *server) applyDefaults() {
	s.HTTP.Host = strings.TrimSpace(s.HTTP.Host)
	if s.HTTP.Host == "" {
		s.HTTP.Host = "localhost"
	}

	if s.HTTP.Port == 0 {
		s.HTTP.Port = 7080
	}
}

func (s *server) validate() error {
	if s.ReadTimeout < 100*time.Millisecond {
		return ErrServerInvalidReadTimeout
	}
	if s.WriteTimeout < 100*time.Millisecond {
		return ErrServerInvalidWriteTimeout
	}
	if s.IdleTimeout < 100*time.Millisecond {
		return ErrServerInvalidIdleTimeout
	}

	return nil
}

// Settings of sso service
type payzSSO struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (p *payzSSO) validate() error {
	p.Host = strings.TrimSpace(p.Host)
	if p.Host == "" {
		return ErrServiceInvalidHost
	}

	if p.Port == 0 {
		return ErrServiceInvalidPort
	}

	return nil
}

// Settings of rate-limiter
type rateLimit struct {
	Window time.Duration `yaml:"window"`
	Limit  int           `yaml:"limit"`
}

func (rl *rateLimit) validate() error {
	if rl.Limit <= 0 {
		return ErrRLInvalidLimit
	}
	if rl.Window < time.Second {
		return ErrRLInvalidWindow
	}

	return nil
}

type Config struct {
	Env       string    `yaml:"env"`
	App       app       `yaml:"app"`
	Server    server    `yaml:"server"`
	RateLimit rateLimit `yaml:"rate-limit"`
	Services  struct {
		PayzSSO payzSSO `yaml:"payz-sso"`
	} `yaml:"services"`
}

// Path to config file
// .env must be loaded
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

	cfg.Server.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if err := c.App.validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Services.PayzSSO.validate(); err != nil {
		return fmt.Errorf("invalid payz-sso: %w", err)
	}
	if err := c.RateLimit.validate(); err != nil {
		return fmt.Errorf("invalid rate-limit: %w", err)
	}

	return nil
}
