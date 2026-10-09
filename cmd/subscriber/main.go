package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/longntv/go-ddd-template/internal/config"
	"github.com/longntv/go-ddd-template/internal/registry"
	"github.com/longntv/go-ddd-template/internal/utils/log"
	zaputil "github.com/longntv/go-ddd-template/internal/utils/zap"
)

const (
	// gracefulShutdownTimeout is the maximum time to wait for graceful
	// shutdown.
	gracefulShutdownTimeout = 30 * time.Second
)

var (
	// Version is the application version.
	Version = "dev"
	// BuildTime is the build timestamp.
	BuildTime = "unknown"
)

func main() {
	// Load configuration.
	cfg := config.Load()

	// Initialize logger.
	logger, err := zaputil.New(cfg)
	if err != nil {
		log.Fatal("failed to create logger", zap.Error(err))
	}
	log.SetLogger(logger)
	// Sync can fail on stderr/stdout (EINVAL on some OSes); nothing useful to do on exit.
	defer func() { _ = logger.Sync() }()

	logger.Info("starting subscriber",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("env", cfg.App.Env),
		zap.String("queue_url", cfg.SQS.QueueURL),
	)

	// Initialize subscriber from the registry.
	subscriber, cleanup, err := registry.InitializeSubscriber(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize subscriber", zap.Error(err))
	}
	defer cleanup()

	// Setup signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize error group for managing goroutines.
	eg, egCtx := errgroup.WithContext(ctx)

	// Run subscriber.
	eg.Go(func() error {
		logger.Info("subscriber starting")
		if err := subscriber.Run(egCtx); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("subscriber stopped gracefully")
				return nil
			}
			return err
		}
		return nil
	})

	// Wait for interrupt signal or error from subscriber.
	<-ctx.Done()
	logger.Info("received shutdown signal, initiating graceful shutdown")
	stop()

	// Create a separate context for shutdown operations with timeout.
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		gracefulShutdownTimeout,
	)
	defer shutdownCancel()

	// Shutdown the subscriber.
	if err := subscriber.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown subscriber gracefully", zap.Error(err))
	}

	// Wait for all goroutines to finish.
	if err := eg.Wait(); err != nil {
		logger.Panic("subscriber returning an error", zap.Error(err))
	}

	logger.Info("subscriber exited successfully")
}
