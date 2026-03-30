package snapshot

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- buildCreateRequest tests ---

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("test-snapshot"),
		VolumeID:    types.StringValue("vol-123"),
		Description: types.StringValue("A test snapshot"),
	}

	body := buildCreateRequest(plan)

	if body.Name != "test-snapshot" {
		t.Errorf("expected name test-snapshot, got %s", body.Name)
	}
	if body.VolumeID != "vol-123" {
		t.Errorf("expected volume_id vol-123, got %s", body.VolumeID)
	}
	if body.Description != "A test snapshot" {
		t.Errorf("expected description 'A test snapshot', got %s", body.Description)
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("minimal-snapshot"),
		VolumeID:    types.StringValue("vol-456"),
		Description: types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.Name != "minimal-snapshot" {
		t.Errorf("expected name minimal-snapshot, got %s", body.Name)
	}
	if body.Description != "" {
		t.Errorf("expected empty description, got %s", body.Description)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("json-snapshot"),
		VolumeID:    types.StringValue("vol-json-123"),
		Description: types.StringValue("JSON test desc"),
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

	// Verify JSON field names match API expectations
	if _, ok := raw["name"]; !ok {
		t.Error("expected JSON key 'name' to be present")
	}
	if _, ok := raw["volume_id"]; !ok {
		t.Error("expected JSON key 'volume_id' to be present")
	}
	if _, ok := raw["description"]; !ok {
		t.Error("expected JSON key 'description' to be present")
	}

	if raw["name"] != "json-snapshot" {
		t.Errorf("expected JSON name 'json-snapshot', got %v", raw["name"])
	}
	if raw["volume_id"] != "vol-json-123" {
		t.Errorf("expected JSON volume_id 'vol-json-123', got %v", raw["volume_id"])
	}
	if raw["description"] != "JSON test desc" {
		t.Errorf("expected JSON description 'JSON test desc', got %v", raw["description"])
	}
}

func TestBuildCreateRequest_WithDescription(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("force-snap"),
		VolumeID:    types.StringValue("vol-force-123"),
		Description: types.StringValue("Snapshot with description"),
	}

	body := buildCreateRequest(plan)

	if body.Name != "force-snap" {
		t.Errorf("expected name force-snap, got %s", body.Name)
	}
	if body.VolumeID != "vol-force-123" {
		t.Errorf("expected volume_id vol-force-123, got %s", body.VolumeID)
	}
	if body.Description != "Snapshot with description" {
		t.Errorf("expected description 'Snapshot with description', got %s", body.Description)
	}
}

func TestBuildCreateRequest_UnknownDescription(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("unknown-desc-snap"),
		VolumeID:    types.StringValue("vol-unknown"),
		Description: types.StringUnknown(),
	}

	body := buildCreateRequest(plan)

	if body.Description != "" {
		t.Errorf("expected empty description for unknown, got %s", body.Description)
	}
}

func TestBuildCreateRequest_EmptyStringDescription(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("empty-desc-snap"),
		VolumeID:    types.StringValue("vol-empty"),
		Description: types.StringValue(""),
	}

	body := buildCreateRequest(plan)

	if body.Description != "" {
		t.Errorf("expected empty description for empty string value, got %q", body.Description)
	}
}

func TestBuildCreateRequest_JSONOmitsEmptyDescription(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("omit-test"),
		VolumeID:    types.StringValue("vol-omit"),
		Description: types.StringNull(),
	}

	body := buildCreateRequest(plan)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// description has omitempty, so it should be absent when empty
	if _, ok := raw["description"]; ok {
		t.Error("expected 'description' to be omitted from JSON when null")
	}
}

