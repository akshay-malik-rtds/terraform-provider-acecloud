package floating_ip_association

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// floatingIPAssociationModel maps the resource schema to a Go struct.
type floatingIPAssociationModel struct {
	ID                types.String `tfsdk:"id"`
	FloatingIPAddress types.String `tfsdk:"floating_ip_address"`
	InstanceID        types.String `tfsdk:"instance_id"`
	FixedIPAddress    types.String `tfsdk:"fixed_ip_address"`
}

func floatingIPAssociationSchema() schema.Schema {
	return schema.Schema{
		Description: "Associates a floating IP address with an instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID of the association (floating_ip/instance_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"floating_ip_address": schema.StringAttribute{
				Description: "The floating IP address to associate.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"instance_id": schema.StringAttribute{
				Description: "ID of the instance to associate the floating IP with.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"fixed_ip_address": schema.StringAttribute{
				Description: "Fixed IP address on the instance to associate with (optional).",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
