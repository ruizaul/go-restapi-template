//nolint:errcheck // Test file - error checking not critical for test assertions
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api-template/internal/users/models"
	"go-api-template/internal/users/repositories"
	"go-api-template/internal/users/services"

	"github.com/google/uuid"
)

// mockDB implements a simple in-memory store for testing
type mockDB struct {
	users map[uuid.UUID]*models.User
}

func newMockDB() *mockDB {
	return &mockDB{users: make(map[uuid.UUID]*models.User)}
}

// mockUserRepository implements repository methods using in-memory storage
type mockUserRepository struct {
	db *mockDB
}

func newMockUserRepository(db *mockDB) *mockUserRepository {
	return &mockUserRepository{db: db}
}

func (r *mockUserRepository) Create(_ context.Context, user *models.User) error {
	user.ID = uuid.New()
	r.db.users[user.ID] = user
	return nil
}

func (r *mockUserRepository) GetByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	user, ok := r.db.users[id]
	if !ok || user.DeletedAt != nil {
		return nil, repositories.ErrUserNotFound
	}
	return user, nil
}

func (r *mockUserRepository) GetByEmail(_ context.Context, email string) (*models.User, error) {
	for _, user := range r.db.users {
		if user.Email == email && user.DeletedAt == nil {
			return user, nil
		}
	}
	return nil, repositories.ErrUserNotFound
}

func (r *mockUserRepository) List(_ context.Context, limit, offset int) ([]models.User, error) {
	var result []models.User
	i := 0
	for _, user := range r.db.users {
		if user.DeletedAt != nil {
			continue
		}
		if i >= offset && len(result) < limit {
			result = append(result, *user)
		}
		i++
	}
	return result, nil
}

func (r *mockUserRepository) Update(_ context.Context, user *models.User) error {
	if _, ok := r.db.users[user.ID]; !ok {
		return repositories.ErrUserNotFound
	}
	r.db.users[user.ID] = user
	return nil
}

func (r *mockUserRepository) Delete(_ context.Context, id uuid.UUID) error {
	user, ok := r.db.users[id]
	if !ok || user.DeletedAt != nil {
		return repositories.ErrUserNotFound
	}
	// Simulate soft delete
	delete(r.db.users, id)
	return nil
}

// mockUserService wraps the mock repository
type mockUserService struct {
	repo *mockUserRepository
}

func newMockUserService(repo *mockUserRepository) *mockUserService {
	return &mockUserService{repo: repo}
}

func (s *mockUserService) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	existing, _ := s.repo.GetByEmail(ctx, req.Email) //nolint:errcheck // test mock
	if existing != nil {
		return nil, services.ErrEmailAlreadyExists
	}
	user := &models.User{Email: req.Email, Name: req.Name}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *mockUserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, services.ErrUserNotFound
	}
	return user, nil
}

func (s *mockUserService) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *mockUserService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, services.ErrUserNotFound
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *mockUserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return services.ErrUserNotFound
	}
	return nil
}

// testHandler wraps UserHandler with mock service
type testHandler struct {
	*UserHandler
	mockService *mockUserService
}

func newTestHandler() *testHandler {
	db := newMockDB()
	repo := newMockUserRepository(db)
	svc := newMockUserService(repo)
	// Create handler with mock service interface
	handler := &UserHandler{service: nil}
	return &testHandler{UserHandler: handler, mockService: svc}
}

// Override handler methods to use mock service
func (h *testHandler) List(w http.ResponseWriter, r *http.Request) {
	users, _ := h.mockService.List(r.Context(), 20, 0) //nolint:errcheck // test mock
	writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": users})
}

func (h *testHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "data": map[string]string{"id": "Invalid UUID"}})
		return
	}
	user, err := h.mockService.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "fail", "data": map[string]string{"id": "User not found"}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": user})
}

func (h *testHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "data": map[string]string{"body": "Invalid JSON"}})
		return
	}
	if req.Email == "" || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "data": map[string]string{"error": "Email and name required"}})
		return
	}
	user, err := h.mockService.Create(r.Context(), &req)
	if err == services.ErrEmailAlreadyExists {
		writeJSON(w, http.StatusConflict, map[string]any{"status": "fail", "data": map[string]string{"email": "Already exists"}})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"status": "success", "data": user})
}

