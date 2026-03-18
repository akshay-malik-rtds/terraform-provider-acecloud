package volume_backup

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- buildCreateRequest tests ---

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("test-backup"),
		VolumeID:    types.StringValue("vol-123"),
		SnapshotID:  types.StringValue("snap-456"),
		Description: types.StringValue("A test backup"),
		Incremental: types.BoolValue(true),
	}

	body := buildCreateRequest(plan)

	if body.Backup.Name != "test-backup" {
		t.Errorf("expected name test-backup, got %s", body.Backup.Name)
	}
	if body.Backup.VolumeID != "vol-123" {
		t.Errorf("expected volume_id vol-123, got %s", body.Backup.VolumeID)
	}
	if body.Backup.SnapshotID != "snap-456" {
		t.Errorf("expected snapshot_id snap-456, got %s", body.Backup.SnapshotID)
	}
	if body.Backup.Description != "A test backup" {
		t.Errorf("expected description 'A test backup', got %s", body.Backup.Description)
	}
	if body.Backup.Incremental != true {
		t.Error("expected Incremental to be true")
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("minimal-backup"),
		VolumeID:    types.StringValue("vol-789"),
		SnapshotID:  types.StringNull(),
		Description: types.StringNull(),
		Incremental: types.BoolValue(false),
	}

	body := buildCreateRequest(plan)

	if body.Backup.Name != "minimal-backup" {
		t.Errorf("expected name minimal-backup, got %s", body.Backup.Name)
	}
	if body.Backup.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id, got %s", body.Backup.SnapshotID)
	}
	if body.Backup.Description != "" {
		t.Errorf("expected empty description, got %s", body.Backup.Description)
	}
	if body.Backup.Incremental != false {
		t.Error("expected Incremental to be false")
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("json-backup"),
		VolumeID:    types.StringValue("vol-json-001"),
		SnapshotID:  types.StringValue("snap-json-002"),
		Description: types.StringValue("JSON backup desc"),
		Incremental: types.BoolValue(true),
	}

	body := buildCreateRequest(plan)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify wrapper format: {"backup": {...}}
	backupRaw, ok := raw["backup"]
	if !ok {
		t.Fatal("expected JSON to have 'backup' wrapper key")
	}

	backup, ok := backupRaw.(map[string]interface{})
	if !ok {
		t.Fatal("expected 'backup' to be an object")
	}

	if backup["name"] != "json-backup" {
		t.Errorf("expected backup.name 'json-backup', got %v", backup["name"])
	}
	if backup["volume_id"] != "vol-json-001" {
		t.Errorf("expected backup.volume_id 'vol-json-001', got %v", backup["volume_id"])
	}
	if backup["snapshot_id"] != "snap-json-002" {
		t.Errorf("expected backup.snapshot_id 'snap-json-002', got %v", backup["snapshot_id"])
	}
	if backup["description"] != "JSON backup desc" {
		t.Errorf("expected backup.description 'JSON backup desc', got %v", backup["description"])
	}
	if backup["incremental"] != true {
		t.Errorf("expected backup.incremental true, got %v", backup["incremental"])
	}
}

func TestBuildCreateRequest_OnlyRequired(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("required-only"),
		VolumeID:    types.StringValue("vol-required"),
		SnapshotID:  types.StringNull(),
		Description: types.StringNull(),
		Incremental: types.BoolValue(false),
	}

	body := buildCreateRequest(plan)

	if body.Backup.Name != "required-only" {
		t.Errorf("expected name required-only, got %s", body.Backup.Name)
	}
	if body.Backup.VolumeID != "vol-required" {
		t.Errorf("expected volume_id vol-required, got %s", body.Backup.VolumeID)
	}
	if body.Backup.Description != "" {
		t.Errorf("expected empty description, got %s", body.Backup.Description)
	}
	if body.Backup.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id, got %s", body.Backup.SnapshotID)
	}
	if body.Backup.Incremental != false {
		t.Error("expected Incremental to be false")
	}

	// Verify JSON omits optional fields
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	backup := raw["backup"].(map[string]interface{})
	if _, ok := backup["description"]; ok {
		t.Error("expected 'description' to be omitted (omitempty)")
	}
	if _, ok := backup["snapshot_id"]; ok {
		t.Error("expected 'snapshot_id' to be omitted (omitempty)")
	}
}

