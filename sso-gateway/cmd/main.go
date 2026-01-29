package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cnt-payz/payz/sso-gateway/internal/app"
)

// 35d094fd092ad5ff000a8a9ed8ddc16b28ad9e2f1bf16b192ead32ee534cd7ad
func main() {
	app, cleanup, err := app.New()
	if err != nil {
		slog.Error("failed to setup app", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer cleanup()

	go func() {
		if err := app.Start(); err != nil {
			slog.Error("failed to start app", slog.String("error", err.Error()))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	app.Shutdown(context.Background())
}
