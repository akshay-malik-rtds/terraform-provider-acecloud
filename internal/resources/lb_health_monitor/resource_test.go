package lb_health_monitor

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("test-hm"),
		PoolID:         types.StringValue("pool-123"),
		Type:           types.StringValue("HTTP"),
		Delay:          types.Int64Value(5),
		Timeout:        types.Int64Value(3),
		MaxRetries:     types.Int64Value(3),
		MaxRetriesDown: types.Int64Value(2),
		URLPath:        types.StringValue("/health"),
		ExpectedCodes:  types.StringValue("200"),
		HTTPMethod:     types.StringValue("GET"),
	}

	body := buildCreateRequest(plan)

	if body.Name != "test-hm" {
		t.Errorf("expected name test-hm, got %s", body.Name)
	}
	if body.PoolID != "pool-123" {
		t.Errorf("expected pool_id pool-123, got %s", body.PoolID)
	}
	if body.Type != "HTTP" {
		t.Errorf("expected type HTTP, got %s", body.Type)
	}
	if body.Delay != 5 {
		t.Errorf("expected delay 5, got %d", body.Delay)
	}
	if body.Timeout != 3 {
		t.Errorf("expected timeout 3, got %d", body.Timeout)
	}
	if body.MaxRetries != 3 {
		t.Errorf("expected max_retries 3, got %d", body.MaxRetries)
	}
	if body.MaxRetriesDown != 2 {
		t.Errorf("expected max_retries_down 2, got %d", body.MaxRetriesDown)
	}
	if body.URLPath != "/health" {
		t.Errorf("expected url_path /health, got %s", body.URLPath)
	}
	if body.ExpectedCodes != "200" {
		t.Errorf("expected expected_codes 200, got %s", body.ExpectedCodes)
	}
	if body.HTTPMethod != "GET" {
		t.Errorf("expected http_method GET, got %s", body.HTTPMethod)
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("tcp-hm"),
		PoolID:         types.StringValue("pool-456"),
		Type:           types.StringValue("TCP"),
		Delay:          types.Int64Value(10),
		Timeout:        types.Int64Value(5),
		MaxRetries:     types.Int64Value(3),
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.MaxRetriesDown != 0 {
		t.Errorf("expected max_retries_down 0, got %d", body.MaxRetriesDown)
	}
	if body.URLPath != "" {
		t.Errorf("expected empty url_path, got %s", body.URLPath)
	}
	if body.ExpectedCodes != "" {
		t.Errorf("expected empty expected_codes, got %s", body.ExpectedCodes)
	}
	if body.HTTPMethod != "" {
		t.Errorf("expected empty http_method, got %s", body.HTTPMethod)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("json-hm"),
		PoolID:         types.StringValue("pool-json-1"),
		Type:           types.StringValue("HTTP"),
		Delay:          types.Int64Value(10),
		Timeout:        types.Int64Value(5),
		MaxRetries:     types.Int64Value(3),
		MaxRetriesDown: types.Int64Value(2),
		URLPath:        types.StringValue("/status"),
		ExpectedCodes:  types.StringValue("200-299"),
		HTTPMethod:     types.StringValue("HEAD"),
	}

	body := buildCreateRequest(plan)

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify JSON field names match API expectations
	if parsed["name"] != "json-hm" {
		t.Errorf("expected name 'json-hm', got %v", parsed["name"])
	}
	if parsed["pool_id"] != "pool-json-1" {
		t.Errorf("expected pool_id 'pool-json-1', got %v", parsed["pool_id"])
	}
	if parsed["type"] != "HTTP" {
		t.Errorf("expected type 'HTTP', got %v", parsed["type"])
	}
	if parsed["delay"] != float64(10) {
		t.Errorf("expected delay 10, got %v", parsed["delay"])
	}
	if parsed["timeout"] != float64(5) {
		t.Errorf("expected timeout 5, got %v", parsed["timeout"])
	}
	if parsed["max_retries"] != float64(3) {
		t.Errorf("expected max_retries 3, got %v", parsed["max_retries"])
	}
	if parsed["max_retries_down"] != float64(2) {
		t.Errorf("expected max_retries_down 2, got %v", parsed["max_retries_down"])
	}
	if parsed["url_path"] != "/status" {
		t.Errorf("expected url_path '/status', got %v", parsed["url_path"])
	}
	if parsed["expected_codes"] != "200-299" {
		t.Errorf("expected expected_codes '200-299', got %v", parsed["expected_codes"])
	}
	if parsed["http_method"] != "HEAD" {
		t.Errorf("expected http_method 'HEAD', got %v", parsed["http_method"])
	}
}

