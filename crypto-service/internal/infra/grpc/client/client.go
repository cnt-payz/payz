package grpcclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	paymentpb "github.com/cnt-payz/payz/crypto-service/api/payment/v1"
	"github.com/cnt-payz/payz/crypto-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func NewPaymentServiceClient(cfg *config.Config) (paymentpb.PaymentClient, error) {
	var paymentServiceClient paymentpb.PaymentClient
	if cfg.GRPC.TLS.Enable {
		cert, err := tls.LoadX509KeyPair(
			cfg.GRPC.TLS.ServerCert,
			cfg.GRPC.TLS.ServerKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert/key: %v", err)
		}

		caCert, err := os.ReadFile(cfg.GRPC.TLS.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %v", err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caPool,
			ServerName:   cfg.GRPC.PaymentService.Host,
		}

		conn, err := grpc.NewClient(
			fmt.Sprintf("%s:%s", cfg.GRPC.PaymentService.Host, cfg.GRPC.PaymentService.Port),
			grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect with mTLS: %v", err)
		}

		paymentServiceClient = paymentpb.NewPaymentClient(conn)
	} else {
		conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", cfg.GRPC.PaymentService.Host, cfg.GRPC.PaymentService.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}

		paymentServiceClient = paymentpb.NewPaymentClient(conn)
	}

	return paymentServiceClient, nil
}
