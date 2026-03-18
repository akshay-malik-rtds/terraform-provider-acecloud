package router

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
	routerNameRegex        = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	routerDescriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// routerResourceModel maps the resource schema to a Go struct.
type routerResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	AdminStateUp              types.Bool   `tfsdk:"admin_state_up"`
	ExternalGatewayNetworkID  types.String `tfsdk:"external_gateway_network_id"`
	Status                    types.String `tfsdk:"status"`
	CreatedAt                 types.String `tfsdk:"created_at"`
	UpdatedAt                 types.String `tfsdk:"updated_at"`
}

func routerSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Router.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the router.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the router. Must be 1-100 characters, alphanumeric with hyphens.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(routerNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the router.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(routerDescriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"admin_state_up": schema.BoolAttribute{
				Description: "Administrative state of the router. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"external_gateway_network_id": schema.StringAttribute{
				Description: "UUID of the external network for the router's gateway.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the router (e.g. ACTIVE, BUILD).",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the router was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the router was last updated.",
				Computed:    true,
			},
		},
	}
}
