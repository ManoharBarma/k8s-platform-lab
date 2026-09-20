package handler

import (
	"context"

	"github.com/shopspring/decimal"

	ledgerv1 "banking-platform/proto/gen/ledger/v1"
	"banking-platform/services/ledger-service/internal/domain"
	"banking-platform/services/ledger-service/internal/service"
)

// GRPC implements the ledger.v1.LedgerService gRPC server.
type GRPC struct {
	ledgerv1.UnimplementedLedgerServiceServer
	svc *service.Service
}

// NewGRPC builds the gRPC handler.
func NewGRPC(svc *service.Service) *GRPC { return &GRPC{svc: svc} }

// OpenAccount creates a ledger account.
func (g *GRPC) OpenAccount(ctx context.Context, req *ledgerv1.OpenAccountRequest) (*ledgerv1.OpenAccountResponse, error) {
	if err := g.svc.OpenAccount(ctx, req.GetAccountId(), req.GetCurrency()); err != nil {
		return nil, err
	}
	return &ledgerv1.OpenAccountResponse{AccountId: req.GetAccountId()}, nil
}

// PostTransaction applies a balanced transaction.
func (g *GRPC) PostTransaction(ctx context.Context, req *ledgerv1.PostTransactionRequest) (*ledgerv1.PostTransactionResponse, error) {
	entries := make([]domain.Entry, 0, len(req.GetEntries()))
	for _, e := range req.GetEntries() {
		amt, err := decimal.NewFromString(e.GetAmount())
		if err != nil {
			return nil, err
		}
		entries = append(entries, domain.Entry{
			AccountID: e.GetAccountId(),
			Direction: directionFromProto(e.GetDirection()),
			Amount:    amt,
		})
	}
	id, dup, err := g.svc.Post(ctx, domain.Transaction{
		ID:             req.GetTransactionId(),
		IdempotencyKey: req.GetIdempotencyKey(),
		Currency:       req.GetCurrency(),
		Reference:      req.GetReference(),
		Entries:        entries,
	})
	if err != nil {
		return nil, err
	}
	return &ledgerv1.PostTransactionResponse{TransactionId: id, Duplicate: dup}, nil
}

// GetBalance returns an account balance.
func (g *GRPC) GetBalance(ctx context.Context, req *ledgerv1.GetBalanceRequest) (*ledgerv1.GetBalanceResponse, error) {
	bal, currency, err := g.svc.Balance(ctx, req.GetAccountId())
	if err != nil {
		return nil, err
	}
	return &ledgerv1.GetBalanceResponse{
		AccountId: req.GetAccountId(),
		Balance:   bal.StringFixed(2),
		Currency:  currency,
	}, nil
}

func directionFromProto(d ledgerv1.Direction) domain.Direction {
	if d == ledgerv1.Direction_DIRECTION_DEBIT {
		return domain.Debit
	}
	return domain.Credit
}
