package snapshot

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	snapshotNameRegex        = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	snapshotDescriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// snapshotResourceModel maps the resource schema to a Go struct.
type snapshotResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	VolumeID    types.String `tfsdk:"volume_id"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	Size        types.Int64  `tfsdk:"size"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func snapshotSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud volume snapshot.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the snapshot.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the snapshot. Must be alphanumeric with hyphens.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(snapshotNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"volume_id": schema.StringAttribute{
				Description: "ID of the volume to create the snapshot from.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the snapshot.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(snapshotDescriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the snapshot (e.g. available, creating).",
				Computed:    true,
			},
			"size": schema.Int64Attribute{
				Description: "Size of the snapshot in GB.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the snapshot was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the snapshot was last updated.",
				Computed:    true,
			},
		},
	}
}
