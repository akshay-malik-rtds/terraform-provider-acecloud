package vpc

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Validation patterns matching ace-cli validate.go and npc-ui constants.
var (
	vpcNameRegex        = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	vpcDescriptionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)
	cidrRegex           = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/([0-9]|[1-2][0-9]|3[0-2])$`)
)

// vpcResourceModel maps the resource schema to a Go struct.
type vpcResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	AdminStateUp types.Bool   `tfsdk:"admin_state_up"`
	Status       types.String `tfsdk:"status"`
	MTU          types.Int64  `tfsdk:"mtu"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`

	// Inline subnet fields (required — backend requires VPC+subnet together)
	SubnetName           types.String `tfsdk:"subnet_name"`
	SubnetCIDR           types.String `tfsdk:"subnet_cidr"`
	SubnetIPVersion      types.Int64  `tfsdk:"subnet_ip_version"`
	SubnetEnableDHCP     types.Bool   `tfsdk:"subnet_enable_dhcp"`
	SubnetDNSNameservers types.List   `tfsdk:"subnet_dns_nameservers"`
	SubnetGatewayIP      types.String `tfsdk:"subnet_gateway_ip"`
	SubnetID             types.String `tfsdk:"subnet_id"`
}

// vpcSchema returns the Terraform resource schema for acecloud_vpc.
func vpcSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud VPC (Virtual Private Cloud) network. The backend requires an initial subnet when creating a VPC, so subnet fields are required.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the VPC.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the VPC. Must be 3-100 characters, alphanumeric with hyphens only.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 100),
					stringvalidator.RegexMatches(vpcNameRegex, "must contain only letters, numbers, and hyphens"),
				},
			},
			"description": schema.StringAttribute{
				Description: "Human-readable description of the VPC (1-100 characters). Only letters, numbers, underscores, hyphens, periods, commas, and spaces allowed.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
					stringvalidator.RegexMatches(vpcDescriptionRegex, "can only contain letters, numbers, underscores, hyphens, periods, commas, and spaces"),
				},
			},
			"admin_state_up": schema.BoolAttribute{
				Description: "Administrative state of the VPC. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"status": schema.StringAttribute{
				Description: "Current status of the VPC (e.g. ACTIVE, BUILD).",
				Computed:    true,
			},
			"mtu": schema.Int64Attribute{
				Description: "Maximum transmission unit of the VPC network.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the VPC was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the VPC was last updated.",
				Computed:    true,
			},

			// --- Inline subnet fields ---
			"subnet_name": schema.StringAttribute{
				Description: "Name of the initial subnet. Required for VPC creation.",
				Required:    true,
			},
			"subnet_cidr": schema.StringAttribute{
				Description: "CIDR block for the initial subnet (e.g. 10.0.0.0/24).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(cidrRegex, "must be a valid CIDR (e.g. 192.168.1.0/24)"),
				},
			},
			"subnet_ip_version": schema.Int64Attribute{
				Description: "IP version for the initial subnet. Must be 4 or 6. Defaults to 4.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(4),
				Validators: []validator.Int64{
					int64validator.OneOf(4, 6),
				},
			},
			"subnet_enable_dhcp": schema.BoolAttribute{
				Description: "Whether DHCP is enabled on the initial subnet. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"subnet_dns_nameservers": schema.ListAttribute{
				Description: "List of DNS nameserver IP addresses for the initial subnet.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"subnet_gateway_ip": schema.StringAttribute{
				Description: "Gateway IP address for the initial subnet. Computed by the backend if not specified.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Description: "Unique identifier of the initial subnet created with the VPC.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
