package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLogin_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/auth/login" {
			t.Errorf("expected /auth/login, got %s", r.URL.Path)
		}

		var body loginRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Email != "user@test.com" {
			t.Errorf("expected email user@test.com, got %s", body.Email)
		}

		resp := loginResponse{}
		resp.Error = false
		resp.Message = "SUCCESS"
		resp.Data.AccessToken = "jwt-token-abc"
		resp.Data.ExpiresInSeconds = 172800
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:    ts.URL,
		Email:     "user@test.com",
		Password:  "pass123",
		Region:    "mumbai",
		ProjectID: "proj-1",
	}

	result, err := loginWithCredentials(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token != "jwt-token-abc" {
		t.Errorf("expected token jwt-token-abc, got %s", result.Token)
	}
	if result.Method != "email_password" {
		t.Errorf("expected method email_password, got %s", result.Method)
	}
	if result.Region != "mumbai" {
		t.Errorf("expected region mumbai, got %s", result.Region)
	}
}

func TestLogin_MFAEnabled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := loginResponse{}
		resp.Error = false
		resp.Message = "SUCCESS"
		resp.Data.MFAEnabled = true
		resp.Data.ExpiresInSeconds = 300
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "mfa@test.com",
		Password: "pass",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for MFA-enabled account")
	}
	if !strings.Contains(err.Error(), "2FA enabled") {
		t.Errorf("expected 2FA error message, got: %s", err.Error())
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := loginResponse{}
		resp.Error = true
		resp.Message = "Invalid credentials"
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "wrong@test.com",
		Password: "wrong",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for bad credentials")
	}
	if !strings.Contains(err.Error(), "Invalid credentials") {
		t.Errorf("expected 'Invalid credentials' in error, got: %s", err.Error())
	}
}

func TestLogin_NoToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := loginResponse{}
		resp.Error = false
		resp.Message = "SUCCESS"
		resp.Data.AccessToken = ""
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for empty access token")
	}
	if !strings.Contains(err.Error(), "no access token") {
		t.Errorf("expected 'no access token' error, got: %s", err.Error())
	}
}

func TestLogin_AutoSelectProject(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Manually build the response with projects.
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"error": false,
			"message": "SUCCESS",
			"data": {
				"accessToken": "jwt-auto",
				"expiresInSeconds": 172800,
				"projects": {
					"proj-auto-123": {
						"name": "Default Project",
						"status": "active",
						"regions": {
							"delhi": {}
						}
					}
				}
			}
		}`))
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
		// Region and ProjectID intentionally empty for auto-select.
	}

	result, err := loginWithCredentials(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectID != "proj-auto-123" {
		t.Errorf("expected auto-selected project_id proj-auto-123, got %s", result.ProjectID)
	}
	if result.Region != "delhi" {
		t.Errorf("expected auto-selected region delhi, got %s", result.Region)
	}
}

func TestLogin_AutoSelectProject_CapitalizedStatus(t *testing.T) {
	// The real npc-api returns "Active" (capitalized), not "active".
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"error": false,
			"message": "SUCCESS",
			"data": {
				"accessToken": "jwt-cap",
				"expiresInSeconds": 172800,
				"projects": {
					"proj-cap-456": {
						"name": "Capitalized Project",
						"status": "Active",
						"regions": {
							"ap-south-mum-1": {}
						}
					}
				}
			}
		}`))
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
	}

	result, err := loginWithCredentials(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectID != "proj-cap-456" {
		t.Errorf("expected auto-selected project_id proj-cap-456, got %s", result.ProjectID)
	}
	if result.Region != "ap-south-mum-1" {
		t.Errorf("expected auto-selected region ap-south-mum-1, got %s", result.Region)
	}
}

func TestLogin_ExplicitOverridesAutoSelect(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"error": false,
			"message": "SUCCESS",
			"data": {
				"accessToken": "jwt-explicit",
				"expiresInSeconds": 172800,
				"projects": {
					"proj-other": {
						"name": "Other Project",
						"status": "active",
						"regions": {
							"noida": {}
						}
					}
				}
			}
		}`))
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:    ts.URL,
		Email:     "user@test.com",
		Password:  "pass",
		Region:    "mumbai",
		ProjectID: "proj-explicit",
	}

	result, err := loginWithCredentials(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Explicit values should NOT be overridden by auto-select.
	if result.Region != "mumbai" {
		t.Errorf("expected explicit region mumbai, got %s", result.Region)
	}
	if result.ProjectID != "proj-explicit" {
		t.Errorf("expected explicit project_id proj-explicit, got %s", result.ProjectID)
	}
}

func TestLogin_MissingAPIURL(t *testing.T) {
	cfg := AuthConfig{
		Email:    "user@test.com",
		Password: "pass",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for missing api_url")
	}
	if !strings.Contains(err.Error(), "api_url is required") {
		t.Errorf("expected api_url error, got: %s", err.Error())
	}
}

func TestLogin_NetworkError(t *testing.T) {
	cfg := AuthConfig{
		APIURL:   "http://127.0.0.1:1", // Unreachable port
		Email:    "user@test.com",
		Password: "pass",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for network failure")
	}
	if !strings.Contains(err.Error(), "login request failed") {
		t.Errorf("expected 'login request failed' error, got: %s", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Extended Login Tests
// ---------------------------------------------------------------------------

func TestLogin_UnparseableResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for unparseable response")
	}
	if !strings.Contains(err.Error(), "unexpected login response format") {
		t.Errorf("expected parse error, got: %s", err.Error())
	}
}

func TestLogin_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never respond — let context cancel
		<-r.Context().Done()
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
	}

	_, err := loginWithCredentials(ctx, cfg)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestLogin_AutoSelectProject_InactiveSkipped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"error": false,
			"message": "SUCCESS",
			"data": {
				"accessToken": "jwt-skip",
				"expiresInSeconds": 172800,
				"projects": {
					"proj-inactive": {
						"name": "Inactive",
						"status": "suspended",
						"regions": {"mumbai": {}}
					}
				}
			}
		}`))
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
	}

	result, err := loginWithCredentials(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Inactive project should be skipped — no auto-select.
	if result.ProjectID != "" {
		t.Errorf("expected empty project (inactive skipped), got %s", result.ProjectID)
	}
	if result.Region != "" {
		t.Errorf("expected empty region (inactive skipped), got %s", result.Region)
	}
}

func TestLogin_AutoSelectProject_NoProjects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"error": false,
			"message": "SUCCESS",
			"data": {
				"accessToken": "jwt-noproj",
				"expiresInSeconds": 172800
			}
		}`))
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
	}

	result, err := loginWithCredentials(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token != "jwt-noproj" {
		t.Errorf("expected jwt-noproj, got %s", result.Token)
	}
}

func TestLogin_RequestBodyFormat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body loginRequest
		json.NewDecoder(r.Body).Decode(&body)
		if body.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", body.Email)
		}
		if body.Password != "secret123" {
			t.Errorf("expected password secret123, got %s", body.Password)
		}

		resp := loginResponse{}
		resp.Data.AccessToken = "ok"
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "test@example.com",
		Password: "secret123",
		Region:   "r",
		ProjectID: "p",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogin_HTTP500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(loginResponse{Error: true, Message: "Internal error"})
	}))
	defer ts.Close()

	cfg := AuthConfig{
		APIURL:   ts.URL,
		Email:    "user@test.com",
		Password: "pass",
	}

	_, err := loginWithCredentials(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !strings.Contains(err.Error(), "Internal error") {
		t.Errorf("expected 'Internal error' in message, got: %s", err.Error())
	}
}
