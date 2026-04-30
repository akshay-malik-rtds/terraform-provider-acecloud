package lb_pool

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
	poolNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

// lbPoolResourceModel maps the resource schema to a Go struct.
type lbPoolResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Protocol           types.String `tfsdk:"protocol"`
	LBAlgorithm        types.String `tfsdk:"lb_algorithm"`
	ListenerID         types.String `tfsdk:"listener_id"`
	LoadBalancerID     types.String `tfsdk:"loadbalancer_id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	HealthMonitorID    types.String `tfsdk:"healthmonitor_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

func lbPoolSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Load Balancer Pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the pool.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the pool. Changing this forces recreation.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(poolNameRegex, "must contain only letters, numbers, and hyphens"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				Description: "Protocol for the pool (HTTP, HTTPS, TCP, UDP, PROXY, PROXYV2).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("HTTP", "HTTPS", "TCP", "UDP", "PROXY", "PROXYV2"),
				},
			},
			"lb_algorithm": schema.StringAttribute{
				Description: "Load balancing algorithm (ROUND_ROBIN, LEAST_CONNECTIONS, SOURCE_IP).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("ROUND_ROBIN", "LEAST_CONNECTIONS", "SOURCE_IP"),
				},
			},
			"listener_id": schema.StringAttribute{
				Description: "ID of the listener this pool belongs to.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"loadbalancer_id": schema.StringAttribute{
				Description: "ID of the load balancer this pool belongs to.",
				Optional:    true,
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
			"healthmonitor_id": schema.StringAttribute{
				Description: "ID of the health monitor associated with this pool.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the pool was created.",
				Computed:    true,
			},
		},
	}
}
