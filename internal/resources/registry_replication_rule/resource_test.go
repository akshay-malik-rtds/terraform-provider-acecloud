package registry_replication_rule

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_Full(t *testing.T) {
	ctx := context.Background()

	srcObj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
		ID:   types.Int64Value(1),
		Name: types.StringValue("docker-hub"),
		URL:  types.StringValue("https://hub.docker.com"),
		Type: types.StringValue("docker-hub"),
	})

	destObj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
		ID:   types.Int64Value(0),
		Name: types.StringValue("local-harbor"),
		URL:  types.StringValue("https://harbor.local"),
		Type: types.StringValue("harbor"),
	})

	trigObj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
		Type: types.StringValue("scheduled"),
		Cron: types.StringValue("0 0 * * *"),
	})

	filterList, _ := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: filterAttrTypes()}, []filterModel{
		{
			Type:       types.StringValue("name"),
			Value:      types.StringValue("library/**"),
			Decoration: types.StringValue("matches"),
		},
	})

	replDel := true
	override := false
	speed := int64(1024)

	plan := &replicationRuleModel{
		Name:              types.StringValue("pull-from-dockerhub"),
		Description:       types.StringValue("Pull images nightly"),
		Enabled:           types.BoolValue(true),
		SrcRegistry:       srcObj,
		DestRegistry:      destObj,
		DestNamespace:     types.StringValue("library"),
		Trigger:           trigObj,
		Filter:            filterList,
		ReplicateDeletion: types.BoolValue(replDel),
		Override:          types.BoolValue(override),
		Speed:             types.Int64Value(speed),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Name != "pull-from-dockerhub" {
		t.Errorf("expected name pull-from-dockerhub, got %s", body.Name)
	}
	if body.Description != "Pull images nightly" {
		t.Errorf("expected description, got %s", body.Description)
	}
	if body.Enabled != true {
		t.Error("expected enabled to be true")
	}
	if body.SrcRegistry == nil {
		t.Fatal("expected src_registry to be set")
	}
	if body.SrcRegistry.ID != 1 {
		t.Errorf("expected src_registry.id 1, got %d", body.SrcRegistry.ID)
	}
	if body.SrcRegistry.Name != "docker-hub" {
		t.Errorf("expected src_registry.name docker-hub, got %s", body.SrcRegistry.Name)
	}
	if body.SrcRegistry.URL != "https://hub.docker.com" {
		t.Errorf("expected src_registry.url, got %s", body.SrcRegistry.URL)
	}
	if body.SrcRegistry.Type != "docker-hub" {
		t.Errorf("expected src_registry.type docker-hub, got %s", body.SrcRegistry.Type)
	}
	if body.DestRegistry == nil {
		t.Fatal("expected dest_registry to be set")
	}
	if body.DestRegistry.Name != "local-harbor" {
		t.Errorf("expected dest_registry.name local-harbor, got %s", body.DestRegistry.Name)
	}
	if body.DestNamespace != "library" {
		t.Errorf("expected dest_namespace library, got %s", body.DestNamespace)
	}
	if body.Trigger == nil {
		t.Fatal("expected trigger to be set")
	}
	if body.Trigger.Type != "scheduled" {
		t.Errorf("expected trigger.type scheduled, got %s", body.Trigger.Type)
	}
	if body.Trigger.TriggerSettings.Cron == "" {
		t.Fatal("expected trigger_settings to be set")
	}
	if body.Trigger.TriggerSettings.Cron != "0 0 * * *" {
		t.Errorf("expected cron '0 0 * * *', got %s", body.Trigger.TriggerSettings.Cron)
	}
	if len(body.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(body.Filters))
	}
	if body.Filters[0].Type != "name" {
		t.Errorf("expected filter type name, got %s", body.Filters[0].Type)
	}
	if body.Filters[0].Value != "library/**" {
		t.Errorf("expected filter value library/**, got %s", body.Filters[0].Value)
	}
	if body.Filters[0].Decoration != "matches" {
		t.Errorf("expected filter decoration matches, got %s", body.Filters[0].Decoration)
	}
	if body.ReplicateDeletion == nil || *body.ReplicateDeletion != true {
		t.Error("expected replicate_deletion to be true")
	}
	if body.Override == nil || *body.Override != false {
		t.Error("expected override to be false")
	}
	if body.Speed == nil || *body.Speed != 1024 {
		t.Errorf("expected speed 1024, got %v", body.Speed)
	}
}

