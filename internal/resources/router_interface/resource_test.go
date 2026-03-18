package router_interface

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestMapInterfaceToState(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:         "port-abc-123",
		SubnetID:   "subnet-xyz",
		IPAddress:  "10.0.0.1",
		Status:     "ACTIVE",
		MACAddress: "fa:16:3e:aa:bb:cc",
	}

	mapInterfaceToState(model, iface)

	if model.ID.ValueString() != "port-abc-123" {
		t.Errorf("expected ID port-abc-123, got %s", model.ID.ValueString())
	}
	if model.IPAddress.ValueString() != "10.0.0.1" {
		t.Errorf("expected IPAddress 10.0.0.1, got %s", model.IPAddress.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
	if model.MACAddress.ValueString() != "fa:16:3e:aa:bb:cc" {
		t.Errorf("expected MACAddress fa:16:3e:aa:bb:cc, got %s", model.MACAddress.ValueString())
	}
}

func TestMapInterfaceToState_FixedIPs(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:     "port-456",
		Status: "ACTIVE",
		FixedIPs: []fixedIP{
			{SubnetID: "subnet-1", IPAddress: "192.168.1.1"},
		},
	}

	mapInterfaceToState(model, iface)

	if model.IPAddress.ValueString() != "192.168.1.1" {
		t.Errorf("expected IPAddress from fixed_ips 192.168.1.1, got %s", model.IPAddress.ValueString())
	}
}

func TestInterfaceDeleteRequest(t *testing.T) {
	req := interfaceDeleteRequest{
		Key:    "id",
		Values: []string{"port-1"},
	}
	if req.Key != "id" {
		t.Errorf("expected key 'id', got %s", req.Key)
	}
	if len(req.Values) != 1 {
		t.Errorf("expected 1 value, got %d", len(req.Values))
	}
}

func TestInterfaceCreateRequest_JSON(t *testing.T) {
	req := interfaceCreateRequest{Subnet: "subnet-1"}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create request: %v", err)
	}

	if parsed["subnet"] != "subnet-1" {
		t.Errorf("expected subnet 'subnet-1', got %v", parsed["subnet"])
	}

	// Verify only one key exists in the JSON.
	if len(parsed) != 1 {
		t.Errorf("expected 1 key in JSON, got %d", len(parsed))
	}
}

func TestMapInterfaceToState_EmptyFields(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:         "port-empty",
		SubnetID:   "subnet-1",
		IPAddress:  "10.0.0.5",
		Status:     "",
		MACAddress: "",
	}

	mapInterfaceToState(model, iface)

	if model.ID.ValueString() != "port-empty" {
		t.Errorf("expected ID port-empty, got %s", model.ID.ValueString())
	}
	// When Status is empty, mapInterfaceToState should not set it.
	if model.Status.ValueString() != "" {
		t.Errorf("expected Status to remain empty/zero, got %s", model.Status.ValueString())
	}
	// When MACAddress is empty, mapInterfaceToState should not set it.
	if model.MACAddress.ValueString() != "" {
		t.Errorf("expected MACAddress to remain empty/zero, got %s", model.MACAddress.ValueString())
	}
	// IPAddress should still be set.
	if model.IPAddress.ValueString() != "10.0.0.5" {
		t.Errorf("expected IPAddress 10.0.0.5, got %s", model.IPAddress.ValueString())
	}
}

func TestMapInterfaceToState_MultipleFixedIPs(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:     "port-multi",
		Status: "ACTIVE",
		FixedIPs: []fixedIP{
			{SubnetID: "subnet-1", IPAddress: "192.168.1.1"},
			{SubnetID: "subnet-2", IPAddress: "192.168.2.1"},
		},
	}

	mapInterfaceToState(model, iface)

	// The first fixed IP should be used when IPAddress is empty.
	if model.IPAddress.ValueString() != "192.168.1.1" {
		t.Errorf("expected IPAddress from first fixed_ip 192.168.1.1, got %s", model.IPAddress.ValueString())
	}
}

