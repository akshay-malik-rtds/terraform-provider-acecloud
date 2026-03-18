package instance

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}

	vol1, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(40),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("hourly"),
	})

	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{vol1})

	networkList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("net-111"),
	})

	sgList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("sg-222"),
	})

	metadataMap, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"env": types.StringValue("prod"),
	})

	tagsList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("web"),
	})

	plan := &instanceResourceModel{
		Name:                types.StringValue("test-server"),
		Description:         types.StringValue("A test server"),
		FlavorID:            types.StringValue("flavor-uuid"),
		BootUUID:            types.StringValue("image-uuid"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(true),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("mumbai-1a"),
		Metadata:            metadataMap,
		KeyName:             types.StringValue("my-key"),
		UserData:            types.StringValue("IyEvYmluL2Jhc2g="),
		ServerGroupID:       types.StringNull(),
		ConfigDrive:         types.BoolValue(false),
		AdminPassword:       types.StringNull(),
		BillingType:         types.StringValue("monthly"),
		Tags:                tagsList,
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if body.Name != "test-server" {
		t.Errorf("expected name test-server, got %s", body.Name)
	}
	if body.Count != 1 {
		t.Errorf("expected count 1, got %d", body.Count)
	}
	if body.Flavor != "flavor-uuid" {
		t.Errorf("expected flavor flavor-uuid, got %s", body.Flavor)
	}
	if body.BootUUID != "image-uuid" {
		t.Errorf("expected boot_uuid image-uuid, got %s", body.BootUUID)
	}
	if body.SourceType != "image" {
		t.Errorf("expected source_type image, got %s", body.SourceType)
	}
	if body.DeleteOnTermination != true {
		t.Error("expected delete_on_termination true")
	}
	if body.AvailabilityZone != "mumbai-1a" {
		t.Errorf("expected availability_zone mumbai-1a, got %s", body.AvailabilityZone)
	}
	if body.Key != "my-key" {
		t.Errorf("expected key my-key, got %s", body.Key)
	}
	if body.Script != "IyEvYmluL2Jhc2g=" {
		t.Errorf("expected script IyEvYmluL2Jhc2g=, got %s", body.Script)
	}
	if body.Description != "A test server" {
		t.Errorf("expected description 'A test server', got %s", body.Description)
	}
	if body.BillingType != "monthly" {
		t.Errorf("expected billing_type monthly, got %s", body.BillingType)
	}

	// Volumes
	if len(body.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(body.Volumes))
	}
	if body.Volumes[0].Size != 40 {
		t.Errorf("expected volume size 40, got %d", body.Volumes[0].Size)
	}
	if body.Volumes[0].Boot != true {
		t.Error("expected volume boot true")
	}
	// buildCreateRequest maps "ssd" → backend name for the API.
	if body.Volumes[0].VolumeType != "NVMe based High IOPS Storage" {
		t.Errorf("expected volume_type 'NVMe based High IOPS Storage', got %s", body.Volumes[0].VolumeType)
	}

	// Network
	if len(body.Network) != 1 || body.Network[0] != "net-111" {
		t.Errorf("expected network [net-111], got %v", body.Network)
	}

	// Security groups
	if len(body.SecurityGroup) != 1 || body.SecurityGroup[0] != "sg-222" {
		t.Errorf("expected security_group [sg-222], got %v", body.SecurityGroup)
	}

	// Metadata
	if body.Metadata["env"] != "prod" {
		t.Errorf("expected metadata env=prod, got %v", body.Metadata)
	}

	// Tags
	if len(body.Tags) != 1 || body.Tags[0] != "web" {
		t.Errorf("expected tags [web], got %v", body.Tags)
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}

	vol1, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(20),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("hdd"),
		"billing_type": types.StringValue("hourly"),
	})

	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{vol1})

	networkList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("net-1"),
	})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("sg-1"),
	})

	plan := &instanceResourceModel{
		Name:                types.StringValue("minimal-server"),
		Description:         types.StringNull(),
		FlavorID:            types.StringValue("flv-1"),
		BootUUID:            types.StringValue("img-1"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(false),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("az-1"),
		Metadata:            types.MapNull(types.StringType),
		KeyName:             types.StringNull(),
		UserData:            types.StringNull(),
		ServerGroupID:       types.StringNull(),
		ConfigDrive:         types.BoolValue(false),
		AdminPassword:       types.StringNull(),
		BillingType:         types.StringValue("monthly"),
		Tags:                types.ListNull(types.StringType),
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if body.Description != "" {
		t.Errorf("expected empty description, got %s", body.Description)
	}
	if body.Key != "" {
		t.Errorf("expected empty key, got %s", body.Key)
	}
	if body.Script != "" {
		t.Errorf("expected empty script, got %s", body.Script)
	}
	if body.Metadata != nil {
		t.Errorf("expected nil metadata, got %v", body.Metadata)
	}
	if body.Tags != nil {
		t.Errorf("expected nil tags, got %v", body.Tags)
	}
}

func TestMapReadResponseToState(t *testing.T) {
	ctx := context.Background()

	// Set up initial state with appropriate types.
	networkList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("net-1"),
	})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("sg-1"),
	})

	state := &instanceResourceModel{
		NetworkIDs:       networkList,
		SecurityGroupIDs: sgList,
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-123",
		Name:             "web-1",
		Description:      "Web server",
		FlavorID:         "flv-big",
		ImageID:          "img-ubuntu",
		SecurityGroups:   []string{"sg-bbb"},
		AvailabilityZone: "mumbai-1a",
		Key:              "deploy-key",
		Status:           "ACTIVE",
		Metadata:         map[string]string{"env": "production"},
		Tags:             []string{"web", "prod"},
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if state.ID.ValueString() != "inst-123" {
		t.Errorf("expected ID inst-123, got %s", state.ID.ValueString())
	}
	if state.Name.ValueString() != "web-1" {
		t.Errorf("expected Name web-1, got %s", state.Name.ValueString())
	}
	if state.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", state.Status.ValueString())
	}
	if state.Description.ValueString() != "Web server" {
		t.Errorf("expected Description 'Web server', got %s", state.Description.ValueString())
	}
	if state.KeyName.ValueString() != "deploy-key" {
		t.Errorf("expected KeyName deploy-key, got %s", state.KeyName.ValueString())
	}
}

