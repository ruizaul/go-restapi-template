package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/joho/godotenv"

	"go-api-template/database"
	"go-api-template/internal/auth"
	"go-api-template/internal/users"
	"go-api-template/pkg/config"
	"go-api-template/pkg/middleware"
	"go-api-template/pkg/response"

	_ "go-api-template/docs"
)

//	@title			Go API Template
//	@version		1.0.0
//	@description	A modern Go REST API template with best practices

//	@contact.name	API Support
//	@contact.url	http://www.example.com/support
//	@contact.email	support@example.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@BasePath	/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

//	@accept		json
//	@produce	json

func main() {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load() //nolint:errcheck // .env file is optional

	// Load configuration
	cfg := config.Load()

	// Setup structured logger
	logger := setupLogger(cfg)

	// Connect to database
	if err := database.Connect(); err != nil {
		logger.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() {
		if err := database.Close(); err != nil {
			logger.Error("database close failed", slog.String("error", err.Error()))
		}
	}()

	// Create HTTP router
	mux := http.NewServeMux()

	// Register routes
	registerRoutes(mux, cfg)

	// Setup middleware chain
	handler := setupMiddleware(mux, logger, cfg)

	// Create HTTP server with production-ready timeouts
	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           handler,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("🚀 Server ready",
			slog.String("port", cfg.Server.Port),
			slog.String("docs", fmt.Sprintf("http://localhost:%s/docs", cfg.Server.Port)),
		)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	gracefulShutdown(server, logger, cfg.Server.ShutdownTimeout)
}

// customTextHandler creates a cleaner text handler for development
type customTextHandler struct {
	w     io.Writer
	level slog.Level
}

func (h *customTextHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *customTextHandler) Handle(_ context.Context, r slog.Record) error {
	// Get level with color
	levelStr := ""
	switch r.Level {
	case slog.LevelDebug:
		levelStr = "\033[36mDEBUG\033[0m"
	case slog.LevelInfo:
		levelStr = "\033[32mINFO\033[0m "
	case slog.LevelWarn:
		levelStr = "\033[33mWARN\033[0m "
	case slog.LevelError:
		levelStr = "\033[31mERROR\033[0m"
	}

	// Build log line
	var buf strings.Builder
	buf.WriteString(r.Time.Format("15:04:05"))
	buf.WriteString(" ")
	buf.WriteString(levelStr)
	buf.WriteString(" ")
	buf.WriteString(r.Message)

	// Add attributes
	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(" ")
		buf.WriteString(a.Key)
		buf.WriteString("=")
		buf.WriteString(a.Value.String())
		return true
	})

	buf.WriteString("\n")
	_, err := h.w.Write([]byte(buf.String()))
	return err
}

func (h *customTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *customTextHandler) WithGroup(name string) slog.Handler {
	return h
}

