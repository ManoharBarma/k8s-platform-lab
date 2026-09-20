package handler

import (
	"context"

	authv1 "banking-platform/proto/gen/auth/v1"
	"banking-platform/services/auth-service/internal/service"
)

// GRPC implements the auth.v1.AuthService gRPC server, letting other services
// validate access tokens through auth-service.
type GRPC struct {
	authv1.UnimplementedAuthServiceServer
	svc *service.Service
}

// NewGRPC builds the gRPC handler.
func NewGRPC(svc *service.Service) *GRPC { return &GRPC{svc: svc} }

// ValidateToken verifies an access token and returns the identity it carries.
func (g *GRPC) ValidateToken(_ context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := g.svc.ValidateAccess(req.GetAccessToken())
	if err != nil {
		return &authv1.ValidateTokenResponse{Valid: false}, nil
	}
	return &authv1.ValidateTokenResponse{
		Valid:   true,
		Subject: claims.Subject,
		Email:   claims.Email,
		Roles:   claims.Roles,
	}, nil
}
