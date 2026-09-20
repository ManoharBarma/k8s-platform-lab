// Package handler contains authz-service transport layers (HTTP + gRPC).
package handler

import (
	"context"

	authzv1 "banking-platform/proto/gen/authz/v1"
	"banking-platform/services/authz-service/internal/service"
)

// GRPC implements the authz.v1.AuthzService gRPC server.
type GRPC struct {
	authzv1.UnimplementedAuthzServiceServer
	svc *service.Service
}

// NewGRPC builds the gRPC handler.
func NewGRPC(svc *service.Service) *GRPC { return &GRPC{svc: svc} }

// Check reports whether the subject holds the requested permission.
func (g *GRPC) Check(ctx context.Context, req *authzv1.CheckRequest) (*authzv1.CheckResponse, error) {
	allowed, err := g.svc.Check(ctx, req.GetSubject(), req.GetPermission())
	if err != nil {
		return nil, err
	}
	return &authzv1.CheckResponse{Allowed: allowed}, nil
}

// GetRoles returns the roles assigned to a subject.
func (g *GRPC) GetRoles(ctx context.Context, req *authzv1.GetRolesRequest) (*authzv1.GetRolesResponse, error) {
	roles, err := g.svc.GetRoles(ctx, req.GetSubject())
	if err != nil {
		return nil, err
	}
	return &authzv1.GetRolesResponse{Roles: roles}, nil
}
