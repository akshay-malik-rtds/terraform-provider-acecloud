package volume_backup

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	backupNameRegex        = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	backupDescriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// volumeBackupResourceModel maps the resource schema to a Go struct.
type volumeBackupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	VolumeID    types.String `tfsdk:"volume_id"`
	SnapshotID  types.String `tfsdk:"snapshot_id"`
	Description types.String `tfsdk:"description"`
	Incremental types.Bool   `tfsdk:"incremental"`
	Status      types.String `tfsdk:"status"`
	Size        types.Int64  `tfsdk:"size"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func volumeBackupSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud volume backup.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the volume backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the backup. Must be alphanumeric with hyphens.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(backupNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"volume_id": schema.StringAttribute{
				Description: "ID of the volume to back up.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"snapshot_id": schema.StringAttribute{
				Description: "ID of a snapshot to create the backup from (optional).",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the backup.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(backupDescriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"incremental": schema.BoolAttribute{
				Description: "Whether this is an incremental backup. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"status": schema.StringAttribute{
				Description: "Current status of the backup (e.g. available, creating).",
				Computed:    true,
			},
			"size": schema.Int64Attribute{
				Description: "Size of the backup in GB.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the backup was created.",
				Computed:    true,
			},
		},
	}
}
