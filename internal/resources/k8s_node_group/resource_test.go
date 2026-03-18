package k8s_node_group

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestBuildCreateRequest(t *testing.T) {
	ctx := context.Background()

	labelsMap, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"env": "prod",
	})
	annotationsMap, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"team": "platform",
	})

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("cluster-uuid-123"),
		SecGroupID:  types.StringValue("sg-uuid-456"),
		Name:        types.StringValue("pool2"),
		Quantity:    types.Int64Value(3),
		FlavorID:    types.StringValue("flavor-uuid-789"),
		Volume:      types.StringValue("50"),
		Labels:      labelsMap,
		Annotations: annotationsMap,
		MinNode:     types.Int64Value(1),
		MaxNode:     types.Int64Value(5),
	}

	body := buildCreateRequest(ctx, plan)

	if body.ClusterID != "cluster-uuid-123" {
		t.Errorf("expected clusterId cluster-uuid-123, got %s", body.ClusterID)
	}
	if body.SecGroupID != "sg-uuid-456" {
		t.Errorf("expected secGroupId sg-uuid-456, got %s", body.SecGroupID)
	}
	if body.Name != "pool2" {
		t.Errorf("expected name pool2, got %s", body.Name)
	}
	if body.Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", body.Quantity)
	}
	if body.FlavorID != "flavor-uuid-789" {
		t.Errorf("expected flavorId flavor-uuid-789, got %s", body.FlavorID)
	}
	if body.Volume != "50" {
		t.Errorf("expected volume 50, got %s", body.Volume)
	}
	if !body.CPU {
		t.Error("expected cpu to be true")
	}
	if body.GPU {
		t.Error("expected gpu to be false")
	}
	if body.Labels["env"] != "prod" {
		t.Errorf("expected label env=prod, got %v", body.Labels)
	}
	if body.Annotations["team"] != "platform" {
		t.Errorf("expected annotation team=platform, got %v", body.Annotations)
	}
	if body.MinNode == nil || *body.MinNode != 1 {
		t.Errorf("expected minNode 1, got %v", body.MinNode)
	}
	if body.MaxNode == nil || *body.MaxNode != 5 {
		t.Errorf("expected maxNode 5, got %v", body.MaxNode)
	}

	// Verify JSON field names match API expectations
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	requiredKeys := []string{"clusterId", "secGroupId", "name", "quantity", "volume", "flavorId", "cpu", "gpu", "labels", "annotations"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' to be present", key)
		}
	}
}

func TestBuildScaleRequest_ScaleUp(t *testing.T) {
	plan := &k8sNodeGroupModel{
		Quantity: types.Int64Value(5),
	}
	state := &k8sNodeGroupModel{
		Quantity:  types.Int64Value(3),
		Name:      types.StringValue("pool2"),
		ClusterID: types.StringValue("cluster-uuid-123"),
	}

	body := buildScaleRequest(plan, state)

	if body.Updates.Type != "add" {
		t.Errorf("expected scale type 'add', got %s", body.Updates.Type)
	}
	if body.Updates.Quantity != 2 {
		t.Errorf("expected scale quantity 2, got %d", body.Updates.Quantity)
	}
	if body.NodeName != "pool2" {
		t.Errorf("expected nodeName pool2, got %s", body.NodeName)
	}
	if body.ClusterID != "cluster-uuid-123" {
		t.Errorf("expected clusterId cluster-uuid-123, got %s", body.ClusterID)
	}

	// Verify JSON structure
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal scale request: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	updates, ok := raw["updates"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'updates' to be an object")
	}
	if updates["type"] != "add" {
		t.Errorf("expected updates.type 'add', got %v", updates["type"])
	}
	if updates["quantity"].(float64) != 2 {
		t.Errorf("expected updates.quantity 2, got %v", updates["quantity"])
	}
}

