package instance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// defaultAvailabilityZone matches the value the platform UI hardcodes when
// creating an instance.
const defaultAvailabilityZone = "nova"

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
	AvailabilityZone    string            `json:"availability_zone"` // hardcoded by buildAPIRequest to match the platform UI default
	Metadata            map[string]string `json:"metadata,omitempty"`
	Key                 string            `json:"key,omitempty"`
	Script              string            `json:"script,omitempty"`
	AdminPassword       string            `json:"admin_password,omitempty"`
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
// - volumes_attached is [{id, device, name, size}], not the create format
// - network info is in addresses{public/private}, not a string array
type readInstanceResponse struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Status           string            `json:"status"`
	Locked           bool              `json:"locked"`
	Metadata         map[string]string `json:"metadata"`
	Key              string            `json:"key"`
	// These are extracted from the raw response in parseInstanceData()
	FlavorID       string
	ImageID        string
	SecurityGroups []string // security group IDs
	Networks       []string // network names from addresses
}

// powerStateFromStatus converts a backend instance status into the
// provider's power_state value. ACTIVE → ON; SHUTOFF/STOPPED/PAUSED/SUSPENDED → OFF.
// Other transient states (BUILD, RESIZE, …) keep the previous value so the
// state doesn't flap during async transitions.
func powerStateFromStatus(status, prev string) string {
	switch status {
	case "ACTIVE":
		return "ON"
	case "SHUTOFF", "STOPPED", "PAUSED", "SUSPENDED":
		return "OFF"
	default:
		return prev
	}
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
	// Save the user-intended values BEFORE calling mapReadResponseToState
	// (which infers power_state/locked from the just-built ACTIVE instance,
	// clobbering any non-default values the user set in their HCL config).
	wantPowerState := plan.PowerState
	wantLocked := plan.Locked

	if result != nil && result.Data != nil {
		instance := result.Data.(*readInstanceResponse)
		plan.Status = types.StringValue(instance.Status)
		mapReadResponseToState(ctx, instance, &plan)
	}

	// Restore user intent so the action checks below see the configured values.
	if !wantPowerState.IsNull() && !wantPowerState.IsUnknown() {
		plan.PowerState = wantPowerState
	}
	if !wantLocked.IsNull() && !wantLocked.IsUnknown() {
		plan.Locked = wantLocked
	}

	// Apply non-default lock state at create time.
	// `locked` defaults to false; only call /lock if user explicitly set true.
	if !plan.Locked.IsNull() && !plan.Locked.IsUnknown() && plan.Locked.ValueBool() {
		lockPath := fmt.Sprintf("/cloud/instances/%s/lock", instanceID)
		if _, err := r.client.PutWithParams(ctx, lockPath, nil, map[string]string{"value": "ON"}); err != nil {
			resp.Diagnostics.AddError("Failed to apply initial lock state", err.Error())
			return
		}
		plan.Locked = types.BoolValue(true)
	}

	// Apply non-default power_state at create time.
	// Backend creates instances ACTIVE; only call /power if user wants OFF.
	// Power transitions are async; wait for SHUTOFF so the next Refresh sees
	// the converged state and Terraform's apply consistency check passes.
	if !plan.PowerState.IsNull() && !plan.PowerState.IsUnknown() && plan.PowerState.ValueString() == "OFF" {
		powerPath := fmt.Sprintf("/cloud/instances/%s/power", instanceID)
		if _, err := r.client.PutWithParams(ctx, powerPath, nil, map[string]string{"value": "OFF"}); err != nil {
			resp.Diagnostics.AddError("Failed to apply initial power state", err.Error())
			return
		}
		off, err := r.waitForInstanceStatus(ctx, instanceID, []string{"SHUTOFF", "STOPPED", "PAUSED", "SUSPENDED"})
		if err != nil {
			resp.Diagnostics.AddError("Failed to power off instance", err.Error())
			return
		}
		plan.PowerState = types.StringValue("OFF")
		plan.Status = types.StringValue(off.Status)
	} else {
		// Always set explicit value so state isn't unknown after create.
		plan.PowerState = types.StringValue("ON")
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

	// 1. Resize (flavor change) — separate endpoint with confirm step.
	if !plan.FlavorID.Equal(state.FlavorID) {
		if err := r.resizeInstance(ctx, id, plan.FlavorID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to resize instance", err.Error())
			return
		}
	}

	// 2. Security groups change — PUT /cloud/instances/{id}/security-groups.
	if !plan.SecurityGroupIDs.Equal(state.SecurityGroupIDs) {
		var sgs []string
		resp.Diagnostics.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &sgs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		sgPath := fmt.Sprintf("/cloud/instances/%s/security-groups", id)
		if _, err := r.client.Put(ctx, sgPath, map[string][]string{"security_groups": sgs}); err != nil {
			resp.Diagnostics.AddError("Failed to update instance security groups", err.Error())
			return
		}
	}

	// 3. Lock toggle — PUT /cloud/instances/{id}/lock?value=ON|OFF.
	// Lock state must be applied BEFORE any operation that mutates the
	// instance (e.g. power_state) since locked instances refuse other
	// actions. Lock is fast and synchronous from the user's perspective.
	if !plan.Locked.IsNull() && !plan.Locked.IsUnknown() && !plan.Locked.Equal(state.Locked) {
		val := "OFF"
		if plan.Locked.ValueBool() {
			val = "ON"
		}
		lockPath := fmt.Sprintf("/cloud/instances/%s/lock", id)
		if _, err := r.client.PutWithParams(ctx, lockPath, nil, map[string]string{"value": val}); err != nil {
			resp.Diagnostics.AddError("Failed to update instance lock state", err.Error())
			return
		}
	}

	// 4. Power state — PUT /cloud/instances/{id}/power?value=ON|OFF.
	// Async; wait for the instance to reach ACTIVE/SHUTOFF before returning
	// so subsequent refreshes don't see stale status.
	if !plan.PowerState.IsNull() && !plan.PowerState.IsUnknown() && !plan.PowerState.Equal(state.PowerState) {
		powerPath := fmt.Sprintf("/cloud/instances/%s/power", id)
		val := plan.PowerState.ValueString() // already validated as ON or OFF
		if _, err := r.client.PutWithParams(ctx, powerPath, nil, map[string]string{"value": val}); err != nil {
			resp.Diagnostics.AddError("Failed to update instance power state", err.Error())
			return
		}
		targets := []string{"SHUTOFF", "STOPPED", "PAUSED", "SUSPENDED"}
		if val == "ON" {
			targets = []string{"ACTIVE"}
		}
		if _, err := r.waitForInstanceStatus(ctx, id, targets); err != nil {
			resp.Diagnostics.AddError("Failed to update instance power state", err.Error())
			return
		}
	}

	// 5. Generic PUT for name + description.
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

// waitForInstanceStatus polls until the instance reports one of the target
// statuses. Used after async actions (power, lock, resize) where the API call
// returns immediately but the state machine takes a few seconds to converge.
func (r *instanceResource) waitForInstanceStatus(ctx context.Context, instanceID string, targets []string) (*readInstanceResponse, error) {
	result, err := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			apiResp, err := r.client.Get(ctx, fmt.Sprintf("/cloud/instances/%s", instanceID), nil)
			if err != nil {
				return nil, err
			}
			inst, perr := parseInstanceData(apiResp.Data)
			if perr != nil {
				return nil, perr
			}
			return &wait.StatusResult{Status: inst.Status, Data: inst}, nil
		},
		TargetStatus: targets,
		ErrorStatus:  []string{"ERROR"},
		Timeout:      3 * time.Minute,
		PollInterval: 4 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	if result == nil || result.Data == nil {
		return nil, nil
	}
	return result.Data.(*readInstanceResponse), nil
}

// resizeInstance performs the two-step resize flow:
//  1. PUT /cloud/instances/{id}/resize {flavor_id: ...} → instance enters
//     RESIZE → VERIFY_RESIZE
//  2. PUT /cloud/instances/{id}/resize-action?value=CONFIRM → instance returns
//     to ACTIVE on the new flavor
//
// Note: backend support for resize requires the cluster to have multiple
// compute hosts (Nova's same-host resize is disabled by default). On
// single-host preprod clusters, the initiate call succeeds but the instance
// bounces back to ACTIVE on the original flavor — there is nothing the
// provider can do about that.
func (r *instanceResource) resizeInstance(ctx context.Context, instanceID, newFlavorID string) error {
	initiate := fmt.Sprintf("/cloud/instances/%s/resize", instanceID)
	body := resizeRequest{FlavorID: newFlavorID}

	if _, err := r.client.Put(ctx, initiate, body); err != nil {
		return fmt.Errorf("initiate resize: %w", err)
	}

	// Wait for VERIFY_RESIZE; tolerate transient ACTIVE while the resize
	// transitions through RESIZE state.
	verifyResult, err := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			apiResp, err := r.client.Get(ctx, fmt.Sprintf("/cloud/instances/%s", instanceID), nil)
			if err != nil {
				return nil, err
			}
			inst, perr := parseInstanceData(apiResp.Data)
			if perr != nil {
				return nil, perr
			}
			return &wait.StatusResult{Status: inst.Status, Data: inst}, nil
		},
		TargetStatus: []string{"VERIFY_RESIZE", "ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      5 * time.Minute,
		PollInterval: 5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("wait for VERIFY_RESIZE: %w", err)
	}

	// If the instance went straight to ACTIVE, the resize was either auto-
	// confirmed by the backend or rejected silently. Verify the flavor
	// actually changed; if not, surface a clear error.
	if verifyResult.Status == "ACTIVE" {
		inst, _ := verifyResult.Data.(*readInstanceResponse)
		if inst != nil && inst.FlavorID != "" && inst.FlavorID != newFlavorID {
			return fmt.Errorf("instance returned to ACTIVE on the original flavor — backend did not perform the resize. This typically means the cluster does not have multiple compute hosts available for resize")
		}
		return nil // backend auto-confirmed; we're done
	}

	// Confirm the resize.
	confirmPath := fmt.Sprintf("/cloud/instances/%s/resize-action", instanceID)
	if _, err := r.client.PutWithParams(ctx, confirmPath, nil, map[string]string{"value": "CONFIRM"}); err != nil {
		return fmt.Errorf("confirm resize: %w", err)
	}

	// Wait for ACTIVE on the new flavor.
	_, err = wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			apiResp, err := r.client.Get(ctx, fmt.Sprintf("/cloud/instances/%s", instanceID), nil)
			if err != nil {
				return nil, err
			}
			inst, perr := parseInstanceData(apiResp.Data)
			if perr != nil {
				return nil, perr
			}
			return &wait.StatusResult{Status: inst.Status, Data: inst}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      5 * time.Minute,
		PollInterval: 5 * time.Second,
	})
	return err
}

