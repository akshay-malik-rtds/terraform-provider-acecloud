package router_interface

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiBasePath = "/os/neutron/interfaces"

var (
	_ resource.Resource              = &routerInterfaceResource{}
	_ resource.ResourceWithConfigure = &routerInterfaceResource{}
)

type routerInterfaceResource struct {
	client *client.Client
}

// --- API types (matching ace-cli/internal/api/router.go) ---

type interfaceCreateRequest struct {
	Subnet string `json:"subnet"`
}

type interfaceAPIResponse struct {
	ID         string    `json:"id"`
	PortID     string    `json:"port_id"`
	SubnetID   string    `json:"subnet_id"`
	IPAddress  string    `json:"ip_address"`
	Status     string    `json:"status"`
	MACAddress string    `json:"mac_address"`
	FixedIPs   []fixedIP `json:"fixed_ips"`
}

type fixedIP struct {
	SubnetID  string `json:"subnet_id"`
	IPAddress string `json:"ip_address"`
}

type interfaceDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &routerInterfaceResource{}
}

func (r *routerInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_router_interface"
}

func (r *routerInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = routerInterfaceSchema()
}

func (r *routerInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *routerInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routerInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiBasePath, plan.RouterID.ValueString())
	body := interfaceCreateRequest{
		Subnet: plan.SubnetID.ValueString(),
	}

	apiResp, err := r.client.Post(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create router interface", err.Error())
		return
	}

	var result interfaceAPIResponse
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	// OpenStack "add router interface" returns id=router_id, port_id=port_id.
	// Use port_id as the resource ID for correct delete operations.
	if result.PortID != "" {
		plan.ID = types.StringValue(result.PortID)
	} else {
		plan.ID = types.StringValue(result.ID)
	}
	plan.Status = types.StringValue(result.Status)
	plan.MACAddress = types.StringValue(result.MACAddress)

	// Extract IP address from fixed_ips or direct field
	if result.IPAddress != "" {
		plan.IPAddress = types.StringValue(result.IPAddress)
	} else if len(result.FixedIPs) > 0 {
		plan.IPAddress = types.StringValue(result.FixedIPs[0].IPAddress)
	} else {
		plan.IPAddress = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routerInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routerInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GET /os/neutron/interfaces/{router_id} returns array of interfaces
	path := fmt.Sprintf("%s/%s", apiBasePath, state.RouterID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read router interfaces", err.Error())
		return
	}

	// API returns {"ports": [...]} wrapper
	var interfaces []interfaceAPIResponse
	var wrapped struct {
		Ports []interfaceAPIResponse `json:"ports"`
	}
	if err := json.Unmarshal(apiResp.Data, &wrapped); err == nil && len(wrapped.Ports) > 0 {
		interfaces = wrapped.Ports
	} else if err := json.Unmarshal(apiResp.Data, &interfaces); err != nil {
		// Try single object format
		var single interfaceAPIResponse
		if err2 := json.Unmarshal(apiResp.Data, &single); err2 != nil {
			resp.Diagnostics.AddError("Failed to parse read response", err.Error())
			return
		}
		interfaces = []interfaceAPIResponse{single}
	}

	// Find the interface matching our subnet_id
	found := false
	for _, iface := range interfaces {
		// Check direct subnet_id match
		if iface.SubnetID == state.SubnetID.ValueString() {
			mapInterfaceToState(&state, &iface)
			found = true
			break
		}
		// Check fixed_ips for subnet match
		for _, fip := range iface.FixedIPs {
			if fip.SubnetID == state.SubnetID.ValueString() {
				mapInterfaceToState(&state, &iface)
				if fip.IPAddress != "" {
					state.IPAddress = types.StringValue(fip.IPAddress)
				}
				found = true
				break
			}
		}
		if found {
			break
		}
		// Also try matching by port ID
		if iface.ID == state.ID.ValueString() {
			mapInterfaceToState(&state, &iface)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routerInterfaceResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Router interfaces do not support update. Changes trigger destroy and recreate.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Router interface resources do not support in-place updates. Changes will trigger a destroy and recreate.",
	)
}

func (r *routerInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routerInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiBasePath, state.RouterID.ValueString())
	body := interfaceDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	_, err := r.client.Delete(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete router interface", err.Error())
		return
	}
}

func mapInterfaceToState(model *routerInterfaceModel, iface *interfaceAPIResponse) {
	model.ID = types.StringValue(iface.ID)
	model.Status = types.StringValue(iface.Status)
	model.MACAddress = types.StringValue(iface.MACAddress)
	if iface.IPAddress != "" {
		model.IPAddress = types.StringValue(iface.IPAddress)
	} else if len(iface.FixedIPs) > 0 {
		model.IPAddress = types.StringValue(iface.FixedIPs[0].IPAddress)
	} else {
		model.IPAddress = types.StringValue("")
	}
}