func TestMapInterfaceToState_DirectIPAndFixedIPs(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:        "port-both",
		IPAddress: "10.0.0.99",
		Status:    "ACTIVE",
		FixedIPs: []fixedIP{
			{SubnetID: "subnet-1", IPAddress: "192.168.1.1"},
		},
	}

	mapInterfaceToState(model, iface)

	// Direct IPAddress should take priority over FixedIPs.
	if model.IPAddress.ValueString() != "10.0.0.99" {
		t.Errorf("expected direct IPAddress 10.0.0.99 to take priority, got %s", model.IPAddress.ValueString())
	}
}

func TestInterfaceAPIResponse_JSON(t *testing.T) {
	jsonData := `{
		"id": "port-json-1",
		"subnet_id": "subnet-abc",
		"ip_address": "10.0.0.1",
		"status": "ACTIVE",
		"mac_address": "fa:16:3e:11:22:33",
		"fixed_ips": [
			{"subnet_id": "subnet-abc", "ip_address": "10.0.0.1"}
		]
	}`

	var resp interfaceAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal API response: %v", err)
	}

	if resp.ID != "port-json-1" {
		t.Errorf("expected ID port-json-1, got %s", resp.ID)
	}
	if resp.SubnetID != "subnet-abc" {
		t.Errorf("expected SubnetID subnet-abc, got %s", resp.SubnetID)
	}
	if resp.IPAddress != "10.0.0.1" {
		t.Errorf("expected IPAddress 10.0.0.1, got %s", resp.IPAddress)
	}
	if resp.Status != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", resp.Status)
	}
	if resp.MACAddress != "fa:16:3e:11:22:33" {
		t.Errorf("expected MACAddress fa:16:3e:11:22:33, got %s", resp.MACAddress)
	}
	if len(resp.FixedIPs) != 1 {
		t.Fatalf("expected 1 fixed IP, got %d", len(resp.FixedIPs))
	}
	if resp.FixedIPs[0].SubnetID != "subnet-abc" {
		t.Errorf("expected fixed_ip subnet_id subnet-abc, got %s", resp.FixedIPs[0].SubnetID)
	}
	if resp.FixedIPs[0].IPAddress != "10.0.0.1" {
		t.Errorf("expected fixed_ip ip_address 10.0.0.1, got %s", resp.FixedIPs[0].IPAddress)
	}
}

func TestFixedIP_JSON(t *testing.T) {
	jsonData := `{"subnet_id": "subnet-fix-1", "ip_address": "172.16.0.1"}`

	var fip fixedIP
	if err := json.Unmarshal([]byte(jsonData), &fip); err != nil {
		t.Fatalf("failed to unmarshal fixedIP: %v", err)
	}

	if fip.SubnetID != "subnet-fix-1" {
		t.Errorf("expected SubnetID subnet-fix-1, got %s", fip.SubnetID)
	}
	if fip.IPAddress != "172.16.0.1" {
		t.Errorf("expected IPAddress 172.16.0.1, got %s", fip.IPAddress)
	}

	// Verify round-trip: marshal back to JSON and check keys.
	data, err := json.Marshal(fip)
	if err != nil {
		t.Fatalf("failed to marshal fixedIP: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal round-trip: %v", err)
	}

	if parsed["subnet_id"] != "subnet-fix-1" {
		t.Errorf("expected JSON key subnet_id with value subnet-fix-1, got %v", parsed["subnet_id"])
	}
	if parsed["ip_address"] != "172.16.0.1" {
		t.Errorf("expected JSON key ip_address with value 172.16.0.1, got %v", parsed["ip_address"])
	}
}

// ---------------------------------------------------------------------------
// mapInterfaceToState edge cases
// ---------------------------------------------------------------------------

