package floating_ip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// floatingIPModel maps the resource schema to Go types.
type floatingIPModel struct {
	ID                types.String `tfsdk:"id"`
	FloatingNetworkID types.String `tfsdk:"floating_network_id"`
	PortID            types.String `tfsdk:"port_id"`
	Description       types.String `tfsdk:"description"`
	FloatingIPAddress types.String `tfsdk:"floating_ip_address"`
	Status            types.String `tfsdk:"status"`
}

func (r *floatingIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Ace Cloud floating IP address.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the floating IP.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"floating_network_id": schema.StringAttribute{
				Description: "UUID of the external network to allocate the floating IP from.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port_id": schema.StringAttribute{
				Description: "ID of the port to associate the floating IP with.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the floating IP.",
				Optional:    true,
			},
			"floating_ip_address": schema.StringAttribute{
				Description: "The allocated floating IP address.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the floating IP (e.g. ACTIVE, DOWN).",
				Computed:    true,
			},
		},
	}
}