func TestBuildCreateRequest_Minimal(t *testing.T) {
	ctx := context.Background()

	srcObj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
		ID:   types.Int64Value(1),
		Name: types.StringValue("hub"),
		URL:  types.StringValue("https://hub.docker.com"),
		Type: types.StringValue("docker-hub"),
	})

	trigObj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
		Type: types.StringValue("manual"),
		Cron: types.StringNull(),
	})

	plan := &replicationRuleModel{
		Name:              types.StringValue("basic-rule"),
		Description:       types.StringNull(),
		Enabled:           types.BoolValue(false),
		SrcRegistry:       srcObj,
		DestRegistry:      types.ObjectNull(registryAttrTypes()),
		DestNamespace:     types.StringNull(),
		Trigger:           trigObj,
		Filter:            types.ListNull(types.ObjectType{AttrTypes: filterAttrTypes()}),
		ReplicateDeletion: types.BoolNull(),
		Override:          types.BoolNull(),
		Speed:             types.Int64Null(),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Name != "basic-rule" {
		t.Errorf("expected name basic-rule, got %s", body.Name)
	}
	if body.Description != "" {
		t.Errorf("expected empty description, got %s", body.Description)
	}
	if body.Enabled != false {
		t.Error("expected enabled to be false")
	}
	if body.DestRegistry != nil {
		t.Error("expected dest_registry to be nil")
	}
	if body.DestNamespace != "" {
		t.Errorf("expected empty dest_namespace, got %s", body.DestNamespace)
	}
	if body.Trigger == nil {
		t.Fatal("expected trigger to be set")
	}
	if body.Trigger.TriggerSettings.Cron != "" {
		t.Error("expected trigger_settings to be nil for manual trigger with null cron")
	}
	if body.Filters != nil {
		t.Errorf("expected nil filters, got %v", body.Filters)
	}
	if body.ReplicateDeletion != nil {
		t.Error("expected replicate_deletion to be nil")
	}
	if body.Override != nil {
		t.Error("expected override to be nil")
	}
	if body.Speed != nil {
		t.Error("expected speed to be nil")
	}
}

func TestCreateRequest_JSON(t *testing.T) {
	replDel := true
	override := false
	speed := int64(-1)

	body := replicationRuleCreateRequest{
		Name:        "json-test-rule",
		Description: "Test JSON serialization",
		Enabled:     true,
		SrcRegistry: &registryRequest{
			ID:   1,
			Name: "docker-hub",
			URL:  "https://hub.docker.com",
			Type: "docker-hub",
		},
		DestRegistry: &registryRequest{
			ID:   0,
			Name: "harbor",
			URL:  "https://harbor.local",
			Type: "harbor",
		},
		DestNamespace: "prod",
		Trigger: &triggerRequest{
			Type: "scheduled",
			TriggerSettings: triggerSettingsRequest{
				Cron: "0 0 * * *",
			},
		},
		Filters: []filterRequest{
			{Type: "name", Value: "library/**", Decoration: "matches"},
			{Type: "tag", Value: "latest", Decoration: "excludes"},
		},
		ReplicateDeletion: &replDel,
		Override:          &override,
		Speed:             &speed,
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify top-level keys
	requiredKeys := []string{"name", "description", "enabled", "src_registry", "dest_registry", "dest_namespace", "trigger", "filters", "replicate_deletion", "override", "speed"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' to be present", key)
		}
	}

	// Verify src_registry structure
	srcReg := raw["src_registry"].(map[string]interface{})
	if srcReg["id"].(float64) != 1 {
		t.Errorf("expected src_registry.id 1, got %v", srcReg["id"])
	}
	if srcReg["name"].(string) != "docker-hub" {
		t.Errorf("expected src_registry.name docker-hub, got %v", srcReg["name"])
	}

	// Verify trigger structure with trigger_settings
	trig := raw["trigger"].(map[string]interface{})
	if trig["type"].(string) != "scheduled" {
		t.Errorf("expected trigger.type scheduled, got %v", trig["type"])
	}
	trigSettings := trig["trigger_settings"].(map[string]interface{})
	if trigSettings["cron"].(string) != "0 0 * * *" {
		t.Errorf("expected trigger_settings.cron '0 0 * * *', got %v", trigSettings["cron"])
	}

	// Verify filters array
	filters := raw["filters"].([]interface{})
	if len(filters) != 2 {
		t.Fatalf("expected 2 filters, got %d", len(filters))
	}
	f0 := filters[0].(map[string]interface{})
	if f0["type"].(string) != "name" {
		t.Errorf("expected filter[0].type name, got %v", f0["type"])
	}
}

