package vpcs

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

func TestVPCModel_AllFields(t *testing.T) {
	model := vpcModel{}
	_ = model.ID
	_ = model.Name
	_ = model.Status
	_ = model.AdminStateUp
	_ = model.RouterID
}

func TestVPCModel_Values(t *testing.T) {
	model := vpcModel{
		ID:           types.StringValue("vpc-123"),
		Name:         types.StringValue("my-vpc"),
		Status:       types.StringValue("ACTIVE"),
		AdminStateUp: types.BoolValue(true),
		RouterID:     types.StringValue("rtr-1"),
	}

	if model.ID.ValueString() != "vpc-123" {
		t.Errorf("expected vpc-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "my-vpc" {
		t.Errorf("expected my-vpc, got %s", model.Name.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s", model.Status.ValueString())
	}
	if !model.AdminStateUp.ValueBool() {
		t.Error("expected AdminStateUp true")
	}
	if model.RouterID.ValueString() != "rtr-1" {
		t.Errorf("expected rtr-1, got %s", model.RouterID.ValueString())
	}
}

func TestVPCsDataSourceModel(t *testing.T) {
	model := vpcsDataSourceModel{
		VPCs: []vpcModel{
			{ID: types.StringValue("vpc-1"), Name: types.StringValue("first")},
			{ID: types.StringValue("vpc-2"), Name: types.StringValue("second")},
			{ID: types.StringValue("vpc-3"), Name: types.StringValue("third")},
		},
	}

	if len(model.VPCs) != 3 {
		t.Fatalf("expected 3 VPCs, got %d", len(model.VPCs))
	}
}

func TestVPCModel_EmptyRouterID(t *testing.T) {
	model := vpcModel{
		ID:       types.StringValue("vpc-no-rtr"),
		RouterID: types.StringValue(""),
	}
	if model.RouterID.ValueString() != "" {
		t.Errorf("expected empty router_id, got %s", model.RouterID.ValueString())
	}
}
