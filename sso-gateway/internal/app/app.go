package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	ssopb "github.com/cnt-payz/payz/sso-gateway/api/sso/v1"
	"github.com/cnt-payz/payz/sso-gateway/internal/application/services"
	"github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/config"
	grpcclient "github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/grpc"
	httpserver "github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/http"
	"github.com/cnt-payz/payz/sso-gateway/internal/infrastructure/http/handlers"
	"github.com/cnt-payz/payz/sso-gateway/pkg/log"
	"github.com/joho/godotenv"
)

type App struct {
	log    *slog.Logger
	server *httpserver.Server
}

func New() (*App, func(), error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env")
	}

	cfg, err := config.New(os.Getenv("PATH_CONFIG"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	clientConn, err := grpcclient.ConnectInsecureSSO(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect sso: %w", err)
	}
	ssoClient := ssopb.NewSSOClient(clientConn)

	service := services.New(cfg, log, ssoClient)
	handler, err := handlers.New(cfg, service)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load handler: %w", err)
	}

	return &App{
			log:    log,
			server: httpserver.New(cfg, handler),
		}, func() {
			if err := clientConn.Close(); err != nil {
				log.Error("failed to close connection with sso", slog.String("error", err.Error()))
			} else {
				log.Info("connection with sso was closed")
			}
		}, nil
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.log.Info("server shutdown")
	return a.server.Shutdown(ctx)
}