func TestCreateRequest_JSON_OmitsEmpty(t *testing.T) {
	body := replicationRuleCreateRequest{
		Name:    "minimal-rule",
		Enabled: false,
		SrcRegistry: &registryRequest{
			ID:   1,
			Name: "hub",
			URL:  "https://hub.docker.com",
			Type: "docker-hub",
		},
		Trigger: &triggerRequest{
			Type: "manual",
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// These should be omitted due to omitempty
	omittedKeys := []string{"description", "dest_registry", "dest_namespace", "filters", "replicate_deletion", "override", "speed"}
	for _, key := range omittedKeys {
		if _, ok := raw[key]; ok {
			t.Errorf("expected JSON key '%s' to be omitted", key)
		}
	}

	// trigger_settings must always be present (npc-api DTO requires it)
	trig := raw["trigger"].(map[string]interface{})
	if _, ok := trig["trigger_settings"]; !ok {
		t.Error("expected trigger_settings to always be present (required by API)")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		// Initialize nullable fields as null
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:          42,
		Name:        "prod-replication",
		Description: "Replicate prod images",
		Enabled:     true,
		SrcRegistry: &registryAPIResponse{
			ID:   1,
			Name: "docker-hub",
			URL:  "https://hub.docker.com",
			Type: "docker-hub",
		},
		DestRegistry: &registryAPIResponse{
			ID:   0,
			Name: "local-harbor",
			URL:  "https://harbor.local",
			Type: "harbor",
		},
		DestNamespace: "library",
		Trigger: &triggerAPIResponse{
			Type: "scheduled",
			TriggerSettings: &triggerSettingsResponse{
				Cron: "0 0 * * *",
			},
		},
		Filters: []filterAPIResponse{
			{Type: "name", Value: "library/**", Decoration: "matches"},
		},
		ReplicateDeletion: true,
		Override:          false,
		Speed:             1024,
		CreatedAt:         "2024-01-15T10:00:00Z",
		UpdatedAt:         "2024-01-16T12:00:00Z",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.ID.ValueString() != "42" {
		t.Errorf("expected ID '42', got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-replication" {
		t.Errorf("expected Name prod-replication, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Replicate prod images" {
		t.Errorf("expected Description, got %s", model.Description.ValueString())
	}
	if model.Enabled.ValueBool() != true {
		t.Error("expected Enabled to be true")
	}
	if model.DestNamespace.ValueString() != "library" {
		t.Errorf("expected DestNamespace library, got %s", model.DestNamespace.ValueString())
	}
	if model.ReplicateDeletion.ValueBool() != true {
		t.Error("expected ReplicateDeletion to be true")
	}
	if model.Override.ValueBool() != false {
		t.Error("expected Override to be false")
	}
	if model.Speed.ValueInt64() != 1024 {
		t.Errorf("expected Speed 1024, got %d", model.Speed.ValueInt64())
	}
	if model.CreatedAt.ValueString() != "2024-01-15T10:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2024-01-16T12:00:00Z" {
		t.Errorf("expected UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}

	// Verify src_registry object is not null
	if model.SrcRegistry.IsNull() {
		t.Error("expected SrcRegistry to not be null")
	}

	// dest_registry stays as user configured (null if not set) — API value not injected
	// No assertion needed — the model preserves user's config

	// Verify trigger object is not null
	if model.Trigger.IsNull() {
		t.Error("expected Trigger to not be null")
	}

	// Verify filter list is not null
	if model.Filter.IsNull() {
		t.Error("expected Filter to not be null")
	}
}

func TestMapAPIResponseToState_Empty(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:      1,
		Name:    "empty-rule",
		Enabled: false,
		SrcRegistry: &registryAPIResponse{
			ID:   1,
			Name: "hub",
			URL:  "https://hub.docker.com",
			Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{
			Type: "manual",
		},
		ReplicateDeletion: false,
		Override:          false,
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.ID.ValueString() != "1" {
		t.Errorf("expected ID '1', got %s", model.ID.ValueString())
	}
	if !model.Description.IsNull() {
		t.Errorf("expected Description to be null, got %s", model.Description.ValueString())
	}
	if !model.DestNamespace.IsNull() {
		t.Errorf("expected DestNamespace to be null, got %s", model.DestNamespace.ValueString())
	}
	if model.DestRegistry.IsNull() != true {
		t.Error("expected DestRegistry to be null")
	}
	if model.Filter.IsNull() != true {
		t.Error("expected Filter to be null")
	}
	if model.Speed.ValueInt64() != 0 {
		t.Errorf("expected Speed to be 0 (API default), got %d", model.Speed.ValueInt64())
	}
	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "" {
		t.Errorf("expected empty UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
}

func TestBuildUpdateRequest(t *testing.T) {
	ctx := context.Background()

	srcObj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
		ID:   types.Int64Value(1),
		Name: types.StringValue("hub"),
		URL:  types.StringValue("https://hub.docker.com"),
		Type: types.StringValue("docker-hub"),
	})

	trigObj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
		Type: types.StringValue("manual"),
		Cron: types.StringNull(),
	})

	plan := &replicationRuleModel{
		Name:              types.StringValue("updated-rule"),
		Description:       types.StringValue("Updated description"),
		Enabled:           types.BoolValue(true),
		SrcRegistry:       srcObj,
		DestRegistry:      types.ObjectNull(registryAttrTypes()),
		DestNamespace:     types.StringNull(),
		Trigger:           trigObj,
		Filter:            types.ListNull(types.ObjectType{AttrTypes: filterAttrTypes()}),
		ReplicateDeletion: types.BoolNull(),
		Override:          types.BoolNull(),
		Speed:             types.Int64Null(),
	}

	body := buildUpdateRequest(ctx, plan)

	if body.Name != "updated-rule" {
		t.Errorf("expected name updated-rule, got %s", body.Name)
	}
	if body.Description != "Updated description" {
		t.Errorf("expected updated description, got %s", body.Description)
	}
	if body.Enabled != true {
		t.Error("expected enabled to be true")
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- Metadata tests ---

func TestMetadata(t *testing.T) {
	r := &registryReplicationRuleResource{}
	req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "acecloud_registry_replication_rule" {
		t.Errorf("expected type name 'acecloud_registry_replication_rule', got %s", resp.TypeName)
	}
}

// --- Configure tests ---

func TestConfigure_NilProviderData(t *testing.T) {
	r := &registryReplicationRuleResource{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("expected no error when ProviderData is nil")
	}
}

func TestConfigure_WrongType(t *testing.T) {
	r := &registryReplicationRuleResource{}
	req := resource.ConfigureRequest{ProviderData: "wrong"}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestConfigure_ValidClient(t *testing.T) {
	r := &registryReplicationRuleResource{}
	c := client.NewClient("http://localhost", "tok", "mumbai", "proj-1")
	req := resource.ConfigureRequest{ProviderData: c}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error, got diagnostics")
	}
	if r.client != c {
		t.Error("expected client to be set")
	}
}

// --- basePath test ---

func TestBasePath(t *testing.T) {
	if basePath != "/ace-registry/replication-rules/policies" {
		t.Errorf("expected basePath '/ace-registry/replication-rules/policies', got %s", basePath)
	}
}

// --- buildRegistryRequest tests ---

func TestBuildRegistryRequest_Null(t *testing.T) {
	result := buildRegistryRequest(context.Background(), types.ObjectNull(registryAttrTypes()))
	if result != nil {
		t.Error("expected nil for null object")
	}
}

func TestBuildRegistryRequest_Valid(t *testing.T) {
	ctx := context.Background()
	obj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
		ID:   types.Int64Value(5),
		Name: types.StringValue("test-reg"),
		URL:  types.StringValue("https://reg.test"),
		Type: types.StringValue("harbor"),
	})

	result := buildRegistryRequest(ctx, obj)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID != 5 {
		t.Errorf("expected ID 5, got %d", result.ID)
	}
	if result.Name != "test-reg" {
		t.Errorf("expected Name 'test-reg', got %s", result.Name)
	}
	if result.URL != "https://reg.test" {
		t.Errorf("expected URL, got %s", result.URL)
	}
	if result.Type != "harbor" {
		t.Errorf("expected Type 'harbor', got %s", result.Type)
	}
}

func TestBuildRegistryRequest_ZeroID(t *testing.T) {
	ctx := context.Background()
	obj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
		ID:   types.Int64Value(0),
		Name: types.StringValue("local"),
		URL:  types.StringValue("https://local.harbor"),
		Type: types.StringValue("harbor"),
	})

	result := buildRegistryRequest(ctx, obj)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.ID != 0 {
		t.Errorf("expected ID 0, got %d", result.ID)
	}
}

// --- buildTriggerRequest tests ---

func TestBuildTriggerRequest_Null(t *testing.T) {
	result := buildTriggerRequest(context.Background(), types.ObjectNull(triggerAttrTypes()))
	if result != nil {
		t.Error("expected nil for null trigger")
	}
}

func TestBuildTriggerRequest_Manual(t *testing.T) {
	ctx := context.Background()
	obj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
		Type: types.StringValue("manual"),
		Cron: types.StringNull(),
	})

	result := buildTriggerRequest(ctx, obj)
	if result == nil {
		t.Fatal("expected non-nil trigger")
	}
	if result.Type != "manual" {
		t.Errorf("expected Type 'manual', got %s", result.Type)
	}
	if result.TriggerSettings.Cron != "" {
		t.Error("expected empty TriggerSettings.Cron for manual trigger with null cron")
	}
}

func TestBuildTriggerRequest_Scheduled(t *testing.T) {
	ctx := context.Background()
	obj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
		Type: types.StringValue("scheduled"),
		Cron: types.StringValue("0 */6 * * *"),
	})

	result := buildTriggerRequest(ctx, obj)
	if result == nil {
		t.Fatal("expected non-nil trigger")
	}
	if result.Type != "scheduled" {
		t.Errorf("expected Type 'scheduled', got %s", result.Type)
	}
	if result.TriggerSettings.Cron == "" {
		t.Fatal("expected non-empty TriggerSettings.Cron for scheduled trigger")
	}
	if result.TriggerSettings.Cron != "0 */6 * * *" {
		t.Errorf("expected cron '0 */6 * * *', got %s", result.TriggerSettings.Cron)
	}
}

func TestBuildTriggerRequest_EventBased(t *testing.T) {
	ctx := context.Background()
	obj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
		Type: types.StringValue("event_based"),
		Cron: types.StringNull(),
	})

	result := buildTriggerRequest(ctx, obj)
	if result == nil {
		t.Fatal("expected non-nil trigger")
	}
	if result.Type != "event_based" {
		t.Errorf("expected Type 'event_based', got %s", result.Type)
	}
}

