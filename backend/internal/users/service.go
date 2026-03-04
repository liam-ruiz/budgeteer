package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken         = errors.New("email already in use")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// Service handles user business logic.
type Service struct {
	repo Repository
}

// NewService creates a new user service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Register creates a new user with a hashed password.
func (s *Service) Register(ctx context.Context, email string, password string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hashing password: %w", err)
	}
	user, err := s.repo.Create(ctx, email, string(hash))
	if err != nil {
		return User{}, fmt.Errorf("creating user: %w", err)
	}
	return user, nil
}

// Authenticate validates credentials and returns the user if valid.
func (s *Service) Authenticate(ctx context.Context, email string, password string) (User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return User{}, fmt.Errorf("getting user: %w", err)
	}
	return user, nil
}
