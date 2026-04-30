package lb_pool_member

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// lbPoolMemberResourceModel maps the resource schema to a Go struct.
type lbPoolMemberResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	PoolID             types.String `tfsdk:"pool_id"`
	Name               types.String `tfsdk:"name"`
	Address            types.String `tfsdk:"address"`
	ProtocolPort       types.Int64  `tfsdk:"protocol_port"`
	Weight             types.Int64  `tfsdk:"weight"`
	MonitorPort        types.Int64  `tfsdk:"monitor_port"`
	MonitorAddress     types.String `tfsdk:"monitor_address"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	AdminStateUp       types.Bool   `tfsdk:"admin_state_up"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

func lbPoolMemberSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Load Balancer Pool Member (backend server).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the pool member.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pool_id": schema.StringAttribute{
				Description: "ID of the pool this member belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the pool member.",
				Optional:    true,
			},
			"address": schema.StringAttribute{
				Description: "IP address of the backend server.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol_port": schema.Int64Attribute{
				Description: "Port number of the backend server (1-65535).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"weight": schema.Int64Attribute{
				Description: "Weight of the member for load balancing. Defaults to 1.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(0, 256),
				},
			},
			"monitor_port": schema.Int64Attribute{
				Description: "Port used for health monitoring (1-65535).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"monitor_address": schema.StringAttribute{
				Description: "IP address used for health monitoring.",
				Optional:    true,
			},
			"provisioning_status": schema.StringAttribute{
				Description: "Provisioning status (e.g. ACTIVE, PENDING_CREATE).",
				Computed:    true,
			},
			"operating_status": schema.StringAttribute{
				Description: "Operating status (e.g. ONLINE, OFFLINE).",
				Computed:    true,
			},
			"admin_state_up": schema.BoolAttribute{
				Description: "Administrative state of the member.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the member was created.",
				Computed:    true,
			},
		},
	}
}
