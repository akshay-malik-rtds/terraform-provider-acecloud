package instance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the resource interfaces.
var (
	_ resource.Resource              = &instanceResource{}
	_ resource.ResourceWithConfigure = &instanceResource{}
)

// instanceResource is the resource implementation.
type instanceResource struct {
	client *client.Client
}

// NewResource is the factory function registered in the provider.
func NewResource() resource.Resource {
	return &instanceResource{}
}

// Metadata returns the resource type name.
func (r *instanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Configure stores the provider-configured client for later use.
func (r *instanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// ---------------------------------------------------------------------------
// API request / response types
// ---------------------------------------------------------------------------

// createInstanceRequest is the JSON body sent to POST /cloud/instances.
type createInstanceRequest struct {
	Name                string            `json:"name"`
	Description         string            `json:"description,omitempty"`
	Count               int               `json:"count"`
	Flavor              string            `json:"flavor"`
	BillingType         string            `json:"billing_type"`
	BootUUID            string            `json:"boot_uuid"`
	SourceType          string            `json:"source_type"`
	DeleteOnTermination bool              `json:"delete_on_termination"`
	Volumes             []volumeRequest   `json:"volumes"`
	Network             []string          `json:"network"`
	SecurityGroup       []string          `json:"security_group"`
	AvailabilityZone    string            `json:"availability_zone"`
	Metadata            map[string]string `json:"metadata,omitempty"`
	Key                 string            `json:"key,omitempty"`
	Script              string            `json:"script,omitempty"`
	ServerGroupID       string            `json:"server_group_id,omitempty"`
	ConfigDrive         bool              `json:"config_drive"`
	AdminPassword       string            `json:"admin_password,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
}

type volumeRequest struct {
	Size        int64  `json:"size"`
	Boot        bool   `json:"boot"`
	VolumeType  string `json:"volume_type"`
	BillingType string `json:"billing_type"`
}

// updateInstanceRequest is the JSON body sent to PUT /cloud/instances/:id.
type updateInstanceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// deleteInstanceRequest is the JSON body sent to DELETE /cloud/instances.
// Matches CLI pattern: {"key": "id", "values": [...]}
type deleteInstanceRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

// createInstanceResponse represents the relevant fields returned by the create
// endpoint. The backend returns a list of created instances.
type createInstanceResponse struct {
	ID string `json:"id"`
}

// readInstanceResponse represents the relevant fields returned by the read
// endpoint GET /cloud/instances/:id.
// Note: The npc-api read response uses a different structure than the create request.
// - flavor/image are objects {id, name}, not strings
// - security_groups are objects [{id, name}], not string arrays
// - config_drive is a string ("" or "True"), not bool
// - volumes_attached is [{id, device, name, size}], not the create format
// - network info is in addresses{public/private}, not a string array
type readInstanceResponse struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Status           string            `json:"status"`
	AvailabilityZone string            `json:"availability_zone"`
	Metadata         map[string]string `json:"metadata"`
	Key              string            `json:"key"`
	Tags             []string          `json:"tags"`
	// These are extracted from the raw response in parseInstanceData()
	FlavorID       string
	ImageID        string
	ConfigDrive    bool
	SecurityGroups []string // security group IDs
	Networks       []string // network names from addresses
}

// parseInstanceData unmarshals the raw API response into readInstanceResponse,
// handling the nested object fields that differ from the create request format.
func parseInstanceData(data json.RawMessage) (*readInstanceResponse, error) {
	var resp readInstanceResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal base fields: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal raw: %w", err)
	}

	// flavor is {id, name} object
	if flavorObj, ok := raw["flavor"].(map[string]interface{}); ok {
		if id, ok := flavorObj["id"].(string); ok {
			resp.FlavorID = id
		}
	}

	// image is {id, name} object
	if imageObj, ok := raw["image"].(map[string]interface{}); ok {
		if id, ok := imageObj["id"].(string); ok {
			resp.ImageID = id
		}
	}

	// config_drive is string "" or "True"
	if cd, ok := raw["config_drive"].(string); ok {
		resp.ConfigDrive = cd == "True"
	}

	// security_groups is [{id, name}, ...]
	if sgList, ok := raw["security_groups"].([]interface{}); ok {
		for _, sg := range sgList {
			if sgObj, ok := sg.(map[string]interface{}); ok {
				if id, ok := sgObj["id"].(string); ok {
					resp.SecurityGroups = append(resp.SecurityGroups, id)
				}
			}
		}
	}

	return &resp, nil
}

// ---------------------------------------------------------------------------
// CRUD operations
// ---------------------------------------------------------------------------

// Create provisions a new instance via POST /cloud/instances.
func (r *instanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the API request body from the Terraform plan.
	body, diags := buildCreateRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Post(ctx, "/cloud/instances", body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create instance", err.Error())
		return
	}

	// The backend returns either a single object or an array.
	// Try single object first (npc-api cloud wrapper), then array fallback.
	var singleCreated createInstanceResponse
	if err := json.Unmarshal(apiResp.Data, &singleCreated); err == nil && singleCreated.ID != "" {
		plan.ID = types.StringValue(singleCreated.ID)
	} else {
		var created []createInstanceResponse
		if err := json.Unmarshal(apiResp.Data, &created); err != nil {
			resp.Diagnostics.AddError(
				"Failed to parse create response",
				fmt.Sprintf("Could not unmarshal response data: %s", err),
			)
			return
		}
		if len(created) == 0 {
			resp.Diagnostics.AddError(
				"Failed to create instance",
				"The API returned an empty list of instances.",
			)
			return
		}
		plan.ID = types.StringValue(created[0].ID)
	}

	// Wait for instance to become ACTIVE.
	instanceID := plan.ID.ValueString()
	plan.Status = types.StringValue("BUILD")
	result, _ := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			readResp, err := r.client.Get(ctx, fmt.Sprintf("/cloud/instances/%s", instanceID), nil)
			if err != nil {
				return nil, err
			}
			instance, err := parseInstanceData(readResp.Data)
			if err != nil {
				return nil, err
			}
			return &wait.StatusResult{Status: instance.Status, Data: instance}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
	})
	if result != nil && result.Data != nil {
		instance := result.Data.(*readInstanceResponse)
		plan.Status = types.StringValue(instance.Status)
		mapReadResponseToState(ctx, instance, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the backend via GET /cloud/instances/:id.
func (r *instanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	apiResp, err := r.client.Get(ctx, fmt.Sprintf("/cloud/instances/%s", id), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read instance",
			fmt.Sprintf("Could not read instance %s: %s", id, err),
		)
		return
	}

	instance, err := parseInstanceData(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse read response",
			fmt.Sprintf("Could not parse instance data: %s", err),
		)
		return
	}

	// Map the API response back into the Terraform state model.
	diags := mapReadResponseToState(ctx, instance, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies name and description via PUT /cloud/instances/:id.
func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body := updateInstanceRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}

	_, err := r.client.Put(ctx, fmt.Sprintf("/cloud/instances/%s", id), body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update instance",
			fmt.Sprintf("Could not update instance %s: %s", id, err),
		)
		return
	}

	// Preserve the existing ID and refresh status from the backend.
	plan.ID = state.ID

	apiResp, err := r.client.Get(ctx, fmt.Sprintf("/cloud/instances/%s", id), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read instance after update",
			fmt.Sprintf("Could not read instance %s: %s", id, err),
		)
		return
	}

	instance, parseErr := parseInstanceData(apiResp.Data)
	if parseErr != nil {
		resp.Diagnostics.AddError(
			"Failed to parse read response after update",
			fmt.Sprintf("Could not parse instance data: %s", parseErr),
		)
		return
	}

	plan.Status = types.StringValue(instance.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the instance via DELETE /cloud/instances.
func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body := deleteInstanceRequest{
		Key:    "id",
		Values: []string{id},
	}

	// Instance deletion can fail transiently when the instance is in a
	// transitional state (e.g. volumes/ports still detaching).
	// npc-api returns: "Cannot perform this action on the instance in current state"
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, "/cloud/instances", body)
			return err
		},
		RetryableErrors: []string{"Cannot perform this action", "in current state"},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete instance",
			fmt.Sprintf("Could not delete instance %s: %s", id, err),
		)
		return
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildCreateRequest converts the Terraform plan model into the API request
// body expected by POST /cloud/instances.
func buildCreateRequest(ctx context.Context, plan *instanceResourceModel) (*createInstanceRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := &createInstanceRequest{
		Name:                plan.Name.ValueString(),
		Count:               1,
		Flavor:              plan.FlavorID.ValueString(),
		BillingType:         plan.BillingType.ValueString(),
		BootUUID:            plan.BootUUID.ValueString(),
		SourceType:          plan.SourceType.ValueString(),
		DeleteOnTermination: plan.DeleteOnTermination.ValueBool(),
		AvailabilityZone:    plan.AvailabilityZone.ValueString(),
		ConfigDrive:         plan.ConfigDrive.ValueBool(),
	}

	// Optional scalar fields.
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.KeyName.IsNull() && !plan.KeyName.IsUnknown() {
		body.Key = plan.KeyName.ValueString()
	}
	if !plan.UserData.IsNull() && !plan.UserData.IsUnknown() {
		body.Script = plan.UserData.ValueString()
	}
	if !plan.ServerGroupID.IsNull() && !plan.ServerGroupID.IsUnknown() {
		body.ServerGroupID = plan.ServerGroupID.ValueString()
	}
	if !plan.AdminPassword.IsNull() && !plan.AdminPassword.IsUnknown() {
		body.AdminPassword = plan.AdminPassword.ValueString()
	}

	// Volumes (nested block -> list of objects).
	var volumeModels []instanceVolumeModel
	diags.Append(plan.Volumes.ElementsAs(ctx, &volumeModels, false)...)
	if diags.HasError() {
		return nil, diags
	}
	body.Volumes = make([]volumeRequest, len(volumeModels))
	for i, v := range volumeModels {
		body.Volumes[i] = volumeRequest{
			Size:        v.Size.ValueInt64(),
			Boot:        v.Boot.ValueBool(),
			VolumeType:  volumeTypeToBackend(v.VolumeType.ValueString()),
			BillingType: v.BillingType.ValueString(),
		}
	}

	// Network IDs.
	var networkIDs []string
	diags.Append(plan.NetworkIDs.ElementsAs(ctx, &networkIDs, false)...)
	if diags.HasError() {
		return nil, diags
	}
	body.Network = networkIDs

	// Security group IDs.
	var sgIDs []string
	diags.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &sgIDs, false)...)
	if diags.HasError() {
		return nil, diags
	}
	body.SecurityGroup = sgIDs

	// Metadata.
	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		metadataMap := make(map[string]string)
		diags.Append(plan.Metadata.ElementsAs(ctx, &metadataMap, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.Metadata = metadataMap
	}

	// Tags.
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		diags.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		if diags.HasError() {
			return nil, diags
		}
		body.Tags = tags
	}

	return body, diags
}

// volumeTypeToBackend maps user-friendly volume type names to the backend API names.
// Consistent with the standalone volume resource mapping.
func volumeTypeToBackend(userType string) string {
	mapping := map[string]string{
		"ssd":  "NVMe based High IOPS Storage",
		"nvme": "NVMe based High IOPS Storage",
		"hdd":  "HDD based Storage",
	}
	if backendType, ok := mapping[userType]; ok {
		return backendType
	}
	return userType
}

// volumeTypeFromBackend maps backend volume type names back to user-friendly aliases.
func volumeTypeFromBackend(backendType string) string {
	mapping := map[string]string{
		"NVMe based High IOPS Storage": "ssd",
		"HDD based Storage":            "hdd",
	}
	if userType, ok := mapping[backendType]; ok {
		return userType
	}
	return backendType
}

// mapReadResponseToState maps the API read response into the Terraform state
// model. The npc-api read response does NOT return many create-time fields
// (source_type, delete_on_termination, volumes spec, network IDs, script,
// server_group_id, admin_password), so those are preserved from the existing
// Terraform state.
func mapReadResponseToState(ctx context.Context, instance *readInstanceResponse, state *instanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(instance.ID)
	state.Name = types.StringValue(instance.Name)
	state.Status = types.StringValue(instance.Status)
	state.AvailabilityZone = types.StringValue(instance.AvailabilityZone)
	state.ConfigDrive = types.BoolValue(instance.ConfigDrive)

	// FlavorID from the nested flavor object.
	if instance.FlavorID != "" {
		state.FlavorID = types.StringValue(instance.FlavorID)
	}

	// BootUUID — use image ID from read response if available.
	if instance.ImageID != "" {
		state.BootUUID = types.StringValue(instance.ImageID)
	}
	// source_type, delete_on_termination — not returned by read, keep state values.

	// Description — preserve null if the backend returns empty string.
	if instance.Description != "" {
		state.Description = types.StringValue(instance.Description)
	} else if state.Description.IsNull() {
		state.Description = types.StringNull()
	}

	// Key name.
	if instance.Key != "" {
		state.KeyName = types.StringValue(instance.Key)
	} else if state.KeyName.IsNull() {
		state.KeyName = types.StringNull()
	}

	// user_data (script), server_group_id, admin_password — not returned by read.
	// Keep existing state values.

	// Security group IDs — read returns [{id, name}] objects, we extracted IDs.
	// Preserve user's ordering if the same set of SGs (API may return different order).
	if len(instance.SecurityGroups) > 0 {
		var existingSGs []string
		state.SecurityGroupIDs.ElementsAs(ctx, &existingSGs, false)

		sameSet := len(existingSGs) == len(instance.SecurityGroups)
		if sameSet {
			apiSet := make(map[string]bool, len(instance.SecurityGroups))
			for _, sg := range instance.SecurityGroups {
				apiSet[sg] = true
			}
			for _, sg := range existingSGs {
				if !apiSet[sg] {
					sameSet = false
					break
				}
			}
		}
		if !sameSet {
			// Different set — use API response order
			sgVals := make([]attr.Value, len(instance.SecurityGroups))
			for i, sg := range instance.SecurityGroups {
				sgVals[i] = types.StringValue(sg)
			}
			sgList, d := types.ListValue(types.StringType, sgVals)
			diags.Append(d...)
			state.SecurityGroupIDs = sgList
		}
		// Same set — keep existing order to prevent drift
	}
	// If empty, keep existing state (don't overwrite with empty list).

	// Network IDs — not directly available in read response (only in addresses).
	// Keep existing state value.

	// Metadata.
	if len(instance.Metadata) > 0 {
		metaVals := make(map[string]attr.Value, len(instance.Metadata))
		for k, v := range instance.Metadata {
			metaVals[k] = types.StringValue(v)
		}
		metaMap, d := types.MapValue(types.StringType, metaVals)
		diags.Append(d...)
		state.Metadata = metaMap
	} else if state.Metadata.IsNull() {
		state.Metadata = types.MapNull(types.StringType)
	}

	// Tags.
	if len(instance.Tags) > 0 {
		tagVals := make([]attr.Value, len(instance.Tags))
		for i, t := range instance.Tags {
			tagVals[i] = types.StringValue(t)
		}
		tagList, d := types.ListValue(types.StringType, tagVals)
		diags.Append(d...)
		state.Tags = tagList
	} else if state.Tags.IsNull() {
		state.Tags = types.ListNull(types.StringType)
	}

	// Volumes — the read response returns volumes_attached as [{id, device, name, size}]
	// which has a completely different format than the create request volumes.
	// Keep existing state value (volumes are ForceNew anyway).

	return diags
}
