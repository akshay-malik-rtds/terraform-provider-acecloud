package volume_attachment

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// volumeAttachmentResourceModel is the Terraform plan/state model.
type volumeAttachmentResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	InstanceID          types.String `tfsdk:"instance_id"`
	VolumeID            types.String `tfsdk:"volume_id"`
	DeleteOnTermination types.Bool   `tfsdk:"delete_on_termination"`
	Device              types.String `tfsdk:"device"`
}

func resourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Attaches an existing volume to an existing instance. Use this when you need to attach a volume to an instance that was created separately (e.g. attaching a data volume to a long-running instance after-the-fact). For volumes attached at instance creation time, use the `volumes` block on `acecloud_instance` instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite identifier `<instance_id>:<volume_id>`.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"instance_id": schema.StringAttribute{
				Description: "UUID of the instance to attach the volume to. Changing this forces re-attachment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"volume_id": schema.StringAttribute{
				Description: "UUID of the volume to attach. Changing this forces re-attachment.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delete_on_termination": schema.BoolAttribute{
				Description: "Whether the volume should be deleted automatically when the instance terminates. Defaults to `false`. Changing this forces re-attachment.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"device": schema.StringAttribute{
				Description: "Device path on the guest (e.g. `/dev/vdb`) once the attachment completes.",
				Computed:    true,
			},
		},
	}
}
