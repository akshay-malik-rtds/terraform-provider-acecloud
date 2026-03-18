package router

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("test-router"),
		Description:              types.StringValue("A test router"),
		AdminStateUp:             types.BoolValue(true),
		ExternalGatewayNetworkID: types.StringValue("ext-net-123"),
	}

	body := buildCreateRequest(plan)

	routerBody, ok := body["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected body to have 'router' key with map[string]interface{} value")
	}

	if routerBody["name"] != "test-router" {
		t.Errorf("expected name test-router, got %v", routerBody["name"])
	}
	// Create request should NOT include description
	if _, exists := routerBody["description"]; exists {
		t.Error("expected description to NOT be present in create request (API rejects it)")
	}
	if routerBody["admin_state_up"] != true {
		t.Error("expected admin_state_up to be true")
	}

	gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected external_gateway_info to be map[string]interface{}")
	}
	if gwInfo["network_id"] != "ext-net-123" {
		t.Errorf("expected network_id ext-net-123, got %v", gwInfo["network_id"])
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("minimal-router"),
		Description:              types.StringNull(),
		AdminStateUp:             types.BoolNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	body := buildCreateRequest(plan)

	routerBody, ok := body["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected body to have 'router' key")
	}

	if routerBody["name"] != "minimal-router" {
		t.Errorf("expected name minimal-router, got %v", routerBody["name"])
	}
	// Description should not be present in create request
	if _, exists := routerBody["description"]; exists {
		t.Error("expected description to not be present in create request")
	}
	// admin_state_up should default to true
	if routerBody["admin_state_up"] != true {
		t.Error("expected admin_state_up to default to true")
	}
	// external_gateway_info should be an empty map (not nil, not omitted)
	gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected external_gateway_info to be map[string]interface{}")
	}
	if len(gwInfo) != 0 {
		t.Error("expected empty external_gateway_info when no network ID is set")
	}
}

func TestBuildCreateRequest_WithGateway(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("gw-router"),
		Description:              types.StringNull(),
		AdminStateUp:             types.BoolValue(true),
		ExternalGatewayNetworkID: types.StringValue("ext-net-gw-456"),
	}

	body := buildCreateRequest(plan)

	routerBody, ok := body["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected body to have 'router' key")
	}

	gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected external_gateway_info to be map[string]interface{}")
	}

	// Verify network_id is inside external_gateway_info
	networkID, exists := gwInfo["network_id"]
	if !exists {
		t.Fatal("expected network_id inside external_gateway_info")
	}
	if networkID != "ext-net-gw-456" {
		t.Errorf("expected network_id 'ext-net-gw-456', got %v", networkID)
	}

	// Verify gateway info has exactly 1 key (network_id only)
	if len(gwInfo) != 1 {
		t.Errorf("expected external_gateway_info to have exactly 1 key, got %d", len(gwInfo))
	}

	// Verify description is NOT in create request
	if _, exists := routerBody["description"]; exists {
		t.Error("expected description to NOT be present in create request")
	}
}

func TestBuildUpdateRequest_AllFields(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("updated-router"),
		Description:              types.StringValue("Updated description"),
		AdminStateUp:             types.BoolValue(false),
		ExternalGatewayNetworkID: types.StringValue("ext-net-789"),
	}

	body := buildUpdateRequest(plan)

	routerBody, ok := body["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected body to have 'router' key")
	}

	if routerBody["name"] != "updated-router" {
		t.Errorf("expected name updated-router, got %v", routerBody["name"])
	}
	// Update request SHOULD include description
	if routerBody["description"] != "Updated description" {
		t.Errorf("expected description 'Updated description', got %v", routerBody["description"])
	}
	if routerBody["admin_state_up"] != false {
		t.Error("expected admin_state_up to be false")
	}

	gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected external_gateway_info to be map[string]interface{}")
	}
	if gwInfo["network_id"] != "ext-net-789" {
		t.Errorf("expected network_id ext-net-789, got %v", gwInfo["network_id"])
	}
}

func TestBuildUpdateRequest_MinimalFields(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("simple-router"),
		Description:              types.StringNull(),
		AdminStateUp:             types.BoolNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	body := buildUpdateRequest(plan)

	routerBody, ok := body["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected body to have 'router' key")
	}

	if routerBody["name"] != "simple-router" {
		t.Errorf("expected name simple-router, got %v", routerBody["name"])
	}

	// Description should NOT be present when null
	if _, exists := routerBody["description"]; exists {
		t.Error("expected description to NOT be present when null")
	}

	// admin_state_up should default to true
	if routerBody["admin_state_up"] != true {
		t.Error("expected admin_state_up to default to true")
	}

	// external_gateway_info should be empty map when no gateway
	gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected external_gateway_info to be map[string]interface{}")
	}
	if len(gwInfo) != 0 {
		t.Error("expected empty external_gateway_info when no network ID is set")
	}
}

func TestBuildUpdateRequest_JSON(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("json-router"),
		Description:              types.StringValue("JSON test desc"),
		AdminStateUp:             types.BoolValue(true),
		ExternalGatewayNetworkID: types.StringValue("ext-net-json"),
	}

	body := buildUpdateRequest(plan)

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal update request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	routerMap, ok := parsed["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'router' key in JSON")
	}

	if routerMap["name"] != "json-router" {
		t.Errorf("expected name 'json-router', got %v", routerMap["name"])
	}
	if routerMap["description"] != "JSON test desc" {
		t.Errorf("expected description 'JSON test desc', got %v", routerMap["description"])
	}
	if routerMap["admin_state_up"] != true {
		t.Errorf("expected admin_state_up true, got %v", routerMap["admin_state_up"])
	}

	gwInfo, ok := routerMap["external_gateway_info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected external_gateway_info in JSON")
	}
	if gwInfo["network_id"] != "ext-net-json" {
		t.Errorf("expected network_id 'ext-net-json', got %v", gwInfo["network_id"])
	}
}