// --- buildFilters tests ---

func TestBuildFilters_Null(t *testing.T) {
	result := buildFilters(context.Background(), types.ListNull(types.ObjectType{AttrTypes: filterAttrTypes()}))
	if result != nil {
		t.Error("expected nil for null filter list")
	}
}

func TestBuildFilters_MultipleFilters(t *testing.T) {
	ctx := context.Background()
	filterList, _ := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: filterAttrTypes()}, []filterModel{
		{
			Type:       types.StringValue("name"),
			Value:      types.StringValue("library/**"),
			Decoration: types.StringValue("matches"),
		},
		{
			Type:       types.StringValue("tag"),
			Value:      types.StringValue("latest"),
			Decoration: types.StringNull(),
		},
		{
			Type:       types.StringValue("resource"),
			Value:      types.StringValue("image"),
			Decoration: types.StringValue("excludes"),
		},
	})

	result := buildFilters(ctx, filterList)
	if len(result) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(result))
	}

	if result[0].Type != "name" {
		t.Errorf("expected first filter type 'name', got %s", result[0].Type)
	}
	if result[0].Value != "library/**" {
		t.Errorf("expected first filter value 'library/**', got %s", result[0].Value)
	}
	if result[0].Decoration != "matches" {
		t.Errorf("expected first filter decoration 'matches', got %s", result[0].Decoration)
	}

	// Second filter has null decoration - should be empty string
	if result[1].Decoration != "" {
		t.Errorf("expected empty decoration for second filter, got %s", result[1].Decoration)
	}

	if result[2].Decoration != "excludes" {
		t.Errorf("expected third filter decoration 'excludes', got %s", result[2].Decoration)
	}
}

