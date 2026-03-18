package subnet

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

var allocationPoolObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"start": types.StringType,
		"end":   types.StringType,
	},
}

var hostRouteObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"destination": types.StringType,
		"nexthop":     types.StringType,
	},
}

// nullPlan returns a subnetModel with required fields set and all optional fields null.
func nullPlan(name, cidr string, ipVersion int64) *subnetModel {
	return &subnetModel{
		Name:            types.StringValue(name),
		CIDR:            types.StringValue(cidr),
		IPVersion:       types.Int64Value(ipVersion),
		VPCID:           types.StringNull(),
		Description:     types.StringNull(),
		EnableDHCP:      types.BoolNull(),
		GatewayIP:       types.StringNull(),
		DNSNameservers:  types.ListNull(types.StringType),
		AllocationPools: types.ListNull(allocationPoolObjectType),
		HostRoutes:      types.ListNull(hostRouteObjectType),
	}
}

// parseReadResponse is a helper that simulates the Read response-parsing logic
// from resource.go so we can unit-test it without an HTTP server.
func parseReadResponse(t *testing.T, ctx context.Context, raw string) subnetModel {
	t.Helper()

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// We need a pre-populated state with list types set so ElementType works.
	state := subnetModel{
		AllocationPools: types.ListNull(allocationPoolObjectType),
		HostRoutes:      types.ListNull(hostRouteObjectType),
	}

	if v, ok := result["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := result["cidr"].(string); ok {
		state.CIDR = types.StringValue(v)
	}
	if v, ok := result["vpc_id"].(string); ok && v != "" {
		state.VPCID = types.StringValue(v)
	} else if v, ok := result["network_id"].(string); ok && v != "" {
		state.VPCID = types.StringValue(v)
	}
	if v, ok := result["ip_version"].(float64); ok {
		state.IPVersion = types.Int64Value(int64(v))
	}
	if v, ok := result["description"].(string); ok && v != "" {
		state.Description = types.StringValue(v)
	} else {
		state.Description = types.StringNull()
	}
	if v, ok := result["enable_dhcp"].(bool); ok {
		state.EnableDHCP = types.BoolValue(v)
	} else {
		state.EnableDHCP = types.BoolNull()
	}
	if v, ok := result["gateway_ip"].(string); ok && v != "" {
		state.GatewayIP = types.StringValue(v)
	} else {
		state.GatewayIP = types.StringNull()
	}

	// DNS nameservers.
	if v, ok := result["dns_nameservers"]; ok && v != nil {
		nsJSON, _ := json.Marshal(v)
		var nameservers []string
		if err := json.Unmarshal(nsJSON, &nameservers); err == nil && len(nameservers) > 0 {
			nsList, _ := types.ListValueFrom(ctx, types.StringType, nameservers)
			state.DNSNameservers = nsList
		} else {
			state.DNSNameservers = types.ListNull(types.StringType)
		}
	} else {
		state.DNSNameservers = types.ListNull(types.StringType)
	}

	// Allocation pools.
	if v, ok := result["allocation_pools"]; ok && v != nil {
		poolsJSON, _ := json.Marshal(v)
		var pools []map[string]string
		if err := json.Unmarshal(poolsJSON, &pools); err == nil && len(pools) > 0 {
			var poolModels []allocationPoolModel
			for _, p := range pools {
				poolModels = append(poolModels, allocationPoolModel{
					Start: types.StringValue(p["start"]),
					End:   types.StringValue(p["end"]),
				})
			}
			poolsList, _ := types.ListValueFrom(ctx, allocationPoolObjectType, poolModels)
			state.AllocationPools = poolsList
		}
	}

	// Host routes.
	if v, ok := result["host_routes"]; ok && v != nil {
		routesJSON, _ := json.Marshal(v)
		var routes []map[string]string
		if err := json.Unmarshal(routesJSON, &routes); err == nil && len(routes) > 0 {
			var routeModels []hostRouteModel
			for _, hr := range routes {
				routeModels = append(routeModels, hostRouteModel{
					Destination: types.StringValue(hr["destination"]),
					Nexthop:     types.StringValue(hr["nexthop"]),
				})
			}
			routesList, _ := types.ListValueFrom(ctx, hostRouteObjectType, routeModels)
			state.HostRoutes = routesList
		}
	}

	return state
}

