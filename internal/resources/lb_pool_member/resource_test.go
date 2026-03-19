package lb_pool_member

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		PoolID:         types.StringValue("pool-123"),
		Name:           types.StringValue("backend-1"),
		Address:        types.StringValue("10.0.0.10"),
		ProtocolPort:   types.Int64Value(8080),
		Weight:         types.Int64Value(5),
		MonitorPort:    types.Int64Value(8081),
		MonitorAddress: types.StringValue("10.0.0.11"),
	}

	body := buildCreateRequest(plan)

	if len(body.BackendServers) != 1 {
		t.Fatalf("expected 1 backend server, got %d", len(body.BackendServers))
	}
	bs := body.BackendServers[0]
	if bs.Name != "backend-1" {
		t.Errorf("expected name backend-1, got %s", bs.Name)
	}
	if bs.Address != "10.0.0.10" {
		t.Errorf("expected address 10.0.0.10, got %s", bs.Address)
	}
	if bs.ProtocolPort != 8080 {
		t.Errorf("expected protocol_port 8080, got %d", bs.ProtocolPort)
	}
	if bs.Weight == nil || *bs.Weight != 5 {
		t.Errorf("expected weight 5, got %v", bs.Weight)
	}
	if bs.MonitorPort != 8081 {
		t.Errorf("expected monitor_port 8081, got %d", bs.MonitorPort)
	}
	if bs.MonitorAddress != "10.0.0.11" {
		t.Errorf("expected monitor_address 10.0.0.11, got %s", bs.MonitorAddress)
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		PoolID:         types.StringValue("pool-456"),
		Name:           types.StringNull(),
		Address:        types.StringValue("10.0.0.20"),
		ProtocolPort:   types.Int64Value(80),
		Weight:         types.Int64Null(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	body := buildCreateRequest(plan)

	bs := body.BackendServers[0]
	if bs.Name != "" {
		t.Errorf("expected empty name, got %s", bs.Name)
	}
	if bs.Weight != nil {
		t.Errorf("expected weight nil (null plan), got %v", bs.Weight)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		PoolID:         types.StringValue("pool-789"),
		Name:           types.StringValue("web-server"),
		Address:        types.StringValue("192.168.1.10"),
		ProtocolPort:   types.Int64Value(443),
		Weight:         types.Int64Value(3),
		MonitorPort:    types.Int64Value(8443),
		MonitorAddress: types.StringValue("192.168.1.11"),
	}

	body := buildCreateRequest(plan)

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Must have the backend_servers wrapper key
	if !strings.Contains(jsonStr, `"backend_servers"`) {
		t.Errorf("expected JSON to contain 'backend_servers' key, got: %s", jsonStr)
	}

	// Parse back and verify structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	serversRaw, ok := parsed["backend_servers"]
	if !ok {
		t.Fatal("expected 'backend_servers' key in JSON")
	}

	servers, ok := serversRaw.([]interface{})
	if !ok {
		t.Fatalf("expected backend_servers to be an array, got %T", serversRaw)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server in array, got %d", len(servers))
	}

	server, ok := servers[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected server to be a map")
	}

	if server["name"] != "web-server" {
		t.Errorf("expected name 'web-server', got %v", server["name"])
	}
	if server["address"] != "192.168.1.10" {
		t.Errorf("expected address '192.168.1.10', got %v", server["address"])
	}
	if server["protocol_port"] != float64(443) {
		t.Errorf("expected protocol_port 443, got %v", server["protocol_port"])
	}
	if server["weight"] != float64(3) {
		t.Errorf("expected weight 3, got %v", server["weight"])
	}
	if server["monitor_port"] != float64(8443) {
		t.Errorf("expected monitor_port 8443, got %v", server["monitor_port"])
	}
	if server["monitor_address"] != "192.168.1.11" {
		t.Errorf("expected monitor_address '192.168.1.11', got %v", server["monitor_address"])
	}
}

func TestBuildUpdateRequest(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		Weight:         types.Int64Value(10),
		MonitorPort:    types.Int64Value(9090),
		MonitorAddress: types.StringValue("10.0.0.50"),
	}

	body := buildUpdateRequest(plan)

	if body.Weight == nil || *body.Weight != 10 {
		t.Errorf("expected weight 10, got %v", body.Weight)
	}
	if body.MonitorPort != 9090 {
		t.Errorf("expected monitor_port 9090, got %d", body.MonitorPort)
	}
	if body.MonitorAddress != "10.0.0.50" {
		t.Errorf("expected monitor_address 10.0.0.50, got %s", body.MonitorAddress)
	}
}