func TestMapReadResponseToState_EmptyOptionals(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-456",
		Name:             "bare-server",
		Status:           "BUILD",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if !state.Description.IsNull() {
		t.Error("expected Description to remain null")
	}
	if !state.KeyName.IsNull() {
		t.Error("expected KeyName to remain null")
	}
	if !state.ServerGroupID.IsNull() {
		t.Error("expected ServerGroupID to remain null")
	}
	if !state.Tags.IsNull() {
		t.Error("expected Tags to remain null")
	}
	if !state.Metadata.IsNull() {
		t.Error("expected Metadata to remain null")
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest – additional scenarios
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_MultipleVolumes(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}

	bootVol, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(40),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("monthly"),
	})
	dataVol, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(100),
		"boot":         types.BoolValue(false),
		"volume_type":  types.StringValue("hdd"),
		"billing_type": types.StringValue("hourly"),
	})

	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{bootVol, dataVol})
	networkList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("net-1")})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("sg-1")})

	plan := &instanceResourceModel{
		Name:                types.StringValue("multi-vol"),
		Description:         types.StringNull(),
		FlavorID:            types.StringValue("flv-1"),
		BootUUID:            types.StringValue("img-1"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(true),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("az-1"),
		Metadata:            types.MapNull(types.StringType),
		KeyName:             types.StringNull(),
		UserData:            types.StringNull(),
		ServerGroupID:       types.StringNull(),
		ConfigDrive:         types.BoolValue(false),
		AdminPassword:       types.StringNull(),
		BillingType:         types.StringValue("monthly"),
		Tags:                types.ListNull(types.StringType),
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if len(body.Volumes) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(body.Volumes))
	}

	// Boot volume.
	if body.Volumes[0].Size != 40 {
		t.Errorf("expected boot volume size 40, got %d", body.Volumes[0].Size)
	}
	if !body.Volumes[0].Boot {
		t.Error("expected boot volume boot=true")
	}
	if body.Volumes[0].VolumeType != "NVMe based High IOPS Storage" {
		t.Errorf("expected boot volume_type 'NVMe based High IOPS Storage', got %s", body.Volumes[0].VolumeType)
	}
	if body.Volumes[0].BillingType != "monthly" {
		t.Errorf("expected boot billing_type 'monthly', got %s", body.Volumes[0].BillingType)
	}

	// Data volume.
	if body.Volumes[1].Size != 100 {
		t.Errorf("expected data volume size 100, got %d", body.Volumes[1].Size)
	}
	if body.Volumes[1].Boot {
		t.Error("expected data volume boot=false")
	}
	if body.Volumes[1].VolumeType != "HDD based Storage" {
		t.Errorf("expected data volume_type 'HDD based Storage', got %s", body.Volumes[1].VolumeType)
	}
	if body.Volumes[1].BillingType != "hourly" {
		t.Errorf("expected data billing_type 'hourly', got %s", body.Volumes[1].BillingType)
	}
}

