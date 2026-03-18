package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolve_StaticToken(t *testing.T) {
	cfg := AuthConfig{
		APIToken:  "static-jwt-token",
		Region:    "mumbai",
		ProjectID: "proj-1",
		// Email/password also set — should be ignored.
		Email:    "user@test.com",
		Password: "pass",
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "static_token" {
		t.Errorf("expected method static_token, got %s", result.Method)
	}
	if result.Token != "static-jwt-token" {
		t.Errorf("expected token static-jwt-token, got %s", result.Token)
	}
}

func TestResolve_EmailPassword(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := loginResponse{}
		resp.Error = false
		resp.Message = "SUCCESS"
		resp.Data.AccessToken = "login-jwt"
		resp.Data.ExpiresInSeconds = 172800
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:    ts.URL,
		Email:     "user@test.com",
		Password:  "pass",
		Region:    "mumbai",
		ProjectID: "proj-1",
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "email_password" {
		t.Errorf("expected method email_password, got %s", result.Method)
	}
	if result.Token != "login-jwt" {
		t.Errorf("expected token login-jwt, got %s", result.Token)
	}
}

func TestResolve_CLIConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"api_base_url": "https://api.acecloud.ai",
		"token": "cli-token-abc",
		"region": "delhi",
		"project_id": "proj-cli"
	}`), 0600)

	cfg := AuthConfig{
		ACEConfigPath: cfgPath,
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "cli_config" {
		t.Errorf("expected method cli_config, got %s", result.Method)
	}
	if result.Token != "cli-token-abc" {
		t.Errorf("expected token cli-token-abc, got %s", result.Token)
	}
	if result.Region != "delhi" {
		t.Errorf("expected region delhi, got %s", result.Region)
	}
	if result.ProjectID != "proj-cli" {
		t.Errorf("expected project_id proj-cli, got %s", result.ProjectID)
	}
}

func TestResolve_Priority_StaticWins(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"token": "cli-token",
		"region": "delhi",
		"project_id": "proj-cli"
	}`), 0600)

	cfg := AuthConfig{
		APIToken:      "static-wins",
		Region:        "mumbai",
		ProjectID:     "proj-static",
		Email:         "user@test.com",
		Password:      "pass",
		ACEConfigPath: cfgPath,
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "static_token" {
		t.Errorf("expected static_token to win, got %s", result.Method)
	}
	if result.Token != "static-wins" {
		t.Errorf("expected token static-wins, got %s", result.Token)
	}
}

func TestResolve_AllFail(t *testing.T) {
	cfg := AuthConfig{
		ACEConfigPath: "/tmp/nonexistent-ace-config-99999/config.json",
	}

	_, err := Resolve(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error when all methods fail")
	}
	if !strings.Contains(err.Error(), "no authentication method succeeded") {
		t.Errorf("expected comprehensive error, got: %s", err.Error())
	}
}

func TestResolve_CLIConfigOverrides(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"token": "cli-token",
		"region": "delhi",
		"project_id": "proj-from-cli"
	}`), 0600)

	cfg := AuthConfig{
		Region:        "mumbai",
		ProjectID:     "proj-explicit",
		ACEConfigPath: cfgPath,
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "cli_config" {
		t.Errorf("expected method cli_config, got %s", result.Method)
	}
	// Explicit values should override CLI config values.
	if result.Region != "mumbai" {
		t.Errorf("expected explicit region mumbai, got %s", result.Region)
	}
	if result.ProjectID != "proj-explicit" {
		t.Errorf("expected explicit project_id proj-explicit, got %s", result.ProjectID)
	}
}

// ---------------------------------------------------------------------------
// Extended Auth Edge Cases
// ---------------------------------------------------------------------------

func TestResolve_EmailOnlyNoPassword(t *testing.T) {
	// Email set without password → should skip email/password method and fall through to CLI config.
	cfg := AuthConfig{
		Email:         "user@test.com",
		ACEConfigPath: "/tmp/nonexistent-ace-config-99999/config.json",
	}

	_, err := Resolve(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error — email-only should not attempt login")
	}
	if !strings.Contains(err.Error(), "no authentication method succeeded") {
		t.Errorf("expected all-fail error, got: %s", err.Error())
	}
}

func TestResolve_PasswordOnlyNoEmail(t *testing.T) {
	// Password set without email → should skip email/password method.
	cfg := AuthConfig{
		Password:      "pass123",
		ACEConfigPath: "/tmp/nonexistent-ace-config-99999/config.json",
	}

	_, err := Resolve(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error — password-only should not attempt login")
	}
	if !strings.Contains(err.Error(), "no authentication method succeeded") {
		t.Errorf("expected all-fail error, got: %s", err.Error())
	}
}

func TestResolve_EmailPasswordOverCLI(t *testing.T) {
	// Verify email/password wins over CLI config.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := loginResponse{}
		resp.Error = false
		resp.Message = "SUCCESS"
		resp.Data.AccessToken = "email-jwt"
		resp.Data.ExpiresInSeconds = 172800
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"token": "cli-token",
		"region": "delhi",
		"project_id": "proj-cli"
	}`), 0600)

	cfg := AuthConfig{
		APIURL:        ts.URL,
		Email:         "user@test.com",
		Password:      "pass",
		Region:        "mumbai",
		ProjectID:     "proj-1",
		ACEConfigPath: cfgPath,
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != "email_password" {
		t.Errorf("expected email_password to win over CLI, got %s", result.Method)
	}
	if result.Token != "email-jwt" {
		t.Errorf("expected email-jwt, got %s", result.Token)
	}
}

func TestResolve_CLIConfigEmptyToken(t *testing.T) {
	// CLI config with empty token → should fail.
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	os.WriteFile(cfgPath, []byte(`{
		"token": "",
		"region": "delhi",
		"project_id": "proj-cli"
	}`), 0600)

	cfg := AuthConfig{
		ACEConfigPath: cfgPath,
	}

	_, err := Resolve(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for empty CLI token")
	}
}

func TestResolve_StaticTokenEmptyRegionProject(t *testing.T) {
	// Static token with no region/project → still succeeds (token is the only requirement).
	cfg := AuthConfig{
		APIToken: "my-token",
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token != "my-token" {
		t.Errorf("expected my-token, got %s", result.Token)
	}
	if result.Region != "" {
		t.Errorf("expected empty region, got %s", result.Region)
	}
	if result.ProjectID != "" {
		t.Errorf("expected empty project_id, got %s", result.ProjectID)
	}
}

func TestResolve_AuthResultFields(t *testing.T) {
	// Verify all fields in AuthResult struct.
	cfg := AuthConfig{
		APIToken:  "tok-123",
		Region:    "ap-south-noi-1",
		ProjectID: "proj-abc-def",
	}

	result, err := Resolve(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token != "tok-123" {
		t.Errorf("expected tok-123, got %s", result.Token)
	}
	if result.Region != "ap-south-noi-1" {
		t.Errorf("expected ap-south-noi-1, got %s", result.Region)
	}
	if result.ProjectID != "proj-abc-def" {
		t.Errorf("expected proj-abc-def, got %s", result.ProjectID)
	}
	if result.Method != "static_token" {
		t.Errorf("expected static_token, got %s", result.Method)
	}
}
