package floating_ip_association

import (
	"context"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/floating-ips/action"

var (
	_ resource.Resource              = &floatingIPAssociationResource{}
	_ resource.ResourceWithConfigure = &floatingIPAssociationResource{}
)

type floatingIPAssociationResource struct {
	client *client.Client
}

// --- API types (matching ace-cli/internal/api/floating_ip.go) ---

type associateRequest struct {
	FloatingIPAddress string `json:"floating_ip_address"`
	InstanceID        string `json:"instance_id"`
	FixedIPAddress    string `json:"fixed_ip_address,omitempty"`
}

type disassociateRequest struct {
	FloatingIPAddress string `json:"floating_ip_address"`
	InstanceID        string `json:"instance_id"`
}

func NewResource() resource.Resource {
	return &floatingIPAssociationResource{}
}

func (r *floatingIPAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip_association"
}

func (r *floatingIPAssociationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = floatingIPAssociationSchema()
}

func (r *floatingIPAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *floatingIPAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan floatingIPAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := associateRequest{
		FloatingIPAddress: plan.FloatingIPAddress.ValueString(),
		InstanceID:        plan.InstanceID.ValueString(),
	}
	if !plan.FixedIPAddress.IsNull() && !plan.FixedIPAddress.IsUnknown() {
		body.FixedIPAddress = plan.FixedIPAddress.ValueString()
	}

	// PUT /cloud/floating-ips/action?type=attach
	_, err := r.client.PutWithParams(ctx, apiPath, body, map[string]string{
		"type": "attach",
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to associate floating IP", err.Error())
		return
	}

	// Composite ID: floating_ip_address/instance_id
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.FloatingIPAddress.ValueString(), plan.InstanceID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *floatingIPAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state floatingIPAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No direct read API for association — the state is maintained by Terraform.
	// We could verify by reading the floating IP and checking its status,
	// but the association is managed as a side effect.
	// Keep existing state as-is.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *floatingIPAssociationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Associations do not support update. Changes trigger destroy and recreate.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Floating IP association does not support in-place updates. Changes will trigger a destroy and recreate.",
	)
}

func (r *floatingIPAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state floatingIPAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := disassociateRequest{
		FloatingIPAddress: state.FloatingIPAddress.ValueString(),
		InstanceID:        state.InstanceID.ValueString(),
	}

	// FIP detach can fail transiently when the instance port is in a
	// transitional state during concurrent destroy operations.
	// npc-api returns: "Cannot perform this action on the instance in current state"
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			// PUT /cloud/floating-ips/action?type=detach
			_, err := r.client.PutWithParams(ctx, apiPath, body, map[string]string{
				"type": "detach",
			})
			return err
		},
		RetryableErrors: []string{"Cannot perform this action", "in current state"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to disassociate floating IP", err.Error())
		return
	}
}