func TestBuildCreateRequest_UnknownDescription(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("unknown-desc"),
		VolumeID:    types.StringValue("vol-unk"),
		SnapshotID:  types.StringNull(),
		Description: types.StringUnknown(),
		Incremental: types.BoolValue(false),
	}

	body := buildCreateRequest(plan)

	if body.Backup.Description != "" {
		t.Errorf("expected empty description for unknown, got %s", body.Backup.Description)
	}
}

func TestBuildCreateRequest_UnknownSnapshotID(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("unknown-snap"),
		VolumeID:    types.StringValue("vol-unk-snap"),
		SnapshotID:  types.StringUnknown(),
		Description: types.StringNull(),
		Incremental: types.BoolValue(false),
	}

	body := buildCreateRequest(plan)

	if body.Backup.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id for unknown, got %s", body.Backup.SnapshotID)
	}
}

func TestBuildCreateRequest_IncrementalTrue(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("incr-backup"),
		VolumeID:    types.StringValue("vol-incr"),
		SnapshotID:  types.StringNull(),
		Description: types.StringNull(),
		Incremental: types.BoolValue(true),
	}

	body := buildCreateRequest(plan)

	if body.Backup.Incremental != true {
		t.Error("expected Incremental to be true")
	}

	// Verify incremental is present in JSON even when true
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	backup := raw["backup"].(map[string]interface{})
	if _, ok := backup["incremental"]; !ok {
		t.Error("expected 'incremental' field to be present in JSON")
	}
}

func TestBuildCreateRequest_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		planName    string
		volumeID    string
		snapshotID  types.String
		description types.String
		incremental bool
		wantName    string
		wantVolID   string
		wantSnapID  string
		wantDesc    string
		wantIncr    bool
	}{
		{
			name:        "all fields",
			planName:    "full-bak",
			volumeID:    "vol-full",
			snapshotID:  types.StringValue("snap-full"),
			description: types.StringValue("full desc"),
			incremental: true,
			wantName:    "full-bak",
			wantVolID:   "vol-full",
			wantSnapID:  "snap-full",
			wantDesc:    "full desc",
			wantIncr:    true,
		},
		{
			name:        "minimal",
			planName:    "min-bak",
			volumeID:    "vol-min",
			snapshotID:  types.StringNull(),
			description: types.StringNull(),
			incremental: false,
			wantName:    "min-bak",
			wantVolID:   "vol-min",
			wantSnapID:  "",
			wantDesc:    "",
			wantIncr:    false,
		},
		{
			name:        "unknown optionals",
			planName:    "unk-bak",
			volumeID:    "vol-unk",
			snapshotID:  types.StringUnknown(),
			description: types.StringUnknown(),
			incremental: false,
			wantName:    "unk-bak",
			wantVolID:   "vol-unk",
			wantSnapID:  "",
			wantDesc:    "",
			wantIncr:    false,
		},
		{
			name:        "with snapshot no desc",
			planName:    "snap-bak",
			volumeID:    "vol-snap",
			snapshotID:  types.StringValue("snap-123"),
			description: types.StringNull(),
			incremental: true,
			wantName:    "snap-bak",
			wantVolID:   "vol-snap",
			wantSnapID:  "snap-123",
			wantDesc:    "",
			wantIncr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := &volumeBackupResourceModel{
				Name:        types.StringValue(tc.planName),
				VolumeID:    types.StringValue(tc.volumeID),
				SnapshotID:  tc.snapshotID,
				Description: tc.description,
				Incremental: types.BoolValue(tc.incremental),
			}

			body := buildCreateRequest(plan)

			if body.Backup.Name != tc.wantName {
				t.Errorf("expected name %q, got %q", tc.wantName, body.Backup.Name)
			}
			if body.Backup.VolumeID != tc.wantVolID {
				t.Errorf("expected volume_id %q, got %q", tc.wantVolID, body.Backup.VolumeID)
			}
			if body.Backup.SnapshotID != tc.wantSnapID {
				t.Errorf("expected snapshot_id %q, got %q", tc.wantSnapID, body.Backup.SnapshotID)
			}
			if body.Backup.Description != tc.wantDesc {
				t.Errorf("expected description %q, got %q", tc.wantDesc, body.Backup.Description)
			}
			if body.Backup.Incremental != tc.wantIncr {
				t.Errorf("expected incremental %v, got %v", tc.wantIncr, body.Backup.Incremental)
			}
		})
	}
}

