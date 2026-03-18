package caas_deployment

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCreateRequest_SharedDeployment(t *testing.T) {
	cpu := float64(0.5)
	body := deploymentCreateRequest{
		Name: "test-app",
		Type: "shared",
		Image: imageCreateRequest{
			Type:      "public",
			Reference: "nginx:latest",
		},
		Resources: resourcesCreateRequest{
			CPU:          &cpu,
			Memory:       "512Mi",
			ReplicaCount: 2,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: true,
			EndpointAccess: "public",
			Ports: []portCreateRequest{
				{Name: "http", Protocol: "HTTP", ContainerPort: 80},
			},
		},
	}

	if body.Name != "test-app" {
		t.Errorf("expected name test-app, got %s", body.Name)
	}
	if body.Type != "shared" {
		t.Errorf("expected type shared, got %s", body.Type)
	}
	if body.Image.Type != "public" {
		t.Errorf("expected image type public, got %s", body.Image.Type)
	}
	if body.Image.Reference != "nginx:latest" {
		t.Errorf("expected image reference nginx:latest, got %s", body.Image.Reference)
	}
	if body.Resources.ReplicaCount != 2 {
		t.Errorf("expected replica_count 2, got %d", body.Resources.ReplicaCount)
	}
	if *body.Resources.CPU != 0.5 {
		t.Errorf("expected cpu 0.5, got %f", *body.Resources.CPU)
	}
	if body.Networking.ExternalAccess != true {
		t.Error("expected external_access to be true")
	}
	if len(body.Networking.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(body.Networking.Ports))
	}
	if body.Networking.Ports[0].ContainerPort != 80 {
		t.Errorf("expected container port 80, got %d", body.Networking.Ports[0].ContainerPort)
	}
}

func TestCreateRequest_DedicatedDeployment(t *testing.T) {
	body := deploymentCreateRequest{
		Name: "dedicated-app",
		Type: "dedicated",
		Image: imageCreateRequest{
			Type:      "private",
			Reference: "myrepo/myimage:v1",
			Secrets:   []string{"my-registry-secret"},
		},
		Resources: resourcesCreateRequest{
			FlavorID:     "flavor-large",
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: false,
		},
	}

	if body.Type != "dedicated" {
		t.Errorf("expected type dedicated, got %s", body.Type)
	}
	if body.Image.Type != "private" {
		t.Errorf("expected image type private, got %s", body.Image.Type)
	}
	if len(body.Image.Secrets) != 1 {
		t.Fatalf("expected 1 image secret, got %d", len(body.Image.Secrets))
	}
	if body.Resources.FlavorID != "flavor-large" {
		t.Errorf("expected flavor_id flavor-large, got %s", body.Resources.FlavorID)
	}
}

