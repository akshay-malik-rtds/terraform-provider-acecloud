package caas_deployment

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var deploymentNameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// --- Model structs ---

type caasDeploymentModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	Command          types.List   `tfsdk:"command"`
	EnvSecrets       types.List   `tfsdk:"env_secrets"`
	Image            types.Object `tfsdk:"image"`
	Resources        types.Object `tfsdk:"resources"`
	Networking       types.Object `tfsdk:"networking"`
	Autoscaling      types.Object `tfsdk:"autoscaling"`
	Env              types.List   `tfsdk:"env"`
	Volume           types.List   `tfsdk:"volume"`
	Status           types.String `tfsdk:"status"`
	PrivateEndpoints types.List   `tfsdk:"private_endpoints"`
	PublicEndpoints  types.List   `tfsdk:"public_endpoints"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

type imageModel struct {
	Type      types.String `tfsdk:"type"`
	Reference types.String `tfsdk:"reference"`
	Secrets   types.List   `tfsdk:"secrets"`
}

type resourcesModel struct {
	CPU          types.Float64 `tfsdk:"cpu"`
	Memory       types.String  `tfsdk:"memory"`
	ReplicaCount types.Int64   `tfsdk:"replica_count"`
	FlavorID     types.String  `tfsdk:"flavor_id"`
}

type networkingModel struct {
	ExternalAccess      types.Bool   `tfsdk:"external_access"`
	EndpointAccess      types.String `tfsdk:"endpoint_access"`
	CIDRBlock           types.List   `tfsdk:"cidr_block"`
	XForwardedFor       types.Bool   `tfsdk:"x_forwarded_for"`
	UseExistingNetwork  types.Bool   `tfsdk:"use_existing_network"`
	NetworkID           types.String `tfsdk:"network_id"`
	CreateNewNetworkCIDR types.String `tfsdk:"create_new_network_cidr"`
	Port                types.List   `tfsdk:"port"`
}

type portModel struct {
	Name          types.String `tfsdk:"name"`
	Protocol      types.String `tfsdk:"protocol"`
	ContainerPort types.Int64  `tfsdk:"container_port"`
	ExposedPort   types.Int64  `tfsdk:"exposed_port"`
}

type autoscalingModel struct {
	Enabled                  types.Bool    `tfsdk:"enabled"`
	MinReplicas              types.Int64   `tfsdk:"min_replicas"`
	MaxReplicas              types.Int64   `tfsdk:"max_replicas"`
	CPUTargetPercentage      types.Float64 `tfsdk:"cpu_target_percentage"`
	MemoryTargetPercentage   types.Float64 `tfsdk:"memory_target_percentage"`
}

type envModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type volumeModel struct {
	Name      types.String `tfsdk:"name"`
	MountPath types.String `tfsdk:"mount_path"`
	Size      types.String `tfsdk:"size"`
}

func caasDeploymentSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud CaaS (Container as a Service) deployment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the deployment (same as name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the deployment. Must be 3-60 characters, lowercase alphanumeric with hyphens.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 60),
					stringvalidator.RegexMatches(deploymentNameRegex, "must be lowercase alphanumeric with hyphens"),
				},
			},
			"type": schema.StringAttribute{
				Description: "Deployment type. Must be 'shared' or 'dedicated'. Defaults to 'shared'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("shared"),
				Validators: []validator.String{
					stringvalidator.OneOf("shared", "dedicated"),
				},
			},
			"command": schema.ListAttribute{
				Description: "Container command override.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"env_secrets": schema.ListAttribute{
				Description: "List of generic secret names to inject as environment variables.",
				Optional:    true,
				ElementType: types.StringType,
			},
			// Computed
			"status": schema.StringAttribute{
				Description: "Current status of the deployment (Active, Provisioning, Error, etc.).",
				Computed:    true,
			},
			"private_endpoints": schema.ListAttribute{
				Description: "Private endpoint URLs for the deployment.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"public_endpoints": schema.ListAttribute{
				Description: "Public endpoint URLs for the deployment.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the deployment was created.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"image": schema.SingleNestedBlock{
				Description: "Container image configuration.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Image type. Must be 'public' or 'private'.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("public", "private"),
						},
					},
					"reference": schema.StringAttribute{
						Description: "Image reference (e.g. nginx:latest, myrepo/myimage:v1).",
						Required:    true,
					},
					"secrets": schema.ListAttribute{
						Description: "Image pull secret names. Required for private images.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"resources": schema.SingleNestedBlock{
				Description: "Resource allocation for the deployment.",
				Attributes: map[string]schema.Attribute{
					"cpu": schema.Float64Attribute{
						Description: "CPU cores for shared deployments (e.g. 0.5, 1.0).",
						Optional:    true,
					},
					"memory": schema.StringAttribute{
						Description: "Memory allocation in K8s format (e.g. 512Mi, 1Gi).",
						Optional:    true,
					},
					"replica_count": schema.Int64Attribute{
						Description: "Number of container replicas (1-12).",
						Required:    true,
						Validators: []validator.Int64{
							int64validator.Between(1, 12),
						},
					},
					"flavor_id": schema.StringAttribute{
						Description: "Flavor ID for dedicated deployments.",
						Optional:    true,
					},
				},
			},
			"networking": schema.SingleNestedBlock{
				Description: "Networking configuration for the deployment.",
				Attributes: map[string]schema.Attribute{
					"external_access": schema.BoolAttribute{
						Description: "Whether to enable external access to the deployment.",
						Required:    true,
					},
					"endpoint_access": schema.StringAttribute{
						Description: "Endpoint access level. Must be 'public' or 'protected'.",
						Optional:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("public", "protected"),
						},
					},
					"cidr_block": schema.ListAttribute{
						Description: "CIDR blocks for protected endpoint access.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"x_forwarded_for": schema.BoolAttribute{
						Description: "Whether to enable X-Forwarded-For header support.",
						Optional:    true,
					},
					"use_existing_network": schema.BoolAttribute{
						Description: "Whether to use an existing VPC network.",
						Optional:    true,
					},
					"network_id": schema.StringAttribute{
						Description: "ID of an existing VPC network to use.",
						Optional:    true,
					},
					"create_new_network_cidr": schema.StringAttribute{
						Description: "CIDR block for creating a new network.",
						Optional:    true,
					},
				},
				Blocks: map[string]schema.Block{
					"port": schema.ListNestedBlock{
						Description: "Port mappings for the deployment.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Name of the port.",
									Required:    true,
								},
								"protocol": schema.StringAttribute{
									Description: "Protocol for the port. Must be HTTP, HTTPS, TCP, or UDP.",
									Required:    true,
									Validators: []validator.String{
										stringvalidator.OneOf("HTTP", "HTTPS", "TCP", "UDP"),
									},
								},
								"container_port": schema.Int64Attribute{
									Description: "Container port number (1-65535).",
									Required:    true,
									Validators: []validator.Int64{
										int64validator.Between(1, 65535),
									},
								},
								"exposed_port": schema.Int64Attribute{
									Description: "Exposed port number (1-65535). Required when external_access is true.",
									Optional:    true,
									Validators: []validator.Int64{
										int64validator.Between(1, 65535),
									},
								},
							},
						},
					},
				},
			},
			"autoscaling": schema.SingleNestedBlock{
				Description: "Horizontal pod autoscaling configuration.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether autoscaling is enabled.",
						Required:    true,
					},
					"min_replicas": schema.Int64Attribute{
						Description: "Minimum number of replicas (1-12).",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.Between(1, 12),
						},
					},
					"max_replicas": schema.Int64Attribute{
						Description: "Maximum number of replicas (1-12).",
						Optional:    true,
						Validators: []validator.Int64{
							int64validator.Between(1, 12),
						},
					},
					"cpu_target_percentage": schema.Float64Attribute{
						Description: "Target CPU utilization percentage (0-100).",
						Optional:    true,
						Validators: []validator.Float64{
							float64validator.Between(0, 100),
						},
					},
					"memory_target_percentage": schema.Float64Attribute{
						Description: "Target memory utilization percentage (0-100).",
						Optional:    true,
						Validators: []validator.Float64{
							float64validator.Between(0, 100),
						},
					},
				},
			},
			"env": schema.ListNestedBlock{
				Description: "Environment variables for the deployment.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Environment variable name.",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Environment variable value.",
							Required:    true,
						},
					},
				},
			},
			"volume": schema.ListNestedBlock{
				Description: "Persistent volume mounts for the deployment.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Volume name.",
							Required:    true,
						},
						"mount_path": schema.StringAttribute{
							Description: "Mount path inside the container.",
							Required:    true,
						},
						"size": schema.StringAttribute{
							Description: "Volume size in K8s format (e.g. 256Mi, 1Gi).",
							Required:    true,
						},
					},
				},
			},
		},
	}
}
