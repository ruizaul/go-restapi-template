package handlers

import (
	"encoding/json"
	"net/http"

	"tacoshare-delivery-api/internal/auth/models"
	"tacoshare-delivery-api/internal/auth/services"
	"tacoshare-delivery-api/pkg/httpx"
	"tacoshare-delivery-api/pkg/middleware"

	"github.com/google/uuid"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate user with email and password. Returns access token (valid for 1 hour) and refresh token (valid for 30 days). Access token should be included in Authorization header as "Bearer {token}" for protected endpoints. The response includes complete user information (ID, name, email, phone, role) along with the authentication tokens.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.LoginRequest		true	"User credentials - email and password"
//	@Success		200		{object}	models.LoginResponse	"Login successful - returns access token (JWT, expires in 1h), refresh token (JWT, expires in 30d), and complete user profile"
//	@Failure		400		{object}	httpx.JSendFail			"Invalid request body (malformed JSON) or missing required fields (email/password)"
//	@Failure		401		{object}	httpx.JSendFail			"Invalid credentials - incorrect email or password combination"
//	@Failure		500		{object}	httpx.JSendError		"Internal server error - database connection failed or unexpected error during authentication"
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	// Validate required fields asda
	if req.Email == "" || req.Password == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "El correo electrónico y la contraseña son requeridos",
		})
		return
	}

	// Extract device info and IP address
	deviceInfo := r.Header.Get("User-Agent")
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	authResp, err := h.authService.Login(&req, deviceInfo, ipAddress)
	if err != nil {
		errMsg := err.Error()
		switch errMsg {
		case "user not found":
			httpx.RespondFail(w, http.StatusUnauthorized, map[string]any{
				"error": "No existe una cuenta con este correo electrónico",
			})
			return
		case "invalid password":
			httpx.RespondFail(w, http.StatusUnauthorized, map[string]any{
				"error": "Contraseña incorrecta",
			})
			return
		default:
			httpx.RespondError(w, http.StatusInternalServerError, "Error al iniciar sesión")
			return
		}
	}

	httpx.RespondSuccess(w, http.StatusOK, authResp)
}

// RefreshToken godoc
//
//	@Summary		Refresh access token
//	@Description	Generate new access and refresh tokens using a valid refresh token. Use this endpoint when the access token expires (after 1 hour). The old refresh token will be invalidated and replaced with a new one for security. Both tokens are JWT format. The response includes the new tokens and updated user information.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.RefreshRequest	true	"Refresh token from previous login or refresh"
//	@Success		200		{object}	models.RefreshResponse	"Tokens refreshed successfully - returns new access token (expires in 1h), new refresh token (expires in 30d), and user profile"
//	@Failure		400		{object}	httpx.JSendFail			"Invalid request body (malformed JSON) or missing refresh_token field"
//	@Failure		401		{object}	httpx.JSendError		"Invalid refresh token (malformed JWT, expired, or revoked) or associated user account no longer exists"
//	@Failure		500		{object}	httpx.JSendError		"Internal server error - database failure or token generation error"
//	@Router			/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	// Validate required field
	if req.RefreshToken == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "El token de actualización es requerido",
		})
		return
	}

	// Extract device info and IP address
	deviceInfo := r.Header.Get("User-Agent")
	ipAddress := r.Header.Get("X-Forwarded-For")
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	// Extract device_id from header (X-Device-ID) or generate from User-Agent
	deviceID := r.Header.Get("X-Device-ID")
	if deviceID == "" {
		// Fallback: use User-Agent as device identifier (not ideal but better than nothing)
		deviceID = deviceInfo
	}

	authResp, err := h.authService.RefreshToken(req.RefreshToken, deviceInfo, ipAddress, deviceID)
	if err != nil {
		if err.Error() == "user not found" {
			httpx.RespondError(w, http.StatusUnauthorized, "Usuario no encontrado")
			return
		}
		httpx.RespondError(w, http.StatusUnauthorized, "Token de actualización inválido o expirado")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, authResp)
}

