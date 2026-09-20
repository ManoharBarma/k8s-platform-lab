// Package ledgerclient adapts the ledger gRPC client for payment-service.
package ledgerclient

import (
	"context"

	"google.golang.org/grpc"

	ledgerv1 "banking-platform/proto/gen/ledger/v1"
)

// Entry mirrors a ledger entry in plain strings.
type Entry struct {
	AccountID string
	Direction string
	Amount    string
}

// Client wraps the generated ledger gRPC client.
type Client struct {
	c ledgerv1.LedgerServiceClient
}

// New builds a ledger client over an existing connection.
func New(conn *grpc.ClientConn) *Client {
	return &Client{c: ledgerv1.NewLedgerServiceClient(conn)}
}

// OpenAccount opens a ledger account (idempotent).
func (c *Client) OpenAccount(ctx context.Context, accountID, currency string) error {
	_, err := c.c.OpenAccount(ctx, &ledgerv1.OpenAccountRequest{AccountId: accountID, Currency: currency})
	return err
}

// Post applies a balanced transaction.
func (c *Client) Post(ctx context.Context, txID, idemKey, currency, reference string, entries []Entry) (string, bool, error) {
	pe := make([]*ledgerv1.Entry, 0, len(entries))
	for _, e := range entries {
		dir := ledgerv1.Direction_DIRECTION_CREDIT
		if e.Direction == "DEBIT" {
			dir = ledgerv1.Direction_DIRECTION_DEBIT
		}
		pe = append(pe, &ledgerv1.Entry{AccountId: e.AccountID, Direction: dir, Amount: e.Amount})
	}
	resp, err := c.c.PostTransaction(ctx, &ledgerv1.PostTransactionRequest{
		TransactionId: txID, IdempotencyKey: idemKey, Currency: currency, Reference: reference, Entries: pe,
	})
	if err != nil {
		return "", false, err
	}
	return resp.GetTransactionId(), resp.GetDuplicate(), nil
}

// Balance returns an account's balance and currency.
func (c *Client) Balance(ctx context.Context, accountID string) (string, string, error) {
	resp, err := c.c.GetBalance(ctx, &ledgerv1.GetBalanceRequest{AccountId: accountID})
	if err != nil {
		return "", "", err
	}
	return resp.GetBalance(), resp.GetCurrency(), nil
}
