package routers

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

func TestRouterModel_AllFields(t *testing.T) {
	model := routerModel{}
	_ = model.ID
	_ = model.Name
	_ = model.Status
	_ = model.AdminStateUp
	_ = model.ExternalGatewayNetworkID
}

func TestRouterModel_Values(t *testing.T) {
	model := routerModel{
		ID:                       types.StringValue("rtr-123"),
		Name:                     types.StringValue("my-router"),
		Status:                   types.StringValue("ACTIVE"),
		AdminStateUp:             types.BoolValue(true),
		ExternalGatewayNetworkID: types.StringValue("net-ext-1"),
	}

	if model.ID.ValueString() != "rtr-123" {
		t.Errorf("expected rtr-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "my-router" {
		t.Errorf("expected my-router, got %s", model.Name.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected ACTIVE, got %s", model.Status.ValueString())
	}
	if !model.AdminStateUp.ValueBool() {
		t.Error("expected AdminStateUp true")
	}
	if model.ExternalGatewayNetworkID.ValueString() != "net-ext-1" {
		t.Errorf("expected net-ext-1, got %s", model.ExternalGatewayNetworkID.ValueString())
	}
}

func TestRouterModel_EmptyGateway(t *testing.T) {
	model := routerModel{
		ID:                       types.StringValue("rtr-456"),
		Name:                     types.StringValue("no-gw"),
		Status:                   types.StringValue("ACTIVE"),
		AdminStateUp:             types.BoolValue(true),
		ExternalGatewayNetworkID: types.StringValue(""),
	}

	if model.ExternalGatewayNetworkID.ValueString() != "" {
		t.Errorf("expected empty gateway, got %s", model.ExternalGatewayNetworkID.ValueString())
	}
}

func TestRoutersDataSourceModel(t *testing.T) {
	model := routersDataSourceModel{
		Routers: []routerModel{
			{ID: types.StringValue("rtr-1"), Name: types.StringValue("first")},
			{ID: types.StringValue("rtr-2"), Name: types.StringValue("second")},
		},
	}

	if len(model.Routers) != 2 {
		t.Fatalf("expected 2 routers, got %d", len(model.Routers))
	}
	if model.Routers[0].ID.ValueString() != "rtr-1" {
		t.Errorf("expected rtr-1, got %s", model.Routers[0].ID.ValueString())
	}
}
