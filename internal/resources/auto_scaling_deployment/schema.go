package auto_scaling_deployment

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	deploymentNameRegex = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)
	descriptionRegex    = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// autoScalingDeploymentModel maps the resource schema to Go types.
type autoScalingDeploymentModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	TemplateID       types.String `tfsdk:"template_id"`
	DesiredCapacity  types.Int64  `tfsdk:"desired_capacity"`
	MaxCapacity      types.Int64  `tfsdk:"max_capacity"`
	NodesScaleCount  types.Int64  `tfsdk:"nodes_scale_count"`
	ScalingParameter types.String `tfsdk:"scaling_parameter"`
	MinThreshold     types.Int64  `tfsdk:"min_threshold"`
	MaxThreshold     types.Int64  `tfsdk:"max_threshold"`
	CoolDownTime     types.Int64  `tfsdk:"cool_down_time"`
	UserEmail        types.List   `tfsdk:"user_email"`
	IsIntegratedLB   types.Bool   `tfsdk:"is_integrated_with_lb"`

	// LB nested block
	LBData types.Object `tfsdk:"lb_data"`

	// Computed
	Status       types.String `tfsdk:"status"`
	ErrorMessage types.String `tfsdk:"error_message"`
	PanelURL     types.String `tfsdk:"panel_url"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

// lbDataModel maps the lb_data nested block.
type lbDataModel struct {
	LBName          types.String `tfsdk:"lb_name"`
	Tags            types.List   `tfsdk:"tags"`
	AssignPublicIP  types.Bool   `tfsdk:"assign_public_ip"`
	IsExistingLB    types.Bool   `tfsdk:"is_existing_lb"`
	LBID            types.String `tfsdk:"lb_id"`
	LBVipPortID     types.String `tfsdk:"lb_vip_port_id"`
	PublicNetworkID types.String `tfsdk:"public_network_id"`

	// Sub-blocks
	Listener      types.Object `tfsdk:"listener"`
	Pool          types.Object `tfsdk:"pool"`
	HealthMonitor types.Object `tfsdk:"health_monitor"`
}

type listenerModel struct {
	ListenerName         types.String `tfsdk:"listener_name"`
	ListenerProtocol     types.String `tfsdk:"listener_protocol"`
	ListenerProtocolPort types.Int64  `tfsdk:"listener_protocol_port"`
}

type poolModel struct {
	PoolProtocol     types.String `tfsdk:"pool_protocol"`
	PoolProtocolPort types.Int64  `tfsdk:"pool_protocol_port"`
	LBAlgorithm      types.String `tfsdk:"lb_algorithm"`
}

type healthMonitorModel struct {
	MonitorProtocol   types.String `tfsdk:"monitor_protocol"`
	MonitorURLPath    types.String `tfsdk:"monitor_url_path"`
	MonitorHTTPMethod types.String `tfsdk:"monitor_http_method"`
}

func autoScalingDeploymentSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Auto Scaling deployment. Deployments define scaling policies and manage instance groups based on CPU/RAM thresholds.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the auto scaling deployment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the deployment (1-255 characters).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
					stringvalidator.RegexMatches(deploymentNameRegex, "must contain only letters, numbers, spaces, underscores, and hyphens"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the deployment (3-255 characters).",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 255),
					stringvalidator.RegexMatches(descriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"template_id": schema.StringAttribute{
				Description: "ID of the auto scaling template to use.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"desired_capacity": schema.Int64Attribute{
				Description: "Desired number of instances (1-30). Changing this forces recreation.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 30),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"max_capacity": schema.Int64Attribute{
				Description: "Maximum number of instances (1-30). Changing this forces recreation.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 30),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"nodes_scale_count": schema.Int64Attribute{
				Description: "Number of nodes to add/remove per scale event (1-30). Changing this forces recreation.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 30),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"scaling_parameter": schema.StringAttribute{
				Description: "Metric to monitor for scaling. Must be 'cpu' or 'ram'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("cpu", "ram"),
				},
			},
			"min_threshold": schema.Int64Attribute{
				Description: "Minimum threshold percentage for scale-in (30-90). Changing this forces recreation.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(30, 90),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"max_threshold": schema.Int64Attribute{
				Description: "Maximum threshold percentage for scale-out (40-95). Changing this forces recreation.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(40, 95),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"cool_down_time": schema.Int64Attribute{
				Description: "Cool-down period in seconds between scale events (60-3600). Changing this forces recreation.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(60, 3600),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"user_email": schema.ListAttribute{
				Description: "List of email addresses to notify on scaling events.",
				Required:    true,
				ElementType: types.StringType,
			},
			"is_integrated_with_lb": schema.BoolAttribute{
				Description: "Whether to integrate with a load balancer.",
				Required:    true,
			},
			// Computed
			"status": schema.StringAttribute{
				Description: "Current status of the deployment (CREATING, ACTIVE, ERROR, etc.).",
				Computed:    true,
			},
			"error_message": schema.StringAttribute{
				Description: "Error message if the deployment is in ERROR status.",
				Computed:    true,
			},
			"panel_url": schema.StringAttribute{
				Description: "URL of the monitoring dashboard for this deployment.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the deployment was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the deployment was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"lb_data": schema.SingleNestedBlock{
				Description: "Load balancer configuration. Required when is_integrated_with_lb is true.",
				Attributes: map[string]schema.Attribute{
					"lb_name": schema.StringAttribute{
						Description: "Name of the new load balancer. Required when is_existing_lb is false.",
						Optional:    true,
					},
					"tags": schema.ListAttribute{
						Description: "Tags for the load balancer.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"assign_public_ip": schema.BoolAttribute{
						Description: "Whether to assign a public IP to the load balancer.",
						Optional:    true,
					},
					"is_existing_lb": schema.BoolAttribute{
						Description: "Whether to use an existing load balancer.",
						Optional:    true,
					},
					"lb_id": schema.StringAttribute{
						Description: "ID of an existing load balancer. Required when is_existing_lb is true.",
						Optional:    true,
					},
					"lb_vip_port_id": schema.StringAttribute{
						Description: "VIP port ID of an existing load balancer.",
						Optional:    true,
					},
					"public_network_id": schema.StringAttribute{
						Description: "Public network ID for floating IP. Required when assign_public_ip is true.",
						Optional:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"listener": schema.SingleNestedBlock{
						Description: "Listener configuration for the load balancer.",
						Attributes: map[string]schema.Attribute{
							"listener_name": schema.StringAttribute{
								Description: "Name of the listener.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.LengthBetween(1, 255),
								},
							},
							"listener_protocol": schema.StringAttribute{
								Description: "Protocol for the listener.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("TCP", "HTTP", "HTTPS", "TERMINATED_HTTPS", "PROXY", "UDP"),
								},
							},
							"listener_protocol_port": schema.Int64Attribute{
								Description: "Port for the listener (1-65535).",
								Optional:    true,
								Validators: []validator.Int64{
									int64validator.Between(1, 65535),
								},
							},
						},
					},
					"pool": schema.SingleNestedBlock{
						Description: "Pool configuration for the load balancer.",
						Attributes: map[string]schema.Attribute{
							"pool_protocol": schema.StringAttribute{
								Description: "Protocol for the pool.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("TCP", "HTTP", "HTTPS", "TERMINATED_HTTPS", "PROXY", "UDP"),
								},
							},
							"pool_protocol_port": schema.Int64Attribute{
								Description: "Port for the pool (1-65535).",
								Optional:    true,
								Validators: []validator.Int64{
									int64validator.Between(1, 65535),
								},
							},
							"lb_algorithm": schema.StringAttribute{
								Description: "Load balancing algorithm.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("ROUND_ROBIN", "LEAST_CONNECTIONS", "SOURCE_IP", "SOURCE_IP_PORT"),
								},
							},
						},
					},
					"health_monitor": schema.SingleNestedBlock{
						Description: "Health monitor configuration for the load balancer.",
						Attributes: map[string]schema.Attribute{
							"monitor_protocol": schema.StringAttribute{
								Description: "Protocol for health monitoring.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("PING", "TCP", "HTTP", "HTTPS", "UDP-CONNECT"),
								},
							},
							"monitor_url_path": schema.StringAttribute{
								Description: "URL path for HTTP/HTTPS health checks.",
								Optional:    true,
							},
							"monitor_http_method": schema.StringAttribute{
								Description: "HTTP method for health checks.",
								Optional:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"),
								},
							},
						},
					},
				},
			},
		},
	}
}
