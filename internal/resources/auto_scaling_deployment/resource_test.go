package auto_scaling_deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &autoScalingDeploymentModel{
		Name:             types.StringValue("test-deployment"),
		Description:      types.StringValue("Test deployment"),
		TemplateID:       types.StringValue("tmpl-123"),
		DesiredCapacity:  types.Int64Value(2),
		MaxCapacity:      types.Int64Value(10),
		NodesScaleCount:  types.Int64Value(1),
		ScalingParameter: types.StringValue("cpu"),
		MinThreshold:     types.Int64Value(30),
		MaxThreshold:     types.Int64Value(80),
		CoolDownTime:     types.Int64Value(300),
		UserEmail: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("admin@example.com"),
		}),
		IsIntegratedLB: types.BoolValue(false),
		LBData:         types.ObjectNull(map[string]attr.Type{}),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Name != "test-deployment" {
		t.Errorf("expected name test-deployment, got %s", body.Name)
	}
	if body.Description != "Test deployment" {
		t.Errorf("expected description 'Test deployment', got %s", body.Description)
	}
	if body.TemplateID != "tmpl-123" {
		t.Errorf("expected template_id tmpl-123, got %s", body.TemplateID)
	}
	if body.DesiredCapacity != 2 {
		t.Errorf("expected desired_capacity 2, got %d", body.DesiredCapacity)
	}
	if body.MaxCapacity != 10 {
		t.Errorf("expected max_capacity 10, got %d", body.MaxCapacity)
	}
	if body.NodesScaleCount != 1 {
		t.Errorf("expected nodes_scale_count 1, got %d", body.NodesScaleCount)
	}
	if body.ScalingParameter != "cpu" {
		t.Errorf("expected scaling_parameter cpu, got %s", body.ScalingParameter)
	}
	if body.MinThreshold != 30 {
		t.Errorf("expected min_threshold 30, got %d", body.MinThreshold)
	}
	if body.MaxThreshold != 80 {
		t.Errorf("expected max_threshold 80, got %d", body.MaxThreshold)
	}
	if body.CoolDownTime != 300 {
		t.Errorf("expected cool_down_time 300, got %d", body.CoolDownTime)
	}
	if len(body.UserEmail) != 1 || body.UserEmail[0] != "admin@example.com" {
		t.Errorf("expected 1 email, got %v", body.UserEmail)
	}
	if body.IsIntegratedLB != false {
		t.Error("expected is_integrated_with_lb to be false")
	}
	if body.LBData != nil {
		t.Error("expected lb_data to be nil when not integrated with LB")
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &autoScalingDeploymentModel{
		Name:             types.StringValue("min-deployment"),
		Description:      types.StringNull(),
		TemplateID:       types.StringValue("tmpl-456"),
		DesiredCapacity:  types.Int64Value(1),
		MaxCapacity:      types.Int64Value(5),
		NodesScaleCount:  types.Int64Value(1),
		ScalingParameter: types.StringValue("ram"),
		MinThreshold:     types.Int64Value(40),
		MaxThreshold:     types.Int64Value(90),
		CoolDownTime:     types.Int64Value(60),
		UserEmail: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("user@example.com"),
		}),
		IsIntegratedLB: types.BoolValue(false),
		LBData:         types.ObjectNull(map[string]attr.Type{}),
	}

	body := buildCreateRequest(context.Background(), plan)

	if body.Description != "" {
		t.Errorf("expected empty description when null, got %s", body.Description)
	}
	if body.ScalingParameter != "ram" {
		t.Errorf("expected scaling_parameter ram, got %s", body.ScalingParameter)
	}
}

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &autoScalingDeploymentModel{
		Name:             types.StringValue("json-deploy"),
		Description:      types.StringValue("JSON test"),
		TemplateID:       types.StringValue("tmpl-1"),
		DesiredCapacity:  types.Int64Value(2),
		MaxCapacity:      types.Int64Value(10),
		NodesScaleCount:  types.Int64Value(1),
		ScalingParameter: types.StringValue("cpu"),
		MinThreshold:     types.Int64Value(30),
		MaxThreshold:     types.Int64Value(80),
		CoolDownTime:     types.Int64Value(120),
		UserEmail: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("test@test.com"),
		}),
		IsIntegratedLB: types.BoolValue(false),
		LBData:         types.ObjectNull(map[string]attr.Type{}),
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

	requiredKeys := []string{"name", "template_id", "desired_capacity", "max_capacity",
		"nodes_scale_count", "scaling_parameter", "min_threshold", "max_threshold",
		"cool_down_time", "user_email", "is_integrated_with_lb"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' to be present", key)
		}
	}

	// lb_data should be omitted when null
	if _, ok := raw["lb_data"]; ok {
		t.Error("expected 'lb_data' to be omitted (omitempty)")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &autoScalingDeploymentModel{
		Description: types.StringNull(),
	}

	apiResp := &deploymentAPIResponse{
		ID:               "dep-abc-123",
		Name:             "prod-deploy",
		Description:      "Production deployment",
		TemplateID:       "tmpl-prod",
		DesiredCapacity:  3,
		MaxCapacity:      15,
		NodesScaleCount:  2,
		ScalingParameter: "cpu",
		MinThreshold:     35,
		MaxThreshold:     85,
		CoolDownTime:     180,
		Status:           "ACTIVE",
		PanelURL:         "https://grafana.example.com/d/123",
		CreatedAt:        "2024-01-01T00:00:00Z",
		UpdatedAt:        "2024-01-02T00:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.ID.ValueString() != "dep-abc-123" {
		t.Errorf("expected ID dep-abc-123, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-deploy" {
		t.Errorf("expected Name prod-deploy, got %s", model.Name.ValueString())
	}
	if model.Description.ValueString() != "Production deployment" {
		t.Errorf("expected Description 'Production deployment', got %s", model.Description.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
	if model.PanelURL.ValueString() != "https://grafana.example.com/d/123" {
		t.Errorf("expected PanelURL, got %s", model.PanelURL.ValueString())
	}
	if model.DesiredCapacity.ValueInt64() != 3 {
		t.Errorf("expected DesiredCapacity 3, got %d", model.DesiredCapacity.ValueInt64())
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	model := &autoScalingDeploymentModel{
		Description: types.StringNull(),
	}

	apiResp := &deploymentAPIResponse{
		ID:     "dep-456",
		Name:   "basic",
		Status: "CREATING",
	}

	mapAPIResponseToState(model, apiResp)

	if !model.Description.IsNull() {
		t.Error("expected Description to remain null when API returns empty string")
	}
	if model.ErrorMessage.ValueString() != "" {
		t.Errorf("expected empty ErrorMessage, got %s", model.ErrorMessage.ValueString())
	}
	if model.PanelURL.ValueString() != "" {
		t.Errorf("expected empty PanelURL, got %s", model.PanelURL.ValueString())
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
	r := &autoScalingDeploymentResource{}
	req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "acecloud_auto_scaling_deployment" {
		t.Errorf("expected type name 'acecloud_auto_scaling_deployment', got %s", resp.TypeName)
	}
}

// --- Configure tests ---

func TestConfigure_NilProviderData(t *testing.T) {
	r := &autoScalingDeploymentResource{}
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Error("expected no error when ProviderData is nil")
	}
}

func TestConfigure_WrongType(t *testing.T) {
	r := &autoScalingDeploymentResource{}
	req := resource.ConfigureRequest{ProviderData: 42}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for wrong provider data type")
	}
}

func TestConfigure_ValidClient(t *testing.T) {
	r := &autoScalingDeploymentResource{}
	c := client.NewClient("http://localhost", "tok", "mumbai", "proj-1")
	req := resource.ConfigureRequest{ProviderData: c}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error, got diagnostics: %v", resp.Diagnostics)
	}
	if r.client != c {
		t.Error("expected client to be set")
	}
}

// --- basePath test ---

func TestBasePath(t *testing.T) {
	if basePath != "/auto-scaling/deployments" {
		t.Errorf("expected basePath '/auto-scaling/deployments', got %s", basePath)
	}
}

// --- Regex validation tests ---

func TestDeploymentNameRegex(t *testing.T) {
	valid := []string{
		"my-deployment",
		"My Deployment 123",
		"test_deploy",
		"a",
		"ABC-123_test",
		"12345",
	}
	invalid := []string{
		"deploy@ment",
		"test!deploy",
		"name#1",
		"test.deploy",
		"test/deploy",
	}

	for _, name := range valid {
		if !deploymentNameRegex.MatchString(name) {
			t.Errorf("expected name '%s' to be valid", name)
		}
	}
	for _, name := range invalid {
		if deploymentNameRegex.MatchString(name) {
			t.Errorf("expected name '%s' to be invalid", name)
		}
	}
}

func TestDescriptionRegex(t *testing.T) {
	valid := []string{
		"Simple description",
		"test-deploy_123",
		"with periods. and commas,",
		"",
		"ABC 123 _ - . ,",
	}
	invalid := []string{
		"has@symbol",
		"has!exclaim",
		"test#hash",
		"with/slash",
		"test<html>",
	}

	for _, desc := range valid {
		if !descriptionRegex.MatchString(desc) {
			t.Errorf("expected description '%s' to be valid", desc)
		}
	}
	for _, desc := range invalid {
		if descriptionRegex.MatchString(desc) {
			t.Errorf("expected description '%s' to be invalid", desc)
		}
	}
}

// --- MapAPIResponseToState edge cases ---

func TestMapAPIResponseToState_ErrorStatus(t *testing.T) {
	model := &autoScalingDeploymentModel{
		Description: types.StringNull(),
	}

	apiResp := &deploymentAPIResponse{
		ID:           "dep-err-1",
		Name:         "failed-deploy",
		Status:       "ERROR",
		ErrorMessage: "Insufficient quota",
	}

	mapAPIResponseToState(model, apiResp)

	if model.Status.ValueString() != "ERROR" {
		t.Errorf("expected Status ERROR, got %s", model.Status.ValueString())
	}
	if model.ErrorMessage.ValueString() != "Insufficient quota" {
		t.Errorf("expected ErrorMessage 'Insufficient quota', got %s", model.ErrorMessage.ValueString())
	}
}

func TestMapAPIResponseToState_DescriptionAlreadySet(t *testing.T) {
	model := &autoScalingDeploymentModel{
		Description: types.StringValue("already-set"),
	}

	apiResp := &deploymentAPIResponse{
		ID:          "dep-1",
		Name:        "test",
		Description: "",
		Status:      "ACTIVE",
	}

	mapAPIResponseToState(model, apiResp)

	// When API returns empty and model has a non-null value, the Description
	// should keep the existing value (because of the "else if model.Description.IsNull()" check)
	if model.Description.ValueString() != "already-set" {
		t.Errorf("expected Description to be preserved as 'already-set', got %s", model.Description.ValueString())
	}
}

func TestMapAPIResponseToState_AllComputedFieldsSet(t *testing.T) {
	model := &autoScalingDeploymentModel{
		Description: types.StringNull(),
	}

	apiResp := &deploymentAPIResponse{
		ID:     "dep-1",
		Name:   "test",
		Status: "",
	}

	mapAPIResponseToState(model, apiResp)

	// All computed fields should be set to known values, even empty strings
	if model.Status.IsNull() || model.Status.IsUnknown() {
		t.Error("expected Status to be a known value, not null/unknown")
	}
	if model.ErrorMessage.IsNull() || model.ErrorMessage.IsUnknown() {
		t.Error("expected ErrorMessage to be a known value, not null/unknown")
	}
	if model.PanelURL.IsNull() || model.PanelURL.IsUnknown() {
		t.Error("expected PanelURL to be a known value, not null/unknown")
	}
	if model.CreatedAt.IsNull() || model.CreatedAt.IsUnknown() {
		t.Error("expected CreatedAt to be a known value, not null/unknown")
	}
	if model.UpdatedAt.IsNull() || model.UpdatedAt.IsUnknown() {
		t.Error("expected UpdatedAt to be a known value, not null/unknown")
	}
}

func TestMapAPIResponseToState_AllFieldsFull(t *testing.T) {
	model := &autoScalingDeploymentModel{
		Description: types.StringNull(),
	}

	apiResp := &deploymentAPIResponse{
		ID:               "dep-full",
		Name:             "full-deployment",
		Description:      "Full test",
		TemplateID:       "tmpl-full",
		DesiredCapacity:  5,
		MaxCapacity:      20,
		NodesScaleCount:  3,
		ScalingParameter: "ram",
		MinThreshold:     40,
		MaxThreshold:     90,
		CoolDownTime:     600,
		Status:           "ACTIVE",
		ErrorMessage:     "",
		PanelURL:         "https://panel.example.com",
		CreatedAt:        "2024-06-01T00:00:00Z",
		UpdatedAt:        "2024-06-02T12:00:00Z",
	}

	mapAPIResponseToState(model, apiResp)

	if model.TemplateID.ValueString() != "tmpl-full" {
		t.Errorf("expected TemplateID 'tmpl-full', got %s", model.TemplateID.ValueString())
	}
	if model.MaxCapacity.ValueInt64() != 20 {
		t.Errorf("expected MaxCapacity 20, got %d", model.MaxCapacity.ValueInt64())
	}
	if model.NodesScaleCount.ValueInt64() != 3 {
		t.Errorf("expected NodesScaleCount 3, got %d", model.NodesScaleCount.ValueInt64())
	}
	if model.ScalingParameter.ValueString() != "ram" {
		t.Errorf("expected ScalingParameter 'ram', got %s", model.ScalingParameter.ValueString())
	}
	if model.MinThreshold.ValueInt64() != 40 {
		t.Errorf("expected MinThreshold 40, got %d", model.MinThreshold.ValueInt64())
	}
	if model.MaxThreshold.ValueInt64() != 90 {
		t.Errorf("expected MaxThreshold 90, got %d", model.MaxThreshold.ValueInt64())
	}
	if model.CoolDownTime.ValueInt64() != 600 {
		t.Errorf("expected CoolDownTime 600, got %d", model.CoolDownTime.ValueInt64())
	}
	if model.CreatedAt.ValueString() != "2024-06-01T00:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", model.CreatedAt.ValueString())
	}
	if model.UpdatedAt.ValueString() != "2024-06-02T12:00:00Z" {
		t.Errorf("expected UpdatedAt, got %s", model.UpdatedAt.ValueString())
	}
}

// --- Build request with multiple emails ---

func TestBuildCreateRequest_MultipleEmails(t *testing.T) {
	plan := &autoScalingDeploymentModel{
		Name:             types.StringValue("multi-email"),
		Description:      types.StringNull(),
		TemplateID:       types.StringValue("tmpl-1"),
		DesiredCapacity:  types.Int64Value(1),
		MaxCapacity:      types.Int64Value(5),
		NodesScaleCount:  types.Int64Value(1),
		ScalingParameter: types.StringValue("cpu"),
		MinThreshold:     types.Int64Value(30),
		MaxThreshold:     types.Int64Value(80),
		CoolDownTime:     types.Int64Value(60),
		UserEmail: types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("admin@example.com"),
			types.StringValue("ops@example.com"),
			types.StringValue("dev@example.com"),
		}),
		IsIntegratedLB: types.BoolValue(false),
		LBData:         types.ObjectNull(map[string]attr.Type{}),
	}

	body := buildCreateRequest(context.Background(), plan)

	if len(body.UserEmail) != 3 {
		t.Fatalf("expected 3 emails, got %d", len(body.UserEmail))
	}
	if body.UserEmail[0] != "admin@example.com" {
		t.Errorf("expected first email admin@example.com, got %s", body.UserEmail[0])
	}
	if body.UserEmail[2] != "dev@example.com" {
		t.Errorf("expected third email dev@example.com, got %s", body.UserEmail[2])
	}
}

