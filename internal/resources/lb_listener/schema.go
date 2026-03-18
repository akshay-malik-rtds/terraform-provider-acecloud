package lb_listener

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	listenerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

// lbListenerResourceModel maps the resource schema to a Go struct.
type lbListenerResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	ProtocolPort       types.Int64  `tfsdk:"protocol_port"`
	LoadBalancerID     types.String `tfsdk:"loadbalancer_id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	DefaultPoolID      types.String `tfsdk:"default_pool_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

func lbListenerSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Load Balancer Listener.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the listener.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the listener.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(listenerNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"protocol": schema.StringAttribute{
				Description: "Protocol for the listener (HTTP, HTTPS, TCP, UDP, TERMINATED_HTTPS).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("HTTP", "HTTPS", "TCP", "UDP", "TERMINATED_HTTPS"),
				},
			},
			"protocol_port": schema.Int64Attribute{
				Description: "Port number for the listener (1-65535).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"loadbalancer_id": schema.StringAttribute{
				Description: "ID of the load balancer this listener belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provisioning_status": schema.StringAttribute{
				Description: "Provisioning status (e.g. ACTIVE, PENDING_CREATE).",
				Computed:    true,
			},
			"operating_status": schema.StringAttribute{
				Description: "Operating status (e.g. ONLINE, OFFLINE).",
				Computed:    true,
			},
			"default_pool_id": schema.StringAttribute{
				Description: "ID of the default pool for this listener.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the listener was created.",
				Computed:    true,
			},
		},
	}
}
