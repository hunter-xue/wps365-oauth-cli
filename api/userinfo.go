// Package api provides WPS365 Open API client functions.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"oauth_tools/sign"
)

// UserInfo holds the current user's profile from GET /v7/users/current.
type UserInfo struct {
	ID        string `json:"id"`
	UserName  string `json:"user_name"`
	Avatar    string `json:"avatar"`
	CompanyID string `json:"company_id"`
}

type userInfoResponse struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data UserInfo `json:"data"`
}

// GetCurrentUser fetches the authenticated user's info using the given access token.
// It applies KSO-1 signing with the provided signer.
func GetCurrentUser(baseURL, accessToken string, signer *sign.KsoSign) (*UserInfo, error) {
	url := baseURL + "/users/current"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	if err := signer.Apply(req, nil); err != nil {
		return nil, fmt.Errorf("signing request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result userInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response (status %d): %w\nbody: %s", resp.StatusCode, err, body)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error (code %d): %s", result.Code, result.Msg)
	}

	return &result.Data, nil
}