// --- mapAPIResponseToState tests ---

func TestMapAPIResponseToState(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:          "bak-abc-123",
		Name:        "prod-backup",
		VolumeID:    "vol-xyz",
		SnapshotID:  "snap-789",
		Description: "Production backup",
		Incremental: true,
		Status:      "available",
		Size:        100,
		CreatedAt:   "2024-01-01T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "bak-abc-123" {
		t.Errorf("expected ID bak-abc-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-backup" {
		t.Errorf("expected Name prod-backup, got %s", model.Name.ValueString())
	}
	if model.VolumeID.ValueString() != "vol-xyz" {
		t.Errorf("expected VolumeID vol-xyz, got %s", model.VolumeID.ValueString())
	}
	if model.SnapshotID.ValueString() != "snap-789" {
		t.Errorf("expected SnapshotID snap-789, got %s", model.SnapshotID.ValueString())
	}
	if model.Incremental.ValueBool() != true {
		t.Error("expected Incremental to be true")
	}
	if model.Status.ValueString() != "available" {
		t.Errorf("expected Status available, got %s", model.Status.ValueString())
	}
	if model.Size.ValueInt64() != 100 {
		t.Errorf("expected Size 100, got %d", model.Size.ValueInt64())
	}
}

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:          "bak-full-001",
		Name:        "full-backup",
		VolumeID:    "vol-full-xyz",
		SnapshotID:  "snap-full-789",
		Description: "Full backup with all fields",
		Incremental: false,
		Status:      "available",
		Size:        500,
		CreatedAt:   "2024-06-15T10:30:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "bak-full-001" {
		t.Errorf("expected ID bak-full-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-backup" {
		t.Errorf("expected Name full-backup, got %s", model.Name.ValueString())
	}
	if model.VolumeID.ValueString() != "vol-full-xyz" {
		t.Errorf("expected VolumeID vol-full-xyz, got %s", model.VolumeID.ValueString())
	}
	if model.SnapshotID.ValueString() != "snap-full-789" {
		t.Errorf("expected SnapshotID snap-full-789, got %s", model.SnapshotID.ValueString())
	}
	if model.Description.ValueString() != "Full backup with all fields" {
		t.Errorf("expected Description 'Full backup with all fields', got %s", model.Description.ValueString())
	}
	if model.Incremental.ValueBool() != false {
		t.Error("expected Incremental to be false")
	}
	if model.Status.ValueString() != "available" {
		t.Errorf("expected Status available, got %s", model.Status.ValueString())
	}
	if model.Size.ValueInt64() != 500 {
		t.Errorf("expected Size 500, got %d", model.Size.ValueInt64())
	}
	if model.CreatedAt.ValueString() != "2024-06-15T10:30:00Z" {
		t.Errorf("expected CreatedAt '2024-06-15T10:30:00Z', got %s", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:          "bak-123",
		Name:        "basic",
		Description: "",
		SnapshotID:  "",
		Status:      "creating",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
	if !model.SnapshotID.IsNull() {
		t.Error("expected SnapshotID to remain null when API returns empty string")
	}
}

func TestMapAPIResponseToState_AllStatuses(t *testing.T) {
	statuses := []string{"available", "creating", "deleting", "error", "restoring"}
	for _, status := range statuses {
		model := &volumeBackupResourceModel{
			Description: types.StringNull(),
			SnapshotID:  types.StringNull(),
		}

		apiResp := &backupAPIResponse{
			ID:     "bak-status-test",
			Name:   "status-backup",
			Status: status,
		}

		mapAPIResponseToState(model, apiResp)

		if model.Status.ValueString() != status {
			t.Errorf("expected status %q, got %q", status, model.Status.ValueString())
		}
	}
}

func TestMapAPIResponseToState_EmptyVolumeID(t *testing.T) {
	model := &volumeBackupResourceModel{
		VolumeID:    types.StringValue("original-vol-id"),
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:     "bak-novol",
		Name:   "no-vol-bak",
		Status: "available",
		// VolumeID intentionally empty
	}

	mapAPIResponseToState(model, apiResp)

	// VolumeID should remain at its original value since API returned empty
	if model.VolumeID.ValueString() != "original-vol-id" {
		t.Errorf("expected VolumeID to remain 'original-vol-id', got %q", model.VolumeID.ValueString())
	}
}

func TestMapAPIResponseToState_ZeroSize(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:     "bak-zero-size",
		Name:   "zero-size-bak",
		Status: "creating",
		Size:   0,
	}

	mapAPIResponseToState(model, apiResp)

	if model.Size.ValueInt64() != 0 {
		t.Errorf("expected Size 0, got %d", model.Size.ValueInt64())
	}
	if model.Size.IsNull() {
		t.Error("expected Size to be known (not null), even when 0")
	}
}

func TestMapAPIResponseToState_EmptyTimestamp(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:     "bak-no-time",
		Name:   "no-time-bak",
		Status: "creating",
		// CreatedAt intentionally empty
	}

	mapAPIResponseToState(model, apiResp)

	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %q", model.CreatedAt.ValueString())
	}
	if model.CreatedAt.IsNull() {
		t.Error("expected CreatedAt to be known (not null)")
	}
}

