// Package domain holds the RBAC model: roles, permissions and their mappings.
package domain

// Wildcard is a permission that grants everything (held by the admin role).
const Wildcard = "*"

// Role is a named collection of permissions.
type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions,omitempty"`
}

// Permission is a single capability, conventionally "<resource>:<action>".
type Permission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
