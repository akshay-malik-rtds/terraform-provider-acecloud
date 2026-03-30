package k8s_cluster

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
// buildCreateRequest — all fields
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_AllFields(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("my-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringValue("No"),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(8),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolValue(true),
		AutoscaleMin:        types.Int64Value(2),
		AutoscaleMax:        types.Int64Value(5),
		WorkerNodeName:      types.StringValue("pool1"),
		WorkerQuantity:      types.Int64Value(3),
		FlavorID:            types.StringValue("flavor-uuid-123"),
		VolumeSize:          types.Int64Value(50),
		SecGroupID:          types.StringValue("sg-uuid-456"),
	}

	body := buildCreateRequest(plan)

	if body.Name != "my-cluster" {
		t.Errorf("expected name my-cluster, got %s", body.Name)
	}
	if body.KubernetesVersion != "v1.32.6+rke2r1" {
		t.Errorf("expected kubernetesVersion v1.32.6+rke2r1, got %s", body.KubernetesVersion)
	}
	if body.EndpointAccess != "Public" {
		t.Errorf("expected endpointAccess Public, got %s", body.EndpointAccess)
	}
	if body.NetworkIsolation != "Disabled" {
		t.Errorf("expected networkIsolation Disabled, got %s", body.NetworkIsolation)
	}
	if body.NginxIngress != "Enabled" {
		t.Errorf("expected nginxIngress Enabled, got %s", body.NginxIngress)
	}
	if body.NginxDefaultBackend != "Enabled" {
		t.Errorf("expected nginxDefaultBackend Enabled, got %s", body.NginxDefaultBackend)
	}
	if body.NetworkProvider != "Calico" {
		t.Errorf("expected networkProvider Calico, got %s", body.NetworkProvider)
	}
	if body.SnapshotBackup != "No" {
		t.Errorf("expected snapshotBackup No, got %s", body.SnapshotBackup)
	}
	if body.SecretsEncryption != "Disabled" {
		t.Errorf("expected secretsEncryption Disabled, got %s", body.SecretsEncryption)
	}
	if body.MaxWorkerNodes != 8 {
		t.Errorf("expected maxWorkerNodes 8, got %d", body.MaxWorkerNodes)
	}
	if body.ClusterType != "rke2" {
		t.Errorf("expected clusterType rke2, got %s", body.ClusterType)
	}
	if body.Autoscale != true {
		t.Error("expected autoscale true")
	}
	if body.AutoscaleMin != 2 {
		t.Errorf("expected autoscaleMin 2, got %d", body.AutoscaleMin)
	}
	if body.AutoscaleMax != 5 {
		t.Errorf("expected autoscaleMax 5, got %d", body.AutoscaleMax)
	}
	if body.WorkerNodeName != "pool1" {
		t.Errorf("expected workerNodeName pool1, got %s", body.WorkerNodeName)
	}
	if body.Quantity != 3 {
		t.Errorf("expected quantity 3, got %d", body.Quantity)
	}
	if body.FlavorID != "flavor-uuid-123" {
		t.Errorf("expected flavorId flavor-uuid-123, got %s", body.FlavorID)
	}
	if body.Volume != 50 {
		t.Errorf("expected volume 50, got %d", body.Volume)
	}

	// Verify sec_group_id is set when provided.
	if body.SecGroupID != "sg-uuid-456" {
		t.Errorf("expected secGroupId sg-uuid-456, got %s", body.SecGroupID)
	}

	// Verify hardcoded defaults.
	if body.SnapshotInterval != 12 {
		t.Errorf("expected snapshotInterval 12, got %d", body.SnapshotInterval)
	}
	if body.SnapshotRetention != 6 {
		t.Errorf("expected snapshotRetention 6, got %d", body.SnapshotRetention)
	}
	if body.SnapshotEnabled != "No" {
		t.Errorf("expected snapshotEnabled No, got %s", body.SnapshotEnabled)
	}
	if body.DrainNodes != "No" {
		t.Errorf("expected drainNodes No, got %s", body.DrainNodes)
	}
	if body.DrainTimeout != "Give up after:" {
		t.Errorf("expected drainTimeout 'Give up after:', got %s", body.DrainTimeout)
	}
	if body.TerminationTime != 30 {
		t.Errorf("expected terminationTime 30, got %d", body.TerminationTime)
	}
	if body.DrainTime != 120 {
		t.Errorf("expected drainTime 120, got %d", body.DrainTime)
	}
	if body.DeleteDirData != "No" {
		t.Errorf("expected deleteDirData No, got %s", body.DeleteDirData)
	}
	if body.Force != "No" {
		t.Errorf("expected force No, got %s", body.Force)
	}
	if body.GracePeriod != "-1" {
		t.Errorf("expected gracePeriod -1, got %s", body.GracePeriod)
	}
	if len(body.Sources) != 1 || body.Sources[0].ID != "default-cidr" || body.Sources[0].Value != "0.0.0.0/0" {
		t.Errorf("expected default sources, got %+v", body.Sources)
	}
	if !body.CPU {
		t.Error("expected cpu true")
	}
	if body.GPU {
		t.Error("expected gpu false")
	}
}

