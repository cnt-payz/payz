package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
)

func LoadTLSConfig(cfg *config.Config) (*tls.Config, error) {
	serverCert, err := tls.LoadX509KeyPair(
		cfg.Server.GRPC.TLS.ServerCert,
		cfg.Server.GRPC.TLS.ServerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load x509 pair of server: %w", err)
	}

	caBytes, err := os.ReadFile(cfg.Server.GRPC.TLS.CaCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca-cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caBytes) {
		return nil, fmt.Errorf("failed to append cert into pool: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
