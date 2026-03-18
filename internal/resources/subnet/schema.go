package subnet

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CIDR validation pattern matching ace-cli validate.go CIDRRegex.
var cidrRegex = regexp.MustCompile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/([0-9]|[1-2][0-9]|3[0-2])$`)

// subnetModel maps the resource schema to Go types.
type subnetModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	CIDR            types.String `tfsdk:"cidr"`
	VPCID           types.String `tfsdk:"vpc_id"`
	IPVersion       types.Int64  `tfsdk:"ip_version"`
	Description     types.String `tfsdk:"description"`
	EnableDHCP      types.Bool   `tfsdk:"enable_dhcp"`
	GatewayIP       types.String `tfsdk:"gateway_ip"`
	DNSNameservers  types.List   `tfsdk:"dns_nameservers"`
	AllocationPools types.List   `tfsdk:"allocation_pools"`
	HostRoutes      types.List   `tfsdk:"host_routes"`
}

// allocationPoolModel maps an allocation pool entry.
type allocationPoolModel struct {
	Start types.String `tfsdk:"start"`
	End   types.String `tfsdk:"end"`
}

// hostRouteModel maps a host route entry.
type hostRouteModel struct {
	Destination types.String `tfsdk:"destination"`
	Nexthop     types.String `tfsdk:"nexthop"`
}

func (r *subnetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ace Cloud subnet within a VPC.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the subnet.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the subnet.",
				Required:    true,
			},
			"cidr": schema.StringAttribute{
				Description: "CIDR block for the subnet (e.g. 10.0.1.0/24). Must be a valid IPv4 CIDR.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(cidrRegex, "must be a valid CIDR (e.g. 192.168.1.0/24)"),
				},
			},
			"vpc_id": schema.StringAttribute{
				Description: "ID of the VPC (network) this subnet belongs to.",
				Optional:    true,
			},
			"ip_version": schema.Int64Attribute{
				Description: "IP version for the subnet. Must be 4 or 6.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.OneOf(4, 6),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the subnet.",
				Optional:    true,
			},
			"enable_dhcp": schema.BoolAttribute{
				Description: "Whether DHCP is enabled on the subnet.",
				Optional:    true,
			},
			"gateway_ip": schema.StringAttribute{
				Description: "Gateway IP address for the subnet.",
				Optional:    true,
			},
			"dns_nameservers": schema.ListAttribute{
				Description: "List of DNS nameserver IP addresses.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"allocation_pools": schema.ListNestedBlock{
				Description: "Allocation pools for the subnet.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"start": schema.StringAttribute{
							Description: "Start IP address of the allocation pool.",
							Required:    true,
						},
						"end": schema.StringAttribute{
							Description: "End IP address of the allocation pool.",
							Required:    true,
						},
					},
				},
			},
			"host_routes": schema.ListNestedBlock{
				Description: "Host routes for the subnet.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"destination": schema.StringAttribute{
							Description: "Destination CIDR of the host route.",
							Required:    true,
						},
						"nexthop": schema.StringAttribute{
							Description: "Next hop IP address of the host route.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}
