package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"item-packing-calculator/internal/adapters/api"
	"item-packing-calculator/internal/adapters/storage"
	"item-packing-calculator/internal/core/service"
)

const _appName = "item-packing-api"

var (
	port         string
	otelEndpoint string
)

func init() {
	flag.StringVar(&port, "port", "8080", "server port to listen on")
	flag.StringVar(&otelEndpoint, "otel-endpoint", "", "OpenTelemetry exporter endpoint (leave empty to disable tracing)")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	// initialize logger
	initLogger()

	// initialize telemetry
	tracerProvider := setupTelemetry(ctx)
	if tracerProvider != nil {
		defer func() {
			if err := tracerProvider.Shutdown(ctx); err != nil {
				logger.Printf("failed to shutdown TracerProvider: %v", err)
			}
		}()
	}

	// initialize api core
	initialSizes := []int{5000, 2000, 1000, 500, 250}
	store := storage.NewMemoryStorage(initialSizes)
	calculatorService := service.NewCalculatorService(store)
	handler := api.NewHandler(calculatorService)

	// setup http routing
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, handler)

	// chain middleware
	middlewares := setUpMiddleware()
	h := Chain(mux, middlewares...)

	// create server
	server := &http.Server{
		Handler:      h,
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// start the server
	go func() {
		logger.Info().
			Str("app", _appName).
			Str("port", port).
			Msg("starting item packing calculator server...")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server crashed")
		}
	}()

	// graceful shutdown
	waitForShutdown(server)
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-interruptChan // Block until signal received

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("graceful shutdown failed")
	}

	logger.Info().Str("app", _appName).Msg("shutting down")
	os.Exit(0)
}
