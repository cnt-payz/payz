package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
)

func LoadTLSConfig(cfg *config.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(
		cfg.Server.GRPC.TLS.ServerCert,
		cfg.Server.GRPC.TLS.ServerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load x509 pair of server: %w", err)
	}

	caCert, err := os.ReadFile(cfg.Server.GRPC.TLS.CaCert)
	if err != nil {
		return nil, fmt.Errorf("faield to read ca cert file: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append ca cert: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	}, nil
}