// ---------------------------------------------------------------------------
// Factory / model sanity
// ---------------------------------------------------------------------------

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// ---------------------------------------------------------------------------
// CIDR regex tests
// ---------------------------------------------------------------------------

func TestCIDRRegex(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		isValid bool
	}{
		// Valid
		{"class A", "10.0.0.0/8", true},
		{"class C", "192.168.1.0/24", true},
		{"class B private", "172.16.0.0/16", true},
		{"default route", "0.0.0.0/0", true},
		{"host route max", "255.255.255.255/32", true},
		{"prefix /1", "128.0.0.0/1", true},
		{"prefix /31", "10.0.0.0/31", true},
		{"single digit octets", "1.2.3.4/24", true},
		{"double digit octets", "10.20.30.40/16", true},
		{"max first octet 255", "255.0.0.0/8", true},
		// Invalid
		{"octet out of range", "256.0.0.0/8", false},
		{"missing prefix length", "10.0.0.0", false},
		{"prefix length too large", "10.0.0.0/33", false},
		{"not a cidr", "not-a-cidr", false},
		{"empty string", "", false},
		{"negative prefix", "10.0.0.0/-1", false},
		{"five octets", "10.0.0.0.0/24", false},
		{"three octets", "10.0.0/24", false},
		{"double slash", "10.0.0.0//24", false},
		{"letters in octet", "abc.0.0.0/24", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matched := cidrRegex.MatchString(tc.cidr)
			if matched != tc.isValid {
				if tc.isValid {
					t.Errorf("expected cidrRegex to match %q, but it did not", tc.cidr)
				} else {
					t.Errorf("expected cidrRegex NOT to match %q, but it did", tc.cidr)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// buildBody tests — table-driven for each optional field combination
// ---------------------------------------------------------------------------

func TestBuildBody_Minimal(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}
	plan := nullPlan("min-subnet", "10.0.0.0/24", 4)

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}

	if body["name"] != "min-subnet" {
		t.Errorf("name: got %v", body["name"])
	}
	if body["cidr"] != "10.0.0.0/24" {
		t.Errorf("cidr: got %v", body["cidr"])
	}
	if body["ip_version"] != int64(4) {
		t.Errorf("ip_version: got %v", body["ip_version"])
	}

	for _, key := range []string{"vpc_id", "description", "enable_dhcp", "gateway_ip", "dns_nameservers", "allocation_pools", "host_routes"} {
		if _, exists := body[key]; exists {
			t.Errorf("key %q should be absent in minimal body", key)
		}
	}
}

func TestBuildBody_WithVPCID(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}
	plan := nullPlan("vpc-subnet", "10.0.1.0/24", 4)
	plan.VPCID = types.StringValue("vpc-abc-123")

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}
	if body["vpc_id"] != "vpc-abc-123" {
		t.Errorf("vpc_id: got %v", body["vpc_id"])
	}
}

func TestBuildBody_WithDescription(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}
	plan := nullPlan("desc-subnet", "10.0.2.0/24", 4)
	plan.Description = types.StringValue("my description")

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}
	if body["description"] != "my description" {
		t.Errorf("description: got %v", body["description"])
	}
}

func TestBuildBody_WithEnableDHCP(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	tests := []struct {
		name   string
		dhcp   bool
		expect bool
	}{
		{"true", true, true},
		{"false", false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := nullPlan("dhcp-subnet", "10.0.3.0/24", 4)
			plan.EnableDHCP = types.BoolValue(tc.dhcp)

			body, err := r.buildBody(ctx, plan)
			if err != nil {
				t.Fatalf("buildBody error: %v", err)
			}
			if body["enable_dhcp"] != tc.expect {
				t.Errorf("enable_dhcp: got %v, want %v", body["enable_dhcp"], tc.expect)
			}
		})
	}
}

func TestBuildBody_WithGatewayIP(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}
	plan := nullPlan("gw-subnet", "10.0.4.0/24", 4)
	plan.GatewayIP = types.StringValue("10.0.4.1")

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}
	if body["gateway_ip"] != "10.0.4.1" {
		t.Errorf("gateway_ip: got %v", body["gateway_ip"])
	}
}

