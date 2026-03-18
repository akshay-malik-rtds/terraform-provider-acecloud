package floating_ip_association

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestAssociateRequest(t *testing.T) {
	req := associateRequest{
		FloatingIPAddress: "203.0.113.10",
		InstanceID:        "instance-123",
		FixedIPAddress:    "10.0.0.5",
	}
	if req.FloatingIPAddress != "203.0.113.10" {
		t.Errorf("expected floating_ip_address 203.0.113.10, got %s", req.FloatingIPAddress)
	}
	if req.InstanceID != "instance-123" {
		t.Errorf("expected instance_id instance-123, got %s", req.InstanceID)
	}
	if req.FixedIPAddress != "10.0.0.5" {
		t.Errorf("expected fixed_ip_address 10.0.0.5, got %s", req.FixedIPAddress)
	}
}

func TestDisassociateRequest(t *testing.T) {
	req := disassociateRequest{
		FloatingIPAddress: "203.0.113.10",
		InstanceID:        "instance-123",
	}
	if req.FloatingIPAddress != "203.0.113.10" {
		t.Errorf("expected floating_ip_address 203.0.113.10, got %s", req.FloatingIPAddress)
	}
	if req.InstanceID != "instance-123" {
		t.Errorf("expected instance_id instance-123, got %s", req.InstanceID)
	}
}

func TestAssociateRequest_JSON(t *testing.T) {
	req := associateRequest{
		FloatingIPAddress: "203.0.113.10",
		InstanceID:        "inst-456",
		FixedIPAddress:    "10.0.0.5",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal associate request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal associate request: %v", err)
	}

	if parsed["floating_ip_address"] != "203.0.113.10" {
		t.Errorf("expected floating_ip_address 203.0.113.10, got %v", parsed["floating_ip_address"])
	}
	if parsed["instance_id"] != "inst-456" {
		t.Errorf("expected instance_id inst-456, got %v", parsed["instance_id"])
	}
	if parsed["fixed_ip_address"] != "10.0.0.5" {
		t.Errorf("expected fixed_ip_address 10.0.0.5, got %v", parsed["fixed_ip_address"])
	}
}

func TestAssociateRequest_WithoutFixedIP(t *testing.T) {
	// FixedIPAddress has omitempty, so it should be absent when empty.
	req := associateRequest{
		FloatingIPAddress: "203.0.113.20",
		InstanceID:        "inst-789",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal associate request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal associate request: %v", err)
	}

	if parsed["floating_ip_address"] != "203.0.113.20" {
		t.Errorf("expected floating_ip_address 203.0.113.20, got %v", parsed["floating_ip_address"])
	}
	if parsed["instance_id"] != "inst-789" {
		t.Errorf("expected instance_id inst-789, got %v", parsed["instance_id"])
	}
	if _, exists := parsed["fixed_ip_address"]; exists {
		t.Error("expected fixed_ip_address to be absent due to omitempty")
	}
}

func TestDisassociateRequest_JSON(t *testing.T) {
	req := disassociateRequest{
		FloatingIPAddress: "203.0.113.30",
		InstanceID:        "inst-disassoc",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal disassociate request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal disassociate request: %v", err)
	}

	if parsed["floating_ip_address"] != "203.0.113.30" {
		t.Errorf("expected floating_ip_address 203.0.113.30, got %v", parsed["floating_ip_address"])
	}
	if parsed["instance_id"] != "inst-disassoc" {
		t.Errorf("expected instance_id inst-disassoc, got %v", parsed["instance_id"])
	}

	// Verify exactly 2 keys (no extra fields).
	if len(parsed) != 2 {
		t.Errorf("expected 2 keys in JSON, got %d", len(parsed))
	}
}

func TestCompositeID(t *testing.T) {
	floatingIP := "203.0.113.10"
	instanceID := "inst-123"

	compositeID := fmt.Sprintf("%s/%s", floatingIP, instanceID)

	expected := "203.0.113.10/inst-123"
	if compositeID != expected {
		t.Errorf("expected composite ID %s, got %s", expected, compositeID)
	}
}

