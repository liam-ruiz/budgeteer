package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/liam-ruiz/budgeteer/internal/api/types"
	"github.com/liam-ruiz/budgeteer/internal/auth"
	"github.com/liam-ruiz/budgeteer/internal/dependencies"
)

// AuthHandler defines the handlers for authentication routes.
type AuthHandler struct {
	container *dependencies.Container
}

func NewAuthHandler(container *dependencies.Container) *AuthHandler {
	return &AuthHandler{
		container: container,
	}
}

// Register creates a new user account.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Register] %s %s", r.Method, r.URL.Path)
	var req types.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.container.UserSvc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("Error registering user: %v\n", err)
		writeError(w, http.StatusConflict, "email already in use")
		return
	}

	token, err := auth.GenerateJWT(user.ID, h.container.Cfg.JWTSecret)
	if err != nil {
		log.Printf("Error generating JWT: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusCreated, types.AuthResponse{
		Token: token,
		User: types.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}

// Login authenticates a user and returns a JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("[Login] %s %s", r.Method, r.URL.Path)
	var req types.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.container.UserSvc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		log.Printf("Error authenticating user: %v\n", err)
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := auth.GenerateJWT(user.ID, h.container.Cfg.JWTSecret)
	if err != nil {
		log.Printf("Error generating JWT: %v\n", err)
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	go func() {
		if err := h.container.AccountSvc.SyncLinkedAccounts(context.Background(), user.ID); err != nil {
			log.Printf("[Login] failed to sync linked accounts for user %s: %v", user.ID, err)
		}
	}()

	writeJSON(w, http.StatusOK, types.AuthResponse{
		Token: token,
		User: types.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}

func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.container.UserSvc.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, types.UserResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}
