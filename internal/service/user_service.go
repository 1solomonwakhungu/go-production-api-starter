package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/1solomonwakhungu/go-production-api-starter/internal/model"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/repository"
	"github.com/google/uuid"
)

// UserService encapsulates business logic for user operations.
// It depends on the UserRepository interface, not a concrete implementation.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new UserService with the given repository.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Create validates and persists a new user.
func (s *UserService) Create(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("service: validate create input: %w", err)
	}

	now := time.Now().UTC()
	user := &model.User{
		ID:        uuid.NewString(),
		Email:     strings.ToLower(input.Email),
		Name:      input.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("service: create user: %w", err)
	}
	return user, nil
}

// GetByID retrieves a user by their ID.
func (s *UserService) GetByID(ctx context.Context, id string) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("service: get user: %w", err)
	}
	return user, nil
}

// GetAll retrieves a paginated list of users.
func (s *UserService) GetAll(ctx context.Context, page, pageSize int) ([]*model.User, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	users, err := s.repo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("service: get all users: %w", err)
	}
	return users, nil
}

// Update modifies an existing user's fields.
func (s *UserService) Update(ctx context.Context, id string, input model.UpdateUserInput) (*model.User, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("service: find user for update: %w", err)
	}

	if input.Email != nil {
		if !isValidEmail(*input.Email) {
			return nil, model.ErrInvalidEmail
		}
		existing.Email = strings.ToLower(*input.Email)
	}
	if input.Name != nil {
		if len(*input.Name) < 2 {
			return nil, model.ErrNameTooShort
		}
		existing.Name = *input.Name
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("service: update user: %w", err)
	}
	return existing, nil
}

// Delete removes a user by ID.
func (s *UserService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("service: delete user: %w", err)
	}
	return nil
}

var (
	ErrUserNotFound = errors.New("user not found")
)

var emailRegexStr = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

func isValidEmail(email string) bool {
	// Simple email check without importing regexp here to avoid circular deps
	// Uses basic validation
	at := strings.Index(email, "@")
	if at < 1 || at == len(email)-1 {
		return false
	}
	dot := strings.LastIndex(email[at:], ".")
	if dot < 2 {
		return false
	}
	return true
}
