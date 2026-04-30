package auto_scaling_template

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const basePath = "/auto-scaling/templates"

var (
	_ resource.Resource              = &autoScalingTemplateResource{}
	_ resource.ResourceWithConfigure = &autoScalingTemplateResource{}
)

type autoScalingTemplateResource struct {
	client *client.Client
}

// --- API types ---

type templateCreateRequest struct {
	Name                string   `json:"name"`
	Type                string   `json:"type"`
	Description         string   `json:"description,omitempty"`
	VolumeSize          int64    `json:"volume_size"`
	VolDelOnTermination bool     `json:"vol_del_on_termination"`
	FlavorID            string   `json:"flavor_id"`
	ImageID             string   `json:"image_id,omitempty"`
	SnapshotID          string   `json:"snapshot_id,omitempty"`
	KeyName             string   `json:"key_name,omitempty"`
	NetworkID           string   `json:"network_id"`
	SubnetID            string   `json:"subnet_id"`
	SecurityGroups      []string `json:"security_groups"`
	IsInstanceSnapshot  bool     `json:"is_instance_snapshot"`
}

type templateUpdateRequest struct {
	Name                string   `json:"name"`
	Type                string   `json:"type"`
	Description         string   `json:"description,omitempty"`
	VolumeSize          int64    `json:"volume_size"`
	VolDelOnTermination bool     `json:"vol_del_on_termination"`
	FlavorID            string   `json:"flavor_id"`
	ImageID             string   `json:"image_id,omitempty"`
	SnapshotID          string   `json:"snapshot_id,omitempty"`
	KeyName             string   `json:"key_name,omitempty"`
	NetworkID           string   `json:"network_id"`
	SubnetID            string   `json:"subnet_id"`
	SecurityGroups      []string `json:"security_groups"`
	IsInstanceSnapshot  bool     `json:"is_instance_snapshot"`
}

// templateAPIResponse matches the actual API response structure.
// Note: The API returns nested objects for flavor, image, network, snapshot,
// and security_groups (each with id+name), not flat IDs.
type templateAPIResponse struct {
	ID                  string                   `json:"id"`
	Name                string                   `json:"name"`
	Type                string                   `json:"type"`
	Description         string                   `json:"description"`
	VolumeSize          int64                    `json:"volume_size"`
	VolDelOnTermination bool                     `json:"vol_del_on_termination"`
	Flavor              *nestedIDName            `json:"flavor"`
	Image               *nestedIDName            `json:"image"`
	Snapshot            *nestedIDName            `json:"snapshot"`
	KeyName             string                   `json:"key_name"`
	Network             *nestedIDName            `json:"network"`
	SubnetID            string                   `json:"subnet_id"`
	SecurityGroups      []nestedIDName           `json:"security_groups"`
	IsInstanceSnapshot  bool                     `json:"is_instance_snapshot"`
	Status              string                   `json:"status"`
	Region              string                   `json:"region"`
	CreatedAt           string                   `json:"created_at"`
	UpdatedAt           string                   `json:"updated_at"`
}

// nestedIDName represents the API's nested {id, name} objects.
type nestedIDName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Helper methods to extract IDs from the nested API response.
func (r *templateAPIResponse) FlavorID() string {
	if r.Flavor != nil {
		return r.Flavor.ID
	}
	return ""
}

func (r *templateAPIResponse) ImageID() string {
	if r.Image != nil {
		return r.Image.ID
	}
	return ""
}

func (r *templateAPIResponse) SnapshotID() string {
	if r.Snapshot != nil {
		return r.Snapshot.ID
	}
	return ""
}

func (r *templateAPIResponse) NetworkID() string {
	if r.Network != nil {
		return r.Network.ID
	}
	return ""
}

func (r *templateAPIResponse) SecurityGroupIDs() []string {
	ids := make([]string, len(r.SecurityGroups))
	for i, sg := range r.SecurityGroups {
		ids[i] = sg.ID
	}
	return ids
}

// createResponseID is used when the create response only returns an id.
type createResponseID struct {
	ID string `json:"id"`
}

func NewResource() resource.Resource {
	return &autoScalingTemplateResource{}
}

func (r *autoScalingTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auto_scaling_template"
}

func (r *autoScalingTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = autoScalingTemplateSchema()
}

