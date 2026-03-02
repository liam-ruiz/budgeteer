package dependencies

import (
	"github.com/liam-ruiz/budget/internal/bank_accounts"
	"github.com/liam-ruiz/budget/internal/plaid_items"
	"github.com/liam-ruiz/budget/internal/budgets"
	"github.com/liam-ruiz/budget/internal/config"
	"github.com/liam-ruiz/budget/internal/plaid"

	"github.com/liam-ruiz/budget/internal/transactions"
	"github.com/liam-ruiz/budget/internal/users"

)

type Container struct {
	UserSvc        *users.Service
	AccountSvc     *bank_accounts.Service
	BudgetSvc      *budgets.Service
	TransactionSvc *transactions.Service
	PlaidAPISvc    *plaid.Service
	Cfg            *config.Config
	PlaidItemSvc   *plaid_items.Service
}

func NewContainer(
	userSvc *users.Service,
	accountSvc *bank_accounts.Service,
	budgetSvc *budgets.Service,
	transactionSvc *transactions.Service,
	plaidAPISvc *plaid.Service,
	cfg *config.Config,
	plaidItemSvc *plaid_items.Service,
) *Container {
	return &Container{
		UserSvc:        userSvc,
		AccountSvc:     accountSvc,
		BudgetSvc:      budgetSvc,
		TransactionSvc: transactionSvc,
		PlaidAPISvc:    plaidAPISvc,
		Cfg:            cfg,
		PlaidItemSvc:   plaidItemSvc,
	}
}