func TestMapInterfaceToState_NoFixedIPs(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:       "port-no-fips",
		Status:   "ACTIVE",
		FixedIPs: nil, // no fixed_ips at all
		// IPAddress is also empty
	}

	mapInterfaceToState(model, iface)

	if model.ID.ValueString() != "port-no-fips" {
		t.Errorf("expected ID port-no-fips, got %s", model.ID.ValueString())
	}
	// No IPAddress and no FixedIPs -> should be set to empty string.
	if model.IPAddress.ValueString() != "" {
		t.Errorf("expected empty IPAddress when no fixed_ips, got %s", model.IPAddress.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
}

func TestMapInterfaceToState_NilIPAddress(t *testing.T) {
	model := &routerInterfaceModel{}

	// fixed_ips with empty ip_address
	iface := &interfaceAPIResponse{
		ID:     "port-nil-ip",
		Status: "DOWN",
		FixedIPs: []fixedIP{
			{SubnetID: "subnet-x", IPAddress: ""},
		},
	}

	mapInterfaceToState(model, iface)

	if model.ID.ValueString() != "port-nil-ip" {
		t.Errorf("expected ID port-nil-ip, got %s", model.ID.ValueString())
	}
	// IPAddress field is empty, FixedIPs[0].IPAddress is also empty -> empty string.
	if model.IPAddress.ValueString() != "" {
		t.Errorf("expected empty IPAddress when fixed_ip has empty ip_address, got %s", model.IPAddress.ValueString())
	}
}

func TestInterfaceCreateRequest_Validation(t *testing.T) {
	// Verify subnet_id is present in the JSON body.
	req := interfaceCreateRequest{Subnet: "subnet-validate-1"}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create request: %v", err)
	}

	// Verify subnet key is present.
	subnet, ok := parsed["subnet"]
	if !ok {
		t.Fatal("expected 'subnet' key to be present in JSON body")
	}
	if subnet != "subnet-validate-1" {
		t.Errorf("expected subnet value 'subnet-validate-1', got %v", subnet)
	}

	// Verify no extra keys.
	if len(parsed) != 1 {
		t.Errorf("expected exactly 1 key in JSON, got %d", len(parsed))
	}

	// Verify empty subnet is still serialized.
	reqEmpty := interfaceCreateRequest{Subnet: ""}
	dataEmpty, err := json.Marshal(reqEmpty)
	if err != nil {
		t.Fatalf("failed to marshal empty create request: %v", err)
	}

	var parsedEmpty map[string]interface{}
	if err := json.Unmarshal(dataEmpty, &parsedEmpty); err != nil {
		t.Fatalf("failed to unmarshal empty create request: %v", err)
	}
	if _, ok := parsedEmpty["subnet"]; !ok {
		t.Error("expected 'subnet' key to be present even when empty (no omitempty)")
	}
}

// ===========================================================================
// NEW TESTS: Read response parsing, Create response parsing, edge cases
// ===========================================================================

// ---------------------------------------------------------------------------
// Read response parsing — {"ports": [...]} envelope, subnet_id filtering
// ---------------------------------------------------------------------------

func TestReadResponse_WrappedPorts_FindBySubnetID(t *testing.T) {
	// Simulate the Read parsing logic: unmarshal {"ports":[...]} and filter by subnet_id.
	rawJSON := `{"ports": [
		{"id": "port-aaa", "subnet_id": "subnet-111", "ip_address": "10.0.0.1", "status": "ACTIVE", "mac_address": "fa:16:3e:aa:aa:aa"},
		{"id": "port-bbb", "subnet_id": "subnet-222", "ip_address": "10.0.0.2", "status": "ACTIVE", "mac_address": "fa:16:3e:bb:bb:bb"},
		{"id": "port-ccc", "subnet_id": "subnet-333", "ip_address": "10.0.0.3", "status": "ACTIVE", "mac_address": "fa:16:3e:cc:cc:cc"}
	]}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal wrapped ports: %v", err)
	}

	if len(wrapped.Ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(wrapped.Ports))
	}

	// Search for subnet-222
	targetSubnet := "subnet-222"
	found := false
	var matched interfaceAPIResponse
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == targetSubnet {
			matched = iface
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected to find port matching subnet-222")
	}
	if matched.ID != "port-bbb" {
		t.Errorf("expected port ID port-bbb, got %s", matched.ID)
	}
	if matched.IPAddress != "10.0.0.2" {
		t.Errorf("expected IPAddress 10.0.0.2, got %s", matched.IPAddress)
	}

	// Map to state and verify
	model := &routerInterfaceModel{
		SubnetID: types.StringValue("subnet-222"),
		RouterID: types.StringValue("router-1"),
	}
	mapInterfaceToState(model, &matched)

	if model.ID.ValueString() != "port-bbb" {
		t.Errorf("expected ID port-bbb, got %s", model.ID.ValueString())
	}
	if model.MACAddress.ValueString() != "fa:16:3e:bb:bb:bb" {
		t.Errorf("expected MACAddress fa:16:3e:bb:bb:bb, got %s", model.MACAddress.ValueString())
	}
}

func TestReadResponse_WrappedPorts_PortNotFound(t *testing.T) {
	rawJSON := `{"ports": [
		{"id": "port-aaa", "subnet_id": "subnet-111", "ip_address": "10.0.0.1", "status": "ACTIVE", "mac_address": "fa:16:3e:aa:aa:aa"}
	]}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Search for a subnet that doesn't exist
	targetSubnet := "subnet-999"
	found := false
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == targetSubnet {
			found = true
			break
		}
		// Also check fixed_ips
		for _, fip := range iface.FixedIPs {
			if fip.SubnetID == targetSubnet {
				found = true
				break
			}
		}
	}

	if found {
		t.Error("expected NOT to find port matching subnet-999")
	}
}