func TestBuildCreateRequest_TCPMonitor(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("tcp-monitor"),
		PoolID:         types.StringValue("pool-tcp-1"),
		Type:           types.StringValue("TCP"),
		Delay:          types.Int64Value(15),
		Timeout:        types.Int64Value(10),
		MaxRetries:     types.Int64Value(5),
		MaxRetriesDown: types.Int64Value(3),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.Name != "tcp-monitor" {
		t.Errorf("expected name tcp-monitor, got %s", body.Name)
	}
	if body.Type != "TCP" {
		t.Errorf("expected type TCP, got %s", body.Type)
	}
	if body.Delay != 15 {
		t.Errorf("expected delay 15, got %d", body.Delay)
	}
	// TCP monitors should not have HTTP-specific fields
	if body.URLPath != "" {
		t.Errorf("expected empty url_path for TCP monitor, got %s", body.URLPath)
	}
	if body.ExpectedCodes != "" {
		t.Errorf("expected empty expected_codes for TCP monitor, got %s", body.ExpectedCodes)
	}
	if body.HTTPMethod != "" {
		t.Errorf("expected empty http_method for TCP monitor, got %s", body.HTTPMethod)
	}

	// Verify JSON omits HTTP fields via omitempty
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, exists := parsed["url_path"]; exists {
		t.Error("TCP monitor JSON should not contain url_path")
	}
	if _, exists := parsed["expected_codes"]; exists {
		t.Error("TCP monitor JSON should not contain expected_codes")
	}
	if _, exists := parsed["http_method"]; exists {
		t.Error("TCP monitor JSON should not contain http_method")
	}
}

func TestBuildCreateRequest_PINGMonitor(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("ping-monitor"),
		PoolID:         types.StringValue("pool-ping-1"),
		Type:           types.StringValue("PING"),
		Delay:          types.Int64Value(5),
		Timeout:        types.Int64Value(3),
		MaxRetries:     types.Int64Value(3),
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.Name != "ping-monitor" {
		t.Errorf("expected name ping-monitor, got %s", body.Name)
	}
	if body.Type != "PING" {
		t.Errorf("expected type PING, got %s", body.Type)
	}
	if body.MaxRetriesDown != 0 {
		t.Errorf("expected max_retries_down 0 for PING, got %d", body.MaxRetriesDown)
	}
	if body.URLPath != "" {
		t.Errorf("expected empty url_path for PING monitor, got %s", body.URLPath)
	}
	if body.ExpectedCodes != "" {
		t.Errorf("expected empty expected_codes for PING monitor, got %s", body.ExpectedCodes)
	}
	if body.HTTPMethod != "" {
		t.Errorf("expected empty http_method for PING monitor, got %s", body.HTTPMethod)
	}

	// Verify JSON omits optional fields
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, exists := parsed["max_retries_down"]; exists {
		t.Error("PING monitor JSON should not contain max_retries_down when null")
	}
	if _, exists := parsed["url_path"]; exists {
		t.Error("PING monitor JSON should not contain url_path")
	}
}

func TestBuildUpdateRequest(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("updated-hm"),
		Delay:          types.Int64Value(15),
		Timeout:        types.Int64Value(10),
		MaxRetries:     types.Int64Value(5),
		MaxRetriesDown: types.Int64Value(3),
		URLPath:        types.StringValue("/status"),
		ExpectedCodes:  types.StringValue("200-299"),
		HTTPMethod:     types.StringValue("HEAD"),
	}

	body := buildUpdateRequest(plan)

	if body.Name != "updated-hm" {
		t.Errorf("expected name updated-hm, got %s", body.Name)
	}
	if body.Delay != 15 {
		t.Errorf("expected delay 15, got %d", body.Delay)
	}
	if body.URLPath != "/status" {
		t.Errorf("expected url_path /status, got %s", body.URLPath)
	}
}

