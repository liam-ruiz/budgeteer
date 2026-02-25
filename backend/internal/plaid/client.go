// internal/plaid/client.go
package plaid

import (
	plaid "github.com/plaid/plaid-go/v20/plaid"
)

func NewPlaidClient(clientID, secret, env string) *plaid.APIClient {
	cfg := plaid.NewConfiguration()
	switch env {
	case "sandbox":
		cfg.UseEnvironment(plaid.Sandbox)
	case "production":
		cfg.UseEnvironment(plaid.Production)
	}
	cfg.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	cfg.AddDefaultHeader("PLAID-SECRET", secret)
	return plaid.NewAPIClient(cfg)
}
