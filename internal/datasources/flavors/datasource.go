package flavors

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &flavorsDataSource{}
var _ datasource.DataSourceWithConfigure = &flavorsDataSource{}

type flavorsDataSource struct {
	client *client.Client
}

type flavorsDataSourceModel struct {
	Flavors []flavorModel `tfsdk:"flavors"`
}

// flavorModel maps the API response fields for a single flavor.
// Matches ace-cli Flavor struct in internal/api/flavor.go:
//
//	ID, Name, VCPUs, Memory, Disk, Price, ExtraSpecs{GPU, IsHourly, Alias, RequestedQuantity}
type flavorModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	RAM      types.Int64  `tfsdk:"ram"`
	VCPUs    types.Int64  `tfsdk:"vcpus"`
	Disk     types.Int64  `tfsdk:"disk"`
	Price    types.String `tfsdk:"price"`
	IsGPU    types.Bool   `tfsdk:"is_gpu"`
	IsHourly types.Bool   `tfsdk:"is_hourly"`
	GPUAlias types.String `tfsdk:"gpu_alias"`
	GPUCount types.Int64  `tfsdk:"gpu_count"`
}

func NewDataSource() datasource.DataSource {
	return &flavorsDataSource{}
}

func (d *flavorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flavors"
}

func (d *flavorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List available compute flavors in the current region.",
		Attributes: map[string]schema.Attribute{
			"flavors": schema.ListNestedAttribute{
				Description: "List of available flavors.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Flavor UUID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Flavor name (e.g. 'ace.small', 'gpu.a100.1').",
							Computed:    true,
						},
						"ram": schema.Int64Attribute{
							Description: "RAM in MB.",
							Computed:    true,
						},
						"vcpus": schema.Int64Attribute{
							Description: "Number of virtual CPUs.",
							Computed:    true,
						},
						"disk": schema.Int64Attribute{
							Description: "Root disk size in GB.",
							Computed:    true,
						},
						"price": schema.StringAttribute{
							Description: "Price per hour as a string (e.g. '0.50').",
							Computed:    true,
						},
						"is_gpu": schema.BoolAttribute{
							Description: "Whether this is a GPU flavor.",
							Computed:    true,
						},
						"is_hourly": schema.BoolAttribute{
							Description: "Whether this flavor supports hourly billing.",
							Computed:    true,
						},
						"gpu_alias": schema.StringAttribute{
							Description: "GPU model alias (e.g. 'A100', 'H100'). Empty for CPU flavors.",
							Computed:    true,
						},
						"gpu_count": schema.Int64Attribute{
							Description: "Number of GPUs. 0 for CPU flavors.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *flavorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *flavorsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.Get(ctx, "/cloud/flavors", nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read flavors", err.Error())
		return
	}

	// The API response matches the ace-cli Flavor struct with nested extra_specs.
	var raw []struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		RAM        int64   `json:"ram"`
		VCPUs      int64   `json:"vcpus"`
		Disk       int64   `json:"disk"`
		Price      float64 `json:"price"`
		ExtraSpecs struct {
			GPU               bool   `json:"gpu"`
			IsHourly          bool   `json:"is_hourly"`
			Alias             string `json:"alias"`
			RequestedQuantity int64  `json:"requested_quantity"`
		} `json:"extra_specs"`
		// Also handle flat field in case backend returns "memory" at top level.
		Memory int64 `json:"memory"`
	}
	if err := json.Unmarshal(apiResp.Data, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to parse flavors response", err.Error())
		return
	}

	var state flavorsDataSourceModel
	for _, f := range raw {
		ram := f.RAM
		if ram == 0 && f.Memory > 0 {
			ram = f.Memory
		}

		priceStr := ""
		if f.Price > 0 {
			priceStr = fmt.Sprintf("%.4f", f.Price)
		}

		state.Flavors = append(state.Flavors, flavorModel{
			ID:       types.StringValue(f.ID),
			Name:     types.StringValue(f.Name),
			RAM:      types.Int64Value(ram),
			VCPUs:    types.Int64Value(f.VCPUs),
			Disk:     types.Int64Value(f.Disk),
			Price:    types.StringValue(priceStr),
			IsGPU:    types.BoolValue(f.ExtraSpecs.GPU),
			IsHourly: types.BoolValue(f.ExtraSpecs.IsHourly),
			GPUAlias: types.StringValue(f.ExtraSpecs.Alias),
			GPUCount: types.Int64Value(f.ExtraSpecs.RequestedQuantity),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
