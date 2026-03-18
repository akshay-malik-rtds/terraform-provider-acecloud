package caas_secret

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- buildCreateRequest tests ---

func TestBuildCreateRequest_RegistrySecret(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("my-registry"),
		Type:     types.StringValue("registry"),
		URL:      types.StringValue("https://registry.example.com"),
		Username: types.StringValue("admin"),
		Password: types.StringValue("secret123"),
		Data:     types.MapNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "my-registry" {
		t.Errorf("expected name my-registry, got %s", body.Name)
	}
	if body.Type != "registry" {
		t.Errorf("expected type registry, got %s", body.Type)
	}
	if body.URL != "https://registry.example.com" {
		t.Errorf("expected url, got %s", body.URL)
	}
	if body.Username != "admin" {
		t.Errorf("expected username admin, got %s", body.Username)
	}
	if body.Password != "secret123" {
		t.Errorf("expected password secret123, got %s", body.Password)
	}
	if body.Data != nil {
		t.Errorf("expected nil data for registry type, got %v", body.Data)
	}
}

func TestBuildCreateRequest_GenericSecret(t *testing.T) {
	dataMap, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{
		"API_KEY":     "abc123",
		"DB_PASSWORD": "dbpass",
	})

	plan := &caasSecretModel{
		Name:     types.StringValue("my-generic"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     dataMap,
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "my-generic" {
		t.Errorf("expected name my-generic, got %s", body.Name)
	}
	if body.Type != "generic" {
		t.Errorf("expected type generic, got %s", body.Type)
	}
	if body.URL != "" {
		t.Errorf("expected empty url for generic, got %s", body.URL)
	}
	if len(body.Data) != 2 {
		t.Errorf("expected 2 data entries, got %d", len(body.Data))
	}
	if body.Data["API_KEY"] != "abc123" {
		t.Errorf("expected API_KEY abc123, got %s", body.Data["API_KEY"])
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("json-secret"),
		Type:     types.StringValue("registry"),
		URL:      types.StringValue("https://hub.docker.com"),
		Username: types.StringValue("user"),
		Password: types.StringValue("pass"),
		Data:     types.MapNull(types.StringType),
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

	for _, key := range []string{"name", "type", "url", "username", "password"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' to be present", key)
		}
	}
	// data should be omitted for registry type
	if _, ok := raw["data"]; ok {
		t.Error("expected 'data' to be omitted for registry secret")
	}
}

func TestBuildCreateRequest_AllTypes(t *testing.T) {
	ctx := context.Background()

	// Test registry type
	registryPlan := &caasSecretModel{
		Name:     types.StringValue("reg-secret"),
		Type:     types.StringValue("registry"),
		URL:      types.StringValue("https://docker.io"),
		Username: types.StringValue("user"),
		Password: types.StringValue("pass"),
		Data:     types.MapNull(types.StringType),
	}

	regBody := buildCreateRequest(ctx, registryPlan)
	if regBody.Type != "registry" {
		t.Errorf("expected type registry, got %s", regBody.Type)
	}
	if regBody.URL != "https://docker.io" {
		t.Errorf("expected url, got %s", regBody.URL)
	}
	if regBody.Username != "user" {
		t.Errorf("expected username user, got %s", regBody.Username)
	}
	if regBody.Password != "pass" {
		t.Errorf("expected password pass, got %s", regBody.Password)
	}
	if regBody.Data != nil {
		t.Errorf("expected nil data for registry, got %v", regBody.Data)
	}

	// Test generic type
	dataMap, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"KEY1": "val1",
		"KEY2": "val2",
	})

	genericPlan := &caasSecretModel{
		Name:     types.StringValue("gen-secret"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     dataMap,
	}

	genBody := buildCreateRequest(ctx, genericPlan)
	if genBody.Type != "generic" {
		t.Errorf("expected type generic, got %s", genBody.Type)
	}
	if genBody.URL != "" {
		t.Errorf("expected empty url for generic, got %s", genBody.URL)
	}
	if genBody.Username != "" {
		t.Errorf("expected empty username for generic, got %s", genBody.Username)
	}
	if len(genBody.Data) != 2 {
		t.Errorf("expected 2 data entries, got %d", len(genBody.Data))
	}
}

