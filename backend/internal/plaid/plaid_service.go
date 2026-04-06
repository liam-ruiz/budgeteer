package plaid

import (
	"context"
	"errors"
	"fmt"
	"log"

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
		// cast the error to get the actual Plaid response body
		if nerr, ok := err.(plaidlib.GenericOpenAPIError); ok {
			log.Printf("Plaid API Error Body: %s", string(nerr.Body()))
			err = fmt.Errorf("plaid api error: %w", errors.New(string(nerr.Body())))
			return plaidlib.LinkTokenCreateResponse{}, err
		}
		log.Printf("Error creating link token: %v\n", err)
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
		// cast the error to get the actual Plaid response body
		if nerr, ok := err.(plaidlib.GenericOpenAPIError); ok {
			log.Printf("Plaid API Error Body: %s", string(nerr.Body()))
			err = fmt.Errorf("plaid api error: %w", errors.New(string(nerr.Body())))
			return plaidlib.ItemPublicTokenExchangeResponse{}, err
		}
		log.Printf("Error exchanging public token: %v\n", err)
		return plaidlib.ItemPublicTokenExchangeResponse{}, err
	}
	return resp, nil
}

func (s *Service) GetAccounts(ctx context.Context, req *plaidlib.AccountsGetRequest) (plaidlib.AccountsGetResponse, error) {
	resp, _, err := s.client.PlaidApi.
		AccountsGet(ctx).
		AccountsGetRequest(*req).
		Execute()
	if err != nil {
		// cast the error to get the actual Plaid response body
		if nerr, ok := err.(plaidlib.GenericOpenAPIError); ok {
			log.Printf("Plaid API Error Body: %s", string(nerr.Body()))
			err = fmt.Errorf("plaid api error: %w", errors.New(string(nerr.Body())))
			return plaidlib.AccountsGetResponse{}, err
		}
		log.Printf("Error getting accounts: %v\n", err)
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
		// cast the error to get the actual Plaid response body
		if nerr, ok := err.(plaidlib.GenericOpenAPIError); ok {
			log.Printf("Plaid API Error Body: %s", string(nerr.Body()))
			err = fmt.Errorf("plaid api error: %w", errors.New(string(nerr.Body())))
			return plaidlib.TransactionsSyncResponse{}, err
		}
		log.Printf("Error syncing transactions: %v\n", err)
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
		// cast the error to get the actual Plaid response body
		if nerr, ok := err.(plaidlib.GenericOpenAPIError); ok {
			log.Printf("Plaid API Error Body: %s", string(nerr.Body()))
			err = fmt.Errorf("plaid api error: %w", errors.New(string(nerr.Body())))
			return plaidlib.ItemGetResponse{}, err
		}
		log.Printf("Error getting item: %v\n", err)
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
		// cast the error to get the actual Plaid response body
		if nerr, ok := err.(plaidlib.GenericOpenAPIError); ok {
			log.Printf("Plaid API Error Body: %s", string(nerr.Body()))
			err = fmt.Errorf("plaid api error: %w", errors.New(string(nerr.Body())))
			return plaidlib.WebhookVerificationKeyGetResponse{}, err
		}
		log.Printf("Error getting webhook public key: %v\n", err)
		return plaidlib.WebhookVerificationKeyGetResponse{}, err
	}

	return resp, nil
}