func TestBuildBody_WithDNSNameservers(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	tests := []struct {
		name   string
		dns    []string
		expect []string
	}{
		{"single", []string{"8.8.8.8"}, []string{"8.8.8.8"}},
		{"multiple", []string{"8.8.8.8", "1.1.1.1", "8.8.4.4"}, []string{"8.8.8.8", "1.1.1.1", "8.8.4.4"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := nullPlan("dns-subnet", "10.0.5.0/24", 4)
			dnsList, diags := types.ListValueFrom(ctx, types.StringType, tc.dns)
			if diags.HasError() {
				t.Fatalf("list build error: %s", diags.Errors())
			}
			plan.DNSNameservers = dnsList

			body, err := r.buildBody(ctx, plan)
			if err != nil {
				t.Fatalf("buildBody error: %v", err)
			}

			ns, ok := body["dns_nameservers"].([]string)
			if !ok {
				t.Fatalf("dns_nameservers type: got %T", body["dns_nameservers"])
			}
			if len(ns) != len(tc.expect) {
				t.Fatalf("dns_nameservers len: got %d, want %d", len(ns), len(tc.expect))
			}
			for i, v := range tc.expect {
				if ns[i] != v {
					t.Errorf("dns_nameservers[%d]: got %s, want %s", i, ns[i], v)
				}
			}
		})
	}
}

func TestBuildBody_EmptyDNS(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}
	plan := nullPlan("empty-dns", "10.0.5.0/24", 4)
	emptyDNS, diags := types.ListValueFrom(ctx, types.StringType, []string{})
	if diags.HasError() {
		t.Fatalf("list build error: %s", diags.Errors())
	}
	plan.DNSNameservers = emptyDNS

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}

	ns, ok := body["dns_nameservers"].([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", body["dns_nameservers"])
	}
	if len(ns) != 0 {
		t.Errorf("expected 0 nameservers, got %d", len(ns))
	}
}

func TestBuildBody_WithAllocationPools(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	tests := []struct {
		name   string
		pools  []allocationPoolModel
		starts []string
		ends   []string
	}{
		{
			"single pool",
			[]allocationPoolModel{{Start: types.StringValue("10.0.0.10"), End: types.StringValue("10.0.0.100")}},
			[]string{"10.0.0.10"}, []string{"10.0.0.100"},
		},
		{
			"multiple pools",
			[]allocationPoolModel{
				{Start: types.StringValue("10.0.0.10"), End: types.StringValue("10.0.0.50")},
				{Start: types.StringValue("10.0.0.100"), End: types.StringValue("10.0.0.200")},
			},
			[]string{"10.0.0.10", "10.0.0.100"}, []string{"10.0.0.50", "10.0.0.200"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := nullPlan("pool-subnet", "10.0.0.0/24", 4)
			poolsList, diags := types.ListValueFrom(ctx, allocationPoolObjectType, tc.pools)
			if diags.HasError() {
				t.Fatalf("list build error: %s", diags.Errors())
			}
			plan.AllocationPools = poolsList

			body, err := r.buildBody(ctx, plan)
			if err != nil {
				t.Fatalf("buildBody error: %v", err)
			}

			ap, ok := body["allocation_pools"].([]map[string]string)
			if !ok {
				t.Fatalf("allocation_pools type: got %T", body["allocation_pools"])
			}
			if len(ap) != len(tc.starts) {
				t.Fatalf("allocation_pools len: got %d, want %d", len(ap), len(tc.starts))
			}
			for i := range tc.starts {
				if ap[i]["start"] != tc.starts[i] {
					t.Errorf("pool[%d].start: got %s, want %s", i, ap[i]["start"], tc.starts[i])
				}
				if ap[i]["end"] != tc.ends[i] {
					t.Errorf("pool[%d].end: got %s, want %s", i, ap[i]["end"], tc.ends[i])
				}
			}
		})
	}
}