// --- mapAPIResponseToState edge cases ---

func TestMapAPIResponseToState_NullDestRegistry(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:      10,
		Name:    "no-dest",
		Enabled: true,
		SrcRegistry: &registryAPIResponse{
			ID:   1,
			Name: "hub",
			URL:  "https://hub.docker.com",
			Type: "docker-hub",
		},
		DestRegistry: nil,
		Trigger: &triggerAPIResponse{
			Type: "manual",
		},
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if !model.DestRegistry.IsNull() {
		t.Error("expected DestRegistry to be null when API returns nil")
	}
}

func TestMapAPIResponseToState_IntegerIDConversion(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	tests := []struct {
		id       int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{999999, "999999"},
	}

	for _, tc := range tests {
		apiResp := &replicationRuleAPIResponse{
			ID:   tc.id,
			Name: "test",
			SrcRegistry: &registryAPIResponse{
				ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
			},
			Trigger: &triggerAPIResponse{Type: "manual"},
		}
		mapAPIResponseToState(ctx, model, apiResp)
		if model.ID.ValueString() != tc.expected {
			t.Errorf("for API ID %d, expected string '%s', got '%s'", tc.id, tc.expected, model.ID.ValueString())
		}
	}
}

func TestMapAPIResponseToState_TriggerWithNilSettings(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:   1,
		Name: "manual-rule",
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{
			Type:            "manual",
			TriggerSettings: nil,
		},
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.Trigger.IsNull() {
		t.Error("expected Trigger to be set")
	}
}

