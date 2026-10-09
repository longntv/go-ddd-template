package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"go-ddd-template/internal/config"
	"go-ddd-template/internal/registry"
	"go-ddd-template/internal/utils/log"
	zaputil "go-ddd-template/internal/utils/zap"
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

//go:generate go run github.com/google/wire/cmd/wire@latest
func main() {
	// Load configuration.
	cfg := config.Load()

	// Initialize logger.
	logger, err := zaputil.New(cfg)
	if err != nil {
		log.Fatal("failed to create logger", zap.Error(err))
	}
	log.SetLogger(logger)
	defer logger.Sync()

	logger.Info("starting server",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// Initialize server from the registry.
	router, cleanup, err := registry.InitializeServer(cfg, logger)
	if err != nil {
		logger.Fatal("failed to initialize server", zap.Error(err))
	}
	defer cleanup()

	// Setup signal handling for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize error group for managing goroutines.
	eg, egCtx := errgroup.WithContext(ctx)

	// Run HTTP server.
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.App.Port),
		Handler: router,
		BaseContext: func(net.Listener) context.Context {
			return egCtx
		},
	}

	eg.Go(func() error {
		logger.Info("http server listening", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Info("http server stopped gracefully")
				return nil
			}
			return err
		}
		return nil
	})

	// Wait for interrupt signal or error from server.
	<-ctx.Done()
	logger.Info("received shutdown signal, initiating graceful shutdown")
	stop()

	// Create a separate context for shutdown operations with timeout.
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		gracefulShutdownTimeout,
	)
	defer shutdownCancel()

	// Gracefully stop the http server with timeout enforcement.
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown server gracefully", zap.Error(err))
	}

	// Wait for all goroutines to finish.
	if err := eg.Wait(); err != nil {
		logger.Panic("server returning an error", zap.Error(err))
	}

	logger.Info("server exited successfully")
}