func TestReadResponse_WrappedPorts_EmptyList(t *testing.T) {
	rawJSON := `{"ports": []}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(wrapped.Ports) != 0 {
		t.Errorf("expected 0 ports, got %d", len(wrapped.Ports))
	}

	// No port should be found
	found := false
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == "subnet-111" {
			found = true
			break
		}
	}
	if found {
		t.Error("expected NOT to find any port in empty list")
	}
}

func TestReadResponse_FindByFixedIPSubnetID(t *testing.T) {
	// Some interfaces have subnet_id only in fixed_ips, not at top level.
	rawJSON := `{"ports": [
		{
			"id": "port-fip",
			"subnet_id": "",
			"ip_address": "",
			"status": "ACTIVE",
			"mac_address": "fa:16:3e:dd:dd:dd",
			"fixed_ips": [
				{"subnet_id": "subnet-hidden", "ip_address": "10.0.1.5"}
			]
		}
	]}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Replicate Read logic: check subnet_id then fixed_ips
	targetSubnet := "subnet-hidden"
	found := false
	var matchedIP string
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == targetSubnet {
			found = true
			break
		}
		for _, fip := range iface.FixedIPs {
			if fip.SubnetID == targetSubnet {
				found = true
				matchedIP = fip.IPAddress
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Fatal("expected to find port by fixed_ips subnet_id")
	}
	if matchedIP != "10.0.1.5" {
		t.Errorf("expected matched IP 10.0.1.5, got %s", matchedIP)
	}
}

func TestReadResponse_FindByPortID(t *testing.T) {
	// When subnet_id doesn't match, the Read logic falls back to matching by port ID.
	rawJSON := `{"ports": [
		{"id": "port-fallback", "subnet_id": "subnet-other", "ip_address": "10.0.2.1", "status": "ACTIVE", "mac_address": "fa:16:3e:ee:ee:ee"}
	]}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// State has a different subnet but matching port ID
	stateID := "port-fallback"
	stateSubnet := "subnet-different"
	found := false
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == stateSubnet {
			found = true
			break
		}
		for _, fip := range iface.FixedIPs {
			if fip.SubnetID == stateSubnet {
				found = true
				break
			}
		}
		if found {
			break
		}
		// Fallback: match by port ID
		if iface.ID == stateID {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected to find port by fallback port ID match")
	}
}

func TestReadResponse_DirectArrayFormat(t *testing.T) {
	// API may return a direct array instead of {"ports": [...]}
	rawJSON := `[
		{"id": "port-direct-1", "subnet_id": "subnet-d1", "ip_address": "10.0.3.1", "status": "ACTIVE", "mac_address": "fa:16:3e:11:11:11"},
		{"id": "port-direct-2", "subnet_id": "subnet-d2", "ip_address": "10.0.3.2", "status": "DOWN", "mac_address": "fa:16:3e:22:22:22"}
	]`

	// Replicate the Read parsing logic
	var interfaces []interfaceAPIResponse
	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err == nil && len(wrapped.Ports) > 0 {
		interfaces = wrapped.Ports
	} else if err := json.Unmarshal([]byte(rawJSON), &interfaces); err != nil {
		t.Fatalf("failed to unmarshal direct array: %v", err)
	}

	if len(interfaces) != 2 {
		t.Fatalf("expected 2 interfaces, got %d", len(interfaces))
	}
	if interfaces[0].ID != "port-direct-1" {
		t.Errorf("expected first port ID port-direct-1, got %s", interfaces[0].ID)
	}
	if interfaces[1].ID != "port-direct-2" {
		t.Errorf("expected second port ID port-direct-2, got %s", interfaces[1].ID)
	}
}

func TestReadResponse_SingleObjectFormat(t *testing.T) {
	// API may return a single object instead of array
	rawJSON := `{"id": "port-single", "subnet_id": "subnet-s1", "ip_address": "10.0.4.1", "status": "ACTIVE", "mac_address": "fa:16:3e:33:33:33"}`

	// Replicate the Read parsing logic
	var interfaces []interfaceAPIResponse
	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err == nil && len(wrapped.Ports) > 0 {
		interfaces = wrapped.Ports
	} else if err := json.Unmarshal([]byte(rawJSON), &interfaces); err != nil {
		// Try single object format
		var single interfaceAPIResponse
		if err2 := json.Unmarshal([]byte(rawJSON), &single); err2 != nil {
			t.Fatalf("failed to unmarshal single object: %v", err2)
		}
		interfaces = []interfaceAPIResponse{single}
	}

	if len(interfaces) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(interfaces))
	}
	if interfaces[0].ID != "port-single" {
		t.Errorf("expected port ID port-single, got %s", interfaces[0].ID)
	}
	if interfaces[0].SubnetID != "subnet-s1" {
		t.Errorf("expected subnet_id subnet-s1, got %s", interfaces[0].SubnetID)
	}
}

// ---------------------------------------------------------------------------
// Create response parsing — port_id as resource ID
// ---------------------------------------------------------------------------

func TestCreateResponse_PortIDAsResourceID(t *testing.T) {
	// API returns id=router_id, port_id=actual_port_id
	createJSON := `{
		"id": "router-uuid-123",
		"port_id": "port-uuid-456",
		"subnet_id": "subnet-abc",
		"ip_address": "10.0.0.1",
		"status": "ACTIVE",
		"mac_address": "fa:16:3e:ff:ff:ff"
	}`

	var result interfaceAPIResponse
	if err := json.Unmarshal([]byte(createJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}

	// Replicate Create logic: use port_id as resource ID if present
	var resourceID string
	if result.PortID != "" {
		resourceID = result.PortID
	} else {
		resourceID = result.ID
	}

	if resourceID != "port-uuid-456" {
		t.Errorf("expected resource ID from port_id 'port-uuid-456', got %s", resourceID)
	}
}

func TestCreateResponse_FallbackToID(t *testing.T) {
	// If port_id is empty, fall back to id
	createJSON := `{
		"id": "router-uuid-789",
		"port_id": "",
		"subnet_id": "subnet-def",
		"ip_address": "10.0.0.2",
		"status": "ACTIVE",
		"mac_address": "fa:16:3e:00:00:00"
	}`

	var result interfaceAPIResponse
	if err := json.Unmarshal([]byte(createJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}

	var resourceID string
	if result.PortID != "" {
		resourceID = result.PortID
	} else {
		resourceID = result.ID
	}

	if resourceID != "router-uuid-789" {
		t.Errorf("expected resource ID from id 'router-uuid-789', got %s", resourceID)
	}
}

func TestCreateResponse_MissingPortID(t *testing.T) {
	// port_id field completely absent
	createJSON := `{
		"id": "router-only-id",
		"subnet_id": "subnet-ghi",
		"ip_address": "10.0.0.3",
		"status": "ACTIVE"
	}`

	var result interfaceAPIResponse
	if err := json.Unmarshal([]byte(createJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}

	// PortID will be "" when not in JSON
	var resourceID string
	if result.PortID != "" {
		resourceID = result.PortID
	} else {
		resourceID = result.ID
	}

	if resourceID != "router-only-id" {
		t.Errorf("expected resource ID from id 'router-only-id', got %s", resourceID)
	}
}

func TestCreateResponse_IPFromFixedIPs(t *testing.T) {
	// Create response where ip_address is empty but fixed_ips has data
	createJSON := `{
		"id": "router-fip-create",
		"port_id": "port-fip-create",
		"subnet_id": "subnet-jkl",
		"ip_address": "",
		"status": "ACTIVE",
		"mac_address": "fa:16:3e:ab:cd:ef",
		"fixed_ips": [{"subnet_id": "subnet-jkl", "ip_address": "10.0.5.1"}]
	}`

	var result interfaceAPIResponse
	if err := json.Unmarshal([]byte(createJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Replicate Create logic for IP extraction
	var ip string
	if result.IPAddress != "" {
		ip = result.IPAddress
	} else if len(result.FixedIPs) > 0 {
		ip = result.FixedIPs[0].IPAddress
	}

	if ip != "10.0.5.1" {
		t.Errorf("expected IP from fixed_ips '10.0.5.1', got %s", ip)
	}
}

func TestCreateResponse_NoIPAnywhere(t *testing.T) {
	// Create response with no IP at all
	createJSON := `{
		"id": "router-no-ip",
		"port_id": "port-no-ip",
		"subnet_id": "subnet-mno",
		"status": "BUILD"
	}`

	var result interfaceAPIResponse
	if err := json.Unmarshal([]byte(createJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	var ip string
	if result.IPAddress != "" {
		ip = result.IPAddress
	} else if len(result.FixedIPs) > 0 {
		ip = result.FixedIPs[0].IPAddress
	}

	if ip != "" {
		t.Errorf("expected empty IP, got %s", ip)
	}
}

// ---------------------------------------------------------------------------
// mapInterfaceToState — full field mapping
// ---------------------------------------------------------------------------

func TestMapInterfaceToState_AllFieldsMapped(t *testing.T) {
	model := &routerInterfaceModel{
		RouterID: types.StringValue("router-999"),
		SubnetID: types.StringValue("subnet-999"),
	}

	iface := &interfaceAPIResponse{
		ID:         "port-full-map",
		SubnetID:   "subnet-999",
		IPAddress:  "172.16.0.100",
		Status:     "ACTIVE",
		MACAddress: "fa:16:3e:de:ad:be",
		FixedIPs: []fixedIP{
			{SubnetID: "subnet-999", IPAddress: "172.16.0.100"},
		},
	}

	mapInterfaceToState(model, iface)

	// ID should be set to port ID from the interface
	if model.ID.ValueString() != "port-full-map" {
		t.Errorf("expected ID port-full-map, got %s", model.ID.ValueString())
	}
	// RouterID should remain unchanged (mapInterfaceToState does not modify it)
	if model.RouterID.ValueString() != "router-999" {
		t.Errorf("expected RouterID preserved as router-999, got %s", model.RouterID.ValueString())
	}
	// SubnetID should remain unchanged (mapInterfaceToState does not modify it)
	if model.SubnetID.ValueString() != "subnet-999" {
		t.Errorf("expected SubnetID preserved as subnet-999, got %s", model.SubnetID.ValueString())
	}
	if model.IPAddress.ValueString() != "172.16.0.100" {
		t.Errorf("expected IPAddress 172.16.0.100, got %s", model.IPAddress.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
	if model.MACAddress.ValueString() != "fa:16:3e:de:ad:be" {
		t.Errorf("expected MACAddress fa:16:3e:de:ad:be, got %s", model.MACAddress.ValueString())
	}
}

func TestMapInterfaceToState_PreservesRouterIDAndSubnetID(t *testing.T) {
	// mapInterfaceToState should NOT overwrite router_id or subnet_id.
	model := &routerInterfaceModel{
		RouterID: types.StringValue("router-preserve"),
		SubnetID: types.StringValue("subnet-preserve"),
	}

	iface := &interfaceAPIResponse{
		ID:        "port-new",
		SubnetID:  "subnet-different", // different from model's SubnetID
		IPAddress: "10.0.0.50",
		Status:    "DOWN",
	}

	mapInterfaceToState(model, iface)

	// router_id should remain untouched
	if model.RouterID.ValueString() != "router-preserve" {
		t.Errorf("expected RouterID to be preserved as 'router-preserve', got %s", model.RouterID.ValueString())
	}
	// subnet_id should remain untouched (mapInterfaceToState does not modify it)
	if model.SubnetID.ValueString() != "subnet-preserve" {
		t.Errorf("expected SubnetID to be preserved as 'subnet-preserve', got %s", model.SubnetID.ValueString())
	}
	// But ID, Status, MACAddress, IPAddress SHOULD be updated
	if model.ID.ValueString() != "port-new" {
		t.Errorf("expected ID port-new, got %s", model.ID.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Read response: multiple interfaces, priority matching
// ---------------------------------------------------------------------------

func TestReadResponse_PortIDMatchBeforeSubnetInLaterPort(t *testing.T) {
	// The Read logic iterates ports in order. For each port it checks:
	// (1) subnet_id match, (2) fixed_ips subnet match, (3) port ID match.
	// If port ID matches in the first port, it wins before the loop reaches
	// a later port with a subnet match.
	rawJSON := `{"ports": [
		{"id": "port-by-id", "subnet_id": "subnet-other", "ip_address": "10.0.0.1", "status": "ACTIVE", "mac_address": "fa:16:3e:11:11:11"},
		{"id": "port-by-subnet", "subnet_id": "subnet-target", "ip_address": "10.0.0.2", "status": "ACTIVE", "mac_address": "fa:16:3e:22:22:22"}
	]}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	stateSubnet := "subnet-target"
	stateID := "port-by-id" // matches first port's ID

	found := false
	var matchedID string
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == stateSubnet {
			matchedID = iface.ID
			found = true
			break
		}
		for _, fip := range iface.FixedIPs {
			if fip.SubnetID == stateSubnet {
				matchedID = iface.ID
				found = true
				break
			}
		}
		if found {
			break
		}
		if iface.ID == stateID {
			matchedID = iface.ID
			found = true
			break
		}
	}

	if !found {
		t.Fatal("expected to find a matching port")
	}
	// Port ID match on the first port is found before the loop reaches the
	// second port with the subnet match.
	if matchedID != "port-by-id" {
		t.Errorf("expected port-by-id (first port, matched by ID), got %s", matchedID)
	}
}

