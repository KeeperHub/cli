package auth

import "time"

// AuthMethod describes how the token was resolved.
type AuthMethod string

const (
	AuthMethodAPIKey AuthMethod = "api-key"
	AuthMethodToken  AuthMethod = "token"
	AuthMethodNone   AuthMethod = "none"
)

// ResolvedToken holds a resolved token and its source.
type ResolvedToken struct {
	Token  string
	Method AuthMethod
	Host   string
}

// TokenInfo holds session details fetched from the server.
type TokenInfo struct {
	UserID    string
	Email     string
	Name      string
	OrgID     string
	OrgName   string
	Role      string
	ExpiresAt time.Time
	Method    AuthMethod
}
