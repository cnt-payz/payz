package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/IBM/sarama"
	paymentpb "github.com/cnt-payz/payz/payment-service/api/payment/v1"
	"github.com/cnt-payz/payz/payment-service/internal/application/services"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/cache/redis"
	idempotencyredis "github.com/cnt-payz/payz/payment-service/internal/infrastructure/cache/redis/idempotency"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/config"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/grpc/server"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/grpc/server/handlers"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/grpc/server/interceptors"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/kafka"
	notificationkafka "github.com/cnt-payz/payz/payment-service/internal/infrastructure/kafka/notification"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/persistence/postgres"
	extransactionpg "github.com/cnt-payz/payz/payment-service/internal/infrastructure/persistence/postgres/extransaction"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/security/mtls"
	"github.com/cnt-payz/payz/payment-service/internal/infrastructure/session/jwt"
	"github.com/cnt-payz/payz/payment-service/pkg/log"
	"github.com/joho/godotenv"
	redissdk "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/gorm"
)

type App struct {
	log    *slog.Logger
	server *server.Server
}

func (a *App) Start() error {
	a.log.Info("server is running")
	return a.server.Start()
}

func (a *App) GracefulStop() {
	a.log.Info("gracefully shutdown")
	a.server.GracefulStop()
}

func New() (*App, func(), error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, nil, fmt.Errorf("failed to load .env: %w", err)
	}

	cfg, err := config.New(os.Getenv("PATH_CONFIG"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	logHandler, err := log.SetupHandler(os.Stdout, cfg.Env)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup log handler: %w", err)
	}
	log := slog.New(logHandler)

	log.Info("config was loaded", slog.Any("server", cfg.Server), slog.Any("app", cfg.App))

	redisClient, idempotencyRepo, err := provideRedis(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide redis: %w", err)
	}

	db, extransactionRepo, err := providePostgres(cfg, log)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide postgres: %w", err)
	}
	go extransactionRepo.HandleDeadlines(context.Background())

	producer, notificationRepo, err := provideKafka(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide kafka: %w", err)
	}

	service := services.New(cfg, log, idempotencyRepo, extransactionRepo, notificationRepo)
	grpcServer, err := provideServer(cfg, service)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to provide server: %w", err)
	}
	server := server.New(cfg, grpcServer)

	return &App{
			log:    log,
			server: server,
		}, func() {
			if err := postgres.Close(db); err != nil {
				log.Error("failed to close connection with postgres", slog.String("error", err.Error()))
			} else {
				log.Info("connection with postgres is closed")
			}

			if err := redis.Close(redisClient); err != nil {
				log.Error("failed to close connection with redis", slog.String("error", err.Error()))
			} else {
				log.Info("connection with redis is closed")
			}

			if err := kafka.Close(producer); err != nil {
				log.Error("failed to close connection with kafka", slog.String("error", err.Error()))
			} else {
				log.Info("connection with kafka was closed")
			}
		}, nil
}

func provideServer(cfg *config.Config, service services.PaymentService) (*grpc.Server, error) {
	jwtMngr, err := jwt.New(cfg)
	if err != nil {
		return nil, err
	}

	packInterceptors := interceptors.New(jwtMngr, map[string]bool{
		paymentpb.Payment_MakeExTransaction_FullMethodName:    true,
		paymentpb.Payment_ConfirmExTransaction_FullMethodName: true,
		paymentpb.Payment_CancelExTransaction_FullMethodName:  true,
	}, map[string]bool{
		paymentpb.Payment_MakeExTransaction_FullMethodName:    true,
		paymentpb.Payment_ConfirmExTransaction_FullMethodName: true,
		paymentpb.Payment_CancelExTransaction_FullMethodName:  true,
	})

	options := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			packInterceptors.AuthInterceptor(),
			packInterceptors.IdempotencyInterceptor(),
		),
	}
	if cfg.Server.GRPC.TLS.Enable {
		tlsConfig, err := mtls.LoadTLSConfig(cfg)
		if err != nil {
			return nil, err
		}

		options = append(options, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	grpcServer := grpc.NewServer(options...)
	api := handlers.New(service)
	paymentpb.RegisterPaymentServer(grpcServer, api)

	return grpcServer, nil
}

func provideRedis(cfg *config.Config) (*redissdk.Client, *idempotencyredis.IdempotencyRepository, error) {
	client, err := redis.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}

	return client, idempotencyredis.New(cfg, client), nil
}

func providePostgres(cfg *config.Config, log *slog.Logger) (*gorm.DB, *extransactionpg.ExTransactionRepository, error) {
	db, err := postgres.Connect(cfg)
	if err != nil {
		return nil, nil, err
	}

	return db, extransactionpg.New(cfg, log, db), nil
}

func provideKafka(cfg *config.Config) (sarama.SyncProducer, *notificationkafka.NotificationClient, error) {
	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, nil, err
	}

	return producer, notificationkafka.New(cfg, producer), nil
}
