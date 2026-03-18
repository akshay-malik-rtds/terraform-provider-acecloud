package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://api.acecloud.ai", "test-token", "mumbai", "proj-123")

	if c.BaseURL != "https://api.acecloud.ai" {
		t.Errorf("expected BaseURL https://api.acecloud.ai, got %s", c.BaseURL)
	}
	if c.APIToken != "test-token" {
		t.Errorf("expected APIToken test-token, got %s", c.APIToken)
	}
	if c.Region != "mumbai" {
		t.Errorf("expected Region mumbai, got %s", c.Region)
	}
	if c.ProjectID != "proj-123" {
		t.Errorf("expected ProjectID proj-123, got %s", c.ProjectID)
	}
	if c.HTTPClient == nil {
		t.Error("expected HTTPClient to be initialized")
	}
}

func TestBuildURL(t *testing.T) {
	c := NewClient("https://api.acecloud.ai", "tok", "mumbai", "proj-1")

	url := c.buildURL("/cloud/vpcs", nil)
	if url == "" {
		t.Fatal("expected non-empty URL")
	}

	// Should contain region and project_id query params.
	if got := url; got == "" {
		t.Fatal("empty url")
	}

	// With extra params.
	url = c.buildURL("/cloud/vpcs", map[string]string{"page": "1"})
	if url == "" {
		t.Fatal("expected non-empty URL with extra params")
	}
}