func TestBuildBody_WithHostRoutes(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	tests := []struct {
		name         string
		routes       []hostRouteModel
		destinations []string
		nexthops     []string
	}{
		{
			"single route",
			[]hostRouteModel{{Destination: types.StringValue("192.168.0.0/16"), Nexthop: types.StringValue("10.0.0.1")}},
			[]string{"192.168.0.0/16"}, []string{"10.0.0.1"},
		},
		{
			"multiple routes",
			[]hostRouteModel{
				{Destination: types.StringValue("192.168.0.0/16"), Nexthop: types.StringValue("10.0.0.1")},
				{Destination: types.StringValue("172.16.0.0/12"), Nexthop: types.StringValue("10.0.0.2")},
			},
			[]string{"192.168.0.0/16", "172.16.0.0/12"}, []string{"10.0.0.1", "10.0.0.2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := nullPlan("route-subnet", "10.0.0.0/24", 4)
			routesList, diags := types.ListValueFrom(ctx, hostRouteObjectType, tc.routes)
			if diags.HasError() {
				t.Fatalf("list build error: %s", diags.Errors())
			}
			plan.HostRoutes = routesList

			body, err := r.buildBody(ctx, plan)
			if err != nil {
				t.Fatalf("buildBody error: %v", err)
			}

			hr, ok := body["host_routes"].([]map[string]string)
			if !ok {
				t.Fatalf("host_routes type: got %T", body["host_routes"])
			}
			if len(hr) != len(tc.destinations) {
				t.Fatalf("host_routes len: got %d, want %d", len(hr), len(tc.destinations))
			}
			for i := range tc.destinations {
				if hr[i]["destination"] != tc.destinations[i] {
					t.Errorf("route[%d].destination: got %s, want %s", i, hr[i]["destination"], tc.destinations[i])
				}
				if hr[i]["nexthop"] != tc.nexthops[i] {
					t.Errorf("route[%d].nexthop: got %s, want %s", i, hr[i]["nexthop"], tc.nexthops[i])
				}
			}
		})
	}
}

func TestBuildBody_AllFields(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	dnsList, _ := types.ListValueFrom(ctx, types.StringType, []string{"8.8.8.8", "1.1.1.1"})
	poolsList, _ := types.ListValueFrom(ctx, allocationPoolObjectType, []allocationPoolModel{
		{Start: types.StringValue("10.0.1.10"), End: types.StringValue("10.0.1.100")},
	})
	routesList, _ := types.ListValueFrom(ctx, hostRouteObjectType, []hostRouteModel{
		{Destination: types.StringValue("192.168.0.0/16"), Nexthop: types.StringValue("10.0.1.1")},
	})

	plan := &subnetModel{
		ID:              types.StringValue("subnet-123"),
		Name:            types.StringValue("full-subnet"),
		CIDR:            types.StringValue("10.0.1.0/24"),
		VPCID:           types.StringValue("vpc-abc"),
		IPVersion:       types.Int64Value(4),
		Description:     types.StringValue("A test subnet"),
		EnableDHCP:      types.BoolValue(true),
		GatewayIP:       types.StringValue("10.0.1.1"),
		DNSNameservers:  dnsList,
		AllocationPools: poolsList,
		HostRoutes:      routesList,
	}

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}

	// Verify every field is present.
	checks := map[string]interface{}{
		"name":       "full-subnet",
		"cidr":       "10.0.1.0/24",
		"vpc_id":     "vpc-abc",
		"ip_version": int64(4),
		"enable_dhcp": true,
		"gateway_ip":  "10.0.1.1",
		"description": "A test subnet",
	}
	for k, want := range checks {
		if body[k] != want {
			t.Errorf("%s: got %v, want %v", k, body[k], want)
		}
	}

	ns := body["dns_nameservers"].([]string)
	if len(ns) != 2 || ns[0] != "8.8.8.8" || ns[1] != "1.1.1.1" {
		t.Errorf("dns_nameservers: got %v", ns)
	}

	ap := body["allocation_pools"].([]map[string]string)
	if len(ap) != 1 || ap[0]["start"] != "10.0.1.10" || ap[0]["end"] != "10.0.1.100" {
		t.Errorf("allocation_pools: got %v", ap)
	}

	hr := body["host_routes"].([]map[string]string)
	if len(hr) != 1 || hr[0]["destination"] != "192.168.0.0/16" || hr[0]["nexthop"] != "10.0.1.1" {
		t.Errorf("host_routes: got %v", hr)
	}
}

