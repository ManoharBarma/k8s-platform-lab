package handler

import (
	"context"

	accountv1 "banking-platform/proto/gen/account/v1"
	"banking-platform/services/account-service/internal/service"
)

// GRPC implements the account.v1.AccountService gRPC server so other services
// can validate accounts before moving money.
type GRPC struct {
	accountv1.UnimplementedAccountServiceServer
	svc *service.Service
}

// NewGRPC builds the gRPC handler.
func NewGRPC(svc *service.Service) *GRPC { return &GRPC{svc: svc} }

// GetAccount returns account details.
func (g *GRPC) GetAccount(ctx context.Context, req *accountv1.GetAccountRequest) (*accountv1.GetAccountResponse, error) {
	acc, err := g.svc.Get(ctx, req.GetAccountId())
	if err != nil {
		return nil, err
	}
	return &accountv1.GetAccountResponse{
		AccountId:     acc.ID,
		AccountNumber: acc.AccountNumber,
		CustomerId:    acc.CustomerID,
		Currency:      acc.Currency,
		Status:        string(acc.Status),
	}, nil
}
