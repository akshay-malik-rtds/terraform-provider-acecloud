package k8s_cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the resource interfaces.
var (
	_ resource.Resource              = &k8sClusterResource{}
	_ resource.ResourceWithConfigure = &k8sClusterResource{}
)

// k8sClusterResource is the resource implementation.
type k8sClusterResource struct {
	client *client.Client
}

// NewResource is the factory function registered in the provider.
func NewResource() resource.Resource {
	return &k8sClusterResource{}
}

// Metadata returns the resource type name.
func (r *k8sClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s_cluster"
}

// Configure stores the provider-configured client for later use.
func (r *k8sClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// ---------------------------------------------------------------------------
// API request / response types
// ---------------------------------------------------------------------------

// createK8sClusterRequest is the JSON body sent to POST /k8s/cluster-overview/create.
// Field names match the npc-api CreateClusterDto exactly.
type createK8sClusterRequest struct {
	Name                string           `json:"name"`
	KubernetesVersion   string           `json:"kubernetesVersion"`
	EndpointAccess      string           `json:"endpointAccess"`
	NetworkIsolation    string           `json:"networkIsolation"`
	NginxIngress        string           `json:"nginxIngress"`
	NginxDefaultBackend string           `json:"nginxDefaultBackend"`
	NetworkProvider     string           `json:"networkProvider"`
	SnapshotBackup      string           `json:"snapshotBackup"`
	SnapshotEnabled     string           `json:"snapshotEnabled"`
	SnapshotInterval    int              `json:"snapshotInterval"`
	SnapshotRetention   int              `json:"snapshotRetention"`
	SecretsEncryption   string           `json:"secretsEncryption"`
	MaxWorkerNodes      int64            `json:"maxWorkerNodes"`
	DrainNodes          string           `json:"drainNodes"`
	DrainTimeout        string           `json:"drainTimeout"`
	TerminationTime     int              `json:"terminationTime"`
	DrainTime           int              `json:"drainTime"`
	DeleteDirData       string           `json:"deleteDirData"`
	Force               string           `json:"force"`
	GracePeriod         string           `json:"gracePeriod"`
	Sources             []clusterSource  `json:"sources"`
	WorkerNodeName      string           `json:"workerNodeName"`
	Quantity            int64            `json:"quantity"`
	FlavorID            string           `json:"flavorId"`
	Volume              int64            `json:"volume"`
	SecGroupID          string           `json:"secGroupId,omitempty"`
	ClusterType         string           `json:"cluster_type"`
	Autoscale           bool             `json:"autoscale"`
	AutoscaleMin        int64            `json:"autoscaleMin"`
	AutoscaleMax        int64            `json:"autoscaleMax"`
	CPU                 bool             `json:"cpu"`
	GPU                 bool             `json:"gpu"`
}

type clusterSource struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// listK8sClusterItem represents a single cluster in the list response from
// GET /k8s/cluster-overview/clusters.
type listK8sClusterItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

// readK8sClusterResponse represents relevant fields returned by the
// GET /k8s/cluster-overview/:id endpoint.
type readK8sClusterResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Created   string `json:"created"`
	CreatedAt string `json:"created_at"`
}

// parseClusterData unmarshals the raw API response into readK8sClusterResponse.
func parseClusterData(data json.RawMessage) (*readK8sClusterResponse, error) {
	var resp readK8sClusterResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal cluster data: %w", err)
	}
	return &resp, nil
}

// ---------------------------------------------------------------------------
// CRUD operations
// ---------------------------------------------------------------------------

