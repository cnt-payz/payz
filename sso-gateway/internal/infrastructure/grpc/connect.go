package grpcclient

import (
	"fmt"
	"net"

	"github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-gateway/pkg/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectInsecureSSO(cfg *config.Config) (*grpc.ClientConn, error) {
	if cfg == nil {
		return nil, consts.ErrNilArgs
	}

	client, err := grpc.NewClient(
		net.JoinHostPort(
			cfg.Services.PayzSSO.Host,
			fmt.Sprint(cfg.Services.PayzSSO.Port),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn: %w", err)
	}

	return client, nil
}