func TestBuildScaleRequest_ScaleDown(t *testing.T) {
	plan := &k8sNodeGroupModel{
		Quantity: types.Int64Value(1),
	}
	state := &k8sNodeGroupModel{
		Quantity:  types.Int64Value(4),
		Name:      types.StringValue("pool-web"),
		ClusterID: types.StringValue("cluster-uuid-789"),
	}

	body := buildScaleRequest(plan, state)

	if body.Updates.Type != "remove" {
		t.Errorf("expected scale type 'remove', got %s", body.Updates.Type)
	}
	if body.Updates.Quantity != 3 {
		t.Errorf("expected scale quantity 3, got %d", body.Updates.Quantity)
	}
	if body.NodeName != "pool-web" {
		t.Errorf("expected nodeName pool-web, got %s", body.NodeName)
	}
	if body.ClusterID != "cluster-uuid-789" {
		t.Errorf("expected clusterId cluster-uuid-789, got %s", body.ClusterID)
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	ctx := context.Background()

	model := &k8sNodeGroupModel{
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	apiResp := &nodeGroupAPIResponse{
		ID:         "ng-abc-123",
		ClusterID:  "cluster-uuid-123",
		SecGroupID: "sg-uuid-456",
		Name:       "pool2",
		Quantity:   3,
		FlavorID:   "flavor-uuid-789",
		Volume:     "50",
		Labels: map[string]string{
			"env": "prod",
		},
		Annotations: map[string]string{
			"team": "platform",
		},
		MinNode: 1,
		MaxNode: 5,
		State:   "ACTIVE",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.ID.ValueString() != "ng-abc-123" {
		t.Errorf("expected ID ng-abc-123, got %s", model.ID.ValueString())
	}
	if model.ClusterID.ValueString() != "cluster-uuid-123" {
		t.Errorf("expected ClusterID cluster-uuid-123, got %s", model.ClusterID.ValueString())
	}
	if model.SecGroupID.ValueString() != "sg-uuid-456" {
		t.Errorf("expected SecGroupID sg-uuid-456, got %s", model.SecGroupID.ValueString())
	}
	if model.Name.ValueString() != "pool2" {
		t.Errorf("expected Name pool2, got %s", model.Name.ValueString())
	}
	if model.Quantity.ValueInt64() != 3 {
		t.Errorf("expected Quantity 3, got %d", model.Quantity.ValueInt64())
	}
	if model.FlavorID.ValueString() != "flavor-uuid-789" {
		t.Errorf("expected FlavorID flavor-uuid-789, got %s", model.FlavorID.ValueString())
	}
	if model.Volume.ValueString() != "50" {
		t.Errorf("expected Volume 50, got %s", model.Volume.ValueString())
	}
	if model.State.ValueString() != "ACTIVE" {
		t.Errorf("expected State ACTIVE, got %s", model.State.ValueString())
	}
	if model.MinNode.ValueInt64() != 1 {
		t.Errorf("expected MinNode 1, got %d", model.MinNode.ValueInt64())
	}
	if model.MaxNode.ValueInt64() != 5 {
		t.Errorf("expected MaxNode 5, got %d", model.MaxNode.ValueInt64())
	}

	// Verify labels
	labelsMap := make(map[string]string)
	model.Labels.ElementsAs(ctx, &labelsMap, false)
	if labelsMap["env"] != "prod" {
		t.Errorf("expected label env=prod, got %v", labelsMap)
	}

	// Verify annotations
	annotationsMap := make(map[string]string)
	model.Annotations.ElementsAs(ctx, &annotationsMap, false)
	if annotationsMap["team"] != "platform" {
		t.Errorf("expected annotation team=platform, got %v", annotationsMap)
	}
}

func TestMapAPIResponseToState_EmptyOptionals(t *testing.T) {
	ctx := context.Background()

	model := &k8sNodeGroupModel{
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	apiResp := &nodeGroupAPIResponse{
		ID:        "ng-456",
		ClusterID: "cluster-1",
		Name:      "basic-pool",
		Quantity:  1,
		FlavorID:  "flv-1",
		Volume:    "20",
		State:     "ACTIVE",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if !model.Labels.IsNull() {
		t.Error("expected Labels to remain null when API returns empty map")
	}
	if !model.Annotations.IsNull() {
		t.Error("expected Annotations to remain null when API returns empty map")
	}
	if !model.MinNode.IsNull() {
		t.Error("expected MinNode to remain null when API returns 0")
	}
	if !model.MaxNode.IsNull() {
		t.Error("expected MaxNode to remain null when API returns 0")
	}
}

func TestMapAPIResponseToState_WithLabelsFromNull(t *testing.T) {
	ctx := context.Background()

	// Simulate a model where labels were set in plan but not yet mapped
	labelsMap, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"key": "val"})
	model := &k8sNodeGroupModel{
		Labels:      labelsMap,
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	apiResp := &nodeGroupAPIResponse{
		ID:        "ng-789",
		ClusterID: "cluster-2",
		Name:      "labeled-pool",
		Quantity:  2,
		FlavorID:  "flv-2",
		Volume:    "100",
		Labels:    map[string]string{"newkey": "newval"},
		State:     "RUNNING",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	resultLabels := make(map[string]string)
	model.Labels.ElementsAs(ctx, &resultLabels, false)
	if resultLabels["newkey"] != "newval" {
		t.Errorf("expected label newkey=newval, got %v", resultLabels)
	}
}

func TestBuildScaleRequest_ZeroDelta(t *testing.T) {
	plan := &k8sNodeGroupModel{
		Quantity: types.Int64Value(3),
	}
	state := &k8sNodeGroupModel{
		Quantity:  types.Int64Value(3),
		Name:      types.StringValue("pool-same"),
		ClusterID: types.StringValue("cluster-same"),
	}

	body := buildScaleRequest(plan, state)

	// Zero delta: type should be "add" (newQty >= oldQty) and quantity should be 0
	if body.Updates.Type != "add" {
		t.Errorf("expected scale type 'add' for zero delta, got %s", body.Updates.Type)
	}
	if body.Updates.Quantity != 0 {
		t.Errorf("expected scale quantity 0 for zero delta, got %d", body.Updates.Quantity)
	}
	if body.NodeName != "pool-same" {
		t.Errorf("expected nodeName pool-same, got %s", body.NodeName)
	}
	if body.ClusterID != "cluster-same" {
		t.Errorf("expected clusterId cluster-same, got %s", body.ClusterID)
	}

	// Verify JSON structure
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	updates := raw["updates"].(map[string]interface{})
	if updates["quantity"].(float64) != 0 {
		t.Errorf("expected updates.quantity 0 in JSON, got %v", updates["quantity"])
	}
}

func TestBuildCreateRequest_AllFields(t *testing.T) {
	ctx := context.Background()

	labelsMap, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"env":  "production",
		"team": "platform",
		"tier": "compute",
	})
	annotationsMap, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"cost-center": "eng-123",
	})
	minNode := int64(2)
	maxNode := int64(8)

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("cluster-full-123"),
		SecGroupID:  types.StringValue("sg-full-456"),
		Name:        types.StringValue("full-pool"),
		Quantity:    types.Int64Value(4),
		FlavorID:    types.StringValue("flavor-full-789"),
		Volume:      types.StringValue("100"),
		Labels:      labelsMap,
		Annotations: annotationsMap,
		MinNode:     types.Int64Value(minNode),
		MaxNode:     types.Int64Value(maxNode),
	}

	body := buildCreateRequest(ctx, plan)

	if body.ClusterID != "cluster-full-123" {
		t.Errorf("expected clusterId cluster-full-123, got %s", body.ClusterID)
	}
	if body.SecGroupID != "sg-full-456" {
		t.Errorf("expected secGroupId sg-full-456, got %s", body.SecGroupID)
	}
	if body.Name != "full-pool" {
		t.Errorf("expected name full-pool, got %s", body.Name)
	}
	if body.Quantity != 4 {
		t.Errorf("expected quantity 4, got %d", body.Quantity)
	}
	if body.FlavorID != "flavor-full-789" {
		t.Errorf("expected flavorId flavor-full-789, got %s", body.FlavorID)
	}
	if body.Volume != "100" {
		t.Errorf("expected volume 100, got %s", body.Volume)
	}
	if !body.CPU {
		t.Error("expected cpu to be true")
	}
	if body.GPU {
		t.Error("expected gpu to be false")
	}

	// Verify labels
	if len(body.Labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(body.Labels))
	}
	if body.Labels["env"] != "production" {
		t.Errorf("expected label env=production, got %v", body.Labels)
	}
	if body.Labels["team"] != "platform" {
		t.Errorf("expected label team=platform, got %v", body.Labels)
	}
	if body.Labels["tier"] != "compute" {
		t.Errorf("expected label tier=compute, got %v", body.Labels)
	}

	// Verify annotations
	if len(body.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(body.Annotations))
	}
	if body.Annotations["cost-center"] != "eng-123" {
		t.Errorf("expected annotation cost-center=eng-123, got %v", body.Annotations)
	}

	// Verify min/max node
	if body.MinNode == nil || *body.MinNode != 2 {
		t.Errorf("expected minNode 2, got %v", body.MinNode)
	}
	if body.MaxNode == nil || *body.MaxNode != 8 {
		t.Errorf("expected maxNode 8, got %v", body.MaxNode)
	}
}

