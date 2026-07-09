package test

import (
	"context"
	"testing"

	"github.com/1solomonwakhungu/go-production-api-starter/internal/model"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository/mock"
)

// TestMockUserRepository_Create verifies the mock repository correctly
// records calls and invokes the configured function.
func TestMockUserRepository_Create(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}

	called := false
	mockRepo.CreateFunc = func(ctx context.Context, user *model.User) error {
		called = true
		if user.Email == "" {
			t.Error("expected non-empty email")
		}
		return nil
	}

	user := &model.User{
		ID:    "test-id",
		Email: "test@example.com",
		Name:  "Test User",
	}

	err := mockRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Error("expected CreateFunc to be called")
	}

	if len(mockRepo.CreateCalls) != 1 {
		t.Fatalf("expected 1 recorded call, got %d", len(mockRepo.CreateCalls))
	}

	if mockRepo.CreateCalls[0].User.ID != "test-id" {
		t.Errorf("expected call to record user ID 'test-id', got %s", mockRepo.CreateCalls[0].User.ID)
	}
}

func TestMockUserRepository_GetByID(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}

	mockRepo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
		if id == "found" {
			return &model.User{ID: id, Email: "found@example.com", Name: "Found"}, nil
		}
		return nil, repository.ErrNotFound
	}

	t.Run("found", func(t *testing.T) {
		user, err := mockRepo.GetByID(context.Background(), "found")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user == nil {
			t.Fatal("expected user to be non-nil")
		}
		if user.Email != "found@example.com" {
			t.Errorf("expected email 'found@example.com', got %s", user.Email)
		}
	})

	t.Run("not found", func(t *testing.T) {
		user, err := mockRepo.GetByID(context.Background(), "missing")
		if err != repository.ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
		if user != nil {
			t.Error("expected user to be nil for not found")
		}
	})

	if len(mockRepo.GetByIDCalls) != 2 {
		t.Errorf("expected 2 recorded calls, got %d", len(mockRepo.GetByIDCalls))
	}
}

func TestMockUserRepository_GetAll(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}

	mockRepo.GetAllFunc = func(ctx context.Context, limit, offset int) ([]*model.User, error) {
		return []*model.User{
			{ID: "1", Email: "a@example.com", Name: "A"},
			{ID: "2", Email: "b@example.com", Name: "B"},
		}, nil
	}

	users, err := mockRepo.GetAll(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}

	call := mockRepo.GetAllCalls[0]
	if call.Limit != 10 {
		t.Errorf("expected limit 10, got %d", call.Limit)
	}
	if call.Offset != 0 {
		t.Errorf("expected offset 0, got %d", call.Offset)
	}
}

func TestMockUserRepository_Update(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}

	mockRepo.UpdateFunc = func(ctx context.Context, user *model.User) error {
		return nil
	}

	user := &model.User{
		ID:    "test-id",
		Email: "updated@example.com",
		Name:  "Updated Name",
	}

	err := mockRepo.Update(context.Background(), user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mockRepo.UpdateCalls) != 1 {
		t.Errorf("expected 1 recorded call, got %d", len(mockRepo.UpdateCalls))
	}
	if mockRepo.UpdateCalls[0].User.Email != "updated@example.com" {
		t.Errorf("expected email 'updated@example.com', got %s", mockRepo.UpdateCalls[0].User.Email)
	}
}

func TestMockUserRepository_Delete(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}

	mockRepo.DeleteFunc = func(ctx context.Context, id string) error {
		return nil
	}

	err := mockRepo.Delete(context.Background(), "test-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(mockRepo.DeleteCalls) != 1 {
		t.Errorf("expected 1 recorded call, got %d", len(mockRepo.DeleteCalls))
	}
	if mockRepo.DeleteCalls[0].ID != "test-id" {
		t.Errorf("expected ID 'test-id', got %s", mockRepo.DeleteCalls[0].ID)
	}
}

// TestMockUserRepository_DefaultBehavior verifies that when no Func is set,
// the mock returns zero values without panicking.
func TestMockUserRepository_DefaultBehavior(t *testing.T) {
	mockRepo := &mock.MockUserRepository{}

	user, err := mockRepo.GetByID(context.Background(), "any")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user != nil {
		t.Error("expected nil user for default behavior")
	}

	users, err := mockRepo.GetAll(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if users != nil {
		t.Error("expected nil users for default behavior")
	}
}