func TestBuildCreateRequest_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		planName    string
		volumeID    string
		description types.String
		wantName    string
		wantVolID   string
		wantDesc    string
	}{
		{
			name:        "all fields populated",
			planName:    "snap-full",
			volumeID:    "vol-full",
			description: types.StringValue("full description"),
			wantName:    "snap-full",
			wantVolID:   "vol-full",
			wantDesc:    "full description",
		},
		{
			name:        "null description",
			planName:    "snap-null",
			volumeID:    "vol-null",
			description: types.StringNull(),
			wantName:    "snap-null",
			wantVolID:   "vol-null",
			wantDesc:    "",
		},
		{
			name:        "unknown description",
			planName:    "snap-unk",
			volumeID:    "vol-unk",
			description: types.StringUnknown(),
			wantName:    "snap-unk",
			wantVolID:   "vol-unk",
			wantDesc:    "",
		},
		{
			name:        "special characters in name",
			planName:    "snap-with-hyphens-123",
			volumeID:    "vol-special-456",
			description: types.StringValue("desc with spaces, periods."),
			wantName:    "snap-with-hyphens-123",
			wantVolID:   "vol-special-456",
			wantDesc:    "desc with spaces, periods.",
		},
		{
			name:        "long name",
			planName:    "a-very-long-snapshot-name-that-has-many-characters",
			volumeID:    "vol-long",
			description: types.StringNull(),
			wantName:    "a-very-long-snapshot-name-that-has-many-characters",
			wantVolID:   "vol-long",
			wantDesc:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := &snapshotResourceModel{
				Name:        types.StringValue(tc.planName),
				VolumeID:    types.StringValue(tc.volumeID),
				Description: tc.description,
			}

			body := buildCreateRequest(plan)

			if body.Name != tc.wantName {
				t.Errorf("expected name %q, got %q", tc.wantName, body.Name)
			}
			if body.VolumeID != tc.wantVolID {
				t.Errorf("expected volume_id %q, got %q", tc.wantVolID, body.VolumeID)
			}
			if body.Description != tc.wantDesc {
				t.Errorf("expected description %q, got %q", tc.wantDesc, body.Description)
			}
		})
	}
}

// --- mapAPIResponseToState tests ---

