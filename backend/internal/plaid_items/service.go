package plaid_items

import (
	"context"

	"github.com/liam-ruiz/budgeteer/internal/db/sqlcdb"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetPlaidItemByPlaidItemID(ctx context.Context, plaidItemID string) (sqlcdb.PlaidItem, error) {
	return s.repo.GetPlaidItemByPlaidItemID(ctx, plaidItemID)
}

func (s *Service) CreatePlaidItem(ctx context.Context, params sqlcdb.CreatePlaidItemParams) (sqlcdb.PlaidItem, error) {
	return s.repo.CreatePlaidItem(ctx, params)
}

func (s *Service) GetCursor(ctx context.Context, plaidItemID string) (string, error) {
	return s.repo.GetCursor(ctx, plaidItemID)
}