func TestBuildBody_IPv6(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}
	plan := nullPlan("ipv6-subnet", "fd00::/64", 6)

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}
	if body["ip_version"] != int64(6) {
		t.Errorf("ip_version: got %v, want 6", body["ip_version"])
	}
}

// ---------------------------------------------------------------------------
// Update immutable field exclusion
// ---------------------------------------------------------------------------

func TestUpdateBody_ImmutableFieldsExcluded(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	plan := nullPlan("update-subnet", "10.0.0.0/24", 4)
	plan.VPCID = types.StringValue("vpc-orig")
	plan.Description = types.StringValue("updated desc")
	plan.EnableDHCP = types.BoolValue(false)

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}

	// Simulate Update: delete immutable fields
	delete(body, "vpc_id")
	delete(body, "ip_version")
	delete(body, "cidr")

	if _, exists := body["vpc_id"]; exists {
		t.Error("vpc_id should be excluded from update body")
	}
	if _, exists := body["ip_version"]; exists {
		t.Error("ip_version should be excluded from update body")
	}
	if _, exists := body["cidr"]; exists {
		t.Error("cidr should be excluded from update body")
	}

	// Mutable fields should remain.
	if body["name"] != "update-subnet" {
		t.Errorf("name should remain, got %v", body["name"])
	}
	if body["description"] != "updated desc" {
		t.Errorf("description should remain, got %v", body["description"])
	}
	if body["enable_dhcp"] != false {
		t.Errorf("enable_dhcp should remain, got %v", body["enable_dhcp"])
	}
}

func TestUpdateBody_MutableFieldsPreserved(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	dnsList, _ := types.ListValueFrom(ctx, types.StringType, []string{"1.1.1.1"})
	poolsList, _ := types.ListValueFrom(ctx, allocationPoolObjectType, []allocationPoolModel{
		{Start: types.StringValue("10.0.0.50"), End: types.StringValue("10.0.0.150")},
	})
	routesList, _ := types.ListValueFrom(ctx, hostRouteObjectType, []hostRouteModel{
		{Destination: types.StringValue("0.0.0.0/0"), Nexthop: types.StringValue("10.0.0.1")},
	})

	plan := nullPlan("update-all", "10.0.0.0/24", 4)
	plan.GatewayIP = types.StringValue("10.0.0.1")
	plan.DNSNameservers = dnsList
	plan.AllocationPools = poolsList
	plan.HostRoutes = routesList

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}

	// Simulate Update: delete immutable fields
	delete(body, "vpc_id")
	delete(body, "ip_version")
	delete(body, "cidr")

	// All mutable fields must survive.
	if body["name"] != "update-all" {
		t.Errorf("name missing")
	}
	if body["gateway_ip"] != "10.0.0.1" {
		t.Errorf("gateway_ip missing")
	}
	if _, ok := body["dns_nameservers"]; !ok {
		t.Error("dns_nameservers missing")
	}
	if _, ok := body["allocation_pools"]; !ok {
		t.Error("allocation_pools missing")
	}
	if _, ok := body["host_routes"]; !ok {
		t.Error("host_routes missing")
	}
}

// ---------------------------------------------------------------------------
// Create response parsing — extracting ID
// ---------------------------------------------------------------------------

