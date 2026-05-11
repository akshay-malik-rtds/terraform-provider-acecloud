package lb_listener

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Protocol             types.String `tfsdk:"protocol"`
	ProtocolPort         types.Int64  `tfsdk:"protocol_port"`
	LoadBalancerID       types.String `tfsdk:"loadbalancer_id"`
	Description          types.String `tfsdk:"description"`
	ConnectionLimit      types.Int64  `tfsdk:"connection_limit"`
	AllowedCIDRs         types.List   `tfsdk:"allowed_cidrs"`
	InsertHeaders        types.Map    `tfsdk:"insert_headers"`
	TimeoutClientData    types.Int64  `tfsdk:"timeout_client_data"`
	TimeoutMemberConnect types.Int64  `tfsdk:"timeout_member_connect"`
	TimeoutMemberData    types.Int64  `tfsdk:"timeout_member_data"`
	TLSCiphers           types.String `tfsdk:"tls_ciphers"`
	ProvisioningStatus   types.String `tfsdk:"provisioning_status"`
	OperatingStatus      types.String `tfsdk:"operating_status"`
	DefaultPoolID        types.String `tfsdk:"default_pool_id"`
	CreatedAt            types.String `tfsdk:"created_at"`
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
				Description: "Name of the listener. Changing this forces recreation.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(listenerNameRegex, "must contain only letters, numbers, and hyphens"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
				Description: "Port number for the listener (1-65535). Changing this forces recreation.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"loadbalancer_id": schema.StringAttribute{
				Description: "ID of the load balancer this listener belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the listener.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"connection_limit": schema.Int64Attribute{
				Description: "Maximum number of concurrent connections. Use `-1` for unlimited (default).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(-1),
			},
			"allowed_cidrs": schema.ListAttribute{
				Description: "List of CIDR blocks allowed to connect to the listener. Empty list means no restriction.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"insert_headers": schema.MapAttribute{
				Description: "HTTP headers to insert. Only applicable for HTTP/HTTPS/TERMINATED_HTTPS protocols. Common keys: `X-Forwarded-Proto`, `X-Forwarded-For`, `X-Forwarded-Port`. Values are stringified booleans (`\"true\"`/`\"false\"`).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"timeout_client_data": schema.Int64Attribute{
				Description: "Frontend client inactivity timeout in milliseconds. Only applicable for TCP/UDP/PROXY protocols.",
				Optional:    true,
				Computed:    true,
			},
			"timeout_member_connect": schema.Int64Attribute{
				Description: "Backend member connection timeout in milliseconds. Only applicable for TCP/UDP/PROXY protocols.",
				Optional:    true,
				Computed:    true,
			},
			"timeout_member_data": schema.Int64Attribute{
				Description: "Backend member inactivity timeout in milliseconds. Only applicable for TCP/UDP/PROXY protocols.",
				Optional:    true,
				Computed:    true,
			},
			"tls_ciphers": schema.StringAttribute{
				Description: "Colon-separated list of TLS ciphers (e.g. `ECDHE-RSA-AES256-GCM-SHA384`). Only applicable for TERMINATED_HTTPS protocol.",
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
