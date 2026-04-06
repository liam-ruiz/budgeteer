package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/liam-ruiz/budgeteer/internal/api/types"
	"github.com/liam-ruiz/budgeteer/internal/auth"
	"github.com/liam-ruiz/budgeteer/internal/dependencies"
	plaidlib "github.com/plaid/plaid-go/v20/plaid"
)

type PlaidHandler struct {
	container *dependencies.Container
}

func NewPlaidHandler(container *dependencies.Container) *PlaidHandler {
	return &PlaidHandler{
		container: container,
	}
}

var (
	webhookTypeTransactions      = "TRANSACTIONS"
	webhookCodeItemSynced        = "SYNC_UPDATES_AVAILABLE"
	webhookCodeItemHistorySynced = "HISTORICAL_UPDATE"
	webhookCodeInitialUpdate     = "INITIAL_UPDATE"
)

func (h *PlaidHandler) ExchangePlaidPublicToken(w http.ResponseWriter, r *http.Request) {

	log.Printf("[ExchangePlaidPublicToken] %s %s", r.Method, r.URL.Path)
	userID, _ := auth.GetUserID(r.Context()) // Assuming middleware handled this

	var req types.ExchangeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[ExchangePlaidPublicToken] error decoding request body: %v", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// exchange public token for access token
	exchangeReq := plaidlib.NewItemPublicTokenExchangeRequest(req.PublicToken)
	resp, err := h.container.PlaidAPISvc.ExchangePublicToken(r.Context(), exchangeReq)
	if err != nil {
		log.Printf("[ExchangePlaidPublicToken] plaid exchange failed: %v", err)
		writeError(w, http.StatusBadGateway, "plaid exchange failed")
		return
	}

	accessToken := resp.GetAccessToken()
	plaidItemID := resp.GetItemId()

	// fetch metadata for this Item via Plaid /accounts/get
	// best to get the account details immediately so the user sees them.
	accountsReq := plaidlib.NewAccountsGetRequest(accessToken)
	accountsResp, err := h.container.PlaidAPISvc.GetAccounts(r.Context(), accountsReq)
	if err != nil {
		log.Printf("[ExchangePlaidPublicToken] failed to fetch account details: %v", err)
		writeError(w, http.StatusBadGateway, "failed to fetch account details")
		return
	}

	// persist the Item and Accounts
	// We do this in the service layer to handle the transaction/upsert logic
	err = h.container.AccountSvc.LinkNewItem(r.Context(), userID, plaidItemID, accessToken, &accountsResp)
	if err != nil {
		log.Printf("[ExchangePlaidPublicToken] failed to save linked accounts: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save linked accounts")
		return
	}

	// trigger background transaction sync (Production-lite approach)
	// empty string for cursor means we want to sync all available transactions
	err = h.container.AccountSvc.SyncTransactions(context.Background(), plaidItemID, "")
	if err != nil {
		log.Printf("[ExchangePlaidPublicToken] failed to sync transactions: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to sync transactions")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "syncing"})
}