func TestBuildCreateRequest_GenericEmptyData(t *testing.T) {
	dataMap, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{})

	plan := &caasSecretModel{
		Name:     types.StringValue("empty-data"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     dataMap,
	}

	body := buildCreateRequest(context.Background(), plan)
	if body.Data == nil {
		t.Error("expected non-nil data map for generic with empty data")
	}
	if len(body.Data) != 0 {
		t.Errorf("expected 0 data entries, got %d", len(body.Data))
	}
}

func TestBuildCreateRequest_NullDataForRegistry(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("reg-null-data"),
		Type:     types.StringValue("registry"),
		URL:      types.StringValue("https://docker.io"),
		Username: types.StringValue("admin"),
		Password: types.StringValue("secret"),
		Data:     types.MapNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)
	if body.Data != nil {
		t.Errorf("expected nil data for registry secret, got %v", body.Data)
	}
}

func TestBuildCreateRequest_UnknownFieldsTreatedAsEmpty(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("unknown-fields"),
		Type:     types.StringValue("registry"),
		URL:      types.StringUnknown(),
		Username: types.StringUnknown(),
		Password: types.StringUnknown(),
		Data:     types.MapNull(types.StringType),
	}

	body := buildCreateRequest(context.Background(), plan)
	if body.URL != "" {
		t.Errorf("expected empty url when unknown, got %s", body.URL)
	}
	if body.Username != "" {
		t.Errorf("expected empty username when unknown, got %s", body.Username)
	}
	if body.Password != "" {
		t.Errorf("expected empty password when unknown, got %s", body.Password)
	}
}

func TestBuildCreateRequest_GenericSingleEntry(t *testing.T) {
	dataMap, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{
		"SINGLE_KEY": "single_value",
	})

	plan := &caasSecretModel{
		Name:     types.StringValue("single-entry"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     dataMap,
	}

	body := buildCreateRequest(context.Background(), plan)
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 data entry, got %d", len(body.Data))
	}
	if body.Data["SINGLE_KEY"] != "single_value" {
		t.Errorf("expected SINGLE_KEY single_value, got %s", body.Data["SINGLE_KEY"])
	}
}

func TestBuildCreateRequest_GenericMultipleEntries(t *testing.T) {
	dataMap, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_USER":     "postgres",
		"DB_PASSWORD": "secret",
		"DB_NAME":     "mydb",
	})

	plan := &caasSecretModel{
		Name:     types.StringValue("multi-entry"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     dataMap,
	}

	body := buildCreateRequest(context.Background(), plan)
	if len(body.Data) != 5 {
		t.Fatalf("expected 5 data entries, got %d", len(body.Data))
	}
	if body.Data["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST localhost, got %s", body.Data["DB_HOST"])
	}
	if body.Data["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT 5432, got %s", body.Data["DB_PORT"])
	}
}

func TestBuildCreateRequest_RegistryJSON_OmitEmpty(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("omit-test"),
		Type:     types.StringValue("registry"),
		URL:      types.StringValue("https://r.io"),
		Username: types.StringValue("u"),
		Password: types.StringValue("p"),
		Data:     types.MapNull(types.StringType),
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

	if _, ok := raw["data"]; ok {
		t.Error("expected 'data' omitted from registry JSON")
	}
}

