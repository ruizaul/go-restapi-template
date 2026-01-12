package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"tacoshare-delivery-api/config"
	"tacoshare-delivery-api/database"
	"tacoshare-delivery-api/docs"
	"tacoshare-delivery-api/pkg/envx"
	"tacoshare-delivery-api/pkg/middleware"
	"tacoshare-delivery-api/pkg/otp"
	"tacoshare-delivery-api/pkg/router"
	"tacoshare-delivery-api/pkg/storage"

	// Features
	"tacoshare-delivery-api/internal/auth"
	authHandlers "tacoshare-delivery-api/internal/auth/handlers"
	authRepos "tacoshare-delivery-api/internal/auth/repositories"
	authServices "tacoshare-delivery-api/internal/auth/services"

	"tacoshare-delivery-api/internal/documents"
	documentHandlers "tacoshare-delivery-api/internal/documents/handlers"
	documentRepos "tacoshare-delivery-api/internal/documents/repositories"
	documentServices "tacoshare-delivery-api/internal/documents/services"

	"tacoshare-delivery-api/internal/drivers"
	driverAdapters "tacoshare-delivery-api/internal/drivers/adapters"
	driverHandlers "tacoshare-delivery-api/internal/drivers/handlers"
	driverRepos "tacoshare-delivery-api/internal/drivers/repositories"
	driverServices "tacoshare-delivery-api/internal/drivers/services"

	"tacoshare-delivery-api/internal/merchants"
	merchantHandlers "tacoshare-delivery-api/internal/merchants/handlers"
	merchantRepos "tacoshare-delivery-api/internal/merchants/repositories"
	merchantServices "tacoshare-delivery-api/internal/merchants/services"

	"tacoshare-delivery-api/internal/notifications"
	notificationHandlers "tacoshare-delivery-api/internal/notifications/handlers"
	notificationRepos "tacoshare-delivery-api/internal/notifications/repositories"
	notificationServices "tacoshare-delivery-api/internal/notifications/services"

	"tacoshare-delivery-api/internal/orders"
	orderHandlers "tacoshare-delivery-api/internal/orders/handlers"
	orderRepos "tacoshare-delivery-api/internal/orders/repositories"
	orderServices "tacoshare-delivery-api/internal/orders/services"

	"tacoshare-delivery-api/internal/users"
	userHandlers "tacoshare-delivery-api/internal/users/handlers"
	userRepos "tacoshare-delivery-api/internal/users/repositories"
	userServices "tacoshare-delivery-api/internal/users/services"

	"tacoshare-delivery-api/internal/websockets"
	wsHandlers "tacoshare-delivery-api/internal/websockets/handlers"
	wsServices "tacoshare-delivery-api/internal/websockets/services"

	"tacoshare-delivery-api/pkg/gmaps"
)

// userRepositoryAdapter adapts the user repository for document service
type userRepositoryAdapter struct {
	userRepo *userRepos.UserRepository
}

func (a *userRepositoryAdapter) FindByID(id uuid.UUID) (*documentServices.User, error) {
	user, err := a.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	return &documentServices.User{ID: user.ID}, nil
}

//	@title			TacoShare Delivery API
//	@version		1.0
//	@description	Delivery marketplace API for customers, merchants, and drivers.

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token. Example: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

//	@accept		json
//	@produce	json

