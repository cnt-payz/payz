package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	shoppb "github.com/cnt-payz/payz/shop-gateway/api/shop/v1"
	"github.com/cnt-payz/payz/shop-gateway/config"
	shopusecase "github.com/cnt-payz/payz/shop-gateway/internal/application/usecase/shop"
	clientgrpc "github.com/cnt-payz/payz/shop-gateway/internal/infra/grpc/client"
	handlershttp "github.com/cnt-payz/payz/shop-gateway/internal/interfaces/http/handlers"
	serverhttp "github.com/cnt-payz/payz/shop-gateway/internal/interfaces/http/server"
	"github.com/cnt-payz/payz/shop-gateway/pkg/log"
	"github.com/joho/godotenv"
)

type App struct {
	log *slog.Logger
	cfg *config.Config

	server *http.Server

	httpHandlers handlershttp.HTTPHandlers
	shopUsecase  shopusecase.ShopUsecase
	client       shoppb.ShopServiceClient
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
		Service:    "shop-gateway",
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

	client, err := clientgrpc.NewShopServiceClient(cfg)
	if err != nil {
		l.Error("failed to create grpc client", slog.String("err", err.Error()))
		os.Exit(1)
	}
	l.Info("grpc client created")

	shopUsecase := shopusecase.NewShopUsecase(client, l)
	l.Info("shop usecase created")

	httpHandlers := handlershttp.NewHTTPHandlers(shopUsecase, l)
	l.Info("http handlers created")

	server := serverhttp.NewServer(cfg, httpHandlers)
	l.Info("http server created")

	return &App{
		log:          l,
		cfg:          cfg,
		server:       server,
		httpHandlers: httpHandlers,
		shopUsecase:  shopUsecase,
		client:       client,
	}
}

func (app *App) Run() {
	go func() {
		app.log.Info("http server started",
			slog.String("host", app.cfg.HTTP.Host),
			slog.String("port", app.cfg.HTTP.Port),
		)
		if err := app.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				app.log.Error("failed to run http server", slog.String("err", err.Error()))
				os.Exit(1)
			}
		}
	}()

	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	app.log.Info("graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gracefulCh := make(chan any, 1)
	defer close(gracefulCh)
	go func() {
		defer func() {
			gracefulCh <- true
		}()

		if err := app.server.Shutdown(ctx); err != nil {
			app.log.Error("failed to gracefully stop http server", slog.String("err", err.Error()))
		}
	}()

	select {
	case <-gracefulCh:
		app.log.Info("successfully graceful shutdown")
	case <-ctx.Done():
		app.log.Error("failed graceful shutdown, terminate")
	}
}
