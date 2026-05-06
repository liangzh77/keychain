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

	"github.com/liangzh77/keychain/internal/auth"
	"github.com/liangzh77/keychain/internal/config"
	keydb "github.com/liangzh77/keychain/internal/db"
	"github.com/liangzh77/keychain/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load(".env")
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	database, err := keydb.Open(context.Background(), cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.Migrate(context.Background()); err != nil {
		logger.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	authService, err := auth.NewService(auth.Options{
		DB:            database.SQL(),
		AdminUsername: cfg.AdminUsername,
		AdminPassword: cfg.AdminPassword,
		SessionSecret: cfg.SessionSecret,
	})
	if err != nil {
		logger.Error("failed to initialize auth service", "error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      server.NewRouter(server.Options{Now: time.Now, HealthCheck: database.Ping, Auth: authService}),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("server shutting down")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
}
