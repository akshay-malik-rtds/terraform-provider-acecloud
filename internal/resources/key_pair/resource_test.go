package key_pair

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestKeyPairModel(t *testing.T) {
	// Verify the model struct has the expected fields.
	model := keyPairModel{}
	_ = model.ID
	_ = model.Name
	_ = model.PublicKey
	_ = model.Fingerprint
}

func TestKeyPairCreateBody_WithPublicKey(t *testing.T) {
	body := map[string]interface{}{
		"name":       "my-keypair",
		"public_key": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB user@host",
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create body: %v", err)
	}

	if parsed["name"] != "my-keypair" {
		t.Errorf("expected name 'my-keypair', got %v", parsed["name"])
	}
	if parsed["public_key"] != "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB user@host" {
		t.Errorf("expected public_key to match, got %v", parsed["public_key"])
	}
}

func TestKeyPairCreateBody_WithoutPublicKey(t *testing.T) {
	// When no public key is provided, the body contains only name.
	body := map[string]interface{}{
		"name": "auto-generated-kp",
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create body: %v", err)
	}

	if parsed["name"] != "auto-generated-kp" {
		t.Errorf("expected name 'auto-generated-kp', got %v", parsed["name"])
	}
	if _, exists := parsed["public_key"]; exists {
		t.Error("expected public_key to be absent when not provided")
	}
}

func TestKeyPairDeleteBody(t *testing.T) {
	body := map[string]interface{}{
		"key":    "id",
		"values": []string{"kp-1"},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal delete body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal delete body: %v", err)
	}

	if parsed["key"] != "id" {
		t.Errorf("expected key 'id', got %v", parsed["key"])
	}

	values, ok := parsed["values"].([]interface{})
	if !ok {
		t.Fatal("expected values to be an array")
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0] != "kp-1" {
		t.Errorf("expected first value 'kp-1', got %v", values[0])
	}
}

// --- Metadata tests ---

func TestMetadata(t *testing.T) {
	r := &keyPairResource{}
	req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "acecloud_key_pair" {
		t.Errorf("expected type name 'acecloud_key_pair', got %s", resp.TypeName)
	}
}

func TestMetadata_EmptyProvider(t *testing.T) {
	r := &keyPairResource{}
	req := resource.MetadataRequest{ProviderTypeName: ""}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "_key_pair" {
		t.Errorf("expected type name '_key_pair', got %s", resp.TypeName)
	}
}

// --- Schema tests ---

func TestSchema_Attributes(t *testing.T) {
	r := &keyPairResource{}
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)

	attrs := resp.Schema.Attributes
	expectedAttrs := []string{"id", "name", "public_key", "fingerprint"}
	for _, name := range expectedAttrs {
		if _, ok := attrs[name]; !ok {
			t.Errorf("expected attribute '%s' in schema", name)
		}
	}

	if len(attrs) != len(expectedAttrs) {
		t.Errorf("expected %d attributes, got %d", len(expectedAttrs), len(attrs))
	}
}

func TestSchema_RequiredOptionalComputed(t *testing.T) {
	r := &keyPairResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	attrs := resp.Schema.Attributes

	// name is Required
	nameAttr := attrs["name"]
	if !nameAttr.IsRequired() {
		t.Error("expected 'name' to be required")
	}

	// public_key is Optional
	pkAttr := attrs["public_key"]
	if !pkAttr.IsOptional() {
		t.Error("expected 'public_key' to be optional")
	}

	// id is Computed
	idAttr := attrs["id"]
	if !idAttr.IsComputed() {
		t.Error("expected 'id' to be computed")
	}

	// fingerprint is Computed
	fpAttr := attrs["fingerprint"]
	if !fpAttr.IsComputed() {
		t.Error("expected 'fingerprint' to be computed")
	}
}

func TestSchema_Description(t *testing.T) {
	r := &keyPairResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)
	if resp.Schema.Description == "" {
		t.Error("expected schema description to be non-empty")
	}
}

// --- Configure tests ---

func TestConfigure_NilProviderData(t *testing.T) {
	r := &keyPairResource{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("expected no error when ProviderData is nil")
	}
	if r.client != nil {
		t.Error("expected client to remain nil")
	}
}

