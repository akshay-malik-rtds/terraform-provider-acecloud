package lb_pool

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	Description        types.String `tfsdk:"description"`
	TLSEnabled         types.Bool   `tfsdk:"tls_enabled"`
	TLSCiphers         types.String `tfsdk:"tls_ciphers"`
	SessionPersistence types.Object `tfsdk:"session_persistence"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	HealthMonitorID    types.String `tfsdk:"healthmonitor_id"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

// sessionPersistenceModel maps the session_persistence nested object.
type sessionPersistenceModel struct {
	Type       types.String `tfsdk:"type"`
	CookieName types.String `tfsdk:"cookie_name"`
}

// sessionPersistenceAttrTypes returns the attribute types for the
// session_persistence nested object.
func sessionPersistenceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":        types.StringType,
		"cookie_name": types.StringType,
	}
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
			"description": schema.StringAttribute{
				Description: "Human-readable description of the pool.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"tls_enabled": schema.BoolAttribute{
				Description: "Whether TLS is enabled for backend member connections. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"tls_ciphers": schema.StringAttribute{
				Description: "Colon-separated list of TLS ciphers for backend connections. Only applicable when `tls_enabled` is true.",
				Optional:    true,
			},
			"session_persistence": schema.SingleNestedAttribute{
				Description: "Session persistence configuration. Omit to disable.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Persistence type: `SOURCE_IP`, `HTTP_COOKIE`, or `APP_COOKIE`.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("SOURCE_IP", "HTTP_COOKIE", "APP_COOKIE"),
						},
					},
					"cookie_name": schema.StringAttribute{
						Description: "Cookie name. Required when `type` is `APP_COOKIE`.",
						Optional:    true,
					},
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