func TestMapAPIResponseToState_WithLabels(t *testing.T) {
	ctx := context.Background()

	model := &k8sNodeGroupModel{
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	apiResp := &nodeGroupAPIResponse{
		ID:         "ng-labels-001",
		ClusterID:  "cluster-labels",
		SecGroupID: "sg-labels",
		Name:       "labeled-pool",
		Quantity:   3,
		FlavorID:   "flv-labels",
		Volume:     "50",
		Labels: map[string]string{
			"env":  "staging",
			"team": "ops",
			"zone": "az-1",
		},
		Annotations: map[string]string{
			"description": "Staging node pool",
		},
		MinNode: 1,
		MaxNode: 5,
		State:   "ACTIVE",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	// Verify labels are mapped correctly
	resultLabels := make(map[string]string)
	model.Labels.ElementsAs(ctx, &resultLabels, false)
	if len(resultLabels) != 3 {
		t.Fatalf("expected 3 labels, got %d", len(resultLabels))
	}
	if resultLabels["env"] != "staging" {
		t.Errorf("expected label env=staging, got %v", resultLabels)
	}
	if resultLabels["team"] != "ops" {
		t.Errorf("expected label team=ops, got %v", resultLabels)
	}
	if resultLabels["zone"] != "az-1" {
		t.Errorf("expected label zone=az-1, got %v", resultLabels)
	}

	// Verify annotations mapped
	resultAnnotations := make(map[string]string)
	model.Annotations.ElementsAs(ctx, &resultAnnotations, false)
	if len(resultAnnotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(resultAnnotations))
	}
	if resultAnnotations["description"] != "Staging node pool" {
		t.Errorf("expected annotation description='Staging node pool', got %v", resultAnnotations)
	}

	// Verify min/max node
	if model.MinNode.ValueInt64() != 1 {
		t.Errorf("expected MinNode 1, got %d", model.MinNode.ValueInt64())
	}
	if model.MaxNode.ValueInt64() != 5 {
		t.Errorf("expected MaxNode 5, got %d", model.MaxNode.ValueInt64())
	}
	if model.State.ValueString() != "ACTIVE" {
		t.Errorf("expected State ACTIVE, got %s", model.State.ValueString())
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- buildCreateRequest additional tests ---

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	ctx := context.Background()

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("cluster-min-1"),
		SecGroupID:  types.StringValue("sg-min-1"),
		Name:        types.StringValue("minimal"),
		Quantity:    types.Int64Value(1),
		FlavorID:    types.StringValue("flv-min-1"),
		Volume:      types.StringValue("20"),
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	body := buildCreateRequest(ctx, plan)

	if body.ClusterID != "cluster-min-1" {
		t.Errorf("expected clusterId cluster-min-1, got %s", body.ClusterID)
	}
	if body.Name != "minimal" {
		t.Errorf("expected name minimal, got %s", body.Name)
	}
	if body.Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", body.Quantity)
	}
	// Labels and Annotations should be empty maps, not nil
	if body.Labels == nil {
		t.Fatal("expected Labels to be empty map, got nil")
	}
	if len(body.Labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(body.Labels))
	}
	if body.Annotations == nil {
		t.Fatal("expected Annotations to be empty map, got nil")
	}
	if len(body.Annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(body.Annotations))
	}
	// MinNode and MaxNode should be nil
	if body.MinNode != nil {
		t.Errorf("expected MinNode nil, got %v", body.MinNode)
	}
	if body.MaxNode != nil {
		t.Errorf("expected MaxNode nil, got %v", body.MaxNode)
	}
}

