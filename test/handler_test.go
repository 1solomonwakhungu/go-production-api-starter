package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/1solomonwakhungu/go-production-api-starter/internal/handler"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/model"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository/mock"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/service"
)

// setupTestHandler creates a handler wired to a mock repository and returns both.
func setupTestHandler() (*handler.Handler, *mock.MockUserRepository) {
	mockRepo := &mock.MockUserRepository{}
	userService := service.NewUserService(mockRepo)
	h := handler.New(userService)
	return h, mockRepo
}

func TestHealthEndpoint(t *testing.T) {
	h, _ := setupTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp handler.Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(repo *mock.MockUserRepository)
		wantStatus int
		wantErr    bool
	}{
		{
			name: "valid user",
			body: `{"email":"user@example.com","name":"John Doe"}`,
			setupMock: func(repo *mock.MockUserRepository) {
				repo.CreateFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid email",
			body:       `{"email":"not-an-email","name":"John Doe"}`,
			setupMock:  func(repo *mock.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "name too short",
			body:       `{"email":"user@example.com","name":"J"}`,
			setupMock:  func(repo *mock.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			setupMock:  func(repo *mock.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "repository error",
			body: `{"email":"user@example.com","name":"John Doe"}`,
			setupMock: func(repo *mock.MockUserRepository) {
				repo.CreateFunc = func(ctx context.Context, user *model.User) error {
					return repository.ErrNotFound
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo := setupTestHandler()
			tt.setupMock(mockRepo)

			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.CreateUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d; body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			var resp handler.Response
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if tt.wantErr && resp.Success {
				t.Error("expected success to be false for error case")
			}
			if !tt.wantErr && !resp.Success {
				t.Error("expected success to be true")
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		setupMock  func(repo *mock.MockUserRepository)
		wantStatus int
	}{
		{
			name:   "user found",
			userID: "550e8400-e29b-41d4-a716-446655440000",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
					return &model.User{
						ID:    id,
						Email: "user@example.com",
						Name:  "John Doe",
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "empty id",
			userID:     "",
			setupMock:  func(repo *mock.MockUserRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo := setupTestHandler()
			tt.setupMock(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/users/{id}", nil)
			req.SetPathValue("id", tt.userID)
			w := httptest.NewRecorder()

			h.GetUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d; body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		setupMock  func(repo *mock.MockUserRepository)
		wantStatus int
	}{
		{
			name:   "successful delete",
			userID: "550e8400-e29b-41d4-a716-446655440000",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.DeleteFunc = func(ctx context.Context, id string) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.DeleteFunc = func(ctx context.Context, id string) error {
					return repository.ErrNotFound
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo := setupTestHandler()
			tt.setupMock(mockRepo)

			req := httptest.NewRequest(http.MethodDelete, "/users/{id}", nil)
			req.SetPathValue("id", tt.userID)
			w := httptest.NewRecorder()

			h.DeleteUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d; body: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
