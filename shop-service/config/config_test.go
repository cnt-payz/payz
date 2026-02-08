package config

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	fullGRPC := GRPCConfig{
		Host: "localhost",
		Port: "50051",
		TLS:  TLSConfig{Enable: false, CACert: "./certs/ca.crt", ServerCert: "./certs/shop-service.crt", ServerKey: "./certs/shop-service.key"},
		SSO:  SSOConfig{Host: "localhost", Port: "8080"},
	}

	fullDB := DBConfig{
		Name:     "shopdb",
		User:     "user",
		Password: "pass",
		Host:     "localhost",
		Port:     "5432",
	}

	fullRedis := RedisConfig{
		Host: "localhost",
		Port: "6379",
	}

	fullLogger := LoggerConfig{
		Level:      "info",
		OutputType: "console",
		JSONFormat: false,
	}

	tests := []struct {
		name string
		cfg  *Config
		err  error
	}{
		{
			name: "ok",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: nil,
		},
		{
			name: "empty grpc host",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: "",
					Port: fullGRPC.Port,
					SSO:  fullGRPC.SSO,
					TLS:  fullGRPC.TLS,
				},
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyGRPCHost,
		},
		{
			name: "empty grpc port",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: "",
					SSO:  fullGRPC.SSO,
					TLS:  fullGRPC.TLS,
				},
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyGRPCPort,
		},
		{
			name: "empty tls ca",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: fullGRPC.Port,
					TLS: TLSConfig{
						CACert:     "",
						ServerCert: fullGRPC.TLS.ServerCert,
						ServerKey:  fullGRPC.TLS.ServerKey,
					},
				},
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyCACert,
		},
		{
			name: "empty tls server cert",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: fullGRPC.Port,
					TLS: TLSConfig{
						CACert:     fullGRPC.TLS.CACert,
						ServerCert: "",
						ServerKey:  fullGRPC.TLS.ServerKey,
					},
				},
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyServerCert,
		},
		{
			name: "empty tls server key",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: fullGRPC.Port,
					TLS: TLSConfig{
						CACert:     fullGRPC.TLS.CACert,
						ServerCert: fullGRPC.TLS.ServerCert,
						ServerKey:  "",
					},
				},
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyServerKey,
		},
		{
			name: "empty sso host",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: fullGRPC.Port,
					TLS:  fullGRPC.TLS,
					SSO: SSOConfig{
						Host: "",
						Port: fullGRPC.SSO.Port,
					},
				},
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptySSOHost,
		},
		{
			name: "empty sso port",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: fullGRPC.Port,
					TLS:  fullGRPC.TLS,
					SSO: SSOConfig{
						Host: fullGRPC.SSO.Host,
						Port: "",
					},
				},
				DB:     fullDB,
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptySSOPort,
		},
		{
			name: "empty db name",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     DBConfig{Name: "", User: fullDB.User, Password: fullDB.Password, Host: fullDB.Host, Port: fullDB.Port},
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyDBName,
		},
		{
			name: "empty db user",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     DBConfig{Name: fullDB.Name, User: "", Password: fullDB.Password, Host: fullDB.Host, Port: fullDB.Port},
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyDBUser,
		},
		{
			name: "empty db password",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     DBConfig{Name: fullDB.Name, User: fullDB.User, Password: "", Host: fullDB.Host, Port: fullDB.Port},
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyDBPassword,
		},
		{
			name: "empty db host",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     DBConfig{Name: fullDB.Name, User: fullDB.User, Password: fullDB.Password, Host: "", Port: fullDB.Port},
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyDBHost,
		},
		{
			name: "empty db port",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     DBConfig{Name: fullDB.Name, User: fullDB.User, Password: fullDB.Password, Host: fullDB.Host, Port: ""},
				Redis:  fullRedis,
				Logger: fullLogger,
			},
			err: ErrEmptyDBPort,
		},
		{
			name: "empty redis host",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     fullDB,
				Redis:  RedisConfig{Host: "", Port: fullRedis.Port},
				Logger: fullLogger,
			},
			err: ErrEmptyRedisHost,
		},
		{
			name: "empty redis port",
			cfg: &Config{
				GRPC:   fullGRPC,
				DB:     fullDB,
				Redis:  RedisConfig{Host: fullRedis.Host, Port: ""},
				Logger: fullLogger,
			},
			err: ErrEmptyRedisPort,
		},
		{
			name: "empty logger level",
			cfg: &Config{
				GRPC:  fullGRPC,
				DB:    fullDB,
				Redis: fullRedis,
				Logger: LoggerConfig{
					Level:      "",
					OutputType: "console",
				},
			},
			err: ErrEmptyLoggerLevel,
		},
		{
			name: "empty logger outputType",
			cfg: &Config{
				GRPC:  fullGRPC,
				DB:    fullDB,
				Redis: fullRedis,
				Logger: LoggerConfig{
					Level:      "info",
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
