package oauth

import "time"

// GrantType enumerates supported OAuth grant types.
type GrantType string

const (
	GrantClientCredentials GrantType = "client_credentials"
	GrantAuthorizationCode GrantType = "authorization_code"
	GrantRefreshToken      GrantType = "refresh_token"
)

// TokenRequest carries all parameters needed for a token request.
// Fields unused by a given GrantType are left at zero value.
type TokenRequest struct {
	GrantType    GrantType
	ClientID     string
	ClientSecret string
	Scopes       string

	// authorization_code flow
	Code        string
	RedirectURI string

	// refresh_token flow
	RefreshToken string
}

// TokenResponse mirrors the standard OAuth 2.0 token response (RFC 6749 §5.1).
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`

	// Error fields (RFC 6749 §5.2)
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`

	// FetchedAt records retrieval time for cache/expiry calculations.
	FetchedAt time.Time `json:"-"`
}

// IsError returns true when the response contains an OAuth error.
func (r *TokenResponse) IsError() bool {
	return r.Error != ""
}

// ExpiresAt returns the absolute expiry time, or zero if unknown.
func (r *TokenResponse) ExpiresAt() time.Time {
	if r.ExpiresIn == 0 || r.FetchedAt.IsZero() {
		return time.Time{}
	}
	return r.FetchedAt.Add(time.Duration(r.ExpiresIn) * time.Second)
}
