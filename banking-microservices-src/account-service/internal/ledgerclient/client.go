// Package ledgerclient adapts the ledger gRPC client to the account service's
// LedgerClient interface.
package ledgerclient

import (
	"context"

	"google.golang.org/grpc"

	ledgerv1 "banking-platform/proto/gen/ledger/v1"
)

// Client wraps the generated ledger gRPC client.
type Client struct {
	c ledgerv1.LedgerServiceClient
}

// New builds a ledger client over an existing connection.
func New(conn *grpc.ClientConn) *Client {
	return &Client{c: ledgerv1.NewLedgerServiceClient(conn)}
}

// OpenAccount opens a ledger account.
func (c *Client) OpenAccount(ctx context.Context, accountID, currency string) error {
	_, err := c.c.OpenAccount(ctx, &ledgerv1.OpenAccountRequest{AccountId: accountID, Currency: currency})
	return err
}

// Balance returns an account's balance and currency.
func (c *Client) Balance(ctx context.Context, accountID string) (string, string, error) {
	resp, err := c.c.GetBalance(ctx, &ledgerv1.GetBalanceRequest{AccountId: accountID})
	if err != nil {
		return "", "", err
	}
	return resp.GetBalance(), resp.GetCurrency(), nil
}
