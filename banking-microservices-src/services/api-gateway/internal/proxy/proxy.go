// Package proxy implements the gateway's routing table and reverse-proxy engine.
// Requests to /api/v1/<resource>/... are forwarded to the owning microservice.
package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
)

// route maps a resource segment to its owning service and whether the gateway
// lets it through without a valid access token (the service still enforces its
// own authorization).
type route struct {
	service string
	public  bool
}

// routes is the platform routing table keyed by the first path segment after
// /api/v1/.
var routes = map[string]route{
	"auth":               {"auth-service", true},
	"authz":              {"authz-service", false},
	"customers":          {"customer-service", false},
	"profiles":           {"profile-service", false},
	"kyc":                {"kyc-service", false},
	"documents":          {"document-service", false},
	"accounts":           {"account-service", false},
	"transactions":       {"transaction-service", false},
	"beneficiaries":      {"beneficiary-service", false},
	"payments":           {"payment-service", false},
	"wallets":            {"wallet-service", false},
	"loans":              {"loan-service", false},
	"emi":                {"emi-service", false},
	"fixed-deposits":     {"fixed-deposit-service", false},
	"recurring-deposits": {"recurring-deposit-service", false},
	"cards":              {"card-service", false},
	"investments":        {"investment-service", false},
	"fx":                 {"currency-exchange-service", true}, // /fx/convert still auth'd by the service
	"notifications":      {"notification-service", false},
	"statements":         {"statement-service", false},
	"fraud":              {"fraud-service", false},
	"audit":              {"audit-service", false},
	"reports":            {"reports-service", false},
	"analytics":          {"analytics-service", false},
	"search":             {"search-service", false},
	"email":              {"email-service", false},
	"sms":                {"sms-service", false},
	"support":            {"support-service", false},
	"admin":              {"admin-service", false},
}

// Gateway routes and proxies API requests to backend services.
type Gateway struct {
	proxies map[string]*httputil.ReverseProxy
	tokens  *auth.Manager
	log     *zap.Logger
}

// New builds a Gateway, pre-creating one reverse proxy per distinct service.
func New(scheme, port string, tokens *auth.Manager, log *zap.Logger) (*Gateway, error) {
	g := &Gateway{proxies: map[string]*httputil.ReverseProxy{}, tokens: tokens, log: log}
	seen := map[string]bool{}
	for _, r := range routes {
		if seen[r.service] {
			continue
		}
		seen[r.service] = true
		target, err := url.Parse(fmt.Sprintf("%s://%s:%s", scheme, r.service, port))
		if err != nil {
			return nil, err
		}
		g.proxies[r.service] = httputil.NewSingleHostReverseProxy(target)
	}
	return g, nil
}

// Handler is the catch-all Gin handler that authenticates (for protected
// resources) and reverse-proxies to the owning service.
func (g *Gateway) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/v1/") {
			c.JSON(http.StatusNotFound, apierror.ErrNotFound("unknown route"))
			return
		}
		resource, _, _ := strings.Cut(strings.TrimPrefix(path, "/api/v1/"), "/")
		r, ok := routes[resource]
		if !ok {
			c.JSON(http.StatusNotFound, apierror.ErrNotFound("unknown route"))
			return
		}

		if !r.public && c.Request.Method != http.MethodOptions {
			claims, err := g.authenticate(c)
			if err != nil {
				c.JSON(http.StatusUnauthorized, apierror.ErrUnauthorized("invalid or missing token"))
				return
			}
			// Propagate identity to the upstream service.
			c.Request.Header.Set("X-User-Id", claims.Subject)
			c.Request.Header.Set("X-User-Roles", strings.Join(claims.Roles, ","))
		}

		g.proxies[r.service].ServeHTTP(c.Writer, c.Request)
	}
}

func (g *Gateway) authenticate(c *gin.Context) (*auth.Claims, error) {
	h := c.GetHeader("Authorization")
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, apierror.ErrUnauthorized("missing bearer token")
	}
	claims, err := g.tokens.Parse(strings.TrimSpace(parts[1]))
	if err != nil || claims.Type != auth.AccessToken {
		return nil, apierror.ErrUnauthorized("invalid token")
	}
	return claims, nil
}
