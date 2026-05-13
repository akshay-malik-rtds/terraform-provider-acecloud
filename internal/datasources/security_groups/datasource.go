package security_groups

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &securityGroupsDataSource{}
var _ datasource.DataSourceWithConfigure = &securityGroupsDataSource{}

type securityGroupsDataSource struct {
	client *client.Client
}

type securityGroupsDataSourceModel struct {
	SecurityGroups []securityGroupModel `tfsdk:"security_groups"`
}

type securityGroupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewDataSource() datasource.DataSource {
	return &securityGroupsDataSource{}
}

func (d *securityGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_groups"
}

func (d *securityGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List all security groups in the current region.",
		Attributes: map[string]schema.Attribute{
			"security_groups": schema.ListNestedAttribute{
				Description: "List of security groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Security group UUID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Security group name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Security group description.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *securityGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityGroupsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.Get(ctx, "/cloud/security-groups", nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read security groups", err.Error())
		return
	}

	var raw []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(apiResp.Data, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to parse security groups response", err.Error())
		return
	}

	state := securityGroupsDataSourceModel{SecurityGroups: []securityGroupModel{}}
	for _, sg := range raw {
		state.SecurityGroups = append(state.SecurityGroups, securityGroupModel{
			ID:          types.StringValue(sg.ID),
			Name:        types.StringValue(sg.Name),
			Description: types.StringValue(sg.Description),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
