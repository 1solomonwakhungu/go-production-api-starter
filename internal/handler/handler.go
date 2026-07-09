package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/1solomonwakhungu/go-production-api-starter/internal/model"
	"github.com/1solomonwakhungu/go-production-api-starter/internal/service"
)

// Handler holds dependencies for HTTP request handlers.
type Handler struct {
	userService *service.UserService
}

// New creates a new Handler with the given user service.
func New(userService *service.UserService) *Handler {
	return &Handler{userService: userService}
}

// Health responds with a simple health check payload.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// CreateUser handles POST /users.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input model.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.Create(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, user)
}

// GetUser handles GET /users/{id}.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing user id")
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

// ListUsers handles GET /users with optional pagination.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	users, err := h.userService.GetAll(r.Context(), page, pageSize)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	WriteJSONWithMeta(w, http.StatusOK, users, &Meta{
		Page:     page,
		PageSize: pageSize,
		Total:    len(users),
	})
}

// UpdateUser handles PUT /users/{id}.
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing user id")
		return
	}

	var input model.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.userService.Update(r.Context(), id, input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}.
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		WriteError(w, http.StatusBadRequest, "missing user id")
		return
	}

	if err := h.userService.Delete(r.Context(), id); err != nil {
		handleServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{
		"status": "deleted",
	})
}

// handleServiceError maps service-layer errors to appropriate HTTP status codes.
// It uses errors.Is for sentinel error matching and falls back to error message
// inspection for validation errors that may be wrapped.
func handleServiceError(w http.ResponseWriter, err error) {
	// Check sentinel errors first using errors.Is for proper unwrapping
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		WriteError(w, http.StatusNotFound, err.Error())

	case errors.Is(err, model.ErrInvalidEmail):
		WriteError(w, http.StatusBadRequest, err.Error())

	case errors.Is(err, model.ErrNameTooShort):
		WriteError(w, http.StatusBadRequest, err.Error())

	default:
		// For wrapped errors, check the error message for known patterns.
		// This handles validation errors that may be wrapped with fmt.Errorf.
		msg := err.Error()
		if strings.Contains(msg, "invalid email") {
			WriteError(w, http.StatusBadRequest, msg)
		} else if strings.Contains(msg, "name must be") {
			WriteError(w, http.StatusBadRequest, msg)
		} else if strings.Contains(msg, "validate") {
			WriteError(w, http.StatusBadRequest, msg)
		} else {
			WriteError(w, http.StatusInternalServerError, "internal server error")
		}
	}
}
