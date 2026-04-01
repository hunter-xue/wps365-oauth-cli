package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client is the HTTP client for OAuth token operations.
type Client struct {
	HTTPClient *http.Client
	Endpoint   string
}

// NewClient creates a Client with a 15-second timeout.
func NewClient(endpoint string) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		Endpoint:   endpoint,
	}
}

// FetchToken executes a token request and returns the parsed response.
func (c *Client) FetchToken(req TokenRequest) (*TokenResponse, error) {
	form := buildForm(req)

	httpResp, err := c.HTTPClient.Post(
		c.Endpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	// Non-2xx: parse error body before returning.
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, parseErrorBody(httpResp.StatusCode, body)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parsing response (status %d): %w\nbody: %s",
			httpResp.StatusCode, err, body)
	}

	tokenResp.FetchedAt = time.Now()

	// Standard OAuth error in a 200 response.
	if tokenResp.IsError() {
		return &tokenResp, fmt.Errorf("OAuth error %q: %s",
			tokenResp.Error, tokenResp.ErrorDescription)
	}

	return &tokenResp, nil
}

// parseErrorBody tries to extract a human-readable error from a non-2xx response.
// It handles both standard OAuth error format and WPS365's custom format.
func parseErrorBody(status int, body []byte) error {
	// Try standard OAuth error: {"error":"..","error_description":".."}
	var stdErr struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if json.Unmarshal(body, &stdErr) == nil && stdErr.Error != "" {
		return fmt.Errorf("HTTP %d: OAuth error %q: %s", status, stdErr.Error, stdErr.Description)
	}

	// Try WPS365 format: {"code":40000005,"msg":"..","debug":{"desc":".."}}
	var wpsErr struct {
		Code  int    `json:"code"`
		Msg   string `json:"msg"`
		Debug struct {
			Desc string `json:"desc"`
		} `json:"debug"`
	}
	if json.Unmarshal(body, &wpsErr) == nil && wpsErr.Msg != "" {
		if wpsErr.Debug.Desc != "" {
			return fmt.Errorf("HTTP %d: %s (code %d): %s", status, wpsErr.Msg, wpsErr.Code, wpsErr.Debug.Desc)
		}
		return fmt.Errorf("HTTP %d: %s (code %d)", status, wpsErr.Msg, wpsErr.Code)
	}

	return fmt.Errorf("HTTP %d: %s", status, body)
}

// buildForm constructs the url.Values for a token request.
func buildForm(req TokenRequest) url.Values {
	form := url.Values{}
	form.Set("grant_type", string(req.GrantType))
	form.Set("client_id", req.ClientID)
	form.Set("client_secret", req.ClientSecret)

	if req.Scopes != "" {
		form.Set("scope", req.Scopes)
	}

	switch req.GrantType {
	case GrantAuthorizationCode:
		form.Set("code", req.Code)
		if req.RedirectURI != "" {
			form.Set("redirect_uri", req.RedirectURI)
		}
	case GrantRefreshToken:
		form.Set("refresh_token", req.RefreshToken)
	}

	return form
}
