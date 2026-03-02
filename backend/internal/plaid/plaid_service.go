package plaid

import (
	"context"

	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

type Service struct {
	client *plaidlib.APIClient
}

func NewService(client *plaidlib.APIClient) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) CreateLinkToken(ctx context.Context, req *plaidlib.LinkTokenCreateRequest) (plaidlib.LinkTokenCreateResponse, error) {
	resp, _, err := s.client.PlaidApi.
		LinkTokenCreate(ctx).
		LinkTokenCreateRequest(*req).
		Execute()
	if err != nil {
		return plaidlib.LinkTokenCreateResponse{}, err
	}
	return resp, nil
}

func (s *Service) ExchangePublicToken(ctx context.Context, req *plaidlib.ItemPublicTokenExchangeRequest) (plaidlib.ItemPublicTokenExchangeResponse, error) {
	resp, _, err := s.client.PlaidApi.
		ItemPublicTokenExchange(ctx).
		ItemPublicTokenExchangeRequest(*req).
		Execute()
	if err != nil {
		return plaidlib.ItemPublicTokenExchangeResponse{}, err
	}
	return resp, nil
}

func (s *Service) GetAccountInfo(ctx context.Context, req *plaidlib.AccountsGetRequest) (plaidlib.AccountsGetResponse, error) {
	resp, _, err := s.client.PlaidApi.
		AccountsGet(ctx).
		AccountsGetRequest(*req).
		Execute()
	if err != nil {
		return plaidlib.AccountsGetResponse{}, err
	}
	return resp, nil
}

func (s *Service) SyncTransactions(ctx context.Context, req *plaidlib.TransactionsSyncRequest) (plaidlib.TransactionsSyncResponse, error) {
	resp, _, err := s.client.PlaidApi.
		TransactionsSync(ctx).
		TransactionsSyncRequest(*req).
		Execute()
	if err != nil {
		return plaidlib.TransactionsSyncResponse{}, err
	}
	return resp, nil
}

func (s *Service) GetItem(ctx context.Context, req *plaidlib.ItemGetRequest) (plaidlib.ItemGetResponse, error) {
	resp, _, err := s.client.PlaidApi.
		ItemGet(ctx).
		ItemGetRequest(*req).
		Execute()
	if err != nil {
		return plaidlib.ItemGetResponse{}, err
	}
	return resp, nil
}

func (s *Service) WebhookPublicKeyGet(ctx context.Context, kid string) (plaidlib.WebhookVerificationKeyGetResponse, error) {
	resp, _, err := s.client.PlaidApi.
		WebhookVerificationKeyGet(ctx).
		WebhookVerificationKeyGetRequest(plaidlib.WebhookVerificationKeyGetRequest{
			KeyId: kid,
		}).
		Execute()
	if err != nil {
		return plaidlib.WebhookVerificationKeyGetResponse{}, err
	}
	return resp, nil
}