func TestBuildCreateRequest_NullLabels(t *testing.T) {
	ctx := context.Background()

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("cluster-nl-1"),
		SecGroupID:  types.StringValue("sg-nl-1"),
		Name:        types.StringValue("null-labels"),
		Quantity:    types.Int64Value(2),
		FlavorID:    types.StringValue("flv-nl-1"),
		Volume:      types.StringValue("30"),
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	body := buildCreateRequest(ctx, plan)

	// When Labels is null in the model, the body should still have an empty map (not nil)
	if body.Labels == nil {
		t.Fatal("expected Labels to be empty map when model.Labels is null, got nil")
	}
	if len(body.Labels) != 0 {
		t.Errorf("expected 0 labels when model.Labels is null, got %d", len(body.Labels))
	}
}

func TestBuildCreateRequest_EmptyLabels(t *testing.T) {
	ctx := context.Background()

	emptyLabels, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{})
	emptyAnnotations, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{})

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("cluster-el-1"),
		SecGroupID:  types.StringValue("sg-el-1"),
		Name:        types.StringValue("empty-labels"),
		Quantity:    types.Int64Value(1),
		FlavorID:    types.StringValue("flv-el-1"),
		Volume:      types.StringValue("20"),
		Labels:      emptyLabels,
		Annotations: emptyAnnotations,
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Labels == nil {
		t.Fatal("expected Labels to be empty map, got nil")
	}
	if len(body.Labels) != 0 {
		t.Errorf("expected 0 labels, got %d", len(body.Labels))
	}
	if body.Annotations == nil {
		t.Fatal("expected Annotations to be empty map, got nil")
	}
	if len(body.Annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(body.Annotations))
	}
}

func TestBuildCreateRequest_JSONFieldNames(t *testing.T) {
	ctx := context.Background()

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("c1"),
		SecGroupID:  types.StringValue("s1"),
		Name:        types.StringValue("json-test"),
		Quantity:    types.Int64Value(1),
		FlavorID:    types.StringValue("f1"),
		Volume:      types.StringValue("10"),
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Value(1),
		MaxNode:     types.Int64Value(3),
	}

	body := buildCreateRequest(ctx, plan)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify camelCase JSON keys
	camelCaseKeys := []string{"clusterId", "secGroupId", "flavorId", "minNode", "maxNode"}
	for _, key := range camelCaseKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected camelCase JSON key '%s' to be present", key)
		}
	}

	// Verify snake_case keys are NOT present
	snakeCaseKeys := []string{"cluster_id", "sec_group_id", "flavor_id", "min_node", "max_node"}
	for _, key := range snakeCaseKeys {
		if _, ok := raw[key]; ok {
			t.Errorf("unexpected snake_case JSON key '%s' found", key)
		}
	}
}

