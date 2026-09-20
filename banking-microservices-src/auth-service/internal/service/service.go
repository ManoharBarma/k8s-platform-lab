// Package service implements auth-service use cases: registration, login,
// token refresh and logout.
package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/password"
	"banking-platform/services/auth-service/internal/domain"
	"banking-platform/services/auth-service/internal/repository"
	"banking-platform/services/auth-service/internal/session"
)

// DefaultRole is granted to every newly registered user.
const DefaultRole = "customer"

// Tokens is the token pair returned on login/register/refresh.
type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Service holds auth-service dependencies.
type Service struct {
	repo     repository.Repository
	sessions session.Store
	tokens   *auth.Manager
	log      *zap.Logger
}

// New builds the Service.
func New(repo repository.Repository, sessions session.Store, tokens *auth.Manager, log *zap.Logger) *Service {
	return &Service{repo: repo, sessions: sessions, tokens: tokens, log: log}
}

// Register creates a new user and returns an initial token pair.
func (s *Service) Register(ctx context.Context, email, pw, fullName string) (domain.User, Tokens, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || pw == "" {
		return domain.User{}, Tokens{}, apierror.ErrBadRequest("email and password are required")
	}
	if len(pw) < 8 {
		return domain.User{}, Tokens{}, apierror.ErrBadRequest("password must be at least 8 characters")
	}

	hash, err := password.Hash(pw)
	if err != nil {
		return domain.User{}, Tokens{}, err
	}
	user := domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hash,
		FullName:     fullName,
		Roles:        []string{DefaultRole},
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return domain.User{}, Tokens{}, err
	}

	tokens, err := s.issue(ctx, user)
	if err != nil {
		return domain.User{}, Tokens{}, err
	}
	return user, tokens, nil
}

// Login verifies credentials and returns a token pair.
func (s *Service) Login(ctx context.Context, email, pw string) (Tokens, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		// Do not leak whether the email exists.
		return Tokens{}, apierror.ErrUnauthorized("invalid credentials")
	}
	if !password.Verify(user.PasswordHash, pw) {
		return Tokens{}, apierror.ErrUnauthorized("invalid credentials")
	}
	return s.issue(ctx, user)
}

// Refresh validates a refresh token (signature, expiry and not-revoked) and
// issues a new token pair, rotating the refresh token.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (Tokens, error) {
	claims, err := s.tokens.Parse(refreshToken)
	if err != nil || claims.Type != auth.RefreshToken {
		return Tokens{}, apierror.ErrUnauthorized("invalid refresh token")
	}
	ok, err := s.sessions.Exists(ctx, refreshToken)
	if err != nil {
		return Tokens{}, err
	}
	if !ok {
		return Tokens{}, apierror.ErrUnauthorized("refresh token revoked or expired")
	}
	user, err := s.repo.GetByID(ctx, claims.Subject)
	if err != nil {
		return Tokens{}, apierror.ErrUnauthorized("invalid refresh token")
	}
	// Rotate: revoke the old refresh token.
	_ = s.sessions.Delete(ctx, refreshToken)
	return s.issue(ctx, user)
}

// Logout revokes a refresh token.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.sessions.Delete(ctx, refreshToken)
}

// ValidateAccess parses and validates an access token, returning its claims.
func (s *Service) ValidateAccess(accessToken string) (*auth.Claims, error) {
	claims, err := s.tokens.Parse(accessToken)
	if err != nil || claims.Type != auth.AccessToken {
		return nil, apierror.ErrUnauthorized("invalid access token")
	}
	return claims, nil
}

func (s *Service) issue(ctx context.Context, user domain.User) (Tokens, error) {
	access, exp, err := s.tokens.GenerateAccess(user.ID, user.Email, user.Roles)
	if err != nil {
		return Tokens{}, err
	}
	refresh, refreshExp, err := s.tokens.GenerateRefresh(user.ID)
	if err != nil {
		return Tokens{}, err
	}
	if err := s.sessions.Save(ctx, refresh, user.ID, time.Until(refreshExp)); err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: refresh, ExpiresAt: exp}, nil
}
