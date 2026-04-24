package volume

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const volumeBasePath = "/cloud/volumes"

// Ensure volumeResource satisfies the Resource interface.
var (
	_ resource.Resource              = &volumeResource{}
	_ resource.ResourceWithConfigure = &volumeResource{}
)

// volumeResource implements the acecloud_volume resource.
type volumeResource struct {
	client *client.Client
}

// NewResource returns a factory function for the volume resource.
func NewResource() resource.Resource {
	return &volumeResource{}
}

func (r *volumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *volumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = volumeSchema()
}

func (r *volumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// volumeAPIRequest is the request body sent to the Ace Cloud API.
type volumeAPIRequest struct {
	Name             string            `json:"name"`
	Size             int64             `json:"size"`
	VolumeType       string            `json:"volume_type"`
	BillingType      string            `json:"billing_type"`
	AvailabilityZone string            `json:"availability_zone,omitempty"`
	Description      string            `json:"description,omitempty"`
	SourceVolID      string            `json:"source_volid,omitempty"`
	SnapshotID       string            `json:"snapshot_id,omitempty"`
	BackupID         string            `json:"backup_id,omitempty"`
	ImageRef         string            `json:"image_ref,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// volumeAPIResponse is the response body returned from the Ace Cloud API.
type volumeAPIResponse struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Size             int64             `json:"size"`
	VolumeType       string            `json:"volume_type"`
	BillingType      string            `json:"billing_type"`
	AvailabilityZone string            `json:"availability_zone"`
	Description      string            `json:"description"`
	SourceVolID      string            `json:"source_volid"`
	SnapshotID       string            `json:"snapshot_id"`
	BackupID         string            `json:"backup_id"`
	ImageRef         string            `json:"image_ref"`
	Metadata         map[string]string `json:"metadata"`
	Status           string            `json:"status"`
}

// volumeDeleteRequest is the request body for volume deletion.
// Matches CLI pattern: {"key": "id", "values": [...]}
type volumeDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func (r *volumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildAPIRequest(&plan, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Post(ctx, volumeBasePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create volume", err.Error())
		return
	}

	var created volumeAPIResponse
	if err := json.Unmarshal(apiResp.Data, &created); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	mapAPIResponseToState(&plan, &created, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", volumeBasePath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read volume", err.Error())
		return
	}

	var volume volumeAPIResponse
	if err := json.Unmarshal(apiResp.Data, &volume); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	// Preserve expected size if async resize hasn't completed yet
	expectedSize := state.Size.ValueInt64()

	mapAPIResponseToState(&state, &volume, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If state had a larger size (pending resize), keep it — the API returns
	// the old size until resize completes
	if expectedSize > state.Size.ValueInt64() {
		state.Size = types.Int64Value(expectedSize)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan volumeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state volumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildAPIRequest(&plan, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", volumeBasePath, state.ID.ValueString())
	apiResp, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update volume", err.Error())
		return
	}

	var updated volumeAPIResponse
	if err := json.Unmarshal(apiResp.Data, &updated); err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}

	// Save plan values before mapping — API response may differ from plan
	plannedSize := plan.Size
	plannedMetadata := plan.Metadata
	plannedDescription := plan.Description

	mapAPIResponseToState(&plan, &updated, ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// If the user requested a size change, keep the planned value.
	// The async resize will eventually converge; the next Read will pick
	// up the final size.
	if !plannedSize.Equal(state.Size) {
		plan.Size = plannedSize
	}

	// If the user explicitly set metadata (even to empty {}), keep the plan value.
	// The API may return old metadata that wasn't cleared, but the plan is the
	// user's desired state.
	if !plannedMetadata.IsNull() && !plannedMetadata.IsUnknown() {
		plan.Metadata = plannedMetadata
	}

	// If the user explicitly set description (even to ""), keep the plan value.
	// The API omitempty skips empty string so old description persists on backend.
	if !plannedDescription.IsNull() && !plannedDescription.IsUnknown() {
		plan.Description = plannedDescription
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := volumeDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	// Volume deletion can fail transiently when the volume is still attached
	// or a snapshot/backup operation is in progress.
	// npc-api returns: "Volume either already in use or volume snapshot is available"
	// npc-api returns: "...is either attached to an instance or is currently in use"
	// npc-api returns: "...status must be available to perform the action"
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, volumeBasePath, body)
			return err
		},
		RetryableErrors: []string{"in use", "attached to an instance", "status must be available"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete volume", err.Error())
		return
	}
}

// volumeTypeToBackend maps user-friendly volume type names to the backend API names.
// Users can specify either the short name or the full backend name.
func volumeTypeToBackend(userType string) string {
	mapping := map[string]string{
		"ssd":  "NVMe based High IOPS Storage",
		"nvme": "NVMe based High IOPS Storage",
		"hdd":  "HDD based Storage",
	}
	if backendType, ok := mapping[userType]; ok {
		return backendType
	}
	// If not in mapping, assume user provided the exact backend name.
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

// buildAPIRequest constructs the API request body from a Terraform plan model.
func buildAPIRequest(plan *volumeResourceModel, ctx context.Context, diags *diag.Diagnostics) volumeAPIRequest {
	body := volumeAPIRequest{
		Name:        plan.Name.ValueString(),
		Size:        plan.Size.ValueInt64(),
		VolumeType:  volumeTypeToBackend(plan.VolumeType.ValueString()),
		BillingType: plan.BillingType.ValueString(),
	}

	if !plan.AvailabilityZone.IsNull() && !plan.AvailabilityZone.IsUnknown() {
		body.AvailabilityZone = plan.AvailabilityZone.ValueString()
	}
	if !plan.Description.IsNull() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.SourceVolID.IsNull() {
		body.SourceVolID = plan.SourceVolID.ValueString()
	}
	if !plan.SnapshotID.IsNull() {
		body.SnapshotID = plan.SnapshotID.ValueString()
	}
	if !plan.BackupID.IsNull() {
		body.BackupID = plan.BackupID.ValueString()
	}
	if !plan.ImageRef.IsNull() {
		body.ImageRef = plan.ImageRef.ValueString()
	}
	if !plan.Metadata.IsNull() {
		metadata := make(map[string]string)
		diags.Append(plan.Metadata.ElementsAs(ctx, &metadata, false)...)
		body.Metadata = metadata
	}

	return body
}

// mapAPIResponseToState maps an API response to the Terraform state model.
func mapAPIResponseToState(model *volumeResourceModel, apiResp *volumeAPIResponse, ctx context.Context, diags *diag.Diagnostics) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)
	model.Size = types.Int64Value(apiResp.Size)
	// Keep the user's volume_type value if it maps to the backend name.
	// This prevents unnecessary diffs (user writes "ssd", backend returns "NVMe based High IOPS Storage").
	if model.VolumeType.ValueString() != "" && volumeTypeToBackend(model.VolumeType.ValueString()) == apiResp.VolumeType {
		// Keep existing value — no change needed.
	} else {
		model.VolumeType = types.StringValue(volumeTypeFromBackend(apiResp.VolumeType))
	}
	// billing_type is a write-only field — the API does not return it.
	// Preserve the user's configured value so terraform state stays consistent.
	if apiResp.BillingType != "" {
		model.BillingType = types.StringValue(apiResp.BillingType)
	}
	// If apiResp.BillingType is empty, keep model.BillingType as-is (from plan).
	model.Status = types.StringValue(apiResp.Status)

	if apiResp.AvailabilityZone != "" {
		model.AvailabilityZone = types.StringValue(apiResp.AvailabilityZone)
	}

	if apiResp.Description != "" {
		// Only update from API if user originally set a non-empty description.
		// If user explicitly set description="" (clearing it), preserve "".
		if !model.Description.IsNull() && model.Description.ValueString() != "" {
			model.Description = types.StringValue(apiResp.Description)
		} else if model.Description.IsNull() {
			// User never set description; API has a value — keep null
		}
		// If model.Description == "", user explicitly cleared it — keep ""
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	// For source fields (source_volid, snapshot_id, backup_id, image_ref):
	// Only update from API if the user originally set the field.
	// The API may return values the user didn't provide (e.g., source_volid when
	// creating from snapshot_id), which causes Terraform "inconsistent result" errors.
	if !model.SourceVolID.IsNull() {
		if apiResp.SourceVolID != "" {
			model.SourceVolID = types.StringValue(apiResp.SourceVolID)
		}
	}

	if !model.SnapshotID.IsNull() {
		if apiResp.SnapshotID != "" {
			model.SnapshotID = types.StringValue(apiResp.SnapshotID)
		}
	}

	if !model.BackupID.IsNull() {
		if apiResp.BackupID != "" {
			model.BackupID = types.StringValue(apiResp.BackupID)
		}
	}

	if !model.ImageRef.IsNull() {
		if apiResp.ImageRef != "" {
			model.ImageRef = types.StringValue(apiResp.ImageRef)
		}
	}

	// For metadata: only update from API if user provided non-empty metadata.
	// API may inject metadata the user didn't set (e.g., src_backup_id when
	// restoring from backup), causing "inconsistent result" errors.
	// If user set metadata = {} (empty map), preserve it — don't overwrite with API data.
	if !model.Metadata.IsNull() && len(model.Metadata.Elements()) > 0 {
		if len(apiResp.Metadata) > 0 {
			metadataMap, d := types.MapValueFrom(ctx, types.StringType, apiResp.Metadata)
			diags.Append(d...)
			model.Metadata = metadataMap
		}
	}
}

// uuidRegex returns a compiled regexp for UUID validation.
func uuidRegex() *regexp.Regexp {
	return regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
}
