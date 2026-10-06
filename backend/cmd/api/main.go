package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"warehouse-management/internal/config"
	"warehouse-management/internal/handler"
	"warehouse-management/internal/middleware"
	"warehouse-management/internal/repository"
	"warehouse-management/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	itemRepository := repository.NewItemRepository(db)
	itemService := service.NewItemService(itemRepository)
	stockRepository := repository.NewStockRepository(db)
	stockService := service.NewStockService(stockRepository)
	router := handler.NewRouter(itemService, stockService)
	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           middleware.RequestLogger(logger)(middleware.CORS(cfg.CORSAllowedOrigin)(router)),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	go func() {
		logger.Info("server started", "addr", cfg.HTTPAddr(), "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
