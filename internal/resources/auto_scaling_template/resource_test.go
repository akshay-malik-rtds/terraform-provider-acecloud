package auto_scaling_template

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- buildCreateRequest tests ---

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("test-template"),
		Type:                types.StringValue("linux"),
		Description:         types.StringValue("Test description"),
		VolumeSize:          types.Int64Value(50),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("flavor-123"),
		ImageID:             types.StringValue("image-456"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringValue("my-key"),
		NetworkID:           types.StringValue("net-789"),
		SubnetID:            types.StringValue("subnet-abc"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
			types.StringValue("sg-2"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "test-template" {
		t.Errorf("expected name test-template, got %s", body.Name)
	}
	if body.Type != "linux" {
		t.Errorf("expected type linux, got %s", body.Type)
	}
	if body.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %s", body.Description)
	}
	if body.VolumeSize != 50 {
		t.Errorf("expected volume_size 50, got %d", body.VolumeSize)
	}
	if body.VolDelOnTermination != true {
		t.Error("expected vol_del_on_termination to be true")
	}
	if body.FlavorID != "flavor-123" {
		t.Errorf("expected flavor_id flavor-123, got %s", body.FlavorID)
	}
	if body.ImageID != "image-456" {
		t.Errorf("expected image_id image-456, got %s", body.ImageID)
	}
	if body.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id when null, got %s", body.SnapshotID)
	}
	if body.KeyName != "my-key" {
		t.Errorf("expected key_name my-key, got %s", body.KeyName)
	}
	if body.NetworkID != "net-789" {
		t.Errorf("expected network_id net-789, got %s", body.NetworkID)
	}
	if body.SubnetID != "subnet-abc" {
		t.Errorf("expected subnet_id subnet-abc, got %s", body.SubnetID)
	}
	if len(body.SecurityGroups) != 2 || body.SecurityGroups[0] != "sg-1" {
		t.Errorf("expected 2 security groups, got %v", body.SecurityGroups)
	}
	if body.IsInstanceSnapshot != false {
		t.Error("expected is_instance_snapshot to be false")
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("min-template"),
		Type:                types.StringValue("windows"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(20),
		VolDelOnTermination: types.BoolValue(false),
		FlavorID:            types.StringValue("flavor-1"),
		ImageID:             types.StringNull(),
		SnapshotID:          types.StringValue("snap-123"),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-default"),
		}),
		IsInstanceSnapshot: types.BoolValue(true),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Description != "" {
		t.Errorf("expected empty description when null, got %s", body.Description)
	}
	if body.ImageID != "" {
		t.Errorf("expected empty image_id when null, got %s", body.ImageID)
	}
	if body.SnapshotID != "snap-123" {
		t.Errorf("expected snapshot_id snap-123, got %s", body.SnapshotID)
	}
	if body.KeyName != "" {
		t.Errorf("expected empty key_name when null, got %s", body.KeyName)
	}
	if body.IsInstanceSnapshot != true {
		t.Error("expected is_instance_snapshot to be true")
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("json-template"),
		Type:                types.StringValue("linux"),
		Description:         types.StringValue("JSON test"),
		VolumeSize:          types.Int64Value(100),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringValue("img-1"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Required fields must be present
	for _, key := range []string{"name", "type", "volume_size", "vol_del_on_termination", "flavor_id", "network_id", "subnet_id", "security_groups", "is_instance_snapshot"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' to be present", key)
		}
	}

	// Optional omitempty fields
	if _, ok := raw["snapshot_id"]; ok {
		t.Error("expected 'snapshot_id' to be omitted (omitempty)")
	}
	if _, ok := raw["key_name"]; ok {
		t.Error("expected 'key_name' to be omitted (omitempty)")
	}
}

func TestBuildCreateRequest_AllSecurityGroups(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("multi-sg-template"),
		Type:                types.StringValue("linux"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(40),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-multi"),
		ImageID:             types.StringValue("img-multi"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-multi"),
		SubnetID:            types.StringValue("sub-multi"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-alpha"),
			types.StringValue("sg-beta"),
			types.StringValue("sg-gamma"),
			types.StringValue("sg-delta"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)

	if len(body.SecurityGroups) != 4 {
		t.Fatalf("expected 4 security groups, got %d", len(body.SecurityGroups))
	}
	expected := []string{"sg-alpha", "sg-beta", "sg-gamma", "sg-delta"}
	for i, sg := range body.SecurityGroups {
		if sg != expected[i] {
			t.Errorf("expected security_groups[%d] %q, got %q", i, expected[i], sg)
		}
	}
}

func TestBuildCreateRequest_SingleSecurityGroup(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("single-sg"),
		Type:                types.StringValue("linux"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(20),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringValue("img-1"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-only"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)
	if len(body.SecurityGroups) != 1 {
		t.Fatalf("expected 1 security group, got %d", len(body.SecurityGroups))
	}
	if body.SecurityGroups[0] != "sg-only" {
		t.Errorf("expected sg-only, got %s", body.SecurityGroups[0])
	}
}

func TestBuildCreateRequest_WindowsType(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("win-template"),
		Type:                types.StringValue("windows"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(100),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-win"),
		ImageID:             types.StringValue("img-win"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)
	if body.Type != "windows" {
		t.Errorf("expected type windows, got %s", body.Type)
	}
}

func TestBuildCreateRequest_SnapshotBoot(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("snap-template"),
		Type:                types.StringValue("linux"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(50),
		VolDelOnTermination: types.BoolValue(false),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringNull(),
		SnapshotID:          types.StringValue("snap-boot-123"),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(true),
	}

	body := buildCreateRequest(context.Background(), plan)
	if body.ImageID != "" {
		t.Errorf("expected empty image_id for snapshot boot, got %s", body.ImageID)
	}
	if body.SnapshotID != "snap-boot-123" {
		t.Errorf("expected snapshot_id snap-boot-123, got %s", body.SnapshotID)
	}
	if body.IsInstanceSnapshot != true {
		t.Error("expected is_instance_snapshot true for snapshot boot")
	}
}

func TestBuildCreateRequest_ImageBoot(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("img-template"),
		Type:                types.StringValue("linux"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(50),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringValue("img-boot-456"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringValue("my-key"),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)
	if body.ImageID != "img-boot-456" {
		t.Errorf("expected image_id img-boot-456, got %s", body.ImageID)
	}
	if body.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id for image boot, got %s", body.SnapshotID)
	}
	if body.IsInstanceSnapshot != false {
		t.Error("expected is_instance_snapshot false for image boot")
	}
}

func TestBuildCreateRequest_UnknownFieldsTreatedAsEmpty(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("unknown-test"),
		Type:                types.StringValue("linux"),
		Description:         types.StringUnknown(),
		VolumeSize:          types.Int64Value(20),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringUnknown(),
		SnapshotID:          types.StringUnknown(),
		KeyName:             types.StringUnknown(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Description != "" {
		t.Errorf("expected empty description when unknown, got %s", body.Description)
	}
	if body.ImageID != "" {
		t.Errorf("expected empty image_id when unknown, got %s", body.ImageID)
	}
	if body.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id when unknown, got %s", body.SnapshotID)
	}
	if body.KeyName != "" {
		t.Errorf("expected empty key_name when unknown, got %s", body.KeyName)
	}
}

func TestBuildCreateRequest_JSON_OmitEmptyOptionals(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("omit-test"),
		Type:                types.StringValue("linux"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(20),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringNull(),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	omitKeys := []string{"description", "image_id", "snapshot_id", "key_name"}
	for _, key := range omitKeys {
		if _, ok := raw[key]; ok {
			t.Errorf("expected '%s' to be omitted when null", key)
		}
	}
}

func TestBuildCreateRequest_JSON_AllOptionalPresent(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("all-opt"),
		Type:                types.StringValue("linux"),
		Description:         types.StringValue("has desc"),
		VolumeSize:          types.Int64Value(20),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringValue("img-1"),
		SnapshotID:          types.StringValue("snap-1"),
		KeyName:             types.StringValue("key-1"),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildCreateRequest(context.Background(), plan)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	presentKeys := []string{"description", "image_id", "snapshot_id", "key_name"}
	for _, key := range presentKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected '%s' to be present when set", key)
		}
	}
}

// --- buildUpdateRequest tests ---

func TestBuildUpdateRequest(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("updated-template"),
		Type:                types.StringValue("linux"),
		Description:         types.StringValue("Updated desc"),
		VolumeSize:          types.Int64Value(80),
		VolDelOnTermination: types.BoolValue(false),
		FlavorID:            types.StringValue("f-new"),
		ImageID:             types.StringValue("img-new"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringValue("new-key"),
		NetworkID:           types.StringValue("net-2"),
		SubnetID:            types.StringValue("sub-2"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-new"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.Name != "updated-template" {
		t.Errorf("expected name updated-template, got %s", body.Name)
	}
	if body.Description != "Updated desc" {
		t.Errorf("expected description 'Updated desc', got %s", body.Description)
	}
	if body.VolumeSize != 80 {
		t.Errorf("expected volume_size 80, got %d", body.VolumeSize)
	}
}

func TestBuildUpdateRequest_AllFields(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("full-update-template"),
		Type:                types.StringValue("windows"),
		Description:         types.StringValue("Updated description"),
		VolumeSize:          types.Int64Value(200),
		VolDelOnTermination: types.BoolValue(false),
		FlavorID:            types.StringValue("f-updated"),
		ImageID:             types.StringValue("img-updated"),
		SnapshotID:          types.StringValue("snap-updated"),
		KeyName:             types.StringValue("updated-key"),
		NetworkID:           types.StringValue("net-updated"),
		SubnetID:            types.StringValue("sub-updated"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-updated-1"),
			types.StringValue("sg-updated-2"),
		}),
		IsInstanceSnapshot: types.BoolValue(true),
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.Name != "full-update-template" {
		t.Errorf("expected name full-update-template, got %s", body.Name)
	}
	if body.Type != "windows" {
		t.Errorf("expected type windows, got %s", body.Type)
	}
	if body.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %s", body.Description)
	}
	if body.VolumeSize != 200 {
		t.Errorf("expected volume_size 200, got %d", body.VolumeSize)
	}
	if body.VolDelOnTermination != false {
		t.Error("expected vol_del_on_termination to be false")
	}
	if body.FlavorID != "f-updated" {
		t.Errorf("expected flavor_id f-updated, got %s", body.FlavorID)
	}
	if body.ImageID != "img-updated" {
		t.Errorf("expected image_id img-updated, got %s", body.ImageID)
	}
	if body.SnapshotID != "snap-updated" {
		t.Errorf("expected snapshot_id snap-updated, got %s", body.SnapshotID)
	}
	if body.KeyName != "updated-key" {
		t.Errorf("expected key_name updated-key, got %s", body.KeyName)
	}
	if body.NetworkID != "net-updated" {
		t.Errorf("expected network_id net-updated, got %s", body.NetworkID)
	}
	if body.SubnetID != "sub-updated" {
		t.Errorf("expected subnet_id sub-updated, got %s", body.SubnetID)
	}
	if len(body.SecurityGroups) != 2 {
		t.Fatalf("expected 2 security groups, got %d", len(body.SecurityGroups))
	}
	if body.SecurityGroups[0] != "sg-updated-1" || body.SecurityGroups[1] != "sg-updated-2" {
		t.Errorf("expected security_groups [sg-updated-1, sg-updated-2], got %v", body.SecurityGroups)
	}
	if body.IsInstanceSnapshot != true {
		t.Error("expected is_instance_snapshot to be true")
	}
}

func TestBuildUpdateRequest_NullOptionals(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("update-null"),
		Type:                types.StringValue("linux"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(30),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringNull(),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.Description != "" {
		t.Errorf("expected empty description, got %s", body.Description)
	}
	if body.ImageID != "" {
		t.Errorf("expected empty image_id, got %s", body.ImageID)
	}
	if body.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id, got %s", body.SnapshotID)
	}
	if body.KeyName != "" {
		t.Errorf("expected empty key_name, got %s", body.KeyName)
	}
}

func TestBuildUpdateRequest_JSON(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("json-update"),
		Type:                types.StringValue("linux"),
		Description:         types.StringNull(),
		VolumeSize:          types.Int64Value(30),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringNull(),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringNull(),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	body := buildUpdateRequest(context.Background(), plan)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Required fields always present
	for _, key := range []string{"name", "type", "volume_size", "vol_del_on_termination", "flavor_id", "network_id", "subnet_id", "security_groups", "is_instance_snapshot"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' in update body", key)
		}
	}
}

func TestBuildUpdateRequest_MatchesCreateStructure(t *testing.T) {
	plan := &autoScalingTemplateModel{
		Name:                types.StringValue("match-test"),
		Type:                types.StringValue("linux"),
		Description:         types.StringValue("desc"),
		VolumeSize:          types.Int64Value(50),
		VolDelOnTermination: types.BoolValue(true),
		FlavorID:            types.StringValue("f-1"),
		ImageID:             types.StringValue("img-1"),
		SnapshotID:          types.StringNull(),
		KeyName:             types.StringValue("key-1"),
		NetworkID:           types.StringValue("net-1"),
		SubnetID:            types.StringValue("sub-1"),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-1"),
		}),
		IsInstanceSnapshot: types.BoolValue(false),
	}

	createBody := buildCreateRequest(context.Background(), plan)
	updateBody := buildUpdateRequest(context.Background(), plan)

	// Fields should match between create and update
	if createBody.Name != updateBody.Name {
		t.Error("name mismatch between create and update")
	}
	if createBody.Type != updateBody.Type {
		t.Error("type mismatch between create and update")
	}
	if createBody.VolumeSize != updateBody.VolumeSize {
		t.Error("volume_size mismatch between create and update")
	}
	if createBody.FlavorID != updateBody.FlavorID {
		t.Error("flavor_id mismatch between create and update")
	}
	if createBody.Description != updateBody.Description {
		t.Error("description mismatch between create and update")
	}
}

// --- mapAPIResponseToState tests ---

func TestMapAPIResponseToState(t *testing.T) {
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	apiResp := &templateAPIResponse{
		ID:                  "tmpl-123",
		Name:                "prod-template",
		Type:                "linux",
		Description:         "Production template",
		VolumeSize:          100,
		VolDelOnTermination: true,
		Flavor:              &nestedIDName{ID: "f-large", Name: "Large"},
		Image:               &nestedIDName{ID: "img-prod", Name: "Ubuntu"},
		Snapshot:            nil,
		KeyName:             "prod-key",
		Network:             &nestedIDName{ID: "net-prod", Name: "Prod Net"},
		SubnetID:            "sub-prod",
		SecurityGroups:      []nestedIDName{{ID: "sg-prod-1", Name: "SG1"}, {ID: "sg-prod-2", Name: "SG2"}},
		IsInstanceSnapshot:  false,
		Status:              "active",
		Region:              "ap-south-noi-1",
		CreatedAt:           "2024-01-01T00:00:00Z",
		UpdatedAt:           "2024-01-02T00:00:00Z",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.ID.ValueString() != "tmpl-123" {
		t.Errorf("expected ID tmpl-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-template" {
		t.Errorf("expected Name prod-template, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Production template" {
		t.Errorf("expected Description 'Production template', got %s", model.Description.ValueString())
	}
	if model.VolumeSize.ValueInt64() != 100 {
		t.Errorf("expected VolumeSize 100, got %d", model.VolumeSize.ValueInt64())
	}
	if model.Status.ValueString() != "active" {
		t.Errorf("expected Status active, got %s", model.Status.ValueString())
	}
	if model.Region.ValueString() != "ap-south-noi-1" {
		t.Errorf("expected Region ap-south-noi-1, got %s", model.Region.ValueString())
	}
	if model.KeyName.ValueString() != "prod-key" {
		t.Errorf("expected KeyName prod-key, got %s", model.KeyName.ValueString())
	}
	if model.SnapshotID.IsNull() != true {
		t.Error("expected SnapshotID to remain null when API returns empty string")
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	apiResp := &templateAPIResponse{
		ID:             "tmpl-456",
		Name:           "basic",
		Type:           "linux",
		VolumeSize:     20,
		Flavor:         &nestedIDName{ID: "f-1", Name: "Small"},
		Network:        &nestedIDName{ID: "net-1", Name: "Net"},
		SubnetID:       "sub-1",
		SecurityGroups: []nestedIDName{{ID: "sg-1", Name: "SG"}},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
	if !model.ImageID.IsNull() {
		t.Error("expected ImageID to remain null when API returns empty string")
	}
	if !model.SnapshotID.IsNull() {
		t.Error("expected SnapshotID to remain null when API returns empty string")
	}
	if !model.KeyName.IsNull() {
		t.Error("expected KeyName to remain null when API returns empty string")
	}
	if model.Status.ValueString() != "" {
		t.Errorf("expected empty Status, got %s", model.Status.ValueString())
	}
}

func TestMapAPIResponseToState_NestedObjects(t *testing.T) {
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	apiResp := &templateAPIResponse{
		ID:                  "tmpl-nested-001",
		Name:                "nested-template",
		Type:                "linux",
		Description:         "Template with nested objects",
		VolumeSize:          80,
		VolDelOnTermination: true,
		Flavor:              &nestedIDName{ID: "flv-nested-1", Name: "4vCPU-8GB"},
		Image:               &nestedIDName{ID: "img-nested-1", Name: "Ubuntu 22.04"},
		Snapshot:            &nestedIDName{ID: "snap-nested-1", Name: "My Snapshot"},
		KeyName:             "my-ssh-key",
		Network:             &nestedIDName{ID: "net-nested-1", Name: "Prod VPC"},
		SubnetID:            "sub-nested-1",
		SecurityGroups: []nestedIDName{
			{ID: "sg-nested-1", Name: "Default SG"},
			{ID: "sg-nested-2", Name: "Web SG"},
			{ID: "sg-nested-3", Name: "DB SG"},
		},
		IsInstanceSnapshot: true,
		Status:              "active",
		Region:              "ap-south-noi-1",
		CreatedAt:           "2026-01-01T00:00:00Z",
		UpdatedAt:           "2026-01-02T00:00:00Z",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.FlavorID.ValueString() != "flv-nested-1" {
		t.Errorf("expected FlavorID flv-nested-1, got %s", model.FlavorID.ValueString())
	}
	if model.ImageID.ValueString() != "img-nested-1" {
		t.Errorf("expected ImageID img-nested-1, got %s", model.ImageID.ValueString())
	}
	if model.SnapshotID.ValueString() != "snap-nested-1" {
		t.Errorf("expected SnapshotID snap-nested-1, got %s", model.SnapshotID.ValueString())
	}
	if model.NetworkID.ValueString() != "net-nested-1" {
		t.Errorf("expected NetworkID net-nested-1, got %s", model.NetworkID.ValueString())
	}
	var sgIDs []string
	model.SecurityGroups.ElementsAs(context.Background(), &sgIDs, false)
	if len(sgIDs) != 3 {
		t.Fatalf("expected 3 security groups, got %d", len(sgIDs))
	}
	if sgIDs[0] != "sg-nested-1" || sgIDs[1] != "sg-nested-2" || sgIDs[2] != "sg-nested-3" {
		t.Errorf("expected sg IDs [sg-nested-1, sg-nested-2, sg-nested-3], got %v", sgIDs)
	}
	if model.KeyName.ValueString() != "my-ssh-key" {
		t.Errorf("expected KeyName my-ssh-key, got %s", model.KeyName.ValueString())
	}
	if model.IsInstanceSnapshot.ValueBool() != true {
		t.Error("expected IsInstanceSnapshot to be true")
	}
	if model.Description.ValueString() != "Template with nested objects" {
		t.Errorf("expected Description 'Template with nested objects', got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_NilFlavorAndNetwork(t *testing.T) {
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	apiResp := &templateAPIResponse{
		ID:             "tmpl-nil",
		Name:           "nil-nested",
		Type:           "linux",
		VolumeSize:     20,
		Flavor:         nil,
		Image:          nil,
		Snapshot:       nil,
		Network:        nil,
		SubnetID:       "sub-1",
		SecurityGroups: []nestedIDName{{ID: "sg-1", Name: "SG"}},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.FlavorID.ValueString() != "" {
		t.Errorf("expected empty FlavorID when Flavor is nil, got %s", model.FlavorID.ValueString())
	}
	if !model.ImageID.IsNull() {
		t.Error("expected ImageID to remain null when Image is nil")
	}
	if model.NetworkID.ValueString() != "" {
		t.Errorf("expected empty NetworkID when Network is nil, got %s", model.NetworkID.ValueString())
	}
}

func TestMapAPIResponseToState_EmptySecurityGroups(t *testing.T) {
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
		SecurityGroups: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("sg-original"),
		}),
	}

	apiResp := &templateAPIResponse{
		ID:             "tmpl-empty-sg",
		Name:           "empty-sg",
		Type:           "linux",
		VolumeSize:     20,
		Flavor:         &nestedIDName{ID: "f-1", Name: "Small"},
		Network:        &nestedIDName{ID: "net-1", Name: "Net"},
		SubnetID:       "sub-1",
		SecurityGroups: []nestedIDName{},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	// When API returns empty SGs, the original list should be preserved
	// (SecurityGroupIDs returns empty slice, len check skips reassignment)
	var sgIDs []string
	model.SecurityGroups.ElementsAs(context.Background(), &sgIDs, false)
	if len(sgIDs) != 1 || sgIDs[0] != "sg-original" {
		t.Errorf("expected original security groups preserved when API returns empty, got %v", sgIDs)
	}
}

func TestMapAPIResponseToState_ComputedFieldsAlwaysSet(t *testing.T) {
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	apiResp := &templateAPIResponse{
		ID:             "tmpl-computed",
		Name:           "computed-test",
		Type:           "linux",
		VolumeSize:     20,
		Flavor:         &nestedIDName{ID: "f-1", Name: "Small"},
		Network:        &nestedIDName{ID: "net-1", Name: "Net"},
		SubnetID:       "sub-1",
		SecurityGroups: []nestedIDName{{ID: "sg-1", Name: "SG"}},
		// Status, Region, CreatedAt, UpdatedAt all empty
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	// All computed fields should be set (never left unknown)
	if model.Status.IsUnknown() {
		t.Error("Status should never be unknown after mapAPIResponseToState")
	}
	if model.Region.IsUnknown() {
		t.Error("Region should never be unknown after mapAPIResponseToState")
	}
	if model.CreatedAt.IsUnknown() {
		t.Error("CreatedAt should never be unknown after mapAPIResponseToState")
	}
	if model.UpdatedAt.IsUnknown() {
		t.Error("UpdatedAt should never be unknown after mapAPIResponseToState")
	}

	// Empty string when API returns empty
	if model.Status.ValueString() != "" {
		t.Errorf("expected empty Status, got %s", model.Status.ValueString())
	}
	if model.Region.ValueString() != "" {
		t.Errorf("expected empty Region, got %s", model.Region.ValueString())
	}
	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "" {
		t.Errorf("expected empty UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionNotNullOverwritesNull(t *testing.T) {
	// When model has null description but API returns a value, it should be set
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	apiResp := &templateAPIResponse{
		ID:             "tmpl-desc",
		Name:           "desc-test",
		Type:           "linux",
		Description:    "New description from API",
		VolumeSize:     20,
		Flavor:         &nestedIDName{ID: "f-1", Name: "Small"},
		Network:        &nestedIDName{ID: "net-1", Name: "Net"},
		SubnetID:       "sub-1",
		SecurityGroups: []nestedIDName{{ID: "sg-1", Name: "SG"}},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.Description.IsNull() {
		t.Error("expected Description to be set when API returns non-empty value")
	}
	if model.Description.ValueString() != "New description from API" {
		t.Errorf("expected Description 'New description from API', got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_VolDelOnTerminationFalse(t *testing.T) {
	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	apiResp := &templateAPIResponse{
		ID:                  "tmpl-vdot",
		Name:                "vdot-false",
		Type:                "linux",
		VolumeSize:          20,
		VolDelOnTermination: false,
		Flavor:              &nestedIDName{ID: "f-1", Name: "Small"},
		Network:             &nestedIDName{ID: "net-1", Name: "Net"},
		SubnetID:            "sub-1",
		SecurityGroups:      []nestedIDName{{ID: "sg-1", Name: "SG"}},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.VolDelOnTermination.ValueBool() != false {
		t.Error("expected VolDelOnTermination false")
	}
}

// --- Helper method tests ---

func TestTemplateAPIResponse_FlavorID(t *testing.T) {
	tests := []struct {
		name     string
		resp     *templateAPIResponse
		expected string
	}{
		{"with flavor", &templateAPIResponse{Flavor: &nestedIDName{ID: "f-1", Name: "Small"}}, "f-1"},
		{"nil flavor", &templateAPIResponse{Flavor: nil}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.resp.FlavorID(); got != tc.expected {
				t.Errorf("FlavorID() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestTemplateAPIResponse_ImageID(t *testing.T) {
	tests := []struct {
		name     string
		resp     *templateAPIResponse
		expected string
	}{
		{"with image", &templateAPIResponse{Image: &nestedIDName{ID: "img-1", Name: "Ubuntu"}}, "img-1"},
		{"nil image", &templateAPIResponse{Image: nil}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.resp.ImageID(); got != tc.expected {
				t.Errorf("ImageID() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestTemplateAPIResponse_SnapshotID(t *testing.T) {
	tests := []struct {
		name     string
		resp     *templateAPIResponse
		expected string
	}{
		{"with snapshot", &templateAPIResponse{Snapshot: &nestedIDName{ID: "snap-1", Name: "Snap"}}, "snap-1"},
		{"nil snapshot", &templateAPIResponse{Snapshot: nil}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.resp.SnapshotID(); got != tc.expected {
				t.Errorf("SnapshotID() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestTemplateAPIResponse_NetworkID(t *testing.T) {
	tests := []struct {
		name     string
		resp     *templateAPIResponse
		expected string
	}{
		{"with network", &templateAPIResponse{Network: &nestedIDName{ID: "net-1", Name: "Net"}}, "net-1"},
		{"nil network", &templateAPIResponse{Network: nil}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.resp.NetworkID(); got != tc.expected {
				t.Errorf("NetworkID() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestTemplateAPIResponse_SecurityGroupIDs(t *testing.T) {
	tests := []struct {
		name     string
		resp     *templateAPIResponse
		expected []string
	}{
		{
			"multiple groups",
			&templateAPIResponse{SecurityGroups: []nestedIDName{{ID: "sg-1", Name: "A"}, {ID: "sg-2", Name: "B"}}},
			[]string{"sg-1", "sg-2"},
		},
		{
			"single group",
			&templateAPIResponse{SecurityGroups: []nestedIDName{{ID: "sg-only", Name: "Only"}}},
			[]string{"sg-only"},
		},
		{
			"empty groups",
			&templateAPIResponse{SecurityGroups: []nestedIDName{}},
			[]string{},
		},
		{
			"nil groups",
			&templateAPIResponse{SecurityGroups: nil},
			[]string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.resp.SecurityGroupIDs()
			if len(got) != len(tc.expected) {
				t.Fatalf("SecurityGroupIDs() len = %d, want %d", len(got), len(tc.expected))
			}
			for i, id := range got {
				if id != tc.expected[i] {
					t.Errorf("SecurityGroupIDs()[%d] = %q, want %q", i, id, tc.expected[i])
				}
			}
		})
	}
}

// --- JSON roundtrip / API response parsing tests ---

func TestTemplateAPIResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"id": "tmpl-json-001",
		"name": "json-parsed",
		"type": "linux",
		"description": "Parsed from JSON",
		"volume_size": 100,
		"vol_del_on_termination": true,
		"flavor": {"id": "f-json", "name": "JSON Flavor"},
		"image": {"id": "img-json", "name": "JSON Image"},
		"snapshot": null,
		"key_name": "json-key",
		"network": {"id": "net-json", "name": "JSON Net"},
		"subnet_id": "sub-json",
		"security_groups": [
			{"id": "sg-json-1", "name": "SG1"},
			{"id": "sg-json-2", "name": "SG2"}
		],
		"is_instance_snapshot": false,
		"status": "active",
		"region": "us-east-1",
		"created_at": "2026-03-01T00:00:00Z",
		"updated_at": "2026-03-02T00:00:00Z"
	}`

	var resp templateAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "tmpl-json-001" {
		t.Errorf("expected ID tmpl-json-001, got %s", resp.ID)
	}
	if resp.FlavorID() != "f-json" {
		t.Errorf("expected FlavorID f-json, got %s", resp.FlavorID())
	}
	if resp.ImageID() != "img-json" {
		t.Errorf("expected ImageID img-json, got %s", resp.ImageID())
	}
	if resp.SnapshotID() != "" {
		t.Errorf("expected empty SnapshotID for null, got %s", resp.SnapshotID())
	}
	if resp.NetworkID() != "net-json" {
		t.Errorf("expected NetworkID net-json, got %s", resp.NetworkID())
	}
	sgIDs := resp.SecurityGroupIDs()
	if len(sgIDs) != 2 {
		t.Fatalf("expected 2 security groups, got %d", len(sgIDs))
	}
	if sgIDs[0] != "sg-json-1" || sgIDs[1] != "sg-json-2" {
		t.Errorf("unexpected security group IDs: %v", sgIDs)
	}
}

func TestTemplateAPIResponse_JSONParsing_MinimalResponse(t *testing.T) {
	jsonData := `{
		"id": "tmpl-min",
		"name": "minimal",
		"type": "linux",
		"volume_size": 20,
		"vol_del_on_termination": false,
		"flavor": {"id": "f-1", "name": "Small"},
		"network": {"id": "net-1", "name": "Default"},
		"subnet_id": "sub-1",
		"security_groups": [{"id": "sg-1", "name": "Default"}],
		"is_instance_snapshot": false
	}`

	var resp templateAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Description != "" {
		t.Errorf("expected empty Description, got %s", resp.Description)
	}
	if resp.Image != nil {
		t.Errorf("expected nil Image, got %v", resp.Image)
	}
	if resp.Snapshot != nil {
		t.Errorf("expected nil Snapshot, got %v", resp.Snapshot)
	}
	if resp.Status != "" {
		t.Errorf("expected empty Status, got %s", resp.Status)
	}
}

func TestMapAPIResponseToState_FromParsedJSON(t *testing.T) {
	jsonData := `{
		"id": "tmpl-e2e",
		"name": "e2e-template",
		"type": "windows",
		"description": "End to end test",
		"volume_size": 200,
		"vol_del_on_termination": false,
		"flavor": {"id": "f-large", "name": "8vCPU-16GB"},
		"image": {"id": "img-win", "name": "Windows Server 2022"},
		"snapshot": null,
		"key_name": "win-key",
		"network": {"id": "net-prod", "name": "Production VPC"},
		"subnet_id": "sub-prod",
		"security_groups": [{"id": "sg-rdp", "name": "RDP Access"}],
		"is_instance_snapshot": false,
		"status": "active",
		"region": "eu-west-1",
		"created_at": "2026-02-01T12:00:00Z",
		"updated_at": "2026-02-15T08:30:00Z"
	}`

	var apiResp templateAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &apiResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	model := &autoScalingTemplateModel{
		Description: types.StringNull(),
		ImageID:     types.StringNull(),
		SnapshotID:  types.StringNull(),
		KeyName:     types.StringNull(),
	}

	mapAPIResponseToState(context.Background(), model, &apiResp)

	if model.ID.ValueString() != "tmpl-e2e" {
		t.Errorf("expected ID tmpl-e2e, got %s", model.ID.ValueString())
	}
	if model.Type.ValueString() != "windows" {
		t.Errorf("expected Type windows, got %s", model.Type.ValueString())
	}
	if model.FlavorID.ValueString() != "f-large" {
		t.Errorf("expected FlavorID f-large, got %s", model.FlavorID.ValueString())
	}
	if model.ImageID.ValueString() != "img-win" {
		t.Errorf("expected ImageID img-win, got %s", model.ImageID.ValueString())
	}
	if !model.SnapshotID.IsNull() {
		t.Error("expected SnapshotID null when API has null snapshot")
	}
	if model.Region.ValueString() != "eu-west-1" {
		t.Errorf("expected Region eu-west-1, got %s", model.Region.ValueString())
	}
	if model.CreatedAt.ValueString() != "2026-02-01T12:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2026-02-15T08:30:00Z" {
		t.Errorf("expected UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
}

// --- NewResource test ---

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- createResponseID test ---

func TestCreateResponseID_Parse(t *testing.T) {
	jsonData := `{"id": "tmpl-new-123"}`
	var resp createResponseID
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.ID != "tmpl-new-123" {
		t.Errorf("expected ID tmpl-new-123, got %s", resp.ID)
	}
}

func TestCreateResponseID_EmptyID(t *testing.T) {
	jsonData := `{"id": ""}`
	var resp createResponseID
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %s", resp.ID)
	}
}
