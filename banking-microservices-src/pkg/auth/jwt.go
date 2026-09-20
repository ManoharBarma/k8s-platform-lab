// Package auth provides JWT issuance/validation and Gin middleware shared by
// every service. Tokens are stateless and signed with a platform-wide secret,
// so any service can authenticate a request without calling auth-service.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType distinguishes access from refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims is the JWT payload carried across the platform.
type Claims struct {
	Email string    `json:"email,omitempty"`
	Roles []string  `json:"roles,omitempty"`
	Type  TokenType `json:"typ"`
	jwt.RegisteredClaims
}

// Manager issues and validates JWTs.
type Manager struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewManager builds a token Manager.
func NewManager(secret, issuer string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccess issues a short-lived access token carrying identity + roles.
func (m *Manager) GenerateAccess(subject, email string, roles []string) (string, time.Time, error) {
	exp := time.Now().Add(m.accessTTL)
	return m.sign(Claims{
		Email: email,
		Roles: roles,
		Type:  AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    m.issuer,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}, exp)
}

// GenerateRefresh issues a long-lived refresh token.
func (m *Manager) GenerateRefresh(subject string) (string, time.Time, error) {
	exp := time.Now().Add(m.refreshTTL)
	return m.sign(Claims{
		Type: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    m.issuer,
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}, exp)
}

func (m *Manager) sign(claims Claims, exp time.Time) (string, time.Time, error) {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(m.secret)
	return signed, exp, err
}

// Parse validates a token's signature and expiry and returns its claims.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer))
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
