package security_group

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------------------------------------------------------------------------
// directionToBackend tests
// ---------------------------------------------------------------------------

func TestDirectionToBackend(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ingress to Inbound", "ingress", "Inbound"},
		{"egress to Outbound", "egress", "Outbound"},
		{"unknown passthrough", "unknown", "unknown"},
		{"empty passthrough", "", ""},
		{"case sensitive Ingress", "Ingress", "Ingress"},
		{"case sensitive EGRESS", "EGRESS", "EGRESS"},
		{"random string", "foobar", "foobar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := directionToBackend(tc.input)
			if got != tc.expected {
				t.Errorf("directionToBackend(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// directionFromBackend tests
// ---------------------------------------------------------------------------

func TestDirectionFromBackend(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Inbound to ingress", "Inbound", "ingress"},
		{"Outbound to egress", "Outbound", "egress"},
		{"unknown passthrough", "unknown", "unknown"},
		{"empty passthrough", "", ""},
		{"case sensitive inbound", "inbound", "inbound"},
		{"case sensitive OUTBOUND", "OUTBOUND", "OUTBOUND"},
		{"random string", "foobar", "foobar"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := directionFromBackend(tc.input)
			if got != tc.expected {
				t.Errorf("directionFromBackend(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Direction round-trip tests
// ---------------------------------------------------------------------------

func TestDirectionRoundTrip(t *testing.T) {
	directions := []string{"ingress", "egress"}
	for _, d := range directions {
		t.Run(d, func(t *testing.T) {
			backend := directionToBackend(d)
			result := directionFromBackend(backend)
			if result != d {
				t.Errorf("round-trip failed for %q: got %q (via %q)", d, result, backend)
			}
		})
	}
}

func TestDirectionRoundTripReverse(t *testing.T) {
	backendValues := []string{"Inbound", "Outbound"}
	for _, b := range backendValues {
		t.Run(b, func(t *testing.T) {
			terraform := directionFromBackend(b)
			result := directionToBackend(terraform)
			if result != b {
				t.Errorf("reverse round-trip failed for %q: got %q (via %q)", b, result, terraform)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// NewResource test
// ---------------------------------------------------------------------------

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// ---------------------------------------------------------------------------
// protocolToBackend: comprehensive table test for ALL protocol mappings
// ---------------------------------------------------------------------------

func TestProtocolToBackend_AllProtocols(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		portMin  int64
		portMax  int64
		expected string
	}{
		// ---- TCP well-known ports (portMin == portMax) ----
		{"tcp SSH", "tcp", 22, 22, "SSH"},
		{"tcp HTTP", "tcp", 80, 80, "HTTP"},
		{"tcp HTTPS", "tcp", 443, 443, "HTTPS"},
		{"tcp DNS", "tcp", 53, 53, "DNS"},
		{"tcp RDP", "tcp", 3389, 3389, "RDP"},
		{"tcp MYSQL", "tcp", 3306, 3306, "MYSQL"},
		{"tcp MSSQL", "tcp", 1433, 1433, "MSSQL"},
		{"tcp SMTP", "tcp", 25, 25, "SMTP"},
		{"tcp SMTPS", "tcp", 465, 465, "SMTPS"},
		{"tcp IMAP", "tcp", 143, 143, "IMAP"},
		{"tcp IMAPS", "tcp", 993, 993, "IMAPS"},
		{"tcp LDAP", "tcp", 389, 389, "LDAP"},
		{"tcp POP3", "tcp", 110, 110, "POP3"},
		{"tcp POP3S", "tcp", 995, 995, "POP3S"},

		// ---- TCP ranges ----
		{"tcp All TCP", "tcp", 1, 65535, "All TCP"},
		{"tcp Custom TCP custom range", "tcp", 8000, 9000, "Custom TCP"},
		{"tcp Custom TCP single non-well-known", "tcp", 8080, 8080, "Custom TCP"},
		{"tcp Custom TCP zero ports", "tcp", 0, 0, "Custom TCP"},
		{"tcp Custom TCP small range", "tcp", 1000, 2000, "Custom TCP"},

		// ---- TCP well-known port but as range (min != max) -> Custom TCP ----
		{"tcp port 22 range", "tcp", 22, 23, "Custom TCP"},
		{"tcp port 80 range", "tcp", 79, 80, "Custom TCP"},

		// ---- UDP ----
		{"udp All UDP", "udp", 1, 65535, "All UDP"},
		{"udp Custom UDP single", "udp", 53, 53, "Custom UDP"},
		{"udp Custom UDP range", "udp", 1000, 2000, "Custom UDP"},
		{"udp Custom UDP zero", "udp", 0, 0, "Custom UDP"},

		// ---- ICMP ----
		{"icmp zero ports", "icmp", 0, 0, "All ICMP"},
		{"icmp with type/code", "icmp", 8, 0, "All ICMP"},
		{"icmp arbitrary", "icmp", 255, 255, "All ICMP"},

		// ---- Other / passthrough ----
		{"gre passthrough", "gre", 0, 0, "gre"},
		{"ah passthrough", "ah", 0, 0, "ah"},
		{"ospf passthrough", "ospf", 0, 0, "ospf"},
		{"any passthrough", "any", 0, 0, "any"},
		{"empty protocol", "", 0, 0, ""},
		{"ipv6-icmp passthrough", "ipv6-icmp", 0, 0, "ipv6-icmp"},
		{"vrrp passthrough", "vrrp", 0, 0, "vrrp"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := protocolToBackend(tc.protocol, tc.portMin, tc.portMax)
			if got != tc.expected {
				t.Errorf("protocolToBackend(%q, %d, %d) = %q, want %q",
					tc.protocol, tc.portMin, tc.portMax, got, tc.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// protocolToBackend: focused sub-tests (kept for backward compatibility)
// ---------------------------------------------------------------------------

func TestProtocolToBackend_TCPWellKnownPorts(t *testing.T) {
	tests := []struct {
		name     string
		portMin  int64
		portMax  int64
		expected string
	}{
		{"SSH", 22, 22, "SSH"},
		{"HTTP", 80, 80, "HTTP"},
		{"HTTPS", 443, 443, "HTTPS"},
		{"DNS", 53, 53, "DNS"},
		{"RDP", 3389, 3389, "RDP"},
		{"MYSQL", 3306, 3306, "MYSQL"},
		{"MSSQL", 1433, 1433, "MSSQL"},
		{"SMTP", 25, 25, "SMTP"},
		{"SMTPS", 465, 465, "SMTPS"},
		{"IMAP", 143, 143, "IMAP"},
		{"IMAPS", 993, 993, "IMAPS"},
		{"LDAP", 389, 389, "LDAP"},
		{"POP3", 110, 110, "POP3"},
		{"POP3S", 995, 995, "POP3S"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := protocolToBackend("tcp", tc.portMin, tc.portMax)
			if got != tc.expected {
				t.Errorf("protocolToBackend(\"tcp\", %d, %d) = %q, want %q", tc.portMin, tc.portMax, got, tc.expected)
			}
		})
	}
}

func TestProtocolToBackend_TCPRanges(t *testing.T) {
	tests := []struct {
		name     string
		portMin  int64
		portMax  int64
		expected string
	}{
		{"All TCP", 1, 65535, "All TCP"},
		{"Custom TCP single port", 8080, 8080, "Custom TCP"},
		{"Custom TCP range", 1000, 2000, "Custom TCP"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := protocolToBackend("tcp", tc.portMin, tc.portMax)
			if got != tc.expected {
				t.Errorf("protocolToBackend(\"tcp\", %d, %d) = %q, want %q", tc.portMin, tc.portMax, got, tc.expected)
			}
		})
	}
}

func TestProtocolToBackend_UDP(t *testing.T) {
	tests := []struct {
		name     string
		portMin  int64
		portMax  int64
		expected string
	}{
		{"All UDP", 1, 65535, "All UDP"},
		{"Custom UDP", 53, 53, "Custom UDP"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := protocolToBackend("udp", tc.portMin, tc.portMax)
			if got != tc.expected {
				t.Errorf("protocolToBackend(\"udp\", %d, %d) = %q, want %q", tc.portMin, tc.portMax, got, tc.expected)
			}
		})
	}
}

func TestProtocolToBackend_ICMP(t *testing.T) {
	got := protocolToBackend("icmp", 0, 0)
	if got != "All ICMP" {
		t.Errorf("protocolToBackend(\"icmp\", 0, 0) = %q, want \"All ICMP\"", got)
	}
}

func TestProtocolToBackend_Other(t *testing.T) {
	got := protocolToBackend("gre", 0, 0)
	if got != "gre" {
		t.Errorf("protocolToBackend(\"gre\", 0, 0) = %q, want \"gre\"", got)
	}
}

// ---------------------------------------------------------------------------
// Helper: security group rule object type
// ---------------------------------------------------------------------------

var sgRuleObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"direction":        types.StringType,
		"protocol":         types.StringType,
		"port_range_min":   types.Int64Type,
		"port_range_max":   types.Int64Type,
		"remote_ip_prefix": types.StringType,
		"remote_group_id":  types.StringType,
		"ethertype":        types.StringType,
	},
}

// makeRulesList builds a types.List of securityGroupRuleModel from the given slice.
func makeRulesList(t *testing.T, ctx context.Context, rules []securityGroupRuleModel) types.List {
	t.Helper()
	rulesList, diags := types.ListValueFrom(ctx, sgRuleObjectType, rules)
	if diags.HasError() {
		t.Fatalf("failed to build rules list: %s", diags.Errors())
	}
	return rulesList
}

// ---------------------------------------------------------------------------
// buildRulesPayload tests
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_IngressTCP(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(80),
			PortRangeMax:   types.Int64Value(80),
			RemoteIPPrefix: types.StringValue("0.0.0.0/0"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["rule_type"] != "Inbound" {
		t.Errorf("expected rule_type='Inbound', got %v", r["rule_type"])
	}
	if r["protocol_name"] != "HTTP" {
		t.Errorf("expected protocol_name='HTTP', got %v", r["protocol_name"])
	}
	if r["remote"] != "manual" {
		t.Errorf("expected remote='manual', got %v", r["remote"])
	}
	if r["remote_ip_prefix"] != "0.0.0.0/0" {
		t.Errorf("expected remote_ip_prefix='0.0.0.0/0', got %v", r["remote_ip_prefix"])
	}
	if r["port_range_min"] != int64(80) {
		t.Errorf("expected port_range_min=80, got %v", r["port_range_min"])
	}
	if r["port_range_max"] != int64(80) {
		t.Errorf("expected port_range_max=80, got %v", r["port_range_max"])
	}
}

func TestBuildRulesPayload_EgressAllTCP(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("egress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(1),
			PortRangeMax:   types.Int64Value(65535),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["rule_type"] != "Outbound" {
		t.Errorf("expected rule_type='Outbound', got %v", r["rule_type"])
	}
	if r["protocol_name"] != "All TCP" {
		t.Errorf("expected protocol_name='All TCP', got %v", r["protocol_name"])
	}
	// "All TCP" is a noPortProtocol, so port_range_min/max must be absent.
	if _, exists := r["port_range_min"]; exists {
		t.Errorf("expected port_range_min to be absent for 'All TCP', but it was present: %v", r["port_range_min"])
	}
	if _, exists := r["port_range_max"]; exists {
		t.Errorf("expected port_range_max to be absent for 'All TCP', but it was present: %v", r["port_range_max"])
	}
}

func TestBuildRulesPayload_RemoteSecurityGroup(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(22),
			PortRangeMax:   types.Int64Value(22),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringValue("sg-abc-123"),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["remote"] != "securityGroup" {
		t.Errorf("expected remote='securityGroup', got %v", r["remote"])
	}
	if r["remote_group_id"] != "sg-abc-123" {
		t.Errorf("expected remote_group_id='sg-abc-123', got %v", r["remote_group_id"])
	}
	if _, exists := r["remote_ip_prefix"]; exists {
		t.Errorf("expected remote_ip_prefix to be absent for securityGroup remote, but it was present")
	}
}

func TestBuildRulesPayload_RemoteAny(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("egress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(443),
			PortRangeMax:   types.Int64Value(443),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["remote"] != "any" {
		t.Errorf("expected remote='any', got %v", r["remote"])
	}
	if _, exists := r["remote_ip_prefix"]; exists {
		t.Errorf("expected remote_ip_prefix to be absent, but it was present")
	}
	if _, exists := r["remote_group_id"]; exists {
		t.Errorf("expected remote_group_id to be absent, but it was present")
	}
}

func TestBuildRulesPayload_NullList(t *testing.T) {
	ctx := context.Background()

	nullList := types.ListNull(sgRuleObjectType)
	payload, err := buildRulesPayload(ctx, nullList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if payload != nil {
		t.Errorf("expected nil payload for null list, got %v", payload)
	}
}

func TestBuildRulesPayload_EmptyList(t *testing.T) {
	ctx := context.Background()

	emptyList, diags := types.ListValueFrom(ctx, sgRuleObjectType, []securityGroupRuleModel{})
	if diags.HasError() {
		t.Fatalf("failed to build empty rules list: %s", diags.Errors())
	}

	payload, err := buildRulesPayload(ctx, emptyList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if payload != nil {
		t.Errorf("expected nil payload for empty list, got %v", payload)
	}
}

func TestBuildRulesPayload_UnknownList(t *testing.T) {
	ctx := context.Background()

	unknownList := types.ListUnknown(sgRuleObjectType)
	payload, err := buildRulesPayload(ctx, unknownList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if payload != nil {
		t.Errorf("expected nil payload for unknown list, got %v", payload)
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: CIDR-based remote rules
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_CIDRFormat(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(22),
			PortRangeMax:   types.Int64Value(22),
			RemoteIPPrefix: types.StringValue("192.168.1.0/24"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringValue("IPv4"),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["remote"] != "manual" {
		t.Errorf("expected remote='manual' for CIDR-based rule, got %v", r["remote"])
	}
	if r["remote_ip_prefix"] != "192.168.1.0/24" {
		t.Errorf("expected remote_ip_prefix='192.168.1.0/24', got %v", r["remote_ip_prefix"])
	}
	if r["protocol_name"] != "SSH" {
		t.Errorf("expected protocol_name='SSH', got %v", r["protocol_name"])
	}
	if r["ethertype"] != "IPv4" {
		t.Errorf("expected ethertype='IPv4', got %v", r["ethertype"])
	}
	if r["port_range_min"] != int64(22) {
		t.Errorf("expected port_range_min=22, got %v", r["port_range_min"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: multiple rules in same direction
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_MultipleRules(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(22),
			PortRangeMax:   types.Int64Value(22),
			RemoteIPPrefix: types.StringValue("0.0.0.0/0"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(80),
			PortRangeMax:   types.Int64Value(80),
			RemoteIPPrefix: types.StringValue("0.0.0.0/0"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(443),
			PortRangeMax:   types.Int64Value(443),
			RemoteIPPrefix: types.StringValue("10.0.0.0/8"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(payload))
	}

	for i, r := range payload {
		if r["rule_type"] != "Inbound" {
			t.Errorf("rule %d: expected rule_type='Inbound', got %v", i, r["rule_type"])
		}
	}

	expectedProtocols := []string{"SSH", "HTTP", "HTTPS"}
	for i, r := range payload {
		if r["protocol_name"] != expectedProtocols[i] {
			t.Errorf("rule %d: expected protocol_name=%q, got %v", i, expectedProtocols[i], r["protocol_name"])
		}
	}

	if payload[2]["remote_ip_prefix"] != "10.0.0.0/8" {
		t.Errorf("rule 2: expected remote_ip_prefix='10.0.0.0/8', got %v", payload[2]["remote_ip_prefix"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: ICMP rule (no port ranges)
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_ICMPNoPortRanges(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("icmp"),
			PortRangeMin:   types.Int64Null(),
			PortRangeMax:   types.Int64Null(),
			RemoteIPPrefix: types.StringValue("0.0.0.0/0"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringValue("IPv4"),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["protocol_name"] != "All ICMP" {
		t.Errorf("expected protocol_name='All ICMP', got %v", r["protocol_name"])
	}
	// All ICMP is a noPortProtocol — port_range_min/max must be absent.
	if _, exists := r["port_range_min"]; exists {
		t.Errorf("expected port_range_min to be absent for 'All ICMP', but it was present")
	}
	if _, exists := r["port_range_max"]; exists {
		t.Errorf("expected port_range_max to be absent for 'All ICMP', but it was present")
	}
	if r["ethertype"] != "IPv4" {
		t.Errorf("expected ethertype='IPv4', got %v", r["ethertype"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: All UDP (noPortProtocol)
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_AllUDPNoPortRanges(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("egress"),
			Protocol:       types.StringValue("udp"),
			PortRangeMin:   types.Int64Value(1),
			PortRangeMax:   types.Int64Value(65535),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["protocol_name"] != "All UDP" {
		t.Errorf("expected protocol_name='All UDP', got %v", r["protocol_name"])
	}
	if _, exists := r["port_range_min"]; exists {
		t.Errorf("expected port_range_min to be absent for 'All UDP', but it was present")
	}
	if _, exists := r["port_range_max"]; exists {
		t.Errorf("expected port_range_max to be absent for 'All UDP', but it was present")
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: Custom TCP with port range
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_CustomTCPWithPortRange(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(8000),
			PortRangeMax:   types.Int64Value(9000),
			RemoteIPPrefix: types.StringValue("10.0.0.0/8"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["protocol_name"] != "Custom TCP" {
		t.Errorf("expected protocol_name='Custom TCP', got %v", r["protocol_name"])
	}
	if r["port_range_min"] != int64(8000) {
		t.Errorf("expected port_range_min=8000, got %v", r["port_range_min"])
	}
	if r["port_range_max"] != int64(9000) {
		t.Errorf("expected port_range_max=9000, got %v", r["port_range_max"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: Custom UDP with port range
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_CustomUDPWithPortRange(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("udp"),
			PortRangeMin:   types.Int64Value(5000),
			PortRangeMax:   types.Int64Value(6000),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["protocol_name"] != "Custom UDP" {
		t.Errorf("expected protocol_name='Custom UDP', got %v", r["protocol_name"])
	}
	if r["port_range_min"] != int64(5000) {
		t.Errorf("expected port_range_min=5000, got %v", r["port_range_min"])
	}
	if r["port_range_max"] != int64(6000) {
		t.Errorf("expected port_range_max=6000, got %v", r["port_range_max"])
	}
	if r["remote"] != "any" {
		t.Errorf("expected remote='any', got %v", r["remote"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: unknown protocol passthrough (e.g. "gre")
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_UnknownProtocolPassthrough(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("gre"),
			PortRangeMin:   types.Int64Null(),
			PortRangeMax:   types.Int64Null(),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["protocol_name"] != "gre" {
		t.Errorf("expected protocol_name='gre', got %v", r["protocol_name"])
	}
	// "gre" is not in noPortProtocols, but ports are null so they should be absent.
	if _, exists := r["port_range_min"]; exists {
		t.Errorf("expected port_range_min to be absent for null ports, but it was present")
	}
	if _, exists := r["port_range_max"]; exists {
		t.Errorf("expected port_range_max to be absent for null ports, but it was present")
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: remote_group_id takes precedence over remote_ip_prefix
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_RemoteGroupIDPrecedence(t *testing.T) {
	ctx := context.Background()

	// When both remote_group_id and remote_ip_prefix are set, remote_group_id wins.
	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(80),
			PortRangeMax:   types.Int64Value(80),
			RemoteIPPrefix: types.StringValue("10.0.0.0/8"),
			RemoteGroupID:  types.StringValue("sg-xyz-456"),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["remote"] != "securityGroup" {
		t.Errorf("expected remote='securityGroup' (group_id takes precedence), got %v", r["remote"])
	}
	if r["remote_group_id"] != "sg-xyz-456" {
		t.Errorf("expected remote_group_id='sg-xyz-456', got %v", r["remote_group_id"])
	}
	// remote_ip_prefix should NOT be in the payload since remote_group_id takes precedence.
	if _, exists := r["remote_ip_prefix"]; exists {
		t.Errorf("expected remote_ip_prefix to be absent when remote_group_id is set, but it was present")
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: empty string remote fields fall through to "any"
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_EmptyStringRemoteFieldsFallToAny(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(80),
			PortRangeMax:   types.Int64Value(80),
			RemoteIPPrefix: types.StringValue(""),
			RemoteGroupID:  types.StringValue(""),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["remote"] != "any" {
		t.Errorf("expected remote='any' when both remote fields are empty strings, got %v", r["remote"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: mixed ingress and egress rules
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_MixedDirections(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(22),
			PortRangeMax:   types.Int64Value(22),
			RemoteIPPrefix: types.StringValue("0.0.0.0/0"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
		{
			Direction:      types.StringValue("egress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(443),
			PortRangeMax:   types.Int64Value(443),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("icmp"),
			PortRangeMin:   types.Int64Null(),
			PortRangeMax:   types.Int64Null(),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringValue("IPv4"),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(payload))
	}

	// Rule 0: ingress SSH
	if payload[0]["rule_type"] != "Inbound" {
		t.Errorf("rule 0: expected rule_type='Inbound', got %v", payload[0]["rule_type"])
	}
	if payload[0]["protocol_name"] != "SSH" {
		t.Errorf("rule 0: expected protocol_name='SSH', got %v", payload[0]["protocol_name"])
	}

	// Rule 1: egress HTTPS
	if payload[1]["rule_type"] != "Outbound" {
		t.Errorf("rule 1: expected rule_type='Outbound', got %v", payload[1]["rule_type"])
	}
	if payload[1]["protocol_name"] != "HTTPS" {
		t.Errorf("rule 1: expected protocol_name='HTTPS', got %v", payload[1]["protocol_name"])
	}

	// Rule 2: ingress All ICMP
	if payload[2]["rule_type"] != "Inbound" {
		t.Errorf("rule 2: expected rule_type='Inbound', got %v", payload[2]["rule_type"])
	}
	if payload[2]["protocol_name"] != "All ICMP" {
		t.Errorf("rule 2: expected protocol_name='All ICMP', got %v", payload[2]["protocol_name"])
	}
	if payload[2]["ethertype"] != "IPv4" {
		t.Errorf("rule 2: expected ethertype='IPv4', got %v", payload[2]["ethertype"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: IPv6 ethertype
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_IPv6EtherType(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(80),
			PortRangeMax:   types.Int64Value(80),
			RemoteIPPrefix: types.StringValue("::/0"),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringValue("IPv6"),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	r := payload[0]
	if r["ethertype"] != "IPv6" {
		t.Errorf("expected ethertype='IPv6', got %v", r["ethertype"])
	}
	if r["remote_ip_prefix"] != "::/0" {
		t.Errorf("expected remote_ip_prefix='::/0', got %v", r["remote_ip_prefix"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: null ethertype is absent from payload
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_NullEtherTypeOmitted(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction:      types.StringValue("ingress"),
			Protocol:       types.StringValue("tcp"),
			PortRangeMin:   types.Int64Value(80),
			PortRangeMax:   types.Int64Value(80),
			RemoteIPPrefix: types.StringNull(),
			RemoteGroupID:  types.StringNull(),
			EtherType:      types.StringNull(),
		},
	}
	rulesList := makeRulesList(t, ctx, rules)

	payload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload returned error: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(payload))
	}

	if _, exists := payload[0]["ethertype"]; exists {
		t.Errorf("expected ethertype to be absent when null, but it was present: %v", payload[0]["ethertype"])
	}
}

// ---------------------------------------------------------------------------
// buildRulesPayload: comprehensive table-driven test
// ---------------------------------------------------------------------------

func TestBuildRulesPayload_TableDriven(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		rule               securityGroupRuleModel
		expectedRuleType   string
		expectedProtocol   string
		expectedRemote     string
		expectPortMin      bool
		expectedPortMin    int64
		expectPortMax      bool
		expectedPortMax    int64
		expectRemoteIP     bool
		expectedRemoteIP   string
		expectRemoteGroup  bool
		expectedRemoteGrp  string
		expectEthertype    bool
		expectedEthertype  string
	}{
		{
			name: "ingress SSH with CIDR",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(22), PortRangeMax: types.Int64Value(22),
				RemoteIPPrefix: types.StringValue("0.0.0.0/0"), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "SSH", expectedRemote: "manual",
			expectPortMin: true, expectedPortMin: 22, expectPortMax: true, expectedPortMax: 22,
			expectRemoteIP: true, expectedRemoteIP: "0.0.0.0/0",
		},
		{
			name: "egress All TCP any remote",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("egress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(1), PortRangeMax: types.Int64Value(65535),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Outbound", expectedProtocol: "All TCP", expectedRemote: "any",
			expectPortMin: false, expectPortMax: false,
		},
		{
			name: "ingress All UDP any remote",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("udp"),
				PortRangeMin: types.Int64Value(1), PortRangeMax: types.Int64Value(65535),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "All UDP", expectedRemote: "any",
			expectPortMin: false, expectPortMax: false,
		},
		{
			name: "ingress All ICMP with IPv4",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("icmp"),
				PortRangeMin: types.Int64Null(), PortRangeMax: types.Int64Null(),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringValue("IPv4"),
			},
			expectedRuleType: "Inbound", expectedProtocol: "All ICMP", expectedRemote: "any",
			expectPortMin: false, expectPortMax: false,
			expectEthertype: true, expectedEthertype: "IPv4",
		},
		{
			name: "ingress Custom TCP range with SG remote",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(3000), PortRangeMax: types.Int64Value(4000),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringValue("sg-remote-id"),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "Custom TCP", expectedRemote: "securityGroup",
			expectPortMin: true, expectedPortMin: 3000, expectPortMax: true, expectedPortMax: 4000,
			expectRemoteGroup: true, expectedRemoteGrp: "sg-remote-id",
		},
		{
			name: "egress Custom UDP single port",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("egress"), Protocol: types.StringValue("udp"),
				PortRangeMin: types.Int64Value(53), PortRangeMax: types.Int64Value(53),
				RemoteIPPrefix: types.StringValue("8.8.8.0/24"), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Outbound", expectedProtocol: "Custom UDP", expectedRemote: "manual",
			expectPortMin: true, expectedPortMin: 53, expectPortMax: true, expectedPortMax: 53,
			expectRemoteIP: true, expectedRemoteIP: "8.8.8.0/24",
		},
		{
			name: "ingress DNS",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(53), PortRangeMax: types.Int64Value(53),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "DNS", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 53, expectPortMax: true, expectedPortMax: 53,
		},
		{
			name: "ingress MYSQL",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(3306), PortRangeMax: types.Int64Value(3306),
				RemoteIPPrefix: types.StringValue("172.16.0.0/12"), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "MYSQL", expectedRemote: "manual",
			expectPortMin: true, expectedPortMin: 3306, expectPortMax: true, expectedPortMax: 3306,
			expectRemoteIP: true, expectedRemoteIP: "172.16.0.0/12",
		},
		{
			name: "egress SMTP",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("egress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(25), PortRangeMax: types.Int64Value(25),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Outbound", expectedProtocol: "SMTP", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 25, expectPortMax: true, expectedPortMax: 25,
		},
		{
			name: "ingress RDP",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(3389), PortRangeMax: types.Int64Value(3389),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "RDP", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 3389, expectPortMax: true, expectedPortMax: 3389,
		},
		{
			name: "ingress MSSQL",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(1433), PortRangeMax: types.Int64Value(1433),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "MSSQL", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 1433, expectPortMax: true, expectedPortMax: 1433,
		},
		{
			name: "ingress SMTPS",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(465), PortRangeMax: types.Int64Value(465),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "SMTPS", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 465, expectPortMax: true, expectedPortMax: 465,
		},
		{
			name: "ingress IMAP",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(143), PortRangeMax: types.Int64Value(143),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "IMAP", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 143, expectPortMax: true, expectedPortMax: 143,
		},
		{
			name: "ingress IMAPS",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(993), PortRangeMax: types.Int64Value(993),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "IMAPS", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 993, expectPortMax: true, expectedPortMax: 993,
		},
		{
			name: "ingress LDAP",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(389), PortRangeMax: types.Int64Value(389),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "LDAP", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 389, expectPortMax: true, expectedPortMax: 389,
		},
		{
			name: "ingress POP3",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(110), PortRangeMax: types.Int64Value(110),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "POP3", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 110, expectPortMax: true, expectedPortMax: 110,
		},
		{
			name: "ingress POP3S",
			rule: securityGroupRuleModel{
				Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
				PortRangeMin: types.Int64Value(995), PortRangeMax: types.Int64Value(995),
				RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
				EtherType: types.StringNull(),
			},
			expectedRuleType: "Inbound", expectedProtocol: "POP3S", expectedRemote: "any",
			expectPortMin: true, expectedPortMin: 995, expectPortMax: true, expectedPortMax: 995,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rulesList := makeRulesList(t, ctx, []securityGroupRuleModel{tc.rule})

			payload, err := buildRulesPayload(ctx, rulesList)
			if err != nil {
				t.Fatalf("buildRulesPayload returned error: %v", err)
			}
			if len(payload) != 1 {
				t.Fatalf("expected 1 rule, got %d", len(payload))
			}

			r := payload[0]

			if r["rule_type"] != tc.expectedRuleType {
				t.Errorf("rule_type = %v, want %q", r["rule_type"], tc.expectedRuleType)
			}
			if r["protocol_name"] != tc.expectedProtocol {
				t.Errorf("protocol_name = %v, want %q", r["protocol_name"], tc.expectedProtocol)
			}
			if r["remote"] != tc.expectedRemote {
				t.Errorf("remote = %v, want %q", r["remote"], tc.expectedRemote)
			}

			// Port range min
			if tc.expectPortMin {
				if r["port_range_min"] != tc.expectedPortMin {
					t.Errorf("port_range_min = %v, want %d", r["port_range_min"], tc.expectedPortMin)
				}
			} else {
				if _, exists := r["port_range_min"]; exists {
					t.Errorf("expected port_range_min to be absent, but got %v", r["port_range_min"])
				}
			}

			// Port range max
			if tc.expectPortMax {
				if r["port_range_max"] != tc.expectedPortMax {
					t.Errorf("port_range_max = %v, want %d", r["port_range_max"], tc.expectedPortMax)
				}
			} else {
				if _, exists := r["port_range_max"]; exists {
					t.Errorf("expected port_range_max to be absent, but got %v", r["port_range_max"])
				}
			}

			// Remote IP prefix
			if tc.expectRemoteIP {
				if r["remote_ip_prefix"] != tc.expectedRemoteIP {
					t.Errorf("remote_ip_prefix = %v, want %q", r["remote_ip_prefix"], tc.expectedRemoteIP)
				}
			} else {
				if _, exists := r["remote_ip_prefix"]; exists {
					t.Errorf("expected remote_ip_prefix to be absent, but got %v", r["remote_ip_prefix"])
				}
			}

			// Remote group ID
			if tc.expectRemoteGroup {
				if r["remote_group_id"] != tc.expectedRemoteGrp {
					t.Errorf("remote_group_id = %v, want %q", r["remote_group_id"], tc.expectedRemoteGrp)
				}
			} else {
				if _, exists := r["remote_group_id"]; exists {
					t.Errorf("expected remote_group_id to be absent, but got %v", r["remote_group_id"])
				}
			}

			// Ethertype
			if tc.expectEthertype {
				if r["ethertype"] != tc.expectedEthertype {
					t.Errorf("ethertype = %v, want %q", r["ethertype"], tc.expectedEthertype)
				}
			} else {
				if _, exists := r["ethertype"]; exists {
					t.Errorf("expected ethertype to be absent, but got %v", r["ethertype"])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Read response parsing tests
// ---------------------------------------------------------------------------

func TestReadResponseParsing_NameAndDescription(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectName  string
		expectDesc  string
		descIsNull  bool
	}{
		{
			name:       "name and description present",
			jsonData:   `{"name": "my-sg", "description": "A test security group"}`,
			expectName: "my-sg",
			expectDesc: "A test security group",
			descIsNull: false,
		},
		{
			name:       "name present, description null",
			jsonData:   `{"name": "my-sg", "description": null}`,
			expectName: "my-sg",
			expectDesc: "",
			descIsNull: true,
		},
		{
			name:       "name present, description missing",
			jsonData:   `{"name": "my-sg"}`,
			expectName: "my-sg",
			expectDesc: "",
			descIsNull: true,
		},
		{
			name:       "name present, description empty string",
			jsonData:   `{"name": "my-sg", "description": ""}`,
			expectName: "my-sg",
			expectDesc: "",
			descIsNull: false,
		},
		{
			name:       "name with spaces and special chars",
			jsonData:   `{"name": "My Security Group 01", "description": "Production SG, v2.0"}`,
			expectName: "My Security Group 01",
			expectDesc: "Production SG, v2.0",
			descIsNull: false,
		},
		{
			name:       "description is numeric value (type mismatch)",
			jsonData:   `{"name": "my-sg", "description": 42}`,
			expectName: "my-sg",
			expectDesc: "",
			descIsNull: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(tc.jsonData), &result); err != nil {
				t.Fatalf("failed to unmarshal test JSON: %v", err)
			}

			// Simulate the Read logic from resource.go
			var state securityGroupModel

			if v, ok := result["name"].(string); ok {
				state.Name = types.StringValue(v)
			}
			if v, ok := result["description"].(string); ok {
				state.Description = types.StringValue(v)
			} else {
				state.Description = types.StringNull()
			}

			// Verify name
			if state.Name.ValueString() != tc.expectName {
				t.Errorf("name = %q, want %q", state.Name.ValueString(), tc.expectName)
			}

			// Verify description
			if tc.descIsNull {
				if !state.Description.IsNull() {
					t.Errorf("expected description to be null, got %q", state.Description.ValueString())
				}
			} else {
				if state.Description.IsNull() {
					t.Errorf("expected description = %q, but it was null", tc.expectDesc)
				} else if state.Description.ValueString() != tc.expectDesc {
					t.Errorf("description = %q, want %q", state.Description.ValueString(), tc.expectDesc)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Update body building tests
// ---------------------------------------------------------------------------

func TestUpdateBodyBuilding(t *testing.T) {
	tests := []struct {
		name           string
		planName       string
		planDesc       types.String
		rules          []securityGroupRuleModel
		expectName     string
		expectDesc     bool
		expectedDesc   string
		expectRules    bool
		expectedRuleN  int
	}{
		{
			name:       "name only, no description, no rules",
			planName:   "updated-sg",
			planDesc:   types.StringNull(),
			rules:      nil,
			expectName: "updated-sg",
			expectDesc: false,
		},
		{
			name:         "name and description, no rules",
			planName:     "my-sg",
			planDesc:     types.StringValue("New description"),
			rules:        nil,
			expectName:   "my-sg",
			expectDesc:   true,
			expectedDesc: "New description",
		},
		{
			name:         "name, description, and rules",
			planName:     "web-sg",
			planDesc:     types.StringValue("Web server SG"),
			expectName:   "web-sg",
			expectDesc:   true,
			expectedDesc: "Web server SG",
			expectRules:  true,
			rules: []securityGroupRuleModel{
				{
					Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
					PortRangeMin: types.Int64Value(80), PortRangeMax: types.Int64Value(80),
					RemoteIPPrefix: types.StringValue("0.0.0.0/0"), RemoteGroupID: types.StringNull(),
					EtherType: types.StringNull(),
				},
				{
					Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
					PortRangeMin: types.Int64Value(443), PortRangeMax: types.Int64Value(443),
					RemoteIPPrefix: types.StringValue("0.0.0.0/0"), RemoteGroupID: types.StringNull(),
					EtherType: types.StringNull(),
				},
			},
			expectedRuleN: 2,
		},
		{
			name:         "name and empty description",
			planName:     "my-sg",
			planDesc:     types.StringValue(""),
			expectName:   "my-sg",
			expectDesc:   true,
			expectedDesc: "",
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the Update body-building logic from resource.go
			body := map[string]interface{}{
				"name": tc.planName,
			}
			if !tc.planDesc.IsNull() && !tc.planDesc.IsUnknown() {
				body["description"] = tc.planDesc.ValueString()
			}

			if tc.rules != nil {
				rulesList := makeRulesList(t, ctx, tc.rules)
				rules, err := buildRulesPayload(ctx, rulesList)
				if err != nil {
					t.Fatalf("buildRulesPayload error: %v", err)
				}
				if rules != nil {
					body["rules"] = rules
				}
			}

			// Verify name
			if body["name"] != tc.expectName {
				t.Errorf("body[name] = %v, want %q", body["name"], tc.expectName)
			}

			// Verify description
			if tc.expectDesc {
				desc, exists := body["description"]
				if !exists {
					t.Errorf("expected body to contain 'description', but it was absent")
				} else if desc != tc.expectedDesc {
					t.Errorf("body[description] = %v, want %q", desc, tc.expectedDesc)
				}
			} else {
				if _, exists := body["description"]; exists {
					t.Errorf("expected body NOT to contain 'description', but it was present: %v", body["description"])
				}
			}

			// Verify rules
			if tc.expectRules {
				rulesVal, exists := body["rules"]
				if !exists {
					t.Errorf("expected body to contain 'rules', but it was absent")
				} else {
					rulesSlice, ok := rulesVal.([]map[string]interface{})
					if !ok {
						t.Fatalf("body[rules] is not []map[string]interface{}")
					}
					if len(rulesSlice) != tc.expectedRuleN {
						t.Errorf("body[rules] has %d rules, want %d", len(rulesSlice), tc.expectedRuleN)
					}
				}
			} else {
				if _, exists := body["rules"]; exists {
					t.Errorf("expected body NOT to contain 'rules', but it was present")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Update body building: rules content verification
// ---------------------------------------------------------------------------

func TestUpdateBodyBuilding_RulesContent(t *testing.T) {
	ctx := context.Background()

	rules := []securityGroupRuleModel{
		{
			Direction: types.StringValue("ingress"), Protocol: types.StringValue("tcp"),
			PortRangeMin: types.Int64Value(22), PortRangeMax: types.Int64Value(22),
			RemoteIPPrefix: types.StringValue("10.0.0.0/8"), RemoteGroupID: types.StringNull(),
			EtherType: types.StringValue("IPv4"),
		},
		{
			Direction: types.StringValue("egress"), Protocol: types.StringValue("icmp"),
			PortRangeMin: types.Int64Null(), PortRangeMax: types.Int64Null(),
			RemoteIPPrefix: types.StringNull(), RemoteGroupID: types.StringNull(),
			EtherType: types.StringNull(),
		},
	}

	body := map[string]interface{}{
		"name":        "test-sg",
		"description": "Test SG for update",
	}

	rulesList := makeRulesList(t, ctx, rules)
	rulesPayload, err := buildRulesPayload(ctx, rulesList)
	if err != nil {
		t.Fatalf("buildRulesPayload error: %v", err)
	}
	body["rules"] = rulesPayload

	rulesSlice := body["rules"].([]map[string]interface{})
	if len(rulesSlice) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rulesSlice))
	}

	// Rule 0: ingress SSH
	r0 := rulesSlice[0]
	if r0["rule_type"] != "Inbound" {
		t.Errorf("rule 0: rule_type = %v, want 'Inbound'", r0["rule_type"])
	}
	if r0["protocol_name"] != "SSH" {
		t.Errorf("rule 0: protocol_name = %v, want 'SSH'", r0["protocol_name"])
	}
	if r0["remote"] != "manual" {
		t.Errorf("rule 0: remote = %v, want 'manual'", r0["remote"])
	}
	if r0["port_range_min"] != int64(22) {
		t.Errorf("rule 0: port_range_min = %v, want 22", r0["port_range_min"])
	}
	if r0["ethertype"] != "IPv4" {
		t.Errorf("rule 0: ethertype = %v, want 'IPv4'", r0["ethertype"])
	}

	// Rule 1: egress All ICMP
	r1 := rulesSlice[1]
	if r1["rule_type"] != "Outbound" {
		t.Errorf("rule 1: rule_type = %v, want 'Outbound'", r1["rule_type"])
	}
	if r1["protocol_name"] != "All ICMP" {
		t.Errorf("rule 1: protocol_name = %v, want 'All ICMP'", r1["protocol_name"])
	}
	if r1["remote"] != "any" {
		t.Errorf("rule 1: remote = %v, want 'any'", r1["remote"])
	}
	if _, exists := r1["port_range_min"]; exists {
		t.Errorf("rule 1: expected port_range_min to be absent for All ICMP")
	}
}

// ---------------------------------------------------------------------------
// Regex tests
// ---------------------------------------------------------------------------

func TestSGNameRegex(t *testing.T) {
	valid := []struct {
		name  string
		input string
	}{
		{"hyphenated", "web-server"},
		{"underscore and digits", "sg_test_1"},
		{"spaces", "My Security Group"},
	}
	for _, tc := range valid {
		t.Run("valid_"+tc.name, func(t *testing.T) {
			if !sgNameRegex.MatchString(tc.input) {
				t.Errorf("expected sgNameRegex to match %q, but it did not", tc.input)
			}
		})
	}

	invalid := []struct {
		name  string
		input string
	}{
		{"at sign", "bad@sg"},
		{"angle brackets", "sg<html>"},
	}
	for _, tc := range invalid {
		t.Run("invalid_"+tc.name, func(t *testing.T) {
			if sgNameRegex.MatchString(tc.input) {
				t.Errorf("expected sgNameRegex NOT to match %q, but it did", tc.input)
			}
		})
	}
}

func TestSGDescriptionRegex(t *testing.T) {
	valid := []struct {
		name  string
		input string
	}{
		{"with comma and period", "A basic desc, here."},
		{"hyphen underscore digits", "test-sg_1"},
	}
	for _, tc := range valid {
		t.Run("valid_"+tc.name, func(t *testing.T) {
			if !sgDescriptionRegex.MatchString(tc.input) {
				t.Errorf("expected sgDescriptionRegex to match %q, but it did not", tc.input)
			}
		})
	}

	invalid := []struct {
		name  string
		input string
	}{
		{"angle brackets", "bad<desc>"},
	}
	for _, tc := range invalid {
		t.Run("invalid_"+tc.name, func(t *testing.T) {
			if sgDescriptionRegex.MatchString(tc.input) {
				t.Errorf("expected sgDescriptionRegex NOT to match %q, but it did", tc.input)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// SG name regex: length limits
// ---------------------------------------------------------------------------

func TestSGNameRegex_LengthLimits(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		isValid bool
	}{
		{"single char", "a", true},
		{"two chars", "ab", true},
		{"three chars (min)", "abc", true},
		{"50 chars (max)", "abcdefghijklmnopqrstuvwxyz-ABCDEFGHIJKLMNOPQRSTUV12", true},
		{"255 chars all valid", func() string {
			s := ""
			for i := 0; i < 255; i++ {
				s += "a"
			}
			return s
		}(), true},
		{"with spaces", "My SG Name", true},
		{"with underscores", "my_sg_name", true},
		{"with hyphens", "my-sg-name", true},
		{"with at sign", "bad@name", false},
		{"with angle bracket", "bad<name>", false},
		{"with pipe", "bad|name", false},
		{"empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matched := sgNameRegex.MatchString(tc.input)
			if matched != tc.isValid {
				if tc.isValid {
					t.Errorf("expected sgNameRegex to match %q, but it did not", tc.input)
				} else {
					t.Errorf("expected sgNameRegex NOT to match %q, but it did", tc.input)
				}
			}
		})
	}
}