// CreateLinkToken generates a Plaid Link token for the authenticated user.
func (h *PlaidHandler) CreateLinkToken(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CreateLinkToken] %s %s", r.Method, r.URL.Path)
	userID, err := auth.GetUserID(r.Context())
	if err != nil {
		log.Printf("Error getting user ID: %v\n", err)
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	products := []plaidlib.Products{plaidlib.PRODUCTS_TRANSACTIONS}
	countryCodes := []plaidlib.CountryCode{plaidlib.COUNTRYCODE_US}

	user := plaidlib.LinkTokenCreateRequestUser{
		ClientUserId: userID.String(),
	}

	linkReq := plaidlib.NewLinkTokenCreateRequest(
		"Budget",
		"en",
		countryCodes,
		user,
	)
	linkReq.SetProducts(products)
	linkReq.SetWebhook(h.container.Cfg.WebhookURL + "/api/plaid/webhook")

	resp, err := h.container.PlaidAPISvc.CreateLinkToken(r.Context(), linkReq)
	if err != nil {
		log.Printf("Error creating link token: %v\n", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create link token: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, types.CreateLinkTokenResponse{
		LinkToken: resp.GetLinkToken(),
	})
}

func (h *PlaidHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	signedJWT := r.Header.Get("Plaid-Verification")
	if signedJWT == "" {
		log.Printf("Error: missing Plaid-Verification header")
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// parse the JWT without verification first to get the 'kid'
	token, _, err := new(jwt.Parser).ParseUnverified(signedJWT, jwt.MapClaims{})
	if err != nil {
		log.Printf("Error parsing unverified JWT: %v", err)
		writeError(w, http.StatusBadRequest, "invalid token format")
		return
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		log.Printf("Error: JWT header missing 'kid'")
		writeError(w, http.StatusBadRequest, "invalid token header")
		return
	}

	// get the public key
	resp, err := h.container.PlaidAPISvc.WebhookPublicKeyGet(r.Context(), kid)
	if err != nil {
		log.Printf("Error getting public key: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	key, ok := resp.GetKeyOk()
	if !ok {
		log.Printf("Error: request is not from plaid\n")
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// read the body for verification
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading body: %v\n", err)
		writeError(w, http.StatusInternalServerError, "could not read body")
		return
	}
	// put the body back so we can JSON decode it later
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// verify the request
	err = verifyPlaidWebhook(r.Context(), signedJWT, *key, bodyBytes)
	if err != nil {
		log.Printf("Error verifying request: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	var webhook types.PlaidWebhook
	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		log.Printf("Error decoding request body: %v\n", err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	switch webhook.WebhookType {
	case webhookTypeTransactions:
		switch webhook.WebhookCode {
		case webhookCodeItemSynced:
			log.Printf("Data is ready for Item: %s", webhook.PlaidItemID)

			go func(itemID string) {
				// check for cursor in db, might be empty
				cursor, err := h.container.PlaidItemSvc.GetCursor(context.Background(), itemID)
				if err != nil {
					log.Printf("Error getting cursor for %s: %v", itemID, err)
					return
				}

				err = h.container.AccountSvc.SyncTransactions(context.Background(), itemID, cursor)
				if err != nil {
					log.Printf("Async sync failed for %s: %v", itemID, err)
				}

			}(webhook.PlaidItemID)

		// ignore initial update and historical sync as these are for backwards compatibility
		case webhookCodeInitialUpdate, webhookCodeItemHistorySynced:
			log.Printf("Item updated for Item, ignoring: %s", webhook.PlaidItemID)

		default:
			log.Printf("Unknown webhook code: %s\n", webhook.WebhookCode)
			writeError(w, http.StatusBadRequest, "unknown webhook code")
			return
		}
	default:
		log.Printf("Unknown webhook type: %s\n", webhook.WebhookType)
		writeError(w, http.StatusBadRequest, "unknown webhook type")
		return
	}
}

func verifyPlaidWebhook(ctx context.Context, signedJWT string, key plaidlib.JWKPublicKey, bodyBytes []byte) error {
	// convert Plaid's JWK response to an ECDSA Public Key
	// plaid returns x and y coordinates in base64url
	pubKey, err := parsePlaidKey(key)
	if err != nil {
		return err
	}

	parser := jwt.NewParser(jwt.WithLeeway(5 * time.Minute))

	// fully verify the JWT signature
	parsedToken, err := parser.Parse(signedJWT, func(t *jwt.Token) (any, error) {
		return pubKey, nil
	})
	if err != nil || !parsedToken.Valid {
		if err != nil {
			log.Printf("Error parsing JWT: %v\n", err)
		}
		return fmt.Errorf("invalid jwt signature")
	}

	// verify the body hash
	claims := parsedToken.Claims.(jwt.MapClaims)
	claimedHash := claims["request_body_sha256"].(string)

	actualHash := sha256.Sum256(bodyBytes)
	actualHashStr := hex.EncodeToString(actualHash[:])

	if claimedHash != actualHashStr {
		return fmt.Errorf("body hash mismatch")
	}

	return nil
}

func parsePlaidKey(jwk plaidlib.JWKPublicKey) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, err
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}
