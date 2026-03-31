package k8s_cluster

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Validation patterns matching ace-cli and npc-ui.
var (
	// Kubernetes cluster name: lowercase alphanumeric + hyphens, max 63 chars.
	// Must start and end with alphanumeric.
	clusterNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

// k8sClusterResourceModel is the Terraform state / plan model for acecloud_k8s_cluster.
type k8sClusterResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	KubernetesVersion   types.String `tfsdk:"kubernetes_version"`
	EndpointAccess      types.String `tfsdk:"endpoint_access"`
	NetworkIsolation    types.String `tfsdk:"network_isolation"`
	NginxIngress        types.String `tfsdk:"nginx_ingress"`
	NginxDefaultBackend types.String `tfsdk:"nginx_default_backend"`
	NetworkProvider     types.String `tfsdk:"network_provider"`
	SnapshotBackup      types.String `tfsdk:"snapshot_backup"`
	SecretsEncryption   types.String `tfsdk:"secrets_encryption"`
	MaxWorkerNodes      types.Int64  `tfsdk:"max_worker_nodes"`
	ClusterType         types.String `tfsdk:"cluster_type"`
	Autoscale           types.Bool   `tfsdk:"autoscale"`
	AutoscaleMin        types.Int64  `tfsdk:"autoscale_min"`
	AutoscaleMax        types.Int64  `tfsdk:"autoscale_max"`
	WorkerNodeName      types.String `tfsdk:"worker_node_name"`
	WorkerQuantity      types.Int64  `tfsdk:"worker_quantity"`
	FlavorID            types.String `tfsdk:"flavor_id"`
	FlavorName          types.String `tfsdk:"flavor_name"`
	VolumeSize          types.Int64  `tfsdk:"volume_size"`
	SecGroupID          types.String `tfsdk:"sec_group_id"`
	Status              types.String `tfsdk:"status"`
	CreatedAt           types.String `tfsdk:"created_at"`
}

// Schema defines the Terraform schema for the acecloud_k8s_cluster resource.
func (r *k8sClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ace Cloud Kubernetes cluster.",
		Attributes: map[string]schema.Attribute{
			// -- Computed identifier --------------------------------------------------
			"id": schema.StringAttribute{
				Description: "Unique identifier of the Kubernetes cluster (assigned by the backend).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// -- Required attributes (all ForceNew) -----------------------------------
			"name": schema.StringAttribute{
				Description: "Cluster name (lowercase alphanumeric and hyphens, max 63 characters).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(clusterNameRegex, "must be lowercase alphanumeric with hyphens, starting and ending with alphanumeric"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kubernetes_version": schema.StringAttribute{
				Description: "Kubernetes version. Use the versions API to list available versions.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint_access": schema.StringAttribute{
				Description: "API server endpoint access. Must be one of: Public, Private, Public and Private.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Public", "Private", "Public and Private"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_isolation": schema.StringAttribute{
				Description: "Network isolation mode. Must be one of: Enabled, Disabled.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Enabled", "Disabled"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nginx_ingress": schema.StringAttribute{
				Description: "NGINX ingress controller. Must be one of: Enabled, Disabled.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Enabled", "Disabled"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nginx_default_backend": schema.StringAttribute{
				Description: "NGINX default backend. Must be one of: Enabled, Disabled.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Enabled", "Disabled"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_provider": schema.StringAttribute{
				Description: "Container network provider. Must be one of: Calico, Flannel.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Calico", "Flannel"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secrets_encryption": schema.StringAttribute{
				Description: "Secrets encryption at rest. Must be one of: Enabled, Disabled.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Enabled", "Disabled"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_worker_nodes": schema.Int64Attribute{
				Description: "Maximum number of worker nodes allowed in the cluster.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"worker_node_name": schema.StringAttribute{
				Description: "Name of the default worker node pool.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"worker_quantity": schema.Int64Attribute{
				Description: "Number of worker nodes to create in the default pool.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"flavor_id": schema.StringAttribute{
				Description: "UUID of the flavor for worker nodes.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"flavor_name": schema.StringAttribute{
				Description: "Name of the flavor for worker nodes (e.g. C4i.medium). Recommended for correct worker node configuration.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"volume_size": schema.Int64Attribute{
				Description: "Volume size in GB for worker nodes.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},

			// -- Optional attributes (all ForceNew) -----------------------------------
			"snapshot_backup": schema.StringAttribute{
				Description: "Snapshot backup. Must be one of: Yes, No.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Yes", "No"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cluster_type": schema.StringAttribute{
				Description: "Cluster provisioning type. Uses the platform default if omitted.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("rke2"), // backend cluster type
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"autoscale": schema.BoolAttribute{
				Description: "Enable cluster autoscaler.",
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"autoscale_min": schema.Int64Attribute{
				Description: "Minimum number of nodes when autoscaling is enabled.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"autoscale_max": schema.Int64Attribute{
				Description: "Maximum number of nodes when autoscaling is enabled.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"sec_group_id": schema.StringAttribute{
				Description: "UUID of the security group to assign to the cluster.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			// -- Computed (read-only) -------------------------------------------------
			"status": schema.StringAttribute{
				Description: "Current status of the Kubernetes cluster.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the cluster was created.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