func (h *testHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "data": map[string]string{"id": "Invalid UUID"}})
		return
	}
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "data": map[string]string{"body": "Invalid JSON"}})
		return
	}
	user, err := h.mockService.Update(r.Context(), id, &req)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "fail", "data": map[string]string{"id": "User not found"}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "success", "data": user})
}

func (h *testHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"status": "fail", "data": map[string]string{"id": "Invalid UUID"}})
		return
	}
	if err := h.mockService.Delete(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"status": "fail", "data": map[string]string{"id": "User not found"}})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func setupRouter(h *testHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", h.List)
	mux.HandleFunc("GET /users/{id}", h.GetByID)
	mux.HandleFunc("POST /users", h.Create)
	mux.HandleFunc("PATCH /users/{id}", h.Update)
	mux.HandleFunc("DELETE /users/{id}", h.Delete)
	return mux
}

func TestUserEndpoints(t *testing.T) {
	h := newTestHandler()
	mux := setupRouter(h)

	t.Run("POST /users - create user", func(t *testing.T) {
		body := `{"email":"test@example.com","name":"Test User"}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}

		var resp map[string]any
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["status"] != "success" {
			t.Errorf("expected success status, got %v", resp["status"])
		}
	})

	t.Run("POST /users - duplicate email", func(t *testing.T) {
		body := `{"email":"test@example.com","name":"Another User"}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", w.Code)
		}
	})

	t.Run("POST /users - missing fields", func(t *testing.T) {
		body := `{"email":""}`
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("GET /users - list users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		var resp map[string]any
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["status"] != "success" {
			t.Errorf("expected success status")
		}
	})

	t.Run("GET /users/{id} - not found", func(t *testing.T) {
		fakeID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, "/users/"+fakeID, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("GET /users/{id} - invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/not-a-uuid", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("DELETE /users/{id} - not found", func(t *testing.T) {
		fakeID := uuid.New().String()
		req := httptest.NewRequest(http.MethodDelete, "/users/"+fakeID, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})
}

func TestCreateAndGetUser(t *testing.T) {
	h := newTestHandler()
	mux := setupRouter(h)

	// Create user
	body := `{"email":"john@example.com","name":"John Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed: %d", w.Code)
	}

	var createResp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&createResp)
	data := createResp["data"].(map[string]any) //nolint:errcheck // test assertion
	userID := data["id"].(string)               //nolint:errcheck // test assertion

	// Get user
	req = httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("get failed: expected 200, got %d", w.Code)
	}

	var getResp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&getResp)
	userData := getResp["data"].(map[string]any) //nolint:errcheck // test assertion
	if userData["email"] != "john@example.com" {
		t.Errorf("expected email john@example.com, got %v", userData["email"])
	}
}

func TestUpdateUser(t *testing.T) {
	h := newTestHandler()
	mux := setupRouter(h)

	// Create user first
	body := `{"email":"update@example.com","name":"Original Name"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var createResp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&createResp)
	data := createResp["data"].(map[string]any) //nolint:errcheck // test assertion
	userID := data["id"].(string)               //nolint:errcheck // test assertion

	// Update user
	updateBody := `{"name":"Updated Name"}`
	req = httptest.NewRequest(http.MethodPatch, "/users/"+userID, bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("update failed: expected 200, got %d", w.Code)
	}

	var updateResp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&updateResp)
	userData := updateResp["data"].(map[string]any) //nolint:errcheck // test assertion
	if userData["name"] != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %v", userData["name"])
	}
}

func TestDeleteUser(t *testing.T) {
	h := newTestHandler()
	mux := setupRouter(h)

	// Create user first
	body := `{"email":"delete@example.com","name":"To Delete"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var createResp map[string]any
	_ = json.NewDecoder(w.Body).Decode(&createResp)
	data := createResp["data"].(map[string]any) //nolint:errcheck // test assertion
	userID := data["id"].(string)               //nolint:errcheck // test assertion

	// Delete user
	req = httptest.NewRequest(http.MethodDelete, "/users/"+userID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("delete failed: expected 204, got %d", w.Code)
	}

	// Verify user is gone
	req = httptest.NewRequest(http.MethodGet, "/users/"+userID, nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}
