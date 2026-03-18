package router_interface

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// routerInterfaceModel maps the resource schema to a Go struct.
type routerInterfaceModel struct {
	ID         types.String `tfsdk:"id"`
	RouterID   types.String `tfsdk:"router_id"`
	SubnetID   types.String `tfsdk:"subnet_id"`
	IPAddress  types.String `tfsdk:"ip_address"`
	Status     types.String `tfsdk:"status"`
	MACAddress types.String `tfsdk:"mac_address"`
}

func routerInterfaceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an Ace Cloud router interface (attaches a subnet to a router).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Port ID of the router interface.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"router_id": schema.StringAttribute{
				Description: "ID of the router to attach the subnet to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subnet_id": schema.StringAttribute{
				Description: "ID of the subnet to attach to the router.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_address": schema.StringAttribute{
				Description: "IP address of the router interface on the subnet.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the router interface.",
				Computed:    true,
			},
			"mac_address": schema.StringAttribute{
				Description: "MAC address of the router interface.",
				Computed:    true,
			},
		},
	}
}
