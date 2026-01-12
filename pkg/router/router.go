// Package router provides system-level route registration for health checks and API documentation
package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"tacoshare-delivery-api/database"
	"tacoshare-delivery-api/docs"
	"tacoshare-delivery-api/pkg/httpx"
)

const (
	protocolHTTPS = "https"
	protocolHTTP  = "http"
)

// RegisterSystemRoutes registers system-level routes (health, docs)
func RegisterSystemRoutes(mux *http.ServeMux) {
	// Health check endpoint (under /api/v1 for proper versioning)
	mux.HandleFunc("GET /api/v1/health", handleHealth)

	// Serve OpenAPI spec JSON (with capitalized tags)
	mux.HandleFunc("GET /swagger/doc.json", handleSwaggerJSON)

	// Scalar API documentation endpoint (replaces Swagger)
	mux.HandleFunc("GET /docs", handleScalarDocs)
	mux.HandleFunc("GET /docs/", handleScalarDocs)

	// Legacy Swagger redirect (for backward compatibility)
	mux.HandleFunc("GET /swagger/{path...}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs", http.StatusMovedPermanently)
	})
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// handleSwaggerJSON serves the OpenAPI specification as JSON with capitalized tags
//
//nolint:gocyclo // Complex JSON transformation with multiple nested iterations
func handleSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	// Get swagger spec from docs package
	specJSON := docs.SwaggerInfo.ReadDoc()

	// Parse spec
	var spec map[string]any
	if err := json.Unmarshal([]byte(specJSON), &spec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update host and schemes based on request
	// Check if we're in localhost or production
	isLocal := r.Host == "localhost:8080" ||
		strings.HasPrefix(r.Host, "localhost:") ||
		strings.HasPrefix(r.Host, "127.0.0.1")

	// Determine the protocol scheme
	schemes := []string{protocolHTTP}
	if !isLocal {
		// In production (Cloud Run), default to HTTPS
		schemes = []string{protocolHTTPS}
		// Check X-Forwarded-Proto header (set by Cloud Run/load balancers)
		forwardedProto := r.Header.Get("X-Forwarded-Proto")
		if forwardedProto == protocolHTTPS {
			schemes = []string{protocolHTTPS}
		}
	}

	spec["host"] = r.Host
	spec["schemes"] = schemes

	// Capitalize tags in the tags array
	if tags, ok := spec["tags"].([]any); ok {
		for _, tag := range tags {
			if tagMap, ok := tag.(map[string]any); ok {
				if name, ok := tagMap["name"].(string); ok {
					tagMap["name"] = capitalizeFirst(name)
				}
			}
		}
	}

	// Capitalize tags in paths
	if paths, ok := spec["paths"].(map[string]any); ok {
		for _, pathItem := range paths {
			if pathMap, ok := pathItem.(map[string]any); ok {
				// Check all HTTP methods
				for _, method := range []string{"get", "post", "put", "patch", "delete", "options", "head"} {
					if operation, ok := pathMap[method].(map[string]any); ok {
						if opTags, ok := operation["tags"].([]any); ok {
							for i, tag := range opTags {
								if tagStr, ok := tag.(string); ok {
									opTags[i] = capitalizeFirst(tagStr)
								}
							}
						}
					}
				}
			}
		}
	}

	// Return modified spec
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if encodeErr := json.NewEncoder(w).Encode(spec); encodeErr != nil {
		http.Error(w, encodeErr.Error(), http.StatusInternalServerError)
		return
	}
}

// ScalarConfig representa la configuración completa de Scalar
type ScalarConfig struct {
	Metadata             map[string]string `json:"metadata"`
	Layout               string            `json:"layout"`
	Theme                string            `json:"theme"`
	OperationTitleSource string            `json:"operationTitleSource"`
	OperationsSorter     string            `json:"operationsSorter"`
	DefaultOpenAllTags   bool              `json:"defaultOpenAllTags"`
	ExpandAllResponses   bool              `json:"expandAllResponses"`
	HideClientButton     bool              `json:"hideClientButton"`
	HideModels           bool              `json:"hideModels"`
	ShowSidebar          bool              `json:"showSidebar"`
	DarkMode             bool              `json:"darkMode"`
}

// Server represents a server configuration for Scalar
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

// handleScalarDocs serves the Scalar API documentation UI
func handleScalarDocs(w http.ResponseWriter, r *http.Request) {
	// Determine the protocol (http or https)
	// Check if we're in localhost or production
	isLocal := r.Host == "localhost:8080" ||
		strings.HasPrefix(r.Host, "localhost:") ||
		strings.HasPrefix(r.Host, "127.0.0.1")

	protocol := protocolHTTP
	if !isLocal {
		// In production (Cloud Run), default to HTTPS
		protocol = protocolHTTPS
		// Check X-Forwarded-Proto header (set by Cloud Run/load balancers)
		forwardedProto := r.Header.Get("X-Forwarded-Proto")
		if forwardedProto == protocolHTTPS {
			protocol = protocolHTTPS
		}
	}

	// Build server URL (all endpoints under /api/v1)
	serverURL := fmt.Sprintf("%s://%s/api/v1", protocol, r.Host)

	// Configuración personalizada completa
	scalarConfig := map[string]any{
		"layout":               "modern",
		"defaultOpenAllTags":   true,
		"expandAllResponses":   true,
		"hideClientButton":     true,
		"hideModels":           true,
		"theme":                "fastify",
		"showSidebar":          true,
		"operationTitleSource": "summary",
		"operationsSorter":     "method",
		"darkMode":             true,
		"metadata": map[string]string{
			"title":       "🌮 TacoShare Delivery API",
			"description": "Delivery marketplace API for customers, merchants, and drivers",
		},
		"servers": []Server{
			{
				URL:         serverURL,
				Description: "API Server",
			},
		},
		"authentication": map[string]any{
			"preferredSecurityScheme": "BearerAuth",
			"securitySchemes": map[string]any{
				"BearerAuth": map[string]any{
					"token": "",
				},
			},
		},
	}

	configJSON, err := json.Marshal(scalarConfig)
	if err != nil {
		configJSON = []byte("{}")
	}

	// Generar HTML personalizado directamente
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>🌮 TacoShare Delivery API</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
    <script
        id="api-reference"
        data-url="%s://%s/swagger/doc.json"
        data-configuration='%s'></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`, protocol, r.Host, string(configJSON))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if _, err := fmt.Fprint(w, html); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// handleHealth godoc
//
//	@Summary		Health check
//	@Description	Check if the API is running and database connectivity
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httpx.JSendSuccess	"API is healthy"
//	@Router			/health [get]
func handleHealth(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if database.DB == nil {
		dbStatus = "disconnected"
	} else if err := database.DB.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	httpx.RespondSuccess(w, http.StatusOK, map[string]string{
		"status":   "healthy",
		"database": dbStatus,
	})
}
