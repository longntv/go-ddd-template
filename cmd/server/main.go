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

	"github.com/longntv/go-ddd-template/internal/config"
	"github.com/longntv/go-ddd-template/internal/registry"
	"github.com/longntv/go-ddd-template/internal/utils/log"
	zaputil "github.com/longntv/go-ddd-template/internal/utils/zap"
)

const (
	// gracefulShutdownTimeout is the maximum time to wait for graceful
	// shutdown.
	gracefulShutdownTimeout = 30 * time.Second

	// readHeaderTimeout bounds how long a client may take to send request
	// headers, protecting against slow-header (Slowloris) attacks.
	readHeaderTimeout = 10 * time.Second
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

	logger.Info("starting server",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.App.Port),
	)

	// Initialize server from the registry.
	server, cleanup, err := registry.InitializeServer(cfg, logger)
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
		Addr:              fmt.Sprintf(":%s", cfg.App.Port),
		Handler:           server.Router,
		ReadHeaderTimeout: readHeaderTimeout,
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

	// Run the outbox relay: it publishes the events that requests saved.
	eg.Go(func() error {
		logger.Info("outbox relay running")
		if err := server.OutboxRelay.Run(egCtx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		logger.Info("outbox relay stopped")
		return nil
	})

	// Wait for an interrupt signal, or for the server or relay to fail.
	<-egCtx.Done()
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
