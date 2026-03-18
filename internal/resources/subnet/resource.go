package subnet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/vpcs/subnets"

// Ensure the resource satisfies the expected interfaces.
var (
	_ resource.Resource              = &subnetResource{}
	_ resource.ResourceWithConfigure = &subnetResource{}
)

// subnetResource is the resource implementation.
type subnetResource struct {
	client *client.Client
}

// NewResource returns a new subnet resource factory.
func NewResource() resource.Resource {
	return &subnetResource{}
}

func (r *subnetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnet"
}

func (r *subnetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// buildBody constructs the API request body from the Terraform plan.
func (r *subnetResource) buildBody(ctx context.Context, plan *subnetModel) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"name":       plan.Name.ValueString(),
		"cidr":       plan.CIDR.ValueString(),
		"ip_version": plan.IPVersion.ValueInt64(),
	}

	if !plan.VPCID.IsNull() && !plan.VPCID.IsUnknown() {
		body["vpc_id"] = plan.VPCID.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}
	if !plan.EnableDHCP.IsNull() && !plan.EnableDHCP.IsUnknown() {
		body["enable_dhcp"] = plan.EnableDHCP.ValueBool()
	}
	if !plan.GatewayIP.IsNull() && !plan.GatewayIP.IsUnknown() {
		body["gateway_ip"] = plan.GatewayIP.ValueString()
	}

	// DNS nameservers.
	if !plan.DNSNameservers.IsNull() && !plan.DNSNameservers.IsUnknown() {
		var nameservers []string
		diags := plan.DNSNameservers.ElementsAs(ctx, &nameservers, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to parse dns_nameservers: %s", diags.Errors())
		}
		body["dns_nameservers"] = nameservers
	}

	// Allocation pools.
	if !plan.AllocationPools.IsNull() && !plan.AllocationPools.IsUnknown() && len(plan.AllocationPools.Elements()) > 0 {
		var pools []allocationPoolModel
		diags := plan.AllocationPools.ElementsAs(ctx, &pools, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to parse allocation_pools: %s", diags.Errors())
		}
		var poolPayload []map[string]string
		for _, p := range pools {
			poolPayload = append(poolPayload, map[string]string{
				"start": p.Start.ValueString(),
				"end":   p.End.ValueString(),
			})
		}
		body["allocation_pools"] = poolPayload
	}

	// Host routes.
	if !plan.HostRoutes.IsNull() && !plan.HostRoutes.IsUnknown() && len(plan.HostRoutes.Elements()) > 0 {
		var routes []hostRouteModel
		diags := plan.HostRoutes.ElementsAs(ctx, &routes, false)
		if diags.HasError() {
			return nil, fmt.Errorf("failed to parse host_routes: %s", diags.Errors())
		}
		var routePayload []map[string]string
		for _, hr := range routes {
			routePayload = append(routePayload, map[string]string{
				"destination": hr.Destination.ValueString(),
				"nexthop":     hr.Nexthop.ValueString(),
			})
		}
		body["host_routes"] = routePayload
	}

	return body, nil
}

func (r *subnetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.buildBody(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build request body", err.Error())
		return
	}

	apiResp, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create subnet", err.Error())
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	id, ok := result["id"].(string)
	if !ok {
		resp.Diagnostics.AddError("Failed to parse subnet ID", "ID not found in response")
		return
	}
	plan.ID = types.StringValue(id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read subnet", err.Error())
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
	if v, ok := result["cidr"].(string); ok {
		state.CIDR = types.StringValue(v)
	}
	if v, ok := result["vpc_id"].(string); ok && v != "" {
		state.VPCID = types.StringValue(v)
	} else if v, ok := result["network_id"].(string); ok && v != "" {
		// Fallback for direct OpenStack responses
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
			nsList, diags := types.ListValueFrom(ctx, types.StringType, nameservers)
			resp.Diagnostics.Append(diags...)
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
			poolType := state.AllocationPools.ElementType(ctx)
			poolsList, diags := types.ListValueFrom(ctx, poolType, poolModels)
			resp.Diagnostics.Append(diags...)
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
			routeType := state.HostRoutes.ElementType(ctx)
			routesList, diags := types.ListValueFrom(ctx, routeType, routeModels)
			resp.Diagnostics.Append(diags...)
			state.HostRoutes = routesList
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subnetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subnetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state subnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.buildBody(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build request body", err.Error())
		return
	}

	// Immutable fields — must not be sent in update requests
	delete(body, "vpc_id")
	delete(body, "ip_version")
	delete(body, "cidr")

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	_, err = r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update subnet", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subnetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subnetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Matches CLI pattern: {"key": "id", "values": [...]}
	body := map[string]interface{}{
		"key":    "id",
		"values": []string{state.ID.ValueString()},
	}

	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(_ context.Context) error {
			_, err := r.client.Delete(ctx, apiPath, body)
			return err
		},
		RetryableErrors: []string{"in use", "IP allocation", "ports have an IP"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete subnet", err.Error())
		return
	}
}