func TestAssociateRequest_AllFields(t *testing.T) {
	req := associateRequest{
		FloatingIPAddress: "198.51.100.5",
		InstanceID:        "inst-all-fields",
		FixedIPAddress:    "10.0.0.99",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal associate request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal associate request: %v", err)
	}

	if parsed["floating_ip_address"] != "198.51.100.5" {
		t.Errorf("expected floating_ip_address 198.51.100.5, got %v", parsed["floating_ip_address"])
	}
	if parsed["instance_id"] != "inst-all-fields" {
		t.Errorf("expected instance_id inst-all-fields, got %v", parsed["instance_id"])
	}
	if parsed["fixed_ip_address"] != "10.0.0.99" {
		t.Errorf("expected fixed_ip_address 10.0.0.99, got %v", parsed["fixed_ip_address"])
	}

	// Verify exactly 3 keys when all fields are set.
	if len(parsed) != 3 {
		t.Errorf("expected 3 keys in JSON, got %d", len(parsed))
	}
}

// ---------------------------------------------------------------------------
// Composite ID format tests
// ---------------------------------------------------------------------------

func TestCompositeID_Format(t *testing.T) {
	tests := []struct {
		name       string
		floatingIP string
		instanceID string
		expected   string
	}{
		{"standard IPs", "203.0.113.50", "inst-abc-123", "203.0.113.50/inst-abc-123"},
		{"IPv4 with UUID", "10.0.0.1", "550e8400-e29b-41d4-a716-446655440000", "10.0.0.1/550e8400-e29b-41d4-a716-446655440000"},
		{"public IP", "198.51.100.1", "i-001", "198.51.100.1/i-001"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			compositeID := fmt.Sprintf("%s/%s", tc.floatingIP, tc.instanceID)
			if compositeID != tc.expected {
				t.Errorf("expected composite ID %s, got %s", tc.expected, compositeID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Associate request with fixed IP
// ---------------------------------------------------------------------------

func TestAssociateRequest_WithFixedIP(t *testing.T) {
	req := associateRequest{
		FloatingIPAddress: "203.0.113.77",
		InstanceID:        "inst-fixed",
		FixedIPAddress:    "10.0.0.42",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal associate request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal associate request: %v", err)
	}

	// Verify fixed_ip_address is present and correct.
	if parsed["fixed_ip_address"] != "10.0.0.42" {
		t.Errorf("expected fixed_ip_address 10.0.0.42, got %v", parsed["fixed_ip_address"])
	}
	// Verify all three fields are present.
	if len(parsed) != 3 {
		t.Errorf("expected 3 keys in JSON when fixed_ip_address is set, got %d", len(parsed))
	}
}

// ---------------------------------------------------------------------------
// Disassociate request: verify no port/fixed_ip fields
// ---------------------------------------------------------------------------

func TestDisassociateRequest_EmptyPortID(t *testing.T) {
	// Disassociate only needs floating_ip_address and instance_id.
	// There is no port_id field in disassociateRequest.
	req := disassociateRequest{
		FloatingIPAddress: "203.0.113.88",
		InstanceID:        "inst-disassoc-test",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal disassociate request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal disassociate request: %v", err)
	}

	// Should have exactly 2 keys: floating_ip_address and instance_id.
	if len(parsed) != 2 {
		t.Errorf("expected 2 keys in disassociate JSON, got %d", len(parsed))
	}
	if parsed["floating_ip_address"] != "203.0.113.88" {
		t.Errorf("expected floating_ip_address 203.0.113.88, got %v", parsed["floating_ip_address"])
	}
	if parsed["instance_id"] != "inst-disassoc-test" {
		t.Errorf("expected instance_id inst-disassoc-test, got %v", parsed["instance_id"])
	}
	// Verify no port_id or fixed_ip_address keys.
	if _, exists := parsed["port_id"]; exists {
		t.Error("expected port_id to NOT be present in disassociate request")
	}
	if _, exists := parsed["fixed_ip_address"]; exists {
		t.Error("expected fixed_ip_address to NOT be present in disassociate request")
	}
}

// ---------------------------------------------------------------------------
// Table-driven: Associate request body construction
// ---------------------------------------------------------------------------

func TestAssociateRequestBody_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		floatingIP      string
		instanceID      string
		fixedIP         string
		expectFixedIP   bool
		expectedKeyCount int
	}{
		{
			name:            "all fields populated",
			floatingIP:      "203.0.113.100",
			instanceID:      "inst-full-001",
			fixedIP:         "10.0.0.50",
			expectFixedIP:   true,
			expectedKeyCount: 3,
		},
		{
			name:            "without fixed IP",
			floatingIP:      "203.0.113.101",
			instanceID:      "inst-no-fixed-002",
			fixedIP:         "",
			expectFixedIP:   false,
			expectedKeyCount: 2,
		},
		{
			name:            "UUID instance ID",
			floatingIP:      "198.51.100.10",
			instanceID:      "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
			fixedIP:         "172.16.0.1",
			expectFixedIP:   true,
			expectedKeyCount: 3,
		},
		{
			name:            "private range floating IP",
			floatingIP:      "10.255.255.1",
			instanceID:      "inst-priv-003",
			fixedIP:         "",
			expectFixedIP:   false,
			expectedKeyCount: 2,
		},
		{
			name:            "loopback-like fixed IP",
			floatingIP:      "203.0.113.200",
			instanceID:      "inst-loop-004",
			fixedIP:         "127.0.0.1",
			expectFixedIP:   true,
			expectedKeyCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := associateRequest{
				FloatingIPAddress: tc.floatingIP,
				InstanceID:        tc.instanceID,
				FixedIPAddress:    tc.fixedIP,
			}

			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if parsed["floating_ip_address"] != tc.floatingIP {
				t.Errorf("expected floating_ip_address %s, got %v", tc.floatingIP, parsed["floating_ip_address"])
			}
			if parsed["instance_id"] != tc.instanceID {
				t.Errorf("expected instance_id %s, got %v", tc.instanceID, parsed["instance_id"])
			}

			_, hasFixedIP := parsed["fixed_ip_address"]
			if tc.expectFixedIP && !hasFixedIP {
				t.Error("expected fixed_ip_address to be present")
			}
			if !tc.expectFixedIP && hasFixedIP {
				t.Error("expected fixed_ip_address to be absent (omitempty)")
			}

			if len(parsed) != tc.expectedKeyCount {
				t.Errorf("expected %d keys, got %d", tc.expectedKeyCount, len(parsed))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Table-driven: Disassociate request body construction
// ---------------------------------------------------------------------------

func TestDisassociateRequestBody_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		floatingIP string
		instanceID string
	}{
		{"standard detach", "203.0.113.50", "inst-detach-001"},
		{"UUID instance", "198.51.100.20", "550e8400-e29b-41d4-a716-446655440000"},
		{"short instance ID", "10.0.0.1", "i-1"},
		{"long instance ID", "172.16.0.1", "instance-very-long-identifier-name-12345678"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := disassociateRequest{
				FloatingIPAddress: tc.floatingIP,
				InstanceID:        tc.instanceID,
			}

			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if parsed["floating_ip_address"] != tc.floatingIP {
				t.Errorf("expected floating_ip_address %s, got %v", tc.floatingIP, parsed["floating_ip_address"])
			}
			if parsed["instance_id"] != tc.instanceID {
				t.Errorf("expected instance_id %s, got %v", tc.instanceID, parsed["instance_id"])
			}
			// Disassociate must never have fixed_ip_address or port_id
			if _, exists := parsed["fixed_ip_address"]; exists {
				t.Error("disassociate must not have fixed_ip_address")
			}
			if _, exists := parsed["port_id"]; exists {
				t.Error("disassociate must not have port_id")
			}
			if len(parsed) != 2 {
				t.Errorf("expected exactly 2 keys, got %d", len(parsed))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Composite ID: parsing and roundtrip
// ---------------------------------------------------------------------------

func TestCompositeID_Parsing(t *testing.T) {
	tests := []struct {
		name               string
		floatingIP         string
		instanceID         string
		expectedComposite  string
		canSplit           bool
	}{
		{
			name:              "simple IP and ID",
			floatingIP:        "203.0.113.1",
			instanceID:        "inst-001",
			expectedComposite: "203.0.113.1/inst-001",
			canSplit:          true,
		},
		{
			name:              "UUID instance",
			floatingIP:        "10.0.0.5",
			instanceID:        "abc12345-def6-7890-abcd-ef1234567890",
			expectedComposite: "10.0.0.5/abc12345-def6-7890-abcd-ef1234567890",
			canSplit:          true,
		},
		{
			name:              "public IP with short ID",
			floatingIP:        "198.51.100.99",
			instanceID:        "i-1",
			expectedComposite: "198.51.100.99/i-1",
			canSplit:          true,
		},
		{
			name:              "class A private IP",
			floatingIP:        "10.255.255.254",
			instanceID:        "inst-classA",
			expectedComposite: "10.255.255.254/inst-classA",
			canSplit:          true,
		},
		{
			name:              "class C private IP",
			floatingIP:        "192.168.1.100",
			instanceID:        "inst-classC",
			expectedComposite: "192.168.1.100/inst-classC",
			canSplit:          true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			composite := fmt.Sprintf("%s/%s", tc.floatingIP, tc.instanceID)
			if composite != tc.expectedComposite {
				t.Errorf("expected %s, got %s", tc.expectedComposite, composite)
			}

			// Verify we can split back into components
			if tc.canSplit {
				parts := splitCompositeID(composite)
				if len(parts) != 2 {
					t.Fatalf("expected 2 parts from split, got %d", len(parts))
				}
				if parts[0] != tc.floatingIP {
					t.Errorf("expected first part %s, got %s", tc.floatingIP, parts[0])
				}
				if parts[1] != tc.instanceID {
					t.Errorf("expected second part %s, got %s", tc.instanceID, parts[1])
				}
			}
		})
	}
}

// splitCompositeID is a test helper to split "ip/instance" into parts.
func splitCompositeID(id string) []string {
	for i, c := range id {
		if c == '/' {
			return []string{id[:i], id[i+1:]}
		}
	}
	return []string{id}
}

// ---------------------------------------------------------------------------
// JSON round-trip: associate request
// ---------------------------------------------------------------------------

func TestAssociateRequest_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		req  associateRequest
	}{
		{
			name: "full request",
			req: associateRequest{
				FloatingIPAddress: "203.0.113.10",
				InstanceID:        "inst-rt-001",
				FixedIPAddress:    "10.0.0.5",
			},
		},
		{
			name: "no fixed IP",
			req: associateRequest{
				FloatingIPAddress: "198.51.100.1",
				InstanceID:        "inst-rt-002",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var decoded associateRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if decoded.FloatingIPAddress != tc.req.FloatingIPAddress {
				t.Errorf("FloatingIPAddress mismatch: got %s, want %s", decoded.FloatingIPAddress, tc.req.FloatingIPAddress)
			}
			if decoded.InstanceID != tc.req.InstanceID {
				t.Errorf("InstanceID mismatch: got %s, want %s", decoded.InstanceID, tc.req.InstanceID)
			}
			if decoded.FixedIPAddress != tc.req.FixedIPAddress {
				t.Errorf("FixedIPAddress mismatch: got %s, want %s", decoded.FixedIPAddress, tc.req.FixedIPAddress)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// JSON round-trip: disassociate request
// ---------------------------------------------------------------------------

func TestDisassociateRequest_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		req  disassociateRequest
	}{
		{
			name: "standard detach",
			req: disassociateRequest{
				FloatingIPAddress: "203.0.113.55",
				InstanceID:        "inst-drt-001",
			},
		},
		{
			name: "UUID instance detach",
			req: disassociateRequest{
				FloatingIPAddress: "10.0.0.99",
				InstanceID:        "550e8400-e29b-41d4-a716-446655440000",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var decoded disassociateRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			if decoded.FloatingIPAddress != tc.req.FloatingIPAddress {
				t.Errorf("FloatingIPAddress mismatch: got %s, want %s", decoded.FloatingIPAddress, tc.req.FloatingIPAddress)
			}
			if decoded.InstanceID != tc.req.InstanceID {
				t.Errorf("InstanceID mismatch: got %s, want %s", decoded.InstanceID, tc.req.InstanceID)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Exact JSON key names (verify snake_case JSON tags)
// ---------------------------------------------------------------------------

func TestAssociateRequest_ExactJSONKeys(t *testing.T) {
	req := associateRequest{
		FloatingIPAddress: "1.2.3.4",
		InstanceID:        "i-1",
		FixedIPAddress:    "5.6.7.8",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	expectedKeys := []string{"floating_ip_address", "instance_id", "fixed_ip_address"}
	for _, key := range expectedKeys {
		if _, exists := raw[key]; !exists {
			t.Errorf("expected JSON key %q to exist", key)
		}
	}

	// Ensure no camelCase or other keys
	unexpectedKeys := []string{"floatingIpAddress", "FloatingIPAddress", "instanceId", "InstanceID", "fixedIpAddress", "FixedIPAddress"}
	for _, key := range unexpectedKeys {
		if _, exists := raw[key]; exists {
			t.Errorf("unexpected JSON key %q found (should be snake_case)", key)
		}
	}
}

func TestDisassociateRequest_ExactJSONKeys(t *testing.T) {
	req := disassociateRequest{
		FloatingIPAddress: "1.2.3.4",
		InstanceID:        "i-1",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	expectedKeys := []string{"floating_ip_address", "instance_id"}
	for _, key := range expectedKeys {
		if _, exists := raw[key]; !exists {
			t.Errorf("expected JSON key %q to exist", key)
		}
	}

	if len(raw) != 2 {
		t.Errorf("expected exactly 2 keys, got %d", len(raw))
	}
}

// ---------------------------------------------------------------------------
// Schema attribute verification
// ---------------------------------------------------------------------------

func TestSchema_RequiredAttributes(t *testing.T) {
	s := floatingIPAssociationSchema()

	requiredFields := []string{"floating_ip_address", "instance_id"}
	for _, field := range requiredFields {
		attr, ok := s.Attributes[field]
		if !ok {
			t.Errorf("expected schema to contain attribute %q", field)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("expected attribute %q to be required", field)
		}
	}
}

func TestSchema_OptionalAttributes(t *testing.T) {
	s := floatingIPAssociationSchema()

	attr, ok := s.Attributes["fixed_ip_address"]
	if !ok {
		t.Fatal("expected schema to contain attribute 'fixed_ip_address'")
	}
	if !attr.IsOptional() {
		t.Error("expected fixed_ip_address to be optional")
	}
}

func TestSchema_ComputedAttributes(t *testing.T) {
	s := floatingIPAssociationSchema()

	attr, ok := s.Attributes["id"]
	if !ok {
		t.Fatal("expected schema to contain attribute 'id'")
	}
	if !attr.IsComputed() {
		t.Error("expected id to be computed")
	}
}

func TestSchema_AllAttributesExist(t *testing.T) {
	s := floatingIPAssociationSchema()

	expectedAttrs := []string{"id", "floating_ip_address", "instance_id", "fixed_ip_address"}
	for _, attrName := range expectedAttrs {
		if _, ok := s.Attributes[attrName]; !ok {
			t.Errorf("expected schema to contain attribute %q", attrName)
		}
	}

	if len(s.Attributes) != len(expectedAttrs) {
		t.Errorf("expected %d attributes, got %d", len(expectedAttrs), len(s.Attributes))
	}
}

func TestSchema_Description(t *testing.T) {
	s := floatingIPAssociationSchema()

	if s.Description == "" {
		t.Error("expected schema to have a non-empty description")
	}
}

// ---------------------------------------------------------------------------
// Model struct field mapping
// ---------------------------------------------------------------------------

func TestModel_FieldTags(t *testing.T) {
	// Verify the model can be used with the schema (tfsdk tags match attribute names)
	model := floatingIPAssociationModel{}
	// Just verify the struct can be instantiated — the tfsdk tags are
	// validated by the framework at runtime.
	_ = model.ID
	_ = model.FloatingIPAddress
	_ = model.InstanceID
	_ = model.FixedIPAddress
}

// ---------------------------------------------------------------------------
// API path constant
// ---------------------------------------------------------------------------

func TestAPIPath(t *testing.T) {
	if apiPath != "/cloud/floating-ips/action" {
		t.Errorf("expected apiPath /cloud/floating-ips/action, got %s", apiPath)
	}
}

// ---------------------------------------------------------------------------
// Associate request: edge cases for fixed IP omitempty
// ---------------------------------------------------------------------------

func TestAssociateRequest_FixedIPOmitempty(t *testing.T) {
	tests := []struct {
		name         string
		fixedIP      string
		shouldOmit   bool
	}{
		{"empty string omitted", "", true},
		{"non-empty included", "10.0.0.1", false},
		{"single char included", "x", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := associateRequest{
				FloatingIPAddress: "1.2.3.4",
				InstanceID:        "i-1",
				FixedIPAddress:    tc.fixedIP,
			}

			data, err := json.Marshal(req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}

			_, hasFixedIP := parsed["fixed_ip_address"]
			if tc.shouldOmit && hasFixedIP {
				t.Error("expected fixed_ip_address to be omitted")
			}
			if !tc.shouldOmit && !hasFixedIP {
				t.Error("expected fixed_ip_address to be present")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Associate/Disassociate: exact JSON output
// ---------------------------------------------------------------------------

func TestAssociateRequest_ExactJSON(t *testing.T) {
	tests := []struct {
		name     string
		req      associateRequest
		expected string
	}{
		{
			name: "with fixed IP",
			req: associateRequest{
				FloatingIPAddress: "1.1.1.1",
				InstanceID:        "i-1",
				FixedIPAddress:    "2.2.2.2",
			},
			expected: `{"floating_ip_address":"1.1.1.1","instance_id":"i-1","fixed_ip_address":"2.2.2.2"}`,
		},
		{
			name: "without fixed IP",
			req: associateRequest{
				FloatingIPAddress: "3.3.3.3",
				InstanceID:        "i-2",
			},
			expected: `{"floating_ip_address":"3.3.3.3","instance_id":"i-2"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("expected JSON:\n  %s\ngot:\n  %s", tc.expected, string(data))
			}
		})
	}
}

func TestDisassociateRequest_ExactJSON(t *testing.T) {
	tests := []struct {
		name     string
		req      disassociateRequest
		expected string
	}{
		{
			name: "standard",
			req: disassociateRequest{
				FloatingIPAddress: "4.4.4.4",
				InstanceID:        "i-3",
			},
			expected: `{"floating_ip_address":"4.4.4.4","instance_id":"i-3"}`,
		},
		{
			name: "UUID instance",
			req: disassociateRequest{
				FloatingIPAddress: "5.5.5.5",
				InstanceID:        "aabbccdd-1122-3344-5566-778899aabbcc",
			},
			expected: `{"floating_ip_address":"5.5.5.5","instance_id":"aabbccdd-1122-3344-5566-778899aabbcc"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.req)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			if string(data) != tc.expected {
				t.Errorf("expected JSON:\n  %s\ngot:\n  %s", tc.expected, string(data))
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Composite ID edge cases
// ---------------------------------------------------------------------------

func TestCompositeID_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		floatingIP string
		instanceID string
	}{
		{"minimum length IP", "0.0.0.0", "i"},
		{"max octet values", "255.255.255.255", "inst-max"},
		{"numeric instance ID", "10.0.0.1", "12345"},
		{"hyphenated instance ID", "10.0.0.2", "my-test-instance-id"},
		{"long UUID", "10.0.0.3", "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			composite := fmt.Sprintf("%s/%s", tc.floatingIP, tc.instanceID)

			// Must contain exactly one slash
			slashCount := 0
			for _, c := range composite {
				if c == '/' {
					slashCount++
				}
			}
			if slashCount != 1 {
				t.Errorf("expected exactly 1 slash in composite ID, got %d in %q", slashCount, composite)
			}

			// Must start with IP
			parts := splitCompositeID(composite)
			if parts[0] != tc.floatingIP {
				t.Errorf("first part should be floating IP %s, got %s", tc.floatingIP, parts[0])
			}
			if parts[1] != tc.instanceID {
				t.Errorf("second part should be instance ID %s, got %s", tc.instanceID, parts[1])
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func TestMetadata(t *testing.T) {
	r := &floatingIPAssociationResource{}
	req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	expected := "acecloud_floating_ip_association"
	if resp.TypeName != expected {
		t.Errorf("expected type name %q, got %q", expected, resp.TypeName)
	}
}

func TestMetadata_DifferentProvider(t *testing.T) {
	r := &floatingIPAssociationResource{}
	req := resource.MetadataRequest{ProviderTypeName: "testprovider"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)

	expected := "testprovider_floating_ip_association"
	if resp.TypeName != expected {
		t.Errorf("expected type name %q, got %q", expected, resp.TypeName)
	}
}