func TestReadResponse_SubnetMatchWinsOnSamePort(t *testing.T) {
	// When the same port matches both by subnet_id AND port ID,
	// subnet_id check comes first in the Read logic.
	rawJSON := `{"ports": [
		{"id": "port-both", "subnet_id": "subnet-target", "ip_address": "10.0.0.1", "status": "ACTIVE", "mac_address": "fa:16:3e:11:11:11"}
	]}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	stateSubnet := "subnet-target"
	stateID := "port-both"

	found := false
	matchedBy := ""
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == stateSubnet {
			found = true
			matchedBy = "subnet"
			break
		}
		for _, fip := range iface.FixedIPs {
			if fip.SubnetID == stateSubnet {
				found = true
				matchedBy = "fixed_ip"
				break
			}
		}
		if found {
			break
		}
		if iface.ID == stateID {
			found = true
			matchedBy = "port_id"
			break
		}
	}

	if !found {
		t.Fatal("expected to find a matching port")
	}
	if matchedBy != "subnet" {
		t.Errorf("expected match by 'subnet', got '%s'", matchedBy)
	}
}

func TestReadResponse_FixedIPSubnetOverrideIP(t *testing.T) {
	// When found via fixed_ips, the fixed_ip's IPAddress should be used
	// (replicating the Read logic at lines 166-170 in resource.go).
	rawJSON := `{"ports": [
		{
			"id": "port-fip-override",
			"subnet_id": "",
			"ip_address": "10.0.0.99",
			"status": "ACTIVE",
			"mac_address": "fa:16:3e:ab:ab:ab",
			"fixed_ips": [
				{"subnet_id": "subnet-target", "ip_address": "10.0.1.55"}
			]
		}
	]}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	stateSubnet := "subnet-target"
	model := &routerInterfaceModel{
		SubnetID: types.StringValue(stateSubnet),
		RouterID: types.StringValue("router-1"),
	}

	found := false
	for _, iface := range wrapped.Ports {
		if iface.SubnetID == stateSubnet {
			mapInterfaceToState(model, &iface)
			found = true
			break
		}
		for _, fip := range iface.FixedIPs {
			if fip.SubnetID == stateSubnet {
				mapInterfaceToState(model, &iface)
				// Override IP from fixed_ips (replicating Read logic)
				if fip.IPAddress != "" {
					model.IPAddress = types.StringValue(fip.IPAddress)
				}
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Fatal("expected to find port via fixed_ips")
	}
	// The IP should be from fixed_ips, not from the top-level ip_address
	if model.IPAddress.ValueString() != "10.0.1.55" {
		t.Errorf("expected IP from fixed_ips '10.0.1.55', got %s", model.IPAddress.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestMapInterfaceToState_EmptyIPAddress(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:        "port-empty-ip",
		IPAddress: "",
		Status:    "BUILD",
		FixedIPs:  []fixedIP{}, // empty slice, not nil
	}

	mapInterfaceToState(model, iface)

	// Should be set to empty string value (not null/unknown)
	if model.IPAddress.ValueString() != "" {
		t.Errorf("expected empty IPAddress, got %s", model.IPAddress.ValueString())
	}
	if model.IPAddress.IsNull() {
		t.Error("expected IPAddress to be empty string value, not null")
	}
}

func TestMapInterfaceToState_FixedIPsEmptyIPAddress(t *testing.T) {
	model := &routerInterfaceModel{}

	iface := &interfaceAPIResponse{
		ID:     "port-fip-empty-ip",
		Status: "ACTIVE",
		FixedIPs: []fixedIP{
			{SubnetID: "subnet-1", IPAddress: ""},
			{SubnetID: "subnet-2", IPAddress: "10.0.0.5"},
		},
	}

	mapInterfaceToState(model, iface)

	// First fixed IP has empty address, should still use it (it's first)
	if model.IPAddress.ValueString() != "" {
		t.Errorf("expected empty IPAddress from first fixed_ip, got %s", model.IPAddress.ValueString())
	}
}

func TestReadResponse_InvalidJSON(t *testing.T) {
	rawJSON := `not valid json`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	err := json.Unmarshal([]byte(rawJSON), &wrapped)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReadResponse_EmptyPortsKey(t *testing.T) {
	// "ports" key exists but maps to null
	rawJSON := `{"ports": null}`

	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapped); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(wrapped.Ports) != 0 {
		t.Errorf("expected 0 ports for null, got %d", len(wrapped.Ports))
	}
}

func TestCreateResponse_FullLifecycleMapping(t *testing.T) {
	// Simulate full Create logic: parse response, set ID, set computed fields
	createJSON := `{
		"id": "router-uuid-abc",
		"port_id": "port-uuid-def",
		"subnet_id": "subnet-uvw",
		"ip_address": "10.0.10.1",
		"status": "ACTIVE",
		"mac_address": "fa:16:3e:12:34:56",
		"fixed_ips": [{"subnet_id": "subnet-uvw", "ip_address": "10.0.10.1"}]
	}`

	var result interfaceAPIResponse
	if err := json.Unmarshal([]byte(createJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	plan := &routerInterfaceModel{
		RouterID: types.StringValue("router-uuid-abc"),
		SubnetID: types.StringValue("subnet-uvw"),
	}

	// Set ID from port_id
	if result.PortID != "" {
		plan.ID = types.StringValue(result.PortID)
	} else {
		plan.ID = types.StringValue(result.ID)
	}
	plan.Status = types.StringValue(result.Status)
	plan.MACAddress = types.StringValue(result.MACAddress)

	// Extract IP
	if result.IPAddress != "" {
		plan.IPAddress = types.StringValue(result.IPAddress)
	} else if len(result.FixedIPs) > 0 {
		plan.IPAddress = types.StringValue(result.FixedIPs[0].IPAddress)
	} else {
		plan.IPAddress = types.StringValue("")
	}

	if plan.ID.ValueString() != "port-uuid-def" {
		t.Errorf("expected ID port-uuid-def, got %s", plan.ID.ValueString())
	}
	if plan.RouterID.ValueString() != "router-uuid-abc" {
		t.Errorf("expected RouterID router-uuid-abc, got %s", plan.RouterID.ValueString())
	}
	if plan.SubnetID.ValueString() != "subnet-uvw" {
		t.Errorf("expected SubnetID subnet-uvw, got %s", plan.SubnetID.ValueString())
	}
	if plan.IPAddress.ValueString() != "10.0.10.1" {
		t.Errorf("expected IPAddress 10.0.10.1, got %s", plan.IPAddress.ValueString())
	}
	if plan.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", plan.Status.ValueString())
	}
	if plan.MACAddress.ValueString() != "fa:16:3e:12:34:56" {
		t.Errorf("expected MACAddress fa:16:3e:12:34:56, got %s", plan.MACAddress.ValueString())
	}
}