func TestConfigure_WrongType(t *testing.T) {
	r := &keyPairResource{}
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestConfigure_ValidClient(t *testing.T) {
	r := &keyPairResource{}
	c := client.NewClient("http://localhost", "token", "mumbai", "proj-1")
	req := resource.ConfigureRequest{ProviderData: c}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error, got: %v", resp.Diagnostics)
	}
	if r.client != c {
		t.Error("expected client to be set")
	}
}

// --- API path and constant tests ---

func TestAPIPath(t *testing.T) {
	if apiPath != "/cloud/key-pairs" {
		t.Errorf("expected apiPath '/cloud/key-pairs', got %s", apiPath)
	}
}

// --- Create response parsing tests ---

func TestParseCreateResponse_ValidWithFingerprint(t *testing.T) {
	data := json.RawMessage(`{"id": "kp-abc-123", "fingerprint": "ab:cd:ef:12:34:56"}`)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	plan := keyPairModel{
		Name:      types.StringValue("test-kp"),
		PublicKey: types.StringValue("ssh-rsa AAAA..."),
	}

	id, ok := result["id"].(string)
	if !ok {
		t.Fatal("expected id as string")
	}
	plan.ID = types.StringValue(id)

	if v, ok := result["fingerprint"].(string); ok {
		plan.Fingerprint = types.StringValue(v)
	} else {
		plan.Fingerprint = types.StringNull()
	}

	if plan.ID.ValueString() != "kp-abc-123" {
		t.Errorf("expected ID 'kp-abc-123', got %s", plan.ID.ValueString())
	}
	if plan.Fingerprint.ValueString() != "ab:cd:ef:12:34:56" {
		t.Errorf("expected fingerprint 'ab:cd:ef:12:34:56', got %s", plan.Fingerprint.ValueString())
	}
}

func TestParseCreateResponse_MissingFingerprint(t *testing.T) {
	data := json.RawMessage(`{"id": "kp-xyz-789"}`)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	plan := keyPairModel{}
	if _, ok := result["fingerprint"].(string); !ok {
		plan.Fingerprint = types.StringNull()
	}
	if !plan.Fingerprint.IsNull() {
		t.Error("expected fingerprint to be null when missing from response")
	}
}

func TestParseCreateResponse_MissingID(t *testing.T) {
	data := json.RawMessage(`{"fingerprint": "ab:cd:ef"}`)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, ok := result["id"].(string)
	if ok {
		t.Error("expected id to not be found when missing from response")
	}
}

func TestParseCreateResponse_EmptyData(t *testing.T) {
	data := json.RawMessage(`{}`)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, ok := result["id"].(string)
	if ok {
		t.Error("expected id to be absent in empty data")
	}
}

func TestParseCreateResponse_NonStringID(t *testing.T) {
	// If the API returns a numeric ID, the type assertion should fail
	data := json.RawMessage(`{"id": 12345}`)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, ok := result["id"].(string)
	if ok {
		t.Error("expected string type assertion to fail for numeric id")
	}
}

// --- Read response parsing tests ---

func TestParseReadResponse_AllFields(t *testing.T) {
	data := json.RawMessage(`{
		"name": "prod-key",
		"public_key": "ssh-rsa AAAA...",
		"fingerprint": "11:22:33:44:55"
	}`)

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	state := keyPairModel{
		ID:        types.StringValue("kp-1"),
		Name:      types.StringValue("old-name"),
		PublicKey: types.StringValue("ssh-rsa OLD..."),
	}

	if v, ok := result["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := result["public_key"].(string); ok && v != "" {
		state.PublicKey = types.StringValue(v)
	}
	if v, ok := result["fingerprint"].(string); ok {
		state.Fingerprint = types.StringValue(v)
	}

	if state.Name.ValueString() != "prod-key" {
		t.Errorf("expected name 'prod-key', got %s", state.Name.ValueString())
	}
	if state.PublicKey.ValueString() != "ssh-rsa AAAA..." {
		t.Errorf("expected public_key 'ssh-rsa AAAA...', got %s", state.PublicKey.ValueString())
	}
	if state.Fingerprint.ValueString() != "11:22:33:44:55" {
		t.Errorf("expected fingerprint '11:22:33:44:55', got %s", state.Fingerprint.ValueString())
	}
}