func TestMapAPIResponseToState_IncrementalViaIsIncremental(t *testing.T) {
	// Test that IsIncremental field is used as fallback
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:            "bak-is-incr",
		Name:          "is-incr-bak",
		Status:        "available",
		Incremental:   false,
		IsIncremental: true,
	}

	mapAPIResponseToState(model, apiResp)

	// Should be true because IsIncremental is true (Incremental || IsIncremental)
	if model.Incremental.ValueBool() != true {
		t.Error("expected Incremental to be true when IsIncremental is true")
	}
}

func TestMapAPIResponseToState_BothIncrementalFalse(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:            "bak-not-incr",
		Name:          "not-incr-bak",
		Status:        "available",
		Incremental:   false,
		IsIncremental: false,
	}

	mapAPIResponseToState(model, apiResp)

	if model.Incremental.ValueBool() != false {
		t.Error("expected Incremental to be false when both fields are false")
	}
}

func TestMapAPIResponseToState_DescriptionSetThenCleared(t *testing.T) {
	// Model had a description previously set
	model := &volumeBackupResourceModel{
		Description: types.StringValue("old description"),
		SnapshotID:  types.StringNull(),
	}

	// API returns empty description
	apiResp := &backupAPIResponse{
		ID:          "bak-cleared",
		Name:        "cleared-bak",
		VolumeID:    "vol-cleared",
		Description: "",
		Status:      "available",
	}

	mapAPIResponseToState(model, apiResp)

	// When Description was previously set (not null) and API returns empty,
	// model.Description should keep its previous value
	if model.Description.ValueString() != "old description" {
		t.Errorf("expected Description to remain 'old description', got %q", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_SnapshotIDSetThenCleared(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringValue("old-snap-id"),
	}

	apiResp := &backupAPIResponse{
		ID:         "bak-snap-cleared",
		Name:       "snap-cleared-bak",
		VolumeID:   "vol-sc",
		SnapshotID: "",
		Status:     "available",
	}

	mapAPIResponseToState(model, apiResp)

	// SnapshotID was previously set, API returns empty, should keep previous value
	if model.SnapshotID.ValueString() != "old-snap-id" {
		t.Errorf("expected SnapshotID to remain 'old-snap-id', got %q", model.SnapshotID.ValueString())
	}
}

func TestMapAPIResponseToState_LargeSize(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:     "bak-large",
		Name:   "large-bak",
		Status: "available",
		Size:   16384,
	}

	mapAPIResponseToState(model, apiResp)

	if model.Size.ValueInt64() != 16384 {
		t.Errorf("expected Size 16384, got %d", model.Size.ValueInt64())
	}
}

func TestMapAPIResponseToState_EmptyStatus(t *testing.T) {
	model := &volumeBackupResourceModel{
		Description: types.StringNull(),
		SnapshotID:  types.StringNull(),
	}

	apiResp := &backupAPIResponse{
		ID:     "bak-empty-status",
		Name:   "empty-status-bak",
		Status: "",
	}

	mapAPIResponseToState(model, apiResp)

	if model.Status.ValueString() != "" {
		t.Errorf("expected empty status, got %q", model.Status.ValueString())
	}
	if model.Status.IsNull() {
		t.Error("expected Status to be known (not null)")
	}
}

func TestMapAPIResponseToState_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		initialDescNull bool
		initialSnapNull bool
		apiResp         backupAPIResponse
		wantID          string
		wantName        string
		wantStatus      string
		wantSize        int64
		wantIncr        bool
		wantDescNull    bool
		wantDescVal     string
		wantSnapNull    bool
		wantSnapVal     string
	}{
		{
			name:            "full response",
			initialDescNull: true,
			initialSnapNull: true,
			apiResp: backupAPIResponse{
				ID: "bak-t1", Name: "t1", VolumeID: "vol-t1",
				SnapshotID: "snap-t1", Description: "desc-t1",
				Status: "available", Size: 10, Incremental: true,
				CreatedAt: "2024-01-01T00:00:00Z",
			},
			wantID: "bak-t1", wantName: "t1", wantStatus: "available", wantSize: 10,
			wantIncr: true, wantDescNull: false, wantDescVal: "desc-t1",
			wantSnapNull: false, wantSnapVal: "snap-t1",
		},
		{
			name:            "empty optionals keep null",
			initialDescNull: true,
			initialSnapNull: true,
			apiResp: backupAPIResponse{
				ID: "bak-t2", Name: "t2", Status: "creating",
			},
			wantID: "bak-t2", wantName: "t2", wantStatus: "creating", wantSize: 0,
			wantIncr: false, wantDescNull: true, wantSnapNull: true,
		},
		{
			name:            "is_incremental fallback",
			initialDescNull: true,
			initialSnapNull: true,
			apiResp: backupAPIResponse{
				ID: "bak-t3", Name: "t3", Status: "available",
				IsIncremental: true,
			},
			wantID: "bak-t3", wantName: "t3", wantStatus: "available", wantSize: 0,
			wantIncr: true, wantDescNull: true, wantSnapNull: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &volumeBackupResourceModel{}
			if tc.initialDescNull {
				model.Description = types.StringNull()
			}
			if tc.initialSnapNull {
				model.SnapshotID = types.StringNull()
			}

			mapAPIResponseToState(model, &tc.apiResp)

			if model.ID.ValueString() != tc.wantID {
				t.Errorf("expected ID %q, got %q", tc.wantID, model.ID.ValueString())
			}
			if model.Name.ValueString() != tc.wantName {
				t.Errorf("expected Name %q, got %q", tc.wantName, model.Name.ValueString())
			}
			if model.Status.ValueString() != tc.wantStatus {
				t.Errorf("expected Status %q, got %q", tc.wantStatus, model.Status.ValueString())
			}
			if model.Size.ValueInt64() != tc.wantSize {
				t.Errorf("expected Size %d, got %d", tc.wantSize, model.Size.ValueInt64())
			}
			if model.Incremental.ValueBool() != tc.wantIncr {
				t.Errorf("expected Incremental %v, got %v", tc.wantIncr, model.Incremental.ValueBool())
			}
			if tc.wantDescNull {
				if !model.Description.IsNull() {
					t.Error("expected Description to be null")
				}
			} else {
				if model.Description.ValueString() != tc.wantDescVal {
					t.Errorf("expected Description %q, got %q", tc.wantDescVal, model.Description.ValueString())
				}
			}
			if tc.wantSnapNull {
				if !model.SnapshotID.IsNull() {
					t.Error("expected SnapshotID to be null")
				}
			} else {
				if model.SnapshotID.ValueString() != tc.wantSnapVal {
					t.Errorf("expected SnapshotID %q, got %q", tc.wantSnapVal, model.SnapshotID.ValueString())
				}
			}
		})
	}
}

