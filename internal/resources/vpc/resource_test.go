package vpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                types.StringValue("test-vpc"),
		Description:         types.StringValue("A test VPC"),
		AdminStateUp:        types.BoolValue(true),
		PortSecurityEnabled: types.BoolValue(false),
		SubnetName:          types.StringValue("test-subnet"),
		SubnetCIDR:          types.StringValue("10.0.0.0/24"),
		SubnetIPVersion:     types.Int64Value(4),
		SubnetEnableDHCP:    types.BoolValue(true),
		SubnetGatewayIP:     types.StringValue("10.0.0.1"),
		SubnetDNSNameservers: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("8.8.8.8"),
			types.StringValue("8.8.4.4"),
		}),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "test-vpc" {
		t.Errorf("expected name test-vpc, got %s", body.Name)
	}
	if body.Description != "A test VPC" {
		t.Errorf("expected description 'A test VPC', got %s", body.Description)
	}
	if body.AdminStateUp == nil || *body.AdminStateUp != true {
		t.Error("expected AdminStateUp to be true")
	}
	if body.PortSecurityEnabled == nil || *body.PortSecurityEnabled != false {
		t.Error("expected PortSecurityEnabled to be false")
	}
	if body.Subnet == nil {
		t.Fatal("expected Subnet to be non-nil")
	}
	if body.Subnet.Name != "test-subnet" {
		t.Errorf("expected subnet name test-subnet, got %s", body.Subnet.Name)
	}
	if body.Subnet.CIDR != "10.0.0.0/24" {
		t.Errorf("expected subnet CIDR 10.0.0.0/24, got %s", body.Subnet.CIDR)
	}
	if body.Subnet.IPVersion != 4 {
		t.Errorf("expected subnet IP version 4, got %d", body.Subnet.IPVersion)
	}
	if body.Subnet.EnableDHCP != true {
		t.Error("expected subnet EnableDHCP to be true")
	}
	if body.Subnet.GatewayIP != "10.0.0.1" {
		t.Errorf("expected subnet gateway 10.0.0.1, got %s", body.Subnet.GatewayIP)
	}
	if len(body.Subnet.DNSNameservers) != 2 || body.Subnet.DNSNameservers[0] != "8.8.8.8" {
		t.Errorf("expected 2 DNS nameservers, got %v", body.Subnet.DNSNameservers)
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                 types.StringValue("minimal-vpc"),
		Description:          types.StringNull(),
		AdminStateUp:         types.BoolNull(),
		PortSecurityEnabled:  types.BoolNull(),
		SubnetName:           types.StringValue("min-subnet"),
		SubnetCIDR:           types.StringValue("10.1.0.0/24"),
		SubnetIPVersion:      types.Int64Value(4),
		SubnetEnableDHCP:     types.BoolValue(true),
		SubnetGatewayIP:      types.StringNull(),
		SubnetDNSNameservers: types.ListNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "minimal-vpc" {
		t.Errorf("expected name minimal-vpc, got %s", body.Name)
	}
	if body.Description != "" {
		t.Errorf("expected empty description, got %s", body.Description)
	}
	if body.AdminStateUp != nil {
		t.Error("expected AdminStateUp to be nil")
	}
	if body.PortSecurityEnabled != nil {
		t.Error("expected PortSecurityEnabled to be nil")
	}
	if body.Subnet == nil {
		t.Fatal("expected Subnet to be non-nil (required)")
	}
	if body.Subnet.GatewayIP != "" {
		t.Errorf("expected empty gateway when null, got %s", body.Subnet.GatewayIP)
	}
	if body.Subnet.DNSNameservers != nil {
		t.Errorf("expected nil DNS when null, got %v", body.Subnet.DNSNameservers)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                 types.StringValue("json-vpc"),
		Description:          types.StringValue("JSON VPC desc"),
		AdminStateUp:         types.BoolValue(true),
		PortSecurityEnabled:  types.BoolNull(),
		SubnetName:           types.StringValue("json-subnet"),
		SubnetCIDR:           types.StringValue("10.2.0.0/24"),
		SubnetIPVersion:      types.Int64Value(4),
		SubnetEnableDHCP:     types.BoolValue(true),
		SubnetGatewayIP:      types.StringNull(),
		SubnetDNSNameservers: types.ListNull(types.StringType),
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

	if _, ok := raw["name"]; !ok {
		t.Error("expected JSON key 'name' to be present")
	}
	if _, ok := raw["description"]; !ok {
		t.Error("expected JSON key 'description' to be present")
	}
	if _, ok := raw["admin_state_up"]; !ok {
		t.Error("expected JSON key 'admin_state_up' to be present")
	}
	// port_security_enabled should be omitted (null in plan)
	if _, ok := raw["port_security_enabled"]; ok {
		t.Error("expected 'port_security_enabled' to be omitted (omitempty)")
	}
	// subnet should always be present
	if _, ok := raw["subnet"]; !ok {
		t.Error("expected JSON key 'subnet' to be present")
	}
}

func TestBuildCreateRequest_JSONNoDescription(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                 types.StringValue("no-desc-vpc"),
		Description:          types.StringNull(),
		AdminStateUp:         types.BoolNull(),
		PortSecurityEnabled:  types.BoolNull(),
		SubnetName:           types.StringValue("sub"),
		SubnetCIDR:           types.StringValue("10.3.0.0/24"),
		SubnetIPVersion:      types.Int64Value(4),
		SubnetEnableDHCP:     types.BoolValue(true),
		SubnetGatewayIP:      types.StringNull(),
		SubnetDNSNameservers: types.ListNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if _, ok := raw["description"]; ok {
		t.Error("expected 'description' to be omitted from JSON when null")
	}
}

func TestBuildUpdateRequest(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                types.StringValue("updated-vpc"),
		Description:         types.StringValue("Updated desc"),
		AdminStateUp:        types.BoolValue(true),
		PortSecurityEnabled: types.BoolValue(false),
	}

	body := buildUpdateRequest(plan)

	if body.Name != "updated-vpc" {
		t.Errorf("expected name updated-vpc, got %s", body.Name)
	}
	if body.Description != "Updated desc" {
		t.Errorf("expected description 'Updated desc', got %s", body.Description)
	}
	if body.AdminStateUp == nil || *body.AdminStateUp != true {
		t.Error("expected AdminStateUp to be true")
	}
	if body.PortSecurityEnabled == nil || *body.PortSecurityEnabled != false {
		t.Error("expected PortSecurityEnabled to be false")
	}
}

