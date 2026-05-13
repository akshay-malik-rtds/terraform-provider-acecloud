package snapshot

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &snapshotResource{}
	_ resource.ResourceWithConfigure = &snapshotResource{}
)

type snapshotResource struct {
	client *client.Client
}

// --- API types ---

type snapshotCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	VolumeID    string `json:"volume_id"`
}

type snapshotUpdateRequest struct {
	Snapshot snapshotUpdateBody `json:"snapshot"`
}

type snapshotUpdateBody struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type snapshotAPIResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	VolumeID    string `json:"volume_id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Size        int64  `json:"size"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type snapshotDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &snapshotResource{}
}

func (r *snapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (r *snapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = snapshotSchema()
}

func (r *snapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(plan *snapshotResourceModel) snapshotCreateRequest {
	body := snapshotCreateRequest{
		Name:     plan.Name.ValueString(),
		VolumeID: plan.VolumeID.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	return body
}

func mapAPIResponseToState(model *snapshotResourceModel, apiResp *snapshotAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.VolumeID != "" {
		model.VolumeID = types.StringValue(apiResp.VolumeID)
	}

	if apiResp.Description != "" {
		// Only update from API if user set a non-empty description.
		// If user cleared description to "" or never set it, leave it alone.
		if !model.Description.IsNull() && model.Description.ValueString() != "" {
			model.Description = types.StringValue(apiResp.Description)
		}
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	model.Status = types.StringValue(apiResp.Status)
	model.Size = types.Int64Value(apiResp.Size)
	model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	model.UpdatedAt = types.StringValue(apiResp.UpdatedAt)
}

// parseSnapshotResponse handles both wrapped {"snapshot": {...}} and direct {...} response formats.
func parseSnapshotResponse(data json.RawMessage) (*snapshotAPIResponse, error) {
	// Try wrapped format first: {"snapshot": {...}}
	var wrapped struct {
		Snapshot snapshotAPIResponse `json:"snapshot"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Snapshot.ID != "" {
		return &wrapped.Snapshot, nil
	}

	// Fallback to direct format
	var direct snapshotAPIResponse
	if err := json.Unmarshal(data, &direct); err != nil {
		return nil, fmt.Errorf("failed to parse snapshot response: %w", err)
	}
	return &direct, nil
}

func (r *snapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(&plan)

	// Snapshot create uses: POST /cloud/volumes/{volume_id}/snapshots
	createPath := fmt.Sprintf("/cloud/volumes/%s/snapshots", plan.VolumeID.ValueString())
	apiResp, err := r.client.Post(ctx, createPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create snapshot", err.Error())
		return
	}

	created, err := parseSnapshotResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	mapAPIResponseToState(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state snapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read uses the storage API: GET snapshots/{id}
	readPath := fmt.Sprintf("/os/cinder/%s/snapshots/%s", r.client.ProjectID, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, readPath, map[string]string{
		"with": `["volume"]`,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to read snapshot", err.Error())
		return
	}

	snapshot, err := parseSnapshotResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	mapAPIResponseToState(&state, snapshot)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *snapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan snapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state snapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update uses the storage API: PUT snapshots/{id}
	updateBody := snapshotUpdateRequest{
		Snapshot: snapshotUpdateBody{
			Name: plan.Name.ValueString(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateBody.Snapshot.Description = plan.Description.ValueString()
	}

	updatePath := fmt.Sprintf("/os/cinder/%s/snapshots/%s", r.client.ProjectID, state.ID.ValueString())
	apiResp, err := r.client.Put(ctx, updatePath, updateBody)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update snapshot", err.Error())
		return
	}

	updated, err := parseSnapshotResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}

	plannedDescription := plan.Description
	mapAPIResponseToState(&plan, updated)

	// Preserve plan description — omitempty skips empty string, so API may return old value
	if !plannedDescription.IsNull() && !plannedDescription.IsUnknown() {
		plan.Description = plannedDescription
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete uses the storage API: DELETE snapshots
	deletePath := fmt.Sprintf("/os/cinder/%s/snapshots", r.client.ProjectID)
	body := snapshotDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	// Snapshot deletion can fail when the source volume is busy or the
	// snapshot is being used by another operation.
	// npc-api returns: "...action not possible due to some process already using the..."
	// npc-api returns: "...status must be available, but current status is in-use"
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, deletePath, body)
			return err
		},
		RetryableErrors: []string{"in use", "status must be available", "already using"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete snapshot", err.Error())
		return
	}
}
