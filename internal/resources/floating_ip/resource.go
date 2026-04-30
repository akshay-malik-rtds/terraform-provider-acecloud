package floating_ip

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/floating-ips"

// Ensure the resource satisfies the expected interfaces.
var (
	_ resource.Resource              = &floatingIPResource{}
	_ resource.ResourceWithConfigure = &floatingIPResource{}
)

// floatingIPResource is the resource implementation.
type floatingIPResource struct {
	client *client.Client
}

// NewResource returns a new floating IP resource factory.
func NewResource() resource.Resource {
	return &floatingIPResource{}
}

func (r *floatingIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip"
}

func (r *floatingIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *floatingIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan floatingIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build request body — billing_type is always "hourly" for floating IPs (only allowed value)
	body := map[string]interface{}{
		"floating_network_id": plan.FloatingNetworkID.ValueString(),
		"billing_type":        "hourly",
	}
	if !plan.PortID.IsNull() && !plan.PortID.IsUnknown() {
		body["port_id"] = plan.PortID.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body["description"] = plan.Description.ValueString()
	}

	apiResp, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create floating IP", err.Error())
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	id, ok := result["id"].(string)
	if !ok {
		resp.Diagnostics.AddError("Failed to parse floating IP ID", "ID not found in response")
		return
	}
	plan.ID = types.StringValue(id)

	if v, ok := result["floating_ip_address"].(string); ok {
		plan.FloatingIPAddress = types.StringValue(v)
	} else {
		plan.FloatingIPAddress = types.StringNull()
	}
	if v, ok := result["status"].(string); ok {
		plan.Status = types.StringValue(v)
	} else {
		plan.Status = types.StringValue("ACTIVE")
	}
	// Ensure port_id is set to a known value (may be null if not associated)
	if v, ok := result["port_id"].(string); ok && v != "" {
		plan.PortID = types.StringValue(v)
	} else if plan.PortID.IsUnknown() {
		plan.PortID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *floatingIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state floatingIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// npc-api has no GET-by-ID for floating IPs. Use list endpoint and filter.
	apiResp, err := r.client.Get(ctx, apiPath, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read floating IP", err.Error())
		return
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &items); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	// Find our floating IP by ID
	targetID := state.ID.ValueString()
	var result map[string]interface{}
	for _, item := range items {
		if id, ok := item["id"].(string); ok && id == targetID {
			result = item
			break
		}
	}

	if result == nil {
		// Floating IP no longer exists
		resp.State.RemoveResource(ctx)
		return
	}

	if v, ok := result["floating_network_id"].(string); ok {
		state.FloatingNetworkID = types.StringValue(v)
	}
	if v, ok := result["port_id"].(string); ok && v != "" {
		state.PortID = types.StringValue(v)
	} else {
		state.PortID = types.StringNull()
	}
	if v, ok := result["description"].(string); ok && v != "" {
		state.Description = types.StringValue(v)
	} else if !state.Description.IsNull() {
		// User set description; API returned empty — preserve empty string
		state.Description = types.StringValue("")
	}
	// If user never set description (null) and API returns "", keep null
	if v, ok := result["floating_ip_address"].(string); ok {
		state.FloatingIPAddress = types.StringValue(v)
	}
	if v, ok := result["status"].(string); ok {
		state.Status = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *floatingIPResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Floating IPs do not support update via PUT. Any change to mutable fields
	// (such as port_id) would require a separate associate/disassociate API.
	// For now, Terraform will destroy and recreate on changes.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Floating IP resources do not support in-place updates. Changes will trigger a destroy and recreate.",
	)
}

func (r *floatingIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state floatingIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Matches CLI pattern: {"key": "id", "values": [...]}
	body := map[string]interface{}{
		"key":    "id",
		"values": []string{state.ID.ValueString()},
	}

	// Floating IP deletion can fail if it's still associated with a port
	// that is draining after instance destruction.
	// npc-api returns: "...is either attached or already released"
	// npc-api returns: "Floating IP is already associated with a port"
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, apiPath, body)
			return err
		},
		RetryableErrors: []string{"either attached", "already associated"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete floating IP", err.Error())
		return
	}
}
