package app

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	ssopb "github.com/cnt-payz/payz/shop-service/api/sso/v1"
	"github.com/cnt-payz/payz/shop-service/config"
	shopusecase "github.com/cnt-payz/payz/shop-service/internal/application/usecase/shop"
	domainrepo "github.com/cnt-payz/payz/shop-service/internal/domain/repository"
	clientgrpc "github.com/cnt-payz/payz/shop-service/internal/infra/grpc/client/sso"
	cacherepo "github.com/cnt-payz/payz/shop-service/internal/infra/repository/cache"
	dbrepo "github.com/cnt-payz/payz/shop-service/internal/infra/repository/database"
	servergrpc "github.com/cnt-payz/payz/shop-service/internal/interfaces/grpc/server"
	"github.com/cnt-payz/payz/shop-service/pkg/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type App struct {
	log *slog.Logger
	cfg *config.Config

	grpcServer *grpc.Server

	shopUsecase shopusecase.ShopUsecase
	dbRepo      domainrepo.DatabaseRepo
	cacheRepo   domainrepo.CacheRepo
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
		Service:    "shop-service",
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
	l.Info("database repo created")

	cacheRepo, err := cacherepo.NewCacheRepo(cfg, l)
	if err != nil {
		l.Error("failed to create cache repo", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("cache repo created")

	shopUsecase := shopusecase.NewShopUsecase(dbRepo, cacheRepo, l)
	l.Info("shop usecase created")

	clientSSO, err := clientgrpc.NewSSOServiceClient(cfg)
	if err != nil {
		l.Error("failed to create sso client grpc", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("client sso created")

	publicKey, err := parsePublicKey(clientSSO)
	if err != nil {
		l.Error("failed to parse public key from sso service", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("public key parsed", slog.String("publicKey", publicKey.N.String()))

	grpcServer, err := servergrpc.NewGRPCServer(publicKey, shopUsecase, cfg, l)
	if err != nil {
		l.Error("failed to create grpc server", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("grpc server created")

	return &App{
		log:         l,
		cfg:         cfg,
		grpcServer:  grpcServer,
		shopUsecase: shopUsecase,
		dbRepo:      dbRepo,
		cacheRepo:   cacheRepo,
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
		app.cacheRepo.Close()
		app.dbRepo.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	select {
	case <-gracefulCh:
		app.log.Info("successfully graceful shutdown")
	case <-ctx.Done():
		app.log.Error("failed graceful shutdown, terminate")
	}
}

func parsePublicKey(client ssopb.SSOClient) (*rsa.PublicKey, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pbPublicKey, err := client.GetPublicKey(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(pbPublicKey.Body)
}
