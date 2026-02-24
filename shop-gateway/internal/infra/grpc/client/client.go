package clientgrpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	shoppb "github.com/cnt-payz/payz/shop-gateway/api/shop/v1"
	"github.com/cnt-payz/payz/shop-gateway/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewShopServiceClient(cfg *config.Config) (shoppb.ShopServiceClient, error) {
	var shopServiceClient shoppb.ShopServiceClient
	if cfg.GRPC.TLS.Enable {
		caCert, err := os.ReadFile(cfg.GRPC.TLS.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %v", err)
		}

		clientCert, err := tls.LoadX509KeyPair(
			cfg.GRPC.TLS.ClientCert,
			cfg.GRPC.TLS.ClientKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert/key: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA cert")
		}

		tlsConfig := &tls.Config{
			ServerName:         cfg.GRPC.ShopService.Host,
			RootCAs:            caCertPool,
			Certificates:       []tls.Certificate{clientCert},
			InsecureSkipVerify: false,
		}

		creds := credentials.NewTLS(tlsConfig)

		conn, err := grpc.NewClient(
			fmt.Sprintf("%s:%s", cfg.GRPC.ShopService.Host, cfg.GRPC.ShopService.Port),
			grpc.WithTransportCredentials(creds),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to shop service: %v", err)
		}

		shopServiceClient = shoppb.NewShopServiceClient(conn)
	} else {
		conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", cfg.GRPC.ShopService.Host, cfg.GRPC.ShopService.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}

		shopServiceClient = shoppb.NewShopServiceClient(conn)
	}

	return shopServiceClient, nil
}
