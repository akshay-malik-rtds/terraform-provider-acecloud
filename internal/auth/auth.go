package auth

import (
	"context"
	"fmt"
)

// AuthResult holds the resolved authentication credentials. Either Token
// (legacy JWT from the ace-cli config) or APIKeyID + APIKeySecret will be
// populated depending on which method succeeded.
type AuthResult struct {
	Token             string
	APIKeyID          string
	APIKeySecret      string
	APIKeyServiceName string
	Region            string
	ProjectID         string
	Method            string // "api_key", "cli_config"
}

// AuthConfig holds all possible auth inputs from the provider schema and env vars.
type AuthConfig struct {
	APIURL            string
	APIKeyID          string
	APIKeySecret      string
	APIKeyServiceName string
	Region            string
	ProjectID         string
	ACEConfigPath     string
}

// Resolve attempts authentication methods in priority order:
//  1. API key (api_key_id + api_key_secret + api_key_service_name, or env vars)
//  2. ace-cli config file (~/.ace/config.json) — supplies an API key or
//     legacy token written by 'ace auth login' / 'ace auth login-api-key'
//
// Returns AuthResult or error describing what failed.
func Resolve(_ context.Context, cfg AuthConfig) (*AuthResult, error) {
	// Method 1: API key — primary, recommended path.
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

	// Method 2: ace-cli config file (supports either API key or legacy token).
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
			"  1. api_key_id + api_key_secret + api_key_service_name in provider config, or the matching ACECLOUD_API_KEY_* environment variables (recommended for automation)\n" +
			"  2. Login via ace-cli first: ace auth login-api-key --service-name <name>",
	)
}
