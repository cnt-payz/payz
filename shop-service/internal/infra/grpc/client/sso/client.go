package clientgrpc

import (
	"fmt"

	ssopb "github.com/cnt-payz/payz/shop-service/api/sso/v1"
	"github.com/cnt-payz/payz/shop-service/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewSSOServiceClient(cfg *config.Config) (ssopb.SSOClient, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", cfg.GRPC.SSO.Host, cfg.GRPC.SSO.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := ssopb.NewSSOClient(conn)

	return client, nil
}
