package security_groups

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

func TestSecurityGroupModel_AllFields(t *testing.T) {
	model := securityGroupModel{}
	_ = model.ID
	_ = model.Name
	_ = model.Description
}

func TestSecurityGroupModel_Values(t *testing.T) {
	model := securityGroupModel{
		ID:          types.StringValue("sg-123"),
		Name:        types.StringValue("web-sg"),
		Description: types.StringValue("Allow HTTP/HTTPS"),
	}

	if model.ID.ValueString() != "sg-123" {
		t.Errorf("expected sg-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "web-sg" {
		t.Errorf("expected web-sg, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Allow HTTP/HTTPS" {
		t.Errorf("expected description 'Allow HTTP/HTTPS', got %s", model.Description.ValueString())
	}
}

func TestSecurityGroupsDataSourceModel(t *testing.T) {
	model := securityGroupsDataSourceModel{
		SecurityGroups: []securityGroupModel{
			{ID: types.StringValue("sg-1"), Name: types.StringValue("default")},
			{ID: types.StringValue("sg-2"), Name: types.StringValue("web")},
		},
	}

	if len(model.SecurityGroups) != 2 {
		t.Fatalf("expected 2 security groups, got %d", len(model.SecurityGroups))
	}
	if model.SecurityGroups[0].Name.ValueString() != "default" {
		t.Errorf("expected 'default', got %s", model.SecurityGroups[0].Name.ValueString())
	}
}

func TestSecurityGroupModel_EmptyDescription(t *testing.T) {
	model := securityGroupModel{
		ID:          types.StringValue("sg-empty"),
		Name:        types.StringValue("no-desc"),
		Description: types.StringValue(""),
	}
	if model.Description.ValueString() != "" {
		t.Errorf("expected empty description, got %s", model.Description.ValueString())
	}
}
