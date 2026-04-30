package lb_pool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/loadbalancers/pools"

var (
	_ resource.Resource              = &lbPoolResource{}
	_ resource.ResourceWithConfigure = &lbPoolResource{}
)

type lbPoolResource struct {
	client *client.Client
}

// --- API types (matching ace-cli/internal/api/loadbalancer.go) ---

type poolCreateRequest struct {
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	LBAlgorithm    string `json:"lb_algorithm"`
	ListenerID     string `json:"listener_id,omitempty"`
	LoadBalancerID string `json:"loadbalancer_id,omitempty"`
}

type poolAPIResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Protocol           string `json:"protocol"`
	LBAlgorithm        string `json:"lb_algorithm"`
	ListenerID         string `json:"listener_id"`
	LoadBalancerID     string `json:"loadbalancer_id"`
	ProvisioningStatus string `json:"provisioning_status"`
	OperatingStatus    string `json:"operating_status"`
	HealthMonitorID    string `json:"healthmonitor_id"`
	CreatedAt          string `json:"created_at"`
}

type poolDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &lbPoolResource{}
}

func (r *lbPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_pool"
}

func (r *lbPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = lbPoolSchema()
}

func (r *lbPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(plan *lbPoolResourceModel) poolCreateRequest {
	body := poolCreateRequest{
		Name:        plan.Name.ValueString(),
		Protocol:    plan.Protocol.ValueString(),
		LBAlgorithm: plan.LBAlgorithm.ValueString(),
	}
	if !plan.ListenerID.IsNull() && !plan.ListenerID.IsUnknown() {
		body.ListenerID = plan.ListenerID.ValueString()
	}
	if !plan.LoadBalancerID.IsNull() && !plan.LoadBalancerID.IsUnknown() {
		body.LoadBalancerID = plan.LoadBalancerID.ValueString()
	}
	return body
}

func mapAPIResponseToState(model *lbPoolResourceModel, apiResp *poolAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.Protocol != "" {
		model.Protocol = types.StringValue(apiResp.Protocol)
	}
	if apiResp.LBAlgorithm != "" {
		model.LBAlgorithm = types.StringValue(apiResp.LBAlgorithm)
	}

	if apiResp.ListenerID != "" {
		model.ListenerID = types.StringValue(apiResp.ListenerID)
	} else if model.ListenerID.IsNull() {
		model.ListenerID = types.StringNull()
	}

	if apiResp.LoadBalancerID != "" {
		model.LoadBalancerID = types.StringValue(apiResp.LoadBalancerID)
	} else if model.LoadBalancerID.IsNull() {
		model.LoadBalancerID = types.StringNull()
	}

	if apiResp.ProvisioningStatus != "" {
		model.ProvisioningStatus = types.StringValue(apiResp.ProvisioningStatus)
	} else {
		model.ProvisioningStatus = types.StringNull()
	}
	if apiResp.OperatingStatus != "" {
		model.OperatingStatus = types.StringValue(apiResp.OperatingStatus)
	} else {
		model.OperatingStatus = types.StringNull()
	}
	if apiResp.HealthMonitorID != "" {
		model.HealthMonitorID = types.StringValue(apiResp.HealthMonitorID)
	} else {
		model.HealthMonitorID = types.StringNull()
	}
	if apiResp.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
}

func parsePoolFromRaw(data json.RawMessage) (*poolAPIResponse, error) {
	var pool poolAPIResponse
	if err := json.Unmarshal(data, &pool); err != nil {
		return nil, err
	}
	// API may return listeners/loadbalancers as [{id:...}] arrays instead of simple ID strings
	if pool.ListenerID == "" || pool.LoadBalancerID == "" {
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err == nil {
			if pool.ListenerID == "" {
				if listeners, ok := raw["listeners"].([]interface{}); ok && len(listeners) > 0 {
					if lObj, ok := listeners[0].(map[string]interface{}); ok {
						pool.ListenerID, _ = lObj["id"].(string)
					}
				}
			}
			if pool.LoadBalancerID == "" {
				if lbs, ok := raw["loadbalancers"].([]interface{}); ok && len(lbs) > 0 {
					if lbObj, ok := lbs[0].(map[string]interface{}); ok {
						pool.LoadBalancerID, _ = lbObj["id"].(string)
					}
				}
			}
		}
	}
	return &pool, nil
}

func (r *lbPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lbPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(&plan)

	_, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create LB pool", err.Error())
		return
	}

	// Pool creation is async (API returns data:{}). Poll until it appears.
	targetName := plan.Name.ValueString()

	item, err := wait.PollForResource(ctx, wait.PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			listResp, err := r.client.Get(ctx, apiPath, nil)
			if err != nil {
				return nil, err
			}
			var rawPools []json.RawMessage
			if err := json.Unmarshal(listResp.Data, &rawPools); err != nil {
				return nil, err
			}
			for _, rawPool := range rawPools {
				pool, err := parsePoolFromRaw(rawPool)
				if err != nil {
					continue
				}
				if pool.Name == targetName {
					return pool, nil
				}
			}
			return nil, nil // not found yet
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create LB pool",
			fmt.Sprintf("Pool %q was not found after polling: %s", targetName, err))
		return
	}
	found := item.(*poolAPIResponse)

	// Wait for pool to reach ACTIVE status (needed for dependent resources like members)
	if found.ProvisioningStatus != "ACTIVE" && found.ID != "" {
		poolPath := fmt.Sprintf("%s/%s", apiPath, found.ID)
		result, _ := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
			Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
				readResp, err := r.client.Get(ctx, poolPath, nil)
				if err != nil {
					return nil, err
				}
				pool, err := parsePoolFromRaw(readResp.Data)
				if err != nil {
					return nil, err
				}
				return &wait.StatusResult{Status: pool.ProvisioningStatus, Data: pool}, nil
			},
			TargetStatus: []string{"ACTIVE"},
			ErrorStatus:  []string{"ERROR"},
		})
		if result != nil && result.Data != nil {
			found = result.Data.(*poolAPIResponse)
		}
	}

	mapAPIResponseToState(&plan, found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lbPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state lbPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read LB pool", err.Error())
		return
	}

	pool, err := parsePoolFromRaw(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	// Preserve listener_id and loadbalancer_id from state if API doesn't return them
	if pool.ListenerID == "" {
		pool.ListenerID = state.ListenerID.ValueString()
	}
	if pool.LoadBalancerID == "" {
		pool.LoadBalancerID = state.LoadBalancerID.ValueString()
	}

	mapAPIResponseToState(&state, pool)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *lbPoolResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Pools do not support update. Changes trigger destroy and recreate.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"LB pool resources do not support in-place updates. Changes will trigger a destroy and recreate.",
	)
}

func (r *lbPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state lbPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := poolDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	// Retry on "immutable" errors (LB may be in PENDING_UPDATE from other sub-resource operations)
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, apiPath, body)
			return err
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete LB pool", err.Error())
	}
}
