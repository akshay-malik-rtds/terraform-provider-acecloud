package auth

import (
	"context"
	"fmt"
)

// AuthResult holds the resolved authentication credentials. Either Token
// (JWT bearer) or APIKeyID + APIKeySecret will be populated depending on
// which method succeeded.
type AuthResult struct {
	Token             string
	APIKeyID          string
	APIKeySecret      string
	APIKeyServiceName string
	Region            string
	ProjectID         string
	Method            string // "api_key", "static_token", "email_password", "cli_config"
}

// AuthConfig holds all possible auth inputs from the provider schema and env vars.
type AuthConfig struct {
	APIURL            string
	APIToken          string
	APIKeyID          string
	APIKeySecret      string
	APIKeyServiceName string
	Region            string
	ProjectID         string
	Email             string
	Password          string
	ACEConfigPath     string
}

// Resolve attempts authentication methods in priority order:
//  1. API key (api_key_id + api_key_secret, or env vars)
//  2. Static API token (explicit or env var)
//  3. Email/password login (POST /auth/login)
//  4. ace-cli config file (~/.ace/config.json) — supports either token or API key
//
// Returns AuthResult or error describing what failed.
func Resolve(ctx context.Context, cfg AuthConfig) (*AuthResult, error) {
	// Method 1: API key — highest priority.
	if cfg.APIKeyID != "" && cfg.APIKeySecret != "" {
		return &AuthResult{
			APIKeyID:          cfg.APIKeyID,
			APIKeySecret:      cfg.APIKeySecret,
			APIKeyServiceName: cfg.APIKeyServiceName,
			Region:            cfg.Region,
			ProjectID:         cfg.ProjectID,
			Method:            "api_key",
		}, nil
	}

	// Method 2: Static token.
	if cfg.APIToken != "" {
		return &AuthResult{
			Token:     cfg.APIToken,
			Region:    cfg.Region,
			ProjectID: cfg.ProjectID,
			Method:    "static_token",
		}, nil
	}

	// Method 3: Email/password login.
	if cfg.Email != "" && cfg.Password != "" {
		return loginWithCredentials(ctx, cfg)
	}

	// Method 4: ace-cli config file (handles either token or API key).
	result, err := readCLIConfig(cfg.ACEConfigPath)
	if err == nil && (result.Token != "" || (result.APIKeyID != "" && result.APIKeySecret != "")) {
		// Explicit config values override CLI config defaults.
		if cfg.Region != "" {
			result.Region = cfg.Region
		}
		if cfg.ProjectID != "" {
			result.ProjectID = cfg.ProjectID
		}
		return result, nil
	}

	return nil, fmt.Errorf(
		"no authentication method succeeded. Provide one of:\n" +
			"  1. api_key_id + api_key_secret in provider config, or ACECLOUD_API_KEY_ID + ACECLOUD_API_KEY_SECRET env vars (recommended for automation)\n" +
			"  2. api_token in provider config or ACECLOUD_API_TOKEN environment variable\n" +
			"  3. email + password in provider config or ACECLOUD_EMAIL + ACECLOUD_PASSWORD environment variables\n" +
			"  4. Login via ace-cli first: ace auth login or ace auth login-api-key",
	)
}
