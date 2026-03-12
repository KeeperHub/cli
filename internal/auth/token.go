package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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

type sessionResponse struct {
	User    *sessionUser    `json:"user"`
	Session *sessionSession `json:"session"`
}

type sessionUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type sessionSession struct {
	ExpiresAt            string `json:"expiresAt"`
	ActiveOrganizationID string `json:"activeOrganizationId"`
}

type orgMembership struct {
	Organization struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"organization"`
	Role string `json:"role"`
}

// FetchTokenInfo queries the server for session details using the given token.
// Returns TokenInfo on success, or an error if the token is invalid.
func FetchTokenInfo(host, token string) (TokenInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, "https://"+host+"/api/auth/get-session", nil)
	if err != nil {
		return TokenInfo{}, fmt.Errorf("creating session request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return TokenInfo{}, fmt.Errorf("fetching session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TokenInfo{}, fmt.Errorf("token is invalid or expired")
	}

	var session sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return TokenInfo{}, fmt.Errorf("decoding session response: %w", err)
	}

	if session.User == nil {
		return TokenInfo{}, fmt.Errorf("token is invalid or expired")
	}

	info := TokenInfo{
		UserID: session.User.ID,
		Email:  session.User.Email,
		Name:   session.User.Name,
	}

	if session.Session != nil {
		if session.Session.ExpiresAt != "" {
			if t, parseErr := time.Parse(time.RFC3339, session.Session.ExpiresAt); parseErr == nil {
				info.ExpiresAt = t
			}
		}
		info.OrgID = session.Session.ActiveOrganizationID
	}

	// Attempt to fetch org name and role; failure is non-fatal.
	if info.OrgID != "" {
		orgName, role := fetchOrgDetails(client, host, token, info.OrgID)
		info.OrgName = orgName
		info.Role = role
	}
	if info.OrgName == "" && info.OrgID != "" {
		info.OrgName = info.OrgID
	}
	if info.Role == "" {
		info.Role = "unknown"
	}

	return info, nil
}

func fetchOrgDetails(client *http.Client, host, token, orgID string) (string, string) {
	req, err := http.NewRequest(http.MethodGet, "https://"+host+"/api/organizations/"+orgID, nil)
	if err != nil {
		return "", ""
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return "", ""
	}
	defer resp.Body.Close()

	var membership orgMembership
	if err := json.NewDecoder(resp.Body).Decode(&membership); err != nil {
		return "", ""
	}

	return membership.Organization.Name, membership.Role
}