// Register godoc
//
//	@Summary		User registration with OTP verification
//	@Description	Two-step registration process with phone verification. **Step 1 (Send OTP)**: Send only phone to receive 6-digit OTP via SMS (expires in 10 min). **Step 2 (Complete Registration)**: After verifying OTP with /auth/verify-otp, send complete registration data including personal info, verified phone, and credentials to create account. Password requirements: 6-72 characters. Age requirement: 18+ years old (calculated from birth_date). Returns access token (1h validity) and refresh token (30d validity) upon successful registration.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.RegisterRequest				true	"Step 1: phone only | Step 2: all fields after OTP verification"
//	@Success		200		{object}	models.RegisterResponse				"OTP sent successfully to phone - valid for 10 minutes"
//	@Success		201		{object}	models.CompleteRegistrationResponse	"User account created successfully with authentication tokens"
//	@Failure		400		{object}	httpx.JSendFailPhoneInvalid			"Phone number format invalid - must be E.164 format (e.g., +525512345678)"
//	@Failure		400		{object}	httpx.JSendFailEmailExists			"Email address already registered in the system"
//	@Failure		400		{object}	httpx.JSendFailPhoneExists			"Phone number already registered to another account"
//	@Failure		400		{object}	httpx.JSendFailOTPInvalid			"OTP code is incorrect or malformed (must be 6 digits)"
//	@Failure		400		{object}	httpx.JSendFailPhoneNotVerified		"Phone number not verified - must call /auth/verify-otp first"
//	@Failure		400		{object}	httpx.JSendFailAgeRestriction		"User age verification failed - must be 18 years or older"
//	@Failure		500		{object}	httpx.JSendError					"Internal server error - SMS service failure or database error"
//	@Router			/auth/register [post]
//
//nolint:gocyclo // Complex registration flow with multiple validation steps
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	// Validate required phone field
	if req.Phone == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"phone": "El número de teléfono es requerido",
		})
		return
	}

	result, err := h.authService.Register(&req)
	if err != nil {
		// Handle specific errors
		switch err.Error() {
		case "invalid phone format":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"phone": "Formato de teléfono inválido (use formato E.164: +525512345678)",
			})
		case "phone number already registered":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"phone": "El número de teléfono ya está registrado",
			})
		case "email already registered":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"email": "El correo electrónico ya está registrado",
			})
		case "invalid email format":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"email": "Formato de correo electrónico inválido",
			})
		case "password must be between 6 and 72 characters":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"password": "La contraseña debe tener entre 6 y 72 caracteres",
			})
		case "invalid OTP format":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"otp": "Formato de OTP inválido (debe tener 6 dígitos)",
			})
		case "phone not verified - please verify OTP first":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"phone": "Teléfono no verificado - por favor verifique el OTP primero",
			})
		case "phone number not found - please request OTP first":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"phone": "Número de teléfono no encontrado - por favor solicite un OTP primero",
			})
		case "user must be at least 18 years old":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"birth_date": "El usuario debe tener al menos 18 años",
			})
		case "invalid birth_date format (use YYYY-MM-DD)":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"birth_date": "Formato de fecha inválido (use AAAA-MM-DD)",
			})
		case "first_name, last_name, and birth_date are required":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"error": "El nombre, apellido y fecha de nacimiento son requeridos",
			})
		default:
			httpx.RespondError(w, http.StatusInternalServerError, "Error al procesar el registro")
		}
		return
	}

	// Check result type to determine response
	switch v := result.(type) {
	case *models.OTPSentResponse:
		// Mode 1: OTP sent
		httpx.RespondSuccess(w, http.StatusOK, v)
	case *models.AuthResponse:
		// Mode 2: Complete registration
		httpx.RespondSuccess(w, http.StatusCreated, v)
	default:
		httpx.RespondError(w, http.StatusInternalServerError, "Tipo de respuesta inesperado")
	}
}

