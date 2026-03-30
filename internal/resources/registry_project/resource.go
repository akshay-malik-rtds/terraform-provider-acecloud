package registry_project

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/ace-registry/projects"

// Ensure the resource satisfies the expected interfaces.
var (
	_ resource.Resource              = &registryProjectResource{}
	_ resource.ResourceWithConfigure = &registryProjectResource{}
)

// registryProjectResource is the resource implementation.
type registryProjectResource struct {
	client *client.Client
}

// NewResource returns a new registry project resource factory.
func NewResource() resource.Resource {
	return &registryProjectResource{}
}

func (r *registryProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_project"
}

func (r *registryProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *registryProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"registry_name":         plan.RegistryName.ValueString(),
		"vulnerability_scanning": plan.VulnerabilityScanning.ValueBool(),
	}

	apiResp, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create registry project", err.Error())
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	// Extract project_id from create response.
	if v, ok := result["project_id"]; ok {
		plan.ID = types.StringValue(fmt.Sprintf("%v", v))
	} else if v, ok := result["id"]; ok {
		plan.ID = types.StringValue(fmt.Sprintf("%v", v))
	} else {
		resp.Diagnostics.AddError("Failed to parse registry project ID", "ID not found in create response")
		return
	}

	// Follow-up read to get full state.
	diags := r.readIntoState(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := r.readIntoState(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *registryProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan registryProjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state registryProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only vulnerability_scanning is updatable.
	if !plan.VulnerabilityScanning.Equal(state.VulnerabilityScanning) {
		body := map[string]interface{}{
			"auto_scan": plan.VulnerabilityScanning.ValueBool(),
		}

		_, err := r.client.Post(ctx, apiPath+"/update_auto_scan", body)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update registry project", err.Error())
			return
		}
	}

	// Preserve computed fields from current state, apply plan values.
	plan.ID = state.ID
	plan.CreatedAt = state.CreatedAt

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryProjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Registry API DELETE requires the registry name, not the numeric ID.
	path := fmt.Sprintf("%s/%s", apiPath, state.RegistryName.ValueString())
	_, err := r.client.Delete(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete registry project", err.Error())
		return
	}
}

// readIntoState fetches the project list, finds our project by registry_name, and
// populates the model fields. Uses the list endpoint since no direct GET-by-ID exists.
func (r *registryProjectResource) readIntoState(ctx context.Context, state *registryProjectModel) diag.Diagnostics {
	var diags diag.Diagnostics

	apiResp, err := r.client.Get(ctx, apiPath, nil)
	if err != nil {
		diags.AddError("Failed to read registry project", err.Error())
		return diags
	}

	var projects []map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &projects); err != nil {
		diags.AddError("Failed to parse registry project list", err.Error())
		return diags
	}

	targetName := state.RegistryName.ValueString()
	var found map[string]interface{}
	for _, p := range projects {
		if name, ok := p["registry_name"].(string); ok && name == targetName {
			found = p
			break
		}
	}

	if found == nil {
		diags.AddError(
			"Failed to read registry project",
			fmt.Sprintf("Registry project %q not found in list response", targetName),
		)
		return diags
	}

	// Map API fields to state.
	if v, ok := found["project_id"]; ok {
		state.ID = types.StringValue(fmt.Sprintf("%v", v))
	} else if v, ok := found["id"]; ok {
		state.ID = types.StringValue(fmt.Sprintf("%v", v))
	}

	if v, ok := found["name"].(string); ok {
		state.RegistryName = types.StringValue(v)
	}

	if v, ok := found["auto_scan"].(bool); ok {
		state.VulnerabilityScanning = types.BoolValue(v)
	} else {
		// Preserve plan value if API doesn't return the field.
		if state.VulnerabilityScanning.IsUnknown() {
			state.VulnerabilityScanning = types.BoolValue(false)
		}
	}

	if v, ok := found["creation_time"].(string); ok {
		state.CreatedAt = types.StringValue(v)
	} else if v, ok := found["created_at"].(string); ok {
		state.CreatedAt = types.StringValue(v)
	} else {
		state.CreatedAt = types.StringValue("")
	}

	return diags
}
