package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gyud-adb/paris-api/internal/infrastructure/di"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// run boots the application server.
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := di.Bootstrap(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = application.Shutdown()
	}()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- application.Server.ListenAndServe()
	}()

	workerCount := 0
	if application.ReActClassificationWorker != nil {
		workerCount++
	}

	var workerErrors chan error
	if workerCount > 0 {
		workerErrors = make(chan error, workerCount)

		if application.ReActClassificationWorker != nil {
			go func() {
				workerErrors <- application.ReActClassificationWorker.Run(ctx)
			}()
		}
	}

	application.Logger.Info("http server started", zap.String("address", application.Server.Addr))

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case err := <-workerErrors:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		application.Logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), application.Config.HTTP.ShutdownTimeout)
	defer cancel()

	if err := application.Server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	application.Logger.Info("http server stopped")

	return nil
}
