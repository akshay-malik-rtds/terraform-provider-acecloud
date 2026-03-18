package lb_listener

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest(t *testing.T) {
	plan := &lbListenerResourceModel{
		Name:           types.StringValue("test-listener"),
		Protocol:       types.StringValue("HTTP"),
		ProtocolPort:   types.Int64Value(80),
		LoadBalancerID: types.StringValue("lb-123"),
	}

	body := buildCreateRequest(plan)

	if body.Name != "test-listener" {
		t.Errorf("expected name test-listener, got %s", body.Name)
	}
	if body.Protocol != "HTTP" {
		t.Errorf("expected protocol HTTP, got %s", body.Protocol)
	}
	if body.ProtocolPort != 80 {
		t.Errorf("expected protocol_port 80, got %d", body.ProtocolPort)
	}
	if body.LoadBalancerID != "lb-123" {
		t.Errorf("expected loadbalancer_id lb-123, got %s", body.LoadBalancerID)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &lbListenerResourceModel{
		Name:           types.StringValue("json-listener"),
		Protocol:       types.StringValue("HTTPS"),
		ProtocolPort:   types.Int64Value(443),
		LoadBalancerID: types.StringValue("lb-json-001"),
	}

	body := buildCreateRequest(plan)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify JSON field names match API expectations
	if _, ok := raw["name"]; !ok {
		t.Error("expected JSON key 'name' to be present")
	}
	if _, ok := raw["protocol"]; !ok {
		t.Error("expected JSON key 'protocol' to be present")
	}
	if _, ok := raw["protocol_port"]; !ok {
		t.Error("expected JSON key 'protocol_port' to be present")
	}
	if _, ok := raw["loadbalancer_id"]; !ok {
		t.Error("expected JSON key 'loadbalancer_id' to be present")
	}

	if raw["name"] != "json-listener" {
		t.Errorf("expected JSON name 'json-listener', got %v", raw["name"])
	}
	if raw["protocol"] != "HTTPS" {
		t.Errorf("expected JSON protocol 'HTTPS', got %v", raw["protocol"])
	}
	// JSON numbers unmarshal as float64
	if raw["protocol_port"] != float64(443) {
		t.Errorf("expected JSON protocol_port 443, got %v", raw["protocol_port"])
	}
	if raw["loadbalancer_id"] != "lb-json-001" {
		t.Errorf("expected JSON loadbalancer_id 'lb-json-001', got %v", raw["loadbalancer_id"])
	}
}

func TestBuildCreateRequest_AllProtocols(t *testing.T) {
	protocols := []struct {
		protocol string
		port     int64
	}{
		{"HTTP", 80},
		{"HTTPS", 443},
		{"TCP", 8080},
		{"UDP", 5353},
		{"TERMINATED_HTTPS", 8443},
	}

	for _, tc := range protocols {
		t.Run(tc.protocol, func(t *testing.T) {
			plan := &lbListenerResourceModel{
				Name:           types.StringValue("listener-" + tc.protocol),
				Protocol:       types.StringValue(tc.protocol),
				ProtocolPort:   types.Int64Value(tc.port),
				LoadBalancerID: types.StringValue("lb-proto-test"),
			}

			body := buildCreateRequest(plan)

			if body.Protocol != tc.protocol {
				t.Errorf("expected protocol %s, got %s", tc.protocol, body.Protocol)
			}
			if body.ProtocolPort != tc.port {
				t.Errorf("expected protocol_port %d, got %d", tc.port, body.ProtocolPort)
			}
			if body.Name != "listener-"+tc.protocol {
				t.Errorf("expected name 'listener-%s', got %s", tc.protocol, body.Name)
			}
		})
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &lbListenerResourceModel{}

	apiResp := &listenerAPIResponse{
		ID:                 "listener-abc-123",
		Name:               "prod-listener",
		Protocol:           "HTTPS",
		ProtocolPort:       443,
		LoadBalancerID:     "lb-xyz",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		DefaultPoolID:      "pool-default",
		CreatedAt:          "2024-01-01T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "listener-abc-123" {
		t.Errorf("expected ID listener-abc-123, got %s", model.ID.ValueString())
	}
	if model.Protocol.ValueString() != "HTTPS" {
		t.Errorf("expected Protocol HTTPS, got %s", model.Protocol.ValueString())
	}
	if model.ProtocolPort.ValueInt64() != 443 {
		t.Errorf("expected ProtocolPort 443, got %d", model.ProtocolPort.ValueInt64())
	}
	if model.ProvisioningStatus.ValueString() != "ACTIVE" {
		t.Errorf("expected ProvisioningStatus ACTIVE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.DefaultPoolID.ValueString() != "pool-default" {
		t.Errorf("expected DefaultPoolID pool-default, got %s", model.DefaultPoolID.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &lbListenerResourceModel{}

	apiResp := &listenerAPIResponse{
		ID:                 "listener-empty-001",
		Name:               "empty-listener",
		Protocol:           "TCP",
		ProtocolPort:       8080,
		LoadBalancerID:     "lb-empty-xyz",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		DefaultPoolID:      "",
		CreatedAt:          "2024-03-01T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "listener-empty-001" {
		t.Errorf("expected ID listener-empty-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "empty-listener" {
		t.Errorf("expected Name empty-listener, got %s", model.Name.ValueString())
	}
	if model.Protocol.ValueString() != "TCP" {
		t.Errorf("expected Protocol TCP, got %s", model.Protocol.ValueString())
	}
	if model.ProtocolPort.ValueInt64() != 8080 {
		t.Errorf("expected ProtocolPort 8080, got %d", model.ProtocolPort.ValueInt64())
	}
	if model.LoadBalancerID.ValueString() != "lb-empty-xyz" {
		t.Errorf("expected LoadBalancerID lb-empty-xyz, got %s", model.LoadBalancerID.ValueString())
	}
	// DefaultPoolID should not be set when API returns empty string
	if model.DefaultPoolID.ValueString() != "" {
		t.Errorf("expected DefaultPoolID to be empty when API returns empty string, got %s", model.DefaultPoolID.ValueString())
	}
}

func TestListenerDeleteRequest(t *testing.T) {
	req := listenerDeleteRequest{
		Key:    "id",
		Values: []string{"listener-del-001"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal delete request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if raw["key"] != "id" {
		t.Errorf("expected key 'id', got %v", raw["key"])
	}

	values, ok := raw["values"].([]interface{})
	if !ok {
		t.Fatal("expected 'values' to be an array")
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0] != "listener-del-001" {
		t.Errorf("expected value 'listener-del-001', got %v", values[0])
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- parseListenerFromRaw tests ---

func TestParseListenerFromRaw_DirectFields(t *testing.T) {
	raw := `{
		"id": "listener-001",
		"name": "my-listener",
		"protocol": "HTTP",
		"protocol_port": 80,
		"loadbalancer_id": "lb-direct-001",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"default_pool_id": "pool-001",
		"created_at": "2024-06-01T12:00:00Z"
	}`

	listener, err := parseListenerFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listener.ID != "listener-001" {
		t.Errorf("expected ID listener-001, got %s", listener.ID)
	}
	if listener.Name != "my-listener" {
		t.Errorf("expected Name my-listener, got %s", listener.Name)
	}
	if listener.Protocol != "HTTP" {
		t.Errorf("expected Protocol HTTP, got %s", listener.Protocol)
	}
	if listener.ProtocolPort != 80 {
		t.Errorf("expected ProtocolPort 80, got %d", listener.ProtocolPort)
	}
	if listener.LoadBalancerID != "lb-direct-001" {
		t.Errorf("expected LoadBalancerID lb-direct-001, got %s", listener.LoadBalancerID)
	}
	if listener.ProvisioningStatus != "ACTIVE" {
		t.Errorf("expected ProvisioningStatus ACTIVE, got %s", listener.ProvisioningStatus)
	}
	if listener.OperatingStatus != "ONLINE" {
		t.Errorf("expected OperatingStatus ONLINE, got %s", listener.OperatingStatus)
	}
	if listener.DefaultPoolID != "pool-001" {
		t.Errorf("expected DefaultPoolID pool-001, got %s", listener.DefaultPoolID)
	}
	if listener.CreatedAt != "2024-06-01T12:00:00Z" {
		t.Errorf("expected CreatedAt 2024-06-01T12:00:00Z, got %s", listener.CreatedAt)
	}
}

func TestParseListenerFromRaw_LoadbalancersArray(t *testing.T) {
	// API returns loadbalancers:[{id:...}] instead of loadbalancer_id
	raw := `{
		"id": "listener-002",
		"name": "array-listener",
		"protocol": "TCP",
		"protocol_port": 443,
		"loadbalancers": [{"id": "lb-from-array"}],
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE"
	}`

	listener, err := parseListenerFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listener.LoadBalancerID != "lb-from-array" {
		t.Errorf("expected LoadBalancerID lb-from-array from loadbalancers array, got %s", listener.LoadBalancerID)
	}
}

func TestParseListenerFromRaw_EmptyLoadbalancersArray(t *testing.T) {
	raw := `{
		"id": "listener-003",
		"name": "empty-lb-array",
		"protocol": "HTTP",
		"protocol_port": 80,
		"loadbalancers": []
	}`

	listener, err := parseListenerFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listener.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID when loadbalancers array is empty, got %s", listener.LoadBalancerID)
	}
}

func TestParseListenerFromRaw_MultipleLoadbalancers(t *testing.T) {
	// Should pick the first one
	raw := `{
		"id": "listener-004",
		"name": "multi-lb-listener",
		"protocol": "HTTP",
		"protocol_port": 80,
		"loadbalancers": [{"id": "lb-first"}, {"id": "lb-second"}]
	}`

	listener, err := parseListenerFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listener.LoadBalancerID != "lb-first" {
		t.Errorf("expected LoadBalancerID lb-first (first in array), got %s", listener.LoadBalancerID)
	}
}

func TestParseListenerFromRaw_DirectIDTakesPrecedence(t *testing.T) {
	// When both loadbalancer_id and loadbalancers array exist, direct ID takes precedence
	raw := `{
		"id": "listener-005",
		"name": "both-listener",
		"protocol": "HTTP",
		"protocol_port": 80,
		"loadbalancer_id": "lb-direct",
		"loadbalancers": [{"id": "lb-array"}]
	}`

	listener, err := parseListenerFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listener.LoadBalancerID != "lb-direct" {
		t.Errorf("expected LoadBalancerID lb-direct (direct field), got %s", listener.LoadBalancerID)
	}
}

func TestParseListenerFromRaw_InvalidJSON(t *testing.T) {
	_, err := parseListenerFromRaw(json.RawMessage(`{invalid json}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseListenerFromRaw_EmptyObject(t *testing.T) {
	listener, err := parseListenerFromRaw(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listener.ID != "" {
		t.Errorf("expected empty ID, got %s", listener.ID)
	}
	if listener.Name != "" {
		t.Errorf("expected empty Name, got %s", listener.Name)
	}
	if listener.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID, got %s", listener.LoadBalancerID)
	}
}

func TestParseListenerFromRaw_LoadbalancersMalformed(t *testing.T) {
	// loadbalancers contains non-object entries
	raw := `{
		"id": "listener-006",
		"name": "malformed-lb",
		"protocol": "HTTP",
		"protocol_port": 80,
		"loadbalancers": ["not-an-object"]
	}`

	listener, err := parseListenerFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should gracefully handle and return empty LB ID
	if listener.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID for malformed array entry, got %s", listener.LoadBalancerID)
	}
}

func TestParseListenerFromRaw_LoadbalancersMissingID(t *testing.T) {
	raw := `{
		"id": "listener-007",
		"name": "no-id-lb",
		"protocol": "HTTP",
		"protocol_port": 80,
		"loadbalancers": [{"name": "lb-without-id"}]
	}`

	listener, err := parseListenerFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listener.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID when lb object has no id field, got %s", listener.LoadBalancerID)
	}
}

// --- Name regex tests ---

func TestListenerNameRegex(t *testing.T) {
	valid := []string{
		"my-listener",
		"listener1",
		"a",
		"A-B-C",
		"test-123-prod",
		"UPPERCASE",
		"MiXeD-CaSe-123",
	}

	for _, name := range valid {
		t.Run("valid_"+name, func(t *testing.T) {
			if !listenerNameRegex.MatchString(name) {
				t.Errorf("expected %q to be valid", name)
			}
		})
	}

	invalid := []string{
		"has spaces",
		"under_score",
		"dot.name",
		"special@char",
		"path/slash",
		"colon:name",
		"",
	}

	for _, name := range invalid {
		label := name
		if label == "" {
			label = "empty"
		}
		t.Run("invalid_"+label, func(t *testing.T) {
			if listenerNameRegex.MatchString(name) {
				t.Errorf("expected %q to be invalid", name)
			}
		})
	}
}

// --- mapAPIResponseToState edge cases ---

func TestMapAPIResponseToState_AllFieldsEmpty(t *testing.T) {
	model := &lbListenerResourceModel{}

	apiResp := &listenerAPIResponse{
		ID:   "listener-empty-all",
		Name: "empty-all",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "listener-empty-all" {
		t.Errorf("expected ID listener-empty-all, got %s", model.ID.ValueString())
	}
	// Protocol, ProtocolPort, LoadBalancerID should remain zero-value (not set)
	// because conditions check for non-empty/non-zero
	if model.ProvisioningStatus.ValueString() != "" {
		if !model.ProvisioningStatus.IsNull() {
			t.Errorf("expected ProvisioningStatus to be null, got %s", model.ProvisioningStatus.ValueString())
		}
	}
	if !model.OperatingStatus.IsNull() {
		t.Error("expected OperatingStatus to be null when API returns empty")
	}
	if !model.DefaultPoolID.IsNull() {
		t.Error("expected DefaultPoolID to be null when API returns empty")
	}
	if !model.CreatedAt.IsNull() {
		t.Error("expected CreatedAt to be null when API returns empty")
	}
}

func TestMapAPIResponseToState_PendingStatuses(t *testing.T) {
	model := &lbListenerResourceModel{}

	apiResp := &listenerAPIResponse{
		ID:                 "listener-pending",
		Name:               "pending-listener",
		Protocol:           "HTTP",
		ProtocolPort:       80,
		LoadBalancerID:     "lb-pending",
		ProvisioningStatus: "PENDING_CREATE",
		OperatingStatus:    "OFFLINE",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ProvisioningStatus.ValueString() != "PENDING_CREATE" {
		t.Errorf("expected ProvisioningStatus PENDING_CREATE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "OFFLINE" {
		t.Errorf("expected OperatingStatus OFFLINE, got %s", model.OperatingStatus.ValueString())
	}
}

// --- API response JSON round-trip ---

func TestListenerAPIResponse_JSONDeserialization(t *testing.T) {
	raw := `{
		"id": "lst-roundtrip-001",
		"name": "roundtrip-listener",
		"protocol": "TERMINATED_HTTPS",
		"protocol_port": 8443,
		"loadbalancer_id": "lb-roundtrip",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"default_pool_id": "pool-rt-001",
		"created_at": "2025-01-15T09:30:00Z"
	}`

	var resp listenerAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "lst-roundtrip-001" {
		t.Errorf("expected ID lst-roundtrip-001, got %s", resp.ID)
	}
	if resp.Protocol != "TERMINATED_HTTPS" {
		t.Errorf("expected Protocol TERMINATED_HTTPS, got %s", resp.Protocol)
	}
	if resp.ProtocolPort != 8443 {
		t.Errorf("expected ProtocolPort 8443, got %d", resp.ProtocolPort)
	}
	if resp.DefaultPoolID != "pool-rt-001" {
		t.Errorf("expected DefaultPoolID pool-rt-001, got %s", resp.DefaultPoolID)
	}
}

func TestListenerAPIResponse_JSONWithExtraFields(t *testing.T) {
	// API may return additional fields not in our struct — should not break parsing
	raw := `{
		"id": "lst-extra-001",
		"name": "extra-fields",
		"protocol": "HTTP",
		"protocol_port": 80,
		"loadbalancer_id": "lb-extra",
		"some_unknown_field": "should-be-ignored",
		"another_field": 42
	}`

	var resp listenerAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal with extra fields: %v", err)
	}

	if resp.ID != "lst-extra-001" {
		t.Errorf("expected ID lst-extra-001, got %s", resp.ID)
	}
	if resp.Name != "extra-fields" {
		t.Errorf("expected Name extra-fields, got %s", resp.Name)
	}
}

func TestListenerAPIResponse_JSONNullFields(t *testing.T) {
	raw := `{
		"id": "lst-null-001",
		"name": "null-fields",
		"protocol": null,
		"protocol_port": 0,
		"loadbalancer_id": null,
		"provisioning_status": null,
		"operating_status": null,
		"default_pool_id": null,
		"created_at": null
	}`

	var resp listenerAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal with null fields: %v", err)
	}

	if resp.Protocol != "" {
		t.Errorf("expected empty Protocol for null, got %s", resp.Protocol)
	}
	if resp.ProtocolPort != 0 {
		t.Errorf("expected 0 ProtocolPort, got %d", resp.ProtocolPort)
	}
	if resp.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID for null, got %s", resp.LoadBalancerID)
	}
}

// --- Schema tests ---

func TestLbListenerSchema(t *testing.T) {
	s := lbListenerSchema()

	if s.Description == "" {
		t.Error("expected non-empty schema description")
	}

	// Verify required attributes
	required := []string{"name", "protocol", "protocol_port", "loadbalancer_id"}
	for _, attr := range required {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
			continue
		}
		sa, ok := a.(schema.StringAttribute)
		if ok {
			if !sa.IsRequired() {
				t.Errorf("expected attribute %q to be required", attr)
			}
		}
	}

	// Verify computed attributes
	computed := []string{"id", "provisioning_status", "operating_status", "default_pool_id", "created_at"}
	for _, attr := range computed {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
			continue
		}
		sa, ok := a.(schema.StringAttribute)
		if ok {
			if !sa.IsComputed() {
				t.Errorf("expected attribute %q to be computed", attr)
			}
		}
	}
}

func TestMapAPIResponseToState_ProtocolPortZero(t *testing.T) {
	// Edge case: protocol_port of 0 should not update the model field
	model := &lbListenerResourceModel{
		ProtocolPort: types.Int64Value(8080),
	}

	apiResp := &listenerAPIResponse{
		ID:           "listener-port-zero",
		Name:         "port-zero",
		ProtocolPort: 0,
	}

	mapAPIResponseToState(model, apiResp)

	// ProtocolPort should stay at original value since API returned 0
	if model.ProtocolPort.ValueInt64() != 8080 {
		t.Errorf("expected ProtocolPort to remain 8080 when API returns 0, got %d", model.ProtocolPort.ValueInt64())
	}
}

func TestMapAPIResponseToState_ProtocolEmpty(t *testing.T) {
	// Edge case: empty protocol should not update the model field
	model := &lbListenerResourceModel{
		Protocol: types.StringValue("TCP"),
	}

	apiResp := &listenerAPIResponse{
		ID:   "listener-proto-empty",
		Name: "proto-empty",
	}

	mapAPIResponseToState(model, apiResp)

	// Protocol should stay at original value since API returned empty
	if model.Protocol.ValueString() != "TCP" {
		t.Errorf("expected Protocol to remain TCP when API returns empty, got %s", model.Protocol.ValueString())
	}
}