func TestCreateResponse_ExtractID(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expectID string
		wantErr  bool
	}{
		{
			"standard response",
			`{"id":"subnet-abc-123","name":"my-subnet","cidr":"10.0.0.0/24"}`,
			"subnet-abc-123",
			false,
		},
		{
			"uuid style",
			`{"id":"d290f1ee-6c54-4b01-90e6-d701748f0851","name":"sub"}`,
			"d290f1ee-6c54-4b01-90e6-d701748f0851",
			false,
		},
		{
			"missing id",
			`{"name":"no-id-subnet"}`,
			"",
			true,
		},
		{
			"id is number not string",
			`{"id":12345}`,
			"",
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(tc.json), &result); err != nil {
				t.Fatalf("json unmarshal error: %v", err)
			}

			id, ok := result["id"].(string)
			if tc.wantErr {
				if ok && id != "" {
					t.Errorf("expected no valid string ID, got %q", id)
				}
			} else {
				if !ok || id != tc.expectID {
					t.Errorf("expected ID=%q, got ok=%v id=%q", tc.expectID, ok, id)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Read response parsing
// ---------------------------------------------------------------------------

func TestReadResponse_FullPayload(t *testing.T) {
	ctx := context.Background()
	raw := `{
		"name":"sub1",
		"cidr":"10.0.0.0/24",
		"vpc_id":"vpc-1",
		"ip_version":4,
		"description":"test",
		"enable_dhcp":true,
		"gateway_ip":"10.0.0.1",
		"dns_nameservers":["8.8.8.8"],
		"allocation_pools":[{"start":"10.0.0.10","end":"10.0.0.200"}],
		"host_routes":[{"destination":"0.0.0.0/0","nexthop":"10.0.0.1"}]
	}`

	state := parseReadResponse(t, ctx, raw)

	if state.Name.ValueString() != "sub1" {
		t.Errorf("Name: got %s", state.Name.ValueString())
	}
	if state.CIDR.ValueString() != "10.0.0.0/24" {
		t.Errorf("CIDR: got %s", state.CIDR.ValueString())
	}
	if state.VPCID.ValueString() != "vpc-1" {
		t.Errorf("VPCID: got %s", state.VPCID.ValueString())
	}
	if state.IPVersion.ValueInt64() != 4 {
		t.Errorf("IPVersion: got %d", state.IPVersion.ValueInt64())
	}
	if state.Description.ValueString() != "test" {
		t.Errorf("Description: got %s", state.Description.ValueString())
	}
	if state.EnableDHCP.ValueBool() != true {
		t.Error("EnableDHCP: expected true")
	}
	if state.GatewayIP.ValueString() != "10.0.0.1" {
		t.Errorf("GatewayIP: got %s", state.GatewayIP.ValueString())
	}

	// DNS
	var ns []string
	state.DNSNameservers.ElementsAs(ctx, &ns, false)
	if len(ns) != 1 || ns[0] != "8.8.8.8" {
		t.Errorf("DNSNameservers: got %v", ns)
	}

	// Allocation pools
	var pools []allocationPoolModel
	state.AllocationPools.ElementsAs(ctx, &pools, false)
	if len(pools) != 1 || pools[0].Start.ValueString() != "10.0.0.10" || pools[0].End.ValueString() != "10.0.0.200" {
		t.Errorf("AllocationPools: got %+v", pools)
	}

	// Host routes
	var routes []hostRouteModel
	state.HostRoutes.ElementsAs(ctx, &routes, false)
	if len(routes) != 1 || routes[0].Destination.ValueString() != "0.0.0.0/0" || routes[0].Nexthop.ValueString() != "10.0.0.1" {
		t.Errorf("HostRoutes: got %+v", routes)
	}
}

func TestReadResponse_EmptyDescription(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"description":""}`

	state := parseReadResponse(t, ctx, raw)

	if !state.Description.IsNull() {
		t.Errorf("empty description should be null, got %q", state.Description.ValueString())
	}
}

func TestReadResponse_MissingGatewayIP(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4}`

	state := parseReadResponse(t, ctx, raw)

	if !state.GatewayIP.IsNull() {
		t.Errorf("missing gateway_ip should be null, got %q", state.GatewayIP.ValueString())
	}
}

func TestReadResponse_EmptyGatewayIP(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"gateway_ip":""}`

	state := parseReadResponse(t, ctx, raw)

	if !state.GatewayIP.IsNull() {
		t.Errorf("empty gateway_ip should be null, got %q", state.GatewayIP.ValueString())
	}
}

func TestReadResponse_EmptyDNSNameservers(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"dns_nameservers":[]}`

	state := parseReadResponse(t, ctx, raw)

	if !state.DNSNameservers.IsNull() {
		t.Errorf("empty dns_nameservers should be null, got %v", state.DNSNameservers)
	}
}

func TestReadResponse_MissingDNSNameservers(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4}`

	state := parseReadResponse(t, ctx, raw)

	if !state.DNSNameservers.IsNull() {
		t.Errorf("missing dns_nameservers should be null")
	}
}

func TestReadResponse_NullDNSNameservers(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"dns_nameservers":null}`

	state := parseReadResponse(t, ctx, raw)

	if !state.DNSNameservers.IsNull() {
		t.Errorf("null dns_nameservers should be null")
	}
}

