package auth

import (
	"strings"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
)

const (
	ctxSubject = "auth_subject"
	ctxEmail   = "auth_email"
	ctxRoles   = "auth_roles"
)

// Middleware validates the Bearer access token and stores identity in the Gin
// context. Requests without a valid access token are rejected with 401.
func (m *Manager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := bearer(c)
		if raw == "" {
			abort(c, apierror.ErrUnauthorized("missing bearer token"))
			return
		}
		claims, err := m.Parse(raw)
		if err != nil {
			abort(c, apierror.ErrUnauthorized("invalid or expired token"))
			return
		}
		if claims.Type != AccessToken {
			abort(c, apierror.ErrUnauthorized("access token required"))
			return
		}
		c.Set(ctxSubject, claims.Subject)
		c.Set(ctxEmail, claims.Email)
		c.Set(ctxRoles, claims.Roles)
		c.Next()
	}
}

// RequireRole ensures the authenticated subject has at least one of roles.
// Must be used after Middleware.
func RequireRole(roles ...string) gin.HandlerFunc {
	want := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		want[r] = struct{}{}
	}
	return func(c *gin.Context) {
		for _, r := range Roles(c) {
			if _, ok := want[r]; ok {
				c.Next()
				return
			}
		}
		abort(c, apierror.ErrForbidden("insufficient role"))
	}
}

// Subject returns the authenticated user's ID (empty if unauthenticated).
func Subject(c *gin.Context) string { return c.GetString(ctxSubject) }

// Email returns the authenticated user's email.
func Email(c *gin.Context) string { return c.GetString(ctxEmail) }

// Roles returns the authenticated user's roles.
func Roles(c *gin.Context) []string {
	v, ok := c.Get(ctxRoles)
	if !ok {
		return nil
	}
	roles, _ := v.([]string)
	return roles
}

func bearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func abort(c *gin.Context, err *apierror.Error) {
	c.AbortWithStatusJSON(err.Status, err)
}
