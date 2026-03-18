package caas_secret

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// k8s-style name pattern.
var secretNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// caasSecretModel maps the resource schema to Go types.
type caasSecretModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	URL       types.String `tfsdk:"url"`
	Username  types.String `tfsdk:"username"`
	Password  types.String `tfsdk:"password"`
	Data      types.Map    `tfsdk:"data"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func caasSecretSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud CaaS secret. Secrets can be of type 'registry' (for private image authentication) or 'generic' (key-value pairs injected as environment variables).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the secret (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the secret. Must be 3-63 characters, lowercase alphanumeric with hyphens.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
					stringvalidator.RegexMatches(secretNameRegex, "must be lowercase alphanumeric with hyphens, starting and ending with alphanumeric"),
				},
			},
			"type": schema.StringAttribute{
				Description: "Type of secret. Must be 'registry' or 'generic'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("registry", "generic"),
				},
			},
			"url": schema.StringAttribute{
				Description: "Docker registry URL. Required when type is 'registry'.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "Registry username. Required when type is 'registry'.",
				Optional:    true,
				Sensitive:   true,
			},
			"password": schema.StringAttribute{
				Description: "Registry password. Required when type is 'registry'.",
				Optional:    true,
				Sensitive:   true,
			},
			"data": schema.MapAttribute{
				Description: "Key-value pairs for a generic secret. Required when type is 'generic'.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the secret (Provisioning, Active, OutOfSync, Deleting).",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the secret was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the secret was last updated.",
				Computed:    true,
			},
		},
	}
}