// Create provisions a new Kubernetes cluster via POST /k8s/cluster-overview/create.
func (r *k8sClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan k8sClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(&plan)

	_, err := r.client.Post(ctx, "/k8s/cluster-overview/create", body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Kubernetes cluster", err.Error())
		return
	}

	// K8s create returns no cluster ID — poll the list endpoint to find it by name.
	clusterName := plan.Name.ValueString()
	clusterID := ""

	pollResult, pollErr := wait.PollForResource(ctx, wait.PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			listResp, err := r.client.Get(ctx, "/k8s/cluster-overview/clusters", nil)
			if err != nil {
				return nil, err
			}
			// While provisioning, the API may return {"clusterStatus":"..."} instead of an array.
			// Try parsing as array first; if that fails, return nil to keep polling.
			var clusters []listK8sClusterItem
			if err := json.Unmarshal(listResp.Data, &clusters); err != nil {
				// Not an array yet — cluster still being set up.
				return nil, nil
			}
			for _, c := range clusters {
				if c.Name == clusterName {
					return &c, nil
				}
			}
			return nil, nil
		},
		Timeout:      10 * time.Minute,
		PollInterval: 30 * time.Second,
	})

	if pollErr != nil {
		resp.Diagnostics.AddError(
			"Failed to find Kubernetes cluster after creation",
			fmt.Sprintf("Cluster '%s' was created but could not be found in the cluster list: %s", clusterName, pollErr),
		)
		return
	}

	if item, ok := pollResult.(*listK8sClusterItem); ok {
		clusterID = item.ID
	}

	plan.ID = types.StringValue(clusterID)
	plan.Status = types.StringValue("provisioning")

	// Wait for the cluster to become active (up to 20 minutes).
	result, waitErr := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			readResp, err := r.client.Get(ctx, fmt.Sprintf("/k8s/cluster-overview/%s", clusterID), nil)
			if err != nil {
				return nil, err
			}
			cluster, err := parseClusterData(readResp.Data)
			if err != nil {
				return nil, err
			}
			return &wait.StatusResult{Status: cluster.State, Data: cluster}, nil
		},
		TargetStatus: []string{"active", "Active"},
		ErrorStatus:  []string{"error", "Error", "failed"},
		Timeout:      20 * time.Minute,
		PollInterval: 15 * time.Second,
	})

	if waitErr != nil {
		resp.Diagnostics.AddWarning(
			"Kubernetes cluster provisioning incomplete",
			fmt.Sprintf("Cluster %s created but did not reach active status: %s", clusterID, waitErr),
		)
	}

	if result != nil && result.Data != nil {
		cluster := result.Data.(*readK8sClusterResponse)
		plan.Status = types.StringValue(cluster.State)
		if cluster.Created != "" {
			plan.CreatedAt = types.StringValue(cluster.Created)
		} else {
			plan.CreatedAt = types.StringValue(cluster.CreatedAt)
		}
	} else {
		// Set computed fields to known values even if polling failed.
		plan.Status = types.StringValue("provisioning")
		plan.CreatedAt = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state from the backend via GET /k8s/cluster-overview/:id.
func (r *k8sClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state k8sClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	apiResp, err := r.client.Get(ctx, fmt.Sprintf("/k8s/cluster-overview/%s", id), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to read Kubernetes cluster",
			fmt.Sprintf("Could not read cluster %s: %s", id, err),
		)
		return
	}

	cluster, err := parseClusterData(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse read response",
			fmt.Sprintf("Could not parse cluster data: %s", err),
		)
		return
	}

	// Map computed fields from the API response.
	// All configurable fields are ForceNew, so they are preserved from state.
	state.ID = types.StringValue(cluster.ID)
	if cluster.Name != "" {
		state.Name = types.StringValue(cluster.Name)
	}
	state.Status = types.StringValue(cluster.State)
	if cluster.Created != "" {
		state.CreatedAt = types.StringValue(cluster.Created)
	} else if cluster.CreatedAt != "" {
		state.CreatedAt = types.StringValue(cluster.CreatedAt)
	} else {
		state.CreatedAt = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported — all configurable fields use RequiresReplace.
// This method exists to satisfy the resource.Resource interface.
func (r *k8sClusterResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Failed to update Kubernetes cluster",
		"Update is not supported for Kubernetes clusters. All configurable fields require resource replacement.",
	)
}

