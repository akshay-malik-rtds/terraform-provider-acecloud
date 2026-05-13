package images

import (
	"context"
	"encoding/json"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &imagesDataSource{}
var _ datasource.DataSourceWithConfigure = &imagesDataSource{}

type imagesDataSource struct {
	client *client.Client
}

type imagesDataSourceModel struct {
	Images []imageModel `tfsdk:"images"`
}

// imageModel maps the API response fields for a single image.
// Matches ace-cli Image struct in internal/api/image.go:
//
//	ID, Name, Status, Size, VirtualSize, DiskFormat, ContainerFormat, Visibility, CreatedAt, UpdatedAt, Tags
//
// Also includes MinDisk and MinRAM which ace-cli parses in the get command.
type imageModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Status          types.String `tfsdk:"status"`
	MinDisk         types.Int64  `tfsdk:"min_disk"`
	MinRAM          types.Int64  `tfsdk:"min_ram"`
	SizeBytes       types.Int64  `tfsdk:"size_bytes"`
	VirtualSize     types.Int64  `tfsdk:"virtual_size"`
	DiskFormat      types.String `tfsdk:"disk_format"`
	ContainerFormat types.String `tfsdk:"container_format"`
	Visibility      types.String `tfsdk:"visibility"`
}

func NewDataSource() datasource.DataSource {
	return &imagesDataSource{}
}

func (d *imagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *imagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List available images in the current region.",
		Attributes: map[string]schema.Attribute{
			"images": schema.ListNestedAttribute{
				Description: "List of available images.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Image UUID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Image name.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Image status (e.g. active, importing, queued, error).",
							Computed:    true,
						},
						"min_disk": schema.Int64Attribute{
							Description: "Minimum disk size in GB required to boot this image.",
							Computed:    true,
						},
						"min_ram": schema.Int64Attribute{
							Description: "Minimum RAM in MB required to boot this image.",
							Computed:    true,
						},
						"size_bytes": schema.Int64Attribute{
							Description: "Image size in bytes.",
							Computed:    true,
						},
						"virtual_size": schema.Int64Attribute{
							Description: "Virtual size of the image in bytes.",
							Computed:    true,
						},
						"disk_format": schema.StringAttribute{
							Description: "Disk format (e.g. qcow2, raw, vdi, vhd, vmdk, iso).",
							Computed:    true,
						},
						"container_format": schema.StringAttribute{
							Description: "Container format (e.g. bare, ami, ari, aki, ovf).",
							Computed:    true,
						},
						"visibility": schema.StringAttribute{
							Description: "Image visibility (e.g. public, private, shared, community).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *imagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *imagesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	apiResp, err := d.client.Get(ctx, "/cloud/images", nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read images", err.Error())
		return
	}

	// The backend response mixes camelCase (diskFormat, containerFormat,
	// virtualSize, createdAt, etc.) and snake_case (size, status) at the
	// top level. Earlier versions used snake_case tags for the camelCase
	// fields, causing them to silently come back as zero values.
	// min_disk and min_ram tags are kept for forward compatibility — the
	// backend's image record does not currently surface them on this route.
	var raw []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Status          string `json:"status"`
		MinDisk         int64  `json:"min_disk"`
		MinRAM          int64  `json:"min_ram"`
		SizeBytes       int64  `json:"size"`
		VirtualSize     int64  `json:"virtualSize"`
		DiskFormat      string `json:"diskFormat"`
		ContainerFormat string `json:"containerFormat"`
		Visibility      string `json:"visibility"`
	}
	if err := json.Unmarshal(apiResp.Data, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to parse images response", err.Error())
		return
	}

	state := imagesDataSourceModel{Images: []imageModel{}}
	for _, img := range raw {
		state.Images = append(state.Images, imageModel{
			ID:              types.StringValue(img.ID),
			Name:            types.StringValue(img.Name),
			Status:          types.StringValue(img.Status),
			MinDisk:         types.Int64Value(img.MinDisk),
			MinRAM:          types.Int64Value(img.MinRAM),
			SizeBytes:       types.Int64Value(img.SizeBytes),
			VirtualSize:     types.Int64Value(img.VirtualSize),
			DiskFormat:      types.StringValue(img.DiskFormat),
			ContainerFormat: types.StringValue(img.ContainerFormat),
			Visibility:      types.StringValue(img.Visibility),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
