package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/forgeflow/usermanagement/internal/config"
	"github.com/forgeflow/usermanagement/internal/repository"
	"github.com/forgeflow/usermanagement/internal/server"
	"github.com/forgeflow/usermanagement/internal/service"
)

func main() {
	cfg, err := config.Load(os.Getenv)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logger.Error("mongodb client setup failed")
		os.Exit(1)
	}
	startCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(startCtx, nil); err != nil {
		logger.Error("mongodb ping failed")
		os.Exit(1)
	}
	repo := repository.NewMongoUserRepository(client.Database(cfg.MongoDB))
	if err := repo.EnsureIndexes(startCtx); err != nil {
		logger.Error("ensuring indexes failed")
		os.Exit(1)
	}
	logger.Info("connected to mongodb")

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.NewRouter(service.NewUserService(repo), cfg, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	stop, stopFn := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopFn()
	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stopFn()
		}
	}()

	<-stop.Done()
	logger.Info("shutting down")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
	if err := client.Disconnect(shutCtx); err != nil {
		logger.Error("mongodb disconnect failed")
	}
}
