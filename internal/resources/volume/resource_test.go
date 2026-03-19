package volume

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildAPIRequest_AllFields(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	metadata, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"env": "prod",
	})

	plan := &volumeResourceModel{
		Name:             types.StringValue("data-vol"),
		Size:             types.Int64Value(100),
		VolumeType:       types.StringValue("ssd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringValue("mumbai-1a"),
		Description:      types.StringValue("Data volume"),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         metadata,
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.Name != "data-vol" {
		t.Errorf("expected name data-vol, got %s", body.Name)
	}
	if body.Size != 100 {
		t.Errorf("expected size 100, got %d", body.Size)
	}
	// buildAPIRequest maps "ssd" → backend name for the API.
	if body.VolumeType != "NVMe based High IOPS Storage" {
		t.Errorf("expected volume_type 'NVMe based High IOPS Storage', got %s", body.VolumeType)
	}
	if body.BillingType != "hourly" {
		t.Errorf("expected billing_type hourly, got %s", body.BillingType)
	}
	if body.AvailabilityZone != "mumbai-1a" {
		t.Errorf("expected availability_zone mumbai-1a, got %s", body.AvailabilityZone)
	}
	if body.Description != "Data volume" {
		t.Errorf("expected description Data volume, got %s", body.Description)
	}
	if body.Metadata["env"] != "prod" {
		t.Errorf("expected metadata env=prod, got %v", body.Metadata)
	}
}

func TestBuildAPIRequest_MinimalFields(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("basic-vol"),
		Size:             types.Int64Value(10),
		VolumeType:       types.StringValue("hdd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringNull(),
		Description:      types.StringNull(),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.BillingType != "hourly" {
		t.Errorf("expected billing_type hourly, got %s", body.BillingType)
	}
	if body.AvailabilityZone != "" {
		t.Errorf("expected empty availability_zone, got %s", body.AvailabilityZone)
	}
	if body.Description != "" {
		t.Errorf("expected empty description, got %s", body.Description)
	}
	if body.Metadata != nil {
		t.Errorf("expected nil metadata, got %v", body.Metadata)
	}
}

func TestBuildAPIRequest_WithSourceVolID(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("clone-vol"),
		Size:             types.Int64Value(50),
		VolumeType:       types.StringValue("ssd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringNull(),
		Description:      types.StringNull(),
		SourceVolID:      types.StringValue("550e8400-e29b-41d4-a716-446655440000"),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.SourceVolID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected source_volid '550e8400-e29b-41d4-a716-446655440000', got %s", body.SourceVolID)
	}
	// Other source fields should remain empty.
	if body.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id, got %s", body.SnapshotID)
	}
	if body.BackupID != "" {
		t.Errorf("expected empty backup_id, got %s", body.BackupID)
	}
	if body.ImageRef != "" {
		t.Errorf("expected empty image_ref, got %s", body.ImageRef)
	}
}

func TestBuildAPIRequest_WithSnapshotID(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("snap-vol"),
		Size:             types.Int64Value(100),
		VolumeType:       types.StringValue("ssd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringNull(),
		Description:      types.StringNull(),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringValue("a1b2c3d4-e5f6-7890-abcd-ef1234567890"),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.SnapshotID != "a1b2c3d4-e5f6-7890-abcd-ef1234567890" {
		t.Errorf("expected snapshot_id 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', got %s", body.SnapshotID)
	}
	if body.SourceVolID != "" {
		t.Errorf("expected empty source_volid, got %s", body.SourceVolID)
	}
}

func TestBuildAPIRequest_WithBackupID(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("restore-vol"),
		Size:             types.Int64Value(200),
		VolumeType:       types.StringValue("hdd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringNull(),
		Description:      types.StringNull(),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringValue("deadbeef-1234-5678-9abc-def012345678"),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.BackupID != "deadbeef-1234-5678-9abc-def012345678" {
		t.Errorf("expected backup_id 'deadbeef-1234-5678-9abc-def012345678', got %s", body.BackupID)
	}
	if body.SourceVolID != "" {
		t.Errorf("expected empty source_volid, got %s", body.SourceVolID)
	}
	if body.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id, got %s", body.SnapshotID)
	}
}

func TestBuildAPIRequest_WithImageRef(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("boot-vol"),
		Size:             types.Int64Value(40),
		VolumeType:       types.StringValue("ssd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringValue("mumbai-1a"),
		Description:      types.StringValue("Bootable volume"),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringValue("12345678-abcd-ef01-2345-6789abcdef01"),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.ImageRef != "12345678-abcd-ef01-2345-6789abcdef01" {
		t.Errorf("expected image_ref '12345678-abcd-ef01-2345-6789abcdef01', got %s", body.ImageRef)
	}
	if body.SourceVolID != "" {
		t.Errorf("expected empty source_volid, got %s", body.SourceVolID)
	}
	if body.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id, got %s", body.SnapshotID)
	}
	if body.BackupID != "" {
		t.Errorf("expected empty backup_id, got %s", body.BackupID)
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-123",
		Name:       "data-vol",
		Size:       200,
		VolumeType: "ssd",
		Status:     "available",
		Metadata:   map[string]string{"env": "prod"},
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if model.ID.ValueString() != "vol-123" {
		t.Errorf("expected ID vol-123, got %s", model.ID.ValueString())
	}
	if model.Size.ValueInt64() != 200 {
		t.Errorf("expected size 200, got %d", model.Size.ValueInt64())
	}
	if model.Status.ValueString() != "available" {
		t.Errorf("expected status available, got %s", model.Status.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-456",
		Name:       "empty-vol",
		Size:       50,
		VolumeType: "hdd",
		Status:     "creating",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if !model.Description.IsNull() {
		t.Error("expected Description to remain null")
	}
	if !model.SourceVolID.IsNull() {
		t.Error("expected SourceVolID to remain null")
	}
	if !model.Metadata.IsNull() {
		t.Error("expected Metadata to remain null")
	}
}

func TestMapAPIResponseToState_PreservesVolumeType(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Model has the user-friendly "ssd" value.
	model := &volumeResourceModel{
		VolumeType:  types.StringValue("ssd"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	// API returns the backend name.
	apiResp := &volumeAPIResponse{
		ID:         "vol-789",
		Name:       "ssd-vol",
		Size:       100,
		VolumeType: "NVMe based High IOPS Storage",
		Status:     "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	// The model should preserve the user's "ssd" value to avoid unnecessary diffs.
	if model.VolumeType.ValueString() != "ssd" {
		t.Errorf("expected VolumeType to remain 'ssd', got %s", model.VolumeType.ValueString())
	}
}

func TestMapAPIResponseToState_WithMetadata(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-meta",
		Name:       "meta-vol",
		Size:       100,
		VolumeType: "HDD based Storage",
		Status:     "available",
		Metadata: map[string]string{
			"env":     "production",
			"team":    "platform",
			"project": "infra",
		},
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if model.Metadata.IsNull() {
		t.Fatal("expected Metadata to be non-null")
	}
	elements := model.Metadata.Elements()
	if len(elements) != 3 {
		t.Fatalf("expected 3 metadata entries, got %d", len(elements))
	}

	// Extract values for verification.
	metadata := make(map[string]string)
	diags.Append(model.Metadata.ElementsAs(ctx, &metadata, false)...)
	if diags.HasError() {
		t.Fatalf("failed to extract metadata: %v", diags.Errors())
	}
	if metadata["env"] != "production" {
		t.Errorf("expected metadata env=production, got %s", metadata["env"])
	}
	if metadata["team"] != "platform" {
		t.Errorf("expected metadata team=platform, got %s", metadata["team"])
	}
	if metadata["project"] != "infra" {
		t.Errorf("expected metadata project=infra, got %s", metadata["project"])
	}
}

func TestMapAPIResponseToState_WithAllSourceFields(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringValue(""),
		SourceVolID: types.StringValue(""),
		SnapshotID:  types.StringValue(""),
		BackupID:    types.StringValue(""),
		ImageRef:    types.StringValue(""),
		Metadata:    types.MapValueMust(types.StringType, map[string]attr.Value{}),
	}

	apiResp := &volumeAPIResponse{
		ID:               "vol-all-sources",
		Name:             "full-source-vol",
		Size:             200,
		VolumeType:       "NVMe based High IOPS Storage",
		BillingType:      "hourly",
		AvailabilityZone: "mumbai-1a",
		Description:      "Volume with all source fields",
		SourceVolID:      "11111111-1111-1111-1111-111111111111",
		SnapshotID:       "22222222-2222-2222-2222-222222222222",
		BackupID:         "33333333-3333-3333-3333-333333333333",
		ImageRef:         "44444444-4444-4444-4444-444444444444",
		Status:           "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if model.SourceVolID.ValueString() != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected SourceVolID '11111111-1111-1111-1111-111111111111', got %s", model.SourceVolID.ValueString())
	}
	if model.SnapshotID.ValueString() != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("expected SnapshotID '22222222-2222-2222-2222-222222222222', got %s", model.SnapshotID.ValueString())
	}
	if model.BackupID.ValueString() != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("expected BackupID '33333333-3333-3333-3333-333333333333', got %s", model.BackupID.ValueString())
	}
	if model.ImageRef.ValueString() != "44444444-4444-4444-4444-444444444444" {
		t.Errorf("expected ImageRef '44444444-4444-4444-4444-444444444444', got %s", model.ImageRef.ValueString())
	}
	if model.Description.ValueString() != "Volume with all source fields" {
		t.Errorf("expected Description 'Volume with all source fields', got %s", model.Description.ValueString())
	}
	if model.AvailabilityZone.ValueString() != "mumbai-1a" {
		t.Errorf("expected AvailabilityZone 'mumbai-1a', got %s", model.AvailabilityZone.ValueString())
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

func TestVolumeTypeMapping(t *testing.T) {
	// Forward mapping: user-friendly → backend
	tests := []struct {
		input    string
		expected string
	}{
		{"ssd", "NVMe based High IOPS Storage"},
		{"nvme", "NVMe based High IOPS Storage"},
		{"hdd", "HDD based Storage"},
		{"NVMe based High IOPS Storage", "NVMe based High IOPS Storage"}, // exact backend name passthrough
		{"HDD based Storage", "HDD based Storage"},                       // exact backend name passthrough
	}
	for _, tc := range tests {
		got := volumeTypeToBackend(tc.input)
		if got != tc.expected {
			t.Errorf("volumeTypeToBackend(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}

	// Reverse mapping: backend → user-friendly
	reverseTests := []struct {
		input    string
		expected string
	}{
		{"NVMe based High IOPS Storage", "ssd"},
		{"HDD based Storage", "hdd"},
		{"unknown-type", "unknown-type"}, // passthrough
	}
	for _, tc := range reverseTests {
		got := volumeTypeFromBackend(tc.input)
		if got != tc.expected {
			t.Errorf("volumeTypeFromBackend(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestUUIDRegex(t *testing.T) {
	re := uuidRegex()

	valid := "550e8400-e29b-41d4-a716-446655440000"
	if !re.MatchString(valid) {
		t.Errorf("expected %s to match UUID regex", valid)
	}

	invalid := "not-a-uuid"
	if re.MatchString(invalid) {
		t.Errorf("expected %s to NOT match UUID regex", invalid)
	}
}

func TestVolumeDeleteRequest_JSON(t *testing.T) {
	req := volumeDeleteRequest{
		Key:    "id",
		Values: []string{"vol-1"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal volumeDeleteRequest: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if parsed["key"] != "id" {
		t.Errorf("expected key 'id', got %v", parsed["key"])
	}
	values, ok := parsed["values"].([]interface{})
	if !ok {
		t.Fatalf("expected values to be an array, got %T", parsed["values"])
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0] != "vol-1" {
		t.Errorf("expected values[0] 'vol-1', got %v", values[0])
	}

	// Verify the exact JSON structure.
	expected := `{"key":"id","values":["vol-1"]}`
	if string(data) != expected {
		t.Errorf("expected JSON %s, got %s", expected, string(data))
	}
}

func TestBuildAPIRequest_VolumeTypeMapping(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("ssd-vol"),
		Size:             types.Int64Value(50),
		VolumeType:       types.StringValue("ssd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringNull(),
		Description:      types.StringNull(),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.VolumeType != "NVMe based High IOPS Storage" {
		t.Errorf("expected volume_type 'NVMe based High IOPS Storage', got %s", body.VolumeType)
	}
	if body.BillingType != "hourly" {
		t.Errorf("expected billing_type hourly, got %s", body.BillingType)
	}
}

func TestBuildAPIRequest_HddType(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("hdd-vol"),
		Size:             types.Int64Value(100),
		VolumeType:       types.StringValue("hdd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringNull(),
		Description:      types.StringNull(),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if body.VolumeType != "HDD based Storage" {
		t.Errorf("expected volume_type 'HDD based Storage', got %s", body.VolumeType)
	}
	if body.BillingType != "hourly" {
		t.Errorf("expected billing_type hourly, got %s", body.BillingType)
	}
}

func TestMapAPIResponseToState_VolumeStatus(t *testing.T) {
	ctx := context.Background()

	statuses := []string{"available", "in-use", "creating", "deleting", "error", "backing-up"}
	for _, status := range statuses {
		var diags diag.Diagnostics

		model := &volumeResourceModel{
			BillingType: types.StringValue("hourly"),
			Description: types.StringNull(),
			SourceVolID: types.StringNull(),
			SnapshotID:  types.StringNull(),
			BackupID:    types.StringNull(),
			ImageRef:    types.StringNull(),
			Metadata:    types.MapNull(types.StringType),
		}

		apiResp := &volumeAPIResponse{
			ID:         "vol-status-test",
			Name:       "status-vol",
			Size:       50,
			VolumeType: "ssd",
			Status:     status,
		}

		mapAPIResponseToState(model, apiResp, ctx, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected diagnostics for status %q: %v", status, diags.Errors())
		}
		if model.Status.ValueString() != status {
			t.Errorf("expected status %q, got %q", status, model.Status.ValueString())
		}
	}
}

func TestVolumeTypeMapping_UnknownType(t *testing.T) {
	// Forward: unknown type passes through unchanged
	unknown := "SomeCustomStorageType"
	got := volumeTypeToBackend(unknown)
	if got != unknown {
		t.Errorf("volumeTypeToBackend(%q) = %q, want %q (passthrough)", unknown, got, unknown)
	}

	// Reverse: unknown backend type passes through unchanged
	got2 := volumeTypeFromBackend(unknown)
	if got2 != unknown {
		t.Errorf("volumeTypeFromBackend(%q) = %q, want %q (passthrough)", unknown, got2, unknown)
	}

	// Empty string passthrough
	empty := ""
	if volumeTypeToBackend(empty) != empty {
		t.Error("expected empty string passthrough for volumeTypeToBackend")
	}
	if volumeTypeFromBackend(empty) != empty {
		t.Error("expected empty string passthrough for volumeTypeFromBackend")
	}
}

func TestUUIDRegex_AdditionalCases(t *testing.T) {
	re := uuidRegex()

	tests := []struct {
		name    string
		input   string
		matches bool
	}{
		// Valid cases
		{"lowercase valid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"uppercase hex", "550E8400-E29B-41D4-A716-446655440000", true},
		{"mixed case", "550e8400-E29B-41d4-A716-446655440000", true},
		{"all zeros", "00000000-0000-0000-0000-000000000000", true},
		{"all f's", "ffffffff-ffff-ffff-ffff-ffffffffffff", true},

		// Invalid cases
		{"too short", "550e8400-e29b-41d4-a716", false},
		{"missing segment", "550e8400-e29b-41d4-446655440000", false},
		{"extra characters at end", "550e8400-e29b-41d4-a716-446655440000x", false},
		{"extra characters at start", "x550e8400-e29b-41d4-a716-446655440000", false},
		{"no dashes", "550e8400e29b41d4a716446655440000", false},
		{"empty string", "", false},
		{"spaces", "550e8400 e29b 41d4 a716 446655440000", false},
		{"invalid hex char g", "g50e8400-e29b-41d4-a716-446655440000", false},
		{"short first segment", "550e840-e29b-41d4-a716-446655440000", false},
		{"long first segment", "550e84000-e29b-41d4-a716-446655440000", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := re.MatchString(tc.input)
			if got != tc.matches {
				t.Errorf("uuidRegex().MatchString(%q) = %v, want %v", tc.input, got, tc.matches)
			}
		})
	}
}

// ===========================================================================
// Additional comprehensive tests
// ===========================================================================

// ---------------------------------------------------------------------------
// mapAPIResponseToState — comprehensive field mapping
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_AllFieldsPopulated(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		VolumeType:  types.StringValue("ssd"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:               "vol-all-fields",
		Name:             "complete-vol",
		Size:             500,
		VolumeType:       "NVMe based High IOPS Storage",
		BillingType:      "hourly",
		AvailabilityZone: "mumbai-1a",
		Description:      "A complete volume",
		SourceVolID:      "11111111-1111-1111-1111-111111111111",
		SnapshotID:       "22222222-2222-2222-2222-222222222222",
		BackupID:         "33333333-3333-3333-3333-333333333333",
		ImageRef:         "44444444-4444-4444-4444-444444444444",
		Metadata:         map[string]string{"env": "prod", "team": "infra"},
		Status:           "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if model.ID.ValueString() != "vol-all-fields" {
		t.Errorf("expected ID vol-all-fields, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "complete-vol" {
		t.Errorf("expected Name complete-vol, got %s", model.Name.ValueString())
	}
	if model.Size.ValueInt64() != 500 {
		t.Errorf("expected Size 500, got %d", model.Size.ValueInt64())
	}
	// Model had "ssd" and backend returns the matching backend name — should preserve "ssd".
	if model.VolumeType.ValueString() != "ssd" {
		t.Errorf("expected VolumeType to remain 'ssd', got %s", model.VolumeType.ValueString())
	}
	if model.BillingType.ValueString() != "hourly" {
		t.Errorf("expected BillingType 'hourly', got %s", model.BillingType.ValueString())
	}
	if model.AvailabilityZone.ValueString() != "mumbai-1a" {
		t.Errorf("expected AvailabilityZone mumbai-1a, got %s", model.AvailabilityZone.ValueString())
	}
	if model.Description.ValueString() != "A complete volume" {
		t.Errorf("expected Description 'A complete volume', got %s", model.Description.ValueString())
	}
	if model.Status.ValueString() != "available" {
		t.Errorf("expected Status available, got %s", model.Status.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Volume type mapping — comprehensive tests
// ---------------------------------------------------------------------------

func TestVolumeTypeMapping_SSD2NotMapped(t *testing.T) {
	// "ssd2" is not in the mapping — should pass through unchanged.
	got := volumeTypeToBackend("ssd2")
	if got != "ssd2" {
		t.Errorf("volumeTypeToBackend(\"ssd2\") = %q, expected passthrough \"ssd2\"", got)
	}
}

func TestVolumeTypeMapping_NVMeAlias(t *testing.T) {
	// "nvme" is an alias for "ssd" — both map to the same backend name.
	ssdBackend := volumeTypeToBackend("ssd")
	nvmeBackend := volumeTypeToBackend("nvme")

	if ssdBackend != nvmeBackend {
		t.Errorf("expected ssd (%s) and nvme (%s) to map to same backend type", ssdBackend, nvmeBackend)
	}
	if ssdBackend != "NVMe based High IOPS Storage" {
		t.Errorf("expected 'NVMe based High IOPS Storage', got %s", ssdBackend)
	}
}

func TestVolumeTypeFromBackend_Preserves_UserValue(t *testing.T) {
	// When user writes "ssd" and backend returns "NVMe based High IOPS Storage",
	// the mapAPIResponseToState logic should keep "ssd" to avoid unnecessary diffs.
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		VolumeType:  types.StringValue("nvme"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-preserve",
		Name:       "preserve-vol",
		Size:       100,
		VolumeType: "NVMe based High IOPS Storage",
		Status:     "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// "nvme" maps to "NVMe based High IOPS Storage", so should be preserved.
	if model.VolumeType.ValueString() != "nvme" {
		t.Errorf("expected VolumeType to remain 'nvme', got %s", model.VolumeType.ValueString())
	}
}

func TestVolumeTypeMapping_HDDPreserved(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		VolumeType:  types.StringValue("hdd"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-hdd",
		Name:       "hdd-vol",
		Size:       100,
		VolumeType: "HDD based Storage",
		Status:     "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if model.VolumeType.ValueString() != "hdd" {
		t.Errorf("expected VolumeType to remain 'hdd', got %s", model.VolumeType.ValueString())
	}
}

func TestVolumeTypeMapping_BackendNameDirect(t *testing.T) {
	// If user writes the exact backend name, it should be preserved.
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		VolumeType:  types.StringValue("NVMe based High IOPS Storage"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-exact",
		Name:       "exact-vol",
		Size:       100,
		VolumeType: "NVMe based High IOPS Storage",
		Status:     "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	if model.VolumeType.ValueString() != "NVMe based High IOPS Storage" {
		t.Errorf("expected VolumeType to remain exact backend name, got %s", model.VolumeType.ValueString())
	}
}

func TestVolumeTypeMapping_EmptyModel(t *testing.T) {
	// If model has empty VolumeType, it should be set from backend.
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		VolumeType:  types.StringValue(""),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-empty-type",
		Name:       "empty-type-vol",
		Size:       50,
		VolumeType: "NVMe based High IOPS Storage",
		Status:     "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Empty string maps to "" via volumeTypeToBackend, which does NOT match
	// "NVMe based High IOPS Storage", so the else branch runs: volumeTypeFromBackend.
	if model.VolumeType.ValueString() != "ssd" {
		t.Errorf("expected VolumeType to be 'ssd' (from backend mapping), got %s", model.VolumeType.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Size preservation during async resize
// ---------------------------------------------------------------------------

func TestSizePreservation_AsyncResize_LargerStateKept(t *testing.T) {
	// Simulates the Read flow: state has size=15 (user requested resize),
	// API returns size=10 (resize not yet complete). The larger value should be preserved.
	ctx := context.Background()
	var diags diag.Diagnostics

	// Before mapping, state has the expected larger size.
	state := &volumeResourceModel{
		ID:          types.StringValue("vol-resize"),
		Name:        types.StringValue("resize-vol"),
		Size:        types.Int64Value(15), // User requested expansion to 15
		VolumeType:  types.StringValue("ssd"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	// API returns old size (resize still in progress).
	apiResp := &volumeAPIResponse{
		ID:         "vol-resize",
		Name:       "resize-vol",
		Size:       10,
		VolumeType: "NVMe based High IOPS Storage",
		Status:     "available",
	}

	expectedSize := state.Size.ValueInt64()

	mapAPIResponseToState(state, apiResp, ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// After mapping, API sets size to 10.
	if state.Size.ValueInt64() != 10 {
		t.Fatalf("expected mapAPIResponseToState to set size to 10, got %d", state.Size.ValueInt64())
	}

	// The Read function applies the preservation logic:
	if expectedSize > state.Size.ValueInt64() {
		state.Size = types.Int64Value(expectedSize)
	}

	if state.Size.ValueInt64() != 15 {
		t.Errorf("expected preserved size 15 (async resize), got %d", state.Size.ValueInt64())
	}
}

func TestSizePreservation_AsyncResize_SameSizeNoChange(t *testing.T) {
	// When state size == API size, no change needed.
	ctx := context.Background()
	var diags diag.Diagnostics

	state := &volumeResourceModel{
		Size:        types.Int64Value(50),
		VolumeType:  types.StringValue("ssd"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-same",
		Name:       "same-vol",
		Size:       50,
		VolumeType: "NVMe based High IOPS Storage",
		Status:     "available",
	}

	expectedSize := state.Size.ValueInt64()

	mapAPIResponseToState(state, apiResp, ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Apply same logic as Read.
	if expectedSize > state.Size.ValueInt64() {
		state.Size = types.Int64Value(expectedSize)
	}

	if state.Size.ValueInt64() != 50 {
		t.Errorf("expected size 50 (no change), got %d", state.Size.ValueInt64())
	}
}

func TestSizePreservation_AsyncResize_APIReturnsBigger(t *testing.T) {
	// If API returns a bigger size than state (shouldn't happen normally, but test the logic).
	ctx := context.Background()
	var diags diag.Diagnostics

	state := &volumeResourceModel{
		Size:        types.Int64Value(10),
		VolumeType:  types.StringValue("ssd"),
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-bigger",
		Name:       "bigger-vol",
		Size:       20,
		VolumeType: "NVMe based High IOPS Storage",
		Status:     "available",
	}

	expectedSize := state.Size.ValueInt64() // 10

	mapAPIResponseToState(state, apiResp, ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Apply same logic as Read — expectedSize (10) is NOT > state.Size (20), so no override.
	if expectedSize > state.Size.ValueInt64() {
		state.Size = types.Int64Value(expectedSize)
	}

	// API size should be used since it's bigger.
	if state.Size.ValueInt64() != 20 {
		t.Errorf("expected size 20 (API value), got %d", state.Size.ValueInt64())
	}
}

// ---------------------------------------------------------------------------
// buildAPIRequest — update-specific field verification
// ---------------------------------------------------------------------------

func TestBuildAPIRequest_UpdateBody_Fields(t *testing.T) {
	// The buildAPIRequest is used for both create and update. Verify it
	// generates the correct JSON with all fields. The Update handler
	// sends this body to PUT which only processes name, description, metadata.
	ctx := context.Background()
	var diags diag.Diagnostics

	metadata, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"updated": "true",
	})

	plan := &volumeResourceModel{
		Name:             types.StringValue("updated-vol"),
		Size:             types.Int64Value(100),
		VolumeType:       types.StringValue("ssd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringValue("mumbai-1a"),
		Description:      types.StringValue("Updated description"),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         metadata,
	}

	body := buildAPIRequest(plan, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	// Verify name and description are set.
	if body.Name != "updated-vol" {
		t.Errorf("expected name updated-vol, got %s", body.Name)
	}
	if body.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %s", body.Description)
	}
	if body.Metadata["updated"] != "true" {
		t.Errorf("expected metadata updated=true, got %v", body.Metadata)
	}

	// Verify JSON shape — size is always present (used in create, ignored in update PUT).
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// These keys should always be present in the serialized body.
	for _, key := range []string{"name", "size", "volume_type", "billing_type"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q to be present", key)
		}
	}
}

func TestBuildAPIRequest_NullOptionalSourceFields(t *testing.T) {
	// When all optional source fields are null, they should not appear in JSON.
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &volumeResourceModel{
		Name:             types.StringValue("no-source-vol"),
		Size:             types.Int64Value(10),
		VolumeType:       types.StringValue("ssd"),
		BillingType:      types.StringValue("hourly"),
		AvailabilityZone: types.StringNull(),
		Description:      types.StringNull(),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	body := buildAPIRequest(plan, ctx, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Optional fields should be omitted via omitempty.
	omittedKeys := []string{"description", "source_volid", "snapshot_id", "backup_id",
		"image_ref", "metadata", "availability_zone"}
	for _, key := range omittedKeys {
		if _, ok := raw[key]; ok {
			t.Errorf("expected %q to be omitted from JSON when null/empty", key)
		}
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState — description edge cases
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_DescriptionPreserveNull(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:          "vol-desc-null",
		Name:        "desc-null-vol",
		Size:        50,
		VolumeType:  "ssd",
		Description: "",
		Status:      "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if !model.Description.IsNull() {
		t.Errorf("expected Description to remain null, got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionOverwrite(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringValue("old desc"),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:          "vol-desc-overwrite",
		Name:        "desc-vol",
		Size:        50,
		VolumeType:  "ssd",
		Description: "new desc",
		Status:      "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if model.Description.ValueString() != "new desc" {
		t.Errorf("expected Description 'new desc', got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionKeepExistingWhenAPIEmpty(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringValue("existing desc"),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:          "vol-desc-keep",
		Name:        "keep-vol",
		Size:        50,
		VolumeType:  "ssd",
		Description: "",
		Status:      "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	// Description is not null, but API returns empty, so neither branch runs.
	if model.Description.ValueString() != "existing desc" {
		t.Errorf("expected Description to remain 'existing desc', got %s", model.Description.ValueString())
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState — availability zone
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_AvailabilityZoneEmpty(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		AvailabilityZone: types.StringValue("mumbai-1a"),
		BillingType:      types.StringValue("hourly"),
		Description:      types.StringNull(),
		SourceVolID:      types.StringNull(),
		SnapshotID:       types.StringNull(),
		BackupID:         types.StringNull(),
		ImageRef:         types.StringNull(),
		Metadata:         types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:               "vol-az",
		Name:             "az-vol",
		Size:             50,
		VolumeType:       "ssd",
		AvailabilityZone: "",
		Status:           "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	// When API returns empty AZ, model keeps existing value.
	if model.AvailabilityZone.ValueString() != "mumbai-1a" {
		t.Errorf("expected AvailabilityZone to remain 'mumbai-1a', got %s", model.AvailabilityZone.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Volume API response JSON unmarshal
// ---------------------------------------------------------------------------

func TestVolumeAPIResponse_Unmarshal(t *testing.T) {
	raw := `{
		"id": "vol-unmarshal-123",
		"name": "unmarshal-vol",
		"size": 100,
		"volume_type": "NVMe based High IOPS Storage",
		"billing_type": "hourly",
		"availability_zone": "mumbai-1a",
		"description": "Test unmarshal",
		"source_volid": "",
		"snapshot_id": "",
		"backup_id": "",
		"image_ref": "",
		"metadata": {"key": "value"},
		"status": "available"
	}`

	var resp volumeAPIResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "vol-unmarshal-123" {
		t.Errorf("expected ID vol-unmarshal-123, got %s", resp.ID)
	}
	if resp.Size != 100 {
		t.Errorf("expected Size 100, got %d", resp.Size)
	}
	if resp.VolumeType != "NVMe based High IOPS Storage" {
		t.Errorf("expected VolumeType 'NVMe based High IOPS Storage', got %s", resp.VolumeType)
	}
	if resp.Metadata["key"] != "value" {
		t.Errorf("expected metadata key=value, got %v", resp.Metadata)
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState — metadata edge cases
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_EmptyMetadataMap(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-empty-meta",
		Name:       "empty-meta-vol",
		Size:       50,
		VolumeType: "ssd",
		Status:     "available",
		Metadata:   map[string]string{}, // empty map
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	// Empty map (len 0) should keep null.
	if !model.Metadata.IsNull() {
		t.Error("expected Metadata to remain null for empty API metadata map")
	}
}

func TestMapAPIResponseToState_MetadataOverwritesExisting(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	existingMeta, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"old_key": "old_value",
	})

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    existingMeta,
	}

	apiResp := &volumeAPIResponse{
		ID:         "vol-meta-overwrite",
		Name:       "meta-overwrite-vol",
		Size:       50,
		VolumeType: "ssd",
		Status:     "available",
		Metadata:   map[string]string{"new_key": "new_value"},
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}

	metadata := make(map[string]string)
	diags.Append(model.Metadata.ElementsAs(ctx, &metadata, false)...)
	if diags.HasError() {
		t.Fatalf("failed to extract metadata: %v", diags.Errors())
	}

	if len(metadata) != 1 {
		t.Fatalf("expected 1 metadata entry, got %d", len(metadata))
	}
	if metadata["new_key"] != "new_value" {
		t.Errorf("expected metadata new_key=new_value, got %v", metadata)
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState — BillingType preservation
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState_BillingTypePreserved(t *testing.T) {
	// When API returns empty BillingType, the model's existing BillingType is preserved.
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:          "vol-bt-preserve",
		Name:        "bt-preserve-vol",
		Size:        50,
		VolumeType:  "ssd",
		BillingType: "",
		Status:      "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if model.BillingType.ValueString() != "hourly" {
		t.Errorf("expected BillingType to remain 'hourly', got %s", model.BillingType.ValueString())
	}
}

func TestMapAPIResponseToState_BillingTypeReturnedFromAPI(t *testing.T) {
	// When API returns a non-empty BillingType, it's set on the model.
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &volumeResourceModel{
		BillingType: types.StringValue("hourly"),
		Description: types.StringNull(),
		SourceVolID: types.StringNull(),
		SnapshotID:  types.StringNull(),
		BackupID:    types.StringNull(),
		ImageRef:    types.StringNull(),
		Metadata:    types.MapNull(types.StringType),
	}

	apiResp := &volumeAPIResponse{
		ID:          "vol-bt-api",
		Name:        "bt-api-vol",
		Size:        100,
		VolumeType:  "ssd",
		BillingType: "monthly",
		Status:      "available",
	}

	mapAPIResponseToState(model, apiResp, ctx, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags.Errors())
	}
	if model.BillingType.ValueString() != "monthly" {
		t.Errorf("expected BillingType 'monthly', got %s", model.BillingType.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Billing type variants — all 5 allowed values
// ---------------------------------------------------------------------------

func TestBuildAPIRequest_BillingTypeVariants(t *testing.T) {
	billingTypes := []string{"hourly", "monthly", "quarterly", "half-yearly", "yearly"}

	for _, bt := range billingTypes {
		t.Run(bt, func(t *testing.T) {
			ctx := context.Background()
			var diags diag.Diagnostics

			plan := &volumeResourceModel{
				Name:             types.StringValue("bt-variant-vol"),
				Size:             types.Int64Value(50),
				VolumeType:       types.StringValue("ssd"),
				BillingType:      types.StringValue(bt),
				AvailabilityZone: types.StringNull(),
				Description:      types.StringNull(),
				SourceVolID:      types.StringNull(),
				SnapshotID:       types.StringNull(),
				BackupID:         types.StringNull(),
				ImageRef:         types.StringNull(),
				Metadata:         types.MapNull(types.StringType),
			}

			body := buildAPIRequest(plan, ctx, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}
			if body.BillingType != bt {
				t.Errorf("expected billing_type %q, got %q", bt, body.BillingType)
			}
		})
	}
}