func TestParseReadResponse_NoPublicKey(t *testing.T) {
	// API often does not return public_key on read - state should preserve the previous value
	data := json.RawMessage(`{
		"name": "prod-key",
		"fingerprint": "11:22:33:44:55"
	}`)

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	state := keyPairModel{
		ID:        types.StringValue("kp-1"),
		Name:      types.StringValue("old-name"),
		PublicKey: types.StringValue("ssh-rsa ORIGINAL..."),
	}

	if v, ok := result["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := result["public_key"].(string); ok && v != "" {
		state.PublicKey = types.StringValue(v)
	}

	// PublicKey should retain previous value
	if state.PublicKey.ValueString() != "ssh-rsa ORIGINAL..." {
		t.Errorf("expected public_key preserved as 'ssh-rsa ORIGINAL...', got %s", state.PublicKey.ValueString())
	}
}

func TestParseReadResponse_EmptyPublicKey(t *testing.T) {
	// If API returns empty string for public_key, state should preserve existing value
	data := json.RawMessage(`{
		"name": "prod-key",
		"public_key": "",
		"fingerprint": "11:22:33:44:55"
	}`)

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	state := keyPairModel{
		PublicKey: types.StringValue("ssh-rsa KEEP-ME..."),
	}

	if v, ok := result["public_key"].(string); ok && v != "" {
		state.PublicKey = types.StringValue(v)
	}

	if state.PublicKey.ValueString() != "ssh-rsa KEEP-ME..." {
		t.Errorf("expected public_key preserved as 'ssh-rsa KEEP-ME...', got %s", state.PublicKey.ValueString())
	}
}

// --- Delete body with multiple IDs ---

func TestKeyPairDeleteBody_MultipleIDs(t *testing.T) {
	body := map[string]interface{}{
		"key":    "id",
		"values": []string{"kp-1", "kp-2", "kp-3"},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	values, ok := parsed["values"].([]interface{})
	if !ok {
		t.Fatal("expected values to be an array")
	}
	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}
}

// --- Create body edge cases ---

func TestCreateBody_EmptyPublicKey(t *testing.T) {
	// Simulates condition where public_key is provided but empty
	body := map[string]interface{}{
		"name":       "empty-pk-test",
		"public_key": "",
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["public_key"] != "" {
		t.Errorf("expected empty public_key, got %v", parsed["public_key"])
	}
}

func TestCreateBody_SpecialCharactersInName(t *testing.T) {
	names := []string{
		"my-key-pair",
		"my_key_pair",
		"my key pair",
		"key-pair-123",
		"KEY_PAIR_ABC",
	}

	for _, name := range names {
		body := map[string]interface{}{"name": name}
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal for name '%s': %v", name, err)
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("failed to unmarshal for name '%s': %v", name, err)
		}
		if parsed["name"] != name {
			t.Errorf("name mismatch: expected '%s', got '%v'", name, parsed["name"])
		}
	}
}

// --- HTTP integration tests using httptest ---

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *client.Client) {
	ts := httptest.NewServer(handler)
	c := client.NewClient(ts.URL, "test-token", "mumbai", "proj-123")
	return ts, c
}

func TestCreateViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/cloud/key-pairs" {
			t.Errorf("expected path /cloud/key-pairs, got %s", r.URL.Path)
		}
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected auth header 'Bearer test-token', got %s", auth)
		}
		// Verify Content-Type
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"id":"kp-new-123","fingerprint":"aa:bb:cc:dd"}}`)
	})
	defer ts.Close()

	apiResp, err := c.Post(context.Background(), apiPath, map[string]interface{}{
		"name":       "test-kp",
		"public_key": "ssh-rsa AAAA...",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to parse response data: %v", err)
	}
	if result["id"] != "kp-new-123" {
		t.Errorf("expected id 'kp-new-123', got %v", result["id"])
	}
	if result["fingerprint"] != "aa:bb:cc:dd" {
		t.Errorf("expected fingerprint 'aa:bb:cc:dd', got %v", result["fingerprint"])
	}
}

func TestCreateViaHTTP_APIError(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"error":true,"message":"Key pair 'test-kp' already exists"}`)
	})
	defer ts.Close()

	_, err := c.Post(context.Background(), apiPath, map[string]interface{}{
		"name": "test-kp",
	})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestReadViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/cloud/key-pairs/kp-123" {
			t.Errorf("expected path /cloud/key-pairs/kp-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"name":"my-kp","fingerprint":"11:22:33"}}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", apiPath, "kp-123")
	apiResp, err := c.Get(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if result["name"] != "my-kp" {
		t.Errorf("expected name 'my-kp', got %v", result["name"])
	}
	if result["fingerprint"] != "11:22:33" {
		t.Errorf("expected fingerprint '11:22:33', got %v", result["fingerprint"])
	}
}

func TestReadViaHTTP_NotFound(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":true,"message":"Key pair not found"}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", apiPath, "nonexistent")
	_, err := c.Get(context.Background(), path, nil)
	if err == nil {
		t.Fatal("expected error for not-found response")
	}
}

func TestDeleteViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/cloud/key-pairs" {
			t.Errorf("expected path /cloud/key-pairs, got %s", r.URL.Path)
		}

		// Verify body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}
		if body["key"] != "id" {
			t.Errorf("expected key 'id', got %v", body["key"])
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{}}`)
	})
	defer ts.Close()

	deleteBody := map[string]interface{}{
		"key":    "id",
		"values": []string{"kp-del-1"},
	}

	_, err := c.Delete(context.Background(), apiPath, deleteBody)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteViaHTTP_ServerError(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"error":true,"message":"Internal server error"}`)
	})
	defer ts.Close()

	deleteBody := map[string]interface{}{
		"key":    "id",
		"values": []string{"kp-del-1"},
	}
	_, err := c.Delete(context.Background(), apiPath, deleteBody)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

// --- Create response with various data shapes ---

func TestParseCreateResponse_ExtraFields(t *testing.T) {
	// API may return extra fields not in schema; they should be ignored
	data := json.RawMessage(`{
		"id": "kp-extra",
		"fingerprint": "ab:cd",
		"user_id": "user-123",
		"created_at": "2024-01-01T00:00:00Z"
	}`)
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if result["id"] != "kp-extra" {
		t.Errorf("expected id 'kp-extra', got %v", result["id"])
	}
	// Extra fields don't break parsing
	if result["user_id"] != "user-123" {
		t.Errorf("expected user_id to be present in raw parse")
	}
}

// --- Query parameters test ---

func TestReadViaHTTP_QueryParams(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		region := r.URL.Query().Get("region")
		projectID := r.URL.Query().Get("project_id")
		if region != "mumbai" {
			t.Errorf("expected region 'mumbai', got %s", region)
		}
		if projectID != "proj-123" {
			t.Errorf("expected project_id 'proj-123', got %s", projectID)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"name":"my-kp","fingerprint":"99:88"}}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", apiPath, "kp-1")
	_, err := c.Get(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateViaHTTP_RequestBodyVerification(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["name"] != "verified-kp" {
			t.Errorf("expected name 'verified-kp', got %v", body["name"])
		}
		if body["public_key"] != "ssh-ed25519 AAAA..." {
			t.Errorf("expected public_key 'ssh-ed25519 AAAA...', got %v", body["public_key"])
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"id":"kp-v-1","fingerprint":"ff:ee:dd"}}`)
	})
	defer ts.Close()

	_, err := c.Post(context.Background(), apiPath, map[string]interface{}{
		"name":       "verified-kp",
		"public_key": "ssh-ed25519 AAAA...",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateViaHTTP_WithoutPublicKey_NoKeyInBody(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		// public_key should not be in the body
		if _, exists := body["public_key"]; exists {
			t.Error("expected public_key to be absent from request body")
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"id":"kp-auto-1","fingerprint":"11:22:33"}}`)
	})
	defer ts.Close()

	_, err := c.Post(context.Background(), apiPath, map[string]interface{}{
		"name": "auto-gen-kp",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
