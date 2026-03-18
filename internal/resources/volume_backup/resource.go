package volume_backup

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &volumeBackupResource{}
	_ resource.ResourceWithConfigure = &volumeBackupResource{}
)

type volumeBackupResource struct {
	client *client.Client
}

// --- API types (matching npc-ui useBackup.ts) ---

type backupCreateWrapper struct {
	Backup backupCreateBody `json:"backup"`
}

type backupCreateBody struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SnapshotID  string `json:"snapshot_id,omitempty"`
	VolumeID    string `json:"volume_id"`
	Incremental bool   `json:"incremental"`
}

type backupUpdateWrapper struct {
	Backup backupUpdateBody `json:"backup"`
}

type backupUpdateBody struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type backupAPIResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	VolumeID      string `json:"volume_id"`
	SnapshotID    string `json:"snapshot_id"`
	Description   string `json:"description"`
	Incremental   bool   `json:"incremental"`
	IsIncremental bool   `json:"is_incremental"`
	Status        string `json:"status"`
	Size          int64  `json:"size"`
	CreatedAt     string `json:"created_at"`
}

type backupDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &volumeBackupResource{}
}

func (r *volumeBackupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_backup"
}

func (r *volumeBackupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = volumeBackupSchema()
}

func (r *volumeBackupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(plan *volumeBackupResourceModel) backupCreateWrapper {
	body := backupCreateWrapper{
		Backup: backupCreateBody{
			Name:        plan.Name.ValueString(),
			VolumeID:    plan.VolumeID.ValueString(),
			Incremental: plan.Incremental.ValueBool(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Backup.Description = plan.Description.ValueString()
	}
	if !plan.SnapshotID.IsNull() && !plan.SnapshotID.IsUnknown() {
		body.Backup.SnapshotID = plan.SnapshotID.ValueString()
	}
	return body
}

func mapAPIResponseToState(model *volumeBackupResourceModel, apiResp *backupAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.VolumeID != "" {
		model.VolumeID = types.StringValue(apiResp.VolumeID)
	}

	if apiResp.SnapshotID != "" {
		model.SnapshotID = types.StringValue(apiResp.SnapshotID)
	} else if model.SnapshotID.IsNull() {
		model.SnapshotID = types.StringNull()
	}

	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	model.Incremental = types.BoolValue(apiResp.Incremental || apiResp.IsIncremental)
	model.Status = types.StringValue(apiResp.Status)
	model.Size = types.Int64Value(apiResp.Size)
	model.CreatedAt = types.StringValue(apiResp.CreatedAt)
}

// parseBackupResponse handles both wrapped {"backup": {...}} and direct {...} response formats.
func parseBackupResponse(data json.RawMessage) (*backupAPIResponse, error) {
	var wrapped struct {
		Backup backupAPIResponse `json:"backup"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Backup.ID != "" {
		return &wrapped.Backup, nil
	}

	var direct backupAPIResponse
	if err := json.Unmarshal(data, &direct); err != nil {
		return nil, fmt.Errorf("failed to parse backup response: %w", err)
	}
	return &direct, nil
}

func (r *volumeBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(&plan)

	// Backup create uses: POST /backups (npc-api routes backups at /backups)
	createPath := "/backups"
	apiResp, err := r.client.Post(ctx, createPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create volume backup", err.Error())
		return
	}

	created, err := parseBackupResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	if created.ID == "" {
		resp.Diagnostics.AddError("Failed to create volume backup", "API returned empty ID")
		return
	}

	// Set ID immediately so we can track the resource even if wait fails
	plan.ID = types.StringValue(created.ID)

	// Wait for backup to become available (backups take ~2 min)
	// Use direct /backups/{id} endpoint — the cloud wrapper returns {} while backup is creating
	waitPath := fmt.Sprintf("/backups/%s", created.ID)
	result, waitErr := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			readResp, err := r.client.Get(ctx, waitPath, nil)
			if err != nil {
				return nil, err
			}
			backup, err := parseBackupResponse(readResp.Data)
			if err != nil {
				return nil, err
			}
			if backup.ID == "" {
				return &wait.StatusResult{Status: "creating"}, nil
			}
			return &wait.StatusResult{
				Status: backup.Status,
				Data:   backup,
			}, nil
		},
		TargetStatus: []string{"available"},
		ErrorStatus:  []string{"error"},
		Timeout:      5 * time.Minute,
	})

	if waitErr != nil {
		// Non-fatal: set what we have from create response
		mapAPIResponseToState(&plan, created)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddWarning("Volume backup still creating", fmt.Sprintf("Backup created but not yet available: %s", waitErr.Error()))
		return
	}

	if backup, ok := result.Data.(*backupAPIResponse); ok {
		mapAPIResponseToState(&plan, backup)
	} else {
		mapAPIResponseToState(&plan, created)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read uses: GET /cloud/volume-backups/{id} (npc-api cloud wrapper, matches npc-ui)
	readPath := fmt.Sprintf("/cloud/volume-backups/%s", state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, readPath, nil)
	if err != nil {
		// If 404, the resource no longer exists
		resp.State.RemoveResource(ctx)
		return
	}

	backup, err := parseBackupResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	// If backup ID is empty, the cloud wrapper returned {} (e.g. backup still initializing).
	// Keep existing state rather than removing — the resource was created successfully.
	if backup.ID == "" {
		return
	}

	mapAPIResponseToState(&state, backup)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan volumeBackupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state volumeBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateBody := backupUpdateWrapper{
		Backup: backupUpdateBody{
			Name: plan.Name.ValueString(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateBody.Backup.Description = plan.Description.ValueString()
	}

	// Update uses: PUT /backups/{id} (npc-api routes backups at /backups)
	updatePath := fmt.Sprintf("/backups/%s", state.ID.ValueString())
	apiResp, err := r.client.Put(ctx, updatePath, updateBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update volume backup", err.Error())
		return
	}

	updated, err := parseBackupResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}

	mapAPIResponseToState(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeBackupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete uses: DELETE /backups (npc-api routes backups at /backups)
	deletePath := "/backups"
	body := backupDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	_, err := r.client.Delete(ctx, deletePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete volume backup", err.Error())
		return
	}
}
