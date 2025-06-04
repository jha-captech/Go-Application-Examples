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
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	_ "github.com/jackc/pgx/v5/stdlib"

	"example.com/examples/api/layered/internal/config"
	"example.com/examples/api/layered/internal/ctxhandler"
	"example.com/examples/api/layered/internal/middleware"
	"example.com/examples/api/layered/internal/routes"
	"example.com/examples/api/layered/internal/services"
	"example.com/examples/api/layered/internal/telemetry"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "server encountered an error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Load and validate environment config
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("[in main.run] failed to load config: %w", err)
	}

	// Create a structured logger, which will print logs in json format to the
	// writer we specify.
	logger := slog.New(
		ctxhandler.WrapSlogHandler(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{
					Level: cfg.LogLevel,
				},
			),
			ctxhandler.WithAttrFunc(middleware.GetTraceIDAsAttr),
		),
	)

	otelShutdownFunc, err := telemetry.SetupOTelSDK(
		ctx,
		telemetry.Config{
			Endpoint:    "jaeger:4317",
			ServiceName: "api-layered-user-service",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to setup OpenTelemetry SDK: %w", err)
	}

	defer func() {
		err = errors.Join(err, otelShutdownFunc(ctx))
	}()

	// Create a new DB connection using environment config
	logger.DebugContext(ctx, "Connecting to and pinging the database")
	db, err := sqlx.Connect(
		"pgx", fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			cfg.DBHost,
			cfg.DBUserName,
			cfg.DBUserPassword,
			cfg.DBName,
			cfg.DBPort,
		),
	)
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

	rdb := redis.NewClient(
		&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", cfg.CacheHost, cfg.CachePort),
			Password: cfg.CachePassword,
			DB:       cfg.CacheDB,
		},
	)

	// Create a new users service
	usersService := services.NewUsersService(
		logger,
		db,
		rdb,
		time.Duration(cfg.CacheExpiration)*time.Second,
	)

	// Create a serve mux to act as our route multiplexer
	mux := telemetry.InstrumentServeMux(http.NewServeMux())

	// Add our routes to the mux
	routes.AddRoutes(mux, logger, usersService, cfg.SwaggerEnabled)

	// Wrap the mux with middleware
	wrappedMux := middleware.WrapHandler(
		mux.InstrumentRootHandler(),
		middleware.TraceID(),
		middleware.Logger(logger),
		middleware.Recover(logger),
	)

	// Create a new http server with our mux as the handler
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           wrappedMux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)

	context.AfterFunc(
		ctx, func() {
			eg.Go(
				func() error {
					<-ctx.Done()
					logger.InfoContext(ctx, "Shutting down server gracefully")

					if err := srv.Shutdown(ctx); err != nil {
						return fmt.Errorf("[in main.run] failed to shutdown server: %w", err)
					}

					if err := otelShutdownFunc(ctx); err != nil {
						return fmt.Errorf(
							"[in main.run] failed to shutdown OpenTelemetry SDK: %w",
							err,
						)
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
		return fmt.Errorf("[in main.run] error waiting for server to shut down: %w", err)
	}

	return nil
}
