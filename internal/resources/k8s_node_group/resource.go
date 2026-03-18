package k8s_node_group

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	createPath    = "/k8s/compute/create-node-group"
	listPath      = "/k8s/compute/node-groups"
	deletePath    = "/k8s/compute/delete"
	scalePath     = "/k8s/compute/scale-node-group"
)

var (
	_ resource.Resource              = &k8sNodeGroupResource{}
	_ resource.ResourceWithConfigure = &k8sNodeGroupResource{}
)

type k8sNodeGroupResource struct {
	client *client.Client
}

// --- API types ---

type nodeGroupCreateRequest struct {
	ClusterID   string            `json:"clusterId"`
	SecGroupID  string            `json:"secGroupId"`
	Name        string            `json:"name"`
	Quantity    int64             `json:"quantity"`
	Volume      string            `json:"volume"`
	FlavorID    string            `json:"flavorId"`
	CPU         bool              `json:"cpu"`
	GPU         bool              `json:"gpu"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	MinNode     *int64            `json:"minNode,omitempty"`
	MaxNode     *int64            `json:"maxNode,omitempty"`
}

type nodeGroupScaleRequest struct {
	Updates   scaleUpdates `json:"updates"`
	NodeName  string       `json:"nodeName"`
	ClusterID string       `json:"clusterId"`
}

type scaleUpdates struct {
	Type     string `json:"type"`
	Quantity int64  `json:"quantity"`
}

type nodeGroupAPIResponse struct {
	ID          string            `json:"id"`
	ClusterID   string            `json:"clusterId"`
	SecGroupID  string            `json:"secGroupId"`
	Name        string            `json:"name"`
	Quantity    int64             `json:"quantity"`
	FlavorID    string            `json:"flavorId"`
	Volume      string            `json:"volume"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	MinNode     int64             `json:"minNode"`
	MaxNode     int64             `json:"maxNode"`
	State       string            `json:"state"`
}

func NewResource() resource.Resource {
	return &k8sNodeGroupResource{}
}

func (r *k8sNodeGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s_node_group"
}

func (r *k8sNodeGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = k8sNodeGroupSchema()
}

func (r *k8sNodeGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func buildCreateRequest(ctx context.Context, plan *k8sNodeGroupModel) nodeGroupCreateRequest {
	body := nodeGroupCreateRequest{
		ClusterID:  plan.ClusterID.ValueString(),
		SecGroupID: plan.SecGroupID.ValueString(),
		Name:       plan.Name.ValueString(),
		Quantity:   plan.Quantity.ValueInt64(),
		Volume:     plan.Volume.ValueString(),
		FlavorID:   plan.FlavorID.ValueString(),
		CPU:        true,
		GPU:        false,
		Labels:     make(map[string]string),
		Annotations: make(map[string]string),
	}

	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
		labels := make(map[string]string)
		plan.Labels.ElementsAs(ctx, &labels, false)
		body.Labels = labels
	}
	if !plan.Annotations.IsNull() && !plan.Annotations.IsUnknown() {
		annotations := make(map[string]string)
		plan.Annotations.ElementsAs(ctx, &annotations, false)
		body.Annotations = annotations
	}

	if !plan.MinNode.IsNull() && !plan.MinNode.IsUnknown() {
		v := plan.MinNode.ValueInt64()
		body.MinNode = &v
	}
	if !plan.MaxNode.IsNull() && !plan.MaxNode.IsUnknown() {
		v := plan.MaxNode.ValueInt64()
		body.MaxNode = &v
	}

	return body
}

func buildScaleRequest(plan *k8sNodeGroupModel, state *k8sNodeGroupModel) nodeGroupScaleRequest {
	newQty := plan.Quantity.ValueInt64()
	oldQty := state.Quantity.ValueInt64()

	scaleType := "add"
	delta := newQty - oldQty
	if newQty < oldQty {
		scaleType = "remove"
		delta = oldQty - newQty
	}

	return nodeGroupScaleRequest{
		Updates: scaleUpdates{
			Type:     scaleType,
			Quantity: delta,
		},
		NodeName:  state.Name.ValueString(),
		ClusterID: state.ClusterID.ValueString(),
	}
}

func mapAPIResponseToState(ctx context.Context, model *k8sNodeGroupModel, apiResp *nodeGroupAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.ClusterID = types.StringValue(apiResp.ClusterID)
	model.SecGroupID = types.StringValue(apiResp.SecGroupID)
	model.Name = types.StringValue(apiResp.Name)
	model.Quantity = types.Int64Value(apiResp.Quantity)
	model.FlavorID = types.StringValue(apiResp.FlavorID)
	model.Volume = types.StringValue(apiResp.Volume)

	if len(apiResp.Labels) > 0 {
		labelsMap, diags := types.MapValueFrom(ctx, types.StringType, apiResp.Labels)
		if !diags.HasError() {
			model.Labels = labelsMap
		}
	} else if model.Labels.IsNull() {
		model.Labels = types.MapNull(types.StringType)
	}

	if len(apiResp.Annotations) > 0 {
		annotationsMap, diags := types.MapValueFrom(ctx, types.StringType, apiResp.Annotations)
		if !diags.HasError() {
			model.Annotations = annotationsMap
		}
	} else if model.Annotations.IsNull() {
		model.Annotations = types.MapNull(types.StringType)
	}

	if apiResp.MinNode > 0 {
		model.MinNode = types.Int64Value(apiResp.MinNode)
	} else if model.MinNode.IsNull() {
		model.MinNode = types.Int64Null()
	}

	if apiResp.MaxNode > 0 {
		model.MaxNode = types.Int64Value(apiResp.MaxNode)
	} else if model.MaxNode.IsNull() {
		model.MaxNode = types.Int64Null()
	}

	if apiResp.State != "" {
		model.State = types.StringValue(apiResp.State)
	} else {
		model.State = types.StringValue("")
	}
}

