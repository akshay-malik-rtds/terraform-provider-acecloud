package registry_replication_rule

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Model structs ---

type replicationRuleModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	SrcRegistry       types.Object `tfsdk:"src_registry"`
	DestRegistry      types.Object `tfsdk:"dest_registry"`
	DestNamespace     types.String `tfsdk:"dest_namespace"`
	Trigger           types.Object `tfsdk:"trigger"`
	Filter            types.List   `tfsdk:"filter"`
	ReplicateDeletion types.Bool   `tfsdk:"replicate_deletion"`
	Override          types.Bool   `tfsdk:"override"`
	Speed             types.Int64  `tfsdk:"speed"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
}

type registryModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
	Type types.String `tfsdk:"type"`
}

type triggerModel struct {
	Type types.String `tfsdk:"type"`
	Cron types.String `tfsdk:"cron"`
}

type filterModel struct {
	Type       types.String `tfsdk:"type"`
	Value      types.String `tfsdk:"value"`
	Decoration types.String `tfsdk:"decoration"`
}

// --- attr.Type maps for nested objects ---

func registryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.Int64Type,
		"name": types.StringType,
		"url":  types.StringType,
		"type": types.StringType,
	}
}

func triggerAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"cron": types.StringType,
	}
}

func filterAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":       types.StringType,
		"value":      types.StringType,
		"decoration": types.StringType,
	}
}

func registryBlockSchema(description string, required bool) schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: description,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "Registry ID.",
				Required:    required,
				Optional:    !required,
			},
			"name": schema.StringAttribute{
				Description: "Registry name.",
				Required:    required,
				Optional:    !required,
			},
			"url": schema.StringAttribute{
				Description: "Registry endpoint URL.",
				Required:    required,
				Optional:    !required,
			},
			"type": schema.StringAttribute{
				Description: "Registry type (e.g. docker-hub, aws-ecr). Defaults to the AceCloud registry type.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func replicationRuleSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages a registry replication rule for Ace Cloud container registry.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the replication rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the replication rule.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the replication rule.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the replication rule is enabled.",
				Required:    true,
			},
			"dest_namespace": schema.StringAttribute{
				Description: "Destination namespace for replicated artifacts.",
				Optional:    true,
			},
			"replicate_deletion": schema.BoolAttribute{
				Description: "Whether to replicate artifact deletion.",
				Optional:    true,
				Computed:    true,
			},
			"override": schema.BoolAttribute{
				Description: "Whether to override resources at the destination.",
				Optional:    true,
				Computed:    true,
			},
			"speed": schema.Int64Attribute{
				Description: "Maximum network bandwidth in Kbps for replication. -1 for unlimited.",
				Optional:    true,
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the replication rule was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the replication rule was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"src_registry":  registryBlockSchema("Source registry configuration.", true),
			"dest_registry": registryBlockSchema("Destination registry configuration.", false),
			"trigger": schema.SingleNestedBlock{
				Description: "Trigger configuration for the replication rule.",
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Trigger type (e.g. manual, scheduled, event_based).",
						Required:    true,
					},
					"cron": schema.StringAttribute{
						Description: "Cron expression for scheduled triggers (e.g. '0 0 * * *').",
						Optional:    true,
					},
				},
			},
			"filter": schema.ListNestedBlock{
				Description: "Filters to apply to the replication rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "Filter type (e.g. name, tag, label, resource).",
							Required:    true,
						},
						"value": schema.StringAttribute{
							Description: "Filter value.",
							Required:    true,
						},
						"decoration": schema.StringAttribute{
							Description: "Filter decoration (e.g. matches, excludes).",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}
