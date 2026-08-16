package main

import (
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"go.opentelemetry.io/otel/trace"
)

var logger zerolog.Logger

// initLogger configures our shared zerolog instance
func initLogger() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Str("role", _appName).
		Logger()
}

// Chain applies middlewares to the http handler
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// setUpMiddleware returns our ordered slice of middleware functions
func setUpMiddleware() []func(http.Handler) http.Handler {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// base logger configuration
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().
		Timestamp().
		Str("role", _appName).
		Logger()

	// 60 requests per second per IP — generous for legitimate use, effective against abuse
	limiter := newRateLimiter(60, time.Second)

	return []func(http.Handler) http.Handler{
		// 1. Rate limit — outermost so abuse is rejected before any other work
		limiter.middleware,

		// 2. Inject the base logger into the request context
		hlog.NewHandler(logger),

		// 3. Extract the trace id and bind it to the logger
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				spanContext := trace.SpanContextFromContext(r.Context())
				if spanContext.HasTraceID() {
					logCtx := zerolog.Ctx(r.Context())
					logCtx.UpdateContext(func(c zerolog.Context) zerolog.Context {
						return c.Str("trace_id", spanContext.TraceID().String())
					})
				}
				next.ServeHTTP(w, r)
			})
		},

		// 4. Automatically log request parameters
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("request completed")
		}),

		// 5. Audit fields
		hlog.RemoteAddrHandler("ip"),
		hlog.UserAgentHandler("user_agent"),
		hlog.RequestIDHandler("req_id", "Request-Id"),

		// 6. Allow browser access (CORS)
		corsMiddleware,
	}
}

// corsMiddleware allows the frontend to communicate with the server
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
