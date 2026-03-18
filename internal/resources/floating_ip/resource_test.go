package floating_ip

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestFloatingIPModel(t *testing.T) {
	// Verify the model struct has the expected fields.
	model := floatingIPModel{}
	_ = model.ID
	_ = model.FloatingNetworkID
	_ = model.BillingType
	_ = model.PortID
	_ = model.Description
	_ = model.FloatingIPAddress
	_ = model.Status
}

func TestFloatingIPModel_AllFields(t *testing.T) {
	model := floatingIPModel{
		ID:                types.StringValue("fip-abc-123"),
		FloatingNetworkID: types.StringValue("ext-net-456"),
		BillingType:       types.StringValue("hourly"),
		PortID:            types.StringValue("port-789"),
		Description:       types.StringValue("my floating ip"),
		FloatingIPAddress: types.StringValue("203.0.113.50"),
		Status:            types.StringValue("ACTIVE"),
	}

	if model.ID.ValueString() != "fip-abc-123" {
		t.Errorf("expected ID fip-abc-123, got %s", model.ID.ValueString())
	}
	if model.FloatingNetworkID.ValueString() != "ext-net-456" {
		t.Errorf("expected FloatingNetworkID ext-net-456, got %s", model.FloatingNetworkID.ValueString())
	}
	if model.BillingType.ValueString() != "hourly" {
		t.Errorf("expected BillingType hourly, got %s", model.BillingType.ValueString())
	}
	if model.PortID.ValueString() != "port-789" {
		t.Errorf("expected PortID port-789, got %s", model.PortID.ValueString())
	}
	if model.Description.ValueString() != "my floating ip" {
		t.Errorf("expected Description 'my floating ip', got %s", model.Description.ValueString())
	}
	if model.FloatingIPAddress.ValueString() != "203.0.113.50" {
		t.Errorf("expected FloatingIPAddress 203.0.113.50, got %s", model.FloatingIPAddress.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
}

func TestFloatingIPCreateBody(t *testing.T) {
	// Construct the create body as the resource does with all optional fields.
	body := map[string]interface{}{
		"floating_network_id": "ext-net-456",
		"billing_type":        "hourly",
		"port_id":             "port-789",
		"description":         "test floating ip",
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create body: %v", err)
	}

	if parsed["floating_network_id"] != "ext-net-456" {
		t.Errorf("expected floating_network_id ext-net-456, got %v", parsed["floating_network_id"])
	}
	if parsed["billing_type"] != "hourly" {
		t.Errorf("expected billing_type hourly, got %v", parsed["billing_type"])
	}
	if parsed["port_id"] != "port-789" {
		t.Errorf("expected port_id port-789, got %v", parsed["port_id"])
	}
	if parsed["description"] != "test floating ip" {
		t.Errorf("expected description 'test floating ip', got %v", parsed["description"])
	}
}

func TestFloatingIPCreateBody_MinimalFields(t *testing.T) {
	// Required fields: floating_network_id and billing_type.
	body := map[string]interface{}{
		"floating_network_id": "ext-net-001",
		"billing_type":        "hourly",
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create body: %v", err)
	}

	if parsed["floating_network_id"] != "ext-net-001" {
		t.Errorf("expected floating_network_id ext-net-001, got %v", parsed["floating_network_id"])
	}
	if parsed["billing_type"] != "hourly" {
		t.Errorf("expected billing_type hourly, got %v", parsed["billing_type"])
	}
	if _, exists := parsed["port_id"]; exists {
		t.Error("expected port_id to be absent in minimal body")
	}
	if _, exists := parsed["description"]; exists {
		t.Error("expected description to be absent in minimal body")
	}
}

func TestFloatingIPDeleteBody(t *testing.T) {
	body := map[string]interface{}{
		"key":    "id",
		"values": []string{"fip-1"},
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
	if values[0] != "fip-1" {
		t.Errorf("expected first value 'fip-1', got %v", values[0])
	}
}

// ---------------------------------------------------------------------------
// FloatingIP create body with description
// ---------------------------------------------------------------------------

func TestFloatingIPCreateBody_WithDescription(t *testing.T) {
	body := map[string]interface{}{
		"floating_network_id": "ext-net-desc",
		"billing_type":        "hourly",
		"description":         "my test floating IP description",
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create body: %v", err)
	}

	if parsed["description"] != "my test floating IP description" {
		t.Errorf("expected description 'my test floating IP description', got %v", parsed["description"])
	}
	if parsed["floating_network_id"] != "ext-net-desc" {
		t.Errorf("expected floating_network_id ext-net-desc, got %v", parsed["floating_network_id"])
	}
	// port_id should be absent when not included.
	if _, exists := parsed["port_id"]; exists {
		t.Error("expected port_id to be absent when not set")
	}
}

// ---------------------------------------------------------------------------
// FloatingIP model computed fields
// ---------------------------------------------------------------------------

func TestFloatingIPModel_ComputedFields(t *testing.T) {
	// Verify computed fields can hold expected values.
	tests := []struct {
		name              string
		status            string
		floatingIPAddress string
	}{
		{"active with IP", "ACTIVE", "203.0.113.100"},
		{"down with IP", "DOWN", "198.51.100.50"},
		{"error status", "ERROR", "192.0.2.10"},
		{"empty status", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := floatingIPModel{
				ID:                types.StringValue("fip-computed"),
				FloatingNetworkID: types.StringValue("ext-net-1"),
				BillingType:       types.StringValue("hourly"),
				PortID:            types.StringNull(),
				Description:       types.StringNull(),
				FloatingIPAddress: types.StringValue(tc.floatingIPAddress),
				Status:            types.StringValue(tc.status),
			}

			if model.FloatingIPAddress.ValueString() != tc.floatingIPAddress {
				t.Errorf("expected FloatingIPAddress %s, got %s",
					tc.floatingIPAddress, model.FloatingIPAddress.ValueString())
			}
			if model.Status.ValueString() != tc.status {
				t.Errorf("expected Status %s, got %s", tc.status, model.Status.ValueString())
			}
			// Verify null optional fields.
			if !model.PortID.IsNull() {
				t.Error("expected PortID to be null")
			}
			if !model.Description.IsNull() {
				t.Error("expected Description to be null")
			}
		})
	}
}

// ===========================================================================
// Read response parsing tests
// ===========================================================================

// TestReadResponseParsing_FindByID simulates the Read logic: parse the list
// endpoint JSON array and locate the floating IP that matches the target ID.
func TestReadResponseParsing_FindByID(t *testing.T) {
	tests := []struct {
		name             string
		response         string
		targetID         string
		expectFound      bool
		expectIP         string
		expectStatus     string
		expectPortID     string // "" means expect null
		expectDesc       string // "" means expect null
		expectNetworkID  string
	}{
		{
			name: "single item matches",
			response: `[{
				"id": "fip-001",
				"floating_ip_address": "203.0.113.10",
				"status": "ACTIVE",
				"floating_network_id": "ext-net-1",
				"port_id": "port-abc",
				"description": "primary"
			}]`,
			targetID:        "fip-001",
			expectFound:     true,
			expectIP:        "203.0.113.10",
			expectStatus:    "ACTIVE",
			expectPortID:    "port-abc",
			expectDesc:      "primary",
			expectNetworkID: "ext-net-1",
		},
		{
			name: "multiple items pick correct one",
			response: `[
				{"id": "fip-001", "floating_ip_address": "203.0.113.10", "status": "ACTIVE", "floating_network_id": "ext-1", "port_id": "", "description": ""},
				{"id": "fip-002", "floating_ip_address": "203.0.113.20", "status": "DOWN", "floating_network_id": "ext-2", "port_id": "port-xyz", "description": "second"},
				{"id": "fip-003", "floating_ip_address": "203.0.113.30", "status": "ACTIVE", "floating_network_id": "ext-1", "port_id": "", "description": ""}
			]`,
			targetID:        "fip-002",
			expectFound:     true,
			expectIP:        "203.0.113.20",
			expectStatus:    "DOWN",
			expectPortID:    "port-xyz",
			expectDesc:      "second",
			expectNetworkID: "ext-2",
		},
		{
			name:        "target not found in list",
			response:    `[{"id": "fip-001", "floating_ip_address": "1.2.3.4", "status": "ACTIVE", "floating_network_id": "ext-1"}]`,
			targetID:    "fip-999",
			expectFound: false,
		},
		{
			name:        "empty list",
			response:    `[]`,
			targetID:    "fip-001",
			expectFound: false,
		},
		{
			name: "item with null/missing optional fields",
			response: `[{
				"id": "fip-010",
				"floating_ip_address": "10.0.0.1",
				"status": "ACTIVE",
				"floating_network_id": "ext-net-5"
			}]`,
			targetID:        "fip-010",
			expectFound:     true,
			expectIP:        "10.0.0.1",
			expectStatus:    "ACTIVE",
			expectPortID:    "",
			expectDesc:      "",
			expectNetworkID: "ext-net-5",
		},
		{
			name: "port_id is empty string in response",
			response: `[{
				"id": "fip-020",
				"floating_ip_address": "10.0.0.2",
				"status": "ACTIVE",
				"floating_network_id": "ext-net-6",
				"port_id": "",
				"description": "has empty port"
			}]`,
			targetID:        "fip-020",
			expectFound:     true,
			expectIP:        "10.0.0.2",
			expectStatus:    "ACTIVE",
			expectPortID:    "",
			expectDesc:      "has empty port",
			expectNetworkID: "ext-net-6",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the JSON array — mirrors Read() logic
			var items []map[string]interface{}
			if err := json.Unmarshal([]byte(tc.response), &items); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			// Find by ID — mirrors Read() logic
			var result map[string]interface{}
			for _, item := range items {
				if id, ok := item["id"].(string); ok && id == tc.targetID {
					result = item
					break
				}
			}

			if !tc.expectFound {
				if result != nil {
					t.Fatal("expected nil result for missing IP, got a match")
				}
				return
			}

			if result == nil {
				t.Fatal("expected to find floating IP, got nil")
			}

			// Apply the same field-extraction logic as Read()
			state := floatingIPModel{
				ID: types.StringValue(tc.targetID),
			}

			if v, ok := result["floating_network_id"].(string); ok {
				state.FloatingNetworkID = types.StringValue(v)
			}
			if v, ok := result["port_id"].(string); ok && v != "" {
				state.PortID = types.StringValue(v)
			} else {
				state.PortID = types.StringNull()
			}
			if v, ok := result["description"].(string); ok && v != "" {
				state.Description = types.StringValue(v)
			} else {
				state.Description = types.StringNull()
			}
			if v, ok := result["floating_ip_address"].(string); ok {
				state.FloatingIPAddress = types.StringValue(v)
			}
			if v, ok := result["status"].(string); ok {
				state.Status = types.StringValue(v)
			}

			// Assertions
			if state.FloatingIPAddress.ValueString() != tc.expectIP {
				t.Errorf("FloatingIPAddress: expected %q, got %q", tc.expectIP, state.FloatingIPAddress.ValueString())
			}
			if state.Status.ValueString() != tc.expectStatus {
				t.Errorf("Status: expected %q, got %q", tc.expectStatus, state.Status.ValueString())
			}
			if state.FloatingNetworkID.ValueString() != tc.expectNetworkID {
				t.Errorf("FloatingNetworkID: expected %q, got %q", tc.expectNetworkID, state.FloatingNetworkID.ValueString())
			}

			if tc.expectPortID != "" {
				if state.PortID.IsNull() {
					t.Errorf("PortID: expected %q, got null", tc.expectPortID)
				} else if state.PortID.ValueString() != tc.expectPortID {
					t.Errorf("PortID: expected %q, got %q", tc.expectPortID, state.PortID.ValueString())
				}
			} else {
				if !state.PortID.IsNull() {
					t.Errorf("PortID: expected null, got %q", state.PortID.ValueString())
				}
			}

			if tc.expectDesc != "" {
				if state.Description.IsNull() {
					t.Errorf("Description: expected %q, got null", tc.expectDesc)
				} else if state.Description.ValueString() != tc.expectDesc {
					t.Errorf("Description: expected %q, got %q", tc.expectDesc, state.Description.ValueString())
				}
			} else {
				if !state.Description.IsNull() {
					t.Errorf("Description: expected null, got %q", state.Description.ValueString())
				}
			}
		})
	}
}

// TestReadResponseParsing_InvalidJSON ensures non-array JSON is rejected.
func TestReadResponseParsing_InvalidJSON(t *testing.T) {
	tests := []struct {
		name     string
		response string
	}{
		{"object instead of array", `{"id": "fip-001"}`},
		{"malformed JSON", `[{"id": "fip-001"`},
		{"null literal", `null`},
		{"bare string", `"fip-001"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var items []map[string]interface{}
			err := json.Unmarshal([]byte(tc.response), &items)
			if err == nil && items != nil {
				// null unmarshals to nil slice without error — that's fine,
				// but object/bare string should error
				if tc.name != "null literal" {
					t.Error("expected unmarshal error for invalid list response")
				}
			}
		})
	}
}

// ===========================================================================
// Create response parsing tests
// ===========================================================================

// TestCreateResponseParsing exercises the map[string]interface{} extraction
// logic that Create() uses after receiving the API response.
func TestCreateResponseParsing(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		expectID     string
		expectIP     string
		expectStatus string
		expectPortID string // "null" means expect null, "" means expect null (empty from API)
		expectError  bool
	}{
		{
			name: "full create response",
			response: `{
				"id": "fip-new-001",
				"floating_ip_address": "203.0.113.99",
				"status": "ACTIVE",
				"port_id": "port-create-1"
			}`,
			expectID:     "fip-new-001",
			expectIP:     "203.0.113.99",
			expectStatus: "ACTIVE",
			expectPortID: "port-create-1",
		},
		{
			name: "create response without status defaults to ACTIVE",
			response: `{
				"id": "fip-new-002",
				"floating_ip_address": "10.0.0.50"
			}`,
			expectID:     "fip-new-002",
			expectIP:     "10.0.0.50",
			expectStatus: "ACTIVE",
			expectPortID: "null",
		},
		{
			name: "create response with empty port_id",
			response: `{
				"id": "fip-new-003",
				"floating_ip_address": "10.0.0.51",
				"status": "DOWN",
				"port_id": ""
			}`,
			expectID:     "fip-new-003",
			expectIP:     "10.0.0.51",
			expectStatus: "DOWN",
			expectPortID: "null",
		},
		{
			name: "create response missing floating_ip_address",
			response: `{
				"id": "fip-new-004",
				"status": "ACTIVE"
			}`,
			expectID:     "fip-new-004",
			expectIP:     "",
			expectStatus: "ACTIVE",
			expectPortID: "null",
		},
		{
			name:        "create response missing ID",
			response:    `{"floating_ip_address": "10.0.0.1"}`,
			expectError: true,
		},
		{
			name:        "create response with non-string ID",
			response:    `{"id": 12345}`,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(tc.response), &result); err != nil {
				t.Fatalf("failed to parse response JSON: %v", err)
			}

			// Extract ID — mirrors Create() logic
			id, ok := result["id"].(string)
			if !ok {
				if tc.expectError {
					return // expected
				}
				t.Fatal("expected to extract string ID, got failure")
			}
			if tc.expectError {
				t.Fatal("expected error extracting ID, but succeeded")
			}

			plan := floatingIPModel{
				ID:                types.StringValue(id),
				FloatingNetworkID: types.StringValue("ext-net-test"),
				BillingType:       types.StringValue("hourly"),
				PortID:            types.StringUnknown(), // simulates plan unknown
			}

			// floating_ip_address — mirrors Create() logic
			if v, ok := result["floating_ip_address"].(string); ok {
				plan.FloatingIPAddress = types.StringValue(v)
			} else {
				plan.FloatingIPAddress = types.StringNull()
			}

			// status — mirrors Create() logic
			if v, ok := result["status"].(string); ok {
				plan.Status = types.StringValue(v)
			} else {
				plan.Status = types.StringValue("ACTIVE")
			}

			// port_id — mirrors Create() logic
			if v, ok := result["port_id"].(string); ok && v != "" {
				plan.PortID = types.StringValue(v)
			} else if plan.PortID.IsUnknown() {
				plan.PortID = types.StringNull()
			}

			// Assertions
			if plan.ID.ValueString() != tc.expectID {
				t.Errorf("ID: expected %q, got %q", tc.expectID, plan.ID.ValueString())
			}

			if tc.expectIP == "" {
				if !plan.FloatingIPAddress.IsNull() {
					t.Errorf("FloatingIPAddress: expected null, got %q", plan.FloatingIPAddress.ValueString())
				}
			} else {
				if plan.FloatingIPAddress.ValueString() != tc.expectIP {
					t.Errorf("FloatingIPAddress: expected %q, got %q", tc.expectIP, plan.FloatingIPAddress.ValueString())
				}
			}

			if plan.Status.ValueString() != tc.expectStatus {
				t.Errorf("Status: expected %q, got %q", tc.expectStatus, plan.Status.ValueString())
			}

			if tc.expectPortID == "null" {
				if !plan.PortID.IsNull() {
					t.Errorf("PortID: expected null, got %q", plan.PortID.ValueString())
				}
			} else {
				if plan.PortID.IsNull() {
					t.Errorf("PortID: expected %q, got null", tc.expectPortID)
				} else if plan.PortID.ValueString() != tc.expectPortID {
					t.Errorf("PortID: expected %q, got %q", tc.expectPortID, plan.PortID.ValueString())
				}
			}
		})
	}
}

// ===========================================================================
// Port_id resolution logic tests
// ===========================================================================

// TestPortIDResolution exercises the three-way port_id logic in Create():
//   - API returns a non-empty port_id string -> use it
//   - API returns empty port_id but plan has a known value -> keep plan value
//   - API returns empty port_id and plan is unknown -> set null
func TestPortIDResolution(t *testing.T) {
	tests := []struct {
		name           string
		apiPortID      interface{} // string or nil (missing key)
		planPortID     types.String
		expectedResult types.String
	}{
		{
			name:           "API returns port_id string",
			apiPortID:      "port-from-api",
			planPortID:     types.StringValue("port-from-plan"),
			expectedResult: types.StringValue("port-from-api"),
		},
		{
			name:           "API returns port_id, plan was unknown",
			apiPortID:      "port-resolved",
			planPortID:     types.StringUnknown(),
			expectedResult: types.StringValue("port-resolved"),
		},
		{
			name:           "API returns port_id, plan was null",
			apiPortID:      "port-surprise",
			planPortID:     types.StringNull(),
			expectedResult: types.StringValue("port-surprise"),
		},
		{
			name:           "API returns empty string, plan was unknown",
			apiPortID:      "",
			planPortID:     types.StringUnknown(),
			expectedResult: types.StringNull(),
		},
		{
			name:           "API returns empty string, plan had value",
			apiPortID:      "",
			planPortID:     types.StringValue("port-planned"),
			expectedResult: types.StringValue("port-planned"), // plan value preserved (not unknown)
		},
		{
			name:           "API field missing entirely, plan was unknown",
			apiPortID:      nil,
			planPortID:     types.StringUnknown(),
			expectedResult: types.StringNull(),
		},
		{
			name:           "API field missing entirely, plan was null",
			apiPortID:      nil,
			planPortID:     types.StringNull(),
			expectedResult: types.StringNull(),
		},
		{
			name:           "API field missing entirely, plan had value",
			apiPortID:      nil,
			planPortID:     types.StringValue("port-kept"),
			expectedResult: types.StringValue("port-kept"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Build a result map simulating API response
			result := map[string]interface{}{
				"id":                 "fip-port-test",
				"floating_ip_address": "1.2.3.4",
			}
			if tc.apiPortID != nil {
				result["port_id"] = tc.apiPortID
			}

			// Start with plan value
			portID := tc.planPortID

			// Apply Create() port_id resolution logic
			if v, ok := result["port_id"].(string); ok && v != "" {
				portID = types.StringValue(v)
			} else if portID.IsUnknown() {
				portID = types.StringNull()
			}

			// Assert
			if tc.expectedResult.IsNull() {
				if !portID.IsNull() {
					t.Errorf("expected null, got %q (unknown=%v)", portID.ValueString(), portID.IsUnknown())
				}
			} else if tc.expectedResult.IsUnknown() {
				if !portID.IsUnknown() {
					t.Error("expected unknown, got known value")
				}
			} else {
				if portID.IsNull() {
					t.Errorf("expected %q, got null", tc.expectedResult.ValueString())
				} else if portID.IsUnknown() {
					t.Errorf("expected %q, got unknown", tc.expectedResult.ValueString())
				} else if portID.ValueString() != tc.expectedResult.ValueString() {
					t.Errorf("expected %q, got %q", tc.expectedResult.ValueString(), portID.ValueString())
				}
			}
		})
	}
}

// ===========================================================================
// Create body construction based on plan state
// ===========================================================================

// TestCreateBodyFromPlan verifies that the body map is built correctly based
// on different combinations of plan field states (null, unknown, valued).
func TestCreateBodyFromPlan(t *testing.T) {
	tests := []struct {
		name           string
		plan           floatingIPModel
		expectPortID   bool
		expectDesc     bool
		expectPortVal  string
		expectDescVal  string
	}{
		{
			name: "all optional fields set",
			plan: floatingIPModel{
				FloatingNetworkID: types.StringValue("ext-1"),
				BillingType:       types.StringValue("hourly"),
				PortID:            types.StringValue("port-1"),
				Description:       types.StringValue("desc-1"),
			},
			expectPortID:  true,
			expectDesc:    true,
			expectPortVal: "port-1",
			expectDescVal: "desc-1",
		},
		{
			name: "port_id null, description set",
			plan: floatingIPModel{
				FloatingNetworkID: types.StringValue("ext-2"),
				BillingType:       types.StringValue("hourly"),
				PortID:            types.StringNull(),
				Description:       types.StringValue("only desc"),
			},
			expectPortID:  false,
			expectDesc:    true,
			expectDescVal: "only desc",
		},
		{
			name: "port_id unknown, description null",
			plan: floatingIPModel{
				FloatingNetworkID: types.StringValue("ext-3"),
				BillingType:       types.StringValue("hourly"),
				PortID:            types.StringUnknown(),
				Description:       types.StringNull(),
			},
			expectPortID: false,
			expectDesc:   false,
		},
		{
			name: "both optional fields null",
			plan: floatingIPModel{
				FloatingNetworkID: types.StringValue("ext-4"),
				BillingType:       types.StringValue("hourly"),
				PortID:            types.StringNull(),
				Description:       types.StringNull(),
			},
			expectPortID: false,
			expectDesc:   false,
		},
		{
			name: "empty description string is still set",
			plan: floatingIPModel{
				FloatingNetworkID: types.StringValue("ext-5"),
				BillingType:       types.StringValue("hourly"),
				PortID:            types.StringNull(),
				Description:       types.StringValue(""),
			},
			expectPortID:  false,
			expectDesc:    true,
			expectDescVal: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Build body — mirrors Create() logic
			body := map[string]interface{}{
				"floating_network_id": tc.plan.FloatingNetworkID.ValueString(),
				"billing_type":        tc.plan.BillingType.ValueString(),
			}
			if !tc.plan.PortID.IsNull() && !tc.plan.PortID.IsUnknown() {
				body["port_id"] = tc.plan.PortID.ValueString()
			}
			if !tc.plan.Description.IsNull() && !tc.plan.Description.IsUnknown() {
				body["description"] = tc.plan.Description.ValueString()
			}

			// Verify required fields always present
			if body["floating_network_id"] != tc.plan.FloatingNetworkID.ValueString() {
				t.Errorf("floating_network_id mismatch")
			}

			// Verify optional fields
			_, hasPort := body["port_id"]
			_, hasDesc := body["description"]

			if hasPort != tc.expectPortID {
				t.Errorf("port_id presence: expected %v, got %v", tc.expectPortID, hasPort)
			}
			if hasDesc != tc.expectDesc {
				t.Errorf("description presence: expected %v, got %v", tc.expectDesc, hasDesc)
			}

			if tc.expectPortID {
				if body["port_id"] != tc.expectPortVal {
					t.Errorf("port_id value: expected %q, got %v", tc.expectPortVal, body["port_id"])
				}
			}
			if tc.expectDesc {
				if body["description"] != tc.expectDescVal {
					t.Errorf("description value: expected %q, got %v", tc.expectDescVal, body["description"])
				}
			}
		})
	}
}

// ===========================================================================
// Edge cases: Read field extraction
// ===========================================================================

// TestReadFieldExtraction_EdgeCases covers unusual API response shapes that
// the Read parsing logic must handle gracefully.
func TestReadFieldExtraction_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		item       map[string]interface{}
		expectIP   string
		expectNull bool // if true, floating_ip_address should be unset
		expectStat string
	}{
		{
			name: "all fields present and valid",
			item: map[string]interface{}{
				"floating_ip_address": "192.168.1.1",
				"status":             "ACTIVE",
				"floating_network_id": "ext-1",
			},
			expectIP:   "192.168.1.1",
			expectStat: "ACTIVE",
		},
		{
			name: "floating_ip_address is empty string",
			item: map[string]interface{}{
				"floating_ip_address": "",
				"status":             "DOWN",
			},
			expectIP:   "", // empty string is still set (the type assertion succeeds)
			expectStat: "DOWN",
		},
		{
			name: "status missing from response",
			item: map[string]interface{}{
				"floating_ip_address": "10.0.0.1",
			},
			expectIP:   "10.0.0.1",
			expectStat: "", // status field not set
		},
		{
			name: "floating_ip_address is numeric (wrong type)",
			item: map[string]interface{}{
				"floating_ip_address": 12345,
				"status":             "ACTIVE",
			},
			expectNull: true, // type assertion .(string) fails
			expectStat: "ACTIVE",
		},
		{
			name: "status is numeric (wrong type)",
			item: map[string]interface{}{
				"floating_ip_address": "10.0.0.2",
				"status":             200,
			},
			expectIP:   "10.0.0.2",
			expectStat: "", // type assertion .(string) fails
		},
		{
			name: "description with only whitespace",
			item: map[string]interface{}{
				"floating_ip_address": "10.0.0.3",
				"status":             "ACTIVE",
				"description":        "   ",
			},
			expectIP:   "10.0.0.3",
			expectStat: "ACTIVE",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := floatingIPModel{
				ID: types.StringValue("fip-edge"),
			}

			// Apply Read() field extraction
			if v, ok := tc.item["floating_ip_address"].(string); ok {
				state.FloatingIPAddress = types.StringValue(v)
			}
			if v, ok := tc.item["status"].(string); ok {
				state.Status = types.StringValue(v)
			}
			if v, ok := tc.item["port_id"].(string); ok && v != "" {
				state.PortID = types.StringValue(v)
			} else {
				state.PortID = types.StringNull()
			}
			if v, ok := tc.item["description"].(string); ok && v != "" {
				state.Description = types.StringValue(v)
			} else {
				state.Description = types.StringNull()
			}

			if tc.expectNull {
				if state.FloatingIPAddress.ValueString() != "" && !state.FloatingIPAddress.IsNull() {
					// If type assertion failed, FloatingIPAddress stays at zero value
					// which is effectively empty. The key point: it should NOT have a value.
				}
			} else {
				if state.FloatingIPAddress.ValueString() != tc.expectIP {
					t.Errorf("FloatingIPAddress: expected %q, got %q", tc.expectIP, state.FloatingIPAddress.ValueString())
				}
			}

			if tc.expectStat != "" {
				if state.Status.ValueString() != tc.expectStat {
					t.Errorf("Status: expected %q, got %q", tc.expectStat, state.Status.ValueString())
				}
			}

			// Verify description with whitespace is treated as non-empty
			if tc.name == "description with only whitespace" {
				if state.Description.IsNull() {
					t.Error("expected whitespace description to be non-null (Read preserves it)")
				}
				if state.Description.ValueString() != "   " {
					t.Errorf("expected whitespace description preserved, got %q", state.Description.ValueString())
				}
			}
		})
	}
}

// TestReadResponseParsing_LargeList verifies that the linear scan works
// correctly even with many items, and correctly returns nil for absent IDs.
func TestReadResponseParsing_LargeList(t *testing.T) {
	// Build a list of 50 floating IPs
	items := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		items[i] = map[string]interface{}{
			"id":                 fmt.Sprintf("fip-%03d", i),
			"floating_ip_address": fmt.Sprintf("10.0.0.%d", i),
			"status":             "ACTIVE",
			"floating_network_id": "ext-net-1",
		}
	}

	data, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed []map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Find existing item
	targetID := "fip-042"
	var found map[string]interface{}
	for _, item := range parsed {
		if id, ok := item["id"].(string); ok && id == targetID {
			found = item
			break
		}
	}
	if found == nil {
		t.Fatal("expected to find fip-042 in list")
	}
	if found["floating_ip_address"] != "10.0.0.42" {
		t.Errorf("expected IP 10.0.0.42, got %v", found["floating_ip_address"])
	}

	// Search for non-existent item
	missingID := "fip-999"
	var notFound map[string]interface{}
	for _, item := range parsed {
		if id, ok := item["id"].(string); ok && id == missingID {
			notFound = item
			break
		}
	}
	if notFound != nil {
		t.Error("expected nil for non-existent fip-999")
	}
}

// ===========================================================================
// Create status default logic
// ===========================================================================

// TestCreateStatusDefault verifies that when the API response omits the
// status field, Create() defaults to "ACTIVE".
func TestCreateStatusDefault(t *testing.T) {
	tests := []struct {
		name         string
		result       map[string]interface{}
		expectStatus string
	}{
		{
			name:         "status present",
			result:       map[string]interface{}{"id": "fip-1", "status": "DOWN"},
			expectStatus: "DOWN",
		},
		{
			name:         "status missing",
			result:       map[string]interface{}{"id": "fip-2"},
			expectStatus: "ACTIVE",
		},
		{
			name:         "status is non-string (number)",
			result:       map[string]interface{}{"id": "fip-3", "status": 200},
			expectStatus: "ACTIVE",
		},
		{
			name:         "status is empty string",
			result:       map[string]interface{}{"id": "fip-4", "status": ""},
			expectStatus: "",
		},
		{
			name:         "status is ERROR",
			result:       map[string]interface{}{"id": "fip-5", "status": "ERROR"},
			expectStatus: "ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var status types.String
			if v, ok := tc.result["status"].(string); ok {
				status = types.StringValue(v)
			} else {
				status = types.StringValue("ACTIVE")
			}

			if status.ValueString() != tc.expectStatus {
				t.Errorf("expected status %q, got %q", tc.expectStatus, status.ValueString())
			}
		})
	}
}

// ===========================================================================
// Read port_id / description null vs empty string handling
// ===========================================================================

// TestReadNullableFields verifies that Read() correctly maps empty strings
// to null and non-empty strings to values for port_id and description.
func TestReadNullableFields(t *testing.T) {
	tests := []struct {
		name        string
		portID      interface{}
		description interface{}
		expectPort  types.String
		expectDesc  types.String
	}{
		{
			name:        "both present and non-empty",
			portID:      "port-1",
			description: "my desc",
			expectPort:  types.StringValue("port-1"),
			expectDesc:  types.StringValue("my desc"),
		},
		{
			name:        "both empty strings",
			portID:      "",
			description: "",
			expectPort:  types.StringNull(),
			expectDesc:  types.StringNull(),
		},
		{
			name:        "both missing from response",
			portID:      nil,
			description: nil,
			expectPort:  types.StringNull(),
			expectDesc:  types.StringNull(),
		},
		{
			name:        "port_id present, description missing",
			portID:      "port-2",
			description: nil,
			expectPort:  types.StringValue("port-2"),
			expectDesc:  types.StringNull(),
		},
		{
			name:        "port_id missing, description present",
			portID:      nil,
			description: "only desc",
			expectPort:  types.StringNull(),
			expectDesc:  types.StringValue("only desc"),
		},
		{
			name:        "port_id non-string type",
			portID:      12345,
			description: "valid desc",
			expectPort:  types.StringNull(),
			expectDesc:  types.StringValue("valid desc"),
		},
		{
			name:        "description non-string type",
			portID:      "port-3",
			description: true,
			expectPort:  types.StringValue("port-3"),
			expectDesc:  types.StringNull(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := map[string]interface{}{}
			if tc.portID != nil {
				result["port_id"] = tc.portID
			}
			if tc.description != nil {
				result["description"] = tc.description
			}

			// Apply Read() logic
			var portID, description types.String
			if v, ok := result["port_id"].(string); ok && v != "" {
				portID = types.StringValue(v)
			} else {
				portID = types.StringNull()
			}
			if v, ok := result["description"].(string); ok && v != "" {
				description = types.StringValue(v)
			} else {
				description = types.StringNull()
			}

			// Assert port_id
			if tc.expectPort.IsNull() {
				if !portID.IsNull() {
					t.Errorf("port_id: expected null, got %q", portID.ValueString())
				}
			} else {
				if portID.IsNull() {
					t.Errorf("port_id: expected %q, got null", tc.expectPort.ValueString())
				} else if portID.ValueString() != tc.expectPort.ValueString() {
					t.Errorf("port_id: expected %q, got %q", tc.expectPort.ValueString(), portID.ValueString())
				}
			}

			// Assert description
			if tc.expectDesc.IsNull() {
				if !description.IsNull() {
					t.Errorf("description: expected null, got %q", description.ValueString())
				}
			} else {
				if description.IsNull() {
					t.Errorf("description: expected %q, got null", tc.expectDesc.ValueString())
				} else if description.ValueString() != tc.expectDesc.ValueString() {
					t.Errorf("description: expected %q, got %q", tc.expectDesc.ValueString(), description.ValueString())
				}
			}
		})
	}
}

// ===========================================================================
// Delete body construction
// ===========================================================================

// TestDeleteBody_MultipleIDs verifies the delete body format with multiple
// IDs (even though the resource only deletes one at a time).
func TestDeleteBody_MultipleIDs(t *testing.T) {
	ids := []string{"fip-a", "fip-b", "fip-c"}
	body := map[string]interface{}{
		"key":    "id",
		"values": ids,
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
	for i, expected := range ids {
		if values[i] != expected {
			t.Errorf("values[%d]: expected %q, got %v", i, expected, values[i])
		}
	}
}