func TestMapAPIResponseToState_TriggerWithEmptyCron(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:   1,
		Name: "empty-cron-rule",
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{
			Type: "scheduled",
			TriggerSettings: &triggerSettingsResponse{
				Cron: "",
			},
		},
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.Trigger.IsNull() {
		t.Error("expected Trigger to be set")
	}
}

func TestMapAPIResponseToState_MultipleFilters(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:   5,
		Name: "multi-filter",
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{Type: "manual"},
		Filters: []filterAPIResponse{
			{Type: "name", Value: "library/**", Decoration: "matches"},
			{Type: "tag", Value: "v*", Decoration: ""},
			{Type: "resource", Value: "image", Decoration: "excludes"},
		},
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.Filter.IsNull() {
		t.Fatal("expected Filter to not be null")
	}

	// Verify it parsed correctly by checking element count
	elems := model.Filter.Elements()
	if len(elems) != 3 {
		t.Errorf("expected 3 filter elements, got %d", len(elems))
	}
}

func TestMapAPIResponseToState_EmptyFilters(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:      1,
		Name:    "no-filters",
		Filters: []filterAPIResponse{},
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{Type: "manual"},
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if !model.Filter.IsNull() {
		t.Error("expected Filter to be null when empty")
	}
}

func TestMapAPIResponseToState_SpeedPreservation(t *testing.T) {
	ctx := context.Background()

	// Case 1: Speed is set in API response
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:    1,
		Name:  "speed-test",
		Speed: 2048,
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{Type: "manual"},
	}

	mapAPIResponseToState(ctx, model, apiResp)
	if model.Speed.ValueInt64() != 2048 {
		t.Errorf("expected Speed 2048, got %d", model.Speed.ValueInt64())
	}

	// Case 2: Speed is 0 and model.Speed was null
	model2 := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}
	apiResp2 := &replicationRuleAPIResponse{
		ID:    2,
		Name:  "no-speed",
		Speed: 0,
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{Type: "manual"},
	}
	mapAPIResponseToState(ctx, model2, apiResp2)
	if model2.Speed.ValueInt64() != 0 {
		t.Errorf("expected Speed to be 0 when API returns 0, got %d", model2.Speed.ValueInt64())
	}

	// Case 3: Speed is 0 but model.Speed was previously set
	model3 := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Value(1024),
	}
	apiResp3 := &replicationRuleAPIResponse{
		ID:    3,
		Name:  "prev-speed",
		Speed: 0,
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{Type: "manual"},
	}
	mapAPIResponseToState(ctx, model3, apiResp3)
	// Speed was previously set and API returns 0 => model.Speed.IsNull() is false,
	// so the else-if branch doesn't trigger, keeping the previous value
	if model3.Speed.ValueInt64() != 1024 {
		t.Errorf("expected Speed to be preserved as 1024, got %d", model3.Speed.ValueInt64())
	}
}

func TestMapAPIResponseToState_NegativeSpeed(t *testing.T) {
	ctx := context.Background()
	model := &replicationRuleModel{
		Description:   types.StringNull(),
		DestNamespace: types.StringNull(),
		Speed:         types.Int64Null(),
	}

	apiResp := &replicationRuleAPIResponse{
		ID:    1,
		Name:  "unlimited-speed",
		Speed: -1,
		SrcRegistry: &registryAPIResponse{
			ID: 1, Name: "hub", URL: "https://hub.docker.com", Type: "docker-hub",
		},
		Trigger: &triggerAPIResponse{Type: "manual"},
	}

	mapAPIResponseToState(ctx, model, apiResp)
	if model.Speed.ValueInt64() != -1 {
		t.Errorf("expected Speed -1 (unlimited), got %d", model.Speed.ValueInt64())
	}
}

// --- API response type deserialization ---

