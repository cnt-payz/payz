package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	paymentpb "github.com/cnt-payz/payz/crypto-service/api/payment/v1"
	"github.com/cnt-payz/payz/crypto-service/config"
	walletusecase "github.com/cnt-payz/payz/crypto-service/internal/application/usecase"
	domainrepo "github.com/cnt-payz/payz/crypto-service/internal/domain/repository"
	grpcclient "github.com/cnt-payz/payz/crypto-service/internal/infra/grpc/client"
	dbrepo "github.com/cnt-payz/payz/crypto-service/internal/infra/repository/database"
	"github.com/cnt-payz/payz/crypto-service/internal/infra/sdk"
	servergrpc "github.com/cnt-payz/payz/crypto-service/internal/interfaces/grpc/server"
	"github.com/cnt-payz/payz/crypto-service/pkg/log"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

type App struct {
	log *slog.Logger
	cfg *config.Config

	grpcServer    *grpc.Server
	paymentClient paymentpb.PaymentClient

	walletUsecase walletusecase.WalletUsecase
	ethSDK        walletusecase.ETHSDK
	dbRepo        domainrepo.DatabaseRepo
}

func Init() *App {
	if err := godotenv.Load(".env"); err != nil {
		slog.Error("failed to load .env file", slog.String("err", err.Error()))
		os.Exit(1)
	}

	cfg, err := config.New(os.Getenv("CONFIG_PATH"))
	if err != nil {
		slog.Error("failed to create config", slog.String("err", err.Error()))
		os.Exit(1)
	}
	slog.Info("config created", slog.Any("cfg", *cfg))

	var loggerOutputType log.OutputType
	switch cfg.Logger.OutputType {
	case "console":
		loggerOutputType = log.Console
	case "file":
		loggerOutputType = log.File
	case "both":
		loggerOutputType = log.Both
	default:
		loggerOutputType = log.Console
	}

	var loggerLevel slog.Level
	switch cfg.Logger.Level {
	case "debug":
		loggerLevel = slog.LevelDebug
	case "info":
		loggerLevel = slog.LevelInfo
	case "warn":
		loggerLevel = slog.LevelWarn
	case "error":
		loggerLevel = slog.LevelError
	}

	logCfg := log.Config{
		Service:    "crypto-service",
		OutputType: loggerOutputType,
		Level:      loggerLevel,
		JSONFormat: cfg.Logger.JSONFormat,
	}
	logHandler, err := log.NewHandler(&logCfg)
	if err != nil {
		slog.Error("failed to create log handler", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l := slog.New(logHandler)

	dbRepo, err := dbrepo.NewDatabaseRepo(cfg, l)
	if err != nil {
		l.Error("failed to create database repo", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("db repo created")

	paymentClient, err := grpcclient.NewPaymentServiceClient(cfg)
	if err != nil {
		l.Error("failed to create payment-service grpcclient", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("payment client created")

	ethSDK, err := sdk.NewETHClient(cfg, paymentClient, dbRepo, l)
	if err != nil {
		l.Error("failed to create eth client", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("ethereum sdk created")

	walletUsecase := walletusecase.NewWalletUsecase(dbRepo, ethSDK, paymentClient, cfg.GRPC.PaymentService.PrivateKey, l)
	l.Info("wallet usecase created")

	grpcServer, err := servergrpc.NewGRPCServer(walletUsecase, cfg, l)
	if err != nil {
		l.Error("failed to create grpc server", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("grpc server created")

	return &App{
		log:           l,
		cfg:           cfg,
		grpcServer:    grpcServer,
		paymentClient: paymentClient,
		walletUsecase: walletUsecase,
		ethSDK:        ethSDK,
		dbRepo:        dbRepo,
	}
}

func (app *App) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", app.cfg.GRPC.Host, app.cfg.GRPC.Port))
	if err != nil {
		app.log.Error("failed to create grpc listener", slog.String("err", err.Error()))
		os.Exit(1)
	}

	go func() {
		app.log.Info("grpc server started", slog.String("address", fmt.Sprintf("%s:%s", app.cfg.GRPC.Host, app.cfg.GRPC.Port)))
		if err := app.grpcServer.Serve(listener); err != nil {
			app.log.Error("failed to serve grpc server", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	app.log.Info("graceful shutdown...")

	gracefulCh := make(chan any, 1)
	defer close(gracefulCh)
	go func() {
		defer func() {
			gracefulCh <- true
		}()

		app.grpcServer.GracefulStop()
		app.dbRepo.Close()
		app.walletUsecase.Wait()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	select {
	case <-gracefulCh:
		app.log.Info("successfully graceful shutdown")
	case <-ctx.Done():
		app.log.Error("failed graceful shutdown, terminate")
	}
}
