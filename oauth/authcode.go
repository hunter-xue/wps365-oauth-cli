package oauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
)

// AuthCodeConfig holds the parameters needed for the authorization_code flow.
type AuthCodeConfig struct {
	// AuthURL is the pre-built authorization URL.
	// The redirect_uri embedded in its query params is used to start the local server.
	AuthURL string

	// Scopes to request. If set, overrides any scope already in AuthURL.
	Scopes string

	ClientID     string
	ClientSecret string
	Endpoint     string
}

// AuthorizationCodeFlow runs the full browser-based OAuth flow:
//  1. Parse redirect_uri from AuthURL to determine local server port
//  2. Start a local HTTP server on that port
//  3. Open the browser to AuthURL
//  4. Wait for the callback with the authorization code
//  5. Exchange the code for a token
func AuthorizationCodeFlow(cfg AuthCodeConfig) (*TokenResponse, error) {
	// Extract redirect_uri from the AUTH_URL query params — single source of truth.
	parsedAuth, err := url.Parse(cfg.AuthURL)
	if err != nil {
		return nil, fmt.Errorf("invalid AUTH_URL %q: %w", cfg.AuthURL, err)
	}
	if parsedAuth.Scheme != "http" && parsedAuth.Scheme != "https" {
		return nil, fmt.Errorf("AUTH_URL must start with http:// or https://, got %q", cfg.AuthURL)
	}
	rawRedirect := parsedAuth.Query().Get("redirect_uri")
	if rawRedirect == "" {
		return nil, fmt.Errorf("AUTH_URL must contain a redirect_uri query parameter")
	}
	redirectURL, err := url.Parse(rawRedirect)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect_uri in AUTH_URL %q: %w", rawRedirect, err)
	}

	// If Scopes is set, inject it into the AUTH_URL (override any existing scope param).
	authURL := cfg.AuthURL
	if cfg.Scopes != "" {
		q := parsedAuth.Query()
		q.Set("scope", cfg.Scopes)
		parsedAuth.RawQuery = q.Encode()
		authURL = parsedAuth.String()
	}

	// Use the port from REDIRECT_URI; fall back to a random port.
	port := redirectURL.Port()
	if port == "" {
		port = "0"
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	callbackPath := redirectURL.Path
	if callbackPath == "" {
		callbackPath = "/"
	}

	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errParam := r.URL.Query().Get("error")

		if errParam != "" {
			desc := r.URL.Query().Get("error_description")
			fmt.Fprintf(w, "<h2>Authorization failed</h2><p>%s: %s</p><p>You may close this tab.</p>",
				errParam, desc)
			errCh <- fmt.Errorf("authorization denied: %s — %s", errParam, desc)
			return
		}
		if code == "" {
			fmt.Fprintf(w, "<h2>No code received</h2><p>You may close this tab.</p>")
			errCh <- fmt.Errorf("callback received no code parameter")
			return
		}

		fmt.Fprintf(w, "<h2>Authorization successful</h2><p>You may close this tab.</p>")
		codeCh <- code
	})

	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return nil, fmt.Errorf("starting callback server on port %s: %w", port, err)
	}

	srv := &http.Server{Handler: mux}
	go srv.Serve(listener) //nolint:errcheck

	actualPort := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Listening for OAuth callback on http://127.0.0.1:%d%s\n", actualPort, callbackPath)
	fmt.Printf("Opening browser: %s\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically.\nPlease open this URL manually:\n  %s\n", authURL)
	}

	// Wait for code or error with a 2-minute timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		srv.Shutdown(context.Background()) //nolint:errcheck
		return nil, err
	case <-ctx.Done():
		srv.Shutdown(context.Background()) //nolint:errcheck
		return nil, fmt.Errorf("timed out waiting for authorization callback")
	}

	srv.Shutdown(context.Background()) //nolint:errcheck

	// Exchange the code for a token.
	client := NewClient(cfg.Endpoint)
	return client.FetchToken(TokenRequest{
		GrantType:    GrantAuthorizationCode,
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Code:         code,
		RedirectURI:  rawRedirect,
	})
}

// openBrowser opens the given URL in the default browser.
func openBrowser(u string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{u}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", u}
	default: // linux, freebsd, etc.
		cmd = "xdg-open"
		args = []string{u}
	}

	return exec.Command(cmd, args...).Start()
}
