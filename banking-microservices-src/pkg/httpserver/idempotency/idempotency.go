// Package idempotency extracts the client-supplied Idempotency-Key used to make
// money-moving endpoints safe to retry. When absent, a fresh key is generated
// (so a single request still works, but the caller forgoes cross-retry dedup).
package idempotency

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Header is the standard idempotency header name.
const Header = "Idempotency-Key"

// Key returns the request's idempotency key, generating one if not provided.
func Key(c *gin.Context) string {
	k := c.GetHeader(Header)
	if k == "" {
		k = uuid.NewString()
	}
	return k
}
