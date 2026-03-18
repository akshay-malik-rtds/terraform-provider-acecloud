package load_balancer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &loadBalancerResourceModel{
		Name:        types.StringValue("test-lb"),
		SubnetID:    types.StringValue("subnet-123"),
		Description: types.StringValue("A test load balancer"),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "test-lb" {
		t.Errorf("expected name test-lb, got %s", body.Name)
	}
	if body.SubnetID != "subnet-123" {
		t.Errorf("expected subnet_id subnet-123, got %s", body.SubnetID)
	}
	if body.Description != "A test load balancer" {
		t.Errorf("expected description 'A test load balancer', got %s", body.Description)
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &loadBalancerResourceModel{
		Name:        types.StringValue("minimal-lb"),
		SubnetID:    types.StringValue("subnet-456"),
		Description: types.StringNull(),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "minimal-lb" {
		t.Errorf("expected name minimal-lb, got %s", body.Name)
	}
	if body.Description != "" {
		t.Errorf("expected empty description, got %s", body.Description)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &loadBalancerResourceModel{
		Name:        types.StringValue("json-lb"),
		SubnetID:    types.StringValue("subnet-json-123"),
		Description: types.StringValue("JSON LB desc"),
	}

	body := buildCreateRequest(context.Background(), plan)

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
	if _, ok := raw["subnet_id"]; !ok {
		t.Error("expected JSON key 'subnet_id' to be present")
	}
	if _, ok := raw["description"]; !ok {
		t.Error("expected JSON key 'description' to be present")
	}

	if raw["name"] != "json-lb" {
		t.Errorf("expected JSON name 'json-lb', got %v", raw["name"])
	}
	if raw["subnet_id"] != "subnet-json-123" {
		t.Errorf("expected JSON subnet_id 'subnet-json-123', got %v", raw["subnet_id"])
	}
	if raw["description"] != "JSON LB desc" {
		t.Errorf("expected JSON description 'JSON LB desc', got %v", raw["description"])
	}
}

func TestBuildUpdateRequest(t *testing.T) {
	plan := &loadBalancerResourceModel{
		Name:        types.StringValue("updated-lb"),
		Description: types.StringValue("Updated desc"),
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.Name != "updated-lb" {
		t.Errorf("expected name updated-lb, got %s", body.Name)
	}
	if body.Description != "Updated desc" {
		t.Errorf("expected description 'Updated desc', got %s", body.Description)
	}
}

func TestBuildUpdateRequest_NoDescription(t *testing.T) {
	plan := &loadBalancerResourceModel{
		Name:        types.StringValue("no-desc-lb"),
		Description: types.StringNull(),
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.Name != "no-desc-lb" {
		t.Errorf("expected name no-desc-lb, got %s", body.Name)
	}
	if body.Description != "" {
		t.Errorf("expected empty description when plan description is null, got %s", body.Description)
	}

	// Also verify JSON omits the description field when empty (omitempty tag)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal update request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if _, ok := raw["description"]; ok {
		t.Error("expected 'description' to be omitted from JSON when empty (omitempty)")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-abc-123",
		Name:               "prod-lb",
		SubnetID:           "subnet-xyz",
		Description:        "Production LB",
		VIPAddress:         "10.0.0.100",
		VIPPortID:          "port-vip",
		VIPNetworkID:       "net-vip",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		Provider:           "octavia",
		CreatedAt:          "2024-01-01T00:00:00Z",
		UpdatedAt:          "2024-01-02T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "lb-abc-123" {
		t.Errorf("expected ID lb-abc-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-lb" {
		t.Errorf("expected Name prod-lb, got %s", model.Name.ValueString())
	}
	if model.SubnetID.ValueString() != "subnet-xyz" {
		t.Errorf("expected SubnetID subnet-xyz, got %s", model.SubnetID.ValueString())
	}
	if model.VIPAddress.ValueString() != "10.0.0.100" {
		t.Errorf("expected VIPAddress 10.0.0.100, got %s", model.VIPAddress.ValueString())
	}
	if model.ProvisioningStatus.ValueString() != "ACTIVE" {
		t.Errorf("expected ProvisioningStatus ACTIVE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "ONLINE" {
		t.Errorf("expected OperatingStatus ONLINE, got %s", model.OperatingStatus.ValueString())
	}
	if model.Provider.ValueString() != "octavia" {
		t.Errorf("expected Provider octavia, got %s", model.Provider.ValueString())
	}
}

func TestMapAPIResponseToState_AllComputedFields(t *testing.T) {
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-full-001",
		Name:               "full-lb",
		SubnetID:           "subnet-full-xyz",
		Description:        "Full LB with all computed fields",
		VIPAddress:         "192.168.1.50",
		VIPPortID:          "port-full-vip",
		VIPNetworkID:       "net-full-vip",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		Provider:           "amphora",
		CreatedAt:          "2024-06-15T10:30:00Z",
		UpdatedAt:          "2024-06-16T14:45:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "lb-full-001" {
		t.Errorf("expected ID lb-full-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-lb" {
		t.Errorf("expected Name full-lb, got %s", model.Name.ValueString())
	}
	if model.SubnetID.ValueString() != "subnet-full-xyz" {
		t.Errorf("expected SubnetID subnet-full-xyz, got %s", model.SubnetID.ValueString())
	}
	if model.Description.ValueString() != "Full LB with all computed fields" {
		t.Errorf("expected Description 'Full LB with all computed fields', got %s", model.Description.ValueString())
	}
	if model.VIPAddress.ValueString() != "192.168.1.50" {
		t.Errorf("expected VIPAddress 192.168.1.50, got %s", model.VIPAddress.ValueString())
	}
	if model.VIPPortID.ValueString() != "port-full-vip" {
		t.Errorf("expected VIPPortID port-full-vip, got %s", model.VIPPortID.ValueString())
	}
	if model.VIPNetworkID.ValueString() != "net-full-vip" {
		t.Errorf("expected VIPNetworkID net-full-vip, got %s", model.VIPNetworkID.ValueString())
	}
	if model.ProvisioningStatus.ValueString() != "ACTIVE" {
		t.Errorf("expected ProvisioningStatus ACTIVE, got %s", model.ProvisioningStatus.ValueString())
	}
	if model.OperatingStatus.ValueString() != "ONLINE" {
		t.Errorf("expected OperatingStatus ONLINE, got %s", model.OperatingStatus.ValueString())
	}
	if model.Provider.ValueString() != "amphora" {
		t.Errorf("expected Provider amphora, got %s", model.Provider.ValueString())
	}
	if model.CreatedAt.ValueString() != "2024-06-15T10:30:00Z" {
		t.Errorf("expected CreatedAt '2024-06-15T10:30:00Z', got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2024-06-16T14:45:00Z" {
		t.Errorf("expected UpdatedAt '2024-06-16T14:45:00Z', got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-123",
		Name:               "basic",
		Description:        "",
		ProvisioningStatus: "PENDING_CREATE",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
}

func TestLBDeleteRequest(t *testing.T) {
	req := lbDeleteRequest{
		Key:    "id",
		Values: []string{"lb-del-001"},
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
	if values[0] != "lb-del-001" {
		t.Errorf("expected value 'lb-del-001', got %v", values[0])
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// ---------------------------------------------------------------------------
// parseLBData Tests
// ---------------------------------------------------------------------------

func TestParseLBData_CreateResponse(t *testing.T) {
	data := json.RawMessage(`{
		"id": "lb-123",
		"name": "my-lb",
		"description": "test",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"provider": "amphora",
		"created_at": "2024-01-01",
		"updated_at": "2024-01-02",
		"tags": ["web"],
		"address": "10.0.0.5",
		"port_id": "port-1",
		"subnet_id": "sub-1",
		"network_id": "net-1"
	}`)

	resp, err := parseLBData(data)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if resp.ID != "lb-123" { t.Errorf("expected lb-123, got %s", resp.ID) }
	if resp.VIPAddress != "10.0.0.5" { t.Errorf("expected 10.0.0.5, got %s", resp.VIPAddress) }
	if resp.VIPPortID != "port-1" { t.Errorf("expected port-1, got %s", resp.VIPPortID) }
	if resp.SubnetID != "sub-1" { t.Errorf("expected sub-1, got %s", resp.SubnetID) }
	if resp.VIPNetworkID != "net-1" { t.Errorf("expected net-1, got %s", resp.VIPNetworkID) }
}

func TestParseLBData_ReadResponse(t *testing.T) {
	data := json.RawMessage(`{
		"id": "lb-456",
		"name": "my-lb",
		"provisioning_status": "ACTIVE",
		"operating_status": "ONLINE",
		"vip_address": "10.0.0.10",
		"vip_port_id": "vport-1",
		"vip_subnet_id": "vsub-1",
		"vip_network_id": "vnet-1"
	}`)

	resp, err := parseLBData(data)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if resp.VIPAddress != "10.0.0.10" { t.Errorf("expected 10.0.0.10, got %s", resp.VIPAddress) }
	if resp.VIPPortID != "vport-1" { t.Errorf("expected vport-1, got %s", resp.VIPPortID) }
	if resp.SubnetID != "vsub-1" { t.Errorf("expected vsub-1, got %s", resp.SubnetID) }
	if resp.VIPNetworkID != "vnet-1" { t.Errorf("expected vnet-1, got %s", resp.VIPNetworkID) }
}

func TestParseLBData_EmptyFields(t *testing.T) {
	data := json.RawMessage(`{"id":"lb-789","name":"empty"}`)
	resp, err := parseLBData(data)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if resp.VIPAddress != "" { t.Errorf("expected empty VIPAddress, got %s", resp.VIPAddress) }
	if resp.SubnetID != "" { t.Errorf("expected empty SubnetID, got %s", resp.SubnetID) }
}

func TestParseLBData_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`not json`)
	_, err := parseLBData(data)
	if err == nil { t.Fatal("expected error for invalid JSON") }
}

// ---------------------------------------------------------------------------
// firstStr Tests
// ---------------------------------------------------------------------------

func TestFirstStr_FirstKeyMatches(t *testing.T) {
	m := map[string]interface{}{"address": "10.0.0.1", "vip_address": "10.0.0.2"}
	got := firstStr(m, "address", "vip_address")
	if got != "10.0.0.1" { t.Errorf("expected 10.0.0.1, got %s", got) }
}

func TestFirstStr_SecondKeyMatches(t *testing.T) {
	m := map[string]interface{}{"vip_address": "10.0.0.2"}
	got := firstStr(m, "address", "vip_address")
	if got != "10.0.0.2" { t.Errorf("expected 10.0.0.2, got %s", got) }
}

func TestFirstStr_NoMatch(t *testing.T) {
	m := map[string]interface{}{"other": "value"}
	got := firstStr(m, "address", "vip_address")
	if got != "" { t.Errorf("expected empty, got %s", got) }
}

func TestFirstStr_EmptyStringSkipped(t *testing.T) {
	m := map[string]interface{}{"address": "", "vip_address": "10.0.0.2"}
	got := firstStr(m, "address", "vip_address")
	if got != "10.0.0.2" { t.Errorf("expected 10.0.0.2 (skip empty), got %s", got) }
}

func TestFirstStr_NonStringSkipped(t *testing.T) {
	m := map[string]interface{}{"address": 12345, "vip_address": "10.0.0.2"}
	got := firstStr(m, "address", "vip_address")
	if got != "10.0.0.2" { t.Errorf("expected 10.0.0.2 (skip non-string), got %s", got) }
}

// ---------------------------------------------------------------------------
// stringValueOrNull Tests
// ---------------------------------------------------------------------------

func TestStringValueOrNull_NonEmpty(t *testing.T) {
	result := stringValueOrNull("hello")
	if result.IsNull() { t.Error("expected non-null for non-empty string") }
	if result.ValueString() != "hello" { t.Errorf("expected 'hello', got %s", result.ValueString()) }
}

func TestStringValueOrNull_Empty(t *testing.T) {
	result := stringValueOrNull("")
	if !result.IsNull() { t.Error("expected null for empty string") }
}

// ===========================================================================
// NEW TESTS: mapAPIResponseToState, tagsEqual, tag ordering, stringValueOrNull
// ===========================================================================

// ---------------------------------------------------------------------------
// mapAPIResponseToState — comprehensive field mapping
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_EmptySubnetID(t *testing.T) {
	// If SubnetID is empty in API response, existing model SubnetID should be preserved.
	model := &loadBalancerResourceModel{
		SubnetID:    types.StringValue("existing-subnet"),
		Description: types.StringNull(),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-no-subnet",
		Name:               "no-subnet-lb",
		SubnetID:           "", // empty
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	// SubnetID should remain the existing value (the if block skips empty)
	if model.SubnetID.ValueString() != "existing-subnet" {
		t.Errorf("expected SubnetID to be preserved as 'existing-subnet', got %s", model.SubnetID.ValueString())
	}
	if model.ID.ValueString() != "lb-no-subnet" {
		t.Errorf("expected ID lb-no-subnet, got %s", model.ID.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyTagsToNull(t *testing.T) {
	// When API returns empty tags and model tags are null, should stay null.
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListNull(types.StringType),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-no-tags",
		Name:               "no-tags-lb",
		Tags:               nil, // empty/nil tags
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Tags.IsNull() {
		t.Error("expected Tags to remain null when API returns no tags")
	}
}

func TestMapAPIResponseToState_EmptyTagsSliceToNull(t *testing.T) {
	// When API returns empty slice [] and model is null, should stay null.
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListNull(types.StringType),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-empty-tags",
		Name:               "empty-tags-lb",
		Tags:               []string{}, // empty slice
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Tags.IsNull() {
		t.Error("expected Tags to remain null when API returns empty tags slice")
	}
}

func TestMapAPIResponseToState_TagsFromNull(t *testing.T) {
	// When model tags are null and API returns tags, should populate.
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListNull(types.StringType),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-new-tags",
		Name:               "new-tags-lb",
		Tags:               []string{"ALB", "web"},
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	if model.Tags.IsNull() {
		t.Fatal("expected Tags to be populated, got null")
	}

	var tags []string
	model.Tags.ElementsAs(context.Background(), &tags, false)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "ALB" {
		t.Errorf("expected first tag 'ALB', got %s", tags[0])
	}
	if tags[1] != "web" {
		t.Errorf("expected second tag 'web', got %s", tags[1])
	}
}

func TestMapAPIResponseToState_TagsFromUnknown(t *testing.T) {
	// When model tags are unknown and API returns tags, should populate.
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListUnknown(types.StringType),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-unknown-tags",
		Name:               "unknown-tags-lb",
		Tags:               []string{"NLB"},
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	if model.Tags.IsUnknown() {
		t.Fatal("expected Tags to be resolved from unknown, still unknown")
	}
	if model.Tags.IsNull() {
		t.Fatal("expected Tags to be populated, got null")
	}

	var tags []string
	model.Tags.ElementsAs(context.Background(), &tags, false)
	if len(tags) != 1 || tags[0] != "NLB" {
		t.Errorf("expected ['NLB'], got %v", tags)
	}
}

func TestMapAPIResponseToState_DescriptionSetThenCleared(t *testing.T) {
	// Model already has a description value; API returns empty string.
	// Description should keep the existing non-null value (the code only sets null
	// if model was already null/unknown).
	model := &loadBalancerResourceModel{
		Description: types.StringValue("old description"),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-desc-clear",
		Name:               "desc-clear-lb",
		Description:        "",
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	// When description is not null/unknown and API returns "", the code doesn't change it.
	if model.Description.ValueString() != "old description" {
		t.Errorf("expected Description preserved as 'old description', got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_AllComputedNulledWhenEmpty(t *testing.T) {
	// All computed fields should be null when API returns empty strings.
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListNull(types.StringType),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-empty-computed",
		Name:               "empty-computed",
		VIPAddress:         "",
		VIPPortID:          "",
		VIPNetworkID:       "",
		ProvisioningStatus: "",
		OperatingStatus:    "",
		Provider:           "",
		CreatedAt:          "",
		UpdatedAt:          "",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.VIPAddress.IsNull() {
		t.Error("expected VIPAddress null for empty string")
	}
	if !model.VIPPortID.IsNull() {
		t.Error("expected VIPPortID null for empty string")
	}
	if !model.VIPNetworkID.IsNull() {
		t.Error("expected VIPNetworkID null for empty string")
	}
	if !model.ProvisioningStatus.IsNull() {
		t.Error("expected ProvisioningStatus null for empty string")
	}
	if !model.OperatingStatus.IsNull() {
		t.Error("expected OperatingStatus null for empty string")
	}
	if !model.Provider.IsNull() {
		t.Error("expected Provider null for empty string")
	}
	if !model.CreatedAt.IsNull() {
		t.Error("expected CreatedAt null for empty string")
	}
	if !model.UpdatedAt.IsNull() {
		t.Error("expected UpdatedAt null for empty string")
	}
}

func TestMapAPIResponseToState_DescriptionUnknownToNull(t *testing.T) {
	// When model Description is unknown and API returns empty, should become null.
	model := &loadBalancerResourceModel{
		Description: types.StringUnknown(),
	}

	apiResp := &lbAPIResponse{
		ID:          "lb-desc-unknown",
		Name:        "desc-unknown-lb",
		Description: "",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to become null when unknown and API returns empty")
	}
}

// ---------------------------------------------------------------------------
// tagsEqual helper
// ---------------------------------------------------------------------------

func TestTagsEqual_SameOrderSameElements(t *testing.T) {
	a := []string{"ALB", "web", "prod"}
	b := []string{"ALB", "web", "prod"}
	if !tagsEqual(a, b) {
		t.Error("expected true for identical slices")
	}
}

func TestTagsEqual_DifferentOrderSameElements(t *testing.T) {
	a := []string{"ALB", "web", "prod"}
	b := []string{"prod", "ALB", "web"}
	if !tagsEqual(a, b) {
		t.Error("expected true for same elements in different order")
	}
}

func TestTagsEqual_DifferentElements(t *testing.T) {
	a := []string{"ALB", "web"}
	b := []string{"NLB", "web"}
	if tagsEqual(a, b) {
		t.Error("expected false for different elements")
	}
}

func TestTagsEqual_DifferentLengths(t *testing.T) {
	a := []string{"ALB", "web"}
	b := []string{"ALB", "web", "extra"}
	if tagsEqual(a, b) {
		t.Error("expected false for different lengths")
	}
}

func TestTagsEqual_DifferentLengths_Reversed(t *testing.T) {
	a := []string{"ALB", "web", "extra"}
	b := []string{"ALB", "web"}
	if tagsEqual(a, b) {
		t.Error("expected false for different lengths (reversed)")
	}
}

func TestTagsEqual_EmptySlices(t *testing.T) {
	a := []string{}
	b := []string{}
	if !tagsEqual(a, b) {
		t.Error("expected true for two empty slices")
	}
}

func TestTagsEqual_NilSlices(t *testing.T) {
	var a, b []string
	if !tagsEqual(a, b) {
		t.Error("expected true for two nil slices")
	}
}

func TestTagsEqual_OneEmpty_OnePop(t *testing.T) {
	a := []string{}
	b := []string{"ALB"}
	if tagsEqual(a, b) {
		t.Error("expected false: empty vs populated")
	}
}

func TestTagsEqual_DuplicateElements_Same(t *testing.T) {
	a := []string{"ALB", "ALB", "web"}
	b := []string{"ALB", "web", "ALB"}
	if !tagsEqual(a, b) {
		t.Error("expected true for same duplicate elements in different order")
	}
}

func TestTagsEqual_DuplicateElements_Different(t *testing.T) {
	// Same elements but different duplicate counts
	a := []string{"ALB", "ALB", "web"}
	b := []string{"ALB", "web", "web"}
	if tagsEqual(a, b) {
		t.Error("expected false for different duplicate counts")
	}
}

func TestTagsEqual_SingleElement(t *testing.T) {
	a := []string{"NLB"}
	b := []string{"NLB"}
	if !tagsEqual(a, b) {
		t.Error("expected true for single identical element")
	}
}

func TestTagsEqual_SingleDifferent(t *testing.T) {
	a := []string{"ALB"}
	b := []string{"NLB"}
	if tagsEqual(a, b) {
		t.Error("expected false for single different element")
	}
}

// ---------------------------------------------------------------------------
// Tags ordering preservation
// ---------------------------------------------------------------------------

func TestTagsOrdering_PreserveExistingOrder(t *testing.T) {
	// When model has ["a","b","c"] and API returns ["c","a","b"], model keeps ["a","b","c"].
	existingTags := []attr.Value{
		types.StringValue("a"),
		types.StringValue("b"),
		types.StringValue("c"),
	}
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListValueMust(types.StringType, existingTags),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-tag-order",
		Name:               "tag-order-lb",
		Tags:               []string{"c", "a", "b"}, // different order
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	// Tags should preserve the original order ["a", "b", "c"]
	var resultTags []string
	model.Tags.ElementsAs(context.Background(), &resultTags, false)

	if len(resultTags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(resultTags))
	}
	if resultTags[0] != "a" || resultTags[1] != "b" || resultTags[2] != "c" {
		t.Errorf("expected tags order [a, b, c] preserved, got %v", resultTags)
	}
}

func TestTagsOrdering_DifferentElements_UpdateOrder(t *testing.T) {
	// When model has ["a","b"] and API returns ["c","d"], should use API order.
	existingTags := []attr.Value{
		types.StringValue("a"),
		types.StringValue("b"),
	}
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListValueMust(types.StringType, existingTags),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-tag-diff",
		Name:               "tag-diff-lb",
		Tags:               []string{"c", "d"}, // completely different elements
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	var resultTags []string
	model.Tags.ElementsAs(context.Background(), &resultTags, false)

	if len(resultTags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(resultTags))
	}
	// Should use API order since elements differ
	if resultTags[0] != "c" || resultTags[1] != "d" {
		t.Errorf("expected tags from API [c, d], got %v", resultTags)
	}
}

func TestTagsOrdering_SingleTag_Preserved(t *testing.T) {
	existingTags := []attr.Value{
		types.StringValue("ALB"),
	}
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListValueMust(types.StringType, existingTags),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-single-tag",
		Name:               "single-tag-lb",
		Tags:               []string{"ALB"},
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	var resultTags []string
	model.Tags.ElementsAs(context.Background(), &resultTags, false)

	if len(resultTags) != 1 || resultTags[0] != "ALB" {
		t.Errorf("expected ['ALB'], got %v", resultTags)
	}
}

func TestTagsOrdering_APIDifferentLength_UpdateFromAPI(t *testing.T) {
	// When model has ["a","b","c"] and API returns ["a","b"], different length = update.
	existingTags := []attr.Value{
		types.StringValue("a"),
		types.StringValue("b"),
		types.StringValue("c"),
	}
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListValueMust(types.StringType, existingTags),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-tag-len",
		Name:               "tag-len-lb",
		Tags:               []string{"a", "b"}, // fewer tags
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	var resultTags []string
	model.Tags.ElementsAs(context.Background(), &resultTags, false)

	if len(resultTags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(resultTags))
	}
	if resultTags[0] != "a" || resultTags[1] != "b" {
		t.Errorf("expected [a, b] from API, got %v", resultTags)
	}
}

func TestTagsOrdering_EmptyAPIResponse_PreserveExisting(t *testing.T) {
	// When API returns empty tags but model has tags, should preserve existing.
	existingTags := []attr.Value{
		types.StringValue("ALB"),
		types.StringValue("web"),
	}
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListValueMust(types.StringType, existingTags),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-empty-api-tags",
		Name:               "empty-api-tags-lb",
		Tags:               nil, // API returns no tags
		ProvisioningStatus: "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	// When API returns empty tags, the code doesn't modify existing non-null tags
	// (the len(apiResp.Tags) > 0 branch is skipped, and the else branch only
	// applies when model.Tags.IsNull() || model.Tags.IsUnknown())
	var resultTags []string
	model.Tags.ElementsAs(context.Background(), &resultTags, false)

	if len(resultTags) != 2 {
		t.Fatalf("expected 2 tags preserved, got %d", len(resultTags))
	}
	if resultTags[0] != "ALB" || resultTags[1] != "web" {
		t.Errorf("expected [ALB, web] preserved, got %v", resultTags)
	}
}

// ---------------------------------------------------------------------------
// stringValueOrNull — additional tests
// ---------------------------------------------------------------------------

func TestStringValueOrNull_Whitespace(t *testing.T) {
	// Whitespace-only string is non-empty, should return a value
	result := stringValueOrNull("  ")
	if result.IsNull() {
		t.Error("expected non-null for whitespace-only string")
	}
	if result.ValueString() != "  " {
		t.Errorf("expected '  ', got '%s'", result.ValueString())
	}
}

func TestStringValueOrNull_SpecialChars(t *testing.T) {
	result := stringValueOrNull("2024-01-01T00:00:00Z")
	if result.IsNull() {
		t.Error("expected non-null for timestamp string")
	}
	if result.ValueString() != "2024-01-01T00:00:00Z" {
		t.Errorf("expected '2024-01-01T00:00:00Z', got %s", result.ValueString())
	}
}

func TestStringValueOrNull_UUID(t *testing.T) {
	result := stringValueOrNull("550e8400-e29b-41d4-a716-446655440000")
	if result.IsNull() {
		t.Error("expected non-null for UUID string")
	}
	if result.ValueString() != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected UUID, got %s", result.ValueString())
	}
}

func TestStringValueOrNull_IsNotUnknown(t *testing.T) {
	// Verify that stringValueOrNull never returns unknown
	result := stringValueOrNull("value")
	if result.IsUnknown() {
		t.Error("stringValueOrNull should never return unknown")
	}

	result2 := stringValueOrNull("")
	if result2.IsUnknown() {
		t.Error("stringValueOrNull should never return unknown for empty string")
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState — complete round-trip with tags
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_WithTags(t *testing.T) {
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
		Tags:        types.ListNull(types.StringType),
	}

	apiResp := &lbAPIResponse{
		ID:                 "lb-with-tags",
		Name:               "tagged-lb",
		SubnetID:           "subnet-tagged",
		Description:        "Tagged LB",
		Tags:               []string{"ALB", "production", "web"},
		VIPAddress:         "10.0.0.200",
		VIPPortID:          "port-tagged",
		VIPNetworkID:       "net-tagged",
		ProvisioningStatus: "ACTIVE",
		OperatingStatus:    "ONLINE",
		Provider:           "octavia",
		CreatedAt:          "2024-03-01T12:00:00Z",
		UpdatedAt:          "2024-03-02T15:30:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	// Verify all fields
	if model.ID.ValueString() != "lb-with-tags" {
		t.Errorf("expected ID lb-with-tags, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "tagged-lb" {
		t.Errorf("expected Name tagged-lb, got %s", model.Name.ValueString())
	}
	if model.SubnetID.ValueString() != "subnet-tagged" {
		t.Errorf("expected SubnetID subnet-tagged, got %s", model.SubnetID.ValueString())
	}
	if model.Description.ValueString() != "Tagged LB" {
		t.Errorf("expected Description 'Tagged LB', got %s", model.Description.ValueString())
	}

	// Verify tags
	var tags []string
	model.Tags.ElementsAs(context.Background(), &tags, false)
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	if tags[0] != "ALB" || tags[1] != "production" || tags[2] != "web" {
		t.Errorf("expected [ALB, production, web], got %v", tags)
	}

	// Verify computed fields
	if model.VIPAddress.ValueString() != "10.0.0.200" {
		t.Errorf("expected VIPAddress 10.0.0.200, got %s", model.VIPAddress.ValueString())
	}
	if model.VIPPortID.ValueString() != "port-tagged" {
		t.Errorf("expected VIPPortID port-tagged, got %s", model.VIPPortID.ValueString())
	}
	if model.VIPNetworkID.ValueString() != "net-tagged" {
		t.Errorf("expected VIPNetworkID net-tagged, got %s", model.VIPNetworkID.ValueString())
	}
	if model.CreatedAt.ValueString() != "2024-03-01T12:00:00Z" {
		t.Errorf("expected CreatedAt 2024-03-01T12:00:00Z, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2024-03-02T15:30:00Z" {
		t.Errorf("expected UpdatedAt 2024-03-02T15:30:00Z, got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionNonEmpty(t *testing.T) {
	// Description goes from null to populated via API
	model := &loadBalancerResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &lbAPIResponse{
		ID:          "lb-desc",
		Name:        "desc-lb",
		Description: "New description",
	}

	mapAPIResponseToState(model, apiResp)

	if model.Description.IsNull() {
		t.Error("expected Description to be set, got null")
	}
	if model.Description.ValueString() != "New description" {
		t.Errorf("expected 'New description', got %s", model.Description.ValueString())
	}
}
