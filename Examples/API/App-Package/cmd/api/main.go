package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"

	_ "github.com/jackc/pgx/v5/stdlib"

	"example.com/examples/api/app-package/internal/app"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %s", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Load and validate environment config
	cfg, err := app.NewConfig()
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load config: %w", err)
	}

	// Create a structured logger, which will print logs in json format to the
	// writer we specify.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	// Create a new DB connection using environment config
	logger.DebugContext(ctx, "Connecting to and pinging the database")
	db, err := sqlx.Connect("pgx", fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBUserName,
		cfg.DBUserPassword,
		cfg.DBName,
		cfg.DBPort,
	))

	defer func() {
		logger.DebugContext(ctx, "Closing database connection")
		if err = db.Close(); err != nil {
			logger.ErrorContext(ctx, "Failed to close database connection", "err", err)
		}
	}()

	// create handler and wrap in middleware
	handler := app.NewHandler(logger, db)
	wrappedHandler := app.WrapHandler(
		handler,
		app.TraceIDMiddleware(),
		app.LoggingMiddleware(logger),
		app.RecoveryMiddleware(logger),
	)

	// Create a new http server with our mux as the handler
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: wrappedHandler,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a new errgroup to handle graceful shutdown
	eg, ctx := errgroup.WithContext(ctx)

	context.AfterFunc(
		ctx, 
		func() {
			eg.Go(
				func() error {
					if err := httpServer.Shutdown(ctx); err != nil {
						return fmt.Errorf("failed to shutdown server: %w", err)
					}

					return nil
				},
			)
		},
	)

	// Start the server
	logger.InfoContext(ctx, "Starting HTTP server", "addr", httpServer.Addr)

	if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("[in main.run] failed to listen and serve: %w", err)
	}

	return nil
}