func TestReplicationRuleAPIResponse_Deserialization(t *testing.T) {
	jsonStr := `{
		"id": 42,
		"name": "full-rule",
		"description": "Full test",
		"enabled": true,
		"src_registry": {"id": 1, "name": "hub", "url": "https://hub.docker.com", "type": "docker-hub"},
		"dest_registry": {"id": 0, "name": "local", "url": "https://local.harbor", "type": "harbor"},
		"dest_namespace": "library",
		"trigger": {"type": "scheduled", "trigger_settings": {"cron": "0 0 * * *"}},
		"filters": [{"type": "name", "value": "lib/**", "decoration": "matches"}],
		"replicate_deletion": true,
		"override": false,
		"speed": 1024,
		"creation_time": "2024-01-15T10:00:00Z",
		"update_time": "2024-01-16T12:00:00Z"
	}`

	var resp replicationRuleAPIResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != 42 {
		t.Errorf("expected ID 42, got %d", resp.ID)
	}
	if resp.Name != "full-rule" {
		t.Errorf("expected Name 'full-rule', got %s", resp.Name)
	}
	if resp.SrcRegistry == nil {
		t.Fatal("expected SrcRegistry to be set")
	}
	if resp.SrcRegistry.ID != 1 {
		t.Errorf("expected SrcRegistry.ID 1, got %d", resp.SrcRegistry.ID)
	}
	if resp.DestRegistry == nil {
		t.Fatal("expected DestRegistry to be set")
	}
	if resp.Trigger == nil {
		t.Fatal("expected Trigger to be set")
	}
	if resp.Trigger.TriggerSettings == nil {
		t.Fatal("expected TriggerSettings to be set")
	}
	if resp.Trigger.TriggerSettings.Cron != "0 0 * * *" {
		t.Errorf("expected cron, got %s", resp.Trigger.TriggerSettings.Cron)
	}
	if len(resp.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(resp.Filters))
	}
	if resp.ReplicateDeletion != true {
		t.Error("expected ReplicateDeletion true")
	}
	if resp.Speed != 1024 {
		t.Errorf("expected Speed 1024, got %d", resp.Speed)
	}
	if resp.CreatedAt != "2024-01-15T10:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", resp.CreatedAt)
	}
}

func TestReplicationRuleAPIResponse_MinimalDeserialization(t *testing.T) {
	jsonStr := `{"id": 1, "name": "minimal", "enabled": false}`

	var resp replicationRuleAPIResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != 1 {
		t.Errorf("expected ID 1, got %d", resp.ID)
	}
	if resp.SrcRegistry != nil {
		t.Error("expected nil SrcRegistry when missing")
	}
	if resp.Trigger != nil {
		t.Error("expected nil Trigger when missing")
	}
	if resp.Filters != nil {
		t.Error("expected nil Filters when missing")
	}
}

// --- HTTP integration tests ---

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *client.Client) {
	ts := httptest.NewServer(handler)
	c := client.NewClient(ts.URL, "test-token", "mumbai", "proj-1")
	return ts, c
}

func TestCreateViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/ace-registry/replication-rules/policies" {
			t.Errorf("expected path %s, got %s", basePath, r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test-rule" {
			t.Errorf("expected name 'test-rule', got %v", body["name"])
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"id":100,"name":"test-rule","enabled":true}}`)
	})
	defer ts.Close()

	apiResp, err := c.Post(context.Background(), basePath, map[string]interface{}{
		"name":    "test-rule",
		"enabled": true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var rule replicationRuleAPIResponse
	if err := json.Unmarshal(apiResp.Data, &rule); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if rule.ID != 100 {
		t.Errorf("expected ID 100, got %d", rule.ID)
	}
}

func TestReadViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/ace-registry/replication-rules/policies/42" {
			t.Errorf("expected path /ace-registry/replication-rules/policies/42, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"id":42,"name":"read-rule","enabled":true,"replicate_deletion":false,"override":true}}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", basePath, "42")
	apiResp, err := c.Get(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var rule replicationRuleAPIResponse
	if err := json.Unmarshal(apiResp.Data, &rule); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if rule.ID != 42 {
		t.Errorf("expected ID 42, got %d", rule.ID)
	}
	if rule.Override != true {
		t.Error("expected Override true")
	}
}

func TestUpdateViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{}}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", basePath, "42")
	_, err := c.Put(context.Background(), path, map[string]interface{}{
		"name":    "updated-rule",
		"enabled": false,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/ace-registry/replication-rules/policies/42" {
			t.Errorf("expected path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{}}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", basePath, "42")
	_, err := c.Delete(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateViaHTTP_APIError(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"error":true,"message":"Rule name already exists"}`)
	})
	defer ts.Close()

	_, err := c.Post(context.Background(), basePath, map[string]interface{}{
		"name": "dup-rule",
	})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

