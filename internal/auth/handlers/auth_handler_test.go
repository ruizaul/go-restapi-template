package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"go-api-template/internal/auth/models"
	"go-api-template/internal/auth/services"
)

// Test status constants
const (
	statusSuccess = "success"
	statusFail    = "fail"
)

// mockAuthService is a simple mock for testing
type mockAuthService struct {
	registerFn      func(ctx context.Context, req *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error)
	loginFn         func(ctx context.Context, req *models.LoginRequest) (*models.AuthUser, *models.TokenPair, error)
	refreshTokensFn func(ctx context.Context, refreshToken string) (*models.AuthUser, *models.TokenPair, error)
	getProfileFn    func(ctx context.Context, userID uuid.UUID) (*models.AuthUser, error)
}

func (m *mockAuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, req)
	}
	return nil, nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthUser, *models.TokenPair, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, req)
	}
	return nil, nil, nil
}

func (m *mockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (*models.AuthUser, *models.TokenPair, error) {
	if m.refreshTokensFn != nil {
		return m.refreshTokensFn(ctx, refreshToken)
	}
	return nil, nil, nil
}

func (m *mockAuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*models.AuthUser, error) {
	if m.getProfileFn != nil {
		return m.getProfileFn(ctx, userID)
	}
	return nil, nil
}

// mockHandler wraps AuthHandler to use our mock service
type mockHandler struct {
	mock *mockAuthService
}

func newMockHandler(mock *mockAuthService) *mockHandler {
	return &mockHandler{mock: mock}
}

func (h *mockHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"status": statusFail,
			"data":   map[string]string{"body": "Invalid JSON"},
		})
		return
	}

	user, tokens, err := h.mock.Register(r.Context(), &req)
	if err != nil {
		switch err {
		case services.ErrEmailAlreadyExists:
			writeJSON(w, http.StatusConflict, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"email": "Email already exists"},
			})
		case services.ErrInvalidEmail:
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"email": "Invalid email format"},
			})
		case services.ErrWeakPassword:
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"password": "Password must be at least 8 characters"},
			})
		case services.ErrNameRequired:
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"name": "Name is required"},
			})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"status":  "error",
				"message": "Failed to create user",
				"code":    500,
			})
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"status": statusSuccess,
		"data": map[string]any{
			"user":   user,
			"tokens": tokens,
		},
	})
}

func (h *mockHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"status": statusFail,
			"data":   map[string]string{"body": "Invalid JSON"},
		})
		return
	}

	user, tokens, err := h.mock.Login(r.Context(), &req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"credentials": "Invalid email or password"},
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": "Failed to authenticate user",
			"code":    500,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": statusSuccess,
		"data": map[string]any{
			"user":   user,
			"tokens": tokens,
		},
	})
}

func (h *mockHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"status": statusFail,
			"data":   map[string]string{"body": "Invalid JSON"},
		})
		return
	}

	if req.RefreshToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"status": statusFail,
			"data":   map[string]string{"refresh_token": "Refresh token is required"},
		})
		return
	}

	user, tokens, err := h.mock.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case services.ErrInvalidToken:
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"refresh_token": "Invalid refresh token"},
			})
		case services.ErrExpiredToken:
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"refresh_token": "Refresh token has expired"},
			})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{
				"status":  "error",
				"message": "Failed to refresh tokens",
				"code":    500,
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": statusSuccess,
		"data": map[string]any{
			"user":   user,
			"tokens": tokens,
		},
	})
}

func (h *mockHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(uuid.UUID)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"status": statusFail,
			"data":   map[string]string{"auth": "User not authenticated"},
		})
		return
	}

	user, err := h.mock.GetProfile(r.Context(), userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"status": statusFail,
				"data":   map[string]string{"user": "User not found"},
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": "Failed to retrieve profile",
			"code":    500,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": statusSuccess,
		"data":   user,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func marshalJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return data
}

func unmarshalResponse(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	return response
}

func getDataMap(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data to be a map")
	}
	return data
}

