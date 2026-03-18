package lb_pool

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &lbPoolResourceModel{
		Name:           types.StringValue("test-pool"),
		Protocol:       types.StringValue("HTTP"),
		LBAlgorithm:    types.StringValue("ROUND_ROBIN"),
		ListenerID:     types.StringValue("listener-123"),
		LoadBalancerID: types.StringValue("lb-456"),
	}

	body := buildCreateRequest(plan)

	if body.Name != "test-pool" {
		t.Errorf("expected name test-pool, got %s", body.Name)
	}
	if body.Protocol != "HTTP" {
		t.Errorf("expected protocol HTTP, got %s", body.Protocol)
	}
	if body.LBAlgorithm != "ROUND_ROBIN" {
		t.Errorf("expected lb_algorithm ROUND_ROBIN, got %s", body.LBAlgorithm)
	}
	if body.ListenerID != "listener-123" {
		t.Errorf("expected listener_id listener-123, got %s", body.ListenerID)
	}
	if body.LoadBalancerID != "lb-456" {
		t.Errorf("expected loadbalancer_id lb-456, got %s", body.LoadBalancerID)
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &lbPoolResourceModel{
		Name:           types.StringValue("minimal-pool"),
		Protocol:       types.StringValue("TCP"),
		LBAlgorithm:    types.StringValue("LEAST_CONNECTIONS"),
		ListenerID:     types.StringNull(),
		LoadBalancerID: types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.ListenerID != "" {
		t.Errorf("expected empty listener_id, got %s", body.ListenerID)
	}
	if body.LoadBalancerID != "" {
		t.Errorf("expected empty loadbalancer_id, got %s", body.LoadBalancerID)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &lbPoolResourceModel{
		Name:           types.StringValue("json-pool"),
		Protocol:       types.StringValue("HTTP"),
		LBAlgorithm:    types.StringValue("ROUND_ROBIN"),
		ListenerID:     types.StringValue("listener-json-001"),
		LoadBalancerID: types.StringValue("lb-json-002"),
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
	if _, ok := raw["lb_algorithm"]; !ok {
		t.Error("expected JSON key 'lb_algorithm' to be present")
	}
	if _, ok := raw["listener_id"]; !ok {
		t.Error("expected JSON key 'listener_id' to be present")
	}
	if _, ok := raw["loadbalancer_id"]; !ok {
		t.Error("expected JSON key 'loadbalancer_id' to be present")
	}

	if raw["name"] != "json-pool" {
		t.Errorf("expected JSON name 'json-pool', got %v", raw["name"])
	}
	if raw["protocol"] != "HTTP" {
		t.Errorf("expected JSON protocol 'HTTP', got %v", raw["protocol"])
	}
	if raw["lb_algorithm"] != "ROUND_ROBIN" {
		t.Errorf("expected JSON lb_algorithm 'ROUND_ROBIN', got %v", raw["lb_algorithm"])
	}
	if raw["listener_id"] != "listener-json-001" {
		t.Errorf("expected JSON listener_id 'listener-json-001', got %v", raw["listener_id"])
	}
	if raw["loadbalancer_id"] != "lb-json-002" {
		t.Errorf("expected JSON loadbalancer_id 'lb-json-002', got %v", raw["loadbalancer_id"])
	}
}

func TestBuildCreateRequest_AllAlgorithms(t *testing.T) {
	algorithms := []string{
		"ROUND_ROBIN",
		"LEAST_CONNECTIONS",
		"SOURCE_IP",
	}

	for _, algo := range algorithms {
		t.Run(algo, func(t *testing.T) {
			plan := &lbPoolResourceModel{
				Name:           types.StringValue("pool-" + algo),
				Protocol:       types.StringValue("HTTP"),
				LBAlgorithm:    types.StringValue(algo),
				ListenerID:     types.StringValue("listener-algo-test"),
				LoadBalancerID: types.StringValue("lb-algo-test"),
			}

			body := buildCreateRequest(plan)

			if body.LBAlgorithm != algo {
				t.Errorf("expected lb_algorithm %s, got %s", algo, body.LBAlgorithm)
			}
			if body.Name != "pool-"+algo {
				t.Errorf("expected name 'pool-%s', got %s", algo, body.Name)
			}
		})
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &lbPoolResourceModel{
		ListenerID:     types.StringNull(),
		LoadBalancerID: types.StringNull(),
	}

	apiResp := &poolAPIResponse{
		ID:                 "pool-abc-123",
		Name:               "prod-pool",
		Protocol:           "HTTP",
		LBAlgorithm:        "ROUND_ROBIN",
		ListenerID:         "listener-xyz",
		LoadBalancerID:     "lb-xyz",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		HealthMonitorID:    "hm-123",
		CreatedAt:          "2024-01-01T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "pool-abc-123" {
		t.Errorf("expected ID pool-abc-123, got %s", model.ID.ValueString())
	}
	if model.LBAlgorithm.ValueString() != "ROUND_ROBIN" {
		t.Errorf("expected LBAlgorithm ROUND_ROBIN, got %s", model.LBAlgorithm.ValueString())
	}
	if model.HealthMonitorID.ValueString() != "hm-123" {
		t.Errorf("expected HealthMonitorID hm-123, got %s", model.HealthMonitorID.ValueString())
	}
}

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	model := &lbPoolResourceModel{
		ListenerID:     types.StringNull(),
		LoadBalancerID: types.StringNull(),
	}

	apiResp := &poolAPIResponse{
		ID:                 "pool-full-001",
		Name:               "full-pool",
		Protocol:           "HTTP",
		LBAlgorithm:        "LEAST_CONNECTIONS",
		ListenerID:         "listener-full-xyz",
		LoadBalancerID:     "lb-full-xyz",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		HealthMonitorID:    "hm-full-001",
		CreatedAt:          "2024-06-15T10:30:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "pool-full-001" {
		t.Errorf("expected ID pool-full-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-pool" {
		t.Errorf("expected Name full-pool, got %s", model.Name.ValueString())
	}
	if model.Protocol.ValueString() != "HTTP" {
		t.Errorf("expected Protocol HTTP, got %s", model.Protocol.ValueString())
	}
	if model.LBAlgorithm.ValueString() != "LEAST_CONNECTIONS" {
		t.Errorf("expected LBAlgorithm LEAST_CONNECTIONS, got %s", model.LBAlgorithm.ValueString())
	}
	if model.ListenerID.ValueString() != "listener-full-xyz" {
		t.Errorf("expected ListenerID listener-full-xyz, got %s", model.ListenerID.ValueString())
	}
	if model.LoadBalancerID.ValueString() != "lb-full-xyz" {
		t.Errorf("expected LoadBalancerID lb-full-xyz, got %s", model.LoadBalancerID.ValueString())
	}
	if model.ProvisioningStatus.ValueString() != "ACTIVE" {
		t.Errorf("expected ProvisioningStatus ACTIVE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "ONLINE" {
		t.Errorf("expected OperatingStatus ONLINE, got %s", model.OperatingStatus.ValueString())
	}
	if model.HealthMonitorID.ValueString() != "hm-full-001" {
		t.Errorf("expected HealthMonitorID hm-full-001, got %s", model.HealthMonitorID.ValueString())
	}
	if model.CreatedAt.ValueString() != "2024-06-15T10:30:00Z" {
		t.Errorf("expected CreatedAt '2024-06-15T10:30:00Z', got %s", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &lbPoolResourceModel{
		ListenerID:     types.StringNull(),
		LoadBalancerID: types.StringNull(),
	}

	apiResp := &poolAPIResponse{
		ID:          "pool-123",
		Name:        "basic",
		ListenerID:  "",
		Protocol:    "TCP",
		LBAlgorithm: "SOURCE_IP",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.ListenerID.IsNull() {
		t.Error("expected ListenerID to remain null when API returns empty string")
	}
}

func TestPoolDeleteRequest(t *testing.T) {
	req := poolDeleteRequest{
		Key:    "id",
		Values: []string{"pool-del-001"},
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
	if values[0] != "pool-del-001" {
		t.Errorf("expected value 'pool-del-001', got %v", values[0])
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- parsePoolFromRaw tests ---

func TestParsePoolFromRaw_DirectFields(t *testing.T) {
	raw := `{
		"id": "pool-001",
		"name": "my-pool",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"listener_id": "listener-direct-001",
		"loadbalancer_id": "lb-direct-001",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"healthmonitor_id": "hm-001",
		"created_at": "2024-06-01T12:00:00Z"
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ID != "pool-001" {
		t.Errorf("expected ID pool-001, got %s", pool.ID)
	}
	if pool.Name != "my-pool" {
		t.Errorf("expected Name my-pool, got %s", pool.Name)
	}
	if pool.Protocol != "HTTP" {
		t.Errorf("expected Protocol HTTP, got %s", pool.Protocol)
	}
	if pool.LBAlgorithm != "ROUND_ROBIN" {
		t.Errorf("expected LBAlgorithm ROUND_ROBIN, got %s", pool.LBAlgorithm)
	}
	if pool.ListenerID != "listener-direct-001" {
		t.Errorf("expected ListenerID listener-direct-001, got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "lb-direct-001" {
		t.Errorf("expected LoadBalancerID lb-direct-001, got %s", pool.LoadBalancerID)
	}
	if pool.HealthMonitorID != "hm-001" {
		t.Errorf("expected HealthMonitorID hm-001, got %s", pool.HealthMonitorID)
	}
}

func TestParsePoolFromRaw_ListenersArray(t *testing.T) {
	raw := `{
		"id": "pool-002",
		"name": "array-pool",
		"protocol": "TCP",
		"lb_algorithm": "LEAST_CONNECTIONS",
		"listeners": [{"id": "listener-from-array"}],
		"loadbalancers": [{"id": "lb-from-array"}],
		"provisioning_status": "ACTIVE"
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ListenerID != "listener-from-array" {
		t.Errorf("expected ListenerID listener-from-array from listeners array, got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "lb-from-array" {
		t.Errorf("expected LoadBalancerID lb-from-array from loadbalancers array, got %s", pool.LoadBalancerID)
	}
}

func TestParsePoolFromRaw_EmptyArrays(t *testing.T) {
	raw := `{
		"id": "pool-003",
		"name": "empty-arrays",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"listeners": [],
		"loadbalancers": []
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ListenerID != "" {
		t.Errorf("expected empty ListenerID when listeners array is empty, got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID when loadbalancers array is empty, got %s", pool.LoadBalancerID)
	}
}

func TestParsePoolFromRaw_MultipleInArrays(t *testing.T) {
	// Should pick the first one
	raw := `{
		"id": "pool-004",
		"name": "multi-array",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"listeners": [{"id": "listener-first"}, {"id": "listener-second"}],
		"loadbalancers": [{"id": "lb-first"}, {"id": "lb-second"}]
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ListenerID != "listener-first" {
		t.Errorf("expected ListenerID listener-first (first), got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "lb-first" {
		t.Errorf("expected LoadBalancerID lb-first (first), got %s", pool.LoadBalancerID)
	}
}

func TestParsePoolFromRaw_DirectIDTakesPrecedence(t *testing.T) {
	raw := `{
		"id": "pool-005",
		"name": "precedence-pool",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"listener_id": "listener-direct",
		"loadbalancer_id": "lb-direct",
		"listeners": [{"id": "listener-array"}],
		"loadbalancers": [{"id": "lb-array"}]
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ListenerID != "listener-direct" {
		t.Errorf("expected ListenerID listener-direct (direct), got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "lb-direct" {
		t.Errorf("expected LoadBalancerID lb-direct (direct), got %s", pool.LoadBalancerID)
	}
}

func TestParsePoolFromRaw_MixedDirectAndArray(t *testing.T) {
	// listener_id is direct, loadbalancer_id comes from array
	raw := `{
		"id": "pool-006",
		"name": "mixed-pool",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"listener_id": "listener-direct-only",
		"loadbalancers": [{"id": "lb-from-array-only"}]
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ListenerID != "listener-direct-only" {
		t.Errorf("expected ListenerID listener-direct-only, got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "lb-from-array-only" {
		t.Errorf("expected LoadBalancerID lb-from-array-only, got %s", pool.LoadBalancerID)
	}
}

func TestParsePoolFromRaw_InvalidJSON(t *testing.T) {
	_, err := parsePoolFromRaw(json.RawMessage(`{invalid}`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParsePoolFromRaw_EmptyObject(t *testing.T) {
	pool, err := parsePoolFromRaw(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ID != "" {
		t.Errorf("expected empty ID, got %s", pool.ID)
	}
	if pool.ListenerID != "" {
		t.Errorf("expected empty ListenerID, got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID, got %s", pool.LoadBalancerID)
	}
}

func TestParsePoolFromRaw_MalformedArrayEntries(t *testing.T) {
	raw := `{
		"id": "pool-007",
		"name": "malformed",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"listeners": ["not-an-object"],
		"loadbalancers": [42]
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ListenerID != "" {
		t.Errorf("expected empty ListenerID for malformed entries, got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID for malformed entries, got %s", pool.LoadBalancerID)
	}
}

func TestParsePoolFromRaw_ArrayEntriesMissingID(t *testing.T) {
	raw := `{
		"id": "pool-008",
		"name": "no-id",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"listeners": [{"name": "listener-no-id"}],
		"loadbalancers": [{"name": "lb-no-id"}]
	}`

	pool, err := parsePoolFromRaw(json.RawMessage(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pool.ListenerID != "" {
		t.Errorf("expected empty ListenerID when object has no id, got %s", pool.ListenerID)
	}
	if pool.LoadBalancerID != "" {
		t.Errorf("expected empty LoadBalancerID when object has no id, got %s", pool.LoadBalancerID)
	}
}

// --- Name regex tests ---

func TestPoolNameRegex(t *testing.T) {
	valid := []string{
		"my-pool",
		"pool1",
		"a",
		"A-B-C",
		"test-123-prod",
		"UPPERCASE",
		"MiXeD-CaSe-123",
	}

	for _, name := range valid {
		t.Run("valid_"+name, func(t *testing.T) {
			if !poolNameRegex.MatchString(name) {
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
		"",
	}

	for _, name := range invalid {
		label := name
		if label == "" {
			label = "empty"
		}
		t.Run("invalid_"+label, func(t *testing.T) {
			if poolNameRegex.MatchString(name) {
				t.Errorf("expected %q to be invalid", name)
			}
		})
	}
}

// --- mapAPIResponseToState edge cases ---

func TestMapAPIResponseToState_ListenerIDNotNullPreserved(t *testing.T) {
	// When model has a non-null ListenerID and API returns empty, it should preserve the model value
	model := &lbPoolResourceModel{
		ListenerID:     types.StringValue("existing-listener"),
		LoadBalancerID: types.StringValue("existing-lb"),
	}

	apiResp := &poolAPIResponse{
		ID:          "pool-preserve",
		Name:        "preserve",
		Protocol:    "HTTP",
		LBAlgorithm: "ROUND_ROBIN",
		ListenerID:  "",
		// API returns empty listener_id and loadbalancer_id
	}

	mapAPIResponseToState(model, apiResp)

	// Since model.ListenerID was NOT null, and API returned empty, the else-if branch checks IsNull()
	// Model was StringValue (not null), so the condition model.ListenerID.IsNull() is false
	// This means the field stays at whatever it was (the model field is not reassigned)
	if model.ListenerID.ValueString() != "existing-listener" {
		t.Errorf("expected ListenerID to remain existing-listener, got %s", model.ListenerID.ValueString())
	}
}

func TestMapAPIResponseToState_AllFieldsNull(t *testing.T) {
	model := &lbPoolResourceModel{
		ListenerID:     types.StringNull(),
		LoadBalancerID: types.StringNull(),
	}

	apiResp := &poolAPIResponse{
		ID:   "pool-all-empty",
		Name: "all-empty",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.ListenerID.IsNull() {
		t.Error("expected ListenerID to remain null")
	}
	if !model.LoadBalancerID.IsNull() {
		t.Error("expected LoadBalancerID to remain null")
	}
	if !model.ProvisioningStatus.IsNull() {
		t.Error("expected ProvisioningStatus to be null")
	}
	if !model.OperatingStatus.IsNull() {
		t.Error("expected OperatingStatus to be null")
	}
	if !model.HealthMonitorID.IsNull() {
		t.Error("expected HealthMonitorID to be null")
	}
	if !model.CreatedAt.IsNull() {
		t.Error("expected CreatedAt to be null")
	}
}

func TestMapAPIResponseToState_PendingStatuses(t *testing.T) {
	model := &lbPoolResourceModel{
		ListenerID:     types.StringNull(),
		LoadBalancerID: types.StringNull(),
	}

	apiResp := &poolAPIResponse{
		ID:                 "pool-pending",
		Name:               "pending-pool",
		Protocol:           "HTTP",
		LBAlgorithm:        "ROUND_ROBIN",
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

// --- JSON round-trip for API response ---

func TestPoolAPIResponse_JSONDeserialization(t *testing.T) {
	raw := `{
		"id": "pool-rt-001",
		"name": "roundtrip-pool",
		"protocol": "PROXY",
		"lb_algorithm": "SOURCE_IP",
		"listener_id": "listener-rt",
		"loadbalancer_id": "lb-rt",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"healthmonitor_id": "hm-rt-001",
		"created_at": "2025-01-15T09:30:00Z"
	}`

	var resp poolAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "pool-rt-001" {
		t.Errorf("expected ID pool-rt-001, got %s", resp.ID)
	}
	if resp.Protocol != "PROXY" {
		t.Errorf("expected Protocol PROXY, got %s", resp.Protocol)
	}
	if resp.LBAlgorithm != "SOURCE_IP" {
		t.Errorf("expected LBAlgorithm SOURCE_IP, got %s", resp.LBAlgorithm)
	}
	if resp.HealthMonitorID != "hm-rt-001" {
		t.Errorf("expected HealthMonitorID hm-rt-001, got %s", resp.HealthMonitorID)
	}
}

func TestPoolAPIResponse_JSONWithExtraFields(t *testing.T) {
	raw := `{
		"id": "pool-extra-001",
		"name": "extra-fields",
		"protocol": "HTTP",
		"lb_algorithm": "ROUND_ROBIN",
		"unknown_field": "ignored",
		"another_extra": 999
	}`

	var resp poolAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal with extra fields: %v", err)
	}

	if resp.ID != "pool-extra-001" {
		t.Errorf("expected ID pool-extra-001, got %s", resp.ID)
	}
}

func TestPoolAPIResponse_JSONNullFields(t *testing.T) {
	raw := `{
		"id": "pool-null-001",
		"name": "null-fields",
		"protocol": null,
		"lb_algorithm": null,
		"listener_id": null,
		"loadbalancer_id": null,
		"healthmonitor_id": null
	}`

	var resp poolAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal with null fields: %v", err)
	}

	if resp.Protocol != "" {
		t.Errorf("expected empty Protocol for null, got %s", resp.Protocol)
	}
	if resp.LBAlgorithm != "" {
		t.Errorf("expected empty LBAlgorithm for null, got %s", resp.LBAlgorithm)
	}
}

// --- buildCreateRequest JSON omitempty ---

func TestBuildCreateRequest_JSONOmitEmpty(t *testing.T) {
	plan := &lbPoolResourceModel{
		Name:           types.StringValue("omit-pool"),
		Protocol:       types.StringValue("TCP"),
		LBAlgorithm:    types.StringValue("ROUND_ROBIN"),
		ListenerID:     types.StringNull(),
		LoadBalancerID: types.StringNull(),
	}

	body := buildCreateRequest(plan)
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// listener_id and loadbalancer_id have omitempty, so empty strings should be omitted
	if _, exists := parsed["listener_id"]; exists {
		t.Error("expected listener_id to be omitted when empty (omitempty)")
	}
	if _, exists := parsed["loadbalancer_id"]; exists {
		t.Error("expected loadbalancer_id to be omitted when empty (omitempty)")
	}
}

// --- Schema tests ---

func TestLbPoolSchema(t *testing.T) {
	s := lbPoolSchema()

	if s.Description == "" {
		t.Error("expected non-empty schema description")
	}

	// Verify required attributes
	requiredStr := []string{"name", "protocol", "lb_algorithm"}
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

	// Verify optional attributes
	optional := []string{"listener_id", "loadbalancer_id"}
	for _, attr := range optional {
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

	// Verify computed attributes
	computed := []string{"id", "provisioning_status", "operating_status", "healthmonitor_id", "created_at"}
	for _, attr := range computed {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
			continue
		}
		sa, ok := a.(schema.StringAttribute)
		if ok && !sa.IsComputed() {
			t.Errorf("expected attribute %q to be computed", attr)
		}
	}
}

func TestBuildCreateRequest_AllProtocols(t *testing.T) {
	protocols := []string{"HTTP", "HTTPS", "TCP", "UDP", "PROXY", "PROXYV2"}

	for _, proto := range protocols {
		t.Run(proto, func(t *testing.T) {
			plan := &lbPoolResourceModel{
				Name:           types.StringValue("pool-" + proto),
				Protocol:       types.StringValue(proto),
				LBAlgorithm:    types.StringValue("ROUND_ROBIN"),
				ListenerID:     types.StringNull(),
				LoadBalancerID: types.StringNull(),
			}

			body := buildCreateRequest(plan)

			if body.Protocol != proto {
				t.Errorf("expected protocol %s, got %s", proto, body.Protocol)
			}
		})
	}
}
