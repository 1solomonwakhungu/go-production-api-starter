package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1solomonwakhungu/go-production-api-starter/internal/config"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/handler"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/server"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/service"
	_ "github.com/lib/pq"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		logger.Error("open database", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("close database", slog.Any("error", err))
		}
	}()

	if err := repository.PingDB(db); err != nil {
		logger.Error("connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	repo := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(repo)
	apiServer := server.New(cfg, handler.New(userService), logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- apiServer.Start()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve http", slog.Any("error", err))
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := apiServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("shut down server", slog.Any("error", err))
			os.Exit(1)
		}
	}
}