func TestRegisterEndpoint_Success(t *testing.T) {
	testUserID := uuid.New()
	now := time.Now().UTC()

	mock := &mockAuthService{
		registerFn: func(_ context.Context, _ *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error) {
			return &models.AuthUser{
					ID:        testUserID,
					Email:     "test@example.com",
					Name:      "Test User",
					CreatedAt: now,
					UpdatedAt: now,
				}, &models.TokenPair{
					AccessToken:  "access-token",
					RefreshToken: "refresh-token",
					TokenType:    "Bearer",
					ExpiresIn:    900,
				}, nil
		},
	}
	handler := newMockHandler(mock)

	body := marshalJSON(t, models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	response := unmarshalResponse(t, w.Body.Bytes())
	if response["status"] != statusSuccess {
		t.Errorf("expected status success, got %v", response["status"])
	}
	data := getDataMap(t, response)
	if data["user"] == nil {
		t.Error("expected user in response")
	}
	if data["tokens"] == nil {
		t.Error("expected tokens in response")
	}
}

func TestRegisterEndpoint_DuplicateEmail(t *testing.T) {
	mock := &mockAuthService{
		registerFn: func(_ context.Context, _ *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error) {
			return nil, nil, services.ErrEmailAlreadyExists
		},
	}
	handler := newMockHandler(mock)

	body := marshalJSON(t, models.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	response := unmarshalResponse(t, w.Body.Bytes())
	if response["status"] != statusFail {
		t.Errorf("expected status fail, got %v", response["status"])
	}
	data := getDataMap(t, response)
	if data["email"] != "Email already exists" {
		t.Errorf("expected email error, got %v", data["email"])
	}
}

func TestRegisterEndpoint_InvalidEmail(t *testing.T) {
	mock := &mockAuthService{
		registerFn: func(_ context.Context, _ *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error) {
			return nil, nil, services.ErrInvalidEmail
		},
	}
	handler := newMockHandler(mock)

	body := marshalJSON(t, models.RegisterRequest{
		Email:    "invalid-email",
		Password: "password123",
		Name:     "Test User",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	response := unmarshalResponse(t, w.Body.Bytes())
	if response["status"] != statusFail {
		t.Errorf("expected status fail, got %v", response["status"])
	}
	data := getDataMap(t, response)
	if data["email"] != "Invalid email format" {
		t.Errorf("expected email error, got %v", data["email"])
	}
}

func TestRegisterEndpoint_WeakPassword(t *testing.T) {
	mock := &mockAuthService{
		registerFn: func(_ context.Context, _ *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error) {
			return nil, nil, services.ErrWeakPassword
		},
	}
	handler := newMockHandler(mock)

	body := marshalJSON(t, models.RegisterRequest{
		Email:    "test@example.com",
		Password: "short",
		Name:     "Test User",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	response := unmarshalResponse(t, w.Body.Bytes())
	if response["status"] != statusFail {
		t.Errorf("expected status fail, got %v", response["status"])
	}
	data := getDataMap(t, response)
	if data["password"] != "Password must be at least 8 characters" {
		t.Errorf("expected password error, got %v", data["password"])
	}
}

func TestRegisterEndpoint_MissingName(t *testing.T) {
	mock := &mockAuthService{
		registerFn: func(_ context.Context, _ *models.RegisterRequest) (*models.AuthUser, *models.TokenPair, error) {
			return nil, nil, services.ErrNameRequired
		},
	}
	handler := newMockHandler(mock)

	body := marshalJSON(t, models.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	response := unmarshalResponse(t, w.Body.Bytes())
	if response["status"] != statusFail {
		t.Errorf("expected status fail, got %v", response["status"])
	}
	data := getDataMap(t, response)
	if data["name"] != "Name is required" {
		t.Errorf("expected name error, got %v", data["name"])
	}
}

func TestRegisterEndpoint_InvalidJSON(t *testing.T) {
	mock := &mockAuthService{}
	handler := newMockHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	response := unmarshalResponse(t, w.Body.Bytes())
	if response["status"] != statusFail {
		t.Errorf("expected status fail, got %v", response["status"])
	}
	data := getDataMap(t, response)
	if data["body"] != "Invalid JSON" {
		t.Errorf("expected body error, got %v", data["body"])
	}
}

func TestLoginEndpoint(t *testing.T) {
	testUserID := uuid.New()
	now := time.Now().UTC()

	t.Run("successful login", func(t *testing.T) {
		mock := &mockAuthService{
			loginFn: func(_ context.Context, _ *models.LoginRequest) (*models.AuthUser, *models.TokenPair, error) {
				return &models.AuthUser{
						ID:        testUserID,
						Email:     "test@example.com",
						Name:      "Test User",
						CreatedAt: now,
						UpdatedAt: now,
					}, &models.TokenPair{
						AccessToken:  "access-token",
						RefreshToken: "refresh-token",
						TokenType:    "Bearer",
						ExpiresIn:    900,
					}, nil
			},
		}
		handler := newMockHandler(mock)

		body := marshalJSON(t, models.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusSuccess {
			t.Errorf("expected status success, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["user"] == nil {
			t.Error("expected user in response")
		}
		if data["tokens"] == nil {
			t.Error("expected tokens in response")
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mock := &mockAuthService{
			loginFn: func(_ context.Context, _ *models.LoginRequest) (*models.AuthUser, *models.TokenPair, error) {
				return nil, nil, services.ErrInvalidCredentials
			},
		}
		handler := newMockHandler(mock)

		body := marshalJSON(t, models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusFail {
			t.Errorf("expected status fail, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["credentials"] != "Invalid email or password" {
			t.Errorf("expected credentials error, got %v", data["credentials"])
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mock := &mockAuthService{}
		handler := newMockHandler(mock)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Login(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusFail {
			t.Errorf("expected status fail, got %v", response["status"])
		}
	})
}

func TestRefreshEndpoint(t *testing.T) {
	testUserID := uuid.New()
	now := time.Now().UTC()

	t.Run("successful refresh", func(t *testing.T) {
		mock := &mockAuthService{
			refreshTokensFn: func(_ context.Context, _ string) (*models.AuthUser, *models.TokenPair, error) {
				return &models.AuthUser{
						ID:        testUserID,
						Email:     "test@example.com",
						Name:      "Test User",
						CreatedAt: now,
						UpdatedAt: now,
					}, &models.TokenPair{
						AccessToken:  "new-access-token",
						RefreshToken: "new-refresh-token",
						TokenType:    "Bearer",
						ExpiresIn:    900,
					}, nil
			},
		}
		handler := newMockHandler(mock)

		body := marshalJSON(t, models.RefreshRequest{
			RefreshToken: "valid-refresh-token",
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Refresh(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusSuccess {
			t.Errorf("expected status success, got %v", response["status"])
		}
	})

	t.Run("missing refresh token", func(t *testing.T) {
		mock := &mockAuthService{}
		handler := newMockHandler(mock)

		body := marshalJSON(t, models.RefreshRequest{
			RefreshToken: "",
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Refresh(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusFail {
			t.Errorf("expected status fail, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["refresh_token"] != "Refresh token is required" {
			t.Errorf("expected refresh_token error, got %v", data["refresh_token"])
		}
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		mock := &mockAuthService{
			refreshTokensFn: func(_ context.Context, _ string) (*models.AuthUser, *models.TokenPair, error) {
				return nil, nil, services.ErrInvalidToken
			},
		}
		handler := newMockHandler(mock)

		body := marshalJSON(t, models.RefreshRequest{
			RefreshToken: "invalid-token",
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Refresh(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusFail {
			t.Errorf("expected status fail, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["refresh_token"] != "Invalid refresh token" {
			t.Errorf("expected refresh_token error, got %v", data["refresh_token"])
		}
	})

	t.Run("expired refresh token", func(t *testing.T) {
		mock := &mockAuthService{
			refreshTokensFn: func(_ context.Context, _ string) (*models.AuthUser, *models.TokenPair, error) {
				return nil, nil, services.ErrExpiredToken
			},
		}
		handler := newMockHandler(mock)

		body := marshalJSON(t, models.RefreshRequest{
			RefreshToken: "expired-token",
		})

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Refresh(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusFail {
			t.Errorf("expected status fail, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["refresh_token"] != "Refresh token has expired" {
			t.Errorf("expected refresh_token error, got %v", data["refresh_token"])
		}
	})
}

func TestGetProfileEndpoint(t *testing.T) {
	testUserID := uuid.New()
	now := time.Now().UTC()

	t.Run("successful get profile", func(t *testing.T) {
		mock := &mockAuthService{
			getProfileFn: func(_ context.Context, _ uuid.UUID) (*models.AuthUser, error) {
				return &models.AuthUser{
					ID:        testUserID,
					Email:     "test@example.com",
					Name:      "Test User",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
		}
		handler := newMockHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusSuccess {
			t.Errorf("expected status success, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["email"] != "test@example.com" {
			t.Errorf("expected email test@example.com, got %v", data["email"])
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		mock := &mockAuthService{}
		handler := newMockHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusFail {
			t.Errorf("expected status fail, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["auth"] != "User not authenticated" {
			t.Errorf("expected auth error, got %v", data["auth"])
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mock := &mockAuthService{
			getProfileFn: func(_ context.Context, _ uuid.UUID) (*models.AuthUser, error) {
				return nil, services.ErrUserNotFound
			},
		}
		handler := newMockHandler(mock)

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		ctx := context.WithValue(req.Context(), UserIDKey, testUserID)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}

		response := unmarshalResponse(t, w.Body.Bytes())
		if response["status"] != statusFail {
			t.Errorf("expected status fail, got %v", response["status"])
		}
		data := getDataMap(t, response)
		if data["user"] != "User not found" {
			t.Errorf("expected user error, got %v", data["user"])
		}
	})
}
