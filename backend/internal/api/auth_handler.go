package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/liam-ruiz/budget/internal/api/types"
	"github.com/liam-ruiz/budget/internal/auth"
	"github.com/liam-ruiz/budget/internal/dependencies"
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

	writeJSON(w, http.StatusOK, types.AuthResponse{
		Token: token,
		User: types.UserResponse{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}

func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
    // If the code gets here, the Middleware already verified the token!
    // We just pull the userID out of the context.
    userIDString := r.Context().Value("userID").(string)
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		log.Printf("Error parsing userID: %v\n", err)
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