func TestBuildUpdateRequest_NoDescription(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                types.StringValue("no-desc-vpc"),
		Description:         types.StringNull(),
		AdminStateUp:        types.BoolNull(),
		PortSecurityEnabled: types.BoolNull(),
	}

	body := buildUpdateRequest(plan)

	if body.Description != "" {
		t.Errorf("expected empty description when null, got %s", body.Description)
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := raw["description"]; ok {
		t.Error("expected 'description' to be omitted from JSON when empty (omitempty)")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &vpcResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &vpcAPIResponse{
		ID:                  "vpc-abc-123",
		Name:                "prod-vpc",
		Description:         "Production VPC",
		AdminStateUp:        true,
		PortSecurityEnabled: true,
		Status:              "ACTIVE",
		MTU:                 1500,
		CreatedAt:           "2024-01-01T00:00:00Z",
		UpdatedAt:           "2024-01-02T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "vpc-abc-123" {
		t.Errorf("expected ID vpc-abc-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-vpc" {
		t.Errorf("expected Name prod-vpc, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Production VPC" {
		t.Errorf("expected Description 'Production VPC', got %s", model.Description.ValueString())
	}
	if model.AdminStateUp.ValueBool() != true {
		t.Error("expected AdminStateUp to be true")
	}
	if model.PortSecurityEnabled.ValueBool() != true {
		t.Error("expected PortSecurityEnabled to be true")
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
	if model.MTU.ValueInt64() != 1500 {
		t.Errorf("expected MTU 1500, got %d", model.MTU.ValueInt64())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &vpcResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &vpcAPIResponse{
		ID:     "vpc-123",
		Name:   "basic-vpc",
		Status: "BUILD",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
	if model.MTU.ValueInt64() != 0 {
		t.Errorf("expected MTU 0 when API returns 0, got %d", model.MTU.ValueInt64())
	}
	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %s", model.CreatedAt.ValueString())
	}
}

func TestParseSubnetFromRaw_CreateResponse(t *testing.T) {
	raw := map[string]json.RawMessage{
		"id":     json.RawMessage(`"vpc-123"`),
		"subnet": json.RawMessage(`{"id":"subnet-456","gateway_ip":"10.0.0.1"}`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "subnet-456" {
		t.Errorf("expected subnet ID subnet-456, got %s", subnetID)
	}
	if gatewayIP != "10.0.0.1" {
		t.Errorf("expected gateway IP 10.0.0.1, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_ReadResponse(t *testing.T) {
	raw := map[string]json.RawMessage{
		"id":      json.RawMessage(`"vpc-123"`),
		"subnets": json.RawMessage(`["subnet-789"]`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "subnet-789" {
		t.Errorf("expected subnet ID subnet-789, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gateway IP from string array, got %s", gatewayIP)
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestVPCDeleteRequest_JSON(t *testing.T) {
	req := vpcDeleteRequest{
		Key:    "id",
		Values: []string{"vpc-abc-123"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal vpcDeleteRequest: %v", err)
	}

	expected := `{"key":"id","values":["vpc-abc-123"]}`
	if string(data) != expected {
		t.Errorf("expected JSON %s, got %s", expected, string(data))
	}
}

// ---------------------------------------------------------------------------
// parseSubnetFromRaw edge cases
// ---------------------------------------------------------------------------

func TestParseSubnetFromRaw_NoSubnetKey(t *testing.T) {
	// JSON with no "subnet" or "subnets" key should return empty strings.
	raw := map[string]json.RawMessage{
		"id":   json.RawMessage(`"vpc-999"`),
		"name": json.RawMessage(`"no-subnet-vpc"`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "" {
		t.Errorf("expected empty subnetID, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_EmptySubnetsArray(t *testing.T) {
	// {"subnets": []} should return empty strings.
	raw := map[string]json.RawMessage{
		"id":      json.RawMessage(`"vpc-empty"`),
		"subnets": json.RawMessage(`[]`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "" {
		t.Errorf("expected empty subnetID for empty subnets array, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP for empty subnets array, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_SubnetsAsObjects(t *testing.T) {
	// {"subnets": [{"id": "sub-1", "gateway_ip": "10.0.0.1"}]} should parse the first object.
	raw := map[string]json.RawMessage{
		"id":      json.RawMessage(`"vpc-obj"`),
		"subnets": json.RawMessage(`[{"id": "sub-1", "gateway_ip": "10.0.0.1"}]`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "sub-1" {
		t.Errorf("expected subnetID sub-1, got %s", subnetID)
	}
	if gatewayIP != "10.0.0.1" {
		t.Errorf("expected gatewayIP 10.0.0.1, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_InvalidJSON(t *testing.T) {
	// Invalid JSON in "subnet" value should fall through gracefully.
	raw := map[string]json.RawMessage{
		"id":     json.RawMessage(`"vpc-bad"`),
		"subnet": json.RawMessage(`{invalid json`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	// Should return empty since it can't parse.
	if subnetID != "" {
		t.Errorf("expected empty subnetID for invalid JSON, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP for invalid JSON, got %s", gatewayIP)
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest additional tests
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_PortSecurity(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                 types.StringValue("portsec-vpc"),
		Description:          types.StringNull(),
		AdminStateUp:         types.BoolNull(),
		PortSecurityEnabled:  types.BoolValue(true),
		SubnetName:           types.StringValue("ps-subnet"),
		SubnetCIDR:           types.StringValue("10.5.0.0/24"),
		SubnetIPVersion:      types.Int64Value(4),
		SubnetEnableDHCP:     types.BoolValue(true),
		SubnetGatewayIP:      types.StringNull(),
		SubnetDNSNameservers: types.ListNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.PortSecurityEnabled == nil {
		t.Fatal("expected PortSecurityEnabled to be non-nil")
	}
	if *body.PortSecurityEnabled != true {
		t.Errorf("expected PortSecurityEnabled to be true, got %v", *body.PortSecurityEnabled)
	}

	// Verify JSON serialization includes port_security_enabled
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := raw["port_security_enabled"]; !ok {
		t.Error("expected port_security_enabled to be present in JSON when set to true")
	}
}

func TestBuildCreateRequest_DNS(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                types.StringValue("dns-vpc"),
		Description:         types.StringNull(),
		AdminStateUp:        types.BoolNull(),
		PortSecurityEnabled: types.BoolNull(),
		SubnetName:          types.StringValue("dns-subnet"),
		SubnetCIDR:          types.StringValue("10.6.0.0/24"),
		SubnetIPVersion:     types.Int64Value(4),
		SubnetEnableDHCP:    types.BoolValue(true),
		SubnetGatewayIP:     types.StringNull(),
		SubnetDNSNameservers: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("1.1.1.1"),
			types.StringValue("9.9.9.9"),
			types.StringValue("8.8.8.8"),
		}),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Subnet == nil {
		t.Fatal("expected Subnet to be non-nil")
	}
	if len(body.Subnet.DNSNameservers) != 3 {
		t.Fatalf("expected 3 DNS nameservers, got %d", len(body.Subnet.DNSNameservers))
	}
	if body.Subnet.DNSNameservers[0] != "1.1.1.1" {
		t.Errorf("expected first DNS 1.1.1.1, got %s", body.Subnet.DNSNameservers[0])
	}
	if body.Subnet.DNSNameservers[1] != "9.9.9.9" {
		t.Errorf("expected second DNS 9.9.9.9, got %s", body.Subnet.DNSNameservers[1])
	}
	if body.Subnet.DNSNameservers[2] != "8.8.8.8" {
		t.Errorf("expected third DNS 8.8.8.8, got %s", body.Subnet.DNSNameservers[2])
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState additional tests
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_MTU(t *testing.T) {
	tests := []struct {
		name        string
		mtuInput    int64
		expectedMTU int64
	}{
		{"standard MTU", 1500, 1500},
		{"jumbo frames", 9000, 9000},
		{"zero MTU", 0, 0},
		{"small MTU", 68, 68},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &vpcResourceModel{
				Description: types.StringNull(),
			}
			apiResp := &vpcAPIResponse{
				ID:   "vpc-mtu-test",
				Name: "mtu-vpc",
				MTU:  tc.mtuInput,
			}

			mapAPIResponseToState(model, apiResp)

			if model.MTU.ValueInt64() != tc.expectedMTU {
				t.Errorf("expected MTU %d, got %d", tc.expectedMTU, model.MTU.ValueInt64())
			}
		})
	}
}

// ===========================================================================
// Additional comprehensive tests
// ===========================================================================

// ---------------------------------------------------------------------------
// parseSubnetFromRaw — comprehensive edge cases
// ---------------------------------------------------------------------------

func TestParseSubnetFromRaw_MultipleSubnetsAsObjects(t *testing.T) {
	// Multi-subnet VPC: should return the first subnet.
	raw := map[string]json.RawMessage{
		"id": json.RawMessage(`"vpc-multi"`),
		"subnets": json.RawMessage(`[
			{"id": "sub-first", "gateway_ip": "10.0.0.1"},
			{"id": "sub-second", "gateway_ip": "10.1.0.1"},
			{"id": "sub-third", "gateway_ip": "10.2.0.1"}
		]`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "sub-first" {
		t.Errorf("expected first subnetID sub-first, got %s", subnetID)
	}
	if gatewayIP != "10.0.0.1" {
		t.Errorf("expected first gatewayIP 10.0.0.1, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_MultipleSubnetsAsStrings(t *testing.T) {
	// Read response with multiple subnet ID strings — should return the first.
	raw := map[string]json.RawMessage{
		"id":      json.RawMessage(`"vpc-multi-str"`),
		"subnets": json.RawMessage(`["sub-aaa", "sub-bbb", "sub-ccc"]`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "sub-aaa" {
		t.Errorf("expected first subnetID sub-aaa, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP from string array, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_EmptySubnetObject(t *testing.T) {
	// "subnet" key present but object has empty ID.
	raw := map[string]json.RawMessage{
		"id":     json.RawMessage(`"vpc-empty-sub"`),
		"subnet": json.RawMessage(`{"id":"","gateway_ip":""}`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	// Empty ID should not be returned — falls through.
	if subnetID != "" {
		t.Errorf("expected empty subnetID for empty subnet object, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP for empty subnet object, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_SubnetObjectNoGateway(t *testing.T) {
	// Create response with subnet that has no gateway_ip field.
	raw := map[string]json.RawMessage{
		"id":     json.RawMessage(`"vpc-no-gw"`),
		"subnet": json.RawMessage(`{"id":"sub-no-gw"}`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "sub-no-gw" {
		t.Errorf("expected subnetID sub-no-gw, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP when not in response, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_SubnetOverridesSubnets(t *testing.T) {
	// Both "subnet" and "subnets" present — "subnet" takes priority.
	raw := map[string]json.RawMessage{
		"id":      json.RawMessage(`"vpc-both"`),
		"subnet":  json.RawMessage(`{"id":"sub-from-create","gateway_ip":"10.0.0.1"}`),
		"subnets": json.RawMessage(`["sub-from-read"]`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "sub-from-create" {
		t.Errorf("expected subnet from create response, got %s", subnetID)
	}
	if gatewayIP != "10.0.0.1" {
		t.Errorf("expected gatewayIP 10.0.0.1, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_SubnetsObjectsNoID(t *testing.T) {
	// Subnets array of objects where objects lack "id" — should fall through to string parsing.
	raw := map[string]json.RawMessage{
		"id":      json.RawMessage(`"vpc-no-id"`),
		"subnets": json.RawMessage(`[{"gateway_ip":"10.0.0.1"}]`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	// Object parse succeeds but ID is empty, so it falls through to string parse which fails.
	if subnetID != "" {
		t.Errorf("expected empty subnetID for object without id, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_SubnetNullValue(t *testing.T) {
	// "subnet" key with JSON null value.
	raw := map[string]json.RawMessage{
		"id":     json.RawMessage(`"vpc-null-sub"`),
		"subnet": json.RawMessage(`null`),
	}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "" {
		t.Errorf("expected empty subnetID for null subnet, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP for null subnet, got %s", gatewayIP)
	}
}

func TestParseSubnetFromRaw_EmptyMap(t *testing.T) {
	// Completely empty raw map.
	raw := map[string]json.RawMessage{}

	subnetID, gatewayIP := parseSubnetFromRaw(raw)

	if subnetID != "" {
		t.Errorf("expected empty subnetID for empty map, got %s", subnetID)
	}
	if gatewayIP != "" {
		t.Errorf("expected empty gatewayIP for empty map, got %s", gatewayIP)
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState — comprehensive field mapping
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	model := &vpcResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &vpcAPIResponse{
		ID:                  "vpc-full-123",
		Name:                "full-vpc",
		Description:         "Full description",
		AdminStateUp:        false,
		PortSecurityEnabled: false,
		Status:              "BUILD",
		MTU:                 9000,
		CreatedAt:           "2025-06-15T10:30:00Z",
		UpdatedAt:           "2025-06-15T11:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "vpc-full-123" {
		t.Errorf("expected ID vpc-full-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-vpc" {
		t.Errorf("expected Name full-vpc, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Full description" {
		t.Errorf("expected Description 'Full description', got %s", model.Description.ValueString())
	}
	if model.AdminStateUp.ValueBool() != false {
		t.Error("expected AdminStateUp false")
	}
	if model.PortSecurityEnabled.ValueBool() != false {
		t.Error("expected PortSecurityEnabled false")
	}
	if model.Status.ValueString() != "BUILD" {
		t.Errorf("expected Status BUILD, got %s", model.Status.ValueString())
	}
	if model.MTU.ValueInt64() != 9000 {
		t.Errorf("expected MTU 9000, got %d", model.MTU.ValueInt64())
	}
	if model.CreatedAt.ValueString() != "2025-06-15T10:30:00Z" {
		t.Errorf("expected CreatedAt 2025-06-15T10:30:00Z, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2025-06-15T11:00:00Z" {
		t.Errorf("expected UpdatedAt 2025-06-15T11:00:00Z, got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionPreserveNull(t *testing.T) {
	// When model.Description is null and API returns empty string, keep null.
	model := &vpcResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &vpcAPIResponse{
		ID:          "vpc-desc-null",
		Name:        "desc-vpc",
		Description: "",
		Status:      "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Errorf("expected Description to remain null when API returns empty, got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionOverwriteExisting(t *testing.T) {
	// When model has a non-null description and API returns a new description, overwrite.
	model := &vpcResourceModel{
		Description: types.StringValue("old description"),
	}

	apiResp := &vpcAPIResponse{
		ID:          "vpc-desc-update",
		Name:        "desc-vpc",
		Description: "new description",
		Status:      "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	if model.Description.ValueString() != "new description" {
		t.Errorf("expected Description 'new description', got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionKeepExistingWhenAPIEmpty(t *testing.T) {
	// When model has a non-null description and API returns empty, keep the existing value.
	model := &vpcResourceModel{
		Description: types.StringValue("existing description"),
	}

	apiResp := &vpcAPIResponse{
		ID:          "vpc-desc-keep",
		Name:        "desc-vpc",
		Description: "",
		Status:      "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	// The logic: if apiResp.Description == "" and model.Description is NOT null,
	// neither branch runs, so the existing value is kept.
	if model.Description.ValueString() != "existing description" {
		t.Errorf("expected Description to keep 'existing description', got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_TimestampEmpty(t *testing.T) {
	model := &vpcResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &vpcAPIResponse{
		ID:        "vpc-ts-empty",
		Name:      "ts-vpc",
		Status:    "ACTIVE",
		CreatedAt: "",
		UpdatedAt: "",
	}

	mapAPIResponseToState(model, apiResp)

	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "" {
		t.Errorf("expected empty UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_StatusVariants(t *testing.T) {
	statuses := []string{"ACTIVE", "BUILD", "DOWN", "ERROR", "PENDING_CREATE"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			model := &vpcResourceModel{
				Description: types.StringNull(),
			}
			apiResp := &vpcAPIResponse{
				ID:     "vpc-status",
				Name:   "status-vpc",
				Status: status,
			}

			mapAPIResponseToState(model, apiResp)

			if model.Status.ValueString() != status {
				t.Errorf("expected Status %s, got %s", status, model.Status.ValueString())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest — additional edge cases
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_IPv6Subnet(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                 types.StringValue("ipv6-vpc"),
		Description:          types.StringNull(),
		AdminStateUp:         types.BoolNull(),
		PortSecurityEnabled:  types.BoolNull(),
		SubnetName:           types.StringValue("ipv6-subnet"),
		SubnetCIDR:           types.StringValue("fd00::/64"),
		SubnetIPVersion:      types.Int64Value(6),
		SubnetEnableDHCP:     types.BoolValue(false),
		SubnetGatewayIP:      types.StringNull(),
		SubnetDNSNameservers: types.ListNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Subnet == nil {
		t.Fatal("expected Subnet to be non-nil")
	}
	if body.Subnet.IPVersion != 6 {
		t.Errorf("expected subnet IP version 6, got %d", body.Subnet.IPVersion)
	}
	if body.Subnet.EnableDHCP != false {
		t.Error("expected subnet EnableDHCP to be false")
	}
	if body.Subnet.CIDR != "fd00::/64" {
		t.Errorf("expected subnet CIDR fd00::/64, got %s", body.Subnet.CIDR)
	}
}

func TestBuildCreateRequest_AdminStateUpFalse(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                 types.StringValue("admin-down-vpc"),
		Description:          types.StringNull(),
		AdminStateUp:         types.BoolValue(false),
		PortSecurityEnabled:  types.BoolNull(),
		SubnetName:           types.StringValue("sub"),
		SubnetCIDR:           types.StringValue("10.0.0.0/24"),
		SubnetIPVersion:      types.Int64Value(4),
		SubnetEnableDHCP:     types.BoolValue(true),
		SubnetGatewayIP:      types.StringNull(),
		SubnetDNSNameservers: types.ListNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.AdminStateUp == nil {
		t.Fatal("expected AdminStateUp to be non-nil")
	}
	if *body.AdminStateUp != false {
		t.Error("expected AdminStateUp to be false")
	}

	// Verify JSON serialization — admin_state_up:false should be present
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := raw["admin_state_up"]; !ok {
		t.Error("expected admin_state_up to be present in JSON when explicitly set to false")
	}
}

// ---------------------------------------------------------------------------
// buildUpdateRequest — comprehensive tests
// ---------------------------------------------------------------------------

func TestBuildUpdateRequest_JSONFieldNames(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                types.StringValue("update-test"),
		Description:         types.StringValue("Updated"),
		AdminStateUp:        types.BoolValue(false),
		PortSecurityEnabled: types.BoolValue(true),
	}

	body := buildUpdateRequest(plan)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify expected JSON keys are present.
	expectedKeys := []string{"name", "description", "admin_state_up", "port_security_enabled"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q to be present in update request", key)
		}
	}

	// Verify no subnet-related keys leak into the update request.
	forbiddenKeys := []string{"subnet", "subnets", "subnet_name", "subnet_cidr",
		"subnet_ip_version", "subnet_enable_dhcp", "subnet_gateway_ip", "subnet_dns_nameservers"}
	for _, key := range forbiddenKeys {
		if _, ok := raw[key]; ok {
			t.Errorf("unexpected key %q found in update request JSON", key)
		}
	}
}

func TestBuildUpdateRequest_OnlyNameWhenOptionalsNull(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                types.StringValue("only-name-vpc"),
		Description:         types.StringNull(),
		AdminStateUp:        types.BoolNull(),
		PortSecurityEnabled: types.BoolNull(),
	}

	body := buildUpdateRequest(plan)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Name must always be present.
	if _, ok := raw["name"]; !ok {
		t.Error("expected JSON key 'name' to always be present")
	}
	if raw["name"] != "only-name-vpc" {
		t.Errorf("expected name only-name-vpc, got %v", raw["name"])
	}

	// Optional fields should be omitted.
	if _, ok := raw["description"]; ok {
		t.Error("expected 'description' to be omitted when null")
	}
	if _, ok := raw["admin_state_up"]; ok {
		t.Error("expected 'admin_state_up' to be omitted when null")
	}
	if _, ok := raw["port_security_enabled"]; ok {
		t.Error("expected 'port_security_enabled' to be omitted when null")
	}
}

func TestBuildUpdateRequest_PortSecurityOnly(t *testing.T) {
	plan := &vpcResourceModel{
		Name:                types.StringValue("ps-only-vpc"),
		Description:         types.StringNull(),
		AdminStateUp:        types.BoolNull(),
		PortSecurityEnabled: types.BoolValue(true),
	}

	body := buildUpdateRequest(plan)

	if body.PortSecurityEnabled == nil {
		t.Fatal("expected PortSecurityEnabled to be non-nil")
	}
	if *body.PortSecurityEnabled != true {
		t.Error("expected PortSecurityEnabled to be true")
	}
	if body.AdminStateUp != nil {
		t.Error("expected AdminStateUp to be nil when null in plan")
	}
}

// ---------------------------------------------------------------------------
// API types JSON serialization
// ---------------------------------------------------------------------------

func TestVPCCreateRequest_JSONRoundTrip(t *testing.T) {
	adminUp := true
	portSec := false
	req := vpcCreateRequest{
		Name:                "roundtrip-vpc",
		Description:         "test round trip",
		AdminStateUp:        &adminUp,
		PortSecurityEnabled: &portSec,
		Subnet: &vpcCreateSubnet{
			Name:           "rt-subnet",
			CIDR:           "192.168.0.0/24",
			IPVersion:      4,
			EnableDHCP:     true,
			DNSNameservers: []string{"8.8.8.8"},
			GatewayIP:      "192.168.0.1",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded vpcCreateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != "roundtrip-vpc" {
		t.Errorf("expected name roundtrip-vpc, got %s", decoded.Name)
	}
	if decoded.Subnet == nil {
		t.Fatal("expected non-nil subnet after round trip")
	}
	if decoded.Subnet.CIDR != "192.168.0.0/24" {
		t.Errorf("expected CIDR 192.168.0.0/24, got %s", decoded.Subnet.CIDR)
	}
	if len(decoded.Subnet.DNSNameservers) != 1 {
		t.Fatalf("expected 1 DNS, got %d", len(decoded.Subnet.DNSNameservers))
	}
}

func TestVPCUpdateRequest_JSONRoundTrip(t *testing.T) {
	adminUp := true
	req := vpcUpdateRequest{
		Name:         "update-rt-vpc",
		Description:  "update round trip",
		AdminStateUp: &adminUp,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded vpcUpdateRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Name != "update-rt-vpc" {
		t.Errorf("expected name update-rt-vpc, got %s", decoded.Name)
	}
	if decoded.AdminStateUp == nil || *decoded.AdminStateUp != true {
		t.Error("expected AdminStateUp true after round trip")
	}
}

func TestVPCAPIResponse_Unmarshal(t *testing.T) {
	raw := `{
		"id": "vpc-unmarshal",
		"name": "unmarshal-vpc",
		"description": "Testing unmarshal",
		"admin_state_up": true,
		"port_security_enabled": false,
		"status": "ACTIVE",
		"mtu": 1450,
		"created_at": "2025-01-01T00:00:00Z",
		"updated_at": "2025-01-02T12:00:00Z"
	}`

	var resp vpcAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "vpc-unmarshal" {
		t.Errorf("expected ID vpc-unmarshal, got %s", resp.ID)
	}
	if resp.AdminStateUp != true {
		t.Error("expected AdminStateUp true")
	}
	if resp.PortSecurityEnabled != false {
		t.Error("expected PortSecurityEnabled false")
	}
	if resp.MTU != 1450 {
		t.Errorf("expected MTU 1450, got %d", resp.MTU)
	}
}