func TestBuildUpdateRequest_MinimalFields(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		Weight:         types.Int64Value(7),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	body := buildUpdateRequest(plan)

	if body.Weight == nil || *body.Weight != 7 {
		t.Errorf("expected weight 7, got %v", body.Weight)
	}
	if body.MonitorPort != 0 {
		t.Errorf("expected monitor_port 0 (omitted), got %d", body.MonitorPort)
	}
	if body.MonitorAddress != "" {
		t.Errorf("expected monitor_address '' (omitted), got %s", body.MonitorAddress)
	}
}

func TestBuildUpdateRequest_JSON(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		Weight:         types.Int64Value(8),
		MonitorPort:    types.Int64Value(9999),
		MonitorAddress: types.StringValue("10.0.0.99"),
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

	if parsed["weight"] != float64(8) {
		t.Errorf("expected weight 8, got %v", parsed["weight"])
	}
	if parsed["monitor_port"] != float64(9999) {
		t.Errorf("expected monitor_port 9999, got %v", parsed["monitor_port"])
	}
	if parsed["monitor_address"] != "10.0.0.99" {
		t.Errorf("expected monitor_address '10.0.0.99', got %v", parsed["monitor_address"])
	}

	// Should NOT have backend_servers wrapper
	if _, exists := parsed["backend_servers"]; exists {
		t.Error("update request should not have backend_servers wrapper")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:                 "member-abc-123",
		Name:               "backend-server-1",
		Address:            "10.0.0.10",
		ProtocolPort:       8080,
		Weight:             5,
		MonitorPort:        8081,
		MonitorAddress:     "10.0.0.11",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		AdminStateUp:       true,
		CreatedAt:          "2024-01-01T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "member-abc-123" {
		t.Errorf("expected ID member-abc-123, got %s", model.ID.ValueString())
	}
	if model.Weight.ValueInt64() != 5 {
		t.Errorf("expected Weight 5, got %d", model.Weight.ValueInt64())
	}
	if model.AdminStateUp.ValueBool() != true {
		t.Error("expected AdminStateUp to be true")
	}
}

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:                 "member-full-456",
		Name:               "full-backend",
		Address:            "172.16.0.100",
		ProtocolPort:       9090,
		Weight:             10,
		MonitorPort:        9091,
		MonitorAddress:     "172.16.0.101",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		AdminStateUp:       true,
		CreatedAt:          "2024-06-15T12:30:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "member-full-456" {
		t.Errorf("expected ID member-full-456, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-backend" {
		t.Errorf("expected Name full-backend, got %s", model.Name.ValueString())
	}
	if model.Address.ValueString() != "172.16.0.100" {
		t.Errorf("expected Address 172.16.0.100, got %s", model.Address.ValueString())
	}
	if model.ProtocolPort.ValueInt64() != 9090 {
		t.Errorf("expected ProtocolPort 9090, got %d", model.ProtocolPort.ValueInt64())
	}
	if model.Weight.ValueInt64() != 10 {
		t.Errorf("expected Weight 10, got %d", model.Weight.ValueInt64())
	}
	if model.MonitorPort.ValueInt64() != 9091 {
		t.Errorf("expected MonitorPort 9091, got %d", model.MonitorPort.ValueInt64())
	}
	if model.MonitorAddress.ValueString() != "172.16.0.101" {
		t.Errorf("expected MonitorAddress 172.16.0.101, got %s", model.MonitorAddress.ValueString())
	}
	if model.ProvisioningStatus.ValueString() != "ACTIVE" {
		t.Errorf("expected ProvisioningStatus ACTIVE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "ONLINE" {
		t.Errorf("expected OperatingStatus ONLINE, got %s", model.OperatingStatus.ValueString())
	}
	if model.AdminStateUp.ValueBool() != true {
		t.Error("expected AdminStateUp to be true")
	}
	if model.CreatedAt.ValueString() != "2024-06-15T12:30:00Z" {
		t.Errorf("expected CreatedAt 2024-06-15T12:30:00Z, got %s", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:           "member-123",
		Address:      "10.0.0.5",
		ProtocolPort: 80,
		Weight:       1,
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Name.IsNull() {
		t.Error("expected Name to remain null")
	}
	if !model.MonitorPort.IsNull() {
		t.Error("expected MonitorPort to remain null")
	}
	if !model.MonitorAddress.IsNull() {
		t.Error("expected MonitorAddress to remain null")
	}
}

func TestMemberDeleteRequest(t *testing.T) {
	req := memberDeleteRequest{
		Key:    "id",
		Values: []string{"member-1"},
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
	if values[0] != "member-1" {
		t.Errorf("expected value 'member-1', got %v", values[0])
	}

	// Verify exact JSON format
	expectedJSON := `{"key":"id","values":["member-1"]}`
	if string(jsonBytes) != expectedJSON {
		t.Errorf("expected JSON %s, got %s", expectedJSON, string(jsonBytes))
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- mapAPIResponseToState edge cases ---

func TestMapAPIResponseToState_AdminStateUpFalse(t *testing.T) {
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:           "member-admin-false",
		Address:      "10.0.0.1",
		ProtocolPort: 80,
		Weight:       1,
		AdminStateUp: false,
	}

	mapAPIResponseToState(model, apiResp)

	if model.AdminStateUp.ValueBool() != false {
		t.Error("expected AdminStateUp to be false")
	}
}

func TestMapAPIResponseToState_ZeroWeight(t *testing.T) {
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:           "member-zero-weight",
		Address:      "10.0.0.1",
		ProtocolPort: 80,
		Weight:       0,
		AdminStateUp: true,
	}

	mapAPIResponseToState(model, apiResp)

	// Weight is always set (no conditional check), so 0 is valid
	if model.Weight.ValueInt64() != 0 {
		t.Errorf("expected Weight 0, got %d", model.Weight.ValueInt64())
	}
}

func TestMapAPIResponseToState_NameSetThenEmpty(t *testing.T) {
	// Model already has a name set (non-null), API returns empty
	model := &lbPoolMemberResourceModel{
		Name:           types.StringValue("existing-name"),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:           "member-name-empty",
		Name:         "",
		Address:      "10.0.0.1",
		ProtocolPort: 80,
		Weight:       1,
		AdminStateUp: true,
	}

	mapAPIResponseToState(model, apiResp)

	// When apiResp.Name is empty and model.Name is NOT null, it stays at original value
	if model.Name.ValueString() != "existing-name" {
		t.Errorf("expected Name to remain existing-name, got %s", model.Name.ValueString())
	}
}

func TestMapAPIResponseToState_MonitorPortSetThenZero(t *testing.T) {
	// Model has existing monitor_port set, API returns 0
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Value(9090),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:           "member-monitor-zero",
		Address:      "10.0.0.1",
		ProtocolPort: 80,
		Weight:       1,
		MonitorPort:  0,
		AdminStateUp: true,
	}

	mapAPIResponseToState(model, apiResp)

	// MonitorPort > 0 check fails, else-if checks IsNull() which is false (was set),
	// so it stays at original value
	if model.MonitorPort.ValueInt64() != 9090 {
		t.Errorf("expected MonitorPort to remain 9090, got %d", model.MonitorPort.ValueInt64())
	}
}

func TestMapAPIResponseToState_MonitorAddressSetThenEmpty(t *testing.T) {
	// Model has existing monitor_address, API returns empty
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringValue("10.0.0.50"),
	}

	apiResp := &memberAPIResponse{
		ID:             "member-addr-empty",
		Address:        "10.0.0.1",
		ProtocolPort:   80,
		Weight:         1,
		MonitorAddress: "",
		AdminStateUp:   true,
	}

	mapAPIResponseToState(model, apiResp)

	// MonitorAddress is empty, model is not null, so stays at original
	if model.MonitorAddress.ValueString() != "10.0.0.50" {
		t.Errorf("expected MonitorAddress to remain 10.0.0.50, got %s", model.MonitorAddress.ValueString())
	}
}

func TestMapAPIResponseToState_AllStatusFieldsEmpty(t *testing.T) {
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:           "member-no-status",
		Address:      "10.0.0.1",
		ProtocolPort: 80,
		Weight:       1,
	}

	mapAPIResponseToState(model, apiResp)

	if !model.ProvisioningStatus.IsNull() {
		t.Error("expected ProvisioningStatus to be null when API returns empty")
	}
	if !model.OperatingStatus.IsNull() {
		t.Error("expected OperatingStatus to be null when API returns empty")
	}
	if !model.CreatedAt.IsNull() {
		t.Error("expected CreatedAt to be null when API returns empty")
	}
}

func TestMapAPIResponseToState_PendingStatuses(t *testing.T) {
	model := &lbPoolMemberResourceModel{
		Name:           types.StringNull(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	apiResp := &memberAPIResponse{
		ID:                 "member-pending",
		Address:            "10.0.0.1",
		ProtocolPort:       80,
		Weight:             1,
		ProvisioningStatus: "PENDING_UPDATE",
		OperatingStatus:    "NO_MONITOR",
		AdminStateUp:       true,
	}

	mapAPIResponseToState(model, apiResp)

	if model.ProvisioningStatus.ValueString() != "PENDING_UPDATE" {
		t.Errorf("expected ProvisioningStatus PENDING_UPDATE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "NO_MONITOR" {
		t.Errorf("expected OperatingStatus NO_MONITOR, got %s", model.OperatingStatus.ValueString())
	}
}

// --- buildCreateRequest edge cases ---

func TestBuildCreateRequest_ZeroWeight(t *testing.T) {
	// Weight of 0 is technically valid (means no traffic)
	plan := &lbPoolMemberResourceModel{
		PoolID:         types.StringValue("pool-zero"),
		Name:           types.StringNull(),
		Address:        types.StringValue("10.0.0.1"),
		ProtocolPort:   types.Int64Value(80),
		Weight:         types.Int64Value(0),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	body := buildCreateRequest(plan)
	bs := body.BackendServers[0]

	// Weight 0 is not null/unknown, so it should be set via pointer
	if bs.Weight == nil || *bs.Weight != 0 {
		t.Errorf("expected weight 0 (pointer), got %v", bs.Weight)
	}
}

func TestBuildCreateRequest_HighWeight(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		PoolID:         types.StringValue("pool-high"),
		Name:           types.StringNull(),
		Address:        types.StringValue("10.0.0.1"),
		ProtocolPort:   types.Int64Value(80),
		Weight:         types.Int64Value(256),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	body := buildCreateRequest(plan)
	bs := body.BackendServers[0]

	if bs.Weight == nil || *bs.Weight != 256 {
		t.Errorf("expected weight 256, got %v", bs.Weight)
	}
}

// --- buildUpdateRequest edge cases ---

func TestBuildUpdateRequest_AllNull(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		Weight:         types.Int64Null(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	body := buildUpdateRequest(plan)

	if body.Weight != nil {
		t.Errorf("expected weight nil when null, got %v", body.Weight)
	}
	if body.MonitorPort != 0 {
		t.Errorf("expected monitor_port 0 when null, got %d", body.MonitorPort)
	}
	if body.MonitorAddress != "" {
		t.Errorf("expected empty monitor_address when null, got %s", body.MonitorAddress)
	}
}

func TestBuildUpdateRequest_JSONOmitEmpty(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		Weight:         types.Int64Null(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	body := buildUpdateRequest(plan)
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// All fields have omitempty, so zero values should be omitted
	if _, exists := parsed["weight"]; exists {
		t.Error("expected weight to be omitted when 0 (omitempty)")
	}
	if _, exists := parsed["monitor_port"]; exists {
		t.Error("expected monitor_port to be omitted when 0 (omitempty)")
	}
	if _, exists := parsed["monitor_address"]; exists {
		t.Error("expected monitor_address to be omitted when empty (omitempty)")
	}
}

// --- API response JSON tests ---

func TestMemberAPIResponse_JSONDeserialization(t *testing.T) {
	raw := `{
		"id": "mem-rt-001",
		"name": "roundtrip-member",
		"address": "192.168.1.100",
		"protocol_port": 8443,
		"weight": 10,
		"monitor_port": 8444,
		"monitor_address": "192.168.1.101",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"admin_state_up": true,
		"created_at": "2025-03-01T10:00:00Z"
	}`

	var resp memberAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "mem-rt-001" {
		t.Errorf("expected ID mem-rt-001, got %s", resp.ID)
	}
	if resp.Address != "192.168.1.100" {
		t.Errorf("expected Address 192.168.1.100, got %s", resp.Address)
	}
	if resp.ProtocolPort != 8443 {
		t.Errorf("expected ProtocolPort 8443, got %d", resp.ProtocolPort)
	}
	if resp.Weight != 10 {
		t.Errorf("expected Weight 10, got %d", resp.Weight)
	}
	if resp.MonitorPort != 8444 {
		t.Errorf("expected MonitorPort 8444, got %d", resp.MonitorPort)
	}
	if resp.MonitorAddress != "192.168.1.101" {
		t.Errorf("expected MonitorAddress 192.168.1.101, got %s", resp.MonitorAddress)
	}
	if resp.AdminStateUp != true {
		t.Error("expected AdminStateUp true")
	}
}

func TestMemberAPIResponse_JSONWithExtraFields(t *testing.T) {
	raw := `{
		"id": "mem-extra-001",
		"name": "extra",
		"address": "10.0.0.1",
		"protocol_port": 80,
		"weight": 1,
		"admin_state_up": true,
		"unknown_field": "should-be-ignored"
	}`

	var resp memberAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal with extra fields: %v", err)
	}

	if resp.ID != "mem-extra-001" {
		t.Errorf("expected ID mem-extra-001, got %s", resp.ID)
	}
}

func TestMemberAPIResponse_JSONAdminStateUpFalse(t *testing.T) {
	raw := `{
		"id": "mem-admin-false",
		"address": "10.0.0.1",
		"protocol_port": 80,
		"weight": 1,
		"admin_state_up": false
	}`

	var resp memberAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.AdminStateUp != false {
		t.Error("expected AdminStateUp false")
	}
}

// --- JSON wrapper format tests ---

func TestBuildCreateRequest_JSONMinimalOmitsEmptyFields(t *testing.T) {
	plan := &lbPoolMemberResourceModel{
		PoolID:         types.StringValue("pool-minimal"),
		Name:           types.StringNull(),
		Address:        types.StringValue("10.0.0.1"),
		ProtocolPort:   types.Int64Value(80),
		Weight:         types.Int64Null(),
		MonitorPort:    types.Int64Null(),
		MonitorAddress: types.StringNull(),
	}

	body := buildCreateRequest(plan)
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Must have backend_servers wrapper
	if !strings.Contains(jsonStr, `"backend_servers"`) {
		t.Error("expected backend_servers wrapper in JSON")
	}

	// Should not contain omitted fields with zero values
	if strings.Contains(jsonStr, `"name"`) {
		t.Error("expected name to be omitted from JSON when null")
	}
	if strings.Contains(jsonStr, `"monitor_port"`) {
		t.Error("expected monitor_port to be omitted from JSON when null")
	}
	if strings.Contains(jsonStr, `"monitor_address"`) {
		t.Error("expected monitor_address to be omitted from JSON when null")
	}
}

// --- Schema tests ---

func TestLbPoolMemberSchema(t *testing.T) {
	s := lbPoolMemberSchema()

	if s.Description == "" {
		t.Error("expected non-empty schema description")
	}

	// Required attributes
	requiredStr := []string{"pool_id", "address"}
	for _, attr := range requiredStr {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
			continue
		}
		sa, ok := a.(schema.StringAttribute)
		if ok && !sa.IsRequired() {
			t.Errorf("expected attribute %q to be required", attr)
		}
	}

	// protocol_port is required (Int64)
	ppAttr, ok := s.Attributes["protocol_port"]
	if !ok {
		t.Fatal("expected protocol_port attribute to exist")
	}
	ia, ok := ppAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("expected protocol_port to be Int64Attribute")
	}
	if !ia.IsRequired() {
		t.Error("expected protocol_port to be required")
	}

	// Optional attributes
	optionalStr := []string{"name", "monitor_address"}
	for _, attr := range optionalStr {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
			continue
		}
		sa, ok := a.(schema.StringAttribute)
		if ok && !sa.IsOptional() {
			t.Errorf("expected attribute %q to be optional", attr)
		}
	}

	// Computed attributes
	computed := []string{"id", "provisioning_status", "operating_status", "admin_state_up", "created_at"}
	for _, attr := range computed {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
		_ = a // existence check is sufficient
	}

	// Weight should be optional + computed with default
	wAttr, ok := s.Attributes["weight"]
	if !ok {
		t.Fatal("expected weight attribute to exist")
	}
	wia, ok := wAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("expected weight to be Int64Attribute")
	}
	if !wia.IsOptional() {
		t.Error("expected weight to be optional")
	}
	if !wia.IsComputed() {
		t.Error("expected weight to be computed")
	}
}

// --- Delete request multiple values ---

func TestMemberDeleteRequest_MultipleValues(t *testing.T) {
	req := memberDeleteRequest{
		Key:    "id",
		Values: []string{"member-1", "member-2", "member-3"},
	}

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	values := parsed["values"].([]interface{})
	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}
	if values[0] != "member-1" {
		t.Errorf("expected first value member-1, got %v", values[0])
	}
	if values[2] != "member-3" {
		t.Errorf("expected third value member-3, got %v", values[2])
	}
}