// Delete removes the cluster via DELETE /k8s/cluster-overview/:id.
func (r *k8sClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state k8sClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	_, err := r.client.Delete(ctx, fmt.Sprintf("/k8s/cluster-overview/%s", id), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to delete Kubernetes cluster",
			fmt.Sprintf("Could not delete cluster %s: %s", id, err),
		)
		return
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildCreateRequest converts the Terraform plan model into the API request
// body expected by POST /k8s/cluster-overview/create.
func buildCreateRequest(plan *k8sClusterResourceModel) *createK8sClusterRequest {
	body := &createK8sClusterRequest{
		Name:                plan.Name.ValueString(),
		KubernetesVersion:   plan.KubernetesVersion.ValueString(),
		EndpointAccess:      plan.EndpointAccess.ValueString(),
		NetworkIsolation:    plan.NetworkIsolation.ValueString(),
		NginxIngress:        plan.NginxIngress.ValueString(),
		NginxDefaultBackend: plan.NginxDefaultBackend.ValueString(),
		NetworkProvider:     plan.NetworkProvider.ValueString(),
		SecretsEncryption:   plan.SecretsEncryption.ValueString(),
		MaxWorkerNodes:      plan.MaxWorkerNodes.ValueInt64(),
		WorkerNodeName:      plan.WorkerNodeName.ValueString(),
		Quantity:            plan.WorkerQuantity.ValueInt64(),
		FlavorID:            plan.FlavorID.ValueString(),
		Volume:              plan.VolumeSize.ValueInt64(),
		ClusterType:         plan.ClusterType.ValueString(),
		// Hardcoded defaults matching npc-api CreateClusterDto.
		SnapshotInterval:  12,
		SnapshotRetention: 6,
		SnapshotEnabled:   "No",
		DrainNodes:        "No",
		DrainTimeout:      "Give up after:",
		TerminationTime:   30,
		DrainTime:         120,
		DeleteDirData:     "No",
		Force:             "No",
		GracePeriod:       "-1",
		Sources:           []clusterSource{{ID: "default-cidr", Value: "0.0.0.0/0"}},
		CPU:               true,
		GPU:               false,
	}

	// Optional: sec_group_id.
	if !plan.SecGroupID.IsNull() && !plan.SecGroupID.IsUnknown() {
		body.SecGroupID = plan.SecGroupID.ValueString()
	}

	// Optional: snapshot_backup.
	if !plan.SnapshotBackup.IsNull() && !plan.SnapshotBackup.IsUnknown() {
		body.SnapshotBackup = plan.SnapshotBackup.ValueString()
		if plan.SnapshotBackup.ValueString() == "Yes" {
			body.SnapshotEnabled = "Yes"
		}
	}

	// Optional: autoscale.
	if !plan.Autoscale.IsNull() && !plan.Autoscale.IsUnknown() {
		body.Autoscale = plan.Autoscale.ValueBool()
	}

	// Optional: autoscale_min.
	if !plan.AutoscaleMin.IsNull() && !plan.AutoscaleMin.IsUnknown() {
		body.AutoscaleMin = plan.AutoscaleMin.ValueInt64()
	}

	// Optional: autoscale_max.
	if !plan.AutoscaleMax.IsNull() && !plan.AutoscaleMax.IsUnknown() {
		body.AutoscaleMax = plan.AutoscaleMax.ValueInt64()
	}

	return body
}

// mapReadResponseToState maps the API read response into the Terraform state model.
// Only computed fields are updated; configurable fields are ForceNew and preserved from state.
func mapReadResponseToState(cluster *readK8sClusterResponse, state *k8sClusterResourceModel) {
	state.ID = types.StringValue(cluster.ID)
	if cluster.Name != "" {
		state.Name = types.StringValue(cluster.Name)
	}
	state.Status = types.StringValue(cluster.State)
	if cluster.Created != "" {
		state.CreatedAt = types.StringValue(cluster.Created)
	} else if cluster.CreatedAt != "" {
		state.CreatedAt = types.StringValue(cluster.CreatedAt)
	} else {
		state.CreatedAt = types.StringValue("")
	}
}
