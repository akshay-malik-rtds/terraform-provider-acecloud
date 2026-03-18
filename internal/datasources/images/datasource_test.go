package images

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewDataSource(t *testing.T) {
	ds := NewDataSource()
	if ds == nil {
		t.Fatal("expected non-nil data source")
	}
}

func TestImageModel(t *testing.T) {
	model := imageModel{}
	_ = model.ID
	_ = model.Name
	_ = model.Status
	_ = model.MinDisk
	_ = model.MinRAM
	_ = model.SizeBytes
	_ = model.VirtualSize
	_ = model.DiskFormat
	_ = model.ContainerFormat
	_ = model.Visibility
}

func TestImageModel_Values(t *testing.T) {
	model := imageModel{
		ID:              types.StringValue("img-123"),
		Name:            types.StringValue("Ubuntu 22.04"),
		Status:          types.StringValue("active"),
		MinDisk:         types.Int64Value(20),
		MinRAM:          types.Int64Value(1024),
		SizeBytes:       types.Int64Value(2147483648),
		VirtualSize:     types.Int64Value(4294967296),
		DiskFormat:      types.StringValue("qcow2"),
		ContainerFormat: types.StringValue("bare"),
		Visibility:      types.StringValue("public"),
	}

	if model.ID.ValueString() != "img-123" {
		t.Errorf("expected img-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "Ubuntu 22.04" {
		t.Errorf("expected 'Ubuntu 22.04', got %s", model.Name.ValueString())
	}
	if model.Status.ValueString() != "active" {
		t.Errorf("expected active, got %s", model.Status.ValueString())
	}
	if model.MinDisk.ValueInt64() != 20 {
		t.Errorf("expected MinDisk 20, got %d", model.MinDisk.ValueInt64())
	}
	if model.MinRAM.ValueInt64() != 1024 {
		t.Errorf("expected MinRAM 1024, got %d", model.MinRAM.ValueInt64())
	}
	if model.DiskFormat.ValueString() != "qcow2" {
		t.Errorf("expected qcow2, got %s", model.DiskFormat.ValueString())
	}
	if model.Visibility.ValueString() != "public" {
		t.Errorf("expected public, got %s", model.Visibility.ValueString())
	}
}

func TestImagesDataSourceModel(t *testing.T) {
	model := imagesDataSourceModel{
		Images: []imageModel{
			{ID: types.StringValue("img-1"), Name: types.StringValue("ubuntu")},
			{ID: types.StringValue("img-2"), Name: types.StringValue("centos")},
			{ID: types.StringValue("img-3"), Name: types.StringValue("debian")},
		},
	}

	if len(model.Images) != 3 {
		t.Fatalf("expected 3 images, got %d", len(model.Images))
	}
}

func TestImageModel_PrivateImage(t *testing.T) {
	model := imageModel{
		ID:         types.StringValue("img-priv"),
		Visibility: types.StringValue("private"),
		Status:     types.StringValue("active"),
	}
	if model.Visibility.ValueString() != "private" {
		t.Errorf("expected private, got %s", model.Visibility.ValueString())
	}
}