func TestBuildCreateRequest_CPUGPUDefaults(t *testing.T) {
	ctx := context.Background()

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("c1"),
		SecGroupID:  types.StringValue("s1"),
		Name:        types.StringValue("cpu-gpu"),
		Quantity:    types.Int64Value(1),
		FlavorID:    types.StringValue("f1"),
		Volume:      types.StringValue("10"),
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	body := buildCreateRequest(ctx, plan)

	if !body.CPU {
		t.Error("expected cpu to always be true")
	}
	if body.GPU {
		t.Error("expected gpu to always be false")
	}

	// Also verify in JSON
	data, _ := json.Marshal(body)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if raw["cpu"] != true {
		t.Errorf("expected cpu=true in JSON, got %v", raw["cpu"])
	}
	if raw["gpu"] != false {
		t.Errorf("expected gpu=false in JSON, got %v", raw["gpu"])
	}
}

// --- buildScaleRequest additional tests ---

func TestBuildScaleRequest_ScaleUpByOne(t *testing.T) {
	plan := &k8sNodeGroupModel{Quantity: types.Int64Value(2)}
	state := &k8sNodeGroupModel{
		Quantity:  types.Int64Value(1),
		Name:      types.StringValue("pool-one"),
		ClusterID: types.StringValue("cluster-one"),
	}

	body := buildScaleRequest(plan, state)

	if body.Updates.Type != "add" {
		t.Errorf("expected type 'add', got %s", body.Updates.Type)
	}
	if body.Updates.Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", body.Updates.Quantity)
	}
}

func TestBuildScaleRequest_ScaleDownByOne(t *testing.T) {
	plan := &k8sNodeGroupModel{Quantity: types.Int64Value(1)}
	state := &k8sNodeGroupModel{
		Quantity:  types.Int64Value(2),
		Name:      types.StringValue("pool-down"),
		ClusterID: types.StringValue("cluster-down"),
	}

	body := buildScaleRequest(plan, state)

	if body.Updates.Type != "remove" {
		t.Errorf("expected type 'remove', got %s", body.Updates.Type)
	}
	if body.Updates.Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", body.Updates.Quantity)
	}
}

func TestBuildScaleRequest_LargeScale(t *testing.T) {
	plan := &k8sNodeGroupModel{Quantity: types.Int64Value(8)}
	state := &k8sNodeGroupModel{
		Quantity:  types.Int64Value(1),
		Name:      types.StringValue("pool-large"),
		ClusterID: types.StringValue("cluster-large"),
	}

	body := buildScaleRequest(plan, state)

	if body.Updates.Type != "add" {
		t.Errorf("expected type 'add', got %s", body.Updates.Type)
	}
	if body.Updates.Quantity != 7 {
		t.Errorf("expected quantity 7, got %d", body.Updates.Quantity)
	}
	if body.NodeName != "pool-large" {
		t.Errorf("expected nodeName pool-large, got %s", body.NodeName)
	}
	if body.ClusterID != "cluster-large" {
		t.Errorf("expected clusterId cluster-large, got %s", body.ClusterID)
	}
}

func TestBuildScaleRequest_JSONStructure(t *testing.T) {
	plan := &k8sNodeGroupModel{Quantity: types.Int64Value(5)}
	state := &k8sNodeGroupModel{
		Quantity:  types.Int64Value(2),
		Name:      types.StringValue("pool-json"),
		ClusterID: types.StringValue("cluster-json"),
	}

	body := buildScaleRequest(plan, state)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify top-level keys
	expectedKeys := []string{"updates", "nodeName", "clusterId"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' to be present", key)
		}
	}

	// Verify updates sub-object
	updates, ok := raw["updates"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'updates' to be an object")
	}
	if _, ok := updates["type"]; !ok {
		t.Error("expected updates.type to be present")
	}
	if _, ok := updates["quantity"]; !ok {
		t.Error("expected updates.quantity to be present")
	}

	// Verify values
	if raw["nodeName"] != "pool-json" {
		t.Errorf("expected nodeName=pool-json, got %v", raw["nodeName"])
	}
	if raw["clusterId"] != "cluster-json" {
		t.Errorf("expected clusterId=cluster-json, got %v", raw["clusterId"])
	}
}

// --- mapAPIResponseToState additional tests ---

func TestMapAPIResponseToState_StateField(t *testing.T) {
	ctx := context.Background()

	states := []string{"ACTIVE", "RUNNING", "ERROR", "CREATING", "DELETING", "SCALING"}
	for _, s := range states {
		t.Run(s, func(t *testing.T) {
			model := &k8sNodeGroupModel{
				Labels:      types.MapNull(types.StringType),
				Annotations: types.MapNull(types.StringType),
				MinNode:     types.Int64Null(),
				MaxNode:     types.Int64Null(),
			}
			apiResp := &nodeGroupAPIResponse{
				ID:        "ng-state-test",
				ClusterID: "c1",
				Name:      "state-pool",
				Quantity:  1,
				FlavorID:  "f1",
				Volume:    "10",
				State:     s,
			}

			mapAPIResponseToState(ctx, model, apiResp)

			if model.State.ValueString() != s {
				t.Errorf("expected State %s, got %s", s, model.State.ValueString())
			}
		})
	}
}

