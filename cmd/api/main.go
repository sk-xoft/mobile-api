package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/smartseaman/mobile-api/internal/app"
	"github.com/smartseaman/mobile-api/internal/config"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load configuration", "error", err)
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		slog.Error("initialize application", "error", err)
		os.Exit(1)
	}
	defer application.Close()

	serverErrors := make(chan error, 1)
	go func() { serverErrors <- application.Listen() }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		slog.Info("shutting down", "signal", sig.String())
	case err := <-serverErrors:
		if !errors.Is(err, context.Canceled) {
			slog.Error("server stopped", "error", err)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := application.Shutdown(shutdownContext); err != nil {
		slog.Error("graceful shutdown", "error", err)
	}
}
