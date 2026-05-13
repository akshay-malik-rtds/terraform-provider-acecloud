package volume

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Validation patterns matching ace-cli validate.go and npc-ui constants.
var (
	// Volume name: alphanumeric + hyphens + underscores (no spaces).
	// Matches CLI: AlphaNumericNoSpaceRegex = /^[a-zA-Z0-9\-_]+$/
	volumeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

	// Description: alphanumeric + underscore, hyphen, period, comma, space.
	// Matches CLI: DescriptionRegex = /^[a-zA-Z0-9_\-., ]*$/
	descriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// volumeResourceModel maps the resource schema to a Go struct.
type volumeResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Size        types.Int64  `tfsdk:"size"`
	VolumeType  types.String `tfsdk:"volume_type"`
	BillingType types.String `tfsdk:"billing_type"`
	Description types.String `tfsdk:"description"`
	SourceVolID types.String `tfsdk:"source_volid"`
	SnapshotID  types.String `tfsdk:"snapshot_id"`
	BackupID    types.String `tfsdk:"backup_id"`
	ImageRef    types.String `tfsdk:"image_ref"`
	Bootable    types.Bool   `tfsdk:"bootable"`
	Status      types.String `tfsdk:"status"`
}

// volumeSchema returns the Terraform resource schema for acecloud_volume.
func volumeSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud block storage volume.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the volume.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the volume. Must be 3-100 characters, alphanumeric with hyphens and underscores only.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 100),
					stringvalidator.RegexMatches(volumeNameRegex, "must contain only letters, numbers, hyphens, and underscores"),
				},
			},
			"size": schema.Int64Attribute{
				Description: "Size of the volume in GB. Must be between 8 and 16384 (16 TB).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(8, 16384),
				},
			},
			"volume_type": schema.StringAttribute{
				Description: "Type of the volume (e.g. 'NVMe based High IOPS Storage').",
				Required:    true,
			},
			"billing_type": schema.StringAttribute{
				Description: "Billing type for the volume. Valid: hourly, monthly, quarterly, half-yearly, yearly. Defaults to 'hourly'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("hourly"),
				Validators: []validator.String{
					stringvalidator.OneOf("hourly", "monthly", "quarterly", "half-yearly", "yearly"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the volume (0-255 characters). Only letters, numbers, underscores, hyphens, periods, commas, and spaces allowed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(descriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"source_volid": schema.StringAttribute{
				Description: "UUID of an existing volume to clone.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex(), "must be a valid UUID"),
				},
			},
			"snapshot_id": schema.StringAttribute{
				Description: "UUID of a snapshot to create the volume from.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex(), "must be a valid UUID"),
				},
			},
			"backup_id": schema.StringAttribute{
				Description: "UUID of a backup to restore the volume from.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex(), "must be a valid UUID"),
				},
			},
			"image_ref": schema.StringAttribute{
				Description: "UUID of an image to create a bootable volume from.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegex(), "must be a valid UUID"),
				},
			},
			"bootable": schema.BoolAttribute{
				Description: "Whether the volume is marked as bootable. Volumes created from `image_ref` are bootable automatically; set this to `true` for volumes created from scratch.",
				Optional:    true,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the volume (e.g. available, in-use).",
				Computed:    true,
			},
		},
	}
}