func (r *autoScalingTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(ctx context.Context, plan *autoScalingTemplateModel) templateCreateRequest {
	body := templateCreateRequest{
		Name:                plan.Name.ValueString(),
		Type:                plan.Type.ValueString(),
		VolumeSize:          plan.VolumeSize.ValueInt64(),
		VolDelOnTermination: plan.VolDelOnTermination.ValueBool(),
		FlavorID:            plan.FlavorID.ValueString(),
		NetworkID:           plan.NetworkID.ValueString(),
		SubnetID:            plan.SubnetID.ValueString(),
		IsInstanceSnapshot:  plan.IsInstanceSnapshot.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.ImageID.IsNull() && !plan.ImageID.IsUnknown() {
		body.ImageID = plan.ImageID.ValueString()
	}
	if !plan.SnapshotID.IsNull() && !plan.SnapshotID.IsUnknown() {
		body.SnapshotID = plan.SnapshotID.ValueString()
	}
	if !plan.KeyName.IsNull() && !plan.KeyName.IsUnknown() {
		body.KeyName = plan.KeyName.ValueString()
	}

	if !plan.SecurityGroups.IsNull() && !plan.SecurityGroups.IsUnknown() {
		var sgs []string
		plan.SecurityGroups.ElementsAs(ctx, &sgs, false)
		body.SecurityGroups = sgs
	}

	return body
}

func buildUpdateRequest(ctx context.Context, plan *autoScalingTemplateModel) templateUpdateRequest {
	body := templateUpdateRequest{
		Name:                plan.Name.ValueString(),
		Type:                plan.Type.ValueString(),
		VolumeSize:          plan.VolumeSize.ValueInt64(),
		VolDelOnTermination: plan.VolDelOnTermination.ValueBool(),
		FlavorID:            plan.FlavorID.ValueString(),
		NetworkID:           plan.NetworkID.ValueString(),
		SubnetID:            plan.SubnetID.ValueString(),
		IsInstanceSnapshot:  plan.IsInstanceSnapshot.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.ImageID.IsNull() && !plan.ImageID.IsUnknown() {
		body.ImageID = plan.ImageID.ValueString()
	}
	if !plan.SnapshotID.IsNull() && !plan.SnapshotID.IsUnknown() {
		body.SnapshotID = plan.SnapshotID.ValueString()
	}
	if !plan.KeyName.IsNull() && !plan.KeyName.IsUnknown() {
		body.KeyName = plan.KeyName.ValueString()
	}

	if !plan.SecurityGroups.IsNull() && !plan.SecurityGroups.IsUnknown() {
		var sgs []string
		plan.SecurityGroups.ElementsAs(ctx, &sgs, false)
		body.SecurityGroups = sgs
	}

	return body
}

func mapAPIResponseToState(ctx context.Context, model *autoScalingTemplateModel, apiResp *templateAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)
	model.Type = types.StringValue(apiResp.Type)
	model.VolumeSize = types.Int64Value(apiResp.VolumeSize)
	model.VolDelOnTermination = types.BoolValue(apiResp.VolDelOnTermination)
	model.IsInstanceSnapshot = types.BoolValue(apiResp.IsInstanceSnapshot)
	model.SubnetID = types.StringValue(apiResp.SubnetID)

	// Nested objects → flat IDs
	model.FlavorID = types.StringValue(apiResp.FlavorID())
	model.NetworkID = types.StringValue(apiResp.NetworkID())

	// Optional string fields — preserve null if plan had null and API returns empty
	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	if id := apiResp.ImageID(); id != "" {
		model.ImageID = types.StringValue(id)
	} else if model.ImageID.IsNull() {
		model.ImageID = types.StringNull()
	}

	if id := apiResp.SnapshotID(); id != "" {
		model.SnapshotID = types.StringValue(id)
	} else if model.SnapshotID.IsNull() {
		model.SnapshotID = types.StringNull()
	}

	if apiResp.KeyName != "" {
		model.KeyName = types.StringValue(apiResp.KeyName)
	} else if model.KeyName.IsNull() {
		model.KeyName = types.StringNull()
	}

	// Security groups — nested {id, name} objects → flat ID list
	sgIDs := apiResp.SecurityGroupIDs()
	if len(sgIDs) > 0 {
		sgList, _ := types.ListValueFrom(ctx, types.StringType, sgIDs)
		model.SecurityGroups = sgList
	}

	// Computed fields
	if apiResp.Status != "" {
		model.Status = types.StringValue(apiResp.Status)
	} else {
		model.Status = types.StringValue("")
	}

	if apiResp.Region != "" {
		model.Region = types.StringValue(apiResp.Region)
	} else {
		model.Region = types.StringValue("")
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

func (r *autoScalingTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan autoScalingTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	apiResp, err := r.client.Post(ctx, basePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create auto scaling template", err.Error())
		return
	}

	// Create response may return just {id: "..."} or full object
	var created createResponseID
	if err := json.Unmarshal(apiResp.Data, &created); err != nil {
		resp.Diagnostics.AddError("Failed to parse auto scaling template response", err.Error())
		return
	}

	if created.ID == "" {
		resp.Diagnostics.AddError("Failed to create auto scaling template", "API returned empty ID")
		return
	}

	plan.ID = types.StringValue(created.ID)

	// Follow-up Read to get full state
	readPath := fmt.Sprintf("%s/%s", basePath, created.ID)
	readResp, err := r.client.Get(ctx, readPath, nil)
	if err != nil {
		// Non-fatal: set what we have
		plan.Status = types.StringValue("")
		plan.Region = types.StringValue("")
		plan.CreatedAt = types.StringValue("")
		plan.UpdatedAt = types.StringValue("")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	var readTemplate templateAPIResponse
	if err := json.Unmarshal(readResp.Data, &readTemplate); err != nil {
		plan.Status = types.StringValue("")
		plan.Region = types.StringValue("")
		plan.CreatedAt = types.StringValue("")
		plan.UpdatedAt = types.StringValue("")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	mapAPIResponseToState(ctx, &plan, &readTemplate)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *autoScalingTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state autoScalingTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read auto scaling template", err.Error())
		return
	}

	var template templateAPIResponse
	if err := json.Unmarshal(apiResp.Data, &template); err != nil {
		resp.Diagnostics.AddError("Failed to parse auto scaling template response", err.Error())
		return
	}

	mapAPIResponseToState(ctx, &state, &template)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *autoScalingTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan autoScalingTemplateModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state autoScalingTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(ctx, &plan)

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	_, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update auto scaling template", err.Error())
		return
	}

	// Update response may return incomplete data — always do a follow-up Read
	readResp, readErr := r.client.Get(ctx, path, nil)
	if readErr != nil {
		resp.Diagnostics.AddError("Failed to read auto scaling template after update", readErr.Error())
		return
	}

	var updated templateAPIResponse
	if err := json.Unmarshal(readResp.Data, &updated); err != nil {
		resp.Diagnostics.AddError("Failed to parse auto scaling template response", err.Error())
		return
	}

	plan.ID = state.ID
	mapAPIResponseToState(ctx, &plan, &updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *autoScalingTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state autoScalingTemplateModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	_, err := r.client.Delete(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete auto scaling template", err.Error())
	}
}