func main() {
	// Load environment variables with priority: .env.local > .env.{ENV} > .env
	if err := envx.LoadEnv(); err != nil {
		panic(err)
	}

	// Update Swagger host dynamically based on BASE_URL
	baseURL := os.Getenv("BASE_URL")
	if baseURL != "" {
		updateSwaggerHost(baseURL)
	}

	// Initialize Twilio client
	twilioConfig := config.LoadTwilioConfig()
	otp.InitializeTwilio(
		twilioConfig.AccountSID,
		twilioConfig.APIKey,
		twilioConfig.APISecret,
		twilioConfig.FromPhone,
		twilioConfig.Enabled,
	)

	// Connect to database
	// Note: We allow the server to start even if DB connection fails
	// This enables Cloud Run health checks to pass while troubleshooting
	dbConnected := false
	if err := database.Connect(); err != nil {
		// Server will start but database operations will fail
	} else {
		dbConnected = true
	}

	// Defer database close only if connected
	if dbConnected {
		defer func() {
			_ = database.Close()
		}()
	}

	// Initialize repositories
	authRepo := authRepos.NewUserRepository(database.DB)
	refreshTokenRepo := authRepos.NewRefreshTokenRepository(database.DB)
	userRepo := userRepos.NewUserRepository(database.DB)
	documentRepo := documentRepos.NewDocumentRepository(database.DB)
	notificationRepo := notificationRepos.NewNotificationRepository(database.DB)
	fcmTokenRepo := notificationRepos.NewFCMTokenRepository(database.DB)
	merchantRepo := merchantRepos.NewMerchantRepository(database.DB)
	orderRepo := orderRepos.NewOrderRepository(database.DB)
	assignmentRepo := orderRepos.NewAssignmentRepository(database.DB)
	locationRepo := driverRepos.NewLocationRepository(database.DB)

	// Initialize services
	authService := authServices.NewAuthService(authRepo, refreshTokenRepo)
	userService := userServices.NewUserService(userRepo)

	// Create adapter for user repository to work with document service
	userRepoAdapter := &userRepositoryAdapter{userRepo: userRepo}
	documentService := documentServices.NewDocumentService(documentRepo, userRepoAdapter)

	// Initialize R2 storage client (optional - for document uploads)
	// Note: We allow the server to start even if R2 is not configured
	// This enables Cloud Run health checks to pass while troubleshooting
	r2Config := storage.LoadConfigFromEnv()
	r2Client, err := storage.NewR2Client(r2Config)
	if err != nil {
		r2Client = nil // Set to nil to indicate R2 is not available
	}

	// Initialize FCM service (optional - will be nil if credentials not provided)
	var notificationService *notificationServices.NotificationService

	// Try JSON credentials first (for Cloud Run with Secret Manager)
	fcmCredentialsJSON := os.Getenv("FCM_CREDENTIALS_JSON")
	if fcmCredentialsJSON != "" {
		fcmService, err := notificationServices.NewFCMServiceFromJSON(context.Background(), fcmCredentialsJSON)
		if err != nil {
			notificationService = notificationServices.NewNotificationService(notificationRepo, fcmTokenRepo, nil)
		} else {
			notificationService = notificationServices.NewNotificationService(notificationRepo, fcmTokenRepo, fcmService)
		}
	} else {
		// Fallback to file path (for local development)
		fcmCredentialsPath := os.Getenv("FCM_CREDENTIALS_PATH")
		if fcmCredentialsPath != "" {
			fcmService, err := notificationServices.NewFCMService(context.Background(), fcmCredentialsPath)
			if err != nil {
				notificationService = notificationServices.NewNotificationService(notificationRepo, fcmTokenRepo, nil)
			} else {
				notificationService = notificationServices.NewNotificationService(notificationRepo, fcmTokenRepo, fcmService)
			}
		} else {
			notificationService = notificationServices.NewNotificationService(notificationRepo, fcmTokenRepo, nil)
		}
	}

	// Initialize Google Maps client
	gmapsClient, _ := gmaps.NewClient()

	// Initialize WebSocket hub
	wsHub := wsServices.NewHub()
	go wsHub.Run()

	// Create WebSocket hub adapter for assignment service
	wsHubAdapter := wsServices.NewHubAdapter(wsHub)

	// Initialize order and driver services
	merchantService := merchantServices.NewMerchantService(merchantRepo)
	orderService := orderServices.NewOrderService(orderRepo, gmapsClient)

	// Initialize route recalculation service
	routeRecalcService := driverServices.NewRouteRecalculationService(gmapsClient)

	// Create adapters for location service
	orderRepoAdapter := driverAdapters.NewOrderRepositoryAdapter(orderRepo)
	wsHubAdapterForLocation := driverAdapters.NewWebSocketHubAdapter(wsHub)

	// Initialize location service with route recalculation
	locationService := driverServices.NewLocationService(locationRepo, orderRepoAdapter, routeRecalcService, wsHubAdapterForLocation)

	// Initialize assignment service (core of the system)
	assignmentService := orderServices.NewAssignmentService(
		orderRepo,
		assignmentRepo,
		locationRepo,
		gmapsClient,
		notificationService,
		wsHubAdapter,
	)

	// Initialize handlers
	authHandler := authHandlers.NewAuthHandler(authService)
	userHandler := userHandlers.NewUserHandler(userService)
	documentHandler := documentHandlers.NewDocumentHandler(documentService)
	uploadHandler := documentHandlers.NewUploadHandler(r2Client)
	notificationHandler := notificationHandlers.NewNotificationHandler(notificationService)
	adminNotificationHandler := notificationHandlers.NewAdminNotificationHandler(notificationService)
	merchantHandler := merchantHandlers.NewMerchantHandler(merchantService)
	orderHandler := orderHandlers.NewOrderHandler(orderService, assignmentService)
	assignmentHandler := orderHandlers.NewAssignmentHandler(assignmentService)
	locationHandler := driverHandlers.NewLocationHandler(locationService)
	wsHandler := wsHandlers.NewWSHandler(wsHub)

	// Create mux and register all routes
	mux := http.NewServeMux()

	// Register system routes (health, swagger)
	router.RegisterSystemRoutes(mux)

	// Serve admin panel
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/admin.html")
	})

	// Serve test page
	mux.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/test.html")
	})

	// Register feature routes
	auth.RegisterRoutes(mux, authHandler)
	users.RegisterRoutes(mux, userHandler)
	documents.RegisterRoutes(mux, documentHandler, uploadHandler)
	notifications.RegisterRoutes(mux, notificationHandler, adminNotificationHandler)
	merchants.RegisterRoutes(mux, merchantHandler)
	orders.RegisterRoutes(mux, orderHandler)
	assignments := router.NewAssignmentRouter(assignmentHandler)
	assignments.RegisterRoutes(mux)
	drivers.RegisterRoutes(mux, locationHandler)
	websockets.RegisterRoutes(mux, wsHandler)

	// Apply global middleware
	handler := middleware.Logger(middleware.CORS(mux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with timeouts to prevent resource exhaustion
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	_ = server.ListenAndServe()
}

// updateSwaggerHost updates the Swagger documentation host dynamically
func updateSwaggerHost(baseURL string) {
	// Remove http:// or https:// prefix
	host := strings.TrimPrefix(baseURL, "http://")
	host = strings.TrimPrefix(host, "https://")

	// Remove trailing slash if present
	host = strings.TrimSuffix(host, "/")

	// Update the Swagger doc
	docs.SwaggerInfo.Host = host
}