// VerifyOTP godoc
//
//	@Summary		Verify OTP code
//	@Description	Verify the 6-digit OTP code sent via SMS to the user's phone number. This endpoint validates the OTP and marks the phone number as verified in the system, which is required before completing registration. OTP codes expire after 10 minutes. After successful verification, proceed to POST /auth/register with complete registration data including the verified phone number.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.VerifyOTPRequest			true	"Phone and 6-digit OTP code"
//	@Success		200		{object}	models.VerifyOTPResponseWrapper	"Phone number verified successfully - can now complete registration"
//	@Failure		400		{object}	httpx.JSendFailPhoneInvalid		"Phone number format invalid - must be E.164 format"
//	@Failure		400		{object}	httpx.JSendFailOTPInvalid		"OTP code is incorrect or format is invalid (must be exactly 6 digits)"
//	@Failure		400		{object}	httpx.JSendFailOTPExpired		"OTP code has expired (10 minute validity) - request a new OTP via POST /auth/register"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error - database error during OTP verification"
//	@Router			/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	// Validate required fields
	if req.Phone == "" || req.OTP == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "El teléfono y el OTP son requeridos",
		})
		return
	}

	result, err := h.authService.VerifyOTP(&req)
	if err != nil {
		// Handle specific errors
		switch err.Error() {
		case "invalid phone format":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"phone": "Formato de teléfono inválido (use formato E.164: +525512345678)",
			})
		case "invalid OTP format":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"otp": "Formato de OTP inválido (debe tener 6 dígitos)",
			})
		case "phone number not found":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"phone": "Número de teléfono no encontrado",
			})
		case "no OTP found for this phone number":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"otp": "No se encontró OTP para este número de teléfono - por favor solicite un OTP primero",
			})
		case "OTP has expired":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"otp": "El OTP ha expirado - por favor solicite uno nuevo",
			})
		case "invalid OTP code":
			httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
				"otp": "Código OTP inválido",
			})
		default:
			httpx.RespondError(w, http.StatusInternalServerError, "Error al verificar el OTP")
		}
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, result)
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Revoke a specific refresh token to logout from a single device. The refresh token will be marked as revoked and cannot be used again. Access tokens remain valid until they expire (15 minutes), but without a valid refresh token, the user cannot obtain new access tokens and will need to login again.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.LogoutRequest			true	"Refresh token to revoke"
//	@Success		200		{object}	models.LogoutResponseWrapper	"Logout successful - refresh token revoked"
//	@Failure		400		{object}	httpx.JSendFail					"Invalid request body or missing refresh_token field"
//	@Failure		401		{object}	httpx.JSendError				"Unauthorized - user not authenticated"
//	@Failure		404		{object}	httpx.JSendFail					"Refresh token not found in database"
//	@Failure		500		{object}	httpx.JSendError				"Internal server error - failed to revoke token"
//	@Security		BearerAuth
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req models.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "Cuerpo de la solicitud inválido",
		})
		return
	}

	// Validate required field
	if req.RefreshToken == "" {
		httpx.RespondFail(w, http.StatusBadRequest, map[string]any{
			"error": "El token de actualización es requerido",
		})
		return
	}

	// Revoke the refresh token
	if err := h.authService.Logout(req.RefreshToken); err != nil {
		if err.Error() == "refresh token not found" {
			httpx.RespondFail(w, http.StatusNotFound, map[string]any{
				"error": "Token de actualización no encontrado",
			})
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "Error al cerrar sesión")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, models.LogoutResponse{
		Message: "Sesión cerrada exitosamente",
	})
}

// LogoutAllDevices godoc
//
//	@Summary		Logout from all devices
//	@Description	Revoke all refresh tokens for the authenticated user, effectively logging out from all devices. This is useful when a user suspects their account has been compromised or wants to terminate all active sessions. All existing refresh tokens will be marked as revoked.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.LogoutResponseWrapper	"All sessions logged out successfully"
//	@Failure		401	{object}	httpx.JSendError				"Unauthorized - user not authenticated"
//	@Failure		500	{object}	httpx.JSendError				"Internal server error - failed to revoke tokens"
//	@Security		BearerAuth
//	@Router			/auth/logout-all [post]
func (h *AuthHandler) LogoutAllDevices(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by RequireAuth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Revoke all user tokens
	if err := h.authService.LogoutAllDevices(userID); err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al cerrar todas las sesiones")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, models.LogoutResponse{
		Message: "Todas las sesiones cerradas exitosamente",
	})
}

// GetActiveSessions godoc
//
//	@Summary		Get active sessions
//	@Description	Retrieve all active sessions (non-revoked, non-expired refresh tokens) for the authenticated user. Returns session details including device information, IP address, creation time, and expiration time. Useful for users to see where they are currently logged in.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	models.ActiveSessionsResponseWrapper	"Active sessions retrieved successfully"
//	@Failure		401	{object}	httpx.JSendError						"Unauthorized - user not authenticated"
//	@Failure		500	{object}	httpx.JSendError						"Internal server error - failed to retrieve sessions"
//	@Security		BearerAuth
//	@Router			/auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by RequireAuth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		httpx.RespondError(w, http.StatusUnauthorized, "ID de usuario inválido")
		return
	}

	// Get active sessions
	sessions, err := h.authService.GetActiveSessions(userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "Error al obtener sesiones activas")
		return
	}

	httpx.RespondSuccess(w, http.StatusOK, models.ActiveSessionsResponse{
		Sessions: sessions,
	})
}