func TestBuildUpdateRequest_MinimalFields(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("minimal-update"),
		Delay:          types.Int64Value(10),
		Timeout:        types.Int64Value(5),
		MaxRetries:     types.Int64Value(3),
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	body := buildUpdateRequest(plan)

	if body.Name != "minimal-update" {
		t.Errorf("expected name minimal-update, got %s", body.Name)
	}
	if body.Delay != 10 {
		t.Errorf("expected delay 10, got %d", body.Delay)
	}
	if body.Timeout != 5 {
		t.Errorf("expected timeout 5, got %d", body.Timeout)
	}
	if body.MaxRetries != 3 {
		t.Errorf("expected max_retries 3, got %d", body.MaxRetries)
	}
	if body.MaxRetriesDown != 0 {
		t.Errorf("expected max_retries_down 0 (null), got %d", body.MaxRetriesDown)
	}
	if body.URLPath != "" {
		t.Errorf("expected empty url_path, got %s", body.URLPath)
	}
	if body.ExpectedCodes != "" {
		t.Errorf("expected empty expected_codes, got %s", body.ExpectedCodes)
	}
	if body.HTTPMethod != "" {
		t.Errorf("expected empty http_method, got %s", body.HTTPMethod)
	}
}

func TestBuildUpdateRequest_JSON(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("json-update"),
		Delay:          types.Int64Value(20),
		Timeout:        types.Int64Value(15),
		MaxRetries:     types.Int64Value(4),
		MaxRetriesDown: types.Int64Value(2),
		URLPath:        types.StringValue("/ready"),
		ExpectedCodes:  types.StringValue("200"),
		HTTPMethod:     types.StringValue("GET"),
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

	if parsed["name"] != "json-update" {
		t.Errorf("expected name 'json-update', got %v", parsed["name"])
	}
	if parsed["delay"] != float64(20) {
		t.Errorf("expected delay 20, got %v", parsed["delay"])
	}
	if parsed["timeout"] != float64(15) {
		t.Errorf("expected timeout 15, got %v", parsed["timeout"])
	}
	if parsed["max_retries"] != float64(4) {
		t.Errorf("expected max_retries 4, got %v", parsed["max_retries"])
	}
	if parsed["max_retries_down"] != float64(2) {
		t.Errorf("expected max_retries_down 2, got %v", parsed["max_retries_down"])
	}
	if parsed["url_path"] != "/ready" {
		t.Errorf("expected url_path '/ready', got %v", parsed["url_path"])
	}
	if parsed["expected_codes"] != "200" {
		t.Errorf("expected expected_codes '200', got %v", parsed["expected_codes"])
	}
	if parsed["http_method"] != "GET" {
		t.Errorf("expected http_method 'GET', got %v", parsed["http_method"])
	}

	// Update request should NOT contain pool_id or type
	if _, exists := parsed["pool_id"]; exists {
		t.Error("update request should not contain pool_id")
	}
	if _, exists := parsed["type"]; exists {
		t.Error("update request should not contain type")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:                 "hm-abc-123",
		Name:               "prod-hm",
		PoolID:             "pool-xyz",
		Type:               "HTTP",
		Delay:              5,
		Timeout:            3,
		MaxRetries:         3,
		MaxRetriesDown:     2,
		URLPath:            "/health",
		ExpectedCodes:      "200",
		HTTPMethod:         "GET",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		AdminStateUp:       true,
		CreatedAt:          "2024-01-01T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "hm-abc-123" {
		t.Errorf("expected ID hm-abc-123, got %s", model.ID.ValueString())
	}
	if model.Type.ValueString() != "HTTP" {
		t.Errorf("expected Type HTTP, got %s", model.Type.ValueString())
	}
	if model.Delay.ValueInt64() != 5 {
		t.Errorf("expected Delay 5, got %d", model.Delay.ValueInt64())
	}
	if model.URLPath.ValueString() != "/health" {
		t.Errorf("expected URLPath /health, got %s", model.URLPath.ValueString())
	}
	if model.AdminStateUp.ValueBool() != true {
		t.Error("expected AdminStateUp to be true")
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:         "hm-123",
		Name:       "tcp-hm",
		Type:       "TCP",
		Delay:      10,
		Timeout:    5,
		MaxRetries: 3,
	}

	mapAPIResponseToState(model, apiResp)

	if !model.MaxRetriesDown.IsNull() {
		t.Error("expected MaxRetriesDown to remain null")
	}
	if !model.URLPath.IsNull() {
		t.Error("expected URLPath to remain null")
	}
	if !model.ExpectedCodes.IsNull() {
		t.Error("expected ExpectedCodes to remain null")
	}
	if !model.HTTPMethod.IsNull() {
		t.Error("expected HTTPMethod to remain null")
	}
}