func TestMapAPIResponseToState_EmptyState(t *testing.T) {
	ctx := context.Background()

	model := &k8sNodeGroupModel{
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}
	apiResp := &nodeGroupAPIResponse{
		ID:        "ng-empty-state",
		ClusterID: "c1",
		Name:      "empty-state-pool",
		Quantity:  1,
		FlavorID:  "f1",
		Volume:    "10",
		State:     "",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.State.ValueString() != "" {
		t.Errorf("expected empty State, got %s", model.State.ValueString())
	}
	if model.State.IsNull() {
		t.Error("expected State to be empty string, not null")
	}
}

func TestMapAPIResponseToState_OverwriteExisting(t *testing.T) {
	ctx := context.Background()

	// Start with existing labels
	existingLabels, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"old": "value"})
	existingAnnotations, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"old-ann": "old-val"})

	model := &k8sNodeGroupModel{
		Labels:      existingLabels,
		Annotations: existingAnnotations,
		MinNode:     types.Int64Value(1),
		MaxNode:     types.Int64Value(3),
	}

	apiResp := &nodeGroupAPIResponse{
		ID:          "ng-overwrite",
		ClusterID:   "c1",
		SecGroupID:  "sg1",
		Name:        "overwrite-pool",
		Quantity:    2,
		FlavorID:    "f1",
		Volume:      "20",
		Labels:      map[string]string{"new": "label"},
		Annotations: map[string]string{"new-ann": "new-val"},
		MinNode:     2,
		MaxNode:     6,
		State:       "ACTIVE",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	// Labels should be overwritten
	resultLabels := make(map[string]string)
	model.Labels.ElementsAs(ctx, &resultLabels, false)
	if len(resultLabels) != 1 {
		t.Errorf("expected 1 label after overwrite, got %d", len(resultLabels))
	}
	if resultLabels["new"] != "label" {
		t.Errorf("expected label new=label, got %v", resultLabels)
	}
	if _, ok := resultLabels["old"]; ok {
		t.Error("old label should be gone after overwrite")
	}

	// Annotations should be overwritten
	resultAnnotations := make(map[string]string)
	model.Annotations.ElementsAs(ctx, &resultAnnotations, false)
	if resultAnnotations["new-ann"] != "new-val" {
		t.Errorf("expected annotation new-ann=new-val, got %v", resultAnnotations)
	}

	// MinNode/MaxNode should be overwritten
	if model.MinNode.ValueInt64() != 2 {
		t.Errorf("expected MinNode 2, got %d", model.MinNode.ValueInt64())
	}
	if model.MaxNode.ValueInt64() != 6 {
		t.Errorf("expected MaxNode 6, got %d", model.MaxNode.ValueInt64())
	}
}