// parseNodeGroupFromRaw parses a single node group from raw JSON.
func parseNodeGroupFromRaw(data json.RawMessage) (*nodeGroupAPIResponse, error) {
	var ng nodeGroupAPIResponse
	if err := json.Unmarshal(data, &ng); err != nil {
		return nil, err
	}
	return &ng, nil
}

// parseNodeGroupListFromData extracts node group list from API response data.
// The response may be {data: [...]} or a bare array.
func parseNodeGroupListFromData(data json.RawMessage) ([]json.RawMessage, error) {
	// Try {data: [...]} envelope first
	var envelope struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && len(envelope.Data) > 0 {
		return envelope.Data, nil
	}

	// Try bare array
	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("unable to parse node group list: %w", err)
	}
	return items, nil
}

func (r *k8sNodeGroupResource) findNodeGroupByID(ctx context.Context, clusterID, nodeGroupID string) (*nodeGroupAPIResponse, error) {
	params := map[string]string{
		"clusterId": clusterID,
		"limit":     "100",
	}
	apiResp, err := r.client.Get(ctx, listPath, params)
	if err != nil {
		return nil, err
	}

	items, err := parseNodeGroupListFromData(apiResp.Data)
	if err != nil {
		return nil, err
	}

	for _, raw := range items {
		ng, err := parseNodeGroupFromRaw(raw)
		if err != nil {
			continue
		}
		if ng.ID == nodeGroupID {
			return ng, nil
		}
	}
	return nil, fmt.Errorf("node group %s not found in cluster %s", nodeGroupID, clusterID)
}

func (r *k8sNodeGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan k8sNodeGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	_, err := r.client.Post(ctx, createPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Kubernetes node group", err.Error())
		return
	}

	// Node group creation is async. Poll until it appears in the list.
	targetName := plan.Name.ValueString()
	clusterID := plan.ClusterID.ValueString()

	item, err := wait.PollForResource(ctx, wait.PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			params := map[string]string{
				"clusterId": clusterID,
				"limit":     "100",
			}
			listResp, err := r.client.Get(ctx, listPath, params)
			if err != nil {
				return nil, err
			}
			items, err := parseNodeGroupListFromData(listResp.Data)
			if err != nil {
				return nil, err
			}
			for _, raw := range items {
				ng, err := parseNodeGroupFromRaw(raw)
				if err != nil {
					continue
				}
				if ng.Name == targetName {
					return ng, nil
				}
			}
			return nil, nil // not found yet
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Kubernetes node group",
			fmt.Sprintf("Node group %q was not found after polling: %s", targetName, err))
		return
	}
	found := item.(*nodeGroupAPIResponse)

	mapAPIResponseToState(ctx, &plan, found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *k8sNodeGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state k8sNodeGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ng, err := r.findNodeGroupByID(ctx, state.ClusterID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Kubernetes node group", err.Error())
		return
	}

	mapAPIResponseToState(ctx, &state, ng)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *k8sNodeGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan k8sNodeGroupModel
	var state k8sNodeGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only quantity changes are supported (all other fields are ForceNew).
	if !plan.Quantity.Equal(state.Quantity) {
		scaleBody := buildScaleRequest(&plan, &state)
		scalePutPath := fmt.Sprintf("%s/%s", scalePath, state.ID.ValueString())

		_, err := r.client.Put(ctx, scalePutPath, scaleBody)
		if err != nil {
			resp.Diagnostics.AddError("Failed to scale Kubernetes node group", err.Error())
			return
		}

		// Poll until state stabilizes after scale.
		clusterID := state.ClusterID.ValueString()
		nodeGroupID := state.ID.ValueString()

		result, err := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
			Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
				ng, err := r.findNodeGroupByID(ctx, clusterID, nodeGroupID)
				if err != nil {
					return nil, err
				}
				return &wait.StatusResult{Status: ng.State, Data: ng}, nil
			},
			TargetStatus: []string{"ACTIVE", "RUNNING"},
			ErrorStatus:  []string{"ERROR", "FAILED"},
		})
		if err != nil {
			resp.Diagnostics.AddError("Failed to scale Kubernetes node group",
				fmt.Sprintf("Timed out waiting for node group to stabilize: %s", err))
			return
		}
		if result != nil && result.Data != nil {
			found := result.Data.(*nodeGroupAPIResponse)
			mapAPIResponseToState(ctx, &plan, found)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
	}

	// No changes needed beyond quantity — just refresh state.
	ng, err := r.findNodeGroupByID(ctx, state.ClusterID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Kubernetes node group", err.Error())
		return
	}

	mapAPIResponseToState(ctx, &plan, ng)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *k8sNodeGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state k8sNodeGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deletePutPath := fmt.Sprintf("%s/%s", deletePath, state.ID.ValueString())

	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, deletePutPath, nil)
			return err
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Kubernetes node group", err.Error())
	}
}
