package servergrpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"

	cryptopb "github.com/cnt-payz/payz/crypto-service/api/crypto/v1"
	"github.com/cnt-payz/payz/crypto-service/config"
	walletusecase "github.com/cnt-payz/payz/crypto-service/internal/application/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type cryptoServiceServer struct {
	cryptopb.UnimplementedCryptoServiceServer
	walletUsecase walletusecase.WalletUsecase
	log           *slog.Logger
}

func NewGRPCServer(walletUsecase walletusecase.WalletUsecase, cfg *config.Config, log *slog.Logger) (*grpc.Server, error) {
	var grpcServer *grpc.Server
	if cfg.GRPC.TLS.Enable {
		cert, err := tls.LoadX509KeyPair(cfg.GRPC.TLS.ServerCert, cfg.GRPC.TLS.ServerKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load server TLS cert/key: %v", err)
		}

		caCert, err := os.ReadFile(cfg.GRPC.TLS.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %v", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    caCertPool,
		}

		grpcServer = grpc.NewServer(
			grpc.Creds(credentials.NewTLS(tlsConfig)),
		)
	} else {
		grpcServer = grpc.NewServer()
	}

	cryptopb.RegisterCryptoServiceServer(grpcServer, &cryptoServiceServer{
		walletUsecase: walletUsecase,
		log:           log,
	})

	return grpcServer, nil
}
