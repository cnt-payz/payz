package app

import (
	"fmt"
	"log/slog"
	"os"

	ssopb "github.com/cnt-payz/payz/sso-service/api/sso/v1"
	"github.com/cnt-payz/payz/sso-service/internal/application/services"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/cache/redis"
	sessionredis "github.com/cnt-payz/payz/sso-service/internal/infrastructure/cache/redis/session"
	userredis "github.com/cnt-payz/payz/sso-service/internal/infrastructure/cache/redis/user"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/filemanager"
	grpcserver "github.com/cnt-payz/payz/sso-service/internal/infrastructure/grpc"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/grpc/handlers"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/grpc/interceptors"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/persistence/postgres"
	userpg "github.com/cnt-payz/payz/sso-service/internal/infrastructure/persistence/postgres/user"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/security/mtls"
	"github.com/cnt-payz/payz/sso-service/internal/infrastructure/session/jwt"
	"github.com/cnt-payz/payz/sso-service/pkg/log"
	"github.com/joho/godotenv"
	redissdk "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/gorm"
)

type App struct {
	log    *slog.Logger
	server *grpcserver.Server
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) Graceful() {
	a.log.Info("graceful stopping")
	a.server.Graceful()
}

func New() (*App, func(), error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg, err := config.New(os.Getenv("PATH_CONFIG"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup log handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	userRepo, db, err := providePersistence(cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide persistence: %w", err)
	}

	sessionRepo, userCache, redisClient, err := provideCache(cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide redis: %w", err)
	}

	jwtMngr, err := jwt.New(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create jwt manager: %w", err)
	}

	fileManager := filemanager.New()

	packInterceptors := interceptors.New(jwtMngr, map[string]bool{
		ssopb.SSO_Refresh_FullMethodName:        true,
		ssopb.SSO_Register_FullMethodName:       true,
		ssopb.SSO_Login_FullMethodName:          true,
		ssopb.SSO_GetPublicKey_FullMethodName:   true,
		ssopb.SSO_GetUserByEmail_FullMethodName: true,
		ssopb.SSO_GetUserByID_FullMethodName:    true,
	})

	grpcServer, err := provideGrpcServer(cfg, packInterceptors)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide grpc server: %w", err)
	}

	service := services.New(cfg, log, userRepo, userCache, sessionRepo, jwtMngr, fileManager)
	api := handlers.New(service)
	ssopb.RegisterSSOServer(grpcServer, api)

	server := grpcserver.New(cfg, grpcServer)

	return &App{
			log:    log,
			server: server,
		}, func() {
			if err := postgres.Close(db); err != nil {
				log.Error("failed to close connection with pg", slog.String("error", err.Error()))
			} else {
				log.Info("connection with pg was closed")
			}

			if err := redis.Close(redisClient); err != nil {
				log.Error("failed to close redis client", slog.String("error", err.Error()))
			} else {
				log.Info("connection with redis was closed")
			}
		}, nil
}

func provideGrpcServer(cfg *config.Config, interceptors *interceptors.PackInterceptors) (*grpc.Server, error) {
	options := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptors.BaseInterceptor(),
			interceptors.AuthInterceptor(),
		),
	}

	if cfg.Server.GRPC.TLS.Enable {
		tlsConfig, err := mtls.LoadTLSConfig(cfg)
		if err != nil {
			return nil, err
		}

		options = append(options, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	return grpc.NewServer(options...), nil
}

func provideCache(cfg *config.Config, log *slog.Logger) (*sessionredis.SessionRepository, *userredis.UserCache, *redissdk.Client, error) {
	client, err := redis.Connect(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	sessionRepo, err := sessionredis.New(cfg, log, client)
	if err != nil {
		return nil, nil, nil, err
	}

	return sessionRepo, userredis.New(cfg, log, client), client, nil
}

func providePersistence(cfg *config.Config, log *slog.Logger) (*userpg.UserRepository, *gorm.DB, error) {
	db, err := postgres.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}
	if err := postgres.Migrate(db); err != nil {
		return nil, nil, err
	}

	userRepo, err := userpg.New(log, db)
	if err != nil {
		return nil, nil, err
	}

	return userRepo, db, nil
}
