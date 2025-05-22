package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"

	"example.com/examples/api/app-package/internal/telemetry"
)

const name = "example.com/examples/api/api-package/cmd/api/main"

var tracer = otel.Tracer(name)

func main() {
	if err := run(context.Background()); err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "Error running server: %s\n", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	otelShutdownFunc, err := telemetry.SetupOTelSDK(
		ctx,
		telemetry.Config{
			JaegerEndpoint: "jaeger:4317",
			ServiceName:    "api-package-user-service",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to setup OpenTelemetry SDK: %w", err)
	}

	defer func() {
		err = errors.Join(err, otelShutdownFunc(ctx))
	}()

	mux := http.NewServeMux()

	telemetry.HandleFunc(mux, "GET /hello-world/{id}", helloHandler)

	handler := telemetry.NewRootInstrumentedHandler(mux, "my-service")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
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

	// Start the server
	log.Println("Starting server on :8080")
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %w", err)
	}

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("error waiting for server to shut down: %w", err)
	}

	return nil
}

// HelloHandler responds with a hello world message
func helloHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "helloHandler")
	defer span.End()

	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"message": "Hello, World v3!",
		},
	)
}