// --- Update body construction tests ---

func TestBuildUpdateRequest(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("updated-backup"),
		Description: types.StringValue("Updated description"),
	}

	updateBody := backupUpdateWrapper{
		Backup: backupUpdateBody{
			Name: plan.Name.ValueString(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateBody.Backup.Description = plan.Description.ValueString()
	}

	data, err := json.Marshal(updateBody)
	if err != nil {
		t.Fatalf("failed to marshal update request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify wrapper format: {"backup": {"name": ..., "description": ...}}
	backupRaw, ok := raw["backup"]
	if !ok {
		t.Fatal("expected JSON to have 'backup' wrapper key")
	}

	backup, ok := backupRaw.(map[string]interface{})
	if !ok {
		t.Fatal("expected 'backup' to be an object")
	}

	if backup["name"] != "updated-backup" {
		t.Errorf("expected backup.name 'updated-backup', got %v", backup["name"])
	}
	if backup["description"] != "Updated description" {
		t.Errorf("expected backup.description 'Updated description', got %v", backup["description"])
	}
}

func TestBuildUpdateRequest_NameOnly(t *testing.T) {
	plan := &volumeBackupResourceModel{
		Name:        types.StringValue("renamed-backup"),
		Description: types.StringNull(),
	}

	updateBody := backupUpdateWrapper{
		Backup: backupUpdateBody{
			Name: plan.Name.ValueString(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateBody.Backup.Description = plan.Description.ValueString()
	}

	data, err := json.Marshal(updateBody)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	backup := raw["backup"].(map[string]interface{})
	if backup["name"] != "renamed-backup" {
		t.Errorf("expected name 'renamed-backup', got %v", backup["name"])
	}
	if _, ok := backup["description"]; ok {
		t.Error("expected 'description' to be omitted from JSON when null")
	}
}

func TestBuildUpdateRequest_JSONStructure(t *testing.T) {
	tests := []struct {
		name     string
		bakName  string
		desc     string
		wantDesc bool
	}{
		{"with description", "bak-a", "some desc", true},
		{"without description", "bak-b", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updateBody := backupUpdateWrapper{
				Backup: backupUpdateBody{
					Name:        tc.bakName,
					Description: tc.desc,
				},
			}

			data, err := json.Marshal(updateBody)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var raw map[string]interface{}
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			backup := raw["backup"].(map[string]interface{})
			if backup["name"] != tc.bakName {
				t.Errorf("expected name %q, got %v", tc.bakName, backup["name"])
			}
			_, hasDesc := backup["description"]
			if hasDesc != tc.wantDesc {
				t.Errorf("expected description present=%v, got %v", tc.wantDesc, hasDesc)
			}
		})
	}
}

// --- parseBackupResponse tests ---

func TestParseBackupResponse_WrappedFormat(t *testing.T) {
	data := json.RawMessage(`{"backup": {"id": "b-123", "name": "test", "status": "available"}}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "b-123" {
		t.Errorf("expected ID b-123, got %s", result.ID)
	}
}

func TestParseBackupResponse_DirectFormat(t *testing.T) {
	data := json.RawMessage(`{"id": "b-456", "name": "direct-test", "status": "available"}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "b-456" {
		t.Errorf("expected ID b-456, got %s", result.ID)
	}
}

func TestParseBackupResponse_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`{invalid json!!!}`)

	_, err := parseBackupResponse(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseBackupResponse_EmptyData(t *testing.T) {
	data := json.RawMessage(`{}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error for empty data: %v", err)
	}
	if result.ID != "" {
		t.Errorf("expected empty ID for empty data, got %s", result.ID)
	}
}

func TestParseBackupResponse_WrappedAllFields(t *testing.T) {
	data := json.RawMessage(`{
		"backup": {
			"id": "b-wrapped-full",
			"name": "wrapped-full",
			"volume_id": "vol-wrapped",
			"snapshot_id": "snap-wrapped",
			"description": "wrapped desc",
			"incremental": true,
			"is_incremental": true,
			"status": "available",
			"size": 100,
			"created_at": "2024-03-01T12:00:00Z"
		}
	}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "b-wrapped-full" {
		t.Errorf("expected ID 'b-wrapped-full', got %q", result.ID)
	}
	if result.VolumeID != "vol-wrapped" {
		t.Errorf("expected VolumeID 'vol-wrapped', got %q", result.VolumeID)
	}
	if result.SnapshotID != "snap-wrapped" {
		t.Errorf("expected SnapshotID 'snap-wrapped', got %q", result.SnapshotID)
	}
	if result.Description != "wrapped desc" {
		t.Errorf("expected Description 'wrapped desc', got %q", result.Description)
	}
	if result.Incremental != true {
		t.Error("expected Incremental true")
	}
	if result.IsIncremental != true {
		t.Error("expected IsIncremental true")
	}
	if result.Size != 100 {
		t.Errorf("expected Size 100, got %d", result.Size)
	}
}

func TestParseBackupResponse_DirectAllFields(t *testing.T) {
	data := json.RawMessage(`{
		"id": "b-direct-full",
		"name": "direct-full",
		"volume_id": "vol-direct",
		"snapshot_id": "snap-direct",
		"description": "direct desc",
		"incremental": false,
		"status": "creating",
		"size": 200,
		"created_at": "2024-04-01T12:00:00Z"
	}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "b-direct-full" {
		t.Errorf("expected ID 'b-direct-full', got %q", result.ID)
	}
	if result.Size != 200 {
		t.Errorf("expected Size 200, got %d", result.Size)
	}
	if result.Incremental != false {
		t.Error("expected Incremental false")
	}
}

func TestParseBackupResponse_WrappedEmptyID(t *testing.T) {
	data := json.RawMessage(`{"backup": {"id": "", "name": "empty-id"}}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "" {
		t.Errorf("expected empty ID, got %q", result.ID)
	}
}

func TestParseBackupResponse_NullValues(t *testing.T) {
	data := json.RawMessage(`{"id": "b-null", "name": "null-test", "description": null, "snapshot_id": null, "status": "available"}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "b-null" {
		t.Errorf("expected ID 'b-null', got %q", result.ID)
	}
	if result.Description != "" {
		t.Errorf("expected empty description for null JSON, got %q", result.Description)
	}
	if result.SnapshotID != "" {
		t.Errorf("expected empty snapshot_id for null JSON, got %q", result.SnapshotID)
	}
}

func TestParseBackupResponse_ExtraFields(t *testing.T) {
	data := json.RawMessage(`{"id": "b-extra", "name": "extra", "status": "available", "unknown_field": "ignore-me"}`)

	result, err := parseBackupResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "b-extra" {
		t.Errorf("expected ID 'b-extra', got %q", result.ID)
	}
}

// --- Delete request tests ---

func TestBackupDeleteRequest(t *testing.T) {
	req := backupDeleteRequest{
		Key:    "id",
		Values: []string{"bak-del-001"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal delete request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if raw["key"] != "id" {
		t.Errorf("expected key 'id', got %v", raw["key"])
	}

	values, ok := raw["values"].([]interface{})
	if !ok {
		t.Fatal("expected 'values' to be an array")
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0] != "bak-del-001" {
		t.Errorf("expected value 'bak-del-001', got %v", values[0])
	}
}

func TestBackupDeleteRequest_MultipleValues(t *testing.T) {
	req := backupDeleteRequest{
		Key:    "id",
		Values: []string{"bak-del-001", "bak-del-002", "bak-del-003"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal delete request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	values, ok := raw["values"].([]interface{})
	if !ok {
		t.Fatal("expected 'values' to be an array")
	}
	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}
}

func TestBackupDeleteRequest_JSONKeys(t *testing.T) {
	req := backupDeleteRequest{
		Key:    "id",
		Values: []string{"bak-123"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(raw) != 2 {
		t.Errorf("expected 2 JSON keys, got %d", len(raw))
	}
	if _, ok := raw["key"]; !ok {
		t.Error("expected 'key' field in JSON")
	}
	if _, ok := raw["values"]; !ok {
		t.Error("expected 'values' field in JSON")
	}
}

// --- Schema / model tests ---

func TestVolumeBackupSchema_AttributesExist(t *testing.T) {
	s := volumeBackupSchema()

	expectedAttrs := []string{"id", "name", "volume_id", "snapshot_id", "description", "incremental", "status", "size", "created_at"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q in schema", attr)
		}
	}
}

func TestVolumeBackupSchema_RequiredFields(t *testing.T) {
	s := volumeBackupSchema()

	nameAttr := s.Attributes["name"].(schema.StringAttribute)
	if !nameAttr.Required {
		t.Error("expected 'name' to be required")
	}

	volAttr := s.Attributes["volume_id"].(schema.StringAttribute)
	if !volAttr.Required {
		t.Error("expected 'volume_id' to be required")
	}
}

func TestVolumeBackupSchema_ComputedFields(t *testing.T) {
	s := volumeBackupSchema()

	computedStringFields := []string{"id", "status", "created_at"}
	for _, field := range computedStringFields {
		attr := s.Attributes[field].(schema.StringAttribute)
		if !attr.Computed {
			t.Errorf("expected %q to be computed", field)
		}
	}

	sizeAttr := s.Attributes["size"].(schema.Int64Attribute)
	if !sizeAttr.Computed {
		t.Error("expected 'size' to be computed")
	}

	incrAttr := s.Attributes["incremental"].(schema.BoolAttribute)
	if !incrAttr.Computed {
		t.Error("expected 'incremental' to be computed")
	}
}

func TestVolumeBackupSchema_OptionalFields(t *testing.T) {
	s := volumeBackupSchema()

	descAttr := s.Attributes["description"].(schema.StringAttribute)
	if !descAttr.Optional {
		t.Error("expected 'description' to be optional")
	}

	snapAttr := s.Attributes["snapshot_id"].(schema.StringAttribute)
	if !snapAttr.Optional {
		t.Error("expected 'snapshot_id' to be optional")
	}

	incrAttr := s.Attributes["incremental"].(schema.BoolAttribute)
	if !incrAttr.Optional {
		t.Error("expected 'incremental' to be optional")
	}
}

func TestVolumeBackupSchema_Description(t *testing.T) {
	s := volumeBackupSchema()

	if s.Description != "Manages an Ace Cloud volume backup." {
		t.Errorf("unexpected schema description: %q", s.Description)
	}
}

// --- NewResource test ---

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}