func TestGet_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected Bearer test-token, got %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("region") != "mumbai" {
			t.Errorf("expected region=mumbai, got %s", r.URL.Query().Get("region"))
		}
		if r.URL.Query().Get("project_id") != "proj-1" {
			t.Errorf("expected project_id=proj-1, got %s", r.URL.Query().Get("project_id"))
		}

		resp := APIResponse{
			Error:   false,
			Message: "success",
			Data:    json.RawMessage(`{"id":"vpc-123","name":"test-vpc"}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	apiResp, err := c.Get(context.Background(), "/cloud/vpcs/vpc-123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if apiResp.Error {
		t.Error("expected no error in response")
	}

	var data map[string]string
	json.Unmarshal(apiResp.Data, &data)
	if data["id"] != "vpc-123" {
		t.Errorf("expected id vpc-123, got %s", data["id"])
	}
}

func TestPost_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test-vpc" {
			t.Errorf("expected name test-vpc, got %s", body["name"])
		}

		resp := APIResponse{
			Error:   false,
			Message: "created",
			Data:    json.RawMessage(`{"id":"vpc-new","name":"test-vpc"}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	apiResp, err := c.Post(context.Background(), "/cloud/vpcs", map[string]string{"name": "test-vpc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]string
	json.Unmarshal(apiResp.Data, &data)
	if data["id"] != "vpc-new" {
		t.Errorf("expected id vpc-new, got %s", data["id"])
	}
}

func TestPut_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		resp := APIResponse{
			Error:   false,
			Message: "updated",
			Data:    json.RawMessage(`{"id":"vpc-123","name":"updated-vpc"}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	apiResp, err := c.Put(context.Background(), "/cloud/vpcs/vpc-123", map[string]string{"name": "updated-vpc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]string
	json.Unmarshal(apiResp.Data, &data)
	if data["name"] != "updated-vpc" {
		t.Errorf("expected name updated-vpc, got %s", data["name"])
	}
}

func TestDelete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		var body map[string][]string
		json.NewDecoder(r.Body).Decode(&body)
		if len(body["values"]) != 1 || body["values"][0] != "vpc-123" {
			t.Errorf("expected values [vpc-123], got %v", body["values"])
		}

		resp := APIResponse{
			Error:   false,
			Message: "deleted",
			Data:    json.RawMessage(`null`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.Delete(context.Background(), "/cloud/vpcs", map[string][]string{"values": {"vpc-123"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"error":   true,
			"message": "VPC not found",
			"data":    nil,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.Get(context.Background(), "/cloud/vpcs/missing", nil)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
	if err.Error() != "API error: VPC not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestAuth401(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]interface{}{
			"error":   true,
			"message": "Token expired",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "bad-token", "mumbai", "proj-1")
	_, err := c.Get(context.Background(), "/cloud/vpcs", nil)
	if err == nil {
		t.Fatal("expected error on 401")
	}
	if err.Error() != "authentication failed (401): Token expired" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestGetData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Error:   false,
			Message: "ok",
			Data:    json.RawMessage(`[{"id":"f1","name":"small","ram":1024,"vcpus":1,"disk":20}]`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")

	var flavors []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		RAM   int64  `json:"ram"`
		VCPUs int64  `json:"vcpus"`
		Disk  int64  `json:"disk"`
	}

	err := c.GetData(context.Background(), "/cloud/flavors", &flavors)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(flavors) != 1 {
		t.Fatalf("expected 1 flavor, got %d", len(flavors))
	}
	if flavors[0].Name != "small" {
		t.Errorf("expected flavor name small, got %s", flavors[0].Name)
	}
}

func TestPostData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := APIResponse{
			Error:   false,
			Message: "created",
			Data:    json.RawMessage(`{"id":"vol-new","name":"data-vol","size":100}`),
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")

	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Size int64  `json:"size"`
	}

	err := c.PostData(context.Background(), "/cloud/volumes", map[string]interface{}{"name": "data-vol", "size": 100}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "vol-new" {
		t.Errorf("expected id vol-new, got %s", result.ID)
	}
}

func TestInvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.Get(context.Background(), "/cloud/vpcs", nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidateToken_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/me" {
			t.Errorf("expected /auth/me, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer valid-token" {
			t.Errorf("expected Bearer valid-token, got %s", r.Header.Get("Authorization"))
		}
		// Verify no region/project_id query params on auth endpoint.
		if r.URL.Query().Get("region") != "" {
			t.Error("expected no region query param on /auth/me")
		}
		if r.URL.Query().Get("project_id") != "" {
			t.Error("expected no project_id query param on /auth/me")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"error":false,"data":{"id":"user-1","email":"user@test.com"}}`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "valid-token", "mumbai", "proj-1")
	err := c.ValidateToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateToken_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":true,"message":"Token expired"}`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "expired-token", "mumbai", "proj-1")
	err := c.ValidateToken(context.Background())
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if err.Error() != "token is invalid or expired (HTTP 401)" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

func TestValidateToken_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":true,"message":"Internal server error"}`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "some-token", "mumbai", "proj-1")
	err := c.ValidateToken(context.Background())
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if err.Error() != "token validation failed (HTTP 500)" {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

// ═══════════════════════════════════════════════════════════════
// Error Sanitization Tests
// ═══════════════════════════════════════════════════════════════

func TestSanitizeErrorMessage_Empty(t *testing.T) {
	if got := sanitizeErrorMessage(""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSanitizeErrorMessage_NoChange(t *testing.T) {
	input := "Instance not found"
	if got := sanitizeErrorMessage(input); got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestSanitizeErrorMessage_OpenStackServiceNames(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Nova reference",
			input: "Nova compute error: instance not found",
			want:  "backend compute error: instance not found",
		},
		{
			name:  "Neutron reference",
			input: "Neutron returned error for port binding",
			want:  "backend returned error for port binding",
		},
		{
			name:  "Cinder reference",
			input: "Cinder volume creation failed",
			want:  "backend volume creation failed",
		},
		{
			name:  "Octavia reference",
			input: "Octavia loadbalancer is immutable",
			want:  "backend loadbalancer is immutable",
		},
		{
			name:  "Glance reference",
			input: "Glance image not found",
			want:  "backend image not found",
		},
		{
			name:  "Keystone reference",
			input: "Keystone token validation failed",
			want:  "backend token validation failed",
		},
		{
			name:  "OpenStack reference",
			input: "OpenStack API returned 503",
			want:  "backend API returned 503",
		},
		{
			name:  "Multiple services",
			input: "Nova called Neutron but Cinder was down",
			want:  "backend called backend but backend was down",
		},
		{
			name:  "Case insensitive",
			input: "NOVA compute error and neutron port error",
			want:  "backend compute error and backend port error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeErrorMessage(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeErrorMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeErrorMessage_KnownReplacements(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "OpenStack auth failed",
			input: "OpenStack authentication failed",
			want:  "Authentication failed",
		},
		{
			name:  "Keystone unavailable",
			input: "Keystone service is temporarily unavailable",
			want:  "Authentication service temporarily unavailable",
		},
		{
			name:  "OpenStack unavailable",
			input: "OpenStack service is temporarily unavailable",
			want:  "Cloud service temporarily unavailable",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeErrorMessage(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeErrorMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeErrorMessage_ErrorStructPrefixes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "NeutronError prefix",
			input: "NeutronError: Port not found",
			want:  "Port not found",
		},
		{
			name:  "computeFault prefix",
			input: "computeFault: Instance quota exceeded",
			want:  "Instance quota exceeded",
		},
		{
			name:  "cinderException prefix",
			input: "cinderException: Volume size exceeds limit",
			want:  "Volume size exceeds limit",
		},
		{
			name:  "itemNotFound prefix",
			input: "itemNotFound: Resource 123 does not exist",
			want:  "Resource 123 does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeErrorMessage(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeErrorMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeErrorMessage_FaultFields(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "faultcode field",
			input: `Error occurred faultcode: "Server" in processing`,
			want:  "Error occurred in processing",
		},
		{
			name:  "faultstring field",
			input: `faultstring: "Connection refused" please retry`,
			want:  "please retry",
		},
		{
			name:  "debuginfo field",
			input: `Internal error debuginfo: "Traceback..."`,
			want:  "Internal error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeErrorMessage(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeErrorMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeErrorMessage_InternalIDs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "request-id",
			input: "Error occurred request-id: req-abc-123-def resource not found",
			want:  "Error occurred resource not found",
		},
		{
			name:  "x-openstack-request-id",
			input: "Failed x-openstack-request-id: req-uuid-here please retry",
			want:  "Failed please retry",
		},
		{
			name:  "transaction_id",
			input: "Error transaction_id: txn-456 volume busy",
			want:  "Error volume busy",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeErrorMessage(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeErrorMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeErrorMessage_CollapseSpaces(t *testing.T) {
	input := "Error    with   extra    spaces"
	want := "Error with extra spaces"
	got := sanitizeErrorMessage(input)
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestSanitizeErrorMessage_ComplexRealWorld(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Neutron port error with request ID",
			input: "NeutronError: Port abc-123 not found request-id: req-xyz-789",
			want:  "Port abc-123 not found",
		},
		{
			name:  "Nova instance error with fault fields",
			input: `Nova error: Instance failed faultcode: "Server" faultstring: "No valid host" debuginfo: "Traceback..."`,
			want:  "backend error: Instance failed",
		},
		{
			name:  "Cinder volume with OpenStack reference",
			input: "cinderException: OpenStack Cinder volume limit exceeded for project",
			want:  "backend backend volume limit exceeded for project",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeErrorMessage(tc.input)
			if got != tc.want {
				t.Errorf("sanitizeErrorMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════════
// Error Sanitization Integration Tests (via mock HTTP server)
// ═══════════════════════════════════════════════════════════════

func TestAPIError_Sanitized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"error":   true,
			"message": "NeutronError: Port abc-123 not found request-id: req-xyz-789",
			"data":    nil,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.Get(context.Background(), "/cloud/security-groups/bad-id", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	errMsg := err.Error()
	// Should NOT contain "NeutronError" or "request-id"
	if strings.Contains(errMsg, "NeutronError") {
		t.Errorf("error message should not contain 'NeutronError': %s", errMsg)
	}
	if strings.Contains(errMsg, "request-id") {
		t.Errorf("error message should not contain 'request-id': %s", errMsg)
	}
	// Should contain the useful part
	if !strings.Contains(errMsg, "Port abc-123 not found") {
		t.Errorf("error message should contain 'Port abc-123 not found': %s", errMsg)
	}
}

func TestAuth401_Sanitized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := map[string]interface{}{
			"error":   true,
			"message": "OpenStack authentication failed: Keystone returned 401",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "bad-token", "mumbai", "proj-1")
	_, err := c.Get(context.Background(), "/cloud/vpcs", nil)
	if err == nil {
		t.Fatal("expected error on 401")
	}
	errMsg := err.Error()
	// Should NOT contain "OpenStack" or "Keystone"
	if strings.Contains(errMsg, "OpenStack") {
		t.Errorf("error should not contain 'OpenStack': %s", errMsg)
	}
	if strings.Contains(errMsg, "Keystone") {
		t.Errorf("error should not contain 'Keystone': %s", errMsg)
	}
	// Should contain sanitized message
	if !strings.Contains(errMsg, "Authentication failed") {
		t.Errorf("error should contain 'Authentication failed': %s", errMsg)
	}
}

func TestParseFailure_NoRawBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html>Internal Server Error</html>"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.Get(context.Background(), "/cloud/vpcs", nil)
	if err == nil {
		t.Fatal("expected error for non-JSON response")
	}
	errMsg := err.Error()
	// Should NOT contain raw HTML body
	if strings.Contains(errMsg, "<html>") {
		t.Errorf("error should not contain raw HTML body: %s", errMsg)
	}
	if strings.Contains(errMsg, "Internal Server Error") {
		t.Errorf("error should not contain raw body text: %s", errMsg)
	}
	// Should contain generic parse failure message
	if !strings.Contains(errMsg, "failed to parse API response") {
		t.Errorf("error should contain 'failed to parse API response': %s", errMsg)
	}
}

func TestRetryableError_Sanitized(t *testing.T) {
	attempt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		w.WriteHeader(http.StatusBadGateway)
		resp := map[string]interface{}{
			"error":   true,
			"message": "Neutron service temporarily unavailable, x-openstack-request-id: req-abcdef",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.Get(context.Background(), "/cloud/vpcs", nil)
	if err == nil {
		t.Fatal("expected error after retries")
	}
	errMsg := err.Error()
	// Should NOT contain "Neutron" or "x-openstack-request-id"
	if strings.Contains(errMsg, "Neutron") {
		t.Errorf("error should not contain 'Neutron': %s", errMsg)
	}
	if strings.Contains(errMsg, "x-openstack-request-id") {
		t.Errorf("error should not contain 'x-openstack-request-id': %s", errMsg)
	}
	// Should have retried (3 attempts)
	if attempt != maxRetries {
		t.Errorf("expected %d retry attempts, got %d", maxRetries, attempt)
	}
}

// ═══════════════════════════════════════════════════════════════
// parseStatus Tests
// ═══════════════════════════════════════════════════════════════

func TestParseStatus_IntegerStatus(t *testing.T) {
	r := &APIResponse{RawStatus: json.RawMessage(`400`)}
	r.parseStatus()
	if r.Status != 400 {
		t.Fatalf("expected 400, got %d", r.Status)
	}
}

func TestParseStatus_StringStatus(t *testing.T) {
	r := &APIResponse{RawStatus: json.RawMessage(`"Cluster creation in process"`)}
	r.parseStatus()
	if r.Status != 0 {
		t.Fatalf("expected 0 for string status, got %d", r.Status)
	}
}

func TestParseStatus_NilRawStatus(t *testing.T) {
	r := &APIResponse{RawStatus: nil}
	r.parseStatus()
	if r.Status != 0 {
		t.Fatalf("expected 0 for nil status, got %d", r.Status)
	}
}

func TestParseStatus_ZeroStatus(t *testing.T) {
	r := &APIResponse{RawStatus: json.RawMessage(`0`)}
	r.parseStatus()
	if r.Status != 0 {
		t.Fatalf("expected 0, got %d", r.Status)
	}
}

func TestParseStatus_200Status(t *testing.T) {
	r := &APIResponse{RawStatus: json.RawMessage(`200`)}
	r.parseStatus()
	if r.Status != 200 {
		t.Fatalf("expected 200, got %d", r.Status)
	}
}

func TestParseStatus_BooleanStatus(t *testing.T) {
	r := &APIResponse{RawStatus: json.RawMessage(`true`)}
	r.parseStatus()
	if r.Status != 0 {
		t.Fatalf("expected 0 for boolean status, got %d", r.Status)
	}
}

// ═══════════════════════════════════════════════════════════════
// parseMessage Tests
// ═══════════════════════════════════════════════════════════════

func TestParseMessage_StringMessage(t *testing.T) {
	r := &APIResponse{RawMessage: json.RawMessage(`"Resource created successfully"`)}
	r.parseMessage()
	if r.Message != "Resource created successfully" {
		t.Fatalf("expected 'Resource created successfully', got %q", r.Message)
	}
}

func TestParseMessage_ObjectMessage(t *testing.T) {
	r := &APIResponse{RawMessage: json.RawMessage(`{"field":"value is invalid"}`)}
	r.parseMessage()
	// Falls back to raw JSON string
	if r.Message != `{"field":"value is invalid"}` {
		t.Fatalf("expected raw JSON fallback, got %q", r.Message)
	}
}

func TestParseMessage_NilMessage(t *testing.T) {
	r := &APIResponse{RawMessage: nil, RawMessages: nil}
	r.parseMessage()
	if r.Message != "" {
		t.Fatalf("expected empty message for nil, got %q", r.Message)
	}
}

func TestParseMessage_MessagesPlural(t *testing.T) {
	r := &APIResponse{
		RawMessage:  nil,
		RawMessages: json.RawMessage(`{"command":["must be an array"]}`),
	}
	r.parseMessage()
	if r.Message != "command: must be an array" {
		t.Fatalf("expected 'command: must be an array', got %q", r.Message)
	}
}

func TestParseMessage_MessageTakesPriority(t *testing.T) {
	r := &APIResponse{
		RawMessage:  json.RawMessage(`"primary message"`),
		RawMessages: json.RawMessage(`{"field":["ignored"]}`),
	}
	r.parseMessage()
	if r.Message != "primary message" {
		t.Fatalf("expected 'primary message', got %q", r.Message)
	}
}

// ═══════════════════════════════════════════════════════════════
// flattenValidationMessages Tests
// ═══════════════════════════════════════════════════════════════

func TestFlattenValidation_StringArray(t *testing.T) {
	raw := json.RawMessage(`{"name":["must not be empty","must be at least 3 characters"]}`)
	got := flattenValidationMessages(raw)
	if !strings.Contains(got, "name: must not be empty") {
		t.Fatalf("expected 'name: must not be empty' in %q", got)
	}
	if !strings.Contains(got, "name: must be at least 3 characters") {
		t.Fatalf("expected 'name: must be at least 3 characters' in %q", got)
	}
}

func TestFlattenValidation_NestedObject(t *testing.T) {
	raw := json.RawMessage(`{"networking":{"xForwardedFor":["must be boolean"]}}`)
	got := flattenValidationMessages(raw)
	if !strings.Contains(got, "networking.xForwardedFor: must be boolean") {
		t.Fatalf("expected nested message in %q", got)
	}
}

func TestFlattenValidation_Mixed(t *testing.T) {
	raw := json.RawMessage(`{"command":["must be an array"],"networking":{"xForwardedFor":["must be boolean"]}}`)
	got := flattenValidationMessages(raw)
	if !strings.Contains(got, "command: must be an array") {
		t.Fatalf("expected command message in %q", got)
	}
	if !strings.Contains(got, "networking.xForwardedFor: must be boolean") {
		t.Fatalf("expected nested message in %q", got)
	}
}

func TestFlattenValidation_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`not json`)
	got := flattenValidationMessages(raw)
	if got != "not json" {
		t.Fatalf("expected raw fallback 'not json', got %q", got)
	}
}

func TestFlattenValidation_EmptyObject(t *testing.T) {
	raw := json.RawMessage(`{}`)
	got := flattenValidationMessages(raw)
	if got != "{}" {
		t.Fatalf("expected raw fallback for empty object, got %q", got)
	}
}

func TestFlattenValidation_NestedNonArrayValue(t *testing.T) {
	raw := json.RawMessage(`{"field":{"sub":"not an array"}}`)
	got := flattenValidationMessages(raw)
	if !strings.Contains(got, "field.sub:") {
		t.Fatalf("expected 'field.sub:' in %q", got)
	}
}

// ═══════════════════════════════════════════════════════════════
// buildURL Content Verification Tests
// ═══════════════════════════════════════════════════════════════

func TestBuildURL_ContainsRegionAndProject(t *testing.T) {
	c := NewClient("https://api.acecloud.ai", "tok", "ap-south-noi-1", "proj-abc")
	url := c.buildURL("/cloud/vpcs", nil)
	if !strings.Contains(url, "region=ap-south-noi-1") {
		t.Errorf("URL missing region param: %s", url)
	}
	if !strings.Contains(url, "project_id=proj-abc") {
		t.Errorf("URL missing project_id param: %s", url)
	}
	if !strings.HasPrefix(url, "https://api.acecloud.ai/cloud/vpcs?") {
		t.Errorf("URL has wrong prefix: %s", url)
	}
}

func TestBuildURL_WithExtraParams(t *testing.T) {
	c := NewClient("https://api.acecloud.ai", "tok", "mumbai", "proj-1")
	url := c.buildURL("/cloud/loadbalancers", map[string]string{"cascade": "true"})
	if !strings.Contains(url, "cascade=true") {
		t.Errorf("URL missing cascade param: %s", url)
	}
	if !strings.Contains(url, "region=mumbai") {
		t.Errorf("URL missing region param: %s", url)
	}
}

func TestBuildURL_PathWithTrailingSlash(t *testing.T) {
	c := NewClient("https://api.acecloud.ai", "tok", "mumbai", "proj-1")
	url := c.buildURL("/cloud/vpcs/vpc-123", nil)
	if !strings.Contains(url, "/cloud/vpcs/vpc-123") {
		t.Errorf("URL should contain path: %s", url)
	}
}

// ═══════════════════════════════════════════════════════════════
// retryableStatus Tests
// ═══════════════════════════════════════════════════════════════

func TestRetryableStatus(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{201, false},
		{204, false},
		{400, true},
		{401, false},
		{403, false},
		{404, false},
		{409, false},
		{500, false},
		{502, true},
		{503, true},
		{504, true},
	}
	for _, tc := range tests {
		got := retryableStatus(tc.code)
		if got != tc.want {
			t.Errorf("retryableStatus(%d) = %v, want %v", tc.code, got, tc.want)
		}
	}
}

// ═══════════════════════════════════════════════════════════════
// DeleteWithParams / PutWithParams Tests
// ═══════════════════════════════════════════════════════════════

func TestDeleteWithParams_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Query().Get("cascade") != "true" {
			t.Errorf("expected cascade=true, got %s", r.URL.Query().Get("cascade"))
		}
		resp := APIResponse{Data: json.RawMessage(`null`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.DeleteWithParams(context.Background(), "/cloud/loadbalancers", map[string]string{"key": "id"}, map[string]string{"cascade": "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPutWithParams_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Query().Get("action") != "associate" {
			t.Errorf("expected action=associate, got %s", r.URL.Query().Get("action"))
		}
		resp := APIResponse{Data: json.RawMessage(`{"id":"fip-123"}`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	apiResp, err := c.PutWithParams(context.Background(), "/cloud/floating-ips/fip-123", map[string]string{"port_id": "port-1"}, map[string]string{"action": "associate"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiResp == nil {
		t.Fatal("expected non-nil response")
	}
}

// ═══════════════════════════════════════════════════════════════
// Delete with nil body (no Content-Type header)
// ═══════════════════════════════════════════════════════════════

func TestDelete_NilBody_NoContentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "" {
			t.Errorf("expected no Content-Type for nil body DELETE, got %s", r.Header.Get("Content-Type"))
		}
		resp := APIResponse{Data: json.RawMessage(`null`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	_, err := c.Delete(context.Background(), "/cloud/secrets/my-secret", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ═══════════════════════════════════════════════════════════════
// DoRequest Retry Behavior Tests
// ═══════════════════════════════════════════════════════════════

func TestDoRequest_RetriesOn502(t *testing.T) {
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": true, "message": "Bad gateway"})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"error": false, "data": map[string]string{"id": "ok"}})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	resp, err := c.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestDoRequest_NoRetryOn404(t *testing.T) {
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		json.NewEncoder(w).Encode(map[string]interface{}{"error": true, "message": "Not found"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt (no retry for non-retryable), got %d", attempts)
	}
}

func TestDoRequest_NoRetryOn401(t *testing.T) {
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": true, "message": "Unauthorized"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected 401 error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt (no retry for 401), got %d", attempts)
	}
}

func TestDoRequest_UnparseableRetryableResponse(t *testing.T) {
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error after retries")
	}
	if attempts != maxRetries {
		t.Fatalf("expected %d attempts, got %d", maxRetries, attempts)
	}
}

func TestDoRequest_401WithUnparseableBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected 401 error")
	}
	if err.Error() != "authentication failed (401)" {
		t.Fatalf("expected generic 401 message, got: %s", err.Error())
	}
}

// ═══════════════════════════════════════════════════════════════
// Integration: parseStatus in DoRequest flow
// ═══════════════════════════════════════════════════════════════

func TestDoRequest_ParsesIntegerStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":false,"status":200,"message":"OK","data":{"id":"1"}}`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	resp, err := c.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != 200 {
		t.Fatalf("expected parsed status 200, got %d", resp.Status)
	}
}

func TestDoRequest_ParsesStringStatusAsZero(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":false,"status":"Cluster creation in process","data":{}}`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	resp, err := c.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != 0 {
		t.Fatalf("expected status 0 for string status, got %d", resp.Status)
	}
}

// ═══════════════════════════════════════════════════════════════
// Integration: parseMessage with CaaS-style messages
// ═══════════════════════════════════════════════════════════════

func TestDoRequest_CaaSValidationMessages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":true,"messages":{"command":["must be an array"],"networking":{"xForwardedFor":["must be boolean"]}}}`))
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "command: must be an array") {
		t.Errorf("expected 'command: must be an array' in error: %s", errMsg)
	}
}

// ═══════════════════════════════════════════════════════════════
// GetData / PostData Error Paths
// ═══════════════════════════════════════════════════════════════

func TestGetData_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": true, "message": "not found"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	var result struct{}
	err := c.GetData(context.Background(), "/test", &result)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %s", err.Error())
	}
}

func TestPostData_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{"error": true, "message": "validation failed"})
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "tok", "r", "p")
	var result struct{}
	err := c.PostData(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("expected 'validation failed' in error, got: %s", err.Error())
	}
}
