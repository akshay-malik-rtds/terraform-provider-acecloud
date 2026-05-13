package security_group

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/security-groups"

// Ensure the resource satisfies the expected interfaces.
var (
	_ resource.Resource              = &securityGroupResource{}
	_ resource.ResourceWithConfigure = &securityGroupResource{}
)

// securityGroupResource is the resource implementation.
type securityGroupResource struct {
	client *client.Client
}

// NewResource returns a new security group resource factory.
func NewResource() resource.Resource {
	return &securityGroupResource{}
}

func (r *securityGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_group"
}

func (r *securityGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

// directionToBackend transforms the Terraform direction value to the backend rule_type.
func directionToBackend(direction string) string {
	switch direction {
	case "ingress":
		return "Inbound"
	case "egress":
		return "Outbound"
	default:
		return direction
	}
}

// protocolToBackend maps user-friendly protocol names to the API's expected protocol names.
// The API accepts: "Custom TCP", "Custom UDP", "Custom ICMP", "All ICMP", "All TCP", "All UDP",
// "Any Protocol", "SSH", "HTTP", "HTTPS", "DNS", "RDP", "MYSQL", "MSSQL", "SMTP", "SMTPS",
// "IMAP", "IMAPS", "LDAP", "POP3", "POP3S", and raw protocol names like "ah", "gre", "ospf", etc.
func protocolToBackend(protocol string, portMin, portMax int64) string {
	switch protocol {
	case "tcp":
		// Map well-known port combinations to named protocols
		if portMin == portMax {
			switch portMin {
			case 22:
				return "SSH"
			case 80:
				return "HTTP"
			case 443:
				return "HTTPS"
			case 53:
				return "DNS"
			case 3389:
				return "RDP"
			case 3306:
				return "MYSQL"
			case 1433:
				return "MSSQL"
			case 25:
				return "SMTP"
			case 465:
				return "SMTPS"
			case 143:
				return "IMAP"
			case 993:
				return "IMAPS"
			case 389:
				return "LDAP"
			case 110:
				return "POP3"
			case 995:
				return "POP3S"
			}
		}
		// Full range = All TCP
		if portMin == 1 && portMax == 65535 {
			return "All TCP"
		}
		return "Custom TCP"
	case "udp":
		if portMin == 1 && portMax == 65535 {
			return "All UDP"
		}
		return "Custom UDP"
	case "icmp":
		return "All ICMP"
	default:
		return protocol
	}
}

// buildRulesPayload converts the Terraform rules list into the backend JSON payload format.
func buildRulesPayload(ctx context.Context, rulesList types.List) ([]map[string]interface{}, error) {
	if rulesList.IsNull() || rulesList.IsUnknown() || len(rulesList.Elements()) == 0 {
		return nil, nil
	}

	var rules []securityGroupRuleModel
	diags := rulesList.ElementsAs(ctx, &rules, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to parse rules: %s", diags.Errors())
	}

	// Protocols that must NOT have port ranges sent.
	noPortProtocols := map[string]bool{
		"All ICMP":     true,
		"All TCP":      true,
		"All UDP":      true,
		"Any Protocol": true,
	}

	var payload []map[string]interface{}
	for _, rule := range rules {
		var portMin, portMax int64
		if !rule.PortRangeMin.IsNull() && !rule.PortRangeMin.IsUnknown() {
			portMin = rule.PortRangeMin.ValueInt64()
		}
		if !rule.PortRangeMax.IsNull() && !rule.PortRangeMax.IsUnknown() {
			portMax = rule.PortRangeMax.ValueInt64()
		}

		protocolName := protocolToBackend(rule.Protocol.ValueString(), portMin, portMax)

		r := map[string]interface{}{
			"rule_type":     directionToBackend(rule.Direction.ValueString()),
			"protocol_name": protocolName,
		}

		// Only include port ranges for protocols that support them.
		if !noPortProtocols[protocolName] {
			if !rule.PortRangeMin.IsNull() && !rule.PortRangeMin.IsUnknown() {
				r["port_range_min"] = rule.PortRangeMin.ValueInt64()
			}
			if !rule.PortRangeMax.IsNull() && !rule.PortRangeMax.IsUnknown() {
				r["port_range_max"] = rule.PortRangeMax.ValueInt64()
			}
		}

		// Determine remote type — required by the API.
		if !rule.RemoteGroupID.IsNull() && !rule.RemoteGroupID.IsUnknown() && rule.RemoteGroupID.ValueString() != "" {
			r["remote"] = "securityGroup"
			r["remote_group_id"] = rule.RemoteGroupID.ValueString()
		} else if !rule.RemoteIPPrefix.IsNull() && !rule.RemoteIPPrefix.IsUnknown() && rule.RemoteIPPrefix.ValueString() != "" {
			r["remote"] = "manual"
			r["remote_ip_prefix"] = rule.RemoteIPPrefix.ValueString()
		} else {
			r["remote"] = "any"
		}

		if !rule.EtherType.IsNull() && !rule.EtherType.IsUnknown() {
			r["ethertype"] = rule.EtherType.ValueString()
		}
		payload = append(payload, r)
	}
	return payload, nil
}

func (r *securityGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	rules, err := buildRulesPayload(ctx, plan.Rules)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build rules payload", err.Error())
		return
	}
	if rules != nil {
		body["rules"] = rules
	}

	apiResp, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create security group", err.Error())
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	id, ok := result["id"].(string)
	if !ok {
		resp.Diagnostics.AddError("Failed to parse security group ID", "ID not found in response")
		return
	}
	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read security group", err.Error())
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	if v, ok := result["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	if v, ok := result["description"].(string); ok && v != "" {
		state.Description = types.StringValue(v)
	} else if !state.Description.IsNull() {
		// User set description — preserve API value (even empty)
		if v, ok := result["description"].(string); ok {
			state.Description = types.StringValue(v)
		}
	}
	// If user never set description (null) and API returns "", keep null

	// Rules: preserve the user's configured rules rather than trying to reconcile
	// with backend rules. The backend transforms protocol names (e.g. "tcp" → "Custom TCP",
	// "SSH"), may strip port ranges, adds default egress rules, and returns rules in a
	// different order. Since updates replace all rules via PUT, preserving the user's
	// config avoids unnecessary diffs and keeps the plan clean.
	// state.Rules is kept as-is from the previous state.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state securityGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	rules, err := buildRulesPayload(ctx, plan.Rules)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build rules payload", err.Error())
		return
	}
	if rules != nil {
		body["rules"] = rules
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	_, err = r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update security group", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"key":    "id",
		"values": []string{state.ID.ValueString()},
	}

	// Retry on "in use" errors (SG may still be attached to a recently deleted instance)
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, apiPath, body)
			return err
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete security group", err.Error())
	}
}
