package load_balancer

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
	lbNameRegex        = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	lbDescriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// loadBalancerResourceModel maps the resource schema to a Go struct.
type loadBalancerResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	SubnetID            types.String `tfsdk:"subnet_id"`
	Description         types.String `tfsdk:"description"`
	Tags                types.List   `tfsdk:"tags"`
	VIPAddress          types.String `tfsdk:"vip_address"`
	VIPPortID           types.String `tfsdk:"vip_port_id"`
	VIPNetworkID        types.String `tfsdk:"vip_network_id"`
	ProvisioningStatus  types.String `tfsdk:"provisioning_status"`
	OperatingStatus     types.String `tfsdk:"operating_status"`
	Provider            types.String `tfsdk:"provider_name"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
}

func loadBalancerSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Load Balancer.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the load balancer.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the load balancer. Must be alphanumeric with hyphens.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(lbNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"subnet_id": schema.StringAttribute{
				Description: "ID of the subnet for the load balancer VIP.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the load balancer.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(lbDescriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"tags": schema.ListAttribute{
				Description: "Tags for the load balancer. Must include exactly one of: ALB, NLB.",
				Required:    true,
				ElementType: types.StringType,
			},
			"vip_address": schema.StringAttribute{
				Description: "Virtual IP address of the load balancer.",
				Computed:    true,
			},
			"vip_port_id": schema.StringAttribute{
				Description: "Port ID of the VIP.",
				Computed:    true,
			},
			"vip_network_id": schema.StringAttribute{
				Description: "Network ID of the VIP.",
				Computed:    true,
			},
			"provisioning_status": schema.StringAttribute{
				Description: "Provisioning status (e.g. ACTIVE, PENDING_CREATE).",
				Computed:    true,
			},
			"operating_status": schema.StringAttribute{
				Description: "Operating status (e.g. ONLINE, OFFLINE).",
				Computed:    true,
			},
			"provider_name": schema.StringAttribute{
				Description: "Load balancer backend provider name.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the load balancer was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the load balancer was last updated.",
				Computed:    true,
			},
		},
	}
}