// --- buildUpdateRequest delegates to buildCreateRequest ---

func TestBuildUpdateRequest_DelegatesToCreate(t *testing.T) {
	ctx := context.Background()

	srcObj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
		ID:   types.Int64Value(1),
		Name: types.StringValue("hub"),
		URL:  types.StringValue("https://hub.docker.com"),
		Type: types.StringValue("docker-hub"),
	})

	trigObj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
		Type: types.StringValue("manual"),
		Cron: types.StringNull(),
	})

	plan := &replicationRuleModel{
		Name:              types.StringValue("delegate-test"),
		Description:       types.StringValue("desc"),
		Enabled:           types.BoolValue(true),
		SrcRegistry:       srcObj,
		DestRegistry:      types.ObjectNull(registryAttrTypes()),
		DestNamespace:     types.StringNull(),
		Trigger:           trigObj,
		Filter:            types.ListNull(types.ObjectType{AttrTypes: filterAttrTypes()}),
		ReplicateDeletion: types.BoolNull(),
		Override:          types.BoolNull(),
		Speed:             types.Int64Null(),
	}

	createBody := buildCreateRequest(ctx, plan)
	updateBody := buildUpdateRequest(ctx, plan)

	// They should produce identical results
	createJSON, _ := json.Marshal(createBody)
	updateJSON, _ := json.Marshal(updateBody)
	if string(createJSON) != string(updateJSON) {
		t.Errorf("expected buildUpdateRequest to produce same result as buildCreateRequest")
	}
}

// --- Schema tests ---

func TestSchema_HasRequiredAttributes(t *testing.T) {
	s := replicationRuleSchema()

	requiredAttrs := []string{"name", "enabled"}
	for _, name := range requiredAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("expected attribute '%s' in schema", name)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("expected attribute '%s' to be required", name)
		}
	}
}

func TestSchema_HasOptionalAttributes(t *testing.T) {
	s := replicationRuleSchema()

	optionalAttrs := []string{"description", "dest_namespace", "replicate_deletion", "override", "speed"}
	for _, name := range optionalAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("expected attribute '%s' in schema", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected attribute '%s' to be optional", name)
		}
	}
}

func TestSchema_HasComputedAttributes(t *testing.T) {
	s := replicationRuleSchema()

	computedAttrs := []string{"id", "created_at", "updated_at"}
	for _, name := range computedAttrs {
		attr, ok := s.Attributes[name]
		if !ok {
			t.Errorf("expected attribute '%s' in schema", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("expected attribute '%s' to be computed", name)
		}
	}
}

func TestSchema_HasExpectedBlocks(t *testing.T) {
	s := replicationRuleSchema()

	expectedBlocks := []string{"src_registry", "dest_registry", "trigger", "filter"}
	for _, name := range expectedBlocks {
		if _, ok := s.Blocks[name]; !ok {
			t.Errorf("expected block '%s' in schema", name)
		}
	}
}

// --- Attr type helpers ---

func TestRegistryAttrTypes(t *testing.T) {
	types := registryAttrTypes()
	expected := []string{"id", "name", "url", "type"}
	for _, key := range expected {
		if _, ok := types[key]; !ok {
			t.Errorf("expected key '%s' in registryAttrTypes", key)
		}
	}
	if len(types) != 4 {
		t.Errorf("expected 4 registry attr types, got %d", len(types))
	}
}

func TestTriggerAttrTypes(t *testing.T) {
	types := triggerAttrTypes()
	expected := []string{"type", "cron"}
	for _, key := range expected {
		if _, ok := types[key]; !ok {
			t.Errorf("expected key '%s' in triggerAttrTypes", key)
		}
	}
	if len(types) != 2 {
		t.Errorf("expected 2 trigger attr types, got %d", len(types))
	}
}

func TestFilterAttrTypes(t *testing.T) {
	types := filterAttrTypes()
	expected := []string{"type", "value", "decoration"}
	for _, key := range expected {
		if _, ok := types[key]; !ok {
			t.Errorf("expected key '%s' in filterAttrTypes", key)
		}
	}
	if len(types) != 3 {
		t.Errorf("expected 3 filter attr types, got %d", len(types))
	}
}
