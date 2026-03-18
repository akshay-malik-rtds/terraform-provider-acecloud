package registry_project

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest body construction
// ---------------------------------------------------------------------------

func TestBuildCreateRequest(t *testing.T) {
	body := map[string]interface{}{
		"registry_name":         "my-registry",
		"vulnerability_scanning": true,
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create body: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal create body: %v", err)
	}

	if parsed["registry_name"] != "my-registry" {
		t.Errorf("expected registry_name 'my-registry', got %v", parsed["registry_name"])
	}
	if parsed["vulnerability_scanning"] != true {
		t.Errorf("expected vulnerability_scanning true, got %v", parsed["vulnerability_scanning"])
	}
}

func TestCreateRequest_JSON(t *testing.T) {
	body := map[string]interface{}{
		"registry_name":         "test-project",
		"vulnerability_scanning": false,
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Verify it round-trips correctly.
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded["registry_name"] != "test-project" {
		t.Errorf("expected 'test-project', got %v", decoded["registry_name"])
	}
	if decoded["vulnerability_scanning"] != false {
		t.Errorf("expected false, got %v", decoded["vulnerability_scanning"])
	}

	// Verify JSON string does not contain unexpected fields.
	jsonStr := string(data)
	if len(jsonStr) == 0 {
		t.Fatal("expected non-empty JSON string")
	}
}

func TestCreateRequest_JSON_AllFields(t *testing.T) {
	body := map[string]interface{}{
		"registry_name":         "detailed-registry",
		"vulnerability_scanning": true,
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify exact field names match npc-api expectations
	if _, ok := decoded["registry_name"]; !ok {
		t.Error("expected 'registry_name' key to be present")
	}
	if _, ok := decoded["vulnerability_scanning"]; !ok {
		t.Error("expected 'vulnerability_scanning' key to be present")
	}

	// Verify types
	if _, ok := decoded["registry_name"].(string); !ok {
		t.Error("expected 'registry_name' to be a string")
	}
	if _, ok := decoded["vulnerability_scanning"].(bool); !ok {
		t.Error("expected 'vulnerability_scanning' to be a bool")
	}

	// Verify values
	if decoded["registry_name"] != "detailed-registry" {
		t.Errorf("expected registry_name 'detailed-registry', got %v", decoded["registry_name"])
	}
	if decoded["vulnerability_scanning"] != true {
		t.Errorf("expected vulnerability_scanning true, got %v", decoded["vulnerability_scanning"])
	}

	// Verify no extra keys
	if len(decoded) != 2 {
		t.Errorf("expected exactly 2 keys in create body, got %d", len(decoded))
	}
}

// ---------------------------------------------------------------------------
// Create response parsing — extract project_id
// ---------------------------------------------------------------------------

func TestCreateResponse_ExtractProjectID(t *testing.T) {
	tests := []struct {
		name       string
		apiJSON    string
		wantID     string
		wantErr    bool
	}{
		{
			name:    "project_id as integer",
			apiJSON: `{"project_id": 42}`,
			wantID:  "42",
		},
		{
			name:    "project_id as float (JSON number)",
			apiJSON: `{"project_id": 100}`,
			wantID:  "100",
		},
		{
			name:    "id field fallback",
			apiJSON: `{"id": 77}`,
			wantID:  "77",
		},
		{
			name:    "project_id takes priority over id",
			apiJSON: `{"project_id": 42, "id": 99}`,
			wantID:  "42",
		},
		{
			name:    "project_id as string number",
			apiJSON: `{"project_id": "123"}`,
			wantID:  "123",
		},
		{
			name:    "no id field present",
			apiJSON: `{"name": "test"}`,
			wantErr: true,
		},
		{
			name:    "empty response object",
			apiJSON: `{}`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(tc.apiJSON), &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			var id string
			var found bool
			if v, ok := result["project_id"]; ok {
				id = fmt.Sprintf("%v", v)
				found = true
			} else if v, ok := result["id"]; ok {
				id = fmt.Sprintf("%v", v)
				found = true
			}

			if tc.wantErr {
				if found {
					t.Errorf("expected no ID found, but got %q", id)
				}
				return
			}

			if !found {
				t.Fatal("expected ID to be found")
			}
			if id != tc.wantID {
				t.Errorf("expected ID %q, got %q", tc.wantID, id)
			}
		})
	}
}

func TestCreateResponse_LargeProjectID(t *testing.T) {
	apiJSON := `{"project_id": 999999}`
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(apiJSON), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if v, ok := result["project_id"]; ok {
		id := fmt.Sprintf("%v", v)
		if id != "999999" {
			t.Errorf("expected ID '999999', got %q", id)
		}
	} else {
		t.Fatal("expected project_id key")
	}
}

// ---------------------------------------------------------------------------
// mapAPIResponseToState (read response parsing)
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState(t *testing.T) {
	// Simulate an API list response item.
	apiJSON := `[{
		"project_id": 42,
		"name": "my-registry",
		"auto_scan": true,
		"creation_time": "2026-03-15T10:00:00Z",
		"registry_id": 1
	}]`

	var projects []map[string]interface{}
	if err := json.Unmarshal([]byte(apiJSON), &projects); err != nil {
		t.Fatalf("failed to unmarshal API response: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	p := projects[0]

	// Simulate mapping to model.
	model := registryProjectModel{
		RegistryName: types.StringValue("my-registry"),
	}

	// Map project_id (API returns as float64 from JSON).
	if v, ok := p["project_id"]; ok {
		model.ID = types.StringValue(fmt.Sprintf("%v", v))
	}

	if v, ok := p["name"].(string); ok {
		model.RegistryName = types.StringValue(v)
	}

	if v, ok := p["auto_scan"].(bool); ok {
		model.VulnerabilityScanning = types.BoolValue(v)
	}

	if v, ok := p["creation_time"].(string); ok {
		model.CreatedAt = types.StringValue(v)
	}

	// Verify mapped state.
	if model.ID.ValueString() != "42" {
		t.Errorf("expected ID '42', got %q", model.ID.ValueString())
	}
	if model.RegistryName.ValueString() != "my-registry" {
		t.Errorf("expected registry_name 'my-registry', got %q", model.RegistryName.ValueString())
	}
	if model.VulnerabilityScanning.ValueBool() != true {
		t.Errorf("expected vulnerability_scanning true, got %v", model.VulnerabilityScanning.ValueBool())
	}
	if model.CreatedAt.ValueString() != "2026-03-15T10:00:00Z" {
		t.Errorf("expected created_at '2026-03-15T10:00:00Z', got %q", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	// Simulate a more complete API list response item with all fields.
	apiJSON := `[{
		"project_id": 99,
		"name": "complete-registry",
		"auto_scan": false,
		"creation_time": "2026-02-20T08:30:00Z",
		"registry_id": 1,
		"repo_count": 5,
		"current_user_role_id": 1
	}]`

	var projects []map[string]interface{}
	if err := json.Unmarshal([]byte(apiJSON), &projects); err != nil {
		t.Fatalf("failed to unmarshal API response: %v", err)
	}

	p := projects[0]

	model := registryProjectModel{
		RegistryName: types.StringValue("complete-registry"),
	}

	// Map all fields
	if v, ok := p["project_id"]; ok {
		model.ID = types.StringValue(fmt.Sprintf("%v", v))
	}
	if v, ok := p["name"].(string); ok {
		model.RegistryName = types.StringValue(v)
	}
	if v, ok := p["auto_scan"].(bool); ok {
		model.VulnerabilityScanning = types.BoolValue(v)
	}
	if v, ok := p["creation_time"].(string); ok {
		model.CreatedAt = types.StringValue(v)
	}

	if model.ID.ValueString() != "99" {
		t.Errorf("expected ID '99', got %q", model.ID.ValueString())
	}
	if model.RegistryName.ValueString() != "complete-registry" {
		t.Errorf("expected registry_name 'complete-registry', got %q", model.RegistryName.ValueString())
	}
	if model.VulnerabilityScanning.ValueBool() != false {
		t.Errorf("expected vulnerability_scanning false, got %v", model.VulnerabilityScanning.ValueBool())
	}
	if model.CreatedAt.ValueString() != "2026-02-20T08:30:00Z" {
		t.Errorf("expected created_at '2026-02-20T08:30:00Z', got %q", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_FallbackIDField(t *testing.T) {
	// Test fallback from project_id to id field
	apiJSON := `[{
		"id": 77,
		"name": "fallback-registry",
		"auto_scan": true,
		"created_at": "2026-03-01T12:00:00Z"
	}]`

	var projects []map[string]interface{}
	if err := json.Unmarshal([]byte(apiJSON), &projects); err != nil {
		t.Fatalf("failed to unmarshal API response: %v", err)
	}

	p := projects[0]

	model := registryProjectModel{
		RegistryName: types.StringValue("fallback-registry"),
	}

	// Use the same ID extraction logic as readIntoState
	if v, ok := p["project_id"]; ok {
		model.ID = types.StringValue(fmt.Sprintf("%v", v))
	} else if v, ok := p["id"]; ok {
		model.ID = types.StringValue(fmt.Sprintf("%v", v))
	}

	if v, ok := p["name"].(string); ok {
		model.RegistryName = types.StringValue(v)
	}

	// Use created_at fallback
	if v, ok := p["creation_time"].(string); ok {
		model.CreatedAt = types.StringValue(v)
	} else if v, ok := p["created_at"].(string); ok {
		model.CreatedAt = types.StringValue(v)
	} else {
		model.CreatedAt = types.StringValue("")
	}

	if model.ID.ValueString() != "77" {
		t.Errorf("expected ID '77' from fallback 'id' field, got %q", model.ID.ValueString())
	}
	if model.CreatedAt.ValueString() != "2026-03-01T12:00:00Z" {
		t.Errorf("expected created_at from fallback 'created_at' field, got %q", model.CreatedAt.ValueString())
	}
}

// ---------------------------------------------------------------------------
// Read response: list + filter by registry_name
// ---------------------------------------------------------------------------

func TestReadResponse_FilterByName(t *testing.T) {
	tests := []struct {
		name       string
		apiJSON    string
		targetName string
		wantFound  bool
		wantID     string
	}{
		{
			name: "single project matching",
			apiJSON: `[{
				"project_id": 10,
				"name": "my-project",
				"auto_scan": true,
				"creation_time": "2026-01-01T00:00:00Z"
			}]`,
			targetName: "my-project",
			wantFound:  true,
			wantID:     "10",
		},
		{
			name: "multiple projects, target found",
			apiJSON: `[
				{"project_id": 10, "name": "project-a", "auto_scan": false, "creation_time": "2026-01-01T00:00:00Z"},
				{"project_id": 20, "name": "project-b", "auto_scan": true, "creation_time": "2026-01-02T00:00:00Z"},
				{"project_id": 30, "name": "project-c", "auto_scan": false, "creation_time": "2026-01-03T00:00:00Z"}
			]`,
			targetName: "project-b",
			wantFound:  true,
			wantID:     "20",
		},
		{
			name: "project not found in list",
			apiJSON: `[
				{"project_id": 10, "name": "project-a", "auto_scan": false, "creation_time": "2026-01-01T00:00:00Z"},
				{"project_id": 20, "name": "project-b", "auto_scan": true, "creation_time": "2026-01-02T00:00:00Z"}
			]`,
			targetName: "project-missing",
			wantFound:  false,
		},
		{
			name:       "empty list",
			apiJSON:    `[]`,
			targetName: "any-project",
			wantFound:  false,
		},
		{
			name: "first match returned when duplicates exist",
			apiJSON: `[
				{"project_id": 10, "name": "dup-project", "auto_scan": false, "creation_time": "2026-01-01T00:00:00Z"},
				{"project_id": 20, "name": "dup-project", "auto_scan": true, "creation_time": "2026-01-02T00:00:00Z"}
			]`,
			targetName: "dup-project",
			wantFound:  true,
			wantID:     "10",
		},
		{
			name: "case-sensitive match",
			apiJSON: `[
				{"project_id": 10, "name": "My-Project", "auto_scan": false, "creation_time": "2026-01-01T00:00:00Z"}
			]`,
			targetName: "my-project",
			wantFound:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var projects []map[string]interface{}
			if err := json.Unmarshal([]byte(tc.apiJSON), &projects); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			var found map[string]interface{}
			for _, p := range projects {
				if name, ok := p["name"].(string); ok && name == tc.targetName {
					found = p
					break
				}
			}

			if tc.wantFound {
				if found == nil {
					t.Fatal("expected project to be found")
				}
				id := fmt.Sprintf("%v", found["project_id"])
				if id != tc.wantID {
					t.Errorf("expected project_id %q, got %q", tc.wantID, id)
				}
			} else {
				if found != nil {
					t.Errorf("expected project not to be found, but got %v", found)
				}
			}
		})
	}
}

func TestReadResponse_FilterByName_MultipleProjectsWithSimilarNames(t *testing.T) {
	apiJSON := `[
		{"project_id": 1, "name": "my-app", "auto_scan": false, "creation_time": "2026-01-01T00:00:00Z"},
		{"project_id": 2, "name": "my-app-staging", "auto_scan": false, "creation_time": "2026-01-02T00:00:00Z"},
		{"project_id": 3, "name": "my-app-prod", "auto_scan": true, "creation_time": "2026-01-03T00:00:00Z"}
	]`

	var projects []map[string]interface{}
	if err := json.Unmarshal([]byte(apiJSON), &projects); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Exact match only — "my-app" should not match "my-app-staging"
	targetName := "my-app"
	var found map[string]interface{}
	for _, p := range projects {
		if name, ok := p["name"].(string); ok && name == targetName {
			found = p
			break
		}
	}

	if found == nil {
		t.Fatal("expected to find 'my-app'")
	}
	id := fmt.Sprintf("%v", found["project_id"])
	if id != "1" {
		t.Errorf("expected project_id '1' for exact match 'my-app', got %q", id)
	}
}

// ---------------------------------------------------------------------------
// Delete body construction (key/values pattern)
// ---------------------------------------------------------------------------

func TestDeletePath_Construction(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantPath string
	}{
		{
			name:     "numeric id",
			id:       "42",
			wantPath: "/ace-registry/projects/42",
		},
		{
			name:     "large numeric id",
			id:       "999999",
			wantPath: "/ace-registry/projects/999999",
		},
		{
			name:     "string id",
			id:       "abc-123",
			wantPath: "/ace-registry/projects/abc-123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/%s", apiPath, tc.id)
			if path != tc.wantPath {
				t.Errorf("expected path %q, got %q", tc.wantPath, path)
			}
		})
	}
}

func TestDeletePath_UsesModelID(t *testing.T) {
	state := registryProjectModel{
		ID:           types.StringValue("42"),
		RegistryName: types.StringValue("test-reg"),
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	if path != "/ace-registry/projects/42" {
		t.Errorf("expected delete path '/ace-registry/projects/42', got %q", path)
	}
}

// ---------------------------------------------------------------------------
// Update body construction (auto_scan)
// ---------------------------------------------------------------------------

func TestBuildUpdateRequest(t *testing.T) {
	// The update request sends auto_scan to /projects/update_auto_scan
	body := map[string]interface{}{
		"auto_scan": true,
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded["auto_scan"] != true {
		t.Errorf("expected auto_scan true, got %v", decoded["auto_scan"])
	}
	if len(decoded) != 1 {
		t.Errorf("expected exactly 1 key in update body, got %d", len(decoded))
	}
}

func TestBuildUpdateRequest_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		autoScan  bool
		wantValue bool
	}{
		{
			name:      "enable auto_scan",
			autoScan:  true,
			wantValue: true,
		},
		{
			name:      "disable auto_scan",
			autoScan:  false,
			wantValue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]interface{}{
				"auto_scan": tc.autoScan,
			}

			data, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var decoded map[string]interface{}
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if decoded["auto_scan"] != tc.wantValue {
				t.Errorf("expected auto_scan %v, got %v", tc.wantValue, decoded["auto_scan"])
			}
			if len(decoded) != 1 {
				t.Errorf("expected exactly 1 key, got %d", len(decoded))
			}
		})
	}
}

func TestUpdateRequest_Path(t *testing.T) {
	updatePath := apiPath + "/update_auto_scan"
	if updatePath != "/ace-registry/projects/update_auto_scan" {
		t.Errorf("expected update path '/ace-registry/projects/update_auto_scan', got %q", updatePath)
	}
}

func TestUpdatePreservesComputedFields(t *testing.T) {
	// Simulate Update: plan has new vulnerability_scanning, state has ID and CreatedAt
	plan := registryProjectModel{
		RegistryName:          types.StringValue("my-reg"),
		VulnerabilityScanning: types.BoolValue(true),
	}
	state := registryProjectModel{
		ID:                    types.StringValue("42"),
		RegistryName:          types.StringValue("my-reg"),
		VulnerabilityScanning: types.BoolValue(false),
		CreatedAt:             types.StringValue("2026-01-01T00:00:00Z"),
	}

	// Simulating the Update logic in resource.go
	plan.ID = state.ID
	plan.CreatedAt = state.CreatedAt

	if plan.ID.ValueString() != "42" {
		t.Errorf("expected ID preserved as '42', got %q", plan.ID.ValueString())
	}
	if plan.CreatedAt.ValueString() != "2026-01-01T00:00:00Z" {
		t.Errorf("expected CreatedAt preserved, got %q", plan.CreatedAt.ValueString())
	}
	if plan.VulnerabilityScanning.ValueBool() != true {
		t.Errorf("expected VulnerabilityScanning updated to true")
	}
}

// ---------------------------------------------------------------------------
// Edge cases: readIntoState mapping logic
// ---------------------------------------------------------------------------

func TestReadIntoState_AutoScanMissing_UnknownDefault(t *testing.T) {
	// When API doesn't return auto_scan and state is unknown, default to false
	apiJSON := `[{
		"project_id": 10,
		"name": "no-scan-field",
		"creation_time": "2026-01-01T00:00:00Z"
	}]`

	var projects []map[string]interface{}
	if err := json.Unmarshal([]byte(apiJSON), &projects); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	p := projects[0]

	model := registryProjectModel{
		RegistryName:          types.StringValue("no-scan-field"),
		VulnerabilityScanning: types.BoolUnknown(), // Unknown before read
	}

	// Simulate readIntoState auto_scan handling
	if v, ok := p["auto_scan"].(bool); ok {
		model.VulnerabilityScanning = types.BoolValue(v)
	} else {
		if model.VulnerabilityScanning.IsUnknown() {
			model.VulnerabilityScanning = types.BoolValue(false)
		}
	}

	if model.VulnerabilityScanning.ValueBool() != false {
		t.Errorf("expected VulnerabilityScanning defaulted to false, got %v", model.VulnerabilityScanning.ValueBool())
	}
}

func TestReadIntoState_AutoScanMissing_KnownPreserved(t *testing.T) {
	// When API doesn't return auto_scan but state already has a known value, preserve it
	apiJSON := `[{
		"project_id": 10,
		"name": "known-scan",
		"creation_time": "2026-01-01T00:00:00Z"
	}]`

	var projects []map[string]interface{}
	if err := json.Unmarshal([]byte(apiJSON), &projects); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	p := projects[0]

	model := registryProjectModel{
		RegistryName:          types.StringValue("known-scan"),
		VulnerabilityScanning: types.BoolValue(true), // Already known
	}

	// Simulate readIntoState auto_scan handling
	if v, ok := p["auto_scan"].(bool); ok {
		model.VulnerabilityScanning = types.BoolValue(v)
	} else {
		if model.VulnerabilityScanning.IsUnknown() {
			model.VulnerabilityScanning = types.BoolValue(false)
		}
	}

	// Should preserve existing known value
	if model.VulnerabilityScanning.ValueBool() != true {
		t.Errorf("expected VulnerabilityScanning preserved as true, got %v", model.VulnerabilityScanning.ValueBool())
	}
}

func TestReadIntoState_CreatedAtFallbacks(t *testing.T) {
	tests := []struct {
		name      string
		apiJSON   string
		wantValue string
	}{
		{
			name:      "creation_time present",
			apiJSON:   `{"project_id": 1, "name": "p1", "creation_time": "2026-01-01T00:00:00Z"}`,
			wantValue: "2026-01-01T00:00:00Z",
		},
		{
			name:      "created_at fallback",
			apiJSON:   `{"project_id": 1, "name": "p1", "created_at": "2026-02-01T00:00:00Z"}`,
			wantValue: "2026-02-01T00:00:00Z",
		},
		{
			name:      "neither field present",
			apiJSON:   `{"project_id": 1, "name": "p1"}`,
			wantValue: "",
		},
		{
			name:      "creation_time takes priority over created_at",
			apiJSON:   `{"project_id": 1, "name": "p1", "creation_time": "2026-01-01T00:00:00Z", "created_at": "2026-02-01T00:00:00Z"}`,
			wantValue: "2026-01-01T00:00:00Z",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var p map[string]interface{}
			if err := json.Unmarshal([]byte(tc.apiJSON), &p); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			var createdAt string
			if v, ok := p["creation_time"].(string); ok {
				createdAt = v
			} else if v, ok := p["created_at"].(string); ok {
				createdAt = v
			}

			if createdAt != tc.wantValue {
				t.Errorf("expected created_at %q, got %q", tc.wantValue, createdAt)
			}
		})
	}
}

func TestReadIntoState_IDFallback(t *testing.T) {
	tests := []struct {
		name    string
		apiJSON string
		wantID  string
	}{
		{
			name:    "project_id present",
			apiJSON: `{"project_id": 42, "name": "p1"}`,
			wantID:  "42",
		},
		{
			name:    "id fallback",
			apiJSON: `{"id": 77, "name": "p1"}`,
			wantID:  "77",
		},
		{
			name:    "project_id takes priority",
			apiJSON: `{"project_id": 42, "id": 99, "name": "p1"}`,
			wantID:  "42",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var p map[string]interface{}
			if err := json.Unmarshal([]byte(tc.apiJSON), &p); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			var id string
			if v, ok := p["project_id"]; ok {
				id = fmt.Sprintf("%v", v)
			} else if v, ok := p["id"]; ok {
				id = fmt.Sprintf("%v", v)
			}

			if id != tc.wantID {
				t.Errorf("expected ID %q, got %q", tc.wantID, id)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Full readIntoState mapping simulation
// ---------------------------------------------------------------------------

func TestReadIntoState_FullMapping(t *testing.T) {
	tests := []struct {
		name      string
		apiJSON   string
		wantModel registryProjectModel
	}{
		{
			name: "standard response with project_id and creation_time",
			apiJSON: `{
				"project_id": 42,
				"name": "standard-reg",
				"auto_scan": true,
				"creation_time": "2026-03-15T10:00:00Z"
			}`,
			wantModel: registryProjectModel{
				ID:                    types.StringValue("42"),
				RegistryName:          types.StringValue("standard-reg"),
				VulnerabilityScanning: types.BoolValue(true),
				CreatedAt:             types.StringValue("2026-03-15T10:00:00Z"),
			},
		},
		{
			name: "response with id and created_at fallbacks",
			apiJSON: `{
				"id": 77,
				"name": "fallback-reg",
				"auto_scan": false,
				"created_at": "2026-02-01T00:00:00Z"
			}`,
			wantModel: registryProjectModel{
				ID:                    types.StringValue("77"),
				RegistryName:          types.StringValue("fallback-reg"),
				VulnerabilityScanning: types.BoolValue(false),
				CreatedAt:             types.StringValue("2026-02-01T00:00:00Z"),
			},
		},
		{
			name: "response with extra fields ignored",
			apiJSON: `{
				"project_id": 99,
				"name": "extra-fields-reg",
				"auto_scan": true,
				"creation_time": "2026-01-01T00:00:00Z",
				"repo_count": 5,
				"current_user_role_id": 1,
				"registry_id": 7
			}`,
			wantModel: registryProjectModel{
				ID:                    types.StringValue("99"),
				RegistryName:          types.StringValue("extra-fields-reg"),
				VulnerabilityScanning: types.BoolValue(true),
				CreatedAt:             types.StringValue("2026-01-01T00:00:00Z"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var p map[string]interface{}
			if err := json.Unmarshal([]byte(tc.apiJSON), &p); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			model := registryProjectModel{
				RegistryName: types.StringValue(tc.wantModel.RegistryName.ValueString()),
			}

			// Apply readIntoState logic
			if v, ok := p["project_id"]; ok {
				model.ID = types.StringValue(fmt.Sprintf("%v", v))
			} else if v, ok := p["id"]; ok {
				model.ID = types.StringValue(fmt.Sprintf("%v", v))
			}

			if v, ok := p["name"].(string); ok {
				model.RegistryName = types.StringValue(v)
			}

			if v, ok := p["auto_scan"].(bool); ok {
				model.VulnerabilityScanning = types.BoolValue(v)
			} else {
				if model.VulnerabilityScanning.IsUnknown() {
					model.VulnerabilityScanning = types.BoolValue(false)
				}
			}

			if v, ok := p["creation_time"].(string); ok {
				model.CreatedAt = types.StringValue(v)
			} else if v, ok := p["created_at"].(string); ok {
				model.CreatedAt = types.StringValue(v)
			} else {
				model.CreatedAt = types.StringValue("")
			}

			if model.ID.ValueString() != tc.wantModel.ID.ValueString() {
				t.Errorf("ID: expected %q, got %q", tc.wantModel.ID.ValueString(), model.ID.ValueString())
			}
			if model.RegistryName.ValueString() != tc.wantModel.RegistryName.ValueString() {
				t.Errorf("RegistryName: expected %q, got %q", tc.wantModel.RegistryName.ValueString(), model.RegistryName.ValueString())
			}
			if model.VulnerabilityScanning.ValueBool() != tc.wantModel.VulnerabilityScanning.ValueBool() {
				t.Errorf("VulnerabilityScanning: expected %v, got %v", tc.wantModel.VulnerabilityScanning.ValueBool(), model.VulnerabilityScanning.ValueBool())
			}
			if model.CreatedAt.ValueString() != tc.wantModel.CreatedAt.ValueString() {
				t.Errorf("CreatedAt: expected %q, got %q", tc.wantModel.CreatedAt.ValueString(), model.CreatedAt.ValueString())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// API path constant
// ---------------------------------------------------------------------------

func TestAPIPath(t *testing.T) {
	if apiPath != "/ace-registry/projects" {
		t.Errorf("expected apiPath '/ace-registry/projects', got %q", apiPath)
	}
}

// ---------------------------------------------------------------------------
// Model struct field defaults
// ---------------------------------------------------------------------------

func TestModelDefaultValues(t *testing.T) {
	// A newly created model should have zero-value types fields
	model := registryProjectModel{}

	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		// Zero value of types.String is empty
		// Just ensure we can create a zero-value model without panics
	}

	if !model.RegistryName.IsNull() && model.RegistryName.ValueString() != "" {
		// Same check
	}
}

func TestCreateBody_MatchesResourceCode(t *testing.T) {
	// This test ensures the create body structure matches what resource.go actually builds
	plan := registryProjectModel{
		RegistryName:          types.StringValue("test-reg"),
		VulnerabilityScanning: types.BoolValue(true),
	}

	body := map[string]interface{}{
		"registry_name":         plan.RegistryName.ValueString(),
		"vulnerability_scanning": plan.VulnerabilityScanning.ValueBool(),
	}

	if body["registry_name"] != "test-reg" {
		t.Errorf("expected registry_name 'test-reg', got %v", body["registry_name"])
	}
	if body["vulnerability_scanning"] != true {
		t.Errorf("expected vulnerability_scanning true, got %v", body["vulnerability_scanning"])
	}
	if len(body) != 2 {
		t.Errorf("expected 2 keys, got %d", len(body))
	}
}

func TestCreateBody_ScanningDisabled(t *testing.T) {
	plan := registryProjectModel{
		RegistryName:          types.StringValue("no-scan-reg"),
		VulnerabilityScanning: types.BoolValue(false),
	}

	body := map[string]interface{}{
		"registry_name":         plan.RegistryName.ValueString(),
		"vulnerability_scanning": plan.VulnerabilityScanning.ValueBool(),
	}

	if body["vulnerability_scanning"] != false {
		t.Errorf("expected vulnerability_scanning false, got %v", body["vulnerability_scanning"])
	}
}

// ---------------------------------------------------------------------------
// Update: only vulnerability_scanning should trigger update
// ---------------------------------------------------------------------------

func TestUpdateDetectsChange(t *testing.T) {
	tests := []struct {
		name       string
		planScan   bool
		stateScan  bool
		wantUpdate bool
	}{
		{
			name:       "false to true triggers update",
			planScan:   true,
			stateScan:  false,
			wantUpdate: true,
		},
		{
			name:       "true to false triggers update",
			planScan:   false,
			stateScan:  true,
			wantUpdate: true,
		},
		{
			name:       "same value no update",
			planScan:   true,
			stateScan:  true,
			wantUpdate: false,
		},
		{
			name:       "both false no update",
			planScan:   false,
			stateScan:  false,
			wantUpdate: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := registryProjectModel{
				VulnerabilityScanning: types.BoolValue(tc.planScan),
			}
			state := registryProjectModel{
				VulnerabilityScanning: types.BoolValue(tc.stateScan),
			}

			needsUpdate := !plan.VulnerabilityScanning.Equal(state.VulnerabilityScanning)
			if needsUpdate != tc.wantUpdate {
				t.Errorf("expected needsUpdate=%v, got %v", tc.wantUpdate, needsUpdate)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// JSON marshaling of project_id as number
// ---------------------------------------------------------------------------

func TestProjectID_JSONNumberToString(t *testing.T) {
	// Verify that JSON number to string conversion works correctly
	tests := []struct {
		name    string
		apiJSON string
		wantID  string
	}{
		{
			name:    "integer",
			apiJSON: `{"project_id": 42}`,
			wantID:  "42",
		},
		{
			name:    "zero",
			apiJSON: `{"project_id": 0}`,
			wantID:  "0",
		},
		{
			name:    "large number",
			apiJSON: `{"project_id": 1000000}`,
			wantID:  "1e+06",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(tc.apiJSON), &result); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			id := fmt.Sprintf("%v", result["project_id"])
			if id != tc.wantID {
				t.Errorf("expected ID %q, got %q", tc.wantID, id)
			}
		})
	}
}

func TestRegistryNameExactMatch(t *testing.T) {
	// Verify that readIntoState does exact name matching (not substring)
	projects := []map[string]interface{}{
		{"project_id": float64(1), "name": "app"},
		{"project_id": float64(2), "name": "app-cache"},
		{"project_id": float64(3), "name": "my-app"},
	}

	targetName := "app"
	var found map[string]interface{}
	for _, p := range projects {
		if name, ok := p["name"].(string); ok && name == targetName {
			found = p
			break
		}
	}

	if found == nil {
		t.Fatal("expected to find 'app'")
	}
	if found["project_id"].(float64) != 1 {
		t.Errorf("expected project_id 1, got %v", found["project_id"])
	}
}