func TestBuildCreateRequest_MinimalFields(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("minimal-cluster"),
		KubernetesVersion:   types.StringValue("v1.30.0+rke2r1"),
		EndpointAccess:      types.StringValue("Private"),
		NetworkIsolation:    types.StringValue("Enabled"),
		NginxIngress:        types.StringValue("Disabled"),
		NginxDefaultBackend: types.StringValue("Disabled"),
		NetworkProvider:     types.StringValue("Flannel"),
		SnapshotBackup:      types.StringNull(), // Optional, omitted.
		SecretsEncryption:   types.StringValue("Enabled"),
		MaxWorkerNodes:      types.Int64Value(1),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolNull(),  // Optional, omitted.
		AutoscaleMin:        types.Int64Null(), // Optional, omitted.
		AutoscaleMax:        types.Int64Null(), // Optional, omitted.
		WorkerNodeName:      types.StringValue("workers"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-small"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringNull(), // Optional, omitted.
	}

	body := buildCreateRequest(plan)

	if body.Name != "minimal-cluster" {
		t.Errorf("expected name minimal-cluster, got %s", body.Name)
	}
	// SnapshotBackup defaults to "No" when plan value is null (required by API DTO).
	if body.SnapshotBackup != "No" {
		t.Errorf("expected snapshotBackup 'No' (default), got %s", body.SnapshotBackup)
	}
	// New required fields always have defaults.
	if body.SnapshotEnabled != "No" {
		t.Errorf("expected snapshotEnabled No, got %s", body.SnapshotEnabled)
	}
	if body.DeleteDirData != "No" {
		t.Errorf("expected deleteDirData No, got %s", body.DeleteDirData)
	}
	if body.Force != "No" {
		t.Errorf("expected force No, got %s", body.Force)
	}
	if body.GracePeriod != "-1" {
		t.Errorf("expected gracePeriod -1, got %s", body.GracePeriod)
	}
	if body.SecGroupID != "" {
		t.Errorf("expected empty secGroupId when null, got %s", body.SecGroupID)
	}
	if body.Autoscale != false {
		t.Error("expected autoscale false when null")
	}
	if body.AutoscaleMin != 0 {
		t.Errorf("expected autoscaleMin 0 when null, got %d", body.AutoscaleMin)
	}
	if body.AutoscaleMax != 0 {
		t.Errorf("expected autoscaleMax 0 when null, got %d", body.AutoscaleMax)
	}
	if body.MaxWorkerNodes != 1 {
		t.Errorf("expected maxWorkerNodes 1, got %d", body.MaxWorkerNodes)
	}
	if body.NetworkProvider != "Flannel" {
		t.Errorf("expected networkProvider Flannel, got %s", body.NetworkProvider)
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest — JSON serialization
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_JSON(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("json-test"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public and Private"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringValue("Yes"),
		SecretsEncryption:   types.StringValue("Enabled"),
		MaxWorkerNodes:      types.Int64Value(10),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolValue(false),
		AutoscaleMin:        types.Int64Value(1),
		AutoscaleMax:        types.Int64Value(3),
		WorkerNodeName:      types.StringValue("pool1"),
		WorkerQuantity:      types.Int64Value(2),
		FlavorID:            types.StringValue("flv-json"),
		VolumeSize:          types.Int64Value(100),
		SecGroupID:          types.StringNull(),
	}

	body := buildCreateRequest(plan)

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal into a generic map to verify camelCase JSON key names.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal into map failed: %v", err)
	}

	// Verify all required JSON keys are present.
	requiredKeys := []string{
		"name",
		"kubernetesVersion",
		"endpointAccess",
		"networkIsolation",
		"nginxIngress",
		"nginxDefaultBackend",
		"networkProvider",
		"secretsEncryption",
		"maxWorkerNodes",
		"workerNodeName",
		"quantity",
		"flavorId",
		"volume",
		"cluster_type",
		"autoscale",
		"autoscaleMin",
		"autoscaleMax",
		"snapshotInterval",
		"snapshotRetention",
		"snapshotEnabled",
		"drainNodes",
		"drainTimeout",
		"terminationTime",
		"drainTime",
		"deleteDirData",
		"force",
		"gracePeriod",
		"sources",
		"cpu",
		"gpu",
	}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected camelCase JSON key %q to be present", key)
		}
	}

	// Verify optional key present when set.
	if _, ok := raw["snapshotBackup"]; !ok {
		t.Error("expected snapshotBackup key to be present when set")
	}

	// Round-trip back into the struct to verify field values.
	var deserialized createK8sClusterRequest
	if err := json.Unmarshal(data, &deserialized); err != nil {
		t.Fatalf("json.Unmarshal into createK8sClusterRequest failed: %v", err)
	}

	if deserialized.Name != "json-test" {
		t.Errorf("JSON round-trip name: expected 'json-test', got %s", deserialized.Name)
	}
	if deserialized.EndpointAccess != "Public and Private" {
		t.Errorf("JSON round-trip endpointAccess: expected 'Public and Private', got %s", deserialized.EndpointAccess)
	}
	if deserialized.MaxWorkerNodes != 10 {
		t.Errorf("JSON round-trip maxWorkerNodes: expected 10, got %d", deserialized.MaxWorkerNodes)
	}
	if deserialized.Volume != 100 {
		t.Errorf("JSON round-trip volume: expected 100, got %d", deserialized.Volume)
	}
	if deserialized.SnapshotBackup != "Yes" {
		t.Errorf("JSON round-trip snapshotBackup: expected 'Yes', got %s", deserialized.SnapshotBackup)
	}
	if deserialized.Quantity != 2 {
		t.Errorf("JSON round-trip quantity: expected 2, got %d", deserialized.Quantity)
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest — hardcoded defaults
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_HardcodedDefaults(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("defaults-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringNull(),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(3),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolNull(),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		WorkerNodeName:      types.StringValue("pool-defaults"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-def"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringNull(),
	}

	body := buildCreateRequest(plan)

	// Verify all hardcoded defaults explicitly
	if body.SnapshotEnabled != "No" {
		t.Errorf("expected SnapshotEnabled 'No', got %q", body.SnapshotEnabled)
	}
	if body.DeleteDirData != "No" {
		t.Errorf("expected DeleteDirData 'No', got %q", body.DeleteDirData)
	}
	if body.Force != "No" {
		t.Errorf("expected Force 'No', got %q", body.Force)
	}
	if body.GracePeriod != "-1" {
		t.Errorf("expected GracePeriod '-1', got %q", body.GracePeriod)
	}
	if body.SnapshotInterval != 12 {
		t.Errorf("expected SnapshotInterval 12, got %d", body.SnapshotInterval)
	}
	if body.SnapshotRetention != 6 {
		t.Errorf("expected SnapshotRetention 6, got %d", body.SnapshotRetention)
	}
	if body.DrainNodes != "No" {
		t.Errorf("expected DrainNodes 'No', got %q", body.DrainNodes)
	}
	if body.DrainTimeout != "Give up after:" {
		t.Errorf("expected DrainTimeout 'Give up after:', got %q", body.DrainTimeout)
	}
	if body.TerminationTime != 30 {
		t.Errorf("expected TerminationTime 30, got %d", body.TerminationTime)
	}
	if body.DrainTime != 120 {
		t.Errorf("expected DrainTime 120, got %d", body.DrainTime)
	}
	if !body.CPU {
		t.Error("expected CPU true (hardcoded default)")
	}
	if body.GPU {
		t.Error("expected GPU false (hardcoded default)")
	}
	if len(body.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(body.Sources))
	}
	if body.Sources[0].ID != "default-cidr" {
		t.Errorf("expected source ID 'default-cidr', got %q", body.Sources[0].ID)
	}
	if body.Sources[0].Value != "0.0.0.0/0" {
		t.Errorf("expected source value '0.0.0.0/0', got %q", body.Sources[0].Value)
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest — secGroupID handling
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_WithSecGroupID(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("sg-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringNull(),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(3),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolNull(),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		WorkerNodeName:      types.StringValue("pool-sg"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-sg"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringValue("sg-abc-123"),
	}

	body := buildCreateRequest(plan)
	if body.SecGroupID != "sg-abc-123" {
		t.Errorf("expected secGroupId sg-abc-123, got %q", body.SecGroupID)
	}

	// Verify JSON: secGroupId should be present (not empty, not omitted)
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if raw["secGroupId"] != "sg-abc-123" {
		t.Errorf("expected secGroupId in JSON, got %v", raw["secGroupId"])
	}

	// Verify omitempty: when null, secGroupId should be absent from JSON
	planNoSg := &k8sClusterResourceModel{
		Name:                types.StringValue("no-sg-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringNull(),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(3),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolNull(),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		WorkerNodeName:      types.StringValue("pool-no-sg"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-no-sg"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringNull(),
	}
	bodyNoSg := buildCreateRequest(planNoSg)
	dataNoSg, _ := json.Marshal(bodyNoSg)
	var rawNoSg map[string]interface{}
	json.Unmarshal(dataNoSg, &rawNoSg)
	if _, ok := rawNoSg["secGroupId"]; ok {
		t.Error("expected secGroupId to be omitted when null (omitempty)")
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest — snapshot_backup "Yes" sets snapshotEnabled
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_SnapshotBackupYes(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("snap-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringValue("Yes"),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(3),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolNull(),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		WorkerNodeName:      types.StringValue("pool1"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-snap"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.SnapshotBackup != "Yes" {
		t.Errorf("expected snapshotBackup 'Yes', got %q", body.SnapshotBackup)
	}
	if body.SnapshotEnabled != "Yes" {
		t.Errorf("expected snapshotEnabled 'Yes' when snapshotBackup is 'Yes', got %q", body.SnapshotEnabled)
	}
}

func TestBuildCreateRequest_SnapshotBackupNo(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("no-snap-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringValue("No"),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(3),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolNull(),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		WorkerNodeName:      types.StringValue("pool1"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-no-snap"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.SnapshotBackup != "No" {
		t.Errorf("expected snapshotBackup 'No', got %q", body.SnapshotBackup)
	}
	// snapshotEnabled stays "No" when backup is "No"
	if body.SnapshotEnabled != "No" {
		t.Errorf("expected snapshotEnabled 'No' when snapshotBackup is 'No', got %q", body.SnapshotEnabled)
	}
}

func TestBuildCreateRequest_SnapshotBackupNull(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("null-snap"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringNull(),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(3),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolNull(),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		WorkerNodeName:      types.StringValue("pool1"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-null"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.SnapshotBackup != "No" {
		t.Errorf("expected snapshotBackup 'No' (default when null), got %q", body.SnapshotBackup)
	}
	if body.SnapshotEnabled != "No" {
		t.Errorf("expected snapshotEnabled 'No' when snapshotBackup is null, got %q", body.SnapshotEnabled)
	}
}

// ---------------------------------------------------------------------------
// buildCreateRequest — autoscale combinations
// ---------------------------------------------------------------------------

func TestBuildCreateRequest_AutoscaleEnabled(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("autoscale-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringNull(),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(10),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolValue(true),
		AutoscaleMin:        types.Int64Value(2),
		AutoscaleMax:        types.Int64Value(8),
		WorkerNodeName:      types.StringValue("pool-auto"),
		WorkerQuantity:      types.Int64Value(3),
		FlavorID:            types.StringValue("flv-auto"),
		VolumeSize:          types.Int64Value(50),
		SecGroupID:          types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.Autoscale != true {
		t.Error("expected autoscale true")
	}
	if body.AutoscaleMin != 2 {
		t.Errorf("expected autoscaleMin 2, got %d", body.AutoscaleMin)
	}
	if body.AutoscaleMax != 8 {
		t.Errorf("expected autoscaleMax 8, got %d", body.AutoscaleMax)
	}
}

func TestBuildCreateRequest_AutoscaleDisabled(t *testing.T) {
	plan := &k8sClusterResourceModel{
		Name:                types.StringValue("no-autoscale"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SnapshotBackup:      types.StringNull(),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(3),
		ClusterType:         types.StringValue("rke2"),
		Autoscale:           types.BoolValue(false),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		WorkerNodeName:      types.StringValue("pool1"),
		WorkerQuantity:      types.Int64Value(1),
		FlavorID:            types.StringValue("flv-1"),
		VolumeSize:          types.Int64Value(20),
		SecGroupID:          types.StringNull(),
	}

	body := buildCreateRequest(plan)

	if body.Autoscale != false {
		t.Error("expected autoscale false")
	}
	if body.AutoscaleMin != 0 {
		t.Errorf("expected autoscaleMin 0 when null, got %d", body.AutoscaleMin)
	}
	if body.AutoscaleMax != 0 {
		t.Errorf("expected autoscaleMax 0 when null, got %d", body.AutoscaleMax)
	}
}

// ---------------------------------------------------------------------------
// mapReadResponseToState
// ---------------------------------------------------------------------------

func TestMapAPIResponseToState(t *testing.T) {
	state := &k8sClusterResourceModel{
		Name:                types.StringValue("existing-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public"),
		NetworkIsolation:    types.StringValue("Disabled"),
		NginxIngress:        types.StringValue("Enabled"),
		NginxDefaultBackend: types.StringValue("Enabled"),
		NetworkProvider:     types.StringValue("Calico"),
		SecretsEncryption:   types.StringValue("Disabled"),
		MaxWorkerNodes:      types.Int64Value(8),
		WorkerNodeName:      types.StringValue("pool1"),
		WorkerQuantity:      types.Int64Value(3),
		FlavorID:            types.StringValue("flavor-uuid"),
		VolumeSize:          types.Int64Value(50),
		ClusterType:         types.StringValue("rke2"),
		SnapshotBackup:      types.StringNull(),
		Autoscale:           types.BoolNull(),
		AutoscaleMin:        types.Int64Null(),
		AutoscaleMax:        types.Int64Null(),
		SecGroupID:          types.StringNull(),
		Status:              types.StringValue("provisioning"),
		CreatedAt:           types.StringValue(""),
	}

	cluster := &readK8sClusterResponse{
		ID:      "cluster-abc-123",
		Name:    "existing-cluster",
		State:   "active",
		Created: "2026-03-15T10:30:00Z",
	}

	mapReadResponseToState(cluster, state)

	if state.ID.ValueString() != "cluster-abc-123" {
		t.Errorf("expected ID cluster-abc-123, got %s", state.ID.ValueString())
	}
	if state.Name.ValueString() != "existing-cluster" {
		t.Errorf("expected Name existing-cluster, got %s", state.Name.ValueString())
	}
	if state.Status.ValueString() != "active" {
		t.Errorf("expected Status active, got %s", state.Status.ValueString())
	}
	if state.CreatedAt.ValueString() != "2026-03-15T10:30:00Z" {
		t.Errorf("expected CreatedAt from 'created' field, got %s", state.CreatedAt.ValueString())
	}

	// Verify configurable fields are preserved (not overwritten).
	if state.KubernetesVersion.ValueString() != "v1.32.6+rke2r1" {
		t.Errorf("expected KubernetesVersion preserved, got %s", state.KubernetesVersion.ValueString())
	}
	if state.EndpointAccess.ValueString() != "Public" {
		t.Errorf("expected EndpointAccess preserved, got %s", state.EndpointAccess.ValueString())
	}
	if state.MaxWorkerNodes.ValueInt64() != 8 {
		t.Errorf("expected MaxWorkerNodes preserved at 8, got %d", state.MaxWorkerNodes.ValueInt64())
	}
	if !state.SnapshotBackup.IsNull() {
		t.Error("expected SnapshotBackup to remain null")
	}
	if !state.Autoscale.IsNull() {
		t.Error("expected Autoscale to remain null")
	}
}

func TestMapAPIResponseToState_AllStatuses(t *testing.T) {
	statuses := []string{"active", "Active", "error", "Error", "failed", "Never Created", "provisioning", "deleting"}
	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			state := &k8sClusterResourceModel{
				Name:                types.StringValue("status-cluster"),
				KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
				EndpointAccess:      types.StringValue("Public"),
				NetworkIsolation:    types.StringValue("Disabled"),
				NginxIngress:        types.StringValue("Enabled"),
				NginxDefaultBackend: types.StringValue("Enabled"),
				NetworkProvider:     types.StringValue("Calico"),
				SecretsEncryption:   types.StringValue("Disabled"),
				MaxWorkerNodes:      types.Int64Value(3),
				WorkerNodeName:      types.StringValue("pool1"),
				WorkerQuantity:      types.Int64Value(1),
				FlavorID:            types.StringValue("flv-status"),
				VolumeSize:          types.Int64Value(20),
				ClusterType:         types.StringValue("rke2"),
				SnapshotBackup:      types.StringNull(),
				Autoscale:           types.BoolNull(),
				AutoscaleMin:        types.Int64Null(),
				AutoscaleMax:        types.Int64Null(),
				SecGroupID:          types.StringNull(),
				Status:              types.StringValue(""),
				CreatedAt:           types.StringValue(""),
			}

			cluster := &readK8sClusterResponse{
				ID:      "cluster-status-test",
				Name:    "status-cluster",
				State:   status,
				Created: "2026-03-15T10:30:00Z",
			}

			mapReadResponseToState(cluster, state)

			if state.Status.ValueString() != status {
				t.Errorf("expected Status %q, got %q", status, state.Status.ValueString())
			}
			if state.ID.ValueString() != "cluster-status-test" {
				t.Errorf("expected ID cluster-status-test, got %s", state.ID.ValueString())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// mapReadResponseToState — CreatedAt fallback logic
// ---------------------------------------------------------------------------

func TestMapReadResponseToState_CreatedAtFallback(t *testing.T) {
	tests := []struct {
		name      string
		cluster   *readK8sClusterResponse
		wantValue string
	}{
		{
			name: "created field",
			cluster: &readK8sClusterResponse{
				ID:      "c1",
				Name:    "c1-name",
				State:   "active",
				Created: "2026-01-01T00:00:00Z",
			},
			wantValue: "2026-01-01T00:00:00Z",
		},
		{
			name: "created_at fallback",
			cluster: &readK8sClusterResponse{
				ID:        "c2",
				Name:      "c2-name",
				State:     "active",
				CreatedAt: "2026-02-01T00:00:00Z",
			},
			wantValue: "2026-02-01T00:00:00Z",
		},
		{
			name: "neither field present",
			cluster: &readK8sClusterResponse{
				ID:    "c3",
				Name:  "c3-name",
				State: "active",
			},
			wantValue: "",
		},
		{
			name: "created takes priority over created_at",
			cluster: &readK8sClusterResponse{
				ID:        "c4",
				Name:      "c4-name",
				State:     "active",
				Created:   "2026-01-01T00:00:00Z",
				CreatedAt: "2026-02-01T00:00:00Z",
			},
			wantValue: "2026-01-01T00:00:00Z",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := &k8sClusterResourceModel{
				Status:    types.StringValue(""),
				CreatedAt: types.StringValue(""),
			}
			mapReadResponseToState(tc.cluster, state)
			if state.CreatedAt.ValueString() != tc.wantValue {
				t.Errorf("expected CreatedAt %q, got %q", tc.wantValue, state.CreatedAt.ValueString())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// mapReadResponseToState — Name handling
// ---------------------------------------------------------------------------

func TestMapReadResponseToState_EmptyName(t *testing.T) {
	state := &k8sClusterResourceModel{
		Name:      types.StringValue("original-name"),
		Status:    types.StringValue(""),
		CreatedAt: types.StringValue(""),
	}

	cluster := &readK8sClusterResponse{
		ID:      "c-empty-name",
		Name:    "",
		State:   "active",
		Created: "2026-01-01T00:00:00Z",
	}

	mapReadResponseToState(cluster, state)

	// When API returns empty name, the original should be preserved
	if state.Name.ValueString() != "original-name" {
		t.Errorf("expected Name preserved as 'original-name', got %q", state.Name.ValueString())
	}
}

func TestMapReadResponseToState_NameOverwrite(t *testing.T) {
	state := &k8sClusterResourceModel{
		Name:      types.StringValue("old-name"),
		Status:    types.StringValue(""),
		CreatedAt: types.StringValue(""),
	}

	cluster := &readK8sClusterResponse{
		ID:      "c-new-name",
		Name:    "new-name",
		State:   "active",
		Created: "2026-01-01T00:00:00Z",
	}

	mapReadResponseToState(cluster, state)

	if state.Name.ValueString() != "new-name" {
		t.Errorf("expected Name updated to 'new-name', got %q", state.Name.ValueString())
	}
}

// ---------------------------------------------------------------------------
// parseClusterData
// ---------------------------------------------------------------------------

func TestParseClusterData(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantID    string
		wantName  string
		wantState string
		wantErr   bool
	}{
		{
			name:      "standard response",
			data:      `{"id":"cluster-123","name":"my-cluster","state":"active","created":"2026-03-15T10:30:00Z"}`,
			wantID:    "cluster-123",
			wantName:  "my-cluster",
			wantState: "active",
		},
		{
			name:      "minimal response",
			data:      `{"id":"c1","name":"c1","state":"provisioning"}`,
			wantID:    "c1",
			wantName:  "c1",
			wantState: "provisioning",
		},
		{
			name:    "invalid JSON",
			data:    `{invalid}`,
			wantErr: true,
		},
		{
			name:      "empty fields",
			data:      `{"id":"","name":"","state":""}`,
			wantID:    "",
			wantName:  "",
			wantState: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseClusterData(json.RawMessage(tc.data))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tc.wantID {
				t.Errorf("expected ID %q, got %q", tc.wantID, result.ID)
			}
			if result.Name != tc.wantName {
				t.Errorf("expected Name %q, got %q", tc.wantName, result.Name)
			}
			if result.State != tc.wantState {
				t.Errorf("expected State %q, got %q", tc.wantState, result.State)
			}
		})
	}
}

func TestParseClusterData_WithCreatedFields(t *testing.T) {
	data := `{"id":"c1","name":"cluster","state":"active","created":"2026-01-01T00:00:00Z","created_at":"2026-02-01T00:00:00Z"}`
	result, err := parseClusterData(json.RawMessage(data))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != "2026-01-01T00:00:00Z" {
		t.Errorf("expected Created '2026-01-01T00:00:00Z', got %q", result.Created)
	}
	if result.CreatedAt != "2026-02-01T00:00:00Z" {
		t.Errorf("expected CreatedAt '2026-02-01T00:00:00Z', got %q", result.CreatedAt)
	}
}

// ---------------------------------------------------------------------------
// Delete path construction
// ---------------------------------------------------------------------------

func TestDeletePath(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		wantPath string
	}{
		{
			name:     "standard uuid",
			id:       "cluster-abc-123",
			wantPath: "/k8s/cluster-overview/cluster-abc-123",
		},
		{
			name:     "numeric id",
			id:       "12345",
			wantPath: "/k8s/cluster-overview/12345",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := fmt.Sprintf("/k8s/cluster-overview/%s", tc.id)
			if path != tc.wantPath {
				t.Errorf("expected path %q, got %q", tc.wantPath, path)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// List response parsing
// ---------------------------------------------------------------------------

func TestListK8sClusterItem_JSON(t *testing.T) {
	data := `[
		{"id":"c1","name":"cluster-a","state":"active"},
		{"id":"c2","name":"cluster-b","state":"provisioning"},
		{"id":"c3","name":"cluster-c","state":"error"}
	]`

	var clusters []listK8sClusterItem
	if err := json.Unmarshal([]byte(data), &clusters); err != nil {
		t.Fatalf("failed to unmarshal list: %v", err)
	}

	if len(clusters) != 3 {
		t.Fatalf("expected 3 clusters, got %d", len(clusters))
	}

	if clusters[0].ID != "c1" || clusters[0].Name != "cluster-a" || clusters[0].State != "active" {
		t.Errorf("unexpected first cluster: %+v", clusters[0])
	}
	if clusters[1].State != "provisioning" {
		t.Errorf("expected second cluster state 'provisioning', got %q", clusters[1].State)
	}
}

func TestListK8sClusterItem_FindByName(t *testing.T) {
	clusters := []listK8sClusterItem{
		{ID: "c1", Name: "alpha", State: "active"},
		{ID: "c2", Name: "beta", State: "active"},
		{ID: "c3", Name: "gamma", State: "error"},
	}

	tests := []struct {
		name      string
		target    string
		wantFound bool
		wantID    string
	}{
		{name: "found first", target: "alpha", wantFound: true, wantID: "c1"},
		{name: "found middle", target: "beta", wantFound: true, wantID: "c2"},
		{name: "found last", target: "gamma", wantFound: true, wantID: "c3"},
		{name: "not found", target: "delta", wantFound: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var found *listK8sClusterItem
			for i := range clusters {
				if clusters[i].Name == tc.target {
					found = &clusters[i]
					break
				}
			}

			if tc.wantFound {
				if found == nil {
					t.Fatal("expected cluster to be found")
				}
				if found.ID != tc.wantID {
					t.Errorf("expected ID %q, got %q", tc.wantID, found.ID)
				}
			} else {
				if found != nil {
					t.Error("expected cluster not to be found")
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// List response — non-array during provisioning
// ---------------------------------------------------------------------------

func TestListResponse_NonArray(t *testing.T) {
	// During provisioning, the API may return a non-array object
	data := `{"clusterStatus":"Creating"}`

	var clusters []listK8sClusterItem
	err := json.Unmarshal([]byte(data), &clusters)

	// This should fail to unmarshal as an array
	if err == nil {
		t.Error("expected error when unmarshaling non-array response, got nil")
	}
}

func TestListResponse_EmptyArray(t *testing.T) {
	data := `[]`

	var clusters []listK8sClusterItem
	if err := json.Unmarshal([]byte(data), &clusters); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters, got %d", len(clusters))
	}
}

// ---------------------------------------------------------------------------
// createK8sClusterRequest — JSON field tags
// ---------------------------------------------------------------------------

func TestCreateRequest_ClusterTypeSnakeCase(t *testing.T) {
	// Verify cluster_type uses snake_case in JSON (not camelCase)
	body := &createK8sClusterRequest{
		Name:        "test",
		ClusterType: "rke2",
		Sources:     []clusterSource{{ID: "default-cidr", Value: "0.0.0.0/0"}},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if _, ok := raw["cluster_type"]; !ok {
		t.Error("expected 'cluster_type' (snake_case) key in JSON")
	}
	if _, ok := raw["clusterType"]; ok {
		t.Error("unexpected 'clusterType' (camelCase) key in JSON — should be snake_case")
	}
}

// ---------------------------------------------------------------------------
// clusterSource struct
// ---------------------------------------------------------------------------

func TestClusterSource_JSON(t *testing.T) {
	sources := []clusterSource{
		{ID: "default-cidr", Value: "0.0.0.0/0"},
		{ID: "custom-cidr", Value: "10.0.0.0/8"},
	}

	data, err := json.Marshal(sources)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded []clusterSource
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	if len(decoded) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(decoded))
	}
	if decoded[0].ID != "default-cidr" || decoded[0].Value != "0.0.0.0/0" {
		t.Errorf("unexpected first source: %+v", decoded[0])
	}
	if decoded[1].ID != "custom-cidr" || decoded[1].Value != "10.0.0.0/8" {
		t.Errorf("unexpected second source: %+v", decoded[1])
	}
}

// ---------------------------------------------------------------------------
// readK8sClusterResponse — JSON parsing
// ---------------------------------------------------------------------------

func TestReadK8sClusterResponse_JSON(t *testing.T) {
	data := `{
		"id": "cluster-xyz",
		"name": "prod-cluster",
		"state": "active",
		"created": "2026-03-15T10:30:00Z",
		"created_at": "2026-03-15T10:30:00Z"
	}`

	var resp readK8sClusterResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.ID != "cluster-xyz" {
		t.Errorf("expected ID 'cluster-xyz', got %q", resp.ID)
	}
	if resp.Name != "prod-cluster" {
		t.Errorf("expected Name 'prod-cluster', got %q", resp.Name)
	}
	if resp.State != "active" {
		t.Errorf("expected State 'active', got %q", resp.State)
	}
	if resp.Created != "2026-03-15T10:30:00Z" {
		t.Errorf("expected Created, got %q", resp.Created)
	}
	if resp.CreatedAt != "2026-03-15T10:30:00Z" {
		t.Errorf("expected CreatedAt, got %q", resp.CreatedAt)
	}
}

func TestReadK8sClusterResponse_ExtraFields(t *testing.T) {
	// API may return extra fields — verify they don't cause errors
	data := `{
		"id": "c1",
		"name": "c1",
		"state": "active",
		"created": "2026-01-01T00:00:00Z",
		"extra_field": "ignored",
		"nested": {"key": "value"}
	}`

	var resp readK8sClusterResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatalf("unexpected error with extra fields: %v", err)
	}

	if resp.ID != "c1" {
		t.Errorf("expected ID 'c1', got %q", resp.ID)
	}
}

// ---------------------------------------------------------------------------
// Configurable fields preserved across mapReadResponseToState
// ---------------------------------------------------------------------------

func TestMapReadResponseToState_PreservesAllConfigurableFields(t *testing.T) {
	state := &k8sClusterResourceModel{
		Name:                types.StringValue("my-cluster"),
		KubernetesVersion:   types.StringValue("v1.32.6+rke2r1"),
		EndpointAccess:      types.StringValue("Public and Private"),
		NetworkIsolation:    types.StringValue("Enabled"),
		NginxIngress:        types.StringValue("Disabled"),
		NginxDefaultBackend: types.StringValue("Disabled"),
		NetworkProvider:     types.StringValue("Flannel"),
		SecretsEncryption:   types.StringValue("Enabled"),
		MaxWorkerNodes:      types.Int64Value(10),
		WorkerNodeName:      types.StringValue("workers"),
		WorkerQuantity:      types.Int64Value(5),
		FlavorID:            types.StringValue("flv-abc"),
		VolumeSize:          types.Int64Value(100),
		ClusterType:         types.StringValue("rke2"),
		SnapshotBackup:      types.StringValue("Yes"),
		Autoscale:           types.BoolValue(true),
		AutoscaleMin:        types.Int64Value(2),
		AutoscaleMax:        types.Int64Value(8),
		SecGroupID:          types.StringValue("sg-123"),
		Status:              types.StringValue("provisioning"),
		CreatedAt:           types.StringValue("old"),
	}

	cluster := &readK8sClusterResponse{
		ID:      "new-id",
		Name:    "my-cluster",
		State:   "active",
		Created: "2026-03-15T00:00:00Z",
	}

	mapReadResponseToState(cluster, state)

	// Verify computed fields are updated
	if state.ID.ValueString() != "new-id" {
		t.Errorf("expected ID updated to 'new-id'")
	}
	if state.Status.ValueString() != "active" {
		t.Errorf("expected Status updated to 'active'")
	}
	if state.CreatedAt.ValueString() != "2026-03-15T00:00:00Z" {
		t.Errorf("expected CreatedAt updated")
	}

	// Verify ALL configurable fields are preserved
	if state.KubernetesVersion.ValueString() != "v1.32.6+rke2r1" {
		t.Error("KubernetesVersion was unexpectedly modified")
	}
	if state.EndpointAccess.ValueString() != "Public and Private" {
		t.Error("EndpointAccess was unexpectedly modified")
	}
	if state.NetworkIsolation.ValueString() != "Enabled" {
		t.Error("NetworkIsolation was unexpectedly modified")
	}
	if state.NginxIngress.ValueString() != "Disabled" {
		t.Error("NginxIngress was unexpectedly modified")
	}
	if state.NginxDefaultBackend.ValueString() != "Disabled" {
		t.Error("NginxDefaultBackend was unexpectedly modified")
	}
	if state.NetworkProvider.ValueString() != "Flannel" {
		t.Error("NetworkProvider was unexpectedly modified")
	}
	if state.SecretsEncryption.ValueString() != "Enabled" {
		t.Error("SecretsEncryption was unexpectedly modified")
	}
	if state.MaxWorkerNodes.ValueInt64() != 10 {
		t.Error("MaxWorkerNodes was unexpectedly modified")
	}
	if state.WorkerNodeName.ValueString() != "workers" {
		t.Error("WorkerNodeName was unexpectedly modified")
	}
	if state.WorkerQuantity.ValueInt64() != 5 {
		t.Error("WorkerQuantity was unexpectedly modified")
	}
	if state.FlavorID.ValueString() != "flv-abc" {
		t.Error("FlavorID was unexpectedly modified")
	}
	if state.VolumeSize.ValueInt64() != 100 {
		t.Error("VolumeSize was unexpectedly modified")
	}
	if state.ClusterType.ValueString() != "rke2" {
		t.Error("ClusterType was unexpectedly modified")
	}
	if state.SnapshotBackup.ValueString() != "Yes" {
		t.Error("SnapshotBackup was unexpectedly modified")
	}
	if state.Autoscale.ValueBool() != true {
		t.Error("Autoscale was unexpectedly modified")
	}
	if state.AutoscaleMin.ValueInt64() != 2 {
		t.Error("AutoscaleMin was unexpectedly modified")
	}
	if state.AutoscaleMax.ValueInt64() != 8 {
		t.Error("AutoscaleMax was unexpectedly modified")
	}
	if state.SecGroupID.ValueString() != "sg-123" {
		t.Error("SecGroupID was unexpectedly modified")
	}
}
