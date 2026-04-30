package auto_scaling_template

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Validation patterns matching ace-cli and npc-ui.
var (
	templateNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)
	descriptionRegex  = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// autoScalingTemplateModel maps the resource schema to Go types.
type autoScalingTemplateModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	Description         types.String `tfsdk:"description"`
	VolumeSize          types.Int64  `tfsdk:"volume_size"`
	VolDelOnTermination types.Bool   `tfsdk:"vol_del_on_termination"`
	FlavorID            types.String `tfsdk:"flavor_id"`
	ImageID             types.String `tfsdk:"image_id"`
	SnapshotID          types.String `tfsdk:"snapshot_id"`
	KeyName             types.String `tfsdk:"key_name"`
	NetworkID           types.String `tfsdk:"network_id"`
	SubnetID            types.String `tfsdk:"subnet_id"`
	SecurityGroups      types.List   `tfsdk:"security_groups"`
	IsInstanceSnapshot  types.Bool   `tfsdk:"is_instance_snapshot"`
	Status              types.String `tfsdk:"status"`
	Region              types.String `tfsdk:"region"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

// autoScalingTemplateSchema returns the Terraform resource schema.
func autoScalingTemplateSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Auto Scaling template. Templates define the instance configuration used when scaling out.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the auto scaling template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the template. Must be 1-255 characters, alphanumeric with spaces, underscores, and hyphens.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(templateNameRegex, "must contain only letters, numbers, spaces, underscores, and hyphens"),
				},
			},
			"type": schema.StringAttribute{
				Description: "Operating system type. Must be 'linux' or 'windows'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("linux", "windows"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the template (3-255 characters).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 255),
					stringvalidator.RegexMatches(descriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"volume_size": schema.Int64Attribute{
				Description: "Volume size in GB for each scaled instance.",
				Required:    true,
			},
			"vol_del_on_termination": schema.BoolAttribute{
				Description: "Whether the volume is deleted when the instance is terminated.",
				Required:    true,
			},
			"flavor_id": schema.StringAttribute{
				Description: "ID of the compute flavor to use.",
				Required:    true,
			},
			"image_id": schema.StringAttribute{
				Description: "ID of the image to boot from. Required when is_instance_snapshot is false.",
				Optional:    true,
			},
			"snapshot_id": schema.StringAttribute{
				Description: "ID of the instance snapshot to boot from. Required when is_instance_snapshot is true.",
				Optional:    true,
			},
			"key_name": schema.StringAttribute{
				Description: "Name of the SSH key pair to inject.",
				Optional:    true,
			},
			"network_id": schema.StringAttribute{
				Description: "ID of the VPC network for scaled instances.",
				Required:    true,
			},
			"subnet_id": schema.StringAttribute{
				Description: "ID of the subnet for scaled instances.",
				Required:    true,
			},
			"security_groups": schema.ListAttribute{
				Description: "List of security group IDs to apply to scaled instances.",
				Required:    true,
				ElementType: types.StringType,
			},
			"is_instance_snapshot": schema.BoolAttribute{
				Description: "Whether to boot from an instance snapshot (true) or an image (false).",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the template.",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region where the template was created.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the template was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the template was last updated.",
				Computed:    true,
			},
		},
	}
}
