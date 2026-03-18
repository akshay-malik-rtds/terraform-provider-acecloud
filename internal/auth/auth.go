package auth

import (
	"context"
	"fmt"
)

// AuthResult holds the resolved authentication credentials.
type AuthResult struct {
	Token     string
	Region    string
	ProjectID string
	Method    string // "static_token", "email_password", "cli_config"
}

// AuthConfig holds all possible auth inputs from the provider schema and env vars.
type AuthConfig struct {
	APIURL        string
	APIToken      string
	Region        string
	ProjectID     string
	Email         string
	Password      string
	ACEConfigPath string
}

// Resolve attempts authentication methods in priority order:
//  1. Static API token (explicit or env var)
//  2. Email/password login (POST /auth/login)
//  3. ace-cli config file (~/.ace/config.json)
//
// Returns AuthResult or error describing what failed.
func Resolve(ctx context.Context, cfg AuthConfig) (*AuthResult, error) {
	// Method 1: Static token — highest priority.
	if cfg.APIToken != "" {
		return &AuthResult{
			Token:     cfg.APIToken,
			Region:    cfg.Region,
			ProjectID: cfg.ProjectID,
			Method:    "static_token",
		}, nil
	}

	// Method 2: Email/password login.
	if cfg.Email != "" && cfg.Password != "" {
		return loginWithCredentials(ctx, cfg)
	}

	// Method 3: ace-cli config file.
	result, err := readCLIConfig(cfg.ACEConfigPath)
	if err == nil && result.Token != "" {
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
			"  1. api_token in provider config or ACECLOUD_API_TOKEN environment variable\n" +
			"  2. email + password in provider config or ACECLOUD_EMAIL + ACECLOUD_PASSWORD environment variables\n" +
			"  3. Login via ace-cli first: ace auth login",
	)
}
