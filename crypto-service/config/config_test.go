package config

import (
	"errors"
	"testing"
)

func TestValidate(t *testing.T) {
	fullGRPC := GRPCConfig{
		Host:           "localhost",
		Port:           "50051",
		TLS:            TLSConfig{Enable: false, CACert: "./certs/ca.crt", ServerCert: "./certs/shop-service.crt", ServerKey: "./certs/shop-service.key"},
		PaymentService: PaymentServiceConfig{Host: "localhost", Port: "50052", PrivateKey: "privatekey"},
	}

	fullNode := NodeConfig{
		PolygonTestnetAddress:          "https://addr",
		PolygonTestnetWebsocketAddress: "https://addr",
	}

	fullDB := DBConfig{
		Name:     "shopdb",
		User:     "user",
		Password: "pass",
		Host:     "localhost",
		Port:     "5432",
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
				Node:   fullNode,
				DB:     fullDB,
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
					TLS:  fullGRPC.TLS,
				},
				Node:   fullNode,
				DB:     fullDB,
				Logger: fullLogger,
			},
			err: ErrEmptyGRPCHost,
		},
		{
			name: "empty grpc port",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host:           fullGRPC.Host,
					Port:           "",
					PaymentService: fullGRPC.PaymentService,
					TLS:            fullGRPC.TLS,
				},
				Node:   fullNode,
				DB:     fullDB,
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
				Node:   fullNode,
				DB:     fullDB,
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
				Node:   fullNode,
				DB:     fullDB,
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
				Node:   fullNode,
				DB:     fullDB,
				Logger: fullLogger,
			},
			err: ErrEmptyServerKey,
		},
		{
			name: "empty payment host",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: fullGRPC.Port,
					TLS:  fullGRPC.TLS,
					PaymentService: PaymentServiceConfig{
						Host: "",
						Port: fullGRPC.PaymentService.Port,
					},
				},
				Node:   fullNode,
				DB:     fullDB,
				Logger: fullLogger,
			},
			err: ErrEmptyPaymentServiceHost,
		},
		{
			name: "empty payment port",
			cfg: &Config{
				GRPC: GRPCConfig{
					Host: fullGRPC.Host,
					Port: fullGRPC.Port,
					TLS:  fullGRPC.TLS,
					PaymentService: PaymentServiceConfig{
						Host: fullGRPC.PaymentService.Host,
						Port: "",
					},
				},
				Node:   fullNode,
				DB:     fullDB,
				Logger: fullLogger,
			},
			err: ErrEmptyPaymentServicePort,
		},
		{
			name: "empty node address",
			cfg: &Config{
				GRPC: fullGRPC,
				Node: NodeConfig{
					PolygonTestnetAddress:          "",
					PolygonTestnetWebsocketAddress: fullNode.PolygonTestnetWebsocketAddress,
				},
				DB:     fullDB,
				Logger: fullLogger,
			},
			err: ErrEmptyNodeAddress,
		},
		{
			name: "empty node ws address",
			cfg: &Config{
				GRPC: fullGRPC,
				Node: NodeConfig{
					PolygonTestnetAddress:          fullNode.PolygonTestnetAddress,
					PolygonTestnetWebsocketAddress: "",
				},
				DB:     fullDB,
				Logger: fullLogger,
			},
			err: ErrEmptyNodeWSAddress,
		},
		{
			name: "empty db name",
			cfg: &Config{
				GRPC:   fullGRPC,
				Node:   fullNode,
				DB:     DBConfig{Name: "", User: fullDB.User, Password: fullDB.Password, Host: fullDB.Host, Port: fullDB.Port},
				Logger: fullLogger,
			},
			err: ErrEmptyDBName,
		},
		{
			name: "empty db user",
			cfg: &Config{
				GRPC:   fullGRPC,
				Node:   fullNode,
				DB:     DBConfig{Name: fullDB.Name, User: "", Password: fullDB.Password, Host: fullDB.Host, Port: fullDB.Port},
				Logger: fullLogger,
			},
			err: ErrEmptyDBUser,
		},
		{
			name: "empty db password",
			cfg: &Config{
				GRPC:   fullGRPC,
				Node:   fullNode,
				DB:     DBConfig{Name: fullDB.Name, User: fullDB.User, Password: "", Host: fullDB.Host, Port: fullDB.Port},
				Logger: fullLogger,
			},
			err: ErrEmptyDBPassword,
		},
		{
			name: "empty db host",
			cfg: &Config{
				GRPC:   fullGRPC,
				Node:   fullNode,
				DB:     DBConfig{Name: fullDB.Name, User: fullDB.User, Password: fullDB.Password, Host: "", Port: fullDB.Port},
				Logger: fullLogger,
			},
			err: ErrEmptyDBHost,
		},
		{
			name: "empty db port",
			cfg: &Config{
				GRPC:   fullGRPC,
				Node:   fullNode,
				DB:     DBConfig{Name: fullDB.Name, User: fullDB.User, Password: fullDB.Password, Host: fullDB.Host, Port: ""},
				Logger: fullLogger,
			},
			err: ErrEmptyDBPort,
		},
		{
			name: "empty logger level",
			cfg: &Config{
				GRPC: fullGRPC,
				Node: fullNode,
				DB:   fullDB,
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
				GRPC: fullGRPC,
				Node: fullNode,
				DB:   fullDB,
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
