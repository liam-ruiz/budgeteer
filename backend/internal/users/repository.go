package users

import (
	"context"

	"github.com/google/uuid"
	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
)

// Repository defines the interface for user data access.
type Repository interface {
	Create(ctx context.Context, email string, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
}

type repository struct {
	q *sqlcdb.Queries
}

// NewRepository creates a new user repository backed by sqlc queries.
func NewRepository(q *sqlcdb.Queries) Repository {
	return &repository{q: q}
}

func (r *repository) Create(ctx context.Context, email string, passwordHash string) (User, error) {
	row, err := r.q.CreateUser(ctx, sqlcdb.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return User{}, err
	}
	return toUser(row), nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}
	return toUser(row), nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return User{}, err
	}
	return toUser(row), nil
}

func toUser(row sqlcdb.User) User {
	return User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt.Time,
	}
}
