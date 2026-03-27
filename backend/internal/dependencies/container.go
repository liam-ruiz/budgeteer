package dependencies

import (
	"github.com/liam-ruiz/budgeteer/internal/bank_accounts"
	"github.com/liam-ruiz/budgeteer/internal/budgets"
	"github.com/liam-ruiz/budgeteer/internal/config"
	"github.com/liam-ruiz/budgeteer/internal/plaid"
	"github.com/liam-ruiz/budgeteer/internal/plaid_items"

	"github.com/liam-ruiz/budgeteer/internal/transactions"
	"github.com/liam-ruiz/budgeteer/internal/users"
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
