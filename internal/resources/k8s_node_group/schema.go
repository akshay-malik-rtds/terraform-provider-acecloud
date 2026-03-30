package k8s_node_group

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	// K8s node group name: lowercase alphanumeric + hyphens, must start/end with alphanumeric.
	// Matches ace-cli K8s naming conventions.
	nodeGroupNameRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

// k8sNodeGroupModel maps the resource schema to a Go struct.
type k8sNodeGroupModel struct {
	ID          types.String `tfsdk:"id"`
	ClusterID   types.String `tfsdk:"cluster_id"`
	SecGroupID  types.String `tfsdk:"sec_group_id"`
	Name        types.String `tfsdk:"name"`
	Quantity    types.Int64  `tfsdk:"quantity"`
	FlavorID    types.String `tfsdk:"flavor_id"`
	Volume      types.String `tfsdk:"volume"`
	Labels      types.Map    `tfsdk:"labels"`
	Annotations types.Map    `tfsdk:"annotations"`
	MinNode     types.Int64  `tfsdk:"min_node"`
	MaxNode     types.Int64  `tfsdk:"max_node"`
	State       types.String `tfsdk:"state"`
}

func k8sNodeGroupSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud Kubernetes node group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the node group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "UUID of the Kubernetes cluster this node group belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sec_group_id": schema.StringAttribute{
				Description: "UUID of the security group for the node group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the node group (lowercase alphanumeric and hyphens, max 30 characters).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 30),
					stringvalidator.RegexMatches(nodeGroupNameRegex, "must be lowercase alphanumeric with hyphens, starting and ending with alphanumeric"),
				},
			},
			"quantity": schema.Int64Attribute{
				Description: "Number of nodes in the node group. This is the only mutable field (triggers scale operation on update).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"flavor_id": schema.StringAttribute{
				Description: "UUID of the compute flavor for nodes in this group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"volume": schema.StringAttribute{
				Description: "Volume size for each node (e.g. \"50\").",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labels": schema.MapAttribute{
				Description: "Key-value labels to apply to the node group.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"annotations": schema.MapAttribute{
				Description: "Key-value annotations to apply to the node group.",
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"min_node": schema.Int64Attribute{
				Description: "Minimum number of nodes for autoscaling (1-8).",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 8),
				},
			},
			"max_node": schema.Int64Attribute{
				Description: "Maximum number of nodes for autoscaling (1-8).",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 8),
				},
			},
			"state": schema.StringAttribute{
				Description: "Current state of the node group.",
				Computed:    true,
			},
		},
	}
}
