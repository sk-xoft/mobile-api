package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/smartseaman/mobile-api/internal/app"
	"github.com/smartseaman/mobile-api/internal/config"
	"github.com/smartseaman/mobile-api/internal/platform/logging"
	"go.uber.org/zap"
)

func main() {
	if envFile := os.Getenv("ENV_FILE"); envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			// Parser errors may contain file contents, including credentials.
			fmt.Fprintln(os.Stderr, "Unable to load environment file:", envFile)
			os.Exit(1)
		}
	} else {
		_ = godotenv.Load()
	}
	logger, err := logging.New(os.Getenv("LOG_TIMEZONE"))
	if err != nil {
		os.Exit(1)
	}
	cleanupLogging := logging.Install(logger)
	defer cleanupLogging()
	defer func() { _ = logger.Sync() }()

	cfg, err := config.Load()
	if err != nil {
		zap.L().Error("load configuration", zap.Error(err))
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		zap.L().Error("initialize application", zap.Error(err))
		os.Exit(1)
	}
	defer application.Close()

	serverErrors := make(chan error, 1)
	go func() { serverErrors <- application.Listen() }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		zap.L().Info("shutting down", zap.String("signal", sig.String()))
	case err := <-serverErrors:
		if !errors.Is(err, context.Canceled) {
			zap.L().Error("server stopped", zap.Error(err))
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := application.Shutdown(shutdownContext); err != nil {
		zap.L().Error("graceful shutdown", zap.Error(err))
	}
}
