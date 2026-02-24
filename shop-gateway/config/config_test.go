package config

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	fullHTTP := HTTPConfig{
		Host: "localhost",
		Port: "7081",
	}

	fullTLS := TLSConfig{
		CACert:     "./certs/ca.crt",
		ClientCert: "./certs/client.crt",
		ClientKey:  "./certs/client.key",
	}

	fullShopService := ShopServiceConfig{
		Host: "localhost",
		Port: "50051",
	}

	fullGRPC := GRPCConfig{
		TLS:         fullTLS,
		ShopService: fullShopService,
	}

	fullLogger := LoggerConfig{
		Level:      "info",
		OutputType: "console",
	}

	tests := []struct {
		name string
		cfg  *Config
		err  error
	}{
		{
			name: "empty http host",
			cfg: &Config{
				HTTP: HTTPConfig{
					Host: "",
					Port: fullHTTP.Port,
				},
				GRPC:   fullGRPC,
				Logger: fullLogger,
			},
			err: ErrEmptyHTTPHost,
		},
		{
			name: "empty http port",
			cfg: &Config{
				HTTP: HTTPConfig{
					Host: fullHTTP.Host,
					Port: "",
				},
				GRPC:   fullGRPC,
				Logger: fullLogger,
			},
			err: ErrEmptyHTTPPort,
		},
		{
			name: "empty tls ca cert",
			cfg: &Config{
				HTTP: fullHTTP,
				GRPC: GRPCConfig{
					TLS: TLSConfig{
						CACert:     "",
						ClientCert: fullTLS.ClientCert,
						ClientKey:  fullTLS.ClientKey,
					},
					ShopService: fullShopService,
				},
				Logger: fullLogger,
			},
			err: ErrEmptyTLSCACert,
		},
		{
			name: "empty tls client cert",
			cfg: &Config{
				HTTP: fullHTTP,
				GRPC: GRPCConfig{
					TLS: TLSConfig{
						CACert:     fullTLS.CACert,
						ClientCert: "",
						ClientKey:  fullTLS.ClientKey,
					},
					ShopService: fullShopService,
				},
				Logger: fullLogger,
			},
			err: ErrEmptyTLSClientCert,
		},
		{
			name: "empty tls client key",
			cfg: &Config{
				HTTP: fullHTTP,
				GRPC: GRPCConfig{
					TLS: TLSConfig{
						CACert:     fullTLS.CACert,
						ClientCert: fullTLS.ClientCert,
						ClientKey:  "",
					},
					ShopService: fullShopService,
				},
				Logger: fullLogger,
			},
			err: ErrEmptyTLSClientKey,
		},
		{
			name: "empty shop service host",
			cfg: &Config{
				HTTP: fullHTTP,
				GRPC: GRPCConfig{
					TLS: fullTLS,
					ShopService: ShopServiceConfig{
						Host: "",
						Port: fullShopService.Port,
					},
				},
				Logger: fullLogger,
			},
			err: ErrEmptyShopServiceHost,
		},
		{
			name: "empty shop service port",
			cfg: &Config{
				HTTP: fullHTTP,
				GRPC: GRPCConfig{
					TLS: fullTLS,
					ShopService: ShopServiceConfig{
						Host: fullShopService.Host,
						Port: "",
					},
				},
				Logger: fullLogger,
			},
			err: ErrEmptyShopServicePort,
		},
		{
			name: "empty logger lever",
			cfg: &Config{
				HTTP: fullHTTP,
				GRPC: fullGRPC,
				Logger: LoggerConfig{
					Level:      "",
					OutputType: fullLogger.OutputType,
				},
			},
			err: ErrEmptyLoggerLevel,
		},
		{
			name: "empty logger output type",
			cfg: &Config{
				HTTP: fullHTTP,
				GRPC: fullGRPC,
				Logger: LoggerConfig{
					Level:      fullLogger.Level,
					OutputType: "",
				},
			},
			err: ErrEmptyLoggerOutputType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if !errors.Is(err, tt.err) {
				t.Errorf("expected: %v, have: %v", tt.err, err)
			}
		})
	}
}
