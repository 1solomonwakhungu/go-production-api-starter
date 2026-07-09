package test

import (
	"context"
	"testing"

	"github.com/1solomonwakhungu/go-production-api-starter/internal/model"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository/mock"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/service"
)

func setupTestService() (*service.UserService, *mock.MockUserRepository) {
	mockRepo := &mock.MockUserRepository{}
	return service.NewUserService(mockRepo), mockRepo
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     model.CreateUserInput
		setupMock func(repo *mock.MockUserRepository)
		wantErr   bool
	}{
		{
			name: "valid user",
			input: model.CreateUserInput{
				Email: "User@Example.com",
				Name:  "John Doe",
			},
			setupMock: func(repo *mock.MockUserRepository) {
				repo.CreateFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			input: model.CreateUserInput{
				Email: "not-an-email",
				Name:  "John Doe",
			},
			setupMock: func(repo *mock.MockUserRepository) {},
			wantErr:   true,
		},
		{
			name: "name too short",
			input: model.CreateUserInput{
				Email: "user@example.com",
				Name:  "J",
			},
			setupMock: func(repo *mock.MockUserRepository) {},
			wantErr:   true,
		},
		{
			name: "repository error",
			input: model.CreateUserInput{
				Email: "user@example.com",
				Name:  "John Doe",
			},
			setupMock: func(repo *mock.MockUserRepository) {
				repo.CreateFunc = func(ctx context.Context, user *model.User) error {
					return repository.ErrNotFound
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupTestService()
			tt.setupMock(mockRepo)

			user, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tt.wantErr && user == nil {
				t.Error("expected user to be non-nil")
			}
			if !tt.wantErr && user != nil && user.Email != "user@example.com" {
				t.Errorf("expected email to be lowercased, got %s", user.Email)
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(repo *mock.MockUserRepository)
		wantErr   bool
		wantNotFound bool
	}{
		{
			name: "user found",
			id:   "test-id",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
					return &model.User{ID: id, Email: "user@example.com", Name: "John Doe"}, nil
				}
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "nonexistent",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantErr: true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupTestService()
			tt.setupMock(mockRepo)

			user, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if tt.wantNotFound && err != nil && err != service.ErrUserNotFound {
				t.Errorf("expected ErrUserNotFound, got %v", err)
			}
			if !tt.wantErr && user == nil {
				t.Error("expected user to be non-nil")
			}
		})
	}
}

func TestUserService_Update(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		input     model.UpdateUserInput
		setupMock func(repo *mock.MockUserRepository)
		wantErr   bool
	}{
		{
			name: "update email only",
			id:   "test-id",
			input: model.UpdateUserInput{
				Email: strPtr("new@example.com"),
			},
			setupMock: func(repo *mock.MockUserRepository) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
					return &model.User{ID: id, Email: "old@example.com", Name: "John Doe"}, nil
				}
				repo.UpdateFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "user not found for update",
			id:   "nonexistent",
			input: model.UpdateUserInput{
				Name: strPtr("New Name"),
			},
			setupMock: func(repo *mock.MockUserRepository) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
					return nil, repository.ErrNotFound
				}
			},
			wantErr: true,
		},
		{
			name: "invalid email on update",
			id:   "test-id",
			input: model.UpdateUserInput{
				Email: strPtr("not-an-email"),
			},
			setupMock: func(repo *mock.MockUserRepository) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
					return &model.User{ID: id, Email: "old@example.com", Name: "John Doe"}, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupTestService()
			tt.setupMock(mockRepo)

			user, err := svc.Update(context.Background(), tt.id, tt.input)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !tt.wantErr && user == nil {
				t.Error("expected user to be non-nil")
			}
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(repo *mock.MockUserRepository)
		wantErr   bool
	}{
		{
			name: "successful delete",
			id:   "test-id",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.DeleteFunc = func(ctx context.Context, id string) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "nonexistent",
			setupMock: func(repo *mock.MockUserRepository) {
				repo.DeleteFunc = func(ctx context.Context, id string) error {
					return repository.ErrNotFound
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupTestService()
			tt.setupMock(mockRepo)

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestUserService_GetAll(t *testing.T) {
	t.Run("default pagination", func(t *testing.T) {
		svc, mockRepo := setupTestService()

		mockRepo.GetAllFunc = func(ctx context.Context, limit, offset int) ([]*model.User, error) {
			if limit != 10 {
				t.Errorf("expected limit 10, got %d", limit)
			}
			if offset != 0 {
				t.Errorf("expected offset 0, got %d", offset)
			}
			return []*model.User{
				{ID: "1", Email: "a@example.com", Name: "A"},
				{ID: "2", Email: "b@example.com", Name: "B"},
			}, nil
		}

		users, err := svc.GetAll(context.Background(), 0, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(users) != 2 {
			t.Errorf("expected 2 users, got %d", len(users))
		}
	})

	t.Run("custom pagination", func(t *testing.T) {
		svc, mockRepo := setupTestService()

		mockRepo.GetAllFunc = func(ctx context.Context, limit, offset int) ([]*model.User, error) {
			if limit != 20 {
				t.Errorf("expected limit 20, got %d", limit)
			}
			if offset != 40 {
				t.Errorf("expected offset 40, got %d", offset)
			}
			return []*model.User{}, nil
		}

		users, err := svc.GetAll(context.Background(), 3, 20)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(users) != 0 {
			t.Errorf("expected 0 users, got %d", len(users))
		}
	})
}

func strPtr(s string) *string {
	return &s
}
