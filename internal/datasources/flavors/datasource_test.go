package flavors

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

func TestFlavorModel(t *testing.T) {
	model := flavorModel{}
	_ = model.ID
	_ = model.Name
	_ = model.RAM
	_ = model.VCPUs
	_ = model.Disk
	_ = model.Price
	_ = model.IsGPU
	_ = model.IsHourly
	_ = model.GPUAlias
	_ = model.GPUCount
}

func TestFlavorModel_CPUFlavor(t *testing.T) {
	model := flavorModel{
		ID:       types.StringValue("flav-cpu"),
		Name:     types.StringValue("C4i.medium"),
		RAM:      types.Int64Value(4096),
		VCPUs:    types.Int64Value(4),
		Disk:     types.Int64Value(20),
		Price:    types.StringValue("0.0500"),
		IsGPU:    types.BoolValue(false),
		IsHourly: types.BoolValue(true),
		GPUAlias: types.StringValue(""),
		GPUCount: types.Int64Value(0),
	}

	if model.ID.ValueString() != "flav-cpu" {
		t.Errorf("expected flav-cpu, got %s", model.ID.ValueString())
	}
	if model.RAM.ValueInt64() != 4096 {
		t.Errorf("expected RAM 4096, got %d", model.RAM.ValueInt64())
	}
	if model.VCPUs.ValueInt64() != 4 {
		t.Errorf("expected VCPUs 4, got %d", model.VCPUs.ValueInt64())
	}
	if model.IsGPU.ValueBool() {
		t.Error("expected IsGPU false for CPU flavor")
	}
	if model.GPUCount.ValueInt64() != 0 {
		t.Errorf("expected GPUCount 0, got %d", model.GPUCount.ValueInt64())
	}
}

func TestFlavorModel_GPUFlavor(t *testing.T) {
	model := flavorModel{
		ID:       types.StringValue("flav-gpu"),
		Name:     types.StringValue("gpu.a100.1"),
		RAM:      types.Int64Value(65536),
		VCPUs:    types.Int64Value(16),
		Disk:     types.Int64Value(100),
		Price:    types.StringValue("2.5000"),
		IsGPU:    types.BoolValue(true),
		IsHourly: types.BoolValue(true),
		GPUAlias: types.StringValue("A100"),
		GPUCount: types.Int64Value(1),
	}

	if !model.IsGPU.ValueBool() {
		t.Error("expected IsGPU true for GPU flavor")
	}
	if model.GPUAlias.ValueString() != "A100" {
		t.Errorf("expected A100, got %s", model.GPUAlias.ValueString())
	}
	if model.GPUCount.ValueInt64() != 1 {
		t.Errorf("expected GPUCount 1, got %d", model.GPUCount.ValueInt64())
	}
}

func TestFlavorsDataSourceModel(t *testing.T) {
	model := flavorsDataSourceModel{
		Flavors: []flavorModel{
			{ID: types.StringValue("f1"), Name: types.StringValue("small")},
			{ID: types.StringValue("f2"), Name: types.StringValue("medium")},
		},
	}

	if len(model.Flavors) != 2 {
		t.Fatalf("expected 2 flavors, got %d", len(model.Flavors))
	}
}