func TestBuildCreateRequest_GenericJSON_OmitEmpty(t *testing.T) {
	dataMap, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{
		"K": "V",
	})

	plan := &caasSecretModel{
		Name:     types.StringValue("gen-omit"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     dataMap,
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

	// url, username, password should be omitted for generic
	for _, key := range []string{"url", "username", "password"} {
		if _, ok := raw[key]; ok {
			t.Errorf("expected '%s' omitted from generic JSON", key)
		}
	}
	// data should be present
	if _, ok := raw["data"]; !ok {
		t.Error("expected 'data' present in generic JSON")
	}
}

// --- buildUpdateRequest tests ---

func TestBuildUpdateRequest_Registry(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("update-reg"),
		Type:     types.StringValue("registry"),
		URL:      types.StringValue("https://new-registry.io"),
		Username: types.StringValue("new-user"),
		Password: types.StringValue("new-pass"),
		Data:     types.MapNull(types.StringType),
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.Name != "update-reg" {
		t.Errorf("expected name update-reg, got %s", body.Name)
	}
	if body.Type != "registry" {
		t.Errorf("expected type registry, got %s", body.Type)
	}
	if body.URL != "https://new-registry.io" {
		t.Errorf("expected url, got %s", body.URL)
	}
	if body.Username != "new-user" {
		t.Errorf("expected username new-user, got %s", body.Username)
	}
	if body.Password != "new-pass" {
		t.Errorf("expected password new-pass, got %s", body.Password)
	}
	if body.Data != nil {
		t.Errorf("expected nil data, got %v", body.Data)
	}
}

func TestBuildUpdateRequest_Generic(t *testing.T) {
	dataMap, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{
		"NEW_KEY": "new_val",
	})

	plan := &caasSecretModel{
		Name:     types.StringValue("update-gen"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     dataMap,
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.Name != "update-gen" {
		t.Errorf("expected name update-gen, got %s", body.Name)
	}
	if body.Type != "generic" {
		t.Errorf("expected type generic, got %s", body.Type)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 data entry, got %d", len(body.Data))
	}
	if body.Data["NEW_KEY"] != "new_val" {
		t.Errorf("expected NEW_KEY new_val, got %s", body.Data["NEW_KEY"])
	}
}

func TestBuildUpdateRequest_NullOptionals(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("null-update"),
		Type:     types.StringValue("generic"),
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	body := buildUpdateRequest(context.Background(), plan)

	if body.URL != "" {
		t.Errorf("expected empty URL, got %s", body.URL)
	}
	if body.Username != "" {
		t.Errorf("expected empty Username, got %s", body.Username)
	}
	if body.Password != "" {
		t.Errorf("expected empty Password, got %s", body.Password)
	}
	if body.Data != nil {
		t.Errorf("expected nil Data, got %v", body.Data)
	}
}

func TestBuildUpdateRequest_MatchesCreateStructure(t *testing.T) {
	plan := &caasSecretModel{
		Name:     types.StringValue("match-test"),
		Type:     types.StringValue("registry"),
		URL:      types.StringValue("https://reg.io"),
		Username: types.StringValue("u"),
		Password: types.StringValue("p"),
		Data:     types.MapNull(types.StringType),
	}

	createBody := buildCreateRequest(context.Background(), plan)
	updateBody := buildUpdateRequest(context.Background(), plan)

	if createBody.Name != updateBody.Name {
		t.Error("name mismatch between create and update")
	}
	if createBody.Type != updateBody.Type {
		t.Error("type mismatch between create and update")
	}
	if createBody.URL != updateBody.URL {
		t.Error("url mismatch between create and update")
	}
	if createBody.Username != updateBody.Username {
		t.Error("username mismatch between create and update")
	}
	if createBody.Password != updateBody.Password {
		t.Error("password mismatch between create and update")
	}
}

// --- mapAPIResponseToState tests ---

func TestMapAPIResponseToState(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringValue("original-pass"),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:      "test-secret",
		Type:      "registry",
		URL:       "https://registry.example.com",
		Username:  "admin",
		Password:  "", // API typically doesn't return password
		Status:    "Active",
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.ID.ValueString() != "test-secret" {
		t.Errorf("expected ID test-secret, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "test-secret" {
		t.Errorf("expected Name test-secret, got %s", model.Name.ValueString())
	}
	if model.URL.ValueString() != "https://registry.example.com" {
		t.Errorf("expected URL, got %s", model.URL.ValueString())
	}
	if model.Username.ValueString() != "admin" {
		t.Errorf("expected Username admin, got %s", model.Username.ValueString())
	}
	// Password should be preserved from plan when API doesn't return it
	if model.Password.ValueString() != "original-pass" {
		t.Errorf("expected Password to be preserved, got %s", model.Password.ValueString())
	}
	if model.Status.ValueString() != "Active" {
		t.Errorf("expected Status Active, got %s", model.Status.ValueString())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:   "basic-secret",
		Type:   "generic",
		Status: "Provisioning",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if !model.URL.IsNull() {
		t.Error("expected URL to remain null")
	}
	if !model.Username.IsNull() {
		t.Error("expected Username to remain null")
	}
}

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	ctx := context.Background()

	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringValue("keep-this"),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:      "full-secret",
		Type:      "registry",
		URL:       "https://registry.example.com",
		Username:  "admin",
		Password:  "",
		Status:    "Active",
		CreatedAt: "2026-01-01T00:00:00Z",
		UpdatedAt: "2026-01-02T00:00:00Z",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.ID.ValueString() != "full-secret" {
		t.Errorf("expected ID full-secret, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "full-secret" {
		t.Errorf("expected Name full-secret, got %s", model.Name.ValueString())
	}
	if model.Type.ValueString() != "registry" {
		t.Errorf("expected Type registry, got %s", model.Type.ValueString())
	}
	if model.URL.ValueString() != "https://registry.example.com" {
		t.Errorf("expected URL, got %s", model.URL.ValueString())
	}
	if model.Username.ValueString() != "admin" {
		t.Errorf("expected Username admin, got %s", model.Username.ValueString())
	}
	// Password should be preserved from model when API returns empty
	if model.Password.ValueString() != "keep-this" {
		t.Errorf("expected Password preserved as 'keep-this', got %s", model.Password.ValueString())
	}
	if model.Status.ValueString() != "Active" {
		t.Errorf("expected Status Active, got %s", model.Status.ValueString())
	}
	if model.CreatedAt.ValueString() != "2026-01-01T00:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2026-01-02T00:00:00Z" {
		t.Errorf("expected UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
	if !model.Data.IsNull() {
		t.Error("expected Data to remain null when API returns empty map")
	}
}

func TestMapAPIResponseToState_IDEqualsName(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:   "my-secret-name",
		Type:   "generic",
		Status: "Active",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	// CaaS secrets: ID == Name
	if model.ID.ValueString() != model.Name.ValueString() {
		t.Errorf("expected ID == Name, got ID=%s Name=%s", model.ID.ValueString(), model.Name.ValueString())
	}
}

func TestMapAPIResponseToState_PasswordReturnedByAPI(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringValue("old-pass"),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:     "pass-test",
		Type:     "registry",
		URL:      "https://r.io",
		Username: "u",
		Password: "new-pass-from-api",
		Status:   "Active",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	// When API actually returns a password, it should be updated
	if model.Password.ValueString() != "new-pass-from-api" {
		t.Errorf("expected Password to be updated from API, got %s", model.Password.ValueString())
	}
}

func TestMapAPIResponseToState_GenericWithData(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:   "data-secret",
		Type:   "generic",
		Status: "Active",
		Data: map[string]string{
			"KEY1": "val1",
			"KEY2": "val2",
		},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.Data.IsNull() {
		t.Fatal("expected Data to be set when API returns data")
	}

	var dataMap map[string]string
	model.Data.ElementsAs(context.Background(), &dataMap, false)
	if len(dataMap) != 2 {
		t.Fatalf("expected 2 data entries, got %d", len(dataMap))
	}
	if dataMap["KEY1"] != "val1" {
		t.Errorf("expected KEY1=val1, got %s", dataMap["KEY1"])
	}
	if dataMap["KEY2"] != "val2" {
		t.Errorf("expected KEY2=val2, got %s", dataMap["KEY2"])
	}
}

func TestMapAPIResponseToState_ComputedFieldsAlwaysSet(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name: "computed-test",
		Type: "generic",
		// Status, CreatedAt, UpdatedAt all empty
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.Status.IsUnknown() {
		t.Error("Status should never be unknown after mapAPIResponseToState")
	}
	if model.CreatedAt.IsUnknown() {
		t.Error("CreatedAt should never be unknown after mapAPIResponseToState")
	}
	if model.UpdatedAt.IsUnknown() {
		t.Error("UpdatedAt should never be unknown after mapAPIResponseToState")
	}

	if model.Status.ValueString() != "" {
		t.Errorf("expected empty Status, got %s", model.Status.ValueString())
	}
	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "" {
		t.Errorf("expected empty UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_URLNotNullOverwritesNull(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:   "url-set",
		Type:   "registry",
		URL:    "https://new-url.io",
		Status: "Active",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.URL.IsNull() {
		t.Error("expected URL to be set when API returns non-empty value")
	}
	if model.URL.ValueString() != "https://new-url.io" {
		t.Errorf("expected URL https://new-url.io, got %s", model.URL.ValueString())
	}
}

func TestMapAPIResponseToState_ProvisioningStatus(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:   "prov-secret",
		Type:   "generic",
		Status: "Provisioning",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.Status.ValueString() != "Provisioning" {
		t.Errorf("expected Status Provisioning, got %s", model.Status.ValueString())
	}
}

func TestMapAPIResponseToState_OutOfSyncStatus(t *testing.T) {
	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	apiResp := &secretAPIResponse{
		Name:   "oos-secret",
		Type:   "generic",
		Status: "OutOfSync",
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.Status.ValueString() != "OutOfSync" {
		t.Errorf("expected Status OutOfSync, got %s", model.Status.ValueString())
	}
}

// --- secretDeleteRequest tests ---

func TestSecretDeleteRequest_JSON(t *testing.T) {
	req := secretDeleteRequest{
		Key:    "name",
		Values: []string{"my-secret"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"key":"name","values":["my-secret"]}`
	if string(data) != expected {
		t.Errorf("expected JSON %s, got %s", expected, string(data))
	}
}

func TestSecretDeleteRequest_Format(t *testing.T) {
	// Verify the CaaS delete uses {key:"name", values:["secret-name"]} format
	// NOT {ids:["name"]} format
	secretName := "my-registry-secret"
	req := secretDeleteRequest{
		Key:    "name",
		Values: []string{secretName},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Must have "key" field with value "name"
	if raw["key"] != "name" {
		t.Errorf("expected key 'name', got %v", raw["key"])
	}

	// Must have "values" field as array
	values, ok := raw["values"].([]interface{})
	if !ok {
		t.Fatalf("expected 'values' to be an array, got %T", raw["values"])
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(values))
	}
	if values[0] != secretName {
		t.Errorf("expected values[0] %q, got %v", secretName, values[0])
	}

	// Must NOT have "ids" field (wrong format)
	if _, ok := raw["ids"]; ok {
		t.Error("expected no 'ids' field -- CaaS uses {key, values} format, not {ids}")
	}
}

func TestSecretDeleteRequest_KeyAlwaysName(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
	}{
		{"simple name", "my-secret"},
		{"hyphenated name", "my-registry-secret"},
		{"short name", "abc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := secretDeleteRequest{
				Key:    "name",
				Values: []string{tc.secretName},
			}
			if req.Key != "name" {
				t.Errorf("expected key 'name', got %s", req.Key)
			}
			if len(req.Values) != 1 || req.Values[0] != tc.secretName {
				t.Errorf("expected values [%s], got %v", tc.secretName, req.Values)
			}
		})
	}
}

// --- secretAPIResponse JSON parsing tests ---

func TestSecretAPIResponse_JSONParsing(t *testing.T) {
	jsonData := `{
		"name": "parsed-secret",
		"type": "registry",
		"url": "https://reg.io",
		"username": "admin",
		"password": "",
		"status": "Active",
		"createdAt": "2026-01-15T10:00:00Z",
		"updatedAt": "2026-01-16T12:00:00Z"
	}`

	var resp secretAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Name != "parsed-secret" {
		t.Errorf("expected Name parsed-secret, got %s", resp.Name)
	}
	if resp.Type != "registry" {
		t.Errorf("expected Type registry, got %s", resp.Type)
	}
	if resp.URL != "https://reg.io" {
		t.Errorf("expected URL, got %s", resp.URL)
	}
	if resp.Status != "Active" {
		t.Errorf("expected Status Active, got %s", resp.Status)
	}
	if resp.CreatedAt != "2026-01-15T10:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", resp.CreatedAt)
	}
}

func TestSecretAPIResponse_JSONParsing_GenericWithData(t *testing.T) {
	jsonData := `{
		"name": "generic-parsed",
		"type": "generic",
		"data": {"API_KEY": "abc123", "SECRET": "xyz"},
		"status": "Active",
		"createdAt": "2026-02-01T00:00:00Z",
		"updatedAt": "2026-02-02T00:00:00Z"
	}`

	var resp secretAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Type != "generic" {
		t.Errorf("expected Type generic, got %s", resp.Type)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 data entries, got %d", len(resp.Data))
	}
	if resp.Data["API_KEY"] != "abc123" {
		t.Errorf("expected API_KEY abc123, got %s", resp.Data["API_KEY"])
	}
}

func TestSecretAPIResponse_JSONParsing_MinimalResponse(t *testing.T) {
	jsonData := `{
		"name": "minimal",
		"type": "generic",
		"status": "Provisioning"
	}`

	var resp secretAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.URL != "" {
		t.Errorf("expected empty URL, got %s", resp.URL)
	}
	if resp.Username != "" {
		t.Errorf("expected empty Username, got %s", resp.Username)
	}
	if resp.Data != nil {
		t.Errorf("expected nil Data, got %v", resp.Data)
	}
}

// --- NewResource test ---

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- End-to-end: parse JSON then map to state ---

func TestMapAPIResponseToState_FromParsedJSON_Registry(t *testing.T) {
	jsonData := `{
		"name": "e2e-registry",
		"type": "registry",
		"url": "https://docker.io",
		"username": "prod-user",
		"password": "",
		"status": "Active",
		"createdAt": "2026-03-01T00:00:00Z",
		"updatedAt": "2026-03-02T00:00:00Z"
	}`

	var apiResp secretAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &apiResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringValue("my-secret-pass"),
		Data:     types.MapNull(types.StringType),
	}

	mapAPIResponseToState(context.Background(), model, &apiResp)

	if model.ID.ValueString() != "e2e-registry" {
		t.Errorf("expected ID e2e-registry, got %s", model.ID.ValueString())
	}
	if model.URL.ValueString() != "https://docker.io" {
		t.Errorf("expected URL, got %s", model.URL.ValueString())
	}
	if model.Username.ValueString() != "prod-user" {
		t.Errorf("expected Username prod-user, got %s", model.Username.ValueString())
	}
	// Password preserved since API returns empty
	if model.Password.ValueString() != "my-secret-pass" {
		t.Errorf("expected Password preserved, got %s", model.Password.ValueString())
	}
}

func TestMapAPIResponseToState_FromParsedJSON_Generic(t *testing.T) {
	jsonData := `{
		"name": "e2e-generic",
		"type": "generic",
		"data": {"TOKEN": "t-123", "WEBHOOK_URL": "https://hooks.io/x"},
		"status": "Active",
		"createdAt": "2026-03-10T00:00:00Z",
		"updatedAt": "2026-03-11T00:00:00Z"
	}`

	var apiResp secretAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &apiResp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	model := &caasSecretModel{
		URL:      types.StringNull(),
		Username: types.StringNull(),
		Password: types.StringNull(),
		Data:     types.MapNull(types.StringType),
	}

	mapAPIResponseToState(context.Background(), model, &apiResp)

	if model.Type.ValueString() != "generic" {
		t.Errorf("expected Type generic, got %s", model.Type.ValueString())
	}
	if model.Data.IsNull() {
		t.Fatal("expected Data to be set")
	}

	var dataMap map[string]string
	model.Data.ElementsAs(context.Background(), &dataMap, false)
	if dataMap["TOKEN"] != "t-123" {
		t.Errorf("expected TOKEN t-123, got %s", dataMap["TOKEN"])
	}
}