func TestBuildCreateRequest_NullEmails(t *testing.T) {
	plan := &autoScalingDeploymentModel{
		Name:             types.StringValue("no-email"),
		Description:      types.StringNull(),
		TemplateID:       types.StringValue("tmpl-1"),
		DesiredCapacity:  types.Int64Value(1),
		MaxCapacity:      types.Int64Value(5),
		NodesScaleCount:  types.Int64Value(1),
		ScalingParameter: types.StringValue("cpu"),
		MinThreshold:     types.Int64Value(30),
		MaxThreshold:     types.Int64Value(80),
		CoolDownTime:     types.Int64Value(60),
		UserEmail:        types.ListNull(types.StringType),
		IsIntegratedLB:   types.BoolValue(false),
		LBData:           types.ObjectNull(map[string]attr.Type{}),
	}

	body := buildCreateRequest(context.Background(), plan)
	if body.UserEmail != nil {
		t.Errorf("expected nil UserEmail when list is null, got %v", body.UserEmail)
	}
}

// --- JSON serialization of API request types ---

func TestDeploymentCreateRequest_JSONSerialization(t *testing.T) {
	req := deploymentCreateRequest{
		Name:             "test",
		Description:      "desc",
		TemplateID:       "tmpl-1",
		DesiredCapacity:  2,
		MaxCapacity:      10,
		NodesScaleCount:  1,
		ScalingParameter: "cpu",
		MinThreshold:     30,
		MaxThreshold:     80,
		CoolDownTime:     120,
		UserEmail:        []string{"a@b.com"},
		IsIntegratedLB:   true,
		LBData: &lbDataCreateRequest{
			LBName:         "my-lb",
			AssignPublicIP: true,
			IsExistingLB:   false,
			Tags:           []string{"env:prod"},
			Listener: &listenerCreateRequest{
				ListenerName:         "http-listener",
				ListenerProtocol:     "HTTP",
				ListenerProtocolPort: 80,
			},
			Pool: &poolCreateRequest{
				PoolProtocol:     "HTTP",
				PoolProtocolPort: 8080,
				LBAlgorithm:      "ROUND_ROBIN",
			},
			HealthMonitor: &healthMonitorCreateRequest{
				MonitorProtocol:   "HTTP",
				MonitorURLPath:    "/health",
				MonitorHTTPMethod: "GET",
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify lb_data is present
	lbData, ok := raw["lb_data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected lb_data to be a JSON object")
	}

	if lbData["lb_name"] != "my-lb" {
		t.Errorf("expected lb_name 'my-lb', got %v", lbData["lb_name"])
	}
	if lbData["assign_public_ip"] != true {
		t.Error("expected assign_public_ip to be true")
	}

	// Verify listener sub-object
	listener, ok := lbData["listener"].(map[string]interface{})
	if !ok {
		t.Fatal("expected listener to be a JSON object")
	}
	if listener["listener_name"] != "http-listener" {
		t.Errorf("expected listener_name 'http-listener', got %v", listener["listener_name"])
	}
	if listener["listener_protocol"] != "HTTP" {
		t.Errorf("expected listener_protocol 'HTTP', got %v", listener["listener_protocol"])
	}
	if listener["listener_protocol_port"].(float64) != 80 {
		t.Errorf("expected listener_protocol_port 80, got %v", listener["listener_protocol_port"])
	}

	// Verify pool sub-object
	pool, ok := lbData["pool"].(map[string]interface{})
	if !ok {
		t.Fatal("expected pool to be a JSON object")
	}
	if pool["lb_algorithm"] != "ROUND_ROBIN" {
		t.Errorf("expected lb_algorithm 'ROUND_ROBIN', got %v", pool["lb_algorithm"])
	}

	// Verify health_monitor sub-object
	hm, ok := lbData["health_monitor"].(map[string]interface{})
	if !ok {
		t.Fatal("expected health_monitor to be a JSON object")
	}
	if hm["monitor_http_method"] != "GET" {
		t.Errorf("expected monitor_http_method 'GET', got %v", hm["monitor_http_method"])
	}
}

func TestDeploymentCreateRequest_JSON_NoLBData(t *testing.T) {
	req := deploymentCreateRequest{
		Name:             "no-lb",
		TemplateID:       "tmpl-1",
		DesiredCapacity:  1,
		MaxCapacity:      5,
		NodesScaleCount:  1,
		ScalingParameter: "cpu",
		MinThreshold:     30,
		MaxThreshold:     80,
		CoolDownTime:     60,
		UserEmail:        []string{"a@b.com"},
		IsIntegratedLB:   false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// lb_data should be omitted
	if _, ok := raw["lb_data"]; ok {
		t.Error("expected lb_data to be omitted when nil")
	}
}

func TestDeploymentCreateRequest_JSON_EmptyDescription(t *testing.T) {
	req := deploymentCreateRequest{
		Name:             "no-desc",
		TemplateID:       "tmpl-1",
		DesiredCapacity:  1,
		MaxCapacity:      5,
		NodesScaleCount:  1,
		ScalingParameter: "cpu",
		MinThreshold:     30,
		MaxThreshold:     80,
		CoolDownTime:     60,
		UserEmail:        []string{"a@b.com"},
		IsIntegratedLB:   false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// description should be omitted with omitempty
	if _, ok := raw["description"]; ok {
		t.Error("expected description to be omitted when empty")
	}
}

// --- API response parsing ---

func TestDeploymentAPIResponse_Deserialization(t *testing.T) {
	jsonStr := `{
		"id": "dep-parse-1",
		"name": "parsed-deploy",
		"description": "Parsed",
		"template_id": "tmpl-p",
		"desired_capacity": 3,
		"max_capacity": 15,
		"nodes_scale_count": 2,
		"scaling_parameter": "cpu",
		"min_threshold": 35,
		"max_threshold": 85,
		"cool_down_time": 300,
		"status": "ACTIVE",
		"error_message": "",
		"panel_url": "https://panel.test",
		"created_at": "2024-03-01T00:00:00Z",
		"updated_at": "2024-03-02T00:00:00Z"
	}`

	var resp deploymentAPIResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "dep-parse-1" {
		t.Errorf("expected ID dep-parse-1, got %s", resp.ID)
	}
	if resp.DesiredCapacity != 3 {
		t.Errorf("expected DesiredCapacity 3, got %d", resp.DesiredCapacity)
	}
	if resp.Status != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", resp.Status)
	}
	if resp.PanelURL != "https://panel.test" {
		t.Errorf("expected PanelURL, got %s", resp.PanelURL)
	}
}

func TestCreateResponseID_Deserialization(t *testing.T) {
	jsonStr := `{"id": "dep-create-1"}`
	var resp createResponseID
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.ID != "dep-create-1" {
		t.Errorf("expected ID 'dep-create-1', got %s", resp.ID)
	}
}

func TestCreateResponseID_EmptyID(t *testing.T) {
	jsonStr := `{"id": ""}`
	var resp createResponseID
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID, got %s", resp.ID)
	}
}

func TestCreateResponseID_MissingID(t *testing.T) {
	jsonStr := `{}`
	var resp createResponseID
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.ID != "" {
		t.Errorf("expected empty ID when missing, got %s", resp.ID)
	}
}

// --- LB data request serialization edge cases ---

func TestLBDataCreateRequest_ExistingLB(t *testing.T) {
	req := lbDataCreateRequest{
		AssignPublicIP: false,
		IsExistingLB:   true,
		LBID:           "lb-existing-1",
		LBVipPortID:    "port-vip-1",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if raw["is_existing_lb"] != true {
		t.Error("expected is_existing_lb to be true")
	}
	if raw["lb_id"] != "lb-existing-1" {
		t.Errorf("expected lb_id, got %v", raw["lb_id"])
	}
	if raw["lb_vip_port_id"] != "port-vip-1" {
		t.Errorf("expected lb_vip_port_id, got %v", raw["lb_vip_port_id"])
	}
	// lb_name should be omitted when using existing LB
	if _, ok := raw["lb_name"]; ok {
		t.Error("expected lb_name to be omitted for existing LB")
	}
}

func TestHealthMonitorCreateRequest_OmitsHTTPMethod(t *testing.T) {
	req := healthMonitorCreateRequest{
		MonitorProtocol: "TCP",
		MonitorURLPath:  "/",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := raw["monitor_http_method"]; ok {
		t.Error("expected monitor_http_method to be omitted when empty")
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
		if r.URL.Path != "/auto-scaling/deployments" {
			t.Errorf("expected path /auto-scaling/deployments, got %s", r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "http-deploy" {
			t.Errorf("expected name 'http-deploy', got %v", body["name"])
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"id":"dep-http-1"}}`)
	})
	defer ts.Close()

	apiResp, err := c.Post(context.Background(), basePath, map[string]interface{}{
		"name":        "http-deploy",
		"template_id": "tmpl-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var created createResponseID
	if err := json.Unmarshal(apiResp.Data, &created); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if created.ID != "dep-http-1" {
		t.Errorf("expected ID 'dep-http-1', got %s", created.ID)
	}
}

func TestReadViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{"id":"dep-read-1","name":"read-deploy","status":"ACTIVE","desired_capacity":2,"max_capacity":10,"nodes_scale_count":1,"scaling_parameter":"cpu","min_threshold":30,"max_threshold":80,"cool_down_time":120,"template_id":"tmpl-1"}}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", basePath, "dep-read-1")
	apiResp, err := c.Get(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var dep deploymentAPIResponse
	if err := json.Unmarshal(apiResp.Data, &dep); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if dep.ID != "dep-read-1" {
		t.Errorf("expected ID 'dep-read-1', got %s", dep.ID)
	}
	if dep.Status != "ACTIVE" {
		t.Errorf("expected Status 'ACTIVE', got %s", dep.Status)
	}
}

func TestDeleteViaHTTP_Success(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/auto-scaling/deployments/dep-del-1" {
			t.Errorf("expected path /auto-scaling/deployments/dep-del-1, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"error":false,"data":{}}`)
	})
	defer ts.Close()

	path := fmt.Sprintf("%s/%s", basePath, "dep-del-1")
	_, err := c.Delete(context.Background(), path, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateViaHTTP_APIError(t *testing.T) {
	ts, c := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, `{"error":true,"message":"Deployment name already exists"}`)
	})
	defer ts.Close()

	_, err := c.Post(context.Background(), basePath, map[string]interface{}{
		"name": "duplicate-deploy",
	})
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

// --- Schema attribute tests ---

func TestSchema_HasRequiredAttributes(t *testing.T) {
	s := autoScalingDeploymentSchema()

	requiredAttrs := []string{
		"name", "template_id", "desired_capacity", "max_capacity",
		"nodes_scale_count", "scaling_parameter", "min_threshold",
		"max_threshold", "cool_down_time", "user_email", "is_integrated_with_lb",
	}

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

func TestSchema_HasComputedAttributes(t *testing.T) {
	s := autoScalingDeploymentSchema()

	computedAttrs := []string{"id", "status", "error_message", "panel_url", "created_at", "updated_at"}

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

func TestSchema_HasOptionalDescription(t *testing.T) {
	s := autoScalingDeploymentSchema()

	descAttr, ok := s.Attributes["description"]
	if !ok {
		t.Fatal("expected 'description' attribute in schema")
	}
	if !descAttr.IsOptional() {
		t.Error("expected 'description' to be optional")
	}
}

func TestSchema_HasLBDataBlock(t *testing.T) {
	s := autoScalingDeploymentSchema()

	if _, ok := s.Blocks["lb_data"]; !ok {
		t.Error("expected 'lb_data' block in schema")
	}
}
