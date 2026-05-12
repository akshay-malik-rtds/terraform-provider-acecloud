package security_group

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Validation patterns matching ace-cli validate.go.
var (
	// Security group name: alphanumeric + spaces, hyphens, underscores.
	// Matches CLI: SecurityGroupName regex = /^[a-zA-Z0-9\s\-_]+$/
	sgNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)

	// Description: alphanumeric + underscore, hyphen, period, comma, space.
	// Matches CLI: DescriptionRegex = /^[a-zA-Z0-9_\-., ]*$/
	sgDescriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
)

// securityGroupModel maps the resource schema to Go types.
type securityGroupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Rules       types.List   `tfsdk:"rules"`
}

// securityGroupRuleModel maps a single rule in the rules list.
type securityGroupRuleModel struct {
	Direction      types.String `tfsdk:"direction"`
	Protocol       types.String `tfsdk:"protocol"`
	PortRangeMin   types.Int64  `tfsdk:"port_range_min"`
	PortRangeMax   types.Int64  `tfsdk:"port_range_max"`
	RemoteIPPrefix types.String `tfsdk:"remote_ip_prefix"`
	RemoteGroupID  types.String `tfsdk:"remote_group_id"`
	EtherType      types.String `tfsdk:"ethertype"`
}

func (r *securityGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ace Cloud security group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the security group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the security group. Must be 3-50 characters, alphanumeric with spaces, hyphens, and underscores.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 50),
					stringvalidator.RegexMatches(sgNameRegex, "can only contain letters, numbers, spaces, hyphens, and underscores"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the security group (0-255 characters). Only letters, numbers, underscores, hyphens, periods, commas, and spaces allowed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
					stringvalidator.RegexMatches(sgDescriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"rules": schema.ListNestedBlock{
				Description: "Security group rules. Maximum 20 rules allowed.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(20),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"direction": schema.StringAttribute{
							Description: "Direction of the rule: 'ingress' or 'egress'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("ingress", "egress"),
							},
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol name (e.g. tcp, udp, icmp, ssh, http, https, rdp, mysql, dns, any).",
							Required:    true,
						},
						"port_range_min": schema.Int64Attribute{
							Description: "Minimum port number (1-65535).",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"port_range_max": schema.Int64Attribute{
							Description: "Maximum port number (1-65535).",
							Optional:    true,
							Validators: []validator.Int64{
								int64validator.Between(1, 65535),
							},
						},
						"remote_ip_prefix": schema.StringAttribute{
							Description: "Remote IP prefix in CIDR notation (e.g. 0.0.0.0/0).",
							Optional:    true,
						},
						"remote_group_id": schema.StringAttribute{
							Description: "Remote security group UUID.",
							Optional:    true,
						},
						"ethertype": schema.StringAttribute{
							Description: "Ethernet type: 'IPv4' or 'IPv6'.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("IPv4", "IPv6"),
							},
						},
					},
				},
			},
		},
	}
}
