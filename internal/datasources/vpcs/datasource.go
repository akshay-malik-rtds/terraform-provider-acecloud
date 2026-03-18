package vpcs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &vpcsDataSource{}
var _ datasource.DataSourceWithConfigure = &vpcsDataSource{}

type vpcsDataSource struct {
	client *client.Client
}

type vpcsDataSourceModel struct {
	VPCs []vpcModel `tfsdk:"vpcs"`
}

type vpcModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Status       types.String `tfsdk:"status"`
	AdminStateUp types.Bool   `tfsdk:"admin_state_up"`
	RouterID     types.String `tfsdk:"router_id"`
}

func NewDataSource() datasource.DataSource {
	return &vpcsDataSource{}
}

func (d *vpcsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

func (d *vpcsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all VPCs in the current region.",
		Attributes: map[string]schema.Attribute{
			"vpcs": schema.ListNestedAttribute{
				Description: "List of VPCs.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "VPC UUID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "VPC name.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "VPC status (e.g. ACTIVE, BUILD).",
							Computed:    true,
						},
						"admin_state_up": schema.BoolAttribute{
							Description: "Administrative state of the VPC.",
							Computed:    true,
						},
						"router_id": schema.StringAttribute{
							Description: "Router ID associated with the VPC.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *vpcsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vpcsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.Get(ctx, "/cloud/vpcs", nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read VPCs", err.Error())
		return
	}

	var raw []struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Status       string `json:"status"`
		AdminStateUp bool   `json:"admin_state_up"`
		RouterID     string `json:"router:external"`
	}
	if err := json.Unmarshal(apiResp.Data, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to parse VPCs response", err.Error())
		return
	}

	var state vpcsDataSourceModel
	for _, v := range raw {
		state.VPCs = append(state.VPCs, vpcModel{
			ID:           types.StringValue(v.ID),
			Name:         types.StringValue(v.Name),
			Status:       types.StringValue(v.Status),
			AdminStateUp: types.BoolValue(v.AdminStateUp),
			RouterID:     types.StringValue(v.RouterID),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
