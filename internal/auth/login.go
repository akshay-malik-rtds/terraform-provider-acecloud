package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// loginRequest is the POST /auth/login request body.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse matches the npc-api response envelope for POST /auth/login.
type loginResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		AccessToken      string          `json:"accessToken"`
		ExpiresInSeconds int             `json:"expiresInSeconds"`
		MFAEnabled       bool            `json:"mfaEnabled"`
		Projects         json.RawMessage `json:"projects"`
	} `json:"data"`
}

// loginWithCredentials calls POST /auth/login and returns an AuthResult.
func loginWithCredentials(ctx context.Context, cfg AuthConfig) (*AuthResult, error) {
	if cfg.APIURL == "" {
		return nil, fmt.Errorf("api_url is required for email/password login")
	}

	body, err := json.Marshal(loginRequest{Email: cfg.Email, Password: cfg.Password})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.APIURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 120 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read login response: %w", err)
	}

	var loginResp loginResponse
	if err := json.Unmarshal(respBody, &loginResp); err != nil {
		return nil, fmt.Errorf("unexpected login response format: %w", err)
	}

	if loginResp.Error {
		return nil, fmt.Errorf("login failed: %s", loginResp.Message)
	}

	if loginResp.Data.MFAEnabled {
		return nil, fmt.Errorf(
			"your account has 2FA enabled. Email/password authentication is not supported " +
				"with 2FA. Please provide a developer token or pre-authenticated JWT via " +
				"the api_token provider attribute or the ACECLOUD_API_TOKEN environment variable",
		)
	}

	if loginResp.Data.AccessToken == "" {
		return nil, fmt.Errorf("login succeeded but no access token was returned")
	}

	result := &AuthResult{
		Token:     loginResp.Data.AccessToken,
		Region:    cfg.Region,
		ProjectID: cfg.ProjectID,
		Method:    "email_password",
	}

	// Auto-select project/region from login response when not explicitly set.
	if loginResp.Data.Projects != nil && (result.Region == "" || result.ProjectID == "") {
		autoSelectProject(loginResp.Data.Projects, result)
	}

	return result, nil
}

// projectInfo holds the shape of a single project entry from the login response.
type projectInfo struct {
	Name    string                     `json:"name"`
	Status  string                     `json:"status"`
	Regions map[string]json.RawMessage `json:"regions"`
}

// autoSelectProject picks the first active project and region from the login
// response, matching the ace-cli behavior.
func autoSelectProject(projectsRaw json.RawMessage, result *AuthResult) {
	var projects map[string]projectInfo
	if err := json.Unmarshal(projectsRaw, &projects); err != nil {
		return
	}

	for projectID, proj := range projects {
		if !strings.EqualFold(proj.Status, "active") {
			continue
		}
		if result.ProjectID == "" {
			result.ProjectID = projectID
		}
		if result.Region == "" {
			for regionName := range proj.Regions {
				result.Region = regionName
				break
			}
		}
		break
	}
}