func TestBuildCreateRequest_JSONFormat(t *testing.T) {
	// Verify JSON serialization matches ace-cli pattern:
	// empty gateway info should serialize as {} not null
	plan := &routerResourceModel{
		Name:                     types.StringValue("json-test"),
		Description:              types.StringNull(),
		AdminStateUp:             types.BoolNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	body := buildCreateRequest(plan)
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Should contain "external_gateway_info":{}
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	routerMap, ok := parsed["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected router key in JSON")
	}

	gwRaw, exists := routerMap["external_gateway_info"]
	if !exists {
		t.Fatal("expected external_gateway_info to be present in JSON")
	}

	gwMap, ok := gwRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("expected external_gateway_info to be an object in JSON, got %T in %s", gwRaw, jsonStr)
	}
	if len(gwMap) != 0 {
		t.Errorf("expected empty object for gateway info, got %v", gwMap)
	}

	// Should NOT contain description
	if _, exists := routerMap["description"]; exists {
		t.Error("create request JSON should not contain description")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &routerResourceModel{
		Description:              types.StringNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	apiResp := &routerAPIResponse{
		ID:           "router-abc-123",
		Name:         "prod-router",
		Description:  "Production router",
		AdminStateUp: true,
		ExternalGatewayInfo: &routerGatewayInfo{
			NetworkID: "ext-net-456",
		},
		Status:    "ACTIVE",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "router-abc-123" {
		t.Errorf("expected ID router-abc-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-router" {
		t.Errorf("expected Name prod-router, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Production router" {
		t.Errorf("expected Description 'Production router', got %s", model.Description.ValueString())
	}
	if model.AdminStateUp.ValueBool() != true {
		t.Error("expected AdminStateUp to be true")
	}
	if model.ExternalGatewayNetworkID.ValueString() != "ext-net-456" {
		t.Errorf("expected ExternalGatewayNetworkID ext-net-456, got %s", model.ExternalGatewayNetworkID.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
	if model.CreatedAt.ValueString() != "2024-01-01T00:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2024-01-02T00:00:00Z" {
		t.Errorf("expected UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &routerResourceModel{
		Description:              types.StringNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	apiResp := &routerAPIResponse{
		ID:                  "router-123",
		Name:                "basic-router",
		Description:         "",
		AdminStateUp:        false,
		ExternalGatewayInfo: nil,
		Status:              "BUILD",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
	if !model.ExternalGatewayNetworkID.IsNull() {
		t.Error("expected ExternalGatewayNetworkID to remain null when API returns nil gateway info")
	}
}

func TestParseRouterResponse_WrappedFormat(t *testing.T) {
	data := json.RawMessage(`{"router": {"id": "r-123", "name": "test", "status": "ACTIVE", "admin_state_up": true}}`)

	result, err := parseRouterResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "r-123" {
		t.Errorf("expected ID r-123, got %s", result.ID)
	}
	if result.Name != "test" {
		t.Errorf("expected Name test, got %s", result.Name)
	}
}

func TestParseRouterResponse_DirectFormat(t *testing.T) {
	data := json.RawMessage(`{"id": "r-456", "name": "direct-test", "status": "ACTIVE", "admin_state_up": true}`)

	result, err := parseRouterResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "r-456" {
		t.Errorf("expected ID r-456, got %s", result.ID)
	}
}

func TestParseRouterResponse_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`{invalid json content`)

	_, err := parseRouterResponse(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestRouterDeleteRequest(t *testing.T) {
	req := routerDeleteRequest{
		Key:    "id",
		Values: []string{"router-1", "router-2"},
	}
	if req.Key != "id" {
		t.Errorf("expected key 'id', got %s", req.Key)
	}
	if len(req.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(req.Values))
	}
}

func TestRouterDeleteRequest_JSON(t *testing.T) {
	req := routerDeleteRequest{
		Key:    "id",
		Values: []string{"r-1"},
	}

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal delete request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed["key"] != "id" {
		t.Errorf("expected key 'id', got %v", parsed["key"])
	}

	valuesRaw, ok := parsed["values"]
	if !ok {
		t.Fatal("expected 'values' key in JSON")
	}

	values, ok := valuesRaw.([]interface{})
	if !ok {
		t.Fatalf("expected values to be an array, got %T", valuesRaw)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0] != "r-1" {
		t.Errorf("expected value 'r-1', got %v", values[0])
	}

	// Verify exact JSON format
	expectedJSON := `{"key":"id","values":["r-1"]}`
	if string(jsonBytes) != expectedJSON {
		t.Errorf("expected JSON %s, got %s", expectedJSON, string(jsonBytes))
	}
}

func TestRouterNameRegex(t *testing.T) {
	validNames := []string{
		"my-router-1",
		"test123",
		"Router",
		"a",
		"ABC-xyz-123",
	}
	for _, name := range validNames {
		if !routerNameRegex.MatchString(name) {
			t.Errorf("expected name %q to be valid, but it was rejected", name)
		}
	}

	invalidNames := []string{
		"has space",
		"has@char",
		"has.dot",
		"has_underscore",
		"",
		"has/slash",
		"has!bang",
	}
	for _, name := range invalidNames {
		if routerNameRegex.MatchString(name) {
			t.Errorf("expected name %q to be invalid, but it was accepted", name)
		}
	}
}

func TestRouterDescriptionRegex(t *testing.T) {
	validDescriptions := []string{
		"A router desc, here.",
		"test-desc_1",
		"Simple description",
		"with.dots.and" + " spaces",
		"",
		"numbers-123_and_more",
		"comma, separated, words",
	}
	for _, desc := range validDescriptions {
		if !routerDescriptionRegex.MatchString(desc) {
			t.Errorf("expected description %q to be valid, but it was rejected", desc)
		}
	}

	invalidDescriptions := []string{
		"bad<html>",
		"has@symbol",
		"has!bang",
		"has#hash",
		"has$dollar",
		"has(parens)",
		"tab\there",
	}
	for _, desc := range invalidDescriptions {
		if routerDescriptionRegex.MatchString(desc) {
			t.Errorf("expected description %q to be invalid, but it was accepted", desc)
		}
	}
}

// ---------------------------------------------------------------------------
// parseRouterResponse edge cases
// ---------------------------------------------------------------------------

func TestParseRouterResponse_EmptyData(t *testing.T) {
	// Empty JSON object should parse without error but return empty fields.
	data := json.RawMessage(`{}`)

	result, err := parseRouterResponse(data)
	if err != nil {
		t.Fatalf("unexpected error for empty JSON: %v", err)
	}
	// Wrapped format fails (no ID), falls through to direct format.
	// Direct format succeeds but all fields are zero-value.
	if result.ID != "" {
		t.Errorf("expected empty ID, got %s", result.ID)
	}
	if result.Name != "" {
		t.Errorf("expected empty Name, got %s", result.Name)
	}
	if result.Status != "" {
		t.Errorf("expected empty Status, got %s", result.Status)
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest additional tests
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_WithAllOptionalFields(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("full-router"),
		Description:              types.StringValue("Full desc"),
		AdminStateUp:             types.BoolValue(false),
		ExternalGatewayNetworkID: types.StringValue("ext-net-full"),
	}

	body := buildCreateRequest(plan)

	routerBody, ok := body["router"].(map[string]interface{})
	if !ok {
		t.Fatal("expected body to have 'router' key")
	}

	if routerBody["name"] != "full-router" {
		t.Errorf("expected name full-router, got %v", routerBody["name"])
	}
	// Description should NOT be in create request (API rejects it).
	if _, exists := routerBody["description"]; exists {
		t.Error("expected description to NOT be in create request")
	}
	if routerBody["admin_state_up"] != false {
		t.Errorf("expected admin_state_up false, got %v", routerBody["admin_state_up"])
	}

	gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
	if !ok {
		t.Fatal("expected external_gateway_info to be map")
	}
	if gwInfo["network_id"] != "ext-net-full" {
		t.Errorf("expected network_id ext-net-full, got %v", gwInfo["network_id"])
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState additional tests
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_WithGateway(t *testing.T) {
	model := &routerResourceModel{
		Description:              types.StringNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	apiResp := &routerAPIResponse{
		ID:           "r-gw-1",
		Name:         "gateway-router",
		Description:  "",
		AdminStateUp: true,
		ExternalGatewayInfo: &routerGatewayInfo{
			NetworkID: "ext-net-gw-999",
		},
		Status:    "ACTIVE",
		CreatedAt: "2025-06-01T12:00:00Z",
		UpdatedAt: "2025-06-02T12:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ExternalGatewayNetworkID.ValueString() != "ext-net-gw-999" {
		t.Errorf("expected ExternalGatewayNetworkID ext-net-gw-999, got %s", model.ExternalGatewayNetworkID.ValueString())
	}
	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}

	// Test with nil gateway info
	model2 := &routerResourceModel{
		Description:              types.StringNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	apiResp2 := &routerAPIResponse{
		ID:                  "r-no-gw",
		Name:                "no-gw-router",
		AdminStateUp:        true,
		ExternalGatewayInfo: nil,
		Status:              "ACTIVE",
	}

	mapAPIResponseToState(model2, apiResp2)

	if !model2.ExternalGatewayNetworkID.IsNull() {
		t.Error("expected ExternalGatewayNetworkID to remain null when gateway info is nil")
	}

	// Test with empty network_id in gateway info
	model3 := &routerResourceModel{
		Description:              types.StringNull(),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	apiResp3 := &routerAPIResponse{
		ID:                  "r-empty-gw",
		Name:                "empty-gw-router",
		AdminStateUp:        true,
		ExternalGatewayInfo: &routerGatewayInfo{NetworkID: ""},
		Status:              "ACTIVE",
	}

	mapAPIResponseToState(model3, apiResp3)

	if !model3.ExternalGatewayNetworkID.IsNull() {
		t.Error("expected ExternalGatewayNetworkID to remain null when gateway network_id is empty")
	}
}

// ---------------------------------------------------------------------------
// Router name regex additional cases
// ---------------------------------------------------------------------------

func TestRouterNameRegex_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		isValid bool
	}{
		{"single char", "a", true},
		{"digits only", "12345", true},
		{"hyphens only", "---", true},
		{"mixed case", "My-Router-1", true},
		{"long valid name", "abcdefghijklmnopqrstuvwxyz-0123456789-ABCDEFGHIJKLMNOPQRSTUVWXYZ", true},
		{"with space", "my router", false},
		{"with underscore", "my_router", false},
		{"with dot", "my.router", false},
		{"with comma", "my,router", false},
		{"with colon", "my:router", false},
		{"with semicolon", "my;router", false},
		{"with equals", "my=router", false},
		{"with plus", "my+router", false},
		{"with tilde", "my~router", false},
		{"with backtick", "my`router", false},
		{"empty string", "", false},
		{"unicode chars", "routerM\u00fcnchen", false},
		{"tab character", "my\trouter", false},
		{"newline character", "my\nrouter", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matched := routerNameRegex.MatchString(tc.input)
			if matched != tc.isValid {
				if tc.isValid {
					t.Errorf("expected routerNameRegex to match %q, but it did not", tc.input)
				} else {
					t.Errorf("expected routerNameRegex NOT to match %q, but it did", tc.input)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Table-driven: buildCreateRequest
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		planName          string
		planDescription   types.String
		planAdminStateUp  types.Bool
		planGatewayNetID  types.String
		expectAdminState  bool
		expectGatewayID   string
		expectDescription bool
	}{
		{
			name:             "all fields set with gateway",
			planName:         "full-router",
			planDescription:  types.StringValue("full desc"),
			planAdminStateUp: types.BoolValue(true),
			planGatewayNetID: types.StringValue("ext-net-1"),
			expectAdminState: true,
			expectGatewayID:  "ext-net-1",
			// Description is NOT in create request
			expectDescription: false,
		},
		{
			name:              "null optionals default admin_state_up to true",
			planName:          "null-opts",
			planDescription:   types.StringNull(),
			planAdminStateUp:  types.BoolNull(),
			planGatewayNetID:  types.StringNull(),
			expectAdminState:  true,
			expectGatewayID:   "",
			expectDescription: false,
		},
		{
			name:              "unknown optionals default admin_state_up to true",
			planName:          "unknown-opts",
			planDescription:   types.StringUnknown(),
			planAdminStateUp:  types.BoolUnknown(),
			planGatewayNetID:  types.StringUnknown(),
			expectAdminState:  true,
			expectGatewayID:   "",
			expectDescription: false,
		},
		{
			name:              "admin_state_up explicitly false",
			planName:          "disabled-router",
			planDescription:   types.StringNull(),
			planAdminStateUp:  types.BoolValue(false),
			planGatewayNetID:  types.StringNull(),
			expectAdminState:  false,
			expectGatewayID:   "",
			expectDescription: false,
		},
		{
			name:              "description set but excluded from create",
			planName:          "desc-router",
			planDescription:   types.StringValue("should not appear"),
			planAdminStateUp:  types.BoolNull(),
			planGatewayNetID:  types.StringNull(),
			expectAdminState:  true,
			expectGatewayID:   "",
			expectDescription: false,
		},
		{
			name:              "gateway UUID-style network ID",
			planName:          "uuid-gw-router",
			planDescription:   types.StringNull(),
			planAdminStateUp:  types.BoolValue(true),
			planGatewayNetID:  types.StringValue("aabbccdd-1122-3344-5566-778899aabbcc"),
			expectAdminState:  true,
			expectGatewayID:   "aabbccdd-1122-3344-5566-778899aabbcc",
			expectDescription: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := &routerResourceModel{
				Name:                     types.StringValue(tc.planName),
				Description:              tc.planDescription,
				AdminStateUp:             tc.planAdminStateUp,
				ExternalGatewayNetworkID: tc.planGatewayNetID,
			}

			body := buildCreateRequest(plan)

			routerBody, ok := body["router"].(map[string]interface{})
			if !ok {
				t.Fatal("expected 'router' key in body")
			}

			if routerBody["name"] != tc.planName {
				t.Errorf("expected name %q, got %v", tc.planName, routerBody["name"])
			}

			if routerBody["admin_state_up"] != tc.expectAdminState {
				t.Errorf("expected admin_state_up %v, got %v", tc.expectAdminState, routerBody["admin_state_up"])
			}

			_, hasDesc := routerBody["description"]
			if tc.expectDescription && !hasDesc {
				t.Error("expected description to be present in create request")
			}
			if !tc.expectDescription && hasDesc {
				t.Error("description must NOT be present in create request (API rejects it)")
			}

			gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
			if !ok {
				t.Fatal("expected external_gateway_info to be a map")
			}

			if tc.expectGatewayID != "" {
				if gwInfo["network_id"] != tc.expectGatewayID {
					t.Errorf("expected gateway network_id %q, got %v", tc.expectGatewayID, gwInfo["network_id"])
				}
			} else {
				if len(gwInfo) != 0 {
					t.Errorf("expected empty gateway info, got %v", gwInfo)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Table-driven: buildUpdateRequest
// ---------------------------------------------------------------------------

func TestBuildUpdateRequest_TableDriven(t *testing.T) {
	tests := []struct {
		name              string
		planName          string
		planDescription   types.String
		planAdminStateUp  types.Bool
		planGatewayNetID  types.String
		expectAdminState  bool
		expectGatewayID   string
		expectDescription bool
		expectedDescValue string
	}{
		{
			name:              "all fields set",
			planName:          "updated-all",
			planDescription:   types.StringValue("Updated desc"),
			planAdminStateUp:  types.BoolValue(false),
			planGatewayNetID:  types.StringValue("ext-upd-1"),
			expectAdminState:  false,
			expectGatewayID:   "ext-upd-1",
			expectDescription: true,
			expectedDescValue: "Updated desc",
		},
		{
			name:              "null description omitted",
			planName:          "no-desc",
			planDescription:   types.StringNull(),
			planAdminStateUp:  types.BoolNull(),
			planGatewayNetID:  types.StringNull(),
			expectAdminState:  true,
			expectGatewayID:   "",
			expectDescription: false,
		},
		{
			name:              "unknown description omitted",
			planName:          "unknown-desc",
			planDescription:   types.StringUnknown(),
			planAdminStateUp:  types.BoolUnknown(),
			planGatewayNetID:  types.StringUnknown(),
			expectAdminState:  true,
			expectGatewayID:   "",
			expectDescription: false,
		},
		{
			name:              "empty string description included",
			planName:          "empty-desc",
			planDescription:   types.StringValue(""),
			planAdminStateUp:  types.BoolValue(true),
			planGatewayNetID:  types.StringNull(),
			expectAdminState:  true,
			expectGatewayID:   "",
			expectDescription: true,
			expectedDescValue: "",
		},
		{
			name:              "admin_state_up explicitly false",
			planName:          "disabled-upd",
			planDescription:   types.StringNull(),
			planAdminStateUp:  types.BoolValue(false),
			planGatewayNetID:  types.StringNull(),
			expectAdminState:  false,
			expectGatewayID:   "",
			expectDescription: false,
		},
		{
			name:              "gateway network removed (null)",
			planName:          "no-gw-upd",
			planDescription:   types.StringNull(),
			planAdminStateUp:  types.BoolValue(true),
			planGatewayNetID:  types.StringNull(),
			expectAdminState:  true,
			expectGatewayID:   "",
			expectDescription: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := &routerResourceModel{
				Name:                     types.StringValue(tc.planName),
				Description:              tc.planDescription,
				AdminStateUp:             tc.planAdminStateUp,
				ExternalGatewayNetworkID: tc.planGatewayNetID,
			}

			body := buildUpdateRequest(plan)

			routerBody, ok := body["router"].(map[string]interface{})
			if !ok {
				t.Fatal("expected 'router' key in body")
			}

			if routerBody["name"] != tc.planName {
				t.Errorf("expected name %q, got %v", tc.planName, routerBody["name"])
			}

			if routerBody["admin_state_up"] != tc.expectAdminState {
				t.Errorf("expected admin_state_up %v, got %v", tc.expectAdminState, routerBody["admin_state_up"])
			}

			descVal, hasDesc := routerBody["description"]
			if tc.expectDescription && !hasDesc {
				t.Error("expected description to be present")
			}
			if !tc.expectDescription && hasDesc {
				t.Error("expected description to NOT be present")
			}
			if tc.expectDescription && hasDesc && descVal != tc.expectedDescValue {
				t.Errorf("expected description %q, got %v", tc.expectedDescValue, descVal)
			}

			gwInfo, ok := routerBody["external_gateway_info"].(map[string]interface{})
			if !ok {
				t.Fatal("expected external_gateway_info to be a map")
			}

			if tc.expectGatewayID != "" {
				if gwInfo["network_id"] != tc.expectGatewayID {
					t.Errorf("expected gateway network_id %q, got %v", tc.expectGatewayID, gwInfo["network_id"])
				}
			} else {
				if len(gwInfo) != 0 {
					t.Errorf("expected empty gateway info, got %v", gwInfo)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Table-driven: mapAPIResponseToState
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_TableDriven(t *testing.T) {
	tests := []struct {
		name                    string
		apiID                   string
		apiName                 string
		apiDescription          string
		apiAdminStateUp         bool
		apiGatewayInfo          *routerGatewayInfo
		apiStatus               string
		apiCreatedAt            string
		apiUpdatedAt            string
		initialDescNull         bool
		initialGWNull           bool
		expectDescNull          bool
		expectDescValue         string
		expectGWNull            bool
		expectGWValue           string
	}{
		{
			name:            "full response with all fields",
			apiID:           "r-full-1",
			apiName:         "full-router",
			apiDescription:  "Full desc",
			apiAdminStateUp: true,
			apiGatewayInfo:  &routerGatewayInfo{NetworkID: "ext-net-full"},
			apiStatus:       "ACTIVE",
			apiCreatedAt:    "2025-01-01T00:00:00Z",
			apiUpdatedAt:    "2025-01-02T00:00:00Z",
			initialDescNull: true,
			initialGWNull:   true,
			expectDescNull:  false,
			expectDescValue: "Full desc",
			expectGWNull:    false,
			expectGWValue:   "ext-net-full",
		},
		{
			name:            "empty description stays null when initial is null",
			apiID:           "r-emptydesc-1",
			apiName:         "no-desc-router",
			apiDescription:  "",
			apiAdminStateUp: false,
			apiGatewayInfo:  nil,
			apiStatus:       "BUILD",
			apiCreatedAt:    "",
			apiUpdatedAt:    "",
			initialDescNull: true,
			initialGWNull:   true,
			expectDescNull:  true,
			expectGWNull:    true,
		},
		{
			name:            "empty description preserves existing value when not null",
			apiID:           "r-existdesc-1",
			apiName:         "exist-desc",
			apiDescription:  "",
			apiAdminStateUp: true,
			apiGatewayInfo:  nil,
			apiStatus:       "ACTIVE",
			initialDescNull: false,
			initialGWNull:   true,
			expectDescNull:  false,
			expectDescValue: "previous-desc",
			expectGWNull:    true,
		},
		{
			name:            "nil gateway stays null",
			apiID:           "r-nilgw-1",
			apiName:         "nil-gw",
			apiDescription:  "desc",
			apiAdminStateUp: true,
			apiGatewayInfo:  nil,
			apiStatus:       "ACTIVE",
			initialDescNull: true,
			initialGWNull:   true,
			expectDescNull:  false,
			expectDescValue: "desc",
			expectGWNull:    true,
		},
		{
			name:            "gateway with empty network_id stays null",
			apiID:           "r-emptygw-1",
			apiName:         "empty-gw",
			apiDescription:  "",
			apiAdminStateUp: true,
			apiGatewayInfo:  &routerGatewayInfo{NetworkID: ""},
			apiStatus:       "ACTIVE",
			initialDescNull: true,
			initialGWNull:   true,
			expectDescNull:  true,
			expectGWNull:    true,
		},
		{
			name:            "admin_state_up false is preserved",
			apiID:           "r-admin-false",
			apiName:         "admin-false",
			apiDescription:  "",
			apiAdminStateUp: false,
			apiGatewayInfo:  nil,
			apiStatus:       "ACTIVE",
			initialDescNull: true,
			initialGWNull:   true,
			expectDescNull:  true,
			expectGWNull:    true,
		},
		{
			name:            "gateway preserves existing when empty network_id",
			apiID:           "r-gw-preserve",
			apiName:         "gw-preserve",
			apiDescription:  "",
			apiAdminStateUp: true,
			apiGatewayInfo:  &routerGatewayInfo{NetworkID: ""},
			apiStatus:       "ACTIVE",
			initialDescNull: true,
			initialGWNull:   false,
			expectDescNull:  true,
			expectGWNull:    false,
			expectGWValue:   "old-gw-id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &routerResourceModel{}
			if tc.initialDescNull {
				model.Description = types.StringNull()
			} else {
				model.Description = types.StringValue("previous-desc")
			}
			if tc.initialGWNull {
				model.ExternalGatewayNetworkID = types.StringNull()
			} else {
				model.ExternalGatewayNetworkID = types.StringValue("old-gw-id")
			}

			apiResp := &routerAPIResponse{
				ID:                  tc.apiID,
				Name:                tc.apiName,
				Description:         tc.apiDescription,
				AdminStateUp:        tc.apiAdminStateUp,
				ExternalGatewayInfo: tc.apiGatewayInfo,
				Status:              tc.apiStatus,
				CreatedAt:           tc.apiCreatedAt,
				UpdatedAt:           tc.apiUpdatedAt,
			}

			mapAPIResponseToState(model, apiResp)

			if model.ID.ValueString() != tc.apiID {
				t.Errorf("expected ID %q, got %q", tc.apiID, model.ID.ValueString())
			}
			if model.Name.ValueString() != tc.apiName {
				t.Errorf("expected Name %q, got %q", tc.apiName, model.Name.ValueString())
			}
			if model.AdminStateUp.ValueBool() != tc.apiAdminStateUp {
				t.Errorf("expected AdminStateUp %v, got %v", tc.apiAdminStateUp, model.AdminStateUp.ValueBool())
			}
			if model.Status.ValueString() != tc.apiStatus {
				t.Errorf("expected Status %q, got %q", tc.apiStatus, model.Status.ValueString())
			}

			// Description checks
			if tc.expectDescNull {
				if !model.Description.IsNull() {
					t.Errorf("expected Description to be null, got %q", model.Description.ValueString())
				}
			} else {
				if model.Description.IsNull() {
					t.Error("expected Description to not be null")
				} else if model.Description.ValueString() != tc.expectDescValue {
					t.Errorf("expected Description %q, got %q", tc.expectDescValue, model.Description.ValueString())
				}
			}

			// Gateway checks
			if tc.expectGWNull {
				if !model.ExternalGatewayNetworkID.IsNull() {
					t.Errorf("expected ExternalGatewayNetworkID to be null, got %q", model.ExternalGatewayNetworkID.ValueString())
				}
			} else {
				if model.ExternalGatewayNetworkID.IsNull() {
					t.Error("expected ExternalGatewayNetworkID to not be null")
				} else if model.ExternalGatewayNetworkID.ValueString() != tc.expectGWValue {
					t.Errorf("expected ExternalGatewayNetworkID %q, got %q", tc.expectGWValue, model.ExternalGatewayNetworkID.ValueString())
				}
			}

			// CreatedAt/UpdatedAt
			if tc.apiCreatedAt != "" {
				if model.CreatedAt.ValueString() != tc.apiCreatedAt {
					t.Errorf("expected CreatedAt %q, got %q", tc.apiCreatedAt, model.CreatedAt.ValueString())
				}
			} else {
				if !model.CreatedAt.IsNull() {
					t.Errorf("expected CreatedAt to be null, got %q", model.CreatedAt.ValueString())
				}
			}
			if tc.apiUpdatedAt != "" {
				if model.UpdatedAt.ValueString() != tc.apiUpdatedAt {
					t.Errorf("expected UpdatedAt %q, got %q", tc.apiUpdatedAt, model.UpdatedAt.ValueString())
				}
			} else {
				if !model.UpdatedAt.IsNull() {
					t.Errorf("expected UpdatedAt to be null, got %q", model.UpdatedAt.ValueString())
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Table-driven: parseRouterResponse
// ---------------------------------------------------------------------------

func TestParseRouterResponse_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectErr   bool
		expectID    string
		expectName  string
		expectAdmin bool
	}{
		{
			name:        "wrapped format with full fields",
			input:       `{"router": {"id": "r-w-1", "name": "wrapped", "status": "ACTIVE", "admin_state_up": true}}`,
			expectErr:   false,
			expectID:    "r-w-1",
			expectName:  "wrapped",
			expectAdmin: true,
		},
		{
			name:        "direct format with full fields",
			input:       `{"id": "r-d-1", "name": "direct", "status": "BUILD", "admin_state_up": false}`,
			expectErr:   false,
			expectID:    "r-d-1",
			expectName:  "direct",
			expectAdmin: false,
		},
		{
			name:      "invalid JSON",
			input:     `{not valid json`,
			expectErr: true,
		},
		{
			name:        "empty object falls back to direct",
			input:       `{}`,
			expectErr:   false,
			expectID:    "",
			expectName:  "",
			expectAdmin: false,
		},
		{
			name:        "wrapped with gateway info",
			input:       `{"router": {"id": "r-gw-1", "name": "gw-router", "status": "ACTIVE", "admin_state_up": true, "external_gateway_info": {"network_id": "ext-net-1"}}}`,
			expectErr:   false,
			expectID:    "r-gw-1",
			expectName:  "gw-router",
			expectAdmin: true,
		},
		{
			name:        "direct with gateway info",
			input:       `{"id": "r-gw-2", "name": "gw-direct", "status": "ACTIVE", "admin_state_up": true, "external_gateway_info": {"network_id": "ext-net-2"}}`,
			expectErr:   false,
			expectID:    "r-gw-2",
			expectName:  "gw-direct",
			expectAdmin: true,
		},
		{
			name:        "wrapped with null gateway",
			input:       `{"router": {"id": "r-ng-1", "name": "no-gw", "status": "ACTIVE", "admin_state_up": true, "external_gateway_info": null}}`,
			expectErr:   false,
			expectID:    "r-ng-1",
			expectName:  "no-gw",
			expectAdmin: true,
		},
		{
			name:        "wrapped with description and timestamps",
			input:       `{"router": {"id": "r-ts-1", "name": "ts-router", "description": "test desc", "status": "ACTIVE", "admin_state_up": true, "created_at": "2025-06-01T00:00:00Z", "updated_at": "2025-06-02T00:00:00Z"}}`,
			expectErr:   false,
			expectID:    "r-ts-1",
			expectName:  "ts-router",
			expectAdmin: true,
		},
		{
			name:        "direct with empty ID falls back to direct parse",
			input:       `{"router": {"id": "", "name": ""}}`,
			expectErr:   false,
			expectID:    "",
			expectName:  "",
			expectAdmin: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseRouterResponse(json.RawMessage(tc.input))

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.ID != tc.expectID {
				t.Errorf("expected ID %q, got %q", tc.expectID, result.ID)
			}
			if result.Name != tc.expectName {
				t.Errorf("expected Name %q, got %q", tc.expectName, result.Name)
			}
			if result.AdminStateUp != tc.expectAdmin {
				t.Errorf("expected AdminStateUp %v, got %v", tc.expectAdmin, result.AdminStateUp)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseRouterResponse: gateway info extraction
// ---------------------------------------------------------------------------

func TestParseRouterResponse_GatewayInfo(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectGateway   bool
		expectNetworkID string
	}{
		{
			name:            "wrapped with gateway",
			input:           `{"router": {"id": "r-1", "name": "gw", "status": "ACTIVE", "admin_state_up": true, "external_gateway_info": {"network_id": "ext-1"}}}`,
			expectGateway:   true,
			expectNetworkID: "ext-1",
		},
		{
			name:          "wrapped without gateway",
			input:         `{"router": {"id": "r-2", "name": "no-gw", "status": "ACTIVE", "admin_state_up": true}}`,
			expectGateway: false,
		},
		{
			name:          "wrapped with null gateway",
			input:         `{"router": {"id": "r-3", "name": "null-gw", "status": "ACTIVE", "admin_state_up": true, "external_gateway_info": null}}`,
			expectGateway: false,
		},
		{
			name:            "direct with gateway",
			input:           `{"id": "r-4", "name": "d-gw", "status": "ACTIVE", "admin_state_up": true, "external_gateway_info": {"network_id": "ext-4"}}`,
			expectGateway:   true,
			expectNetworkID: "ext-4",
		},
		{
			name:            "wrapped with empty network_id",
			input:           `{"router": {"id": "r-5", "name": "empty-gw", "status": "ACTIVE", "admin_state_up": true, "external_gateway_info": {"network_id": ""}}}`,
			expectGateway:   true,
			expectNetworkID: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseRouterResponse(json.RawMessage(tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.expectGateway {
				if result.ExternalGatewayInfo == nil {
					t.Fatal("expected ExternalGatewayInfo to not be nil")
				}
				if result.ExternalGatewayInfo.NetworkID != tc.expectNetworkID {
					t.Errorf("expected NetworkID %q, got %q", tc.expectNetworkID, result.ExternalGatewayInfo.NetworkID)
				}
			} else {
				if result.ExternalGatewayInfo != nil {
					t.Errorf("expected ExternalGatewayInfo to be nil, got %+v", result.ExternalGatewayInfo)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Delete request: table-driven
// ---------------------------------------------------------------------------

func TestRouterDeleteRequest_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		values       []string
		expectedJSON string
	}{
		{
			name:         "single ID",
			key:          "id",
			values:       []string{"r-1"},
			expectedJSON: `{"key":"id","values":["r-1"]}`,
		},
		{
			name:         "multiple IDs",
			key:          "id",
			values:       []string{"r-1", "r-2", "r-3"},
			expectedJSON: `{"key":"id","values":["r-1","r-2","r-3"]}`,
		},
		{
			name:         "UUID values",
			key:          "id",
			values:       []string{"aabbccdd-1122-3344-5566-778899aabbcc"},
			expectedJSON: `{"key":"id","values":["aabbccdd-1122-3344-5566-778899aabbcc"]}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := routerDeleteRequest{
				Key:    tc.key,
				Values: tc.values,
			}

			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			if string(data) != tc.expectedJSON {
				t.Errorf("expected JSON:\n  %s\ngot:\n  %s", tc.expectedJSON, string(data))
			}

			// Verify round-trip
			var decoded routerDeleteRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			if decoded.Key != tc.key {
				t.Errorf("expected key %q, got %q", tc.key, decoded.Key)
			}
			if len(decoded.Values) != len(tc.values) {
				t.Fatalf("expected %d values, got %d", len(tc.values), len(decoded.Values))
			}
			for i, v := range tc.values {
				if decoded.Values[i] != v {
					t.Errorf("expected values[%d] %q, got %q", i, v, decoded.Values[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Create vs Update: description handling difference
// ---------------------------------------------------------------------------

func TestCreateVsUpdate_DescriptionDifference(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("desc-test-router"),
		Description:              types.StringValue("My description"),
		AdminStateUp:             types.BoolValue(true),
		ExternalGatewayNetworkID: types.StringNull(),
	}

	createBody := buildCreateRequest(plan)
	updateBody := buildUpdateRequest(plan)

	createRouter := createBody["router"].(map[string]interface{})
	updateRouter := updateBody["router"].(map[string]interface{})

	// Create must NOT have description
	if _, exists := createRouter["description"]; exists {
		t.Error("create request must NOT include description")
	}

	// Update MUST have description
	desc, exists := updateRouter["description"]
	if !exists {
		t.Error("update request must include description")
	}
	if desc != "My description" {
		t.Errorf("expected description 'My description', got %v", desc)
	}
}

// ---------------------------------------------------------------------------
// JSON serialization format: create and update
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_JSONSerialization(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("json-create"),
		Description:              types.StringNull(),
		AdminStateUp:             types.BoolValue(true),
		ExternalGatewayNetworkID: types.StringValue("ext-json"),
	}

	body := buildCreateRequest(plan)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify the JSON can be parsed as nested map
	var parsed map[string]map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	router, ok := parsed["router"]
	if !ok {
		t.Fatal("expected 'router' key in JSON")
	}

	if router["name"] != "json-create" {
		t.Errorf("expected name 'json-create', got %v", router["name"])
	}

	// Verify gateway info is a nested object
	gwRaw, exists := router["external_gateway_info"]
	if !exists {
		t.Fatal("expected external_gateway_info in JSON")
	}
	gw, ok := gwRaw.(map[string]interface{})
	if !ok {
		t.Fatalf("expected external_gateway_info to be object, got %T", gwRaw)
	}
	if gw["network_id"] != "ext-json" {
		t.Errorf("expected network_id 'ext-json', got %v", gw["network_id"])
	}
}

func TestBuildUpdateRequest_JSONSerialization(t *testing.T) {
	plan := &routerResourceModel{
		Name:                     types.StringValue("json-update"),
		Description:              types.StringValue("Update desc"),
		AdminStateUp:             types.BoolValue(false),
		ExternalGatewayNetworkID: types.StringValue("ext-upd-json"),
	}

	body := buildUpdateRequest(plan)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	router := parsed["router"]
	if router["description"] != "Update desc" {
		t.Errorf("expected description 'Update desc', got %v", router["description"])
	}
	if router["admin_state_up"] != false {
		t.Errorf("expected admin_state_up false, got %v", router["admin_state_up"])
	}
}

// ---------------------------------------------------------------------------
// Schema verification
// ---------------------------------------------------------------------------

func TestRouterSchema_RequiredAttributes(t *testing.T) {
	s := routerSchema()

	attr, ok := s.Attributes["name"]
	if !ok {
		t.Fatal("expected 'name' attribute in schema")
	}
	if !attr.IsRequired() {
		t.Error("expected 'name' to be required")
	}
}

func TestRouterSchema_OptionalAttributes(t *testing.T) {
	s := routerSchema()

	optionalFields := []string{"description", "admin_state_up", "external_gateway_network_id"}
	for _, field := range optionalFields {
		attr, ok := s.Attributes[field]
		if !ok {
			t.Errorf("expected %q attribute in schema", field)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected %q to be optional", field)
		}
	}
}

func TestRouterSchema_ComputedAttributes(t *testing.T) {
	s := routerSchema()

	computedOnly := []string{"id", "status", "created_at", "updated_at"}
	for _, field := range computedOnly {
		attr, ok := s.Attributes[field]
		if !ok {
			t.Errorf("expected %q attribute in schema", field)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("expected %q to be computed", field)
		}
	}
}

func TestRouterSchema_AllAttributes(t *testing.T) {
	s := routerSchema()

	expected := []string{"id", "name", "description", "admin_state_up", "external_gateway_network_id", "status", "created_at", "updated_at"}
	for _, name := range expected {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("expected attribute %q in schema", name)
		}
	}
	if len(s.Attributes) != len(expected) {
		t.Errorf("expected %d attributes, got %d", len(expected), len(s.Attributes))
	}
}

func TestRouterSchema_Description(t *testing.T) {
	s := routerSchema()
	if s.Description == "" {
		t.Error("expected schema to have a non-empty description")
	}
}

// ---------------------------------------------------------------------------
// API path and Metadata
// ---------------------------------------------------------------------------

func TestAPIPath(t *testing.T) {
	if apiPath != "/os/neutron/routers" {
		t.Errorf("expected apiPath /os/neutron/routers, got %s", apiPath)
	}
}

func TestMetadata(t *testing.T) {
	r := &routerResource{}
	req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	expected := "acecloud_router"
	if resp.TypeName != expected {
		t.Errorf("expected type name %q, got %q", expected, resp.TypeName)
	}
}

func TestMetadata_DifferentProvider(t *testing.T) {
	r := &routerResource{}
	req := resource.MetadataRequest{ProviderTypeName: "testprovider"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	expected := "testprovider_router"
	if resp.TypeName != expected {
		t.Errorf("expected type name %q, got %q", expected, resp.TypeName)
	}
}

// ---------------------------------------------------------------------------
// API type JSON round-trip
// ---------------------------------------------------------------------------

func TestRouterGatewayInfo_JSON(t *testing.T) {
	gw := routerGatewayInfo{NetworkID: "ext-net-rt-1"}

	data, err := json.Marshal(gw)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded routerGatewayInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.NetworkID != "ext-net-rt-1" {
		t.Errorf("expected NetworkID ext-net-rt-1, got %s", decoded.NetworkID)
	}
}

func TestRouterAPIResponse_JSONRoundTrip(t *testing.T) {
	resp := routerAPIResponse{
		ID:           "r-rt-1",
		Name:         "roundtrip-router",
		Description:  "Roundtrip desc",
		AdminStateUp: true,
		ExternalGatewayInfo: &routerGatewayInfo{
			NetworkID: "ext-net-rt",
		},
		Status:    "ACTIVE",
		CreatedAt: "2025-01-01T00:00:00Z",
		UpdatedAt: "2025-01-02T00:00:00Z",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded routerAPIResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ID != resp.ID {
		t.Errorf("ID mismatch: got %q, want %q", decoded.ID, resp.ID)
	}
	if decoded.Name != resp.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, resp.Name)
	}
	if decoded.Description != resp.Description {
		t.Errorf("Description mismatch: got %q, want %q", decoded.Description, resp.Description)
	}
	if decoded.AdminStateUp != resp.AdminStateUp {
		t.Errorf("AdminStateUp mismatch: got %v, want %v", decoded.AdminStateUp, resp.AdminStateUp)
	}
	if decoded.ExternalGatewayInfo == nil {
		t.Fatal("expected ExternalGatewayInfo to not be nil")
	}
	if decoded.ExternalGatewayInfo.NetworkID != "ext-net-rt" {
		t.Errorf("NetworkID mismatch: got %q, want %q", decoded.ExternalGatewayInfo.NetworkID, "ext-net-rt")
	}
	if decoded.Status != resp.Status {
		t.Errorf("Status mismatch: got %q, want %q", decoded.Status, resp.Status)
	}
	if decoded.CreatedAt != resp.CreatedAt {
		t.Errorf("CreatedAt mismatch: got %q, want %q", decoded.CreatedAt, resp.CreatedAt)
	}
	if decoded.UpdatedAt != resp.UpdatedAt {
		t.Errorf("UpdatedAt mismatch: got %q, want %q", decoded.UpdatedAt, resp.UpdatedAt)
	}
}

func TestRouterAPIResponse_JSONNilGateway(t *testing.T) {
	resp := routerAPIResponse{
		ID:                  "r-nil-gw",
		Name:                "nil-gw-router",
		AdminStateUp:        true,
		ExternalGatewayInfo: nil,
		Status:              "ACTIVE",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded routerAPIResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded.ExternalGatewayInfo != nil {
		t.Error("expected ExternalGatewayInfo to be nil after round-trip")
	}
}

// ---------------------------------------------------------------------------
// Description regex: additional edge cases
// ---------------------------------------------------------------------------

func TestRouterDescriptionRegex_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		isValid bool
	}{
		{"only spaces", "   ", true},
		{"only periods", "...", true},
		{"only commas", ",,,", true},
		{"only underscores", "___", true},
		{"only hyphens", "---", true},
		{"mixed punctuation", "test-desc_1, here.", true},
		{"numbers", "1234567890", true},
		{"with forward slash", "test/desc", false},
		{"with backslash", "test\\desc", false},
		{"with ampersand", "test&desc", false},
		{"with percent", "test%desc", false},
		{"with caret", "test^desc", false},
		{"with curly braces", "test{desc}", false},
		{"with square brackets", "test[desc]", false},
		{"with pipe", "test|desc", false},
		{"with question mark", "test?desc", false},
		{"with double quotes", `test"desc"`, false},
		{"with single quotes", "test'desc'", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matched := routerDescriptionRegex.MatchString(tc.input)
			if matched != tc.isValid {
				if tc.isValid {
					t.Errorf("expected routerDescriptionRegex to match %q", tc.input)
				} else {
					t.Errorf("expected routerDescriptionRegex NOT to match %q", tc.input)
				}
			}
		})
	}
}
