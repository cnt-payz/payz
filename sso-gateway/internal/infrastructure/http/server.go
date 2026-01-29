package httpserver

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/config"
)

type Server struct {
	srv *http.Server
}

func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr: net.JoinHostPort(
				cfg.Server.HTTP.Host,
				fmt.Sprint(cfg.Server.HTTP.Port),
			),
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
