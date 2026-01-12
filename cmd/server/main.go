package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	scalargo "github.com/bdpiprava/scalar-go"

	"go-api-template/internal/test"
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

//	@host		localhost:8080
//	@BasePath	/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

//	@accept		json
//	@produce	json

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs all HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := newResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log the request
		log.Printf(
			"[%s] %s %s %d %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			wrapped.statusCode,
			time.Since(start),
		)
	})
}

func main() {
	// Configure logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetPrefix("[API] ")

	// Create HTTP router
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		response.Success(w, map[string]string{"message": "Server is running"})
	})

	// Serve swagger.json directly
	mux.HandleFunc("GET /docs/swagger.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, nil, "./docs/swagger.json")
	})

	// API documentation with Scalar
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, _ *http.Request) {
		html, err := scalargo.NewV2(
			scalargo.WithSpecDir("./docs"),
			scalargo.WithBaseFileName("openapi.json"),
			scalargo.WithTheme(scalargo.ThemeAlternate),
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

	// Register feature routes
	test.RegisterRoutes(mux)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with production-ready timeouts
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           loggingMiddleware(mux),
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("🚀 Server starting on http://localhost:%s", port)
	log.Printf("📚 API Documentation: http://localhost:%s/docs", port)
	log.Printf("💚 Health Check: http://localhost:%s/health", port)

	if err := server.ListenAndServe(); err != nil {
		log.Printf("❌ Server error: %v", err)
		os.Exit(1)
	}
}