func TestMapAPIResponseToState(t *testing.T) {
	model := &snapshotResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:          "snap-abc-123",
		Name:        "prod-snapshot",
		VolumeID:    "vol-xyz",
		Description: "Production snapshot",
		Status:      "available",
		Size:        50,
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-02T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "snap-abc-123" {
		t.Errorf("expected ID snap-abc-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-snapshot" {
		t.Errorf("expected Name prod-snapshot, got %s", model.Name.ValueString())
	}
	if model.VolumeID.ValueString() != "vol-xyz" {
		t.Errorf("expected VolumeID vol-xyz, got %s", model.VolumeID.ValueString())
	}
	if model.Status.ValueString() != "available" {
		t.Errorf("expected Status available, got %s", model.Status.ValueString())
	}
	if model.Size.ValueInt64() != 50 {
		t.Errorf("expected Size 50, got %d", model.Size.ValueInt64())
	}
}

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	model := &snapshotResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:          "snap-full-001",
		Name:        "full-snapshot",
		VolumeID:    "vol-full-xyz",
		Description: "Full snapshot with all fields",
		Status:      "available",
		Size:        250,
		CreatedAt:   "2024-06-15T10:30:00Z",
		UpdatedAt:   "2024-06-16T14:45:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "snap-full-001" {
		t.Errorf("expected ID snap-full-001, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-snapshot" {
		t.Errorf("expected Name full-snapshot, got %s", model.Name.ValueString())
	}
	if model.VolumeID.ValueString() != "vol-full-xyz" {
		t.Errorf("expected VolumeID vol-full-xyz, got %s", model.VolumeID.ValueString())
	}
	// When model Description is null (user didn't set), API value is NOT injected.
	if !model.Description.IsNull() {
		t.Errorf("expected Description null (user didn't set), got %s", model.Description.ValueString())
	}
	if model.Status.ValueString() != "available" {
		t.Errorf("expected Status available, got %s", model.Status.ValueString())
	}
	if model.Size.ValueInt64() != 250 {
		t.Errorf("expected Size 250, got %d", model.Size.ValueInt64())
	}
	if model.CreatedAt.ValueString() != "2024-06-15T10:30:00Z" {
		t.Errorf("expected CreatedAt '2024-06-15T10:30:00Z', got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2024-06-16T14:45:00Z" {
		t.Errorf("expected UpdatedAt '2024-06-16T14:45:00Z', got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &snapshotResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:          "snap-123",
		Name:        "basic",
		Description: "",
		Status:      "creating",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
}

func TestMapAPIResponseToState_Status(t *testing.T) {
	statuses := []string{"available", "creating", "deleting", "error", "error_deleting"}
	for _, status := range statuses {
		model := &snapshotResourceModel{
			Description: types.StringNull(),
		}

		apiResp := &snapshotAPIResponse{
			ID:     "snap-status-test",
			Name:   "status-snap",
			Status: status,
		}

		mapAPIResponseToState(model, apiResp)

		if model.Status.ValueString() != status {
			t.Errorf("expected status %q, got %q", status, model.Status.ValueString())
		}
	}
}

func TestMapAPIResponseToState_EmptyVolumeID(t *testing.T) {
	// When VolumeID is empty in response, model.VolumeID should NOT be overwritten
	model := &snapshotResourceModel{
		VolumeID:    types.StringValue("original-vol-id"),
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:     "snap-novol",
		Name:   "no-vol-snap",
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
	model := &snapshotResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:     "snap-zero-size",
		Name:   "zero-size-snap",
		Status: "creating",
		Size:   0,
	}

	mapAPIResponseToState(model, apiResp)

	if model.Size.ValueInt64() != 0 {
		t.Errorf("expected Size 0, got %d", model.Size.ValueInt64())
	}
}

func TestMapAPIResponseToState_EmptyTimestamps(t *testing.T) {
	model := &snapshotResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:     "snap-no-time",
		Name:   "no-time-snap",
		Status: "creating",
		// CreatedAt and UpdatedAt intentionally empty
	}

	mapAPIResponseToState(model, apiResp)

	// Even empty timestamps should be set as known values (empty strings)
	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %q", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "" {
		t.Errorf("expected empty UpdatedAt, got %q", model.UpdatedAt.ValueString())
	}
	if model.CreatedAt.IsNull() {
		t.Error("expected CreatedAt to be known (not null)")
	}
	if model.UpdatedAt.IsNull() {
		t.Error("expected UpdatedAt to be known (not null)")
	}
}

func TestMapAPIResponseToState_DescriptionSetThenCleared(t *testing.T) {
	// Model had a description previously set
	model := &snapshotResourceModel{
		Description: types.StringValue("old description"),
	}

	// API returns empty description
	apiResp := &snapshotAPIResponse{
		ID:          "snap-cleared",
		Name:        "cleared-snap",
		VolumeID:    "vol-cleared",
		Description: "",
		Status:      "available",
	}

	mapAPIResponseToState(model, apiResp)

	// When Description was previously set (not null) and API returns empty,
	// model.Description should keep its previous value since the else-if only applies when IsNull
	if model.Description.ValueString() != "old description" {
		t.Errorf("expected Description to remain 'old description', got %q", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyStatus(t *testing.T) {
	model := &snapshotResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:     "snap-empty-status",
		Name:   "empty-status-snap",
		Status: "",
	}

	mapAPIResponseToState(model, apiResp)

	// Empty status should still be set as a known value
	if model.Status.ValueString() != "" {
		t.Errorf("expected empty status, got %q", model.Status.ValueString())
	}
	if model.Status.IsNull() {
		t.Error("expected Status to be known (not null)")
	}
}

func TestMapAPIResponseToState_LargeSize(t *testing.T) {
	model := &snapshotResourceModel{
		Description: types.StringNull(),
	}

	apiResp := &snapshotAPIResponse{
		ID:     "snap-large",
		Name:   "large-snap",
		Status: "available",
		Size:   16384, // 16TB
	}

	mapAPIResponseToState(model, apiResp)

	if model.Size.ValueInt64() != 16384 {
		t.Errorf("expected Size 16384, got %d", model.Size.ValueInt64())
	}
}

func TestMapAPIResponseToState_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		initialDescNull bool
		apiResp         snapshotAPIResponse
		wantID          string
		wantName        string
		wantStatus      string
		wantSize        int64
		wantDescNull    bool
		wantDescVal     string
	}{
		{
			name:            "full response with description null model",
			initialDescNull: true,
			apiResp: snapshotAPIResponse{
				ID: "snap-t1", Name: "t1", VolumeID: "vol-t1",
				Description: "desc-t1", Status: "available", Size: 10,
				CreatedAt: "2024-01-01T00:00:00Z", UpdatedAt: "2024-01-02T00:00:00Z",
			},
			wantID: "snap-t1", wantName: "t1", wantStatus: "available", wantSize: 10,
			wantDescNull: true,
		},
		{
			name:            "empty desc keeps null",
			initialDescNull: true,
			apiResp: snapshotAPIResponse{
				ID: "snap-t2", Name: "t2", Status: "creating",
			},
			wantID: "snap-t2", wantName: "t2", wantStatus: "creating", wantSize: 0,
			wantDescNull: true,
		},
		{
			name:            "zero size reported",
			initialDescNull: true,
			apiResp: snapshotAPIResponse{
				ID: "snap-t3", Name: "t3", Status: "available", Size: 0,
			},
			wantID: "snap-t3", wantName: "t3", wantStatus: "available", wantSize: 0,
			wantDescNull: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &snapshotResourceModel{}
			if tc.initialDescNull {
				model.Description = types.StringNull()
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
			if tc.wantDescNull {
				if !model.Description.IsNull() {
					t.Error("expected Description to be null")
				}
			} else {
				if model.Description.ValueString() != tc.wantDescVal {
					t.Errorf("expected Description %q, got %q", tc.wantDescVal, model.Description.ValueString())
				}
			}
		})
	}
}

// --- Update body construction tests ---

func TestBuildUpdateRequest_NameAndDescription(t *testing.T) {
	updateBody := snapshotUpdateRequest{
		Snapshot: snapshotUpdateBody{
			Name:        "updated-snapshot",
			Description: "Updated description",
		},
	}

	data, err := json.Marshal(updateBody)
	if err != nil {
		t.Fatalf("failed to marshal update request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	snapshotRaw, ok := raw["snapshot"]
	if !ok {
		t.Fatal("expected JSON to have 'snapshot' wrapper key")
	}

	snap, ok := snapshotRaw.(map[string]interface{})
	if !ok {
		t.Fatal("expected 'snapshot' to be an object")
	}

	if snap["name"] != "updated-snapshot" {
		t.Errorf("expected snapshot.name 'updated-snapshot', got %v", snap["name"])
	}
	if snap["description"] != "Updated description" {
		t.Errorf("expected snapshot.description 'Updated description', got %v", snap["description"])
	}
}

func TestBuildUpdateRequest_NameOnly(t *testing.T) {
	updateBody := snapshotUpdateRequest{
		Snapshot: snapshotUpdateBody{
			Name: "renamed-snapshot",
			// Description omitted (empty string with omitempty)
		},
	}

	data, err := json.Marshal(updateBody)
	if err != nil {
		t.Fatalf("failed to marshal update request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	snap := raw["snapshot"].(map[string]interface{})

	if snap["name"] != "renamed-snapshot" {
		t.Errorf("expected name 'renamed-snapshot', got %v", snap["name"])
	}
	if _, ok := snap["description"]; ok {
		t.Error("expected 'description' to be omitted from JSON when empty")
	}
}

func TestBuildUpdateRequest_FromModel(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("model-update"),
		Description: types.StringValue("model desc"),
	}

	updateBody := snapshotUpdateRequest{
		Snapshot: snapshotUpdateBody{
			Name: plan.Name.ValueString(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateBody.Snapshot.Description = plan.Description.ValueString()
	}

	if updateBody.Snapshot.Name != "model-update" {
		t.Errorf("expected name 'model-update', got %s", updateBody.Snapshot.Name)
	}
	if updateBody.Snapshot.Description != "model desc" {
		t.Errorf("expected description 'model desc', got %s", updateBody.Snapshot.Description)
	}
}

func TestBuildUpdateRequest_FromModelNullDescription(t *testing.T) {
	plan := &snapshotResourceModel{
		Name:        types.StringValue("no-desc-update"),
		Description: types.StringNull(),
	}

	updateBody := snapshotUpdateRequest{
		Snapshot: snapshotUpdateBody{
			Name: plan.Name.ValueString(),
		},
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateBody.Snapshot.Description = plan.Description.ValueString()
	}

	if updateBody.Snapshot.Description != "" {
		t.Errorf("expected empty description, got %q", updateBody.Snapshot.Description)
	}
}

func TestBuildUpdateRequest_JSONStructure(t *testing.T) {
	tests := []struct {
		name     string
		snapName string
		desc     string
		wantDesc bool
	}{
		{"with description", "snap-a", "some desc", true},
		{"without description", "snap-b", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updateBody := snapshotUpdateRequest{
				Snapshot: snapshotUpdateBody{
					Name:        tc.snapName,
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

			snap := raw["snapshot"].(map[string]interface{})
			if snap["name"] != tc.snapName {
				t.Errorf("expected name %q, got %v", tc.snapName, snap["name"])
			}
			_, hasDesc := snap["description"]
			if hasDesc != tc.wantDesc {
				t.Errorf("expected description present=%v, got %v", tc.wantDesc, hasDesc)
			}
		})
	}
}

// --- parseSnapshotResponse tests ---

func TestParseSnapshotResponse_WrappedFormat(t *testing.T) {
	data := json.RawMessage(`{"snapshot": {"id": "s-123", "name": "test", "status": "available"}}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "s-123" {
		t.Errorf("expected ID s-123, got %s", result.ID)
	}
}

func TestParseSnapshotResponse_DirectFormat(t *testing.T) {
	data := json.RawMessage(`{"id": "s-456", "name": "direct-test", "status": "available"}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "s-456" {
		t.Errorf("expected ID s-456, got %s", result.ID)
	}
}

func TestParseSnapshotResponse_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`{invalid json!!!}`)

	_, err := parseSnapshotResponse(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseSnapshotResponse_EmptyData(t *testing.T) {
	data := json.RawMessage(`{}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error for empty data: %v", err)
	}
	if result.ID != "" {
		t.Errorf("expected empty ID for empty data, got %s", result.ID)
	}
}

func TestParseSnapshotResponse_WrappedAllFields(t *testing.T) {
	data := json.RawMessage(`{
		"snapshot": {
			"id": "s-wrapped-full",
			"name": "wrapped-full",
			"volume_id": "vol-wrapped",
			"description": "wrapped desc",
			"status": "available",
			"size": 100,
			"created_at": "2024-03-01T12:00:00Z",
			"updated_at": "2024-03-02T12:00:00Z"
		}
	}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "s-wrapped-full" {
		t.Errorf("expected ID 's-wrapped-full', got %q", result.ID)
	}
	if result.VolumeID != "vol-wrapped" {
		t.Errorf("expected VolumeID 'vol-wrapped', got %q", result.VolumeID)
	}
	if result.Description != "wrapped desc" {
		t.Errorf("expected Description 'wrapped desc', got %q", result.Description)
	}
	if result.Size != 100 {
		t.Errorf("expected Size 100, got %d", result.Size)
	}
	if result.CreatedAt != "2024-03-01T12:00:00Z" {
		t.Errorf("expected CreatedAt '2024-03-01T12:00:00Z', got %q", result.CreatedAt)
	}
	if result.UpdatedAt != "2024-03-02T12:00:00Z" {
		t.Errorf("expected UpdatedAt '2024-03-02T12:00:00Z', got %q", result.UpdatedAt)
	}
}

func TestParseSnapshotResponse_DirectAllFields(t *testing.T) {
	data := json.RawMessage(`{
		"id": "s-direct-full",
		"name": "direct-full",
		"volume_id": "vol-direct",
		"description": "direct desc",
		"status": "creating",
		"size": 200,
		"created_at": "2024-04-01T12:00:00Z",
		"updated_at": "2024-04-02T12:00:00Z"
	}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "s-direct-full" {
		t.Errorf("expected ID 's-direct-full', got %q", result.ID)
	}
	if result.VolumeID != "vol-direct" {
		t.Errorf("expected VolumeID 'vol-direct', got %q", result.VolumeID)
	}
	if result.Size != 200 {
		t.Errorf("expected Size 200, got %d", result.Size)
	}
}

func TestParseSnapshotResponse_WrappedEmptyID(t *testing.T) {
	// Wrapped format but with empty ID should fall back to direct parse
	data := json.RawMessage(`{"snapshot": {"id": "", "name": "empty-id"}}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Falls through to direct parse, which finds no direct ID
	if result.ID != "" {
		t.Errorf("expected empty ID, got %q", result.ID)
	}
}

func TestParseSnapshotResponse_NullValue(t *testing.T) {
	data := json.RawMessage(`{"id": "s-null", "name": "null-test", "description": null, "status": "available"}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "s-null" {
		t.Errorf("expected ID 's-null', got %q", result.ID)
	}
	if result.Description != "" {
		t.Errorf("expected empty description for null JSON, got %q", result.Description)
	}
}

func TestParseSnapshotResponse_ExtraFields(t *testing.T) {
	// API may return extra fields not in our struct — should be ignored
	data := json.RawMessage(`{"id": "s-extra", "name": "extra", "status": "available", "unknown_field": "ignore-me", "count": 42}`)

	result, err := parseSnapshotResponse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "s-extra" {
		t.Errorf("expected ID 's-extra', got %q", result.ID)
	}
}

// --- Delete request tests ---

func TestSnapshotDeleteRequest(t *testing.T) {
	req := snapshotDeleteRequest{
		Key:    "id",
		Values: []string{"snap-del-001"},
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
	if values[0] != "snap-del-001" {
		t.Errorf("expected value 'snap-del-001', got %v", values[0])
	}
}

func TestSnapshotDeleteRequest_MultipleValues(t *testing.T) {
	req := snapshotDeleteRequest{
		Key:    "id",
		Values: []string{"snap-del-001", "snap-del-002", "snap-del-003"},
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

func TestSnapshotDeleteRequest_JSONKeys(t *testing.T) {
	req := snapshotDeleteRequest{
		Key:    "id",
		Values: []string{"snap-123"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify only expected keys are present
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

func TestSnapshotSchema_AttributesExist(t *testing.T) {
	s := snapshotSchema()

	expectedAttrs := []string{"id", "name", "volume_id", "description", "status", "size", "created_at", "updated_at"}
	for _, attr := range expectedAttrs {
		if _, ok := s.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q in schema", attr)
		}
	}
}

func TestSnapshotSchema_RequiredFields(t *testing.T) {
	s := snapshotSchema()

	nameAttr := s.Attributes["name"].(schema.StringAttribute)
	if !nameAttr.Required {
		t.Error("expected 'name' to be required")
	}

	volAttr := s.Attributes["volume_id"].(schema.StringAttribute)
	if !volAttr.Required {
		t.Error("expected 'volume_id' to be required")
	}
}

func TestSnapshotSchema_ComputedFields(t *testing.T) {
	s := snapshotSchema()

	computedStringFields := []string{"id", "status", "created_at", "updated_at"}
	for _, field := range computedStringFields {
		attr := s.Attributes[field].(schema.StringAttribute)
		if !attr.Computed {
			t.Errorf("expected %q to be computed", field)
		}
	}

	// Test size separately since it's Int64Attribute
	sizeAttr := s.Attributes["size"].(schema.Int64Attribute)
	if !sizeAttr.Computed {
		t.Error("expected 'size' to be computed")
	}
}

func TestSnapshotSchema_OptionalFields(t *testing.T) {
	s := snapshotSchema()

	descAttr := s.Attributes["description"].(schema.StringAttribute)
	if !descAttr.Optional {
		t.Error("expected 'description' to be optional")
	}
}

func TestSnapshotSchema_Description(t *testing.T) {
	s := snapshotSchema()

	if s.Description != "Manages an Ace Cloud volume snapshot." {
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