func TestReadResponse_NetworkIDFallback(t *testing.T) {
	ctx := context.Background()
	// No vpc_id, but network_id is present — should be used as fallback.
	raw := `{"name":"sub","cidr":"10.0.0.0/24","network_id":"net-fallback-123","ip_version":4}`

	state := parseReadResponse(t, ctx, raw)

	if state.VPCID.ValueString() != "net-fallback-123" {
		t.Errorf("VPCID from network_id fallback: got %q, want net-fallback-123", state.VPCID.ValueString())
	}
}

func TestReadResponse_VPCIDPrecedenceOverNetworkID(t *testing.T) {
	ctx := context.Background()
	// Both vpc_id and network_id present — vpc_id wins.
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"vpc-primary","network_id":"net-secondary","ip_version":4}`

	state := parseReadResponse(t, ctx, raw)

	if state.VPCID.ValueString() != "vpc-primary" {
		t.Errorf("VPCID should prefer vpc_id over network_id: got %q", state.VPCID.ValueString())
	}
}

func TestReadResponse_EnableDHCPFalse(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"enable_dhcp":false}`

	state := parseReadResponse(t, ctx, raw)

	if state.EnableDHCP.IsNull() {
		t.Error("enable_dhcp=false should not be null")
	}
	if state.EnableDHCP.ValueBool() != false {
		t.Error("enable_dhcp should be false")
	}
}

func TestReadResponse_MissingEnableDHCP(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4}`

	state := parseReadResponse(t, ctx, raw)

	if !state.EnableDHCP.IsNull() {
		t.Errorf("missing enable_dhcp should be null, got %v", state.EnableDHCP.ValueBool())
	}
}

func TestReadResponse_MultipleDNS(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"dns_nameservers":["8.8.8.8","1.1.1.1","9.9.9.9"]}`

	state := parseReadResponse(t, ctx, raw)

	var ns []string
	state.DNSNameservers.ElementsAs(ctx, &ns, false)
	if len(ns) != 3 {
		t.Fatalf("expected 3 DNS nameservers, got %d", len(ns))
	}
	expected := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}
	for i, v := range expected {
		if ns[i] != v {
			t.Errorf("dns[%d]: got %s, want %s", i, ns[i], v)
		}
	}
}

func TestReadResponse_MultipleAllocationPools(t *testing.T) {
	ctx := context.Background()
	raw := `{
		"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,
		"allocation_pools":[
			{"start":"10.0.0.10","end":"10.0.0.50"},
			{"start":"10.0.0.100","end":"10.0.0.200"}
		]
	}`

	state := parseReadResponse(t, ctx, raw)

	var pools []allocationPoolModel
	state.AllocationPools.ElementsAs(ctx, &pools, false)
	if len(pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(pools))
	}
	if pools[0].Start.ValueString() != "10.0.0.10" || pools[0].End.ValueString() != "10.0.0.50" {
		t.Errorf("pool[0]: got %+v", pools[0])
	}
	if pools[1].Start.ValueString() != "10.0.0.100" || pools[1].End.ValueString() != "10.0.0.200" {
		t.Errorf("pool[1]: got %+v", pools[1])
	}
}

func TestReadResponse_MultipleHostRoutes(t *testing.T) {
	ctx := context.Background()
	raw := `{
		"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,
		"host_routes":[
			{"destination":"192.168.0.0/16","nexthop":"10.0.0.1"},
			{"destination":"172.16.0.0/12","nexthop":"10.0.0.2"}
		]
	}`

	state := parseReadResponse(t, ctx, raw)

	var routes []hostRouteModel
	state.HostRoutes.ElementsAs(ctx, &routes, false)
	if len(routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(routes))
	}
	if routes[0].Destination.ValueString() != "192.168.0.0/16" || routes[0].Nexthop.ValueString() != "10.0.0.1" {
		t.Errorf("route[0]: got %+v", routes[0])
	}
	if routes[1].Destination.ValueString() != "172.16.0.0/12" || routes[1].Nexthop.ValueString() != "10.0.0.2" {
		t.Errorf("route[1]: got %+v", routes[1])
	}
}