func TestBuildCreateRequest_MultipleNetworks(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}
	vol, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(20),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("hourly"),
	})
	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{vol})

	networkList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("net-aaa"),
		types.StringValue("net-bbb"),
		types.StringValue("net-ccc"),
	})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("sg-1")})

	plan := &instanceResourceModel{
		Name:                types.StringValue("multi-net"),
		Description:         types.StringNull(),
		FlavorID:            types.StringValue("flv-1"),
		BootUUID:            types.StringValue("img-1"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(false),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("az-1"),
		Metadata:            types.MapNull(types.StringType),
		KeyName:             types.StringNull(),
		UserData:            types.StringNull(),
		ServerGroupID:       types.StringNull(),
		ConfigDrive:         types.BoolValue(false),
		AdminPassword:       types.StringNull(),
		BillingType:         types.StringValue("monthly"),
		Tags:                types.ListNull(types.StringType),
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if len(body.Network) != 3 {
		t.Fatalf("expected 3 networks, got %d", len(body.Network))
	}
	expected := []string{"net-aaa", "net-bbb", "net-ccc"}
	for i, exp := range expected {
		if body.Network[i] != exp {
			t.Errorf("network[%d]: expected %s, got %s", i, exp, body.Network[i])
		}
	}
}

func TestBuildCreateRequest_ServerGroupID(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}
	vol, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(20),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("hourly"),
	})
	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{vol})
	networkList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("net-1")})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("sg-1")})

	plan := &instanceResourceModel{
		Name:                types.StringValue("sg-test"),
		Description:         types.StringNull(),
		FlavorID:            types.StringValue("flv-1"),
		BootUUID:            types.StringValue("img-1"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(true),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("az-1"),
		Metadata:            types.MapNull(types.StringType),
		KeyName:             types.StringNull(),
		UserData:            types.StringNull(),
		ServerGroupID:       types.StringValue("sgrp-uuid-123"),
		ConfigDrive:         types.BoolValue(false),
		AdminPassword:       types.StringNull(),
		BillingType:         types.StringValue("monthly"),
		Tags:                types.ListNull(types.StringType),
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if body.ServerGroupID != "sgrp-uuid-123" {
		t.Errorf("expected server_group_id 'sgrp-uuid-123', got %s", body.ServerGroupID)
	}
}

func TestBuildCreateRequest_ConfigDrive(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}
	vol, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(20),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("hourly"),
	})
	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{vol})
	networkList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("net-1")})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("sg-1")})

	plan := &instanceResourceModel{
		Name:                types.StringValue("cfg-drive-test"),
		Description:         types.StringNull(),
		FlavorID:            types.StringValue("flv-1"),
		BootUUID:            types.StringValue("img-1"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(true),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("az-1"),
		Metadata:            types.MapNull(types.StringType),
		KeyName:             types.StringNull(),
		UserData:            types.StringNull(),
		ServerGroupID:       types.StringNull(),
		ConfigDrive:         types.BoolValue(true),
		AdminPassword:       types.StringNull(),
		BillingType:         types.StringValue("monthly"),
		Tags:                types.ListNull(types.StringType),
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if !body.ConfigDrive {
		t.Error("expected config_drive true, got false")
	}
}

func TestBuildCreateRequest_AdminPassword(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}
	vol, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(20),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("hourly"),
	})
	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{vol})
	networkList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("net-1")})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("sg-1")})

	plan := &instanceResourceModel{
		Name:                types.StringValue("admin-pw-test"),
		Description:         types.StringNull(),
		FlavorID:            types.StringValue("flv-1"),
		BootUUID:            types.StringValue("img-1"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(true),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("az-1"),
		Metadata:            types.MapNull(types.StringType),
		KeyName:             types.StringNull(),
		UserData:            types.StringNull(),
		ServerGroupID:       types.StringNull(),
		ConfigDrive:         types.BoolValue(false),
		AdminPassword:       types.StringValue("c2VjcmV0cGFzcw=="),
		BillingType:         types.StringValue("monthly"),
		Tags:                types.ListNull(types.StringType),
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if body.AdminPassword != "c2VjcmV0cGFzcw==" {
		t.Errorf("expected admin_password 'c2VjcmV0cGFzcw==', got %s", body.AdminPassword)
	}
}

// ---------------------------------------------------------------------------
// Volume type mapping helpers
// ---------------------------------------------------------------------------

func TestVolumeTypeToBackend(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ssd", "NVMe based High IOPS Storage"},
		{"nvme", "NVMe based High IOPS Storage"},
		{"hdd", "HDD based Storage"},
		{"unknown-type", "unknown-type"},       // passthrough for unknown
		{"CustomStorage", "CustomStorage"},      // passthrough for unknown
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := volumeTypeToBackend(tc.input)
			if got != tc.expected {
				t.Errorf("volumeTypeToBackend(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestVolumeTypeFromBackend(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NVMe based High IOPS Storage", "ssd"},
		{"HDD based Storage", "hdd"},
		{"UnknownBackendType", "UnknownBackendType"}, // passthrough
		{"ssd", "ssd"},                               // passthrough (already user-friendly)
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := volumeTypeFromBackend(tc.input)
			if got != tc.expected {
				t.Errorf("volumeTypeFromBackend(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestVolumeTypeRoundTrip(t *testing.T) {
	// Round-trip: volumeTypeFromBackend(volumeTypeToBackend(x)) == x
	userTypes := []string{"ssd", "hdd"}
	for _, ut := range userTypes {
		t.Run(ut, func(t *testing.T) {
			backend := volumeTypeToBackend(ut)
			roundTripped := volumeTypeFromBackend(backend)
			if roundTripped != ut {
				t.Errorf("round-trip failed for %q: toBackend=%q, fromBackend=%q", ut, backend, roundTripped)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Regex validation tests
// ---------------------------------------------------------------------------

func TestResourceNameRegex(t *testing.T) {
	tests := []struct {
		input string
		match bool
	}{
		{"valid-name-1", true},
		{"test123", true},
		{"A", true},
		{"my-server-01", true},
		{"invalid name!", false},
		{"has space", false},
		{"has@symbol", false},
		{"with.dot", false},
		{"under_score", false},
		{"", false},
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := re.MatchString(tc.input)
			if got != tc.match {
				t.Errorf("resourceNameRegex.MatchString(%q) = %v, want %v", tc.input, got, tc.match)
			}
		})
	}

	// Also verify the package-level regex behaves identically.
	for _, tc := range tests {
		got := resourceNameRegex.MatchString(tc.input)
		if got != tc.match {
			t.Errorf("package resourceNameRegex.MatchString(%q) = %v, want %v", tc.input, got, tc.match)
		}
	}
}

func TestDescriptionRegex(t *testing.T) {
	tests := []struct {
		input string
		match bool
	}{
		{"valid desc, here.", true},
		{"test-desc_1", true},
		{"simple", true},
		{"with spaces and commas, periods.", true},
		{"", true}, // empty is valid (the * quantifier allows zero matches)
		{"invalid<html>", false},
		{"has>bracket", false},
		{"semi;colon", false},
		{"quote\"inside", false},
		{"back\\slash", false},
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9_\-., ]*$`)

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := re.MatchString(tc.input)
			if got != tc.match {
				t.Errorf("descriptionRegex.MatchString(%q) = %v, want %v", tc.input, got, tc.match)
			}
		})
	}

	// Also verify the package-level regex behaves identically.
	for _, tc := range tests {
		got := descriptionRegex.MatchString(tc.input)
		if got != tc.match {
			t.Errorf("package descriptionRegex.MatchString(%q) = %v, want %v", tc.input, got, tc.match)
		}
	}
}

// ---------------------------------------------------------------------------
// mapReadResponseToState – additional scenarios
// ---------------------------------------------------------------------------

func TestMapReadResponseToState_MultipleVolumes(t *testing.T) {
	ctx := context.Background()

	// Set up initial state with two volumes so the mapping has existing data.
	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}
	existVol1, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(40),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("hourly"),
	})
	existVol2, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(200),
		"boot":         types.BoolValue(false),
		"volume_type":  types.StringValue("hdd"),
		"billing_type": types.StringValue("hourly"),
	})
	existVolList, _ := types.ListValue(volumeObjType, []attr.Value{existVol1, existVol2})

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
		Volumes:          existVolList,
	}

	instance := &readInstanceResponse{
		ID:               "inst-multi-vol",
		Name:             "multi-vol-server",
		Status:           "ACTIVE",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Extract volume models from state.
	var volumes []instanceVolumeModel
	diags.Append(state.Volumes.ElementsAs(ctx, &volumes, false)...)
	if diags.HasError() {
		t.Fatalf("error extracting volumes: %v", diags.Errors())
	}

	if len(volumes) != 2 {
		t.Fatalf("expected 2 volumes in state, got %d", len(volumes))
	}

	// Boot volume.
	if volumes[0].Size.ValueInt64() != 40 {
		t.Errorf("volume[0] size: expected 40, got %d", volumes[0].Size.ValueInt64())
	}
	if !volumes[0].Boot.ValueBool() {
		t.Error("volume[0] boot: expected true")
	}
	// The user had "ssd" and backend returns the matching backend name, so it should preserve "ssd".
	if volumes[0].VolumeType.ValueString() != "ssd" {
		t.Errorf("volume[0] volume_type: expected 'ssd', got %s", volumes[0].VolumeType.ValueString())
	}

	// Data volume.
	if volumes[1].Size.ValueInt64() != 200 {
		t.Errorf("volume[1] size: expected 200, got %d", volumes[1].Size.ValueInt64())
	}
	if volumes[1].Boot.ValueBool() {
		t.Error("volume[1] boot: expected false")
	}
	if volumes[1].VolumeType.ValueString() != "hdd" {
		t.Errorf("volume[1] volume_type: expected 'hdd', got %s", volumes[1].VolumeType.ValueString())
	}
}

func TestMapReadResponseToState_WithMetadata(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-meta",
		Name:             "meta-server",
		Status:           "ACTIVE",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Metadata: map[string]string{
			"env":     "production",
			"team":    "platform",
			"version": "2.1.0",
		},
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if state.Metadata.IsNull() {
		t.Fatal("expected metadata to be populated, got null")
	}

	metaMap := make(map[string]string)
	diags.Append(state.Metadata.ElementsAs(ctx, &metaMap, false)...)
	if diags.HasError() {
		t.Fatalf("error extracting metadata: %v", diags.Errors())
	}

	if len(metaMap) != 3 {
		t.Fatalf("expected 3 metadata entries, got %d", len(metaMap))
	}
	expectedMeta := map[string]string{
		"env":     "production",
		"team":    "platform",
		"version": "2.1.0",
	}
	for k, v := range expectedMeta {
		if metaMap[k] != v {
			t.Errorf("metadata[%s]: expected %s, got %s", k, v, metaMap[k])
		}
	}
}

func TestMapReadResponseToState_WithTags(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-tags",
		Name:             "tagged-server",
		Status:           "ACTIVE",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Tags:             []string{"web", "prod", "us-east"},
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if state.Tags.IsNull() {
		t.Fatal("expected tags to be populated, got null")
	}

	var tags []string
	diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	if diags.HasError() {
		t.Fatalf("error extracting tags: %v", diags.Errors())
	}

	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	expectedTags := []string{"web", "prod", "us-east"}
	for i, exp := range expectedTags {
		if tags[i] != exp {
			t.Errorf("tags[%d]: expected %s, got %s", i, exp, tags[i])
		}
	}
}

// ---------------------------------------------------------------------------
// JSON serialization – verify API field names
// ---------------------------------------------------------------------------

func TestCreateRequestBody_JSONSerialization(t *testing.T) {
	ctx := context.Background()

	volumeObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"size":         types.Int64Type,
			"boot":         types.BoolType,
			"volume_type":  types.StringType,
			"billing_type": types.StringType,
		},
	}
	vol, _ := types.ObjectValue(volumeObjType.AttrTypes, map[string]attr.Value{
		"size":         types.Int64Value(40),
		"boot":         types.BoolValue(true),
		"volume_type":  types.StringValue("ssd"),
		"billing_type": types.StringValue("hourly"),
	})
	volumesList, _ := types.ListValue(volumeObjType, []attr.Value{vol})
	networkList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("net-1")})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("sg-1")})
	metadataMap, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"env": types.StringValue("test"),
	})
	tagsList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("ci"),
	})

	plan := &instanceResourceModel{
		Name:                types.StringValue("json-test"),
		Description:         types.StringValue("JSON test instance"),
		FlavorID:            types.StringValue("flv-json"),
		BootUUID:            types.StringValue("img-json"),
		SourceType:          types.StringValue("image"),
		DeleteOnTermination: types.BoolValue(true),
		Volumes:             volumesList,
		NetworkIDs:          networkList,
		SecurityGroupIDs:    sgList,
		AvailabilityZone:    types.StringValue("az-1"),
		Metadata:            metadataMap,
		KeyName:             types.StringValue("deploy-key"),
		UserData:            types.StringValue("dXNlcmRhdGE="),
		ServerGroupID:       types.StringValue("sgrp-1"),
		ConfigDrive:         types.BoolValue(true),
		AdminPassword:       types.StringValue("cGFzc3dvcmQ="),
		BillingType:         types.StringValue("monthly"),
		Tags:                tagsList,
	}

	body, diags := buildCreateRequest(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal into a generic map to verify JSON key names.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal into map failed: %v", err)
	}

	// Verify required top-level JSON field names match API expectations.
	requiredKeys := []string{
		"name",
		"count",
		"flavor",
		"billing_type",
		"boot_uuid",
		"source_type",
		"delete_on_termination",
		"volumes",
		"network",
		"security_group",
		"availability_zone",
		"config_drive",
	}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}

	// Verify optional keys present when set.
	optionalKeys := []string{
		"description",
		"metadata",
		"key",
		"script",
		"server_group_id",
		"admin_password",
		"tags",
	}
	for _, key := range optionalKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected optional JSON key %q to be present when set", key)
		}
	}

	// Verify concrete field values through the JSON.
	var deserialized createInstanceRequest
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("json.Unmarshal into createInstanceRequest failed: %v", err)
	}

	if deserialized.Name != "json-test" {
		t.Errorf("JSON round-trip name: expected 'json-test', got %s", deserialized.Name)
	}
	if deserialized.Count != 1 {
		t.Errorf("JSON round-trip count: expected 1, got %d", deserialized.Count)
	}
	if deserialized.Flavor != "flv-json" {
		t.Errorf("JSON round-trip flavor: expected 'flv-json', got %s", deserialized.Flavor)
	}
	if deserialized.BootUUID != "img-json" {
		t.Errorf("JSON round-trip boot_uuid: expected 'img-json', got %s", deserialized.BootUUID)
	}
	if deserialized.SourceType != "image" {
		t.Errorf("JSON round-trip source_type: expected 'image', got %s", deserialized.SourceType)
	}
	if !deserialized.ConfigDrive {
		t.Error("JSON round-trip config_drive: expected true")
	}
	if deserialized.AdminPassword != "cGFzc3dvcmQ=" {
		t.Errorf("JSON round-trip admin_password: expected 'cGFzc3dvcmQ=', got %s", deserialized.AdminPassword)
	}
	if deserialized.ServerGroupID != "sgrp-1" {
		t.Errorf("JSON round-trip server_group_id: expected 'sgrp-1', got %s", deserialized.ServerGroupID)
	}
	if deserialized.Key != "deploy-key" {
		t.Errorf("JSON round-trip key: expected 'deploy-key', got %s", deserialized.Key)
	}
	if deserialized.Script != "dXNlcmRhdGE=" {
		t.Errorf("JSON round-trip script: expected 'dXNlcmRhdGE=', got %s", deserialized.Script)
	}
}

// ---------------------------------------------------------------------------
// parseInstanceData Tests
// ---------------------------------------------------------------------------

func TestParseInstanceData_FullResponse(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-123",
		"name": "test-instance",
		"description": "my instance",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"key": "my-key",
		"metadata": {"env": "prod"},
		"tags": ["web", "production"],
		"flavor": {"id": "flav-1", "name": "m1.small"},
		"image": {"id": "img-1", "name": "ubuntu-22"},
		"config_drive": "True",
		"security_groups": [{"id": "sg-1", "name": "default"}, {"id": "sg-2", "name": "web"}]
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "inst-123" { t.Errorf("expected inst-123, got %s", resp.ID) }
	if resp.Name != "test-instance" { t.Errorf("expected test-instance, got %s", resp.Name) }
	if resp.Description != "my instance" { t.Errorf("expected 'my instance', got %s", resp.Description) }
	if resp.Status != "ACTIVE" { t.Errorf("expected ACTIVE, got %s", resp.Status) }
	if resp.AvailabilityZone != "nova" { t.Errorf("expected nova, got %s", resp.AvailabilityZone) }
	if resp.Key != "my-key" { t.Errorf("expected my-key, got %s", resp.Key) }
	if resp.FlavorID != "flav-1" { t.Errorf("expected flav-1, got %s", resp.FlavorID) }
	if resp.ImageID != "img-1" { t.Errorf("expected img-1, got %s", resp.ImageID) }
	if resp.ConfigDrive != true { t.Error("expected ConfigDrive true") }
	if len(resp.SecurityGroups) != 2 { t.Fatalf("expected 2 SGs, got %d", len(resp.SecurityGroups)) }
	if resp.SecurityGroups[0] != "sg-1" { t.Errorf("expected sg-1, got %s", resp.SecurityGroups[0]) }
	if resp.SecurityGroups[1] != "sg-2" { t.Errorf("expected sg-2, got %s", resp.SecurityGroups[1]) }
	if resp.Metadata["env"] != "prod" { t.Errorf("expected metadata env=prod") }
	if len(resp.Tags) != 2 { t.Errorf("expected 2 tags, got %d", len(resp.Tags)) }
}

func TestParseInstanceData_MinimalResponse(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-456",
		"name": "minimal",
		"status": "BUILD",
		"availability_zone": "nova",
		"config_drive": "",
		"flavor": {},
		"image": {}
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "inst-456" { t.Errorf("expected inst-456, got %s", resp.ID) }
	if resp.FlavorID != "" { t.Errorf("expected empty FlavorID, got %s", resp.FlavorID) }
	if resp.ImageID != "" { t.Errorf("expected empty ImageID, got %s", resp.ImageID) }
	if resp.ConfigDrive != false { t.Error("expected ConfigDrive false for empty string") }
	if len(resp.SecurityGroups) != 0 { t.Errorf("expected 0 SGs, got %d", len(resp.SecurityGroups)) }
}

func TestParseInstanceData_ConfigDriveTrue(t *testing.T) {
	data := json.RawMessage(`{"id":"i1","name":"n","status":"ACTIVE","availability_zone":"az","config_drive":"True"}`)
	resp, err := parseInstanceData(data)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if !resp.ConfigDrive { t.Error("expected ConfigDrive true for 'True'") }
}

func TestParseInstanceData_ConfigDriveFalse(t *testing.T) {
	data := json.RawMessage(`{"id":"i1","name":"n","status":"ACTIVE","availability_zone":"az","config_drive":""}`)
	resp, err := parseInstanceData(data)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if resp.ConfigDrive { t.Error("expected ConfigDrive false for empty string") }
}

func TestParseInstanceData_NoFlavorImage(t *testing.T) {
	data := json.RawMessage(`{"id":"i1","name":"n","status":"ACTIVE","availability_zone":"az"}`)
	resp, err := parseInstanceData(data)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if resp.FlavorID != "" { t.Errorf("expected empty FlavorID, got %s", resp.FlavorID) }
	if resp.ImageID != "" { t.Errorf("expected empty ImageID, got %s", resp.ImageID) }
}

func TestParseInstanceData_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`not json`)
	_, err := parseInstanceData(data)
	if err == nil { t.Fatal("expected error for invalid JSON") }
}

func TestParseInstanceData_SecurityGroupsNoID(t *testing.T) {
	data := json.RawMessage(`{"id":"i1","name":"n","status":"ACTIVE","availability_zone":"az","security_groups":[{"name":"default"}]}`)
	resp, err := parseInstanceData(data)
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	// SG without "id" key should be skipped
	if len(resp.SecurityGroups) != 0 { t.Errorf("expected 0 SGs (no id), got %d", len(resp.SecurityGroups)) }
}

// ===========================================================================
// Additional comprehensive tests
// ===========================================================================

// ---------------------------------------------------------------------------
// parseInstanceData — nested object extraction
// ---------------------------------------------------------------------------

func TestParseInstanceData_FlavorExtraction(t *testing.T) {
	tests := []struct {
		name       string
		flavorJSON string
		expectedID string
	}{
		{"full flavor object", `{"id":"flav-abc","name":"m1.large","ram":8192}`, "flav-abc"},
		{"flavor with only id", `{"id":"flav-only-id"}`, "flav-only-id"},
		{"empty flavor object", `{}`, ""},
		{"flavor with null id", `{"id":null,"name":"m1.small"}`, ""},
		{"flavor id as number", `{"id":123,"name":"m1.small"}`, ""}, // id must be string
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := json.RawMessage(`{"id":"inst-1","name":"test","status":"ACTIVE","availability_zone":"az","flavor":` + tc.flavorJSON + `}`)
			resp, err := parseInstanceData(data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.FlavorID != tc.expectedID {
				t.Errorf("expected FlavorID %q, got %q", tc.expectedID, resp.FlavorID)
			}
		})
	}
}

func TestParseInstanceData_ImageExtraction(t *testing.T) {
	tests := []struct {
		name       string
		imageJSON  string
		expectedID string
	}{
		{"full image object", `{"id":"img-abc","name":"ubuntu-22.04"}`, "img-abc"},
		{"image with only id", `{"id":"img-only-id"}`, "img-only-id"},
		{"empty image object", `{}`, ""},
		{"image with null id", `{"id":null}`, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := json.RawMessage(`{"id":"inst-1","name":"test","status":"ACTIVE","availability_zone":"az","image":` + tc.imageJSON + `}`)
			resp, err := parseInstanceData(data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.ImageID != tc.expectedID {
				t.Errorf("expected ImageID %q, got %q", tc.expectedID, resp.ImageID)
			}
		})
	}
}

func TestParseInstanceData_BootFromVolume_NullImage(t *testing.T) {
	// Boot-from-volume instances have "image": null or "image": "" in the API response.
	data := json.RawMessage(`{
		"id": "inst-bfv",
		"name": "boot-from-vol",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"flavor": {"id": "flav-1", "name": "m1.large"},
		"image": "",
		"config_drive": ""
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ImageID != "" {
		t.Errorf("expected empty ImageID for boot-from-volume, got %s", resp.ImageID)
	}
	if resp.FlavorID != "flav-1" {
		t.Errorf("expected FlavorID flav-1, got %s", resp.FlavorID)
	}
}

func TestParseInstanceData_BootFromVolume_NullImageJSON(t *testing.T) {
	// Some backends return "image": null.
	data := json.RawMessage(`{
		"id": "inst-bfv-null",
		"name": "boot-from-vol-null",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"flavor": {"id": "flav-2", "name": "m1.small"},
		"image": null,
		"config_drive": ""
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ImageID != "" {
		t.Errorf("expected empty ImageID for null image, got %s", resp.ImageID)
	}
}

// ---------------------------------------------------------------------------
// parseInstanceData — config_drive variations
// ---------------------------------------------------------------------------

func TestParseInstanceData_ConfigDriveVariations(t *testing.T) {
	tests := []struct {
		name     string
		cdValue  string
		expected bool
	}{
		{"True string", `"True"`, true},
		{"empty string", `""`, false},
		{"false string", `"false"`, false},
		{"true lowercase", `"true"`, false}, // Only "True" with capital T
		{"1 string", `"1"`, false},          // Only "True" is truthy
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := json.RawMessage(`{"id":"i1","name":"n","status":"ACTIVE","availability_zone":"az","config_drive":` + tc.cdValue + `}`)
			resp, err := parseInstanceData(data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.ConfigDrive != tc.expected {
				t.Errorf("expected ConfigDrive %v for %s, got %v", tc.expected, tc.cdValue, resp.ConfigDrive)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseInstanceData — security_groups extraction
// ---------------------------------------------------------------------------

func TestParseInstanceData_MultipleSGs(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-sgs",
		"name": "sg-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"security_groups": [
			{"id": "sg-1", "name": "default"},
			{"id": "sg-2", "name": "web"},
			{"id": "sg-3", "name": "db"}
		]
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.SecurityGroups) != 3 {
		t.Fatalf("expected 3 SGs, got %d", len(resp.SecurityGroups))
	}

	expected := []string{"sg-1", "sg-2", "sg-3"}
	for i, exp := range expected {
		if resp.SecurityGroups[i] != exp {
			t.Errorf("SecurityGroups[%d]: expected %s, got %s", i, exp, resp.SecurityGroups[i])
		}
	}
}

func TestParseInstanceData_EmptySecurityGroups(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-no-sgs",
		"name": "no-sg-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"security_groups": []
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.SecurityGroups) != 0 {
		t.Errorf("expected 0 SGs, got %d", len(resp.SecurityGroups))
	}
}

func TestParseInstanceData_SGsMixedWithAndWithoutID(t *testing.T) {
	// Some SGs have id, some don't — only those with id should be extracted.
	data := json.RawMessage(`{
		"id": "inst-mixed-sgs",
		"name": "mixed-sg-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"security_groups": [
			{"id": "sg-valid", "name": "default"},
			{"name": "no-id-sg"},
			{"id": "sg-also-valid", "name": "web"}
		]
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.SecurityGroups) != 2 {
		t.Fatalf("expected 2 SGs (skipping no-id), got %d", len(resp.SecurityGroups))
	}
	if resp.SecurityGroups[0] != "sg-valid" {
		t.Errorf("expected first SG sg-valid, got %s", resp.SecurityGroups[0])
	}
	if resp.SecurityGroups[1] != "sg-also-valid" {
		t.Errorf("expected second SG sg-also-valid, got %s", resp.SecurityGroups[1])
	}
}

// ---------------------------------------------------------------------------
// parseInstanceData — metadata and tags
// ---------------------------------------------------------------------------

func TestParseInstanceData_MetadataExtraction(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-meta",
		"name": "meta-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"metadata": {
			"env": "production",
			"team": "platform",
			"version": "3.2.1"
		}
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Metadata) != 3 {
		t.Fatalf("expected 3 metadata entries, got %d", len(resp.Metadata))
	}
	if resp.Metadata["env"] != "production" {
		t.Errorf("expected metadata env=production, got %s", resp.Metadata["env"])
	}
	if resp.Metadata["team"] != "platform" {
		t.Errorf("expected metadata team=platform, got %s", resp.Metadata["team"])
	}
}

func TestParseInstanceData_EmptyMetadata(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-empty-meta",
		"name": "empty-meta-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"metadata": {}
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Metadata) != 0 {
		t.Errorf("expected 0 metadata entries, got %d", len(resp.Metadata))
	}
}

func TestParseInstanceData_NullMetadata(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-null-meta",
		"name": "null-meta-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"metadata": null
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Metadata != nil {
		t.Errorf("expected nil metadata, got %v", resp.Metadata)
	}
}

func TestParseInstanceData_TagsExtraction(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-tags",
		"name": "tags-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"tags": ["web", "prod", "us-east"]
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(resp.Tags))
	}
	expected := []string{"web", "prod", "us-east"}
	for i, exp := range expected {
		if resp.Tags[i] != exp {
			t.Errorf("Tags[%d]: expected %s, got %s", i, exp, resp.Tags[i])
		}
	}
}

func TestParseInstanceData_EmptyTags(t *testing.T) {
	data := json.RawMessage(`{
		"id": "inst-empty-tags",
		"name": "empty-tags-test",
		"status": "ACTIVE",
		"availability_zone": "nova",
		"tags": []
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Tags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(resp.Tags))
	}
}

// ---------------------------------------------------------------------------
// mapReadResponseToState — comprehensive field mapping
// ---------------------------------------------------------------------------

func TestMapReadResponseToState_AllFieldsPopulated(t *testing.T) {
	ctx := context.Background()

	networkList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("net-1"),
	})
	sgList, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("sg-old"),
	})

	state := &instanceResourceModel{
		NetworkIDs:       networkList,
		SecurityGroupIDs: sgList,
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringValue("script-data"),
		ServerGroupID:    types.StringValue("sgrp-1"),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-all-fields",
		Name:             "complete-server",
		Description:      "A complete server",
		FlavorID:         "flav-big",
		ImageID:          "img-ubuntu",
		SecurityGroups:   []string{"sg-new-1", "sg-new-2"},
		AvailabilityZone: "mumbai-1a",
		Key:              "my-key",
		Status:           "ACTIVE",
		ConfigDrive:      true,
		Metadata:         map[string]string{"env": "prod", "team": "infra"},
		Tags:             []string{"web", "api"},
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Core fields.
	if state.ID.ValueString() != "inst-all-fields" {
		t.Errorf("expected ID inst-all-fields, got %s", state.ID.ValueString())
	}
	if state.Name.ValueString() != "complete-server" {
		t.Errorf("expected Name complete-server, got %s", state.Name.ValueString())
	}
	if state.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", state.Status.ValueString())
	}
	if state.AvailabilityZone.ValueString() != "mumbai-1a" {
		t.Errorf("expected AZ mumbai-1a, got %s", state.AvailabilityZone.ValueString())
	}
	if state.ConfigDrive.ValueBool() != true {
		t.Error("expected ConfigDrive true")
	}

	// Description.
	if state.Description.ValueString() != "A complete server" {
		t.Errorf("expected Description 'A complete server', got %s", state.Description.ValueString())
	}

	// FlavorID.
	if state.FlavorID.ValueString() != "flav-big" {
		t.Errorf("expected FlavorID flav-big, got %s", state.FlavorID.ValueString())
	}

	// BootUUID (from image).
	if state.BootUUID.ValueString() != "img-ubuntu" {
		t.Errorf("expected BootUUID img-ubuntu, got %s", state.BootUUID.ValueString())
	}

	// KeyName.
	if state.KeyName.ValueString() != "my-key" {
		t.Errorf("expected KeyName my-key, got %s", state.KeyName.ValueString())
	}

	// Security groups — should be overwritten with new values from API.
	var sgIDs []string
	diags.Append(state.SecurityGroupIDs.ElementsAs(ctx, &sgIDs, false)...)
	if len(sgIDs) != 2 {
		t.Fatalf("expected 2 SGs, got %d", len(sgIDs))
	}
	if sgIDs[0] != "sg-new-1" {
		t.Errorf("expected sg-new-1, got %s", sgIDs[0])
	}
	if sgIDs[1] != "sg-new-2" {
		t.Errorf("expected sg-new-2, got %s", sgIDs[1])
	}

	// Metadata.
	metaMap := make(map[string]string)
	diags.Append(state.Metadata.ElementsAs(ctx, &metaMap, false)...)
	if len(metaMap) != 2 {
		t.Fatalf("expected 2 metadata entries, got %d", len(metaMap))
	}
	if metaMap["env"] != "prod" {
		t.Errorf("expected metadata env=prod, got %s", metaMap["env"])
	}

	// Tags.
	var tags []string
	diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "web" {
		t.Errorf("expected first tag web, got %s", tags[0])
	}

	// Preserved fields — UserData and ServerGroupID should not be changed.
	if state.UserData.ValueString() != "script-data" {
		t.Errorf("expected UserData to remain 'script-data', got %s", state.UserData.ValueString())
	}
	if state.ServerGroupID.ValueString() != "sgrp-1" {
		t.Errorf("expected ServerGroupID to remain 'sgrp-1', got %s", state.ServerGroupID.ValueString())
	}
}

func TestMapReadResponseToState_DescriptionPreserveNull(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-desc-null",
		Name:             "desc-test",
		Description:      "", // API returns empty
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Description was null before, API returns empty — should stay null.
	if !state.Description.IsNull() {
		t.Errorf("expected Description to remain null when API returns empty, got %s", state.Description.ValueString())
	}
}

func TestMapReadResponseToState_DescriptionKeepExisting(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringValue("existing desc"),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-desc-keep",
		Name:             "desc-keep-test",
		Description:      "", // API returns empty
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Description was non-null, API returns empty — neither branch runs, existing preserved.
	if state.Description.ValueString() != "existing desc" {
		t.Errorf("expected Description 'existing desc', got %s", state.Description.ValueString())
	}
}

func TestMapReadResponseToState_BootFromVolume_NoImage(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
		BootUUID:         types.StringValue("vol-uuid-boot"),
	}

	instance := &readInstanceResponse{
		ID:               "inst-bfv",
		Name:             "boot-from-vol",
		FlavorID:         "flv-1",
		ImageID:          "", // Boot from volume — no image
		AvailabilityZone: "nova",
		Status:           "ACTIVE",
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// BootUUID should remain as the original volume UUID since ImageID is empty.
	if state.BootUUID.ValueString() != "vol-uuid-boot" {
		t.Errorf("expected BootUUID to remain 'vol-uuid-boot' for BFV, got %s", state.BootUUID.ValueString())
	}
}

func TestMapReadResponseToState_EmptyMetadataPreservesNull(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-empty-meta",
		Name:             "empty-meta",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
		Metadata:         map[string]string{}, // Empty map
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Empty metadata (len 0) should keep null.
	if !state.Metadata.IsNull() {
		t.Error("expected Metadata to remain null for empty API metadata map")
	}
}

func TestMapReadResponseToState_NilMetadataPreservesNull(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-nil-meta",
		Name:             "nil-meta",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
		Metadata:         nil,
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if !state.Metadata.IsNull() {
		t.Error("expected Metadata to remain null for nil API metadata")
	}
}

func TestMapReadResponseToState_EmptyTagsPreservesNull(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-empty-tags",
		Name:             "empty-tags",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
		Tags:             []string{},
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Empty tags should keep null.
	if !state.Tags.IsNull() {
		t.Error("expected Tags to remain null for empty tags array")
	}
}

func TestMapReadResponseToState_NilTagsPreservesNull(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-nil-tags",
		Name:             "nil-tags",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
		Tags:             nil,
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if !state.Tags.IsNull() {
		t.Error("expected Tags to remain null for nil tags")
	}
}

func TestMapReadResponseToState_EmptySGsPreserveExisting(t *testing.T) {
	ctx := context.Background()

	existingSGs, _ := types.ListValue(types.StringType, []attr.Value{
		types.StringValue("sg-existing"),
	})

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: existingSGs,
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-empty-sgs",
		Name:             "empty-sgs",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
		SecurityGroups:   []string{}, // Empty SGs from API
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Empty SGs should keep existing state value.
	var sgIDs []string
	diags.Append(state.SecurityGroupIDs.ElementsAs(ctx, &sgIDs, false)...)
	if len(sgIDs) != 1 {
		t.Fatalf("expected 1 existing SG preserved, got %d", len(sgIDs))
	}
	if sgIDs[0] != "sg-existing" {
		t.Errorf("expected sg-existing, got %s", sgIDs[0])
	}
}

func TestMapReadResponseToState_KeyNamePreserveNull(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
	}

	instance := &readInstanceResponse{
		ID:               "inst-no-key",
		Name:             "no-key-server",
		FlavorID:         "flv-1",
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
		Key:              "", // No key set
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if !state.KeyName.IsNull() {
		t.Errorf("expected KeyName to remain null when API returns empty, got %s", state.KeyName.ValueString())
	}
}

func TestMapReadResponseToState_FlavorIDPreserved(t *testing.T) {
	ctx := context.Background()

	state := &instanceResourceModel{
		NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
		SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
		Description:      types.StringNull(),
		KeyName:          types.StringNull(),
		UserData:         types.StringNull(),
		ServerGroupID:    types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
		Tags:             types.ListNull(types.StringType),
		FlavorID:         types.StringValue("old-flavor"),
	}

	instance := &readInstanceResponse{
		ID:               "inst-flavor",
		Name:             "flavor-test",
		FlavorID:         "", // Empty from API
		ImageID:          "img-1",
		AvailabilityZone: "az-1",
		Status:           "ACTIVE",
	}

	diags := mapReadResponseToState(ctx, instance, state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Empty FlavorID from API — the if branch doesn't run, so existing value is preserved.
	if state.FlavorID.ValueString() != "old-flavor" {
		t.Errorf("expected FlavorID to remain 'old-flavor', got %s", state.FlavorID.ValueString())
	}
}

// ---------------------------------------------------------------------------
// parseInstanceData — complete real-world response
// ---------------------------------------------------------------------------

func TestParseInstanceData_RealWorldResponse(t *testing.T) {
	// Simulates a realistic full API response with all fields.
	data := json.RawMessage(`{
		"id": "550e8400-e29b-41d4-a716-446655440000",
		"name": "production-web-01",
		"description": "Primary web server",
		"status": "ACTIVE",
		"availability_zone": "mumbai-1a",
		"key": "deploy-prod",
		"metadata": {
			"env": "production",
			"team": "platform",
			"cost_center": "eng-001",
			"terraform": "true"
		},
		"tags": ["web", "production", "critical"],
		"flavor": {
			"id": "64b9c2fc-1234-5678-abcd-111111111111",
			"name": "m1.xlarge",
			"ram": 16384,
			"vcpus": 4,
			"disk": 0
		},
		"image": {
			"id": "aabbccdd-1234-5678-abcd-222222222222",
			"name": "Ubuntu-22.04-LTS"
		},
		"config_drive": "True",
		"security_groups": [
			{"id": "sg-11111111", "name": "default"},
			{"id": "sg-22222222", "name": "web-access"},
			{"id": "sg-33333333", "name": "monitoring"}
		],
		"volumes_attached": [
			{"id": "vol-aaaa", "device": "/dev/vda"},
			{"id": "vol-bbbb", "device": "/dev/vdb"}
		],
		"addresses": {
			"private": [{"addr": "10.0.0.5", "version": 4}],
			"public": [{"addr": "203.0.113.10", "version": 4}]
		}
	}`)

	resp, err := parseInstanceData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected UUID ID, got %s", resp.ID)
	}
	if resp.Name != "production-web-01" {
		t.Errorf("expected name production-web-01, got %s", resp.Name)
	}
	if resp.Description != "Primary web server" {
		t.Errorf("expected description 'Primary web server', got %s", resp.Description)
	}
	if resp.FlavorID != "64b9c2fc-1234-5678-abcd-111111111111" {
		t.Errorf("expected flavor UUID, got %s", resp.FlavorID)
	}
	if resp.ImageID != "aabbccdd-1234-5678-abcd-222222222222" {
		t.Errorf("expected image UUID, got %s", resp.ImageID)
	}
	if resp.ConfigDrive != true {
		t.Error("expected ConfigDrive true")
	}
	if len(resp.SecurityGroups) != 3 {
		t.Fatalf("expected 3 SGs, got %d", len(resp.SecurityGroups))
	}
	if len(resp.Metadata) != 4 {
		t.Fatalf("expected 4 metadata entries, got %d", len(resp.Metadata))
	}
	if resp.Metadata["terraform"] != "true" {
		t.Errorf("expected metadata terraform=true, got %s", resp.Metadata["terraform"])
	}
	if len(resp.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(resp.Tags))
	}
}

// ---------------------------------------------------------------------------
// mapReadResponseToState — status variations
// ---------------------------------------------------------------------------

func TestMapReadResponseToState_StatusVariants(t *testing.T) {
	statuses := []string{"ACTIVE", "BUILD", "SHUTOFF", "ERROR", "SUSPENDED", "PAUSED", "VERIFY_RESIZE", "REBOOT"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			state := &instanceResourceModel{
				NetworkIDs:       types.ListValueMust(types.StringType, []attr.Value{}),
				SecurityGroupIDs: types.ListValueMust(types.StringType, []attr.Value{}),
				Description:      types.StringNull(),
				KeyName:          types.StringNull(),
				UserData:         types.StringNull(),
				ServerGroupID:    types.StringNull(),
				Metadata:         types.MapNull(types.StringType),
				Tags:             types.ListNull(types.StringType),
			}

			instance := &readInstanceResponse{
				ID:               "inst-status",
				Name:             "status-test",
				FlavorID:         "flv-1",
				ImageID:          "img-1",
				AvailabilityZone: "az-1",
				Status:           status,
			}

			diags := mapReadResponseToState(ctx, instance, state)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}

			if state.Status.ValueString() != status {
				t.Errorf("expected Status %s, got %s", status, state.Status.ValueString())
			}
		})
	}
}