// resizeRequest is the body for PUT /cloud/instances/{id}/resize.
type resizeRequest struct {
	FlavorID string `json:"flavor_id"`
}

// Delete removes the instance via DELETE /cloud/instances.
func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	// Locked instances reject delete; unlock first.
	if !state.Locked.IsNull() && state.Locked.ValueBool() {
		lockPath := fmt.Sprintf("/cloud/instances/%s/lock", id)
		if _, err := r.client.PutWithParams(ctx, lockPath, nil, map[string]string{"value": "OFF"}); err != nil {
			resp.Diagnostics.AddError("Failed to unlock instance before delete", err.Error())
			return
		}
	}

	// Powered-off instances on this backend cannot detach ports cleanly,
	// which causes the subsequent VPC/SG delete to fail with "ports still
	// in use". Power the instance back on briefly before issuing delete.
	if !state.PowerState.IsNull() && state.PowerState.ValueString() == "OFF" {
		powerPath := fmt.Sprintf("/cloud/instances/%s/power", id)
		if _, err := r.client.PutWithParams(ctx, powerPath, nil, map[string]string{"value": "ON"}); err == nil {
			// Best effort wait for ACTIVE; ignore timeout — delete will retry.
			_, _ = r.waitForInstanceStatus(ctx, id, []string{"ACTIVE"})
		}
	}

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
		RetryableErrors: []string{"Cannot perform this action", "in current state", "PENDING_"},
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
		AvailabilityZone:    defaultAvailabilityZone,
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

// mapReadResponseToState maps the API read response into the Terraform state
// model. The npc-api read response does NOT return many create-time fields
// (source_type, delete_on_termination, volumes spec, network IDs, script,
// admin_password), so those are preserved from the existing Terraform state.
func mapReadResponseToState(ctx context.Context, instance *readInstanceResponse, state *instanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(instance.ID)
	state.Name = types.StringValue(instance.Name)
	state.Status = types.StringValue(instance.Status)

	// power_state derived from instance status; locked mirrored from API.
	state.PowerState = types.StringValue(powerStateFromStatus(instance.Status, state.PowerState.ValueString()))
	state.Locked = types.BoolValue(instance.Locked)

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

	// user_data (script) and admin_password — not returned by read.
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

	// Volumes — the read response returns volumes_attached as [{id, device, name, size}]
	// which has a completely different format than the create request volumes.
	// Keep existing state value (volumes are ForceNew anyway).

	return diags
}