// setupLogger creates a structured logger based on configuration
func setupLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	// Set log level
	var level slog.Level
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Set log format
	if cfg.Log.Format == "json" || cfg.IsProduction() {
		opts := &slog.HandlerOptions{
			Level:     level,
			AddSource: cfg.Log.AddSource,
		}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Use custom text handler for cleaner development logs
		handler = &customTextHandler{
			w:     os.Stdout,
			level: level,
		}
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

// setupMiddleware chains all middleware in the correct order
func setupMiddleware(handler http.Handler, logger *slog.Logger, cfg *config.Config) http.Handler {
	// Build middleware chain (order matters - first is outermost)
	middlewares := []func(http.Handler) http.Handler{
		middleware.Recovery(logger),                         // Recover from panics first
		middleware.Logging(logger),                          // Log all requests
		middleware.CORS(middleware.CORSConfig{               // Handle CORS
			AllowedOrigins:   cfg.CORS.AllowedOrigins,
			AllowedMethods:   cfg.CORS.AllowedMethods,
			AllowedHeaders:   cfg.CORS.AllowedHeaders,
			AllowCredentials: cfg.CORS.AllowCredentials,
			MaxAge:           cfg.CORS.MaxAge,
		}),
	}

	// Add rate limiting if enabled
	if cfg.RateLimit.Enabled {
		middlewares = append(middlewares, middleware.RateLimit(middleware.RateLimitConfig{
			Rate:            cfg.RateLimit.Rate,
			Window:          cfg.RateLimit.Window,
			CleanupInterval: 5 * time.Minute,
		}))
	}

	return middleware.Chain(handler, middlewares...)
}

// registerRoutes registers all application routes
func registerRoutes(mux *http.ServeMux, cfg *config.Config) {
	// Health check endpoint (checks database connectivity)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		health := map[string]any{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}

		// Check database health
		if err := database.Health(r.Context()); err != nil {
			health["status"] = "unhealthy"
			health["database"] = map[string]string{
				"status": "unhealthy",
				"error":  err.Error(),
			}
			response.Error(w, http.StatusServiceUnavailable, "Service unhealthy")
			return
		}

		health["database"] = map[string]string{"status": "healthy"}
		response.Success(w, health)
	})

	// Liveness probe (simple check - server is running)
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) {
		response.Success(w, map[string]string{"status": "alive"})
	})

	// Readiness probe (checks if ready to accept traffic)
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		if err := database.Health(r.Context()); err != nil {
			response.Error(w, http.StatusServiceUnavailable, "Not ready")
			return
		}
		response.Success(w, map[string]string{"status": "ready"})
	})

	// Serve swagger.json directly
	mux.HandleFunc("GET /docs/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "./docs/swagger.json")
	})

	// API documentation with Scalar
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, _ *http.Request) {
		html, err := scalargo.NewV2(
			scalargo.WithSpecDir("./docs"),
			scalargo.WithBaseFileName("openapi.json"),
			scalargo.WithDarkMode(),
			scalargo.WithLayout(scalargo.LayoutModern),
			scalargo.WithMetaDataOpts(
				scalargo.WithTitle("Go API Template - Documentation"),
				scalargo.WithKeyValue("defaultOpenAllTags", true),
				scalargo.WithKeyValue("expandAllModelSections", true),
				scalargo.WithKeyValue("expandAllResponses", true),
			),
			scalargo.WithSidebarVisibility(true),
			scalargo.WithShowToolbar(scalargo.ShowToolbarLocalhost),
			scalargo.WithOperationTitleSource(scalargo.OperationTitleSourceSummary),
			scalargo.WithPersistAuth(false),
			scalargo.WithHideSearch(false),
			scalargo.WithShowOperationID(false),
			scalargo.WithOrderSchemaPropertiesBy(scalargo.SchemaPropertiesOrderAlpha),
			scalargo.WithDefaultFonts(),
		)
		if err != nil {
			response.InternalError(w, fmt.Sprintf("Error generating documentation: %v", err))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//nolint:errcheck // Response write errors are not recoverable
		fmt.Fprint(w, html)
	})

	// Register auth routes (returns jwtService for protecting other routes)
	jwtService := auth.RegisterRoutes(mux, database.DB, cfg)

	// Register feature routes (protected with auth)
	users.RegisterRoutes(mux, database.DB, jwtService)
}

// gracefulShutdown handles graceful server shutdown on interrupt signals
func gracefulShutdown(server *http.Server, logger *slog.Logger, timeout time.Duration) {
	// Create channel to listen for signals
	quit := make(chan os.Signal, 1)

	// Notify on SIGINT (Ctrl+C) and SIGTERM (Docker/K8s stop)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until signal received
	sig := <-quit
	logger.Info("⏹️  Shutting down", slog.String("signal", sig.String()))

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	var shutdownErr error
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
		shutdownErr = err
	}

	// Cancel context after shutdown attempt
	cancel()

	// Close database connection
	if err := database.Close(); err != nil {
		logger.Error("database close error", slog.String("error", err.Error()))
	}

	if shutdownErr != nil {
		os.Exit(1)
	}

	logger.Info("✓ Shutdown complete")
}
