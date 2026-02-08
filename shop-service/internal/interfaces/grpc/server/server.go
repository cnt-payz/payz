package servergrpc

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"

	shoppb "github.com/cnt-payz/payz/shop-service/api/shop/v1"
	"github.com/cnt-payz/payz/shop-service/config"
	shopusecase "github.com/cnt-payz/payz/shop-service/internal/application/usecase/shop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type shopServiceServer struct {
	shoppb.UnimplementedShopServiceServer
	shopUsecase shopusecase.ShopUsecase
	log         *slog.Logger
}

func NewGRPCServer(publicKey *rsa.PublicKey, shopUsecase shopusecase.ShopUsecase, cfg *config.Config, log *slog.Logger) (*grpc.Server, error) {
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
			grpc.UnaryInterceptor(AuthInterceptor(publicKey)),
		)
	} else {
		grpcServer = grpc.NewServer(grpc.UnaryInterceptor(AuthInterceptor(publicKey)))
	}

	shoppb.RegisterShopServiceServer(grpcServer, &shopServiceServer{
		shopUsecase: shopUsecase,
		log:         log,
	})

	return grpcServer, nil
}