func TestCreateRequest_JSON(t *testing.T) {
	cpu := float64(0.5)
	exposed := int64(8080)
	minReplicas := int64(1)
	maxReplicas := int64(5)
	cpuTarget := float64(80)

	body := deploymentCreateRequest{
		Name: "json-app",
		Type: "shared",
		Image: imageCreateRequest{
			Type:      "public",
			Reference: "nginx:latest",
		},
		Resources: resourcesCreateRequest{
			CPU:          &cpu,
			Memory:       "512Mi",
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: true,
			EndpointAccess: "public",
			Ports: []portCreateRequest{
				{Name: "http", Protocol: "HTTP", ContainerPort: 80, ExposedPort: &exposed},
			},
		},
		Autoscaling: &autoscalingCreateRequest{
			Enabled:             true,
			MinReplicas:         &minReplicas,
			MaxReplicas:         &maxReplicas,
			CPUTargetPercentage: &cpuTarget,
		},
		Env: []envCreateRequest{
			{Name: "ENV_VAR", Value: "test"},
		},
		Volumes: []volumeCreateRequest{
			{Name: "data", MountPath: "/data", Size: "1Gi"},
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

	requiredKeys := []string{"name", "image", "resources", "networking"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key '%s' to be present", key)
		}
	}

	if _, ok := raw["autoscaling"]; !ok {
		t.Error("expected 'autoscaling' key")
	}

	// Verify resources uses camelCase
	res := raw["resources"].(map[string]interface{})
	if _, ok := res["replicaCount"]; !ok {
		t.Error("expected 'replicaCount' in resources (camelCase)")
	}

	// Verify networking ports uses camelCase
	net := raw["networking"].(map[string]interface{})
	ports := net["ports"].([]interface{})
	if len(ports) != 1 {
		t.Fatalf("expected 1 port in JSON, got %d", len(ports))
	}
	port := ports[0].(map[string]interface{})
	if _, ok := port["containerPort"]; !ok {
		t.Error("expected 'containerPort' in port (camelCase)")
	}
}

func TestCreateRequest_JSONOmitEmpty(t *testing.T) {
	body := deploymentCreateRequest{
		Name:    "minimal-app",
		Command: []string{}, // API requires command as an array (even empty)
		Image: imageCreateRequest{
			Type:      "public",
			Reference: "nginx:latest",
		},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: false,
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

	// command must always be present (API requires it as an array)
	if cmd, ok := raw["command"]; !ok {
		t.Error("expected 'command' to be present")
	} else if arr, ok := cmd.([]interface{}); !ok || len(arr) != 0 {
		t.Error("expected 'command' to be an empty array")
	}

	// Optional fields with omitempty should be absent
	if _, ok := raw["envSecrets"]; ok {
		t.Error("expected 'envSecrets' to be omitted")
	}
	if _, ok := raw["autoscaling"]; ok {
		t.Error("expected 'autoscaling' to be omitted")
	}
	if _, ok := raw["volumes"]; ok {
		t.Error("expected 'volumes' to be omitted")
	}
}

func TestMapAPIResponseToState(t *testing.T) {
	model := &caasDeploymentModel{}

	apiResp := &deploymentAPIResponse{
		Name: "prod-app",
		Type: "shared",
		DeploymentDetails: deploymentDetailsResp{
			Status:           "Active",
			CreatedAt:        "2024-01-01T00:00:00Z",
			PrivateEndpoints: []string{"http://prod-app:8080"},
			PublicEndpoints:  []string{"http://prod-app.cloud.example.com:8080"},
		},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.ID.ValueString() != "prod-app" {
		t.Errorf("expected ID prod-app, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "prod-app" {
		t.Errorf("expected Name prod-app, got %s", model.Name.ValueString())
	}
	if model.Status.ValueString() != "Active" {
		t.Errorf("expected Status Active, got %s", model.Status.ValueString())
	}
	if model.CreatedAt.ValueString() != "2024-01-01T00:00:00Z" {
		t.Errorf("expected CreatedAt, got %s", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_Empty(t *testing.T) {
	model := &caasDeploymentModel{}

	apiResp := &deploymentAPIResponse{
		Name: "basic-app",
		Type: "shared",
		DeploymentDetails: deploymentDetailsResp{
			Status: "Provisioning",
		},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.Status.ValueString() != "Provisioning" {
		t.Errorf("expected Status Provisioning, got %s", model.Status.ValueString())
	}
	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty CreatedAt, got %s", model.CreatedAt.ValueString())
	}
}

func TestCreateRequest_WithVolumes(t *testing.T) {
	body := deploymentCreateRequest{
		Name:    "vol-app",
		Type:    "shared",
		Command: []string{},
		Image: imageCreateRequest{
			Type:      "public",
			Reference: "nginx:latest",
		},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: false,
		},
		Volumes: []volumeCreateRequest{
			{Name: "data-vol", MountPath: "/data", Size: "1Gi"},
			{Name: "cache-vol", MountPath: "/cache", Size: "256Mi"},
			{Name: "logs-vol", MountPath: "/var/log", Size: "512Mi"},
		},
	}

	if len(body.Volumes) != 3 {
		t.Fatalf("expected 3 volumes, got %d", len(body.Volumes))
	}
	if body.Volumes[0].Name != "data-vol" {
		t.Errorf("expected volume[0] name data-vol, got %s", body.Volumes[0].Name)
	}
	if body.Volumes[1].MountPath != "/cache" {
		t.Errorf("expected volume[1] mountPath /cache, got %s", body.Volumes[1].MountPath)
	}
	if body.Volumes[2].Size != "512Mi" {
		t.Errorf("expected volume[2] size 512Mi, got %s", body.Volumes[2].Size)
	}

	// Verify JSON roundtrip
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	vols := raw["volumes"].([]interface{})
	if len(vols) != 3 {
		t.Errorf("expected 3 volumes in JSON, got %d", len(vols))
	}
	vol0 := vols[0].(map[string]interface{})
	if vol0["mountPath"] != "/data" {
		t.Errorf("expected volume[0] mountPath /data in JSON, got %v", vol0["mountPath"])
	}
}

func TestCreateRequest_WithEnvironmentVars(t *testing.T) {
	body := deploymentCreateRequest{
		Name:    "env-app",
		Type:    "shared",
		Command: []string{},
		Image: imageCreateRequest{
			Type:      "public",
			Reference: "python:3.11",
		},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: false,
		},
		Env: []envCreateRequest{
			{Name: "DATABASE_URL", Value: "postgres://localhost/db"},
			{Name: "APP_ENV", Value: "production"},
			{Name: "LOG_LEVEL", Value: "info"},
		},
	}

	if len(body.Env) != 3 {
		t.Fatalf("expected 3 env vars, got %d", len(body.Env))
	}
	if body.Env[0].Name != "DATABASE_URL" {
		t.Errorf("expected env[0] name DATABASE_URL, got %s", body.Env[0].Name)
	}
	if body.Env[1].Value != "production" {
		t.Errorf("expected env[1] value production, got %s", body.Env[1].Value)
	}

	// Verify JSON
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	envs := raw["env"].([]interface{})
	if len(envs) != 3 {
		t.Errorf("expected 3 env vars in JSON, got %d", len(envs))
	}
}

func TestCreateRequest_CommandAlwaysArray(t *testing.T) {
	// When command is set
	body1 := deploymentCreateRequest{
		Name:    "cmd-app",
		Command: []string{"python", "-m", "app"},
		Image:   imageCreateRequest{Type: "public", Reference: "python:3.11"},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{},
	}

	data1, err := json.Marshal(body1)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw1 map[string]interface{}
	if err := json.Unmarshal(data1, &raw1); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	cmd1 := raw1["command"].([]interface{})
	if len(cmd1) != 3 {
		t.Errorf("expected 3 command elements, got %d", len(cmd1))
	}
	if cmd1[0] != "python" {
		t.Errorf("expected command[0] python, got %v", cmd1[0])
	}

	// When command is empty array (API requires it present)
	body2 := deploymentCreateRequest{
		Name:    "no-cmd-app",
		Command: []string{},
		Image:   imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{},
	}

	data2, err := json.Marshal(body2)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw2 map[string]interface{}
	if err := json.Unmarshal(data2, &raw2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// command must be present even if empty
	cmd2, ok := raw2["command"]
	if !ok {
		t.Fatal("expected 'command' key to be present even when empty array")
	}
	arr, ok := cmd2.([]interface{})
	if !ok {
		t.Fatalf("expected 'command' to be an array, got %T", cmd2)
	}
	if len(arr) != 0 {
		t.Errorf("expected empty command array, got %d elements", len(arr))
	}
}

func TestCreateRequest_XForwardedForAlwaysBool(t *testing.T) {
	// XForwardedFor = true
	body1 := deploymentCreateRequest{
		Name:    "xff-true-app",
		Command: []string{},
		Image:   imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: true,
			XForwardedFor:  true,
		},
	}

	data1, err := json.Marshal(body1)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw1 map[string]interface{}
	if err := json.Unmarshal(data1, &raw1); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	net1 := raw1["networking"].(map[string]interface{})
	xff1, ok := net1["xForwardedFor"]
	if !ok {
		t.Fatal("expected 'xForwardedFor' to be present in networking")
	}
	if xff1 != true {
		t.Errorf("expected xForwardedFor true, got %v", xff1)
	}

	// XForwardedFor = false (must be sent as false, not omitted)
	body2 := deploymentCreateRequest{
		Name:    "xff-false-app",
		Command: []string{},
		Image:   imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess: false,
			XForwardedFor:  false,
		},
	}

	data2, err := json.Marshal(body2)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw2 map[string]interface{}
	if err := json.Unmarshal(data2, &raw2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	net2 := raw2["networking"].(map[string]interface{})
	xff2, ok := net2["xForwardedFor"]
	if !ok {
		t.Fatal("expected 'xForwardedFor' to be present even when false (API requires it)")
	}
	if xff2 != false {
		t.Errorf("expected xForwardedFor false, got %v", xff2)
	}
}

func TestNewResource(t *testing.T) {
	r := NewResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
}

// --- buildCreateRequest tests (from Terraform model) ---

func TestBuildCreateRequest_MinimalShared(t *testing.T) {
	ctx := context.Background()

	plan := &caasDeploymentModel{
		Name:        types.StringValue("minimal-shared"),
		Type:        types.StringValue("shared"),
		Command:     types.ListNull(types.StringType),
		EnvSecrets:  types.ListNull(types.StringType),
		Image:       buildImageObject("public", "nginx:latest", nil),
		Resources:   buildResourcesObject(nil, "", 1, ""),
		Networking:  buildNetworkingObject(false, "", nil, false, nil, "", "", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Name != "minimal-shared" {
		t.Errorf("expected name minimal-shared, got %s", body.Name)
	}
	if body.Type != "shared" {
		t.Errorf("expected type shared, got %s", body.Type)
	}
	if body.Image.Type != "public" {
		t.Errorf("expected image type public, got %s", body.Image.Type)
	}
	if body.Image.Reference != "nginx:latest" {
		t.Errorf("expected image ref nginx:latest, got %s", body.Image.Reference)
	}
	if body.Resources.ReplicaCount != 1 {
		t.Errorf("expected replica_count 1, got %d", body.Resources.ReplicaCount)
	}
	if body.Resources.CPU != nil {
		t.Errorf("expected nil CPU, got %v", body.Resources.CPU)
	}
	if body.Resources.Memory != "" {
		t.Errorf("expected empty memory, got %s", body.Resources.Memory)
	}
	if body.Resources.FlavorID != "" {
		t.Errorf("expected empty flavor_id, got %s", body.Resources.FlavorID)
	}
	if body.Networking.ExternalAccess != false {
		t.Error("expected external_access false")
	}
	if len(body.Command) != 0 {
		t.Errorf("expected empty command, got %v", body.Command)
	}
	if body.Autoscaling != nil {
		t.Errorf("expected nil autoscaling, got %+v", body.Autoscaling)
	}
	if len(body.EnvSecrets) != 0 {
		t.Errorf("expected no env secrets, got %v", body.EnvSecrets)
	}
	if len(body.Env) != 0 {
		t.Errorf("expected no env vars, got %v", body.Env)
	}
	if len(body.Volumes) != 0 {
		t.Errorf("expected no volumes, got %v", body.Volumes)
	}
}

func TestBuildCreateRequest_FullDedicated(t *testing.T) {
	ctx := context.Background()

	plan := &caasDeploymentModel{
		Name:        types.StringValue("dedicated-full"),
		Type:        types.StringValue("dedicated"),
		Command:     types.ListNull(types.StringType),
		EnvSecrets:  types.ListNull(types.StringType),
		Image:       buildImageObject("private", "myrepo/app:v2", []string{"reg-secret-1"}),
		Resources:   buildResourcesObject(nil, "", 3, "flavor-xl"),
		Networking:  buildNetworkingObject(true, "public", nil, true, nil, "", "", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Type != "dedicated" {
		t.Errorf("expected type dedicated, got %s", body.Type)
	}
	if body.Resources.FlavorID != "flavor-xl" {
		t.Errorf("expected flavor_id flavor-xl, got %s", body.Resources.FlavorID)
	}
	if body.Resources.ReplicaCount != 3 {
		t.Errorf("expected replica_count 3, got %d", body.Resources.ReplicaCount)
	}
	if body.Resources.CPU != nil {
		t.Errorf("expected nil CPU for dedicated, got %v", body.Resources.CPU)
	}
	if body.Image.Type != "private" {
		t.Errorf("expected private image, got %s", body.Image.Type)
	}
	if len(body.Image.Secrets) != 1 || body.Image.Secrets[0] != "reg-secret-1" {
		t.Errorf("expected image secrets [reg-secret-1], got %v", body.Image.Secrets)
	}
	if body.Networking.ExternalAccess != true {
		t.Error("expected external_access true")
	}
	if body.Networking.XForwardedFor != true {
		t.Error("expected xForwardedFor true")
	}
}

func TestBuildCreateRequest_WithAutoscaling(t *testing.T) {
	ctx := context.Background()

	plan := &caasDeploymentModel{
		Name:       types.StringValue("autoscale-app"),
		Type:       types.StringValue("shared"),
		Command:    types.ListNull(types.StringType),
		EnvSecrets: types.ListNull(types.StringType),
		Image:      buildImageObject("public", "nginx:latest", nil),
		Resources:  buildResourcesObject(ptrFloat64(1.0), "1Gi", 2, ""),
		Networking: buildNetworkingObject(false, "", nil, false, nil, "", "", nil),
		Autoscaling: buildAutoscalingObject(true, ptrInt64(1), ptrInt64(10),
			ptrFloat64(75), ptrFloat64(80)),
		Env:    types.ListNull(envObjectType()),
		Volume: types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Autoscaling == nil {
		t.Fatal("expected autoscaling to be set")
	}
	if body.Autoscaling.Enabled != true {
		t.Error("expected autoscaling enabled=true")
	}
	if body.Autoscaling.MinReplicas == nil || *body.Autoscaling.MinReplicas != 1 {
		t.Errorf("expected minReplicas 1, got %v", body.Autoscaling.MinReplicas)
	}
	if body.Autoscaling.MaxReplicas == nil || *body.Autoscaling.MaxReplicas != 10 {
		t.Errorf("expected maxReplicas 10, got %v", body.Autoscaling.MaxReplicas)
	}
	if body.Autoscaling.CPUTargetPercentage == nil || *body.Autoscaling.CPUTargetPercentage != 75 {
		t.Errorf("expected cpuTarget 75, got %v", body.Autoscaling.CPUTargetPercentage)
	}
	if body.Autoscaling.MemoryTargetPercentage == nil || *body.Autoscaling.MemoryTargetPercentage != 80 {
		t.Errorf("expected memTarget 80, got %v", body.Autoscaling.MemoryTargetPercentage)
	}
}

func TestBuildCreateRequest_WithNetworkingCIDR(t *testing.T) {
	ctx := context.Background()
	useExisting := true

	plan := &caasDeploymentModel{
		Name:       types.StringValue("net-app"),
		Type:       types.StringValue("shared"),
		Command:    types.ListNull(types.StringType),
		EnvSecrets: types.ListNull(types.StringType),
		Image:      buildImageObject("public", "nginx:latest", nil),
		Resources:  buildResourcesObject(ptrFloat64(0.5), "512Mi", 1, ""),
		Networking: buildNetworkingObject(true, "protected",
			[]string{"10.0.0.0/24", "192.168.1.0/24"},
			false, &useExisting, "net-12345", "", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Networking.EndpointAccess != "protected" {
		t.Errorf("expected endpoint access protected, got %s", body.Networking.EndpointAccess)
	}
	if len(body.Networking.CIDRBlock) != 2 {
		t.Fatalf("expected 2 CIDR blocks, got %d", len(body.Networking.CIDRBlock))
	}
	if body.Networking.CIDRBlock[0] != "10.0.0.0/24" {
		t.Errorf("expected CIDR[0] 10.0.0.0/24, got %s", body.Networking.CIDRBlock[0])
	}
	if body.Networking.CIDRBlock[1] != "192.168.1.0/24" {
		t.Errorf("expected CIDR[1] 192.168.1.0/24, got %s", body.Networking.CIDRBlock[1])
	}
	if body.Networking.UseExistingNetwork == nil || *body.Networking.UseExistingNetwork != true {
		t.Error("expected useExistingNetwork true")
	}
	if body.Networking.NetworkID != "net-12345" {
		t.Errorf("expected networkId net-12345, got %s", body.Networking.NetworkID)
	}
}

func TestBuildCreateRequest_WithMultiplePorts(t *testing.T) {
	ctx := context.Background()

	ports := []portCreateRequest{
		{Name: "http", Protocol: "HTTP", ContainerPort: 80, ExposedPort: ptrInt64(8080)},
		{Name: "https", Protocol: "HTTPS", ContainerPort: 443, ExposedPort: ptrInt64(8443)},
		{Name: "grpc", Protocol: "TCP", ContainerPort: 9090},
	}

	plan := &caasDeploymentModel{
		Name:        types.StringValue("multiport-app"),
		Type:        types.StringValue("shared"),
		Command:     types.ListNull(types.StringType),
		EnvSecrets:  types.ListNull(types.StringType),
		Image:       buildImageObject("public", "myapp:latest", nil),
		Resources:   buildResourcesObject(ptrFloat64(1.0), "1Gi", 2, ""),
		Networking:  buildNetworkingObjectWithPorts(true, "public", nil, true, nil, "", "", ports),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if len(body.Networking.Ports) != 3 {
		t.Fatalf("expected 3 ports, got %d", len(body.Networking.Ports))
	}
	if body.Networking.Ports[0].Name != "http" {
		t.Errorf("expected port[0] name http, got %s", body.Networking.Ports[0].Name)
	}
	if body.Networking.Ports[0].ExposedPort == nil || *body.Networking.Ports[0].ExposedPort != 8080 {
		t.Errorf("expected port[0] exposedPort 8080, got %v", body.Networking.Ports[0].ExposedPort)
	}
	if body.Networking.Ports[1].Protocol != "HTTPS" {
		t.Errorf("expected port[1] protocol HTTPS, got %s", body.Networking.Ports[1].Protocol)
	}
	if body.Networking.Ports[2].ExposedPort != nil {
		t.Errorf("expected port[2] nil exposedPort, got %v", body.Networking.Ports[2].ExposedPort)
	}
	if body.Networking.Ports[2].ContainerPort != 9090 {
		t.Errorf("expected port[2] containerPort 9090, got %d", body.Networking.Ports[2].ContainerPort)
	}
}

func TestBuildCreateRequest_WithEnvSecrets(t *testing.T) {
	ctx := context.Background()

	secretsList, _ := types.ListValueFrom(ctx, types.StringType, []string{"db-secret", "api-key-secret", "tls-secret"})

	plan := &caasDeploymentModel{
		Name:        types.StringValue("secrets-app"),
		Type:        types.StringValue("shared"),
		Command:     types.ListNull(types.StringType),
		EnvSecrets:  secretsList,
		Image:       buildImageObject("public", "myapp:latest", nil),
		Resources:   buildResourcesObject(ptrFloat64(0.5), "256Mi", 1, ""),
		Networking:  buildNetworkingObject(false, "", nil, false, nil, "", "", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if len(body.EnvSecrets) != 3 {
		t.Fatalf("expected 3 env secrets, got %d", len(body.EnvSecrets))
	}
	if body.EnvSecrets[0] != "db-secret" {
		t.Errorf("expected envSecrets[0] db-secret, got %s", body.EnvSecrets[0])
	}
	if body.EnvSecrets[1] != "api-key-secret" {
		t.Errorf("expected envSecrets[1] api-key-secret, got %s", body.EnvSecrets[1])
	}
	if body.EnvSecrets[2] != "tls-secret" {
		t.Errorf("expected envSecrets[2] tls-secret, got %s", body.EnvSecrets[2])
	}
}

func TestBuildCreateRequest_WithImageSecrets(t *testing.T) {
	ctx := context.Background()

	plan := &caasDeploymentModel{
		Name:        types.StringValue("private-img-app"),
		Type:        types.StringValue("shared"),
		Command:     types.ListNull(types.StringType),
		EnvSecrets:  types.ListNull(types.StringType),
		Image:       buildImageObject("private", "registry.example.com/myapp:v3", []string{"harbor-cred", "docker-cred"}),
		Resources:   buildResourcesObject(ptrFloat64(0.5), "512Mi", 1, ""),
		Networking:  buildNetworkingObject(false, "", nil, false, nil, "", "", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Image.Type != "private" {
		t.Errorf("expected image type private, got %s", body.Image.Type)
	}
	if body.Image.Reference != "registry.example.com/myapp:v3" {
		t.Errorf("expected image ref, got %s", body.Image.Reference)
	}
	if len(body.Image.Secrets) != 2 {
		t.Fatalf("expected 2 image secrets, got %d", len(body.Image.Secrets))
	}
	if body.Image.Secrets[0] != "harbor-cred" {
		t.Errorf("expected secret[0] harbor-cred, got %s", body.Image.Secrets[0])
	}
	if body.Image.Secrets[1] != "docker-cred" {
		t.Errorf("expected secret[1] docker-cred, got %s", body.Image.Secrets[1])
	}
}

func TestBuildCreateRequest_NullOptionalFields(t *testing.T) {
	ctx := context.Background()

	plan := &caasDeploymentModel{
		Name:        types.StringValue("null-opts"),
		Type:        types.StringValue("shared"),
		Command:     types.ListNull(types.StringType),
		EnvSecrets:  types.ListNull(types.StringType),
		Image:       buildImageObject("public", "nginx:latest", nil),
		Resources:   buildResourcesObject(nil, "", 1, ""),
		Networking:  buildNetworkingObject(false, "", nil, false, nil, "", "", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	// Verify optional fields are empty/nil
	if body.Autoscaling != nil {
		t.Errorf("expected nil autoscaling, got %+v", body.Autoscaling)
	}
	if len(body.EnvSecrets) != 0 {
		t.Errorf("expected empty envSecrets, got %v", body.EnvSecrets)
	}
	if len(body.Env) != 0 {
		t.Errorf("expected empty env, got %v", body.Env)
	}
	if len(body.Volumes) != 0 {
		t.Errorf("expected empty volumes, got %v", body.Volumes)
	}
	if body.Resources.CPU != nil {
		t.Errorf("expected nil CPU, got %v", body.Resources.CPU)
	}
	if body.Resources.FlavorID != "" {
		t.Errorf("expected empty flavorId, got %s", body.Resources.FlavorID)
	}
	if body.Resources.Memory != "" {
		t.Errorf("expected empty memory, got %s", body.Resources.Memory)
	}

	// Verify JSON omits empty optional fields
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := raw["autoscaling"]; ok {
		t.Error("expected autoscaling omitted from JSON")
	}
	if _, ok := raw["envSecrets"]; ok {
		t.Error("expected envSecrets omitted from JSON")
	}
	if _, ok := raw["env"]; ok {
		t.Error("expected env omitted from JSON")
	}
	if _, ok := raw["volumes"]; ok {
		t.Error("expected volumes omitted from JSON")
	}
}

// --- mapAPIResponseToState tests ---

func TestMapAPIResponseToState_Endpoints(t *testing.T) {
	model := &caasDeploymentModel{}
	apiResp := &deploymentAPIResponse{
		Name: "ep-app",
		Type: "shared",
		DeploymentDetails: deploymentDetailsResp{
			Status:    "Active",
			CreatedAt: "2024-06-15T12:00:00Z",
			PrivateEndpoints: []string{
				"http://ep-app:8080",
				"http://ep-app:9090",
			},
			PublicEndpoints: []string{
				"https://ep-app.cloud.example.com:8080",
				"https://ep-app.cloud.example.com:9090",
				"https://ep-app.cloud.example.com:3000",
			},
		},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	// Check private endpoints list
	var privateEps []string
	model.PrivateEndpoints.ElementsAs(context.Background(), &privateEps, false)
	if len(privateEps) != 2 {
		t.Fatalf("expected 2 private endpoints, got %d", len(privateEps))
	}
	if privateEps[0] != "http://ep-app:8080" {
		t.Errorf("expected private[0], got %s", privateEps[0])
	}
	if privateEps[1] != "http://ep-app:9090" {
		t.Errorf("expected private[1], got %s", privateEps[1])
	}

	// Check public endpoints list
	var publicEps []string
	model.PublicEndpoints.ElementsAs(context.Background(), &publicEps, false)
	if len(publicEps) != 3 {
		t.Fatalf("expected 3 public endpoints, got %d", len(publicEps))
	}
	if publicEps[2] != "https://ep-app.cloud.example.com:3000" {
		t.Errorf("expected public[2], got %s", publicEps[2])
	}
}

func TestMapAPIResponseToState_EmptyEndpoints(t *testing.T) {
	model := &caasDeploymentModel{}
	apiResp := &deploymentAPIResponse{
		Name: "no-ep-app",
		Type: "shared",
		DeploymentDetails: deploymentDetailsResp{
			Status:           "Active",
			PrivateEndpoints: []string{},
			PublicEndpoints:  []string{},
		},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	// Empty endpoints should produce empty lists, not null
	if model.PrivateEndpoints.IsNull() {
		t.Error("expected non-null private_endpoints (empty list, not null)")
	}
	if model.PublicEndpoints.IsNull() {
		t.Error("expected non-null public_endpoints (empty list, not null)")
	}

	var privateEps []string
	model.PrivateEndpoints.ElementsAs(context.Background(), &privateEps, false)
	if len(privateEps) != 0 {
		t.Errorf("expected 0 private endpoints, got %d", len(privateEps))
	}

	var publicEps []string
	model.PublicEndpoints.ElementsAs(context.Background(), &publicEps, false)
	if len(publicEps) != 0 {
		t.Errorf("expected 0 public endpoints, got %d", len(publicEps))
	}
}

func TestMapAPIResponseToState_NilEndpoints(t *testing.T) {
	model := &caasDeploymentModel{}
	apiResp := &deploymentAPIResponse{
		Name: "nil-ep-app",
		Type: "shared",
		DeploymentDetails: deploymentDetailsResp{
			Status:           "Provisioning",
			PrivateEndpoints: nil,
			PublicEndpoints:  nil,
		},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	// nil endpoints should also produce empty lists, not null
	if model.PrivateEndpoints.IsNull() {
		t.Error("expected non-null private_endpoints even with nil input")
	}
	if model.PublicEndpoints.IsNull() {
		t.Error("expected non-null public_endpoints even with nil input")
	}
}

func TestMapAPIResponseToState_TypePreserved(t *testing.T) {
	tests := []struct {
		name     string
		respType string
	}{
		{"shared type", "shared"},
		{"dedicated type", "dedicated"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			model := &caasDeploymentModel{}
			apiResp := &deploymentAPIResponse{
				Name: "type-app",
				Type: tc.respType,
				DeploymentDetails: deploymentDetailsResp{
					Status: "Active",
				},
			}
			mapAPIResponseToState(context.Background(), model, apiResp)
			if model.Type.ValueString() != tc.respType {
				t.Errorf("expected type %s, got %s", tc.respType, model.Type.ValueString())
			}
		})
	}
}

func TestMapAPIResponseToState_AllFieldsEmpty(t *testing.T) {
	model := &caasDeploymentModel{}
	apiResp := &deploymentAPIResponse{
		Name: "empty-app",
		Type: "",
		DeploymentDetails: deploymentDetailsResp{
			Status:    "",
			CreatedAt: "",
		},
	}

	mapAPIResponseToState(context.Background(), model, apiResp)

	if model.ID.ValueString() != "empty-app" {
		t.Errorf("expected ID empty-app, got %s", model.ID.ValueString())
	}
	if model.Name.ValueString() != "empty-app" {
		t.Errorf("expected Name empty-app, got %s", model.Name.ValueString())
	}
	// Empty type should not override
	if model.Type.ValueString() != "" && !model.Type.IsNull() {
		// Type is not set when empty — this is expected behavior
	}
	if model.Status.ValueString() != "" {
		t.Errorf("expected empty status, got %s", model.Status.ValueString())
	}
	if model.CreatedAt.ValueString() != "" {
		t.Errorf("expected empty created_at, got %s", model.CreatedAt.ValueString())
	}
}

func TestMapAPIResponseToState_StatusVariants(t *testing.T) {
	statuses := []string{"Active", "Provisioning", "Error", "DeletionFailed", "Updating", "OutOfSync"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			model := &caasDeploymentModel{}
			apiResp := &deploymentAPIResponse{
				Name: "status-app",
				Type: "shared",
				DeploymentDetails: deploymentDetailsResp{
					Status: status,
				},
			}
			mapAPIResponseToState(context.Background(), model, apiResp)
			if model.Status.ValueString() != status {
				t.Errorf("expected status %s, got %s", status, model.Status.ValueString())
			}
		})
	}
}

// --- JSON serialization tests ---

func TestCreateRequest_NetworkingJSON(t *testing.T) {
	useExisting := true
	body := deploymentCreateRequest{
		Name:    "net-json-app",
		Command: []string{},
		Image:   imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources: resourcesCreateRequest{
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{
			ExternalAccess:       true,
			EndpointAccess:       "protected",
			CIDRBlock:            []string{"10.0.0.0/16"},
			XForwardedFor:        true,
			UseExistingNetwork:   &useExisting,
			NetworkID:            "vpc-abc123",
			CreateNewNetworkCIDR: "172.16.0.0/16",
			Ports: []portCreateRequest{
				{Name: "http", Protocol: "HTTP", ContainerPort: 80},
			},
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

	net := raw["networking"].(map[string]interface{})

	// Verify all camelCase keys exist
	expectedKeys := []string{
		"externalAccess", "endpointAccess", "xForwardedFor",
		"cidrBlock", "useExistingNetwork", "networkId",
		"createNewNetworkCidr", "ports",
	}
	for _, key := range expectedKeys {
		if _, ok := net[key]; !ok {
			t.Errorf("expected JSON key '%s' in networking", key)
		}
	}

	// Verify values
	if net["externalAccess"] != true {
		t.Errorf("expected externalAccess true, got %v", net["externalAccess"])
	}
	if net["endpointAccess"] != "protected" {
		t.Errorf("expected endpointAccess protected, got %v", net["endpointAccess"])
	}
	if net["xForwardedFor"] != true {
		t.Errorf("expected xForwardedFor true, got %v", net["xForwardedFor"])
	}
	if net["networkId"] != "vpc-abc123" {
		t.Errorf("expected networkId vpc-abc123, got %v", net["networkId"])
	}
	if net["createNewNetworkCidr"] != "172.16.0.0/16" {
		t.Errorf("expected createNewNetworkCidr 172.16.0.0/16, got %v", net["createNewNetworkCidr"])
	}
	if net["useExistingNetwork"] != true {
		t.Errorf("expected useExistingNetwork true, got %v", net["useExistingNetwork"])
	}

	cidrs := net["cidrBlock"].([]interface{})
	if len(cidrs) != 1 || cidrs[0] != "10.0.0.0/16" {
		t.Errorf("expected cidrBlock [10.0.0.0/16], got %v", cidrs)
	}
}

func TestCreateRequest_AutoscalingJSON(t *testing.T) {
	minR := int64(2)
	maxR := int64(8)
	cpuT := float64(70)
	memT := float64(85)

	body := deploymentCreateRequest{
		Name:    "autoscale-json-app",
		Command: []string{},
		Image:   imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources: resourcesCreateRequest{
			ReplicaCount: 2,
		},
		Networking: networkingCreateRequest{},
		Autoscaling: &autoscalingCreateRequest{
			Enabled:                true,
			MinReplicas:            &minR,
			MaxReplicas:            &maxR,
			CPUTargetPercentage:    &cpuT,
			MemoryTargetPercentage: &memT,
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

	as := raw["autoscaling"].(map[string]interface{})

	expectedKeys := []string{"enabled", "minReplicas", "maxReplicas", "cpuTargetPercentage", "memoryTargetPercentage"}
	for _, key := range expectedKeys {
		if _, ok := as[key]; !ok {
			t.Errorf("expected JSON key '%s' in autoscaling", key)
		}
	}

	if as["enabled"] != true {
		t.Errorf("expected enabled true, got %v", as["enabled"])
	}
	if as["minReplicas"] != float64(2) {
		t.Errorf("expected minReplicas 2, got %v", as["minReplicas"])
	}
	if as["maxReplicas"] != float64(8) {
		t.Errorf("expected maxReplicas 8, got %v", as["maxReplicas"])
	}
	if as["cpuTargetPercentage"] != float64(70) {
		t.Errorf("expected cpuTargetPercentage 70, got %v", as["cpuTargetPercentage"])
	}
	if as["memoryTargetPercentage"] != float64(85) {
		t.Errorf("expected memoryTargetPercentage 85, got %v", as["memoryTargetPercentage"])
	}
}

func TestCreateRequest_ResourcesJSONDedicated(t *testing.T) {
	body := deploymentCreateRequest{
		Name:    "ded-json-app",
		Command: []string{},
		Image:   imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources: resourcesCreateRequest{
			FlavorID:     "flavor-gpu-large",
			ReplicaCount: 1,
		},
		Networking: networkingCreateRequest{},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	res := raw["resources"].(map[string]interface{})

	// flavorId should be present
	if res["flavorId"] != "flavor-gpu-large" {
		t.Errorf("expected flavorId flavor-gpu-large, got %v", res["flavorId"])
	}
	// cpu and memory should be omitted (omitempty)
	if _, ok := res["cpu"]; ok {
		t.Error("expected cpu to be omitted for dedicated")
	}
	if _, ok := res["memory"]; ok {
		t.Error("expected memory to be omitted for dedicated")
	}
	// replicaCount is always present (no omitempty)
	if res["replicaCount"] != float64(1) {
		t.Errorf("expected replicaCount 1, got %v", res["replicaCount"])
	}
}

func TestCreateRequest_ResourcesJSONShared(t *testing.T) {
	cpu := float64(2.0)
	body := deploymentCreateRequest{
		Name:    "shared-res-json",
		Command: []string{},
		Image:   imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources: resourcesCreateRequest{
			CPU:          &cpu,
			Memory:       "2Gi",
			ReplicaCount: 4,
		},
		Networking: networkingCreateRequest{},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	res := raw["resources"].(map[string]interface{})

	if res["cpu"] != float64(2.0) {
		t.Errorf("expected cpu 2.0, got %v", res["cpu"])
	}
	if res["memory"] != "2Gi" {
		t.Errorf("expected memory 2Gi, got %v", res["memory"])
	}
	if res["replicaCount"] != float64(4) {
		t.Errorf("expected replicaCount 4, got %v", res["replicaCount"])
	}
	// flavorId should be omitted for shared
	if _, ok := res["flavorId"]; ok {
		t.Error("expected flavorId to be omitted for shared")
	}
}

func TestCreateRequest_EnvSecretsJSON(t *testing.T) {
	body := deploymentCreateRequest{
		Name:       "envsecrets-json",
		Command:    []string{},
		EnvSecrets: []string{"secret-a", "secret-b"},
		Image:      imageCreateRequest{Type: "public", Reference: "nginx:latest"},
		Resources:  resourcesCreateRequest{ReplicaCount: 1},
		Networking: networkingCreateRequest{},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	secrets := raw["envSecrets"].([]interface{})
	if len(secrets) != 2 {
		t.Fatalf("expected 2 envSecrets, got %d", len(secrets))
	}
	if secrets[0] != "secret-a" {
		t.Errorf("expected envSecrets[0] secret-a, got %v", secrets[0])
	}
	if secrets[1] != "secret-b" {
		t.Errorf("expected envSecrets[1] secret-b, got %v", secrets[1])
	}
}

func TestBuildCreateRequest_WithCommandFromModel(t *testing.T) {
	ctx := context.Background()

	cmdList, _ := types.ListValueFrom(ctx, types.StringType, []string{"python", "-u", "main.py"})

	plan := &caasDeploymentModel{
		Name:        types.StringValue("cmd-model-app"),
		Type:        types.StringValue("shared"),
		Command:     cmdList,
		EnvSecrets:  types.ListNull(types.StringType),
		Image:       buildImageObject("public", "python:3.11", nil),
		Resources:   buildResourcesObject(ptrFloat64(0.5), "512Mi", 1, ""),
		Networking:  buildNetworkingObject(false, "", nil, false, nil, "", "", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if len(body.Command) != 3 {
		t.Fatalf("expected 3 command elements, got %d", len(body.Command))
	}
	if body.Command[0] != "python" {
		t.Errorf("expected command[0] python, got %s", body.Command[0])
	}
	if body.Command[1] != "-u" {
		t.Errorf("expected command[1] -u, got %s", body.Command[1])
	}
	if body.Command[2] != "main.py" {
		t.Errorf("expected command[2] main.py, got %s", body.Command[2])
	}
}

func TestBuildCreateRequest_NewNetworkCIDR(t *testing.T) {
	ctx := context.Background()
	useExisting := false

	plan := &caasDeploymentModel{
		Name:        types.StringValue("new-net-app"),
		Type:        types.StringValue("shared"),
		Command:     types.ListNull(types.StringType),
		EnvSecrets:  types.ListNull(types.StringType),
		Image:       buildImageObject("public", "nginx:latest", nil),
		Resources:   buildResourcesObject(ptrFloat64(0.5), "512Mi", 1, ""),
		Networking:  buildNetworkingObject(false, "", nil, false, &useExisting, "", "10.244.0.0/16", nil),
		Autoscaling: types.ObjectNull(autoscalingAttrTypes()),
		Env:         types.ListNull(envObjectType()),
		Volume:      types.ListNull(volumeObjectType()),
	}

	body := buildCreateRequest(ctx, plan)

	if body.Networking.UseExistingNetwork == nil || *body.Networking.UseExistingNetwork != false {
		t.Error("expected useExistingNetwork false")
	}
	if body.Networking.CreateNewNetworkCIDR != "10.244.0.0/16" {
		t.Errorf("expected createNewNetworkCidr 10.244.0.0/16, got %s", body.Networking.CreateNewNetworkCIDR)
	}
	if body.Networking.NetworkID != "" {
		t.Errorf("expected empty networkId, got %s", body.Networking.NetworkID)
	}
}

func TestDeploymentAPIResponse_JSONDeserialization(t *testing.T) {
	jsonData := `{
		"name": "deserialized-app",
		"type": "dedicated",
		"deploymentDetails": {
			"status": "Active",
			"createdAt": "2024-12-01T10:30:00Z",
			"privateEndpoints": ["http://deserialized-app:80"],
			"publicEndpoints": ["https://deserialized-app.example.com"]
		}
	}`

	var resp deploymentAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Name != "deserialized-app" {
		t.Errorf("expected name deserialized-app, got %s", resp.Name)
	}
	if resp.Type != "dedicated" {
		t.Errorf("expected type dedicated, got %s", resp.Type)
	}
	if resp.DeploymentDetails.Status != "Active" {
		t.Errorf("expected status Active, got %s", resp.DeploymentDetails.Status)
	}
	if resp.DeploymentDetails.CreatedAt != "2024-12-01T10:30:00Z" {
		t.Errorf("expected createdAt, got %s", resp.DeploymentDetails.CreatedAt)
	}
	if len(resp.DeploymentDetails.PrivateEndpoints) != 1 {
		t.Fatalf("expected 1 private endpoint, got %d", len(resp.DeploymentDetails.PrivateEndpoints))
	}
	if len(resp.DeploymentDetails.PublicEndpoints) != 1 {
		t.Fatalf("expected 1 public endpoint, got %d", len(resp.DeploymentDetails.PublicEndpoints))
	}
}

func TestDeploymentAPIResponse_EmptyJSON(t *testing.T) {
	jsonData := `{
		"name": "empty-resp",
		"type": "",
		"deploymentDetails": {}
	}`

	var resp deploymentAPIResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if resp.Name != "empty-resp" {
		t.Errorf("expected name empty-resp, got %s", resp.Name)
	}
	if resp.DeploymentDetails.Status != "" {
		t.Errorf("expected empty status, got %s", resp.DeploymentDetails.Status)
	}
	if resp.DeploymentDetails.PrivateEndpoints != nil {
		t.Errorf("expected nil privateEndpoints, got %v", resp.DeploymentDetails.PrivateEndpoints)
	}
}

// --- Helper functions for building Terraform type objects ---

func ptrFloat64(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64       { return &v }

func autoscalingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":                    types.BoolType,
		"min_replicas":               types.Int64Type,
		"max_replicas":               types.Int64Type,
		"cpu_target_percentage":      types.Float64Type,
		"memory_target_percentage":   types.Float64Type,
	}
}

func envObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":  types.StringType,
			"value": types.StringType,
		},
	}
}

func volumeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":       types.StringType,
			"mount_path": types.StringType,
			"size":       types.StringType,
		},
	}
}

func portAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":           types.StringType,
		"protocol":       types.StringType,
		"container_port": types.Int64Type,
		"exposed_port":   types.Int64Type,
	}
}

func buildImageObject(imgType, reference string, secrets []string) types.Object {
	attrTypes := map[string]attr.Type{
		"type":      types.StringType,
		"reference": types.StringType,
		"secrets":   types.ListType{ElemType: types.StringType},
	}

	secretsVal := types.ListNull(types.StringType)
	if secrets != nil {
		secretsVal, _ = types.ListValueFrom(context.Background(), types.StringType, secrets)
	}

	obj, _ := types.ObjectValue(attrTypes, map[string]attr.Value{
		"type":      types.StringValue(imgType),
		"reference": types.StringValue(reference),
		"secrets":   secretsVal,
	})
	return obj
}

func buildResourcesObject(cpu *float64, memory string, replicaCount int64, flavorID string) types.Object {
	attrTypes := map[string]attr.Type{
		"cpu":           types.Float64Type,
		"memory":        types.StringType,
		"replica_count": types.Int64Type,
		"flavor_id":     types.StringType,
	}

	cpuVal := types.Float64Null()
	if cpu != nil {
		cpuVal = types.Float64Value(*cpu)
	}

	memVal := types.StringNull()
	if memory != "" {
		memVal = types.StringValue(memory)
	}

	flavorVal := types.StringNull()
	if flavorID != "" {
		flavorVal = types.StringValue(flavorID)
	}

	obj, _ := types.ObjectValue(attrTypes, map[string]attr.Value{
		"cpu":           cpuVal,
		"memory":        memVal,
		"replica_count": types.Int64Value(replicaCount),
		"flavor_id":     flavorVal,
	})
	return obj
}

func buildNetworkingObject(extAccess bool, endpointAccess string, cidrBlocks []string, xff bool, useExisting *bool, networkID, newNetworkCIDR string, ports []portCreateRequest) types.Object {
	return buildNetworkingObjectWithPorts(extAccess, endpointAccess, cidrBlocks, xff, useExisting, networkID, newNetworkCIDR, ports)
}

func buildNetworkingObjectWithPorts(extAccess bool, endpointAccess string, cidrBlocks []string, xff bool, useExisting *bool, networkID, newNetworkCIDR string, ports []portCreateRequest) types.Object {
	portObjType := types.ObjectType{AttrTypes: portAttrTypes()}

	attrTypes := map[string]attr.Type{
		"external_access":        types.BoolType,
		"endpoint_access":        types.StringType,
		"cidr_block":             types.ListType{ElemType: types.StringType},
		"x_forwarded_for":        types.BoolType,
		"use_existing_network":   types.BoolType,
		"network_id":             types.StringType,
		"create_new_network_cidr": types.StringType,
		"port":                   types.ListType{ElemType: portObjType},
	}

	epVal := types.StringNull()
	if endpointAccess != "" {
		epVal = types.StringValue(endpointAccess)
	}

	cidrVal := types.ListNull(types.StringType)
	if cidrBlocks != nil {
		cidrVal, _ = types.ListValueFrom(context.Background(), types.StringType, cidrBlocks)
	}

	useExistVal := types.BoolNull()
	if useExisting != nil {
		useExistVal = types.BoolValue(*useExisting)
	}

	netIDVal := types.StringNull()
	if networkID != "" {
		netIDVal = types.StringValue(networkID)
	}

	newCIDRVal := types.StringNull()
	if newNetworkCIDR != "" {
		newCIDRVal = types.StringValue(newNetworkCIDR)
	}

	portListVal := types.ListNull(portObjType)
	if ports != nil {
		portObjs := []attr.Value{}
		for _, p := range ports {
			expVal := types.Int64Null()
			if p.ExposedPort != nil {
				expVal = types.Int64Value(*p.ExposedPort)
			}
			pObj, _ := types.ObjectValue(portAttrTypes(), map[string]attr.Value{
				"name":           types.StringValue(p.Name),
				"protocol":       types.StringValue(p.Protocol),
				"container_port": types.Int64Value(p.ContainerPort),
				"exposed_port":   expVal,
			})
			portObjs = append(portObjs, pObj)
		}
		portListVal, _ = types.ListValue(portObjType, portObjs)
	}

	obj, _ := types.ObjectValue(attrTypes, map[string]attr.Value{
		"external_access":        types.BoolValue(extAccess),
		"endpoint_access":        epVal,
		"cidr_block":             cidrVal,
		"x_forwarded_for":        types.BoolValue(xff),
		"use_existing_network":   useExistVal,
		"network_id":             netIDVal,
		"create_new_network_cidr": newCIDRVal,
		"port":                   portListVal,
	})
	return obj
}

func buildAutoscalingObject(enabled bool, minR, maxR *int64, cpuT, memT *float64) types.Object {
	minVal := types.Int64Null()
	if minR != nil {
		minVal = types.Int64Value(*minR)
	}
	maxVal := types.Int64Null()
	if maxR != nil {
		maxVal = types.Int64Value(*maxR)
	}
	cpuVal := types.Float64Null()
	if cpuT != nil {
		cpuVal = types.Float64Value(*cpuT)
	}
	memVal := types.Float64Null()
	if memT != nil {
		memVal = types.Float64Value(*memT)
	}

	obj, _ := types.ObjectValue(autoscalingAttrTypes(), map[string]attr.Value{
		"enabled":                  types.BoolValue(enabled),
		"min_replicas":             minVal,
		"max_replicas":             maxVal,
		"cpu_target_percentage":    cpuVal,
		"memory_target_percentage": memVal,
	})
	return obj
}
