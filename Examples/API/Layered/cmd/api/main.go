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

	"example.com/examples/api/layered/internal/config"
	"example.com/examples/api/layered/internal/middleware"
	"example.com/examples/api/layered/internal/routes"
	"example.com/examples/api/layered/internal/services"
	"golang.org/x/sync/errgroup"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "server encountered an error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Load and validate environment config
	cfg, err := config.New()
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
	if err != nil {
		return fmt.Errorf("[in main.run] failed to open/ping database: %w", err)
	}

	defer func() {
		logger.DebugContext(ctx, "Closing database connection")
		if err = db.Close(); err != nil {
			logger.ErrorContext(ctx, "Failed to close database connection", "err", err)
		}
	}()

	logger.InfoContext(ctx, "Connected successfully to the database")

	// Create a new users service
	usersService := services.NewUsersService(logger, db)

	// Create a serve mux to act as our route multiplexer
	mux := http.NewServeMux()

	// Add our routes to the mux
	routes.AddRoutes(mux, logger, usersService)

	// Wrap the mux with middleware
	wrappedMux := middleware.WrapHandler(mux, middleware.Logger(logger), middleware.Recover(logger))

	// Create a new http server with our mux as the handler
	srv := &http.Server{
		Addr:    ":8080",
		Handler: wrappedMux,
	}

	eg, ctx := errgroup.WithContext(ctx)

	context.AfterFunc(
		ctx, func() {
			eg.Go(
				func() error {
					if err := srv.Shutdown(ctx); err != nil {
						return fmt.Errorf("failed to shutdown server: %w", err)
					}

					return nil
				},
			)
		},
	)

	// Start the http server
	//
	// once srv.Shutdown is called, it will always return a
	// http.ErrServerClosed error and we don't care about that error.
	logger.InfoContext(ctx, "listening", slog.String("address", srv.Addr))
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("[in main.run] failed to listen and serve: %w", err)
	}

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("error waiting for server to shut down: %w", err)
	}

	return nil
}
