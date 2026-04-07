package server

import (
	"fmt"
	"net"

	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
	"google.golang.org/grpc"
)

type Server struct {
	cfg        *config.Config
	grpcServer *grpc.Server
}

func New(cfg *config.Config, grpcServer *grpc.Server) *Server {
	return &Server{
		cfg:        cfg,
		grpcServer: grpcServer,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen(
		s.cfg.Server.GRPC.Protocol,
		net.JoinHostPort(
			s.cfg.Server.GRPC.Host,
			fmt.Sprint(s.cfg.Server.GRPC.Port),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve listener: %w", err)
	}

	return nil
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}
