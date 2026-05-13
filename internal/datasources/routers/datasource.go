package routers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &routersDataSource{}
var _ datasource.DataSourceWithConfigure = &routersDataSource{}

type routersDataSource struct {
	client *client.Client
}

type routersDataSourceModel struct {
	Routers []routerModel `tfsdk:"routers"`
}

type routerModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Status                   types.String `tfsdk:"status"`
	AdminStateUp             types.Bool   `tfsdk:"admin_state_up"`
	ExternalGatewayNetworkID types.String `tfsdk:"external_gateway_network_id"`
}

func NewDataSource() datasource.DataSource {
	return &routersDataSource{}
}

func (d *routersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_routers"
}

func (d *routersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all routers in the current region.",
		Attributes: map[string]schema.Attribute{
			"routers": schema.ListNestedAttribute{
				Description: "List of routers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Router UUID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Router name.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Router status (e.g. ACTIVE, BUILD).",
							Computed:    true,
						},
						"admin_state_up": schema.BoolAttribute{
							Description: "Administrative state of the router.",
							Computed:    true,
						},
						"external_gateway_network_id": schema.StringAttribute{
							Description: "External gateway network UUID.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *routersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *routersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.Get(ctx, "/os/neutron/routers", nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read routers", err.Error())
		return
	}

	// Response may be wrapped in {"routers": [...]} or be a direct array
	var routers []struct {
		ID                  string `json:"id"`
		Name                string `json:"name"`
		Status              string `json:"status"`
		AdminStateUp        bool   `json:"admin_state_up"`
		ExternalGatewayInfo *struct {
			NetworkID string `json:"network_id"`
		} `json:"external_gateway_info"`
	}

	// Try wrapped format
	var wrapped struct {
		Routers json.RawMessage `json:"routers"`
	}
	if err := json.Unmarshal(apiResp.Data, &wrapped); err == nil && wrapped.Routers != nil {
		if err := json.Unmarshal(wrapped.Routers, &routers); err != nil {
			resp.Diagnostics.AddError("Failed to parse routers response", err.Error())
			return
		}
	} else {
		// Try direct array
		if err := json.Unmarshal(apiResp.Data, &routers); err != nil {
			resp.Diagnostics.AddError("Failed to parse routers response", err.Error())
			return
		}
	}

	// Initialise to an empty slice so an empty response renders as [] rather
	// than null in Terraform output.
	state := routersDataSourceModel{Routers: []routerModel{}}
	for _, r := range routers {
		gatewayNetworkID := ""
		if r.ExternalGatewayInfo != nil {
			gatewayNetworkID = r.ExternalGatewayInfo.NetworkID
		}

		state.Routers = append(state.Routers, routerModel{
			ID:                       types.StringValue(r.ID),
			Name:                     types.StringValue(r.Name),
			Status:                   types.StringValue(r.Status),
			AdminStateUp:             types.BoolValue(r.AdminStateUp),
			ExternalGatewayNetworkID: types.StringValue(gatewayNetworkID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
