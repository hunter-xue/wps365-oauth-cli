package cmd

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"oauth_tools/api"
	"oauth_tools/config"
	"oauth_tools/oauth"
	"oauth_tools/sign"
)

func tokenUsage(fs *flag.FlagSet) {
	fmt.Fprint(os.Stderr, `USAGE:
  oauth_tools token [flags]

FLAGS:
`)
	fs.PrintDefaults()
	fmt.Fprint(os.Stderr, `
EXAMPLES:
  oauth_tools token
      获取用户 token（打开浏览器授权，并打印当前用户信息）

  oauth_tools token -type tenant
      获取租户 token（直接用 AppID/Secret，无需浏览器）

  oauth_tools token -token-only
      仅输出 access_token（适合脚本）：
      export TOKEN=$(oauth_tools token -token-only)

  oauth_tools token -type tenant -json
      以 JSON 格式输出租户 token
`)
}

// RunToken handles the "token" subcommand.
func RunToken(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("token", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	tokenType := fs.String("type", "user", "Token type: user (authorization_code) | tenant (client_credentials)")
	rawJSON := fs.Bool("json", false, "Output raw JSON response")
	tokenOnly := fs.Bool("token-only", false, "Output only the access_token value (pipe-friendly)")

	fs.Usage = func() { tokenUsage(fs) }

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	var (
		resp *oauth.TokenResponse
		err  error
	)

	switch *tokenType {
	case "user":
		if cfg.AuthURL == "" {
			return fmt.Errorf("AUTH_URL must be set in .env for user token")
		}
		resp, err = oauth.AuthorizationCodeFlow(oauth.AuthCodeConfig{
			AuthURL:      cfg.AuthURL,
			Scopes:       cfg.Scopes,
			ClientID:     cfg.AppID,
			ClientSecret: cfg.Secret,
			Endpoint:     cfg.Endpoint,
		})

	case "tenant":
		client := oauth.NewClient(cfg.Endpoint)
		resp, err = client.FetchToken(oauth.TokenRequest{
			GrantType:    oauth.GrantClientCredentials,
			ClientID:     cfg.AppID,
			ClientSecret: cfg.Secret,
		})

	default:
		return fmt.Errorf("unsupported token type: %q (supported: user, tenant)", *tokenType)
	}

	if err != nil {
		if resp != nil && *rawJSON {
			printJSON(resp)
		}
		return err
	}

	switch {
	case *tokenOnly:
		fmt.Println(resp.AccessToken)
	case *rawJSON:
		printJSON(resp)
	default:
		printPretty(resp, cfg.APIBaseURL)
		if *tokenType == "user" {
			fetchAndPrintUserInfo(cfg, resp.AccessToken)
		}
	}

	return nil
}

func fetchAndPrintUserInfo(cfg *config.Config, accessToken string) {
	if cfg.APIBaseURL == "" {
		return
	}

	signer, err := sign.New(cfg.AppID, cfg.Secret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sign init error: %v\n", err)
		return
	}

	user, err := api.GetCurrentUser(cfg.APIBaseURL, accessToken, signer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get user info error: %v\n", err)
		return
	}

	fmt.Println("────────────────────────────────────────")
	fmt.Printf("user_id:       %s\n", user.ID)
	fmt.Printf("user_name:     %s\n", user.UserName)
	fmt.Printf("company_id:    %s\n", user.CompanyID)
	if user.Avatar != "" {
		fmt.Printf("avatar:        %s\n", user.Avatar)
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func printJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func printPretty(resp *oauth.TokenResponse, apiBaseURL string) {
	fmt.Println("────────────────────────────────────────")
	fmt.Printf("access_token:  %s\n", resp.AccessToken)
	fmt.Printf("token_type:    %s\n", resp.TokenType)
	fmt.Printf("expires_in:    %ds\n", resp.ExpiresIn)
	if resp.RefreshToken != "" {
		fmt.Printf("refresh_token: %s\n", resp.RefreshToken)
	}
	if resp.Scope != "" {
		fmt.Printf("scope:         %s\n", resp.Scope)
	}
	if !resp.ExpiresAt().IsZero() {
		fmt.Printf("expires_at:    %s\n", resp.ExpiresAt().Format("2006-01-02 15:04:05 UTC"))
	}
	if apiBaseURL != "" {
		fmt.Printf("api_base_url:  %s\n", apiBaseURL)
	}
	fmt.Println("────────────────────────────────────────")
	fmt.Printf("Authorization: %s %s\n", capitalize(resp.TokenType), resp.AccessToken)
}
