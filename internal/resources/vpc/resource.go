package vpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const vpcBasePath = "/cloud/vpcs"

var (
	_ resource.Resource              = &vpcResource{}
	_ resource.ResourceWithConfigure = &vpcResource{}
)

type vpcResource struct {
	client *client.Client
}

// --- API types ---

type vpcCreateSubnet struct {
	Name           string   `json:"name"`
	CIDR           string   `json:"cidr"`
	IPVersion      int64    `json:"ip_version"`
	EnableDHCP     bool     `json:"enable_dhcp"`
	DNSNameservers []string `json:"dns_nameservers,omitempty"`
	GatewayIP      string   `json:"gateway_ip,omitempty"`
}

type vpcCreateRequest struct {
	Name                string           `json:"name"`
	Description         string           `json:"description,omitempty"`
	AdminStateUp        *bool            `json:"admin_state_up,omitempty"`
	PortSecurityEnabled *bool            `json:"port_security_enabled,omitempty"`
	Subnet              *vpcCreateSubnet `json:"subnet,omitempty"`
}

type vpcUpdateRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description,omitempty"`
	AdminStateUp        *bool  `json:"admin_state_up,omitempty"`
	PortSecurityEnabled *bool  `json:"port_security_enabled,omitempty"`
}

type vpcAPIResponse struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	AdminStateUp        bool   `json:"admin_state_up"`
	PortSecurityEnabled bool   `json:"port_security_enabled"`
	Status              string `json:"status"`
	MTU                 int64  `json:"mtu"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type subnetAPIResponse struct {
	ID        string `json:"id"`
	GatewayIP string `json:"gateway_ip"`
}

type vpcDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &vpcResource{}
}

func (r *vpcResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc"
}

func (r *vpcResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = vpcSchema()
}

func (r *vpcResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected *client.Client, got an unexpected type.",
		)
		return
	}
	r.client = c
}

func buildCreateRequest(ctx context.Context, plan *vpcResourceModel) vpcCreateRequest {
	body := vpcCreateRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.AdminStateUp.IsNull() && !plan.AdminStateUp.IsUnknown() {
		v := plan.AdminStateUp.ValueBool()
		body.AdminStateUp = &v
	}
	if !plan.PortSecurityEnabled.IsNull() && !plan.PortSecurityEnabled.IsUnknown() {
		v := plan.PortSecurityEnabled.ValueBool()
		body.PortSecurityEnabled = &v
	}

	// Build inline subnet (required by backend)
	subnet := vpcCreateSubnet{
		Name:       plan.SubnetName.ValueString(),
		CIDR:       plan.SubnetCIDR.ValueString(),
		IPVersion:  plan.SubnetIPVersion.ValueInt64(),
		EnableDHCP: plan.SubnetEnableDHCP.ValueBool(),
	}

	if !plan.SubnetGatewayIP.IsNull() && !plan.SubnetGatewayIP.IsUnknown() {
		subnet.GatewayIP = plan.SubnetGatewayIP.ValueString()
	}

	if !plan.SubnetDNSNameservers.IsNull() && !plan.SubnetDNSNameservers.IsUnknown() {
		var dns []string
		plan.SubnetDNSNameservers.ElementsAs(ctx, &dns, false)
		subnet.DNSNameservers = dns
	}

	body.Subnet = &subnet
	return body
}

func buildUpdateRequest(plan *vpcResourceModel) vpcUpdateRequest {
	body := vpcUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.AdminStateUp.IsNull() && !plan.AdminStateUp.IsUnknown() {
		v := plan.AdminStateUp.ValueBool()
		body.AdminStateUp = &v
	}
	if !plan.PortSecurityEnabled.IsNull() && !plan.PortSecurityEnabled.IsUnknown() {
		v := plan.PortSecurityEnabled.ValueBool()
		body.PortSecurityEnabled = &v
	}
	return body
}

func mapAPIResponseToState(model *vpcResourceModel, apiResp *vpcAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)
	model.AdminStateUp = types.BoolValue(apiResp.AdminStateUp)
	model.PortSecurityEnabled = types.BoolValue(apiResp.PortSecurityEnabled)
	model.Status = types.StringValue(apiResp.Status)

	if apiResp.MTU > 0 {
		model.MTU = types.Int64Value(apiResp.MTU)
	} else {
		model.MTU = types.Int64Value(0)
	}

	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	if apiResp.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	} else {
		model.CreatedAt = types.StringValue("")
	}
	if apiResp.UpdatedAt != "" {
		model.UpdatedAt = types.StringValue(apiResp.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringValue("")
	}
}

// parseSubnetFromRaw extracts subnet ID and gateway_ip from the raw API response.
// Create response returns "subnet" (object); read response returns "subnets" (array of IDs).
func parseSubnetFromRaw(raw map[string]json.RawMessage) (subnetID, gatewayIP string) {
	// Try "subnet" key first (create response)
	if subnetRaw, ok := raw["subnet"]; ok {
		var sub subnetAPIResponse
		if json.Unmarshal(subnetRaw, &sub) == nil && sub.ID != "" {
			return sub.ID, sub.GatewayIP
		}
	}

	// Try "subnets" key (read response — array of subnet IDs or objects)
	if subnetsRaw, ok := raw["subnets"]; ok {
		// First try as array of objects
		var subs []subnetAPIResponse
		if json.Unmarshal(subnetsRaw, &subs) == nil && len(subs) > 0 && subs[0].ID != "" {
			return subs[0].ID, subs[0].GatewayIP
		}

		// Then try as array of strings (just IDs)
		var ids []string
		if json.Unmarshal(subnetsRaw, &ids) == nil && len(ids) > 0 {
			return ids[0], ""
		}
	}

	return "", ""
}

func (r *vpcResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	apiResp, err := r.client.Post(ctx, vpcBasePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create VPC", err.Error())
		return
	}

	// Parse raw response for VPC + subnet data
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(apiResp.Data, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to parse VPC response", err.Error())
		return
	}

	var created vpcAPIResponse
	if err := json.Unmarshal(apiResp.Data, &created); err != nil {
		resp.Diagnostics.AddError("Failed to parse VPC response", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)

	// Parse subnet from create response
	subnetID, gatewayIP := parseSubnetFromRaw(raw)
	if subnetID != "" {
		plan.SubnetID = types.StringValue(subnetID)
	} else {
		plan.SubnetID = types.StringValue("")
	}
	if gatewayIP != "" {
		plan.SubnetGatewayIP = types.StringValue(gatewayIP)
	} else if plan.SubnetGatewayIP.IsUnknown() {
		plan.SubnetGatewayIP = types.StringValue("")
	}

	// Save subnet fields from create response — follow-up Read returns all
	// subnet IDs in an array but no gateway_ip, so create response is more reliable
	savedSubnetID := plan.SubnetID
	savedSubnetGatewayIP := plan.SubnetGatewayIP

	// Follow-up Read to get accurate state (create response may have stale
	// values for fields like port_security_enabled)
	readPath := fmt.Sprintf("%s/%s", vpcBasePath, created.ID)
	readResp, err := r.client.Get(ctx, readPath, nil)
	if err != nil {
		// Non-fatal: fall back to create response data
		mapAPIResponseToState(&plan, &created)
		plan.SubnetID = savedSubnetID
		plan.SubnetGatewayIP = savedSubnetGatewayIP
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	var readVPC vpcAPIResponse
	if err := json.Unmarshal(readResp.Data, &readVPC); err != nil {
		mapAPIResponseToState(&plan, &created)
		plan.SubnetID = savedSubnetID
		plan.SubnetGatewayIP = savedSubnetGatewayIP
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	mapAPIResponseToState(&plan, &readVPC)

	// Restore subnet fields from create response (more reliable than Read)
	plan.SubnetID = savedSubnetID
	plan.SubnetGatewayIP = savedSubnetGatewayIP

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vpcResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", vpcBasePath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read VPC", err.Error())
		return
	}

	// Save subnet fields before mapping — Read response returns all subnet IDs
	// in an array but no gateway_ip, so we preserve the known values from state
	savedSubnetID := state.SubnetID
	savedSubnetName := state.SubnetName
	savedSubnetCIDR := state.SubnetCIDR
	savedSubnetIPVersion := state.SubnetIPVersion
	savedSubnetGatewayIP := state.SubnetGatewayIP
	savedSubnetEnableDHCP := state.SubnetEnableDHCP
	savedSubnetDNS := state.SubnetDNSNameservers

	var vpc vpcAPIResponse
	if err := json.Unmarshal(apiResp.Data, &vpc); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	mapAPIResponseToState(&state, &vpc)

	// Restore subnet fields — they are managed at create time and don't change
	// via VPC read/update. The Read response only has subnet IDs, not details.
	state.SubnetID = savedSubnetID
	state.SubnetName = savedSubnetName
	state.SubnetCIDR = savedSubnetCIDR
	state.SubnetIPVersion = savedSubnetIPVersion
	state.SubnetGatewayIP = savedSubnetGatewayIP
	state.SubnetEnableDHCP = savedSubnetEnableDHCP
	state.SubnetDNSNameservers = savedSubnetDNS

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vpcResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vpcResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state vpcResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(&plan)

	path := fmt.Sprintf("%s/%s", vpcBasePath, state.ID.ValueString())
	_, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update VPC", err.Error())
		return
	}

	// Follow-up Read to get accurate state (update response may be stale
	// for fields like port_security_enabled and subnet data)
	readResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read VPC after update", err.Error())
		return
	}

	var readVPC vpcAPIResponse
	if err := json.Unmarshal(readResp.Data, &readVPC); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response after update", err.Error())
		return
	}

	mapAPIResponseToState(&plan, &readVPC)

	// Subnet fields are immutable via VPC update — always preserve from prior state.
	// The Read response returns "subnets":["id1","id2"] (just IDs) when multiple subnets
	// exist, which causes parseSubnetFromRaw to pick the wrong subnet's data.
	plan.SubnetID = state.SubnetID
	plan.SubnetName = state.SubnetName
	plan.SubnetCIDR = state.SubnetCIDR
	plan.SubnetIPVersion = state.SubnetIPVersion
	plan.SubnetGatewayIP = state.SubnetGatewayIP
	plan.SubnetEnableDHCP = state.SubnetEnableDHCP
	plan.SubnetDNSNameservers = state.SubnetDNSNameservers

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vpcResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := vpcDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	_, err := r.client.Delete(ctx, vpcBasePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete VPC", err.Error())
	}
}