func TestMapAPIResponseToState_TCPMonitor(t *testing.T) {
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:                 "hm-tcp-789",
		Name:               "tcp-health",
		PoolID:             "pool-tcp-xyz",
		Type:               "TCP",
		Delay:              10,
		Timeout:            5,
		MaxRetries:         3,
		MaxRetriesDown:     2,
		URLPath:            "",
		ExpectedCodes:      "",
		HTTPMethod:         "",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		AdminStateUp:       true,
		CreatedAt:          "2024-03-15T08:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "hm-tcp-789" {
		t.Errorf("expected ID hm-tcp-789, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "tcp-health" {
		t.Errorf("expected Name tcp-health, got %s", model.Name.ValueString())
	}
	if model.PoolID.ValueString() != "pool-tcp-xyz" {
		t.Errorf("expected PoolID pool-tcp-xyz, got %s", model.PoolID.ValueString())
	}
	if model.Type.ValueString() != "TCP" {
		t.Errorf("expected Type TCP, got %s", model.Type.ValueString())
	}
	if model.Delay.ValueInt64() != 10 {
		t.Errorf("expected Delay 10, got %d", model.Delay.ValueInt64())
	}
	if model.Timeout.ValueInt64() != 5 {
		t.Errorf("expected Timeout 5, got %d", model.Timeout.ValueInt64())
	}
	if model.MaxRetries.ValueInt64() != 3 {
		t.Errorf("expected MaxRetries 3, got %d", model.MaxRetries.ValueInt64())
	}
	if model.MaxRetriesDown.ValueInt64() != 2 {
		t.Errorf("expected MaxRetriesDown 2, got %d", model.MaxRetriesDown.ValueInt64())
	}
	// TCP monitor: url_path, expected_codes, http_method should remain null
	if !model.URLPath.IsNull() {
		t.Error("expected URLPath to remain null for TCP monitor")
	}
	if !model.ExpectedCodes.IsNull() {
		t.Error("expected ExpectedCodes to remain null for TCP monitor")
	}
	if !model.HTTPMethod.IsNull() {
		t.Error("expected HTTPMethod to remain null for TCP monitor")
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
	if model.CreatedAt.ValueString() != "2024-03-15T08:00:00Z" {
		t.Errorf("expected CreatedAt 2024-03-15T08:00:00Z, got %s", model.CreatedAt.ValueString())
	}
}

func TestHMDeleteRequest(t *testing.T) {
	req := hmDeleteRequest{
		Key:    "id",
		Values: []string{"hm-del-1"},
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
	if values[0] != "hm-del-1" {
		t.Errorf("expected value 'hm-del-1', got %v", values[0])
	}

	// Verify exact JSON format
	expectedJSON := `{"key":"id","values":["hm-del-1"]}`
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

// --- Name regex tests ---

func TestHMNameRegex(t *testing.T) {
	valid := []string{
		"my-monitor",
		"monitor1",
		"a",
		"A-B-C",
		"test-123-prod",
		"UPPERCASE",
		"MiXeD-CaSe-123",
	}

	for _, name := range valid {
		t.Run("valid_"+name, func(t *testing.T) {
			if !hmNameRegex.MatchString(name) {
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
			if hmNameRegex.MatchString(name) {
				t.Errorf("expected %q to be invalid", name)
			}
		})
	}
}

// --- mapAPIResponseToState edge cases ---

func TestMapAPIResponseToState_AdminStateUpFalse(t *testing.T) {
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:           "hm-admin-false",
		Name:         "admin-false",
		PoolID:       "pool-001",
		Type:         "TCP",
		Delay:        5,
		Timeout:      3,
		MaxRetries:   3,
		AdminStateUp: false,
	}

	mapAPIResponseToState(model, apiResp)

	if model.AdminStateUp.ValueBool() != false {
		t.Error("expected AdminStateUp to be false")
	}
}

func TestMapAPIResponseToState_AllFieldsFull(t *testing.T) {
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:                 "hm-full-001",
		Name:               "full-hm",
		PoolID:             "pool-full-001",
		Type:               "HTTPS",
		Delay:              10,
		Timeout:            5,
		MaxRetries:         5,
		MaxRetriesDown:     3,
		URLPath:            "/healthz",
		ExpectedCodes:      "200-299",
		HTTPMethod:         "POST",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		AdminStateUp:       true,
		CreatedAt:          "2025-06-01T12:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "hm-full-001" {
		t.Errorf("expected ID hm-full-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-hm" {
		t.Errorf("expected Name full-hm, got %s", model.Name.ValueString())
	}
	if model.PoolID.ValueString() != "pool-full-001" {
		t.Errorf("expected PoolID pool-full-001, got %s", model.PoolID.ValueString())
	}
	if model.Type.ValueString() != "HTTPS" {
		t.Errorf("expected Type HTTPS, got %s", model.Type.ValueString())
	}
	if model.Delay.ValueInt64() != 10 {
		t.Errorf("expected Delay 10, got %d", model.Delay.ValueInt64())
	}
	if model.Timeout.ValueInt64() != 5 {
		t.Errorf("expected Timeout 5, got %d", model.Timeout.ValueInt64())
	}
	if model.MaxRetries.ValueInt64() != 5 {
		t.Errorf("expected MaxRetries 5, got %d", model.MaxRetries.ValueInt64())
	}
	if model.MaxRetriesDown.ValueInt64() != 3 {
		t.Errorf("expected MaxRetriesDown 3, got %d", model.MaxRetriesDown.ValueInt64())
	}
	if model.URLPath.ValueString() != "/healthz" {
		t.Errorf("expected URLPath /healthz, got %s", model.URLPath.ValueString())
	}
	if model.ExpectedCodes.ValueString() != "200-299" {
		t.Errorf("expected ExpectedCodes 200-299, got %s", model.ExpectedCodes.ValueString())
	}
	if model.HTTPMethod.ValueString() != "POST" {
		t.Errorf("expected HTTPMethod POST, got %s", model.HTTPMethod.ValueString())
	}
	if model.ProvisioningStatus.ValueString() != "ACTIVE" {
		t.Errorf("expected ProvisioningStatus ACTIVE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "ONLINE" {
		t.Errorf("expected OperatingStatus ONLINE, got %s", model.OperatingStatus.ValueString())
	}
	if model.AdminStateUp.ValueBool() != true {
		t.Error("expected AdminStateUp true")
	}
	if model.CreatedAt.ValueString() != "2025-06-01T12:00:00Z" {
		t.Errorf("expected CreatedAt 2025-06-01T12:00:00Z, got %s", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_NonNullModelFieldsAPIEmpty(t *testing.T) {
	// When model has non-null optional fields but API returns empty, they should stay unchanged
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Value(5),
		URLPath:        types.StringValue("/existing"),
		ExpectedCodes:  types.StringValue("200"),
		HTTPMethod:     types.StringValue("GET"),
	}

	apiResp := &hmAPIResponse{
		ID:         "hm-preserve",
		Name:       "preserve-hm",
		PoolID:     "pool-preserve",
		Type:       "HTTP",
		Delay:      10,
		Timeout:    5,
		MaxRetries: 3,
		// All optional fields are empty/zero
	}

	mapAPIResponseToState(model, apiResp)

	// MaxRetriesDown is 0, model is not null => else-if (model.MaxRetriesDown.IsNull()) is false => stays
	if model.MaxRetriesDown.ValueInt64() != 5 {
		t.Errorf("expected MaxRetriesDown to remain 5, got %d", model.MaxRetriesDown.ValueInt64())
	}
	// URLPath is empty, model is not null => stays
	if model.URLPath.ValueString() != "/existing" {
		t.Errorf("expected URLPath to remain /existing, got %s", model.URLPath.ValueString())
	}
	// ExpectedCodes is empty, model is not null => stays
	if model.ExpectedCodes.ValueString() != "200" {
		t.Errorf("expected ExpectedCodes to remain 200, got %s", model.ExpectedCodes.ValueString())
	}
	// HTTPMethod is empty, model is not null => stays
	if model.HTTPMethod.ValueString() != "GET" {
		t.Errorf("expected HTTPMethod to remain GET, got %s", model.HTTPMethod.ValueString())
	}
}

func TestMapAPIResponseToState_PendingStatuses(t *testing.T) {
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:                 "hm-pending",
		Name:               "pending-hm",
		Type:               "TCP",
		Delay:              10,
		Timeout:            5,
		MaxRetries:         3,
		ProvisioningStatus: "PENDING_DELETE",
		OperatingStatus:    "OFFLINE",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ProvisioningStatus.ValueString() != "PENDING_DELETE" {
		t.Errorf("expected ProvisioningStatus PENDING_DELETE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "OFFLINE" {
		t.Errorf("expected OperatingStatus OFFLINE, got %s", model.OperatingStatus.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyPoolIDAndType(t *testing.T) {
	model := &lbHealthMonitorResourceModel{
		PoolID:         types.StringValue("existing-pool"),
		Type:           types.StringValue("HTTP"),
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:         "hm-no-pool",
		Name:       "no-pool-hm",
		Delay:      10,
		Timeout:    5,
		MaxRetries: 3,
	}

	mapAPIResponseToState(model, apiResp)

	// PoolID is empty but the condition is `if apiResp.PoolID != ""`, so model stays
	if model.PoolID.ValueString() != "existing-pool" {
		t.Errorf("expected PoolID to remain existing-pool, got %s", model.PoolID.ValueString())
	}
	if model.Type.ValueString() != "HTTP" {
		t.Errorf("expected Type to remain HTTP, got %s", model.Type.ValueString())
	}
}

func TestMapAPIResponseToState_ZeroDelay(t *testing.T) {
	// Delay/Timeout/MaxRetries are always set (no conditional check)
	model := &lbHealthMonitorResourceModel{
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
	}

	apiResp := &hmAPIResponse{
		ID:         "hm-zero-delay",
		Name:       "zero-delay",
		Delay:      0,
		Timeout:    0,
		MaxRetries: 0,
	}

	mapAPIResponseToState(model, apiResp)

	if model.Delay.ValueInt64() != 0 {
		t.Errorf("expected Delay 0, got %d", model.Delay.ValueInt64())
	}
	if model.Timeout.ValueInt64() != 0 {
		t.Errorf("expected Timeout 0, got %d", model.Timeout.ValueInt64())
	}
	if model.MaxRetries.ValueInt64() != 0 {
		t.Errorf("expected MaxRetries 0, got %d", model.MaxRetries.ValueInt64())
	}
}

// --- buildCreateRequest all types ---

func TestBuildCreateRequest_AllMonitorTypes(t *testing.T) {
	monitorTypes := []string{"HTTP", "HTTPS", "TCP", "PING", "TLS-HELLO", "UDP-CONNECT", "SCTP"}

	for _, mt := range monitorTypes {
		t.Run(mt, func(t *testing.T) {
			plan := &lbHealthMonitorResourceModel{
				Name:           types.StringValue("hm-" + mt),
				PoolID:         types.StringValue("pool-type-test"),
				Type:           types.StringValue(mt),
				Delay:          types.Int64Value(5),
				Timeout:        types.Int64Value(3),
				MaxRetries:     types.Int64Value(3),
				MaxRetriesDown: types.Int64Null(),
				URLPath:        types.StringNull(),
				ExpectedCodes:  types.StringNull(),
				HTTPMethod:     types.StringNull(),
			}

			body := buildCreateRequest(plan)

			if body.Type != mt {
				t.Errorf("expected type %s, got %s", mt, body.Type)
			}
			if body.Name != "hm-"+mt {
				t.Errorf("expected name hm-%s, got %s", mt, body.Name)
			}
		})
	}
}

// --- buildUpdateRequest all HTTP methods ---

func TestBuildUpdateRequest_AllHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "CONNECT", "TRACE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			plan := &lbHealthMonitorResourceModel{
				Name:           types.StringValue("update-" + method),
				Delay:          types.Int64Value(10),
				Timeout:        types.Int64Value(5),
				MaxRetries:     types.Int64Value(3),
				MaxRetriesDown: types.Int64Null(),
				URLPath:        types.StringValue("/health"),
				ExpectedCodes:  types.StringValue("200"),
				HTTPMethod:     types.StringValue(method),
			}

			body := buildUpdateRequest(plan)

			if body.HTTPMethod != method {
				t.Errorf("expected http_method %s, got %s", method, body.HTTPMethod)
			}
		})
	}
}

// --- JSON round-trip ---

func TestHMAPIResponse_JSONDeserialization(t *testing.T) {
	raw := `{
		"id": "hm-rt-001",
		"name": "roundtrip-hm",
		"pool_id": "pool-rt-001",
		"type": "HTTPS",
		"delay": 15,
		"timeout": 10,
		"max_retries": 5,
		"max_retries_down": 3,
		"url_path": "/ready",
		"expected_codes": "200-299",
		"http_method": "HEAD",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"admin_state_up": true,
		"created_at": "2025-03-01T10:00:00Z"
	}`

	var resp hmAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "hm-rt-001" {
		t.Errorf("expected ID hm-rt-001, got %s", resp.ID)
	}
	if resp.PoolID != "pool-rt-001" {
		t.Errorf("expected PoolID pool-rt-001, got %s", resp.PoolID)
	}
	if resp.Type != "HTTPS" {
		t.Errorf("expected Type HTTPS, got %s", resp.Type)
	}
	if resp.Delay != 15 {
		t.Errorf("expected Delay 15, got %d", resp.Delay)
	}
	if resp.MaxRetriesDown != 3 {
		t.Errorf("expected MaxRetriesDown 3, got %d", resp.MaxRetriesDown)
	}
	if resp.URLPath != "/ready" {
		t.Errorf("expected URLPath /ready, got %s", resp.URLPath)
	}
	if resp.HTTPMethod != "HEAD" {
		t.Errorf("expected HTTPMethod HEAD, got %s", resp.HTTPMethod)
	}
}

func TestHMAPIResponse_JSONWithExtraFields(t *testing.T) {
	raw := `{
		"id": "hm-extra-001",
		"name": "extra",
		"type": "TCP",
		"delay": 5,
		"timeout": 3,
		"max_retries": 3,
		"unknown_field": "ignored",
		"extra_number": 999
	}`

	var resp hmAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal with extra fields: %v", err)
	}

	if resp.ID != "hm-extra-001" {
		t.Errorf("expected ID hm-extra-001, got %s", resp.ID)
	}
}

func TestHMAPIResponse_JSONNullFields(t *testing.T) {
	raw := `{
		"id": "hm-null-001",
		"name": "null-hm",
		"pool_id": null,
		"type": null,
		"delay": 0,
		"timeout": 0,
		"max_retries": 0,
		"max_retries_down": 0,
		"url_path": null,
		"expected_codes": null,
		"http_method": null,
		"admin_state_up": false
	}`

	var resp hmAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal with null fields: %v", err)
	}

	if resp.PoolID != "" {
		t.Errorf("expected empty PoolID for null, got %s", resp.PoolID)
	}
	if resp.Type != "" {
		t.Errorf("expected empty Type for null, got %s", resp.Type)
	}
	if resp.URLPath != "" {
		t.Errorf("expected empty URLPath for null, got %s", resp.URLPath)
	}
	if resp.AdminStateUp != false {
		t.Error("expected AdminStateUp false")
	}
}

// --- Schema tests ---

func TestLbHealthMonitorSchema(t *testing.T) {
	s := lbHealthMonitorSchema()

	if s.Description == "" {
		t.Error("expected non-empty schema description")
	}

	// Verify required string attributes
	requiredStr := []string{"name", "pool_id", "type"}
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

	// Verify required int64 attributes
	requiredInt := []string{"delay", "timeout", "max_retries"}
	for _, attr := range requiredInt {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
			continue
		}
		ia, ok := a.(schema.Int64Attribute)
		if ok && !ia.IsRequired() {
			t.Errorf("expected attribute %q to be required", attr)
		}
	}

	// Verify optional string attributes
	optionalStr := []string{"url_path", "expected_codes", "http_method"}
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

	// Verify computed attributes
	computed := []string{"id", "provisioning_status", "operating_status", "admin_state_up", "created_at"}
	for _, attr := range computed {
		a, ok := s.Attributes[attr]
		if !ok {
			t.Errorf("expected attribute %q to exist", attr)
		}
		_ = a
	}

	// max_retries_down should be optional + computed
	mrdAttr, ok := s.Attributes["max_retries_down"]
	if !ok {
		t.Fatal("expected max_retries_down attribute to exist")
	}
	mrdIA, ok := mrdAttr.(schema.Int64Attribute)
	if !ok {
		t.Fatal("expected max_retries_down to be Int64Attribute")
	}
	if !mrdIA.IsOptional() {
		t.Error("expected max_retries_down to be optional")
	}
	if !mrdIA.IsComputed() {
		t.Error("expected max_retries_down to be computed")
	}
}

// --- buildCreateRequest JSON omitempty ---

func TestBuildCreateRequest_JSONOmitEmpty(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("omit-hm"),
		PoolID:         types.StringValue("pool-omit"),
		Type:           types.StringValue("TCP"),
		Delay:          types.Int64Value(10),
		Timeout:        types.Int64Value(5),
		MaxRetries:     types.Int64Value(3),
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
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

	// Fields with omitempty should not appear when zero/empty
	if _, exists := parsed["max_retries_down"]; exists {
		t.Error("expected max_retries_down to be omitted when 0 (omitempty)")
	}
	if _, exists := parsed["url_path"]; exists {
		t.Error("expected url_path to be omitted when empty (omitempty)")
	}
	if _, exists := parsed["expected_codes"]; exists {
		t.Error("expected expected_codes to be omitted when empty (omitempty)")
	}
	if _, exists := parsed["http_method"]; exists {
		t.Error("expected http_method to be omitted when empty (omitempty)")
	}

	// Required fields should always be present
	if _, exists := parsed["name"]; !exists {
		t.Error("expected name to be present")
	}
	if _, exists := parsed["pool_id"]; !exists {
		t.Error("expected pool_id to be present")
	}
	if _, exists := parsed["type"]; !exists {
		t.Error("expected type to be present")
	}
	if _, exists := parsed["delay"]; !exists {
		t.Error("expected delay to be present")
	}
}

// --- buildUpdateRequest does not include pool_id or type ---

func TestBuildUpdateRequest_ExcludesImmutableFields(t *testing.T) {
	plan := &lbHealthMonitorResourceModel{
		Name:           types.StringValue("update-test"),
		PoolID:         types.StringValue("pool-should-not-appear"),
		Type:           types.StringValue("TCP"),
		Delay:          types.Int64Value(10),
		Timeout:        types.Int64Value(5),
		MaxRetries:     types.Int64Value(3),
		MaxRetriesDown: types.Int64Null(),
		URLPath:        types.StringNull(),
		ExpectedCodes:  types.StringNull(),
		HTTPMethod:     types.StringNull(),
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

	if _, exists := parsed["pool_id"]; exists {
		t.Error("update request should NOT contain pool_id (immutable)")
	}
	if _, exists := parsed["type"]; exists {
		t.Error("update request should NOT contain type (immutable)")
	}
}

// --- Delete request tests ---

func TestHMDeleteRequest_MultipleValues(t *testing.T) {
	req := hmDeleteRequest{
		Key:    "id",
		Values: []string{"hm-1", "hm-2", "hm-3"},
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
}

func TestHMDeleteRequest_EmptyValues(t *testing.T) {
	req := hmDeleteRequest{
		Key:    "id",
		Values: []string{},
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
	if len(values) != 0 {
		t.Fatalf("expected 0 values, got %d", len(values))
	}
}
