package lb_health_monitor

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
	hmNameRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
)

// lbHealthMonitorResourceModel maps the resource schema to a Go struct.
type lbHealthMonitorResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	PoolID             types.String `tfsdk:"pool_id"`
	Type               types.String `tfsdk:"type"`
	Delay              types.Int64  `tfsdk:"delay"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	MaxRetries         types.Int64  `tfsdk:"max_retries"`
	MaxRetriesDown     types.Int64  `tfsdk:"max_retries_down"`
	URLPath            types.String `tfsdk:"url_path"`
	ExpectedCodes      types.String `tfsdk:"expected_codes"`
	HTTPMethod         types.String `tfsdk:"http_method"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
	OperatingStatus    types.String `tfsdk:"operating_status"`
	AdminStateUp       types.Bool   `tfsdk:"admin_state_up"`
	CreatedAt          types.String `tfsdk:"created_at"`
}

func lbHealthMonitorSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Load Balancer Health Monitor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the health monitor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the health monitor.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(hmNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"pool_id": schema.StringAttribute{
				Description: "ID of the pool to monitor.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Description: "Health monitor type (HTTP, HTTPS, TCP, PING, TLS-HELLO, UDP-CONNECT, SCTP).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("HTTP", "HTTPS", "TCP", "PING", "TLS-HELLO", "UDP-CONNECT", "SCTP"),
				},
			},
			"delay": schema.Int64Attribute{
				Description: "Delay between health checks in seconds.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout for each health check in seconds.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_retries": schema.Int64Attribute{
				Description: "Number of successful checks before marking as healthy.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"max_retries_down": schema.Int64Attribute{
				Description: "Number of failed checks before marking as unhealthy.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 10),
				},
			},
			"url_path": schema.StringAttribute{
				Description: "URL path for HTTP/HTTPS health checks (e.g. /health). Defaults to / for HTTP/HTTPS monitors.",
				Optional:    true,
				Computed:    true,
			},
			"expected_codes": schema.StringAttribute{
				Description: "Expected HTTP response codes for HTTP/HTTPS checks (e.g. 200, 200-299). Defaults to 200.",
				Optional:    true,
				Computed:    true,
			},
			"http_method": schema.StringAttribute{
				Description: "HTTP method for HTTP/HTTPS checks (GET, POST, etc.). Defaults to GET.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "CONNECT", "TRACE"),
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
			"admin_state_up": schema.BoolAttribute{
				Description: "Administrative state of the health monitor.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the health monitor was created.",
				Computed:    true,
			},
		},
	}
}