func TestMapAPIResponseToState_ZeroQuantity(t *testing.T) {
	ctx := context.Background()

	model := &k8sNodeGroupModel{
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}
	apiResp := &nodeGroupAPIResponse{
		ID:        "ng-zero-qty",
		ClusterID: "c1",
		Name:      "zero-pool",
		Quantity:  0,
		FlavorID:  "f1",
		Volume:    "10",
		State:     "ACTIVE",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if model.Quantity.ValueInt64() != 0 {
		t.Errorf("expected Quantity 0, got %d", model.Quantity.ValueInt64())
	}
}

func TestMapAPIResponseToState_AllFields(t *testing.T) {
	ctx := context.Background()

	model := &k8sNodeGroupModel{
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	apiResp := &nodeGroupAPIResponse{
		ID:         "ng-all-fields-001",
		ClusterID:  "cluster-all-001",
		SecGroupID: "sg-all-001",
		Name:       "all-fields-pool",
		Quantity:   4,
		FlavorID:   "flavor-all-001",
		Volume:     "200",
		Labels: map[string]string{
			"a": "1",
			"b": "2",
		},
		Annotations: map[string]string{
			"x": "10",
			"y": "20",
			"z": "30",
		},
		MinNode: 2,
		MaxNode: 7,
		State:   "RUNNING",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	// Verify every field
	if model.ID.ValueString() != "ng-all-fields-001" {
		t.Errorf("ID: expected ng-all-fields-001, got %s", model.ID.ValueString())
	}
	if model.ClusterID.ValueString() != "cluster-all-001" {
		t.Errorf("ClusterID: expected cluster-all-001, got %s", model.ClusterID.ValueString())
	}
	if model.SecGroupID.ValueString() != "sg-all-001" {
		t.Errorf("SecGroupID: expected sg-all-001, got %s", model.SecGroupID.ValueString())
	}
	if model.Name.ValueString() != "all-fields-pool" {
		t.Errorf("Name: expected all-fields-pool, got %s", model.Name.ValueString())
	}
	if model.Quantity.ValueInt64() != 4 {
		t.Errorf("Quantity: expected 4, got %d", model.Quantity.ValueInt64())
	}
	if model.FlavorID.ValueString() != "flavor-all-001" {
		t.Errorf("FlavorID: expected flavor-all-001, got %s", model.FlavorID.ValueString())
	}
	if model.Volume.ValueString() != "200" {
		t.Errorf("Volume: expected 200, got %s", model.Volume.ValueString())
	}
	if model.State.ValueString() != "RUNNING" {
		t.Errorf("State: expected RUNNING, got %s", model.State.ValueString())
	}
	if model.MinNode.ValueInt64() != 2 {
		t.Errorf("MinNode: expected 2, got %d", model.MinNode.ValueInt64())
	}
	if model.MaxNode.ValueInt64() != 7 {
		t.Errorf("MaxNode: expected 7, got %d", model.MaxNode.ValueInt64())
	}

	// Labels
	resultLabels := make(map[string]string)
	model.Labels.ElementsAs(ctx, &resultLabels, false)
	if len(resultLabels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(resultLabels))
	}
	if resultLabels["a"] != "1" || resultLabels["b"] != "2" {
		t.Errorf("labels mismatch: %v", resultLabels)
	}

	// Annotations
	resultAnnotations := make(map[string]string)
	model.Annotations.ElementsAs(ctx, &resultAnnotations, false)
	if len(resultAnnotations) != 3 {
		t.Errorf("expected 3 annotations, got %d", len(resultAnnotations))
	}
	if resultAnnotations["x"] != "10" || resultAnnotations["y"] != "20" || resultAnnotations["z"] != "30" {
		t.Errorf("annotations mismatch: %v", resultAnnotations)
	}
}

// --- parseNodeGroupFromRaw tests ---

func TestParseNodeGroupFromRaw_Valid(t *testing.T) {
	raw := json.RawMessage(`{
		"id": "ng-raw-1",
		"clusterId": "c-raw-1",
		"secGroupId": "sg-raw-1",
		"name": "raw-pool",
		"quantity": 3,
		"flavorId": "f-raw-1",
		"volume": "50",
		"labels": {"env": "dev"},
		"annotations": {},
		"minNode": 1,
		"maxNode": 5,
		"state": "ACTIVE"
	}`)

	ng, err := parseNodeGroupFromRaw(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ng.ID != "ng-raw-1" {
		t.Errorf("expected ID ng-raw-1, got %s", ng.ID)
	}
	if ng.ClusterID != "c-raw-1" {
		t.Errorf("expected ClusterID c-raw-1, got %s", ng.ClusterID)
	}
	if ng.Name != "raw-pool" {
		t.Errorf("expected Name raw-pool, got %s", ng.Name)
	}
	if ng.Quantity != 3 {
		t.Errorf("expected Quantity 3, got %d", ng.Quantity)
	}
	if ng.Labels["env"] != "dev" {
		t.Errorf("expected label env=dev, got %v", ng.Labels)
	}
	if ng.State != "ACTIVE" {
		t.Errorf("expected State ACTIVE, got %s", ng.State)
	}
}

func TestParseNodeGroupFromRaw_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`{not valid json}`)

	_, err := parseNodeGroupFromRaw(raw)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseNodeGroupFromRaw_EmptyJSON(t *testing.T) {
	raw := json.RawMessage(`{}`)

	ng, err := parseNodeGroupFromRaw(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All fields should be zero values
	if ng.ID != "" {
		t.Errorf("expected empty ID, got %s", ng.ID)
	}
	if ng.Name != "" {
		t.Errorf("expected empty Name, got %s", ng.Name)
	}
	if ng.Quantity != 0 {
		t.Errorf("expected Quantity 0, got %d", ng.Quantity)
	}
	if ng.State != "" {
		t.Errorf("expected empty State, got %s", ng.State)
	}
}

func TestParseNodeGroupFromRaw_MinimalFields(t *testing.T) {
	raw := json.RawMessage(`{"id":"ng-min","name":"min-pool","quantity":1}`)

	ng, err := parseNodeGroupFromRaw(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ng.ID != "ng-min" {
		t.Errorf("expected ID ng-min, got %s", ng.ID)
	}
	if ng.Name != "min-pool" {
		t.Errorf("expected Name min-pool, got %s", ng.Name)
	}
	if ng.Quantity != 1 {
		t.Errorf("expected Quantity 1, got %d", ng.Quantity)
	}
	// Unset fields should be zero
	if ng.ClusterID != "" {
		t.Errorf("expected empty ClusterID, got %s", ng.ClusterID)
	}
	if ng.MinNode != 0 {
		t.Errorf("expected MinNode 0, got %d", ng.MinNode)
	}
}

// --- parseNodeGroupListFromData tests ---

func TestParseNodeGroupListFromData_BareArray(t *testing.T) {
	data := json.RawMessage(`[{"id":"ng1","name":"pool1"},{"id":"ng2","name":"pool2"}]`)

	items, err := parseNodeGroupListFromData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Parse first item to verify content
	ng, err := parseNodeGroupFromRaw(items[0])
	if err != nil {
		t.Fatalf("failed to parse first item: %v", err)
	}
	if ng.ID != "ng1" {
		t.Errorf("expected first item ID ng1, got %s", ng.ID)
	}
}

func TestParseNodeGroupListFromData_Envelope(t *testing.T) {
	data := json.RawMessage(`{"data":[{"id":"ng-env-1","name":"env-pool-1"},{"id":"ng-env-2","name":"env-pool-2"},{"id":"ng-env-3","name":"env-pool-3"}]}`)

	items, err := parseNodeGroupListFromData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Parse last item to verify
	ng, err := parseNodeGroupFromRaw(items[2])
	if err != nil {
		t.Fatalf("failed to parse item: %v", err)
	}
	if ng.ID != "ng-env-3" {
		t.Errorf("expected ID ng-env-3, got %s", ng.ID)
	}
}

func TestParseNodeGroupListFromData_EmptyArray(t *testing.T) {
	data := json.RawMessage(`[]`)

	items, err := parseNodeGroupListFromData(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestParseNodeGroupListFromData_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`{not valid}`)

	_, err := parseNodeGroupListFromData(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseNodeGroupListFromData_EmptyEnvelope(t *testing.T) {
	// Envelope with empty data array — len(envelope.Data) == 0, falls through
	// to bare array parsing which fails since input is an object, not an array
	data := json.RawMessage(`{"data":[]}`)

	_, err := parseNodeGroupListFromData(data)
	if err == nil {
		t.Fatal("expected error for empty envelope (falls through to bare array parse of object), got nil")
	}
}

func TestBuildCreateRequest_MinNodeOnly(t *testing.T) {
	ctx := context.Background()

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("c1"),
		SecGroupID:  types.StringValue("s1"),
		Name:        types.StringValue("min-only"),
		Quantity:    types.Int64Value(2),
		FlavorID:    types.StringValue("f1"),
		Volume:      types.StringValue("20"),
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Value(1),
		MaxNode:     types.Int64Null(),
	}

	body := buildCreateRequest(ctx, plan)

	if body.MinNode == nil || *body.MinNode != 1 {
		t.Errorf("expected MinNode 1, got %v", body.MinNode)
	}
	if body.MaxNode != nil {
		t.Errorf("expected MaxNode nil, got %v", body.MaxNode)
	}

	// Verify JSON: minNode present, maxNode omitted
	data, _ := json.Marshal(body)
	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, ok := raw["minNode"]; !ok {
		t.Error("expected minNode in JSON")
	}
	if _, ok := raw["maxNode"]; ok {
		t.Error("expected maxNode to be omitted from JSON")
	}
}

func TestBuildCreateRequest_MaxNodeOnly(t *testing.T) {
	ctx := context.Background()

	plan := &k8sNodeGroupModel{
		ClusterID:   types.StringValue("c1"),
		SecGroupID:  types.StringValue("s1"),
		Name:        types.StringValue("max-only"),
		Quantity:    types.Int64Value(2),
		FlavorID:    types.StringValue("f1"),
		Volume:      types.StringValue("20"),
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Value(8),
	}

	body := buildCreateRequest(ctx, plan)

	if body.MinNode != nil {
		t.Errorf("expected MinNode nil, got %v", body.MinNode)
	}
	if body.MaxNode == nil || *body.MaxNode != 8 {
		t.Errorf("expected MaxNode 8, got %v", body.MaxNode)
	}
}

func TestMapAPIResponseToState_PreserveNullLabelsWhenAPIEmpty(t *testing.T) {
	ctx := context.Background()

	// Model has null labels; API returns empty labels — should stay null
	model := &k8sNodeGroupModel{
		Labels:      types.MapNull(types.StringType),
		Annotations: types.MapNull(types.StringType),
		MinNode:     types.Int64Null(),
		MaxNode:     types.Int64Null(),
	}

	apiResp := &nodeGroupAPIResponse{
		ID:          "ng-preserve",
		ClusterID:   "c1",
		SecGroupID:  "sg1",
		Name:        "preserve-pool",
		Quantity:    1,
		FlavorID:    "f1",
		Volume:      "10",
		Labels:      map[string]string{},
		Annotations: map[string]string{},
		MinNode:     0,
		MaxNode:     0,
		State:       "ACTIVE",
	}

	mapAPIResponseToState(ctx, model, apiResp)

	if !model.Labels.IsNull() {
		t.Error("expected Labels to remain null when API returns empty map")
	}
	if !model.Annotations.IsNull() {
		t.Error("expected Annotations to remain null when API returns empty map")
	}
	if !model.MinNode.IsNull() {
		t.Error("expected MinNode to remain null when API returns 0")
	}
	if !model.MaxNode.IsNull() {
		t.Error("expected MaxNode to remain null when API returns 0")
	}
}
