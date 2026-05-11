// Package volume_attachment implements the acecloud_volume_attachment Terraform
// resource, which attaches an existing volume to an existing instance via
// `PUT /cloud/instances/{instance_id}/attach-volume` and detaches via
// `PUT /cloud/instances/{instance_id}/detach-volume/{volume_id}`.
package volume_attachment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	volumesPath   = "/cloud/volumes"
	instancesPath = "/cloud/instances"
)

var (
	_ resource.Resource              = &volumeAttachmentResource{}
	_ resource.ResourceWithConfigure = &volumeAttachmentResource{}
)

type volumeAttachmentResource struct {
	client *client.Client
}

// NewResource is the factory used by the provider registration.
func NewResource() resource.Resource {
	return &volumeAttachmentResource{}
}

func (r *volumeAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_attachment"
}

func (r *volumeAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema()
}

func (r *volumeAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

// attachRequest is the body for PUT /cloud/instances/{id}/attach-volume.
type attachRequest struct {
	VolumeID            string `json:"volume_id"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
}

// volumeReadResponse mirrors the subset of the volume GET response we care
// about for attachment status.
type volumeReadResponse struct {
	ID          string             `json:"id"`
	Status      string             `json:"status"`
	Attachments []volumeAttachment `json:"attachments"`
}

type volumeAttachment struct {
	ServerID string `json:"server_id"`
	Device   string `json:"device"`
}

func (r *volumeAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan volumeAttachmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := attachRequest{
		VolumeID:            plan.VolumeID.ValueString(),
		DeleteOnTermination: plan.DeleteOnTermination.ValueBool(),
	}
	path := fmt.Sprintf("%s/%s/attach-volume", instancesPath, plan.InstanceID.ValueString())

	// Retry on transient 409s (volume creating, instance pending, etc.)
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Put(ctx, path, body)
			return err
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to attach volume", err.Error())
		return
	}

	// Poll the volume's status until it reaches "in-use" with a matching
	// attachment record. Backend attaches asynchronously.
	device, err := waitForAttachment(ctx, r.client, plan.VolumeID.ValueString(), plan.InstanceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to attach volume", err.Error())
		return
	}

	plan.ID = types.StringValue(buildID(plan.InstanceID.ValueString(), plan.VolumeID.ValueString()))
	plan.Device = types.StringValue(device)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *volumeAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state volumeAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vol, err := readVolume(ctx, r.client, state.VolumeID.ValueString())
	if err != nil {
		// Volume gone → attachment gone.
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read volume", err.Error())
		return
	}

	// Detect detach drift: if the volume is no longer attached to the
	// expected instance, remove the resource from state.
	device := ""
	for _, a := range vol.Attachments {
		if a.ServerID == state.InstanceID.ValueString() {
			device = a.Device
			break
		}
	}
	if device == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Device = types.StringValue(device)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *volumeAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All writable attributes use RequiresReplace, so Update should never
	// be called. If the framework ever calls it, surface a clear error.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"All attributes on acecloud_volume_attachment are immutable; changes trigger replacement.",
	)
}

func (r *volumeAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state volumeAttachmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s/detach-volume/%s",
		instancesPath,
		state.InstanceID.ValueString(),
		state.VolumeID.ValueString(),
	)
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Put(ctx, path, nil)
			return err
		},
		RetryableErrors: []string{"in use", "PENDING_", "is busy", "currently in use", "Please try again"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to detach volume", err.Error())
	}
}

// waitForAttachment polls the volume until the expected instance appears in
// its attachments[] list, returning the device path. Times out after 5 min.
func waitForAttachment(ctx context.Context, c *client.Client, volumeID, instanceID string) (string, error) {
	var device string
	_, err := wait.PollForResource(ctx, wait.PollForResourceOpts{
		Timeout:      5 * time.Minute,
		PollInterval: 3 * time.Second,
		List: func(ctx context.Context) (interface{}, error) {
			vol, err := readVolume(ctx, c, volumeID)
			if err != nil {
				return nil, err
			}
			for _, a := range vol.Attachments {
				if a.ServerID == instanceID {
					device = a.Device
					return vol, nil
				}
			}
			return nil, nil
		},
	})
	if err != nil {
		return "", err
	}
	return device, nil
}

func readVolume(ctx context.Context, c *client.Client, volumeID string) (*volumeReadResponse, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("%s/%s", volumesPath, volumeID), nil)
	if err != nil {
		return nil, err
	}
	var v volumeReadResponse
	if err := json.Unmarshal(resp.Data, &v); err != nil {
		return nil, fmt.Errorf("parse volume response: %w", err)
	}
	return &v, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "404") || strings.Contains(strings.ToLower(s), "not found")
}

// buildID composes the Terraform resource ID from instance + volume.
func buildID(instanceID, volumeID string) string {
	return fmt.Sprintf("%s:%s", instanceID, volumeID)
}