func TestReadResponse_EmptyAllocationPools(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"allocation_pools":[]}`

	state := parseReadResponse(t, ctx, raw)

	// Empty allocation_pools should keep the null list (no pools parsed).
	if !state.AllocationPools.IsNull() {
		t.Errorf("empty allocation_pools should remain null, got %v", state.AllocationPools)
	}
}

func TestReadResponse_EmptyHostRoutes(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"sub","cidr":"10.0.0.0/24","vpc_id":"v1","ip_version":4,"host_routes":[]}`

	state := parseReadResponse(t, ctx, raw)

	if !state.HostRoutes.IsNull() {
		t.Errorf("empty host_routes should remain null, got %v", state.HostRoutes)
	}
}

func TestReadResponse_IPVersionFloat64(t *testing.T) {
	ctx := context.Background()
	// JSON numbers are always float64 — verify int64 conversion.
	raw := `{"name":"sub","cidr":"fd00::/64","vpc_id":"v1","ip_version":6}`

	state := parseReadResponse(t, ctx, raw)

	if state.IPVersion.ValueInt64() != 6 {
		t.Errorf("IPVersion: got %d, want 6", state.IPVersion.ValueInt64())
	}
}

func TestReadResponse_MinimalPayload(t *testing.T) {
	ctx := context.Background()
	raw := `{"name":"min","cidr":"10.0.0.0/24","ip_version":4}`

	state := parseReadResponse(t, ctx, raw)

	if state.Name.ValueString() != "min" {
		t.Errorf("Name: got %s", state.Name.ValueString())
	}
	if state.CIDR.ValueString() != "10.0.0.0/24" {
		t.Errorf("CIDR: got %s", state.CIDR.ValueString())
	}
	if state.IPVersion.ValueInt64() != 4 {
		t.Errorf("IPVersion: got %d", state.IPVersion.ValueInt64())
	}
	// All optionals should be null.
	if !state.Description.IsNull() {
		t.Error("Description should be null")
	}
	if !state.EnableDHCP.IsNull() {
		t.Error("EnableDHCP should be null")
	}
	if !state.GatewayIP.IsNull() {
		t.Error("GatewayIP should be null")
	}
	if !state.DNSNameservers.IsNull() {
		t.Error("DNSNameservers should be null")
	}
	if !state.AllocationPools.IsNull() {
		t.Error("AllocationPools should be null")
	}
	if !state.HostRoutes.IsNull() {
		t.Error("HostRoutes should be null")
	}
}

// ---------------------------------------------------------------------------
// buildBody — unknown fields should be excluded (same as null)
// ---------------------------------------------------------------------------

func TestBuildBody_UnknownFieldsExcluded(t *testing.T) {
	ctx := context.Background()
	r := &subnetResource{}

	plan := &subnetModel{
		Name:            types.StringValue("unk-subnet"),
		CIDR:            types.StringValue("10.0.0.0/24"),
		IPVersion:       types.Int64Value(4),
		VPCID:           types.StringUnknown(),
		Description:     types.StringUnknown(),
		EnableDHCP:      types.BoolUnknown(),
		GatewayIP:       types.StringUnknown(),
		DNSNameservers:  types.ListUnknown(types.StringType),
		AllocationPools: types.ListUnknown(allocationPoolObjectType),
		HostRoutes:      types.ListUnknown(hostRouteObjectType),
	}

	body, err := r.buildBody(ctx, plan)
	if err != nil {
		t.Fatalf("buildBody error: %v", err)
	}

	for _, key := range []string{"vpc_id", "description", "enable_dhcp", "gateway_ip", "dns_nameservers", "allocation_pools", "host_routes"} {
		if _, exists := body[key]; exists {
			t.Errorf("unknown field %q should be absent from body", key)
		}
	}
}
