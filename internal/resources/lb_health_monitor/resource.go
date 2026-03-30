package lb_health_monitor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/loadbalancers/pools/health-monitors"

var (
	_ resource.Resource              = &lbHealthMonitorResource{}
	_ resource.ResourceWithConfigure = &lbHealthMonitorResource{}
)

type lbHealthMonitorResource struct {
	client *client.Client
}

// --- API types (matching npc-ui useLoadBalancerHealthMonitors.ts) ---

type hmCreateRequest struct {
	Name           string `json:"name"`
	PoolID         string `json:"pool_id"`
	Type           string `json:"type"`
	Delay          int64  `json:"delay"`
	Timeout        int64  `json:"timeout"`
	MaxRetries     int64  `json:"max_retries"`
	MaxRetriesDown int64  `json:"max_retries_down,omitempty"`
	URLPath        string `json:"url_path,omitempty"`
	ExpectedCodes  string `json:"expected_codes,omitempty"`
	HTTPMethod     string `json:"http_method,omitempty"`
}

type hmUpdateRequest struct {
	Name           string `json:"name"`
	Delay          int64  `json:"delay"`
	Timeout        int64  `json:"timeout"`
	MaxRetries     int64  `json:"max_retries"`
	MaxRetriesDown int64  `json:"max_retries_down,omitempty"`
	URLPath        string `json:"url_path,omitempty"`
	ExpectedCodes  string `json:"expected_codes,omitempty"`
	HTTPMethod     string `json:"http_method,omitempty"`
}

type hmAPIResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	PoolID             string `json:"pool_id"`
	Type               string `json:"type"`
	Delay              int64  `json:"delay"`
	Timeout            int64  `json:"timeout"`
	MaxRetries         int64  `json:"max_retries"`
	MaxRetriesDown     int64  `json:"max_retries_down"`
	URLPath            string `json:"url_path"`
	ExpectedCodes      string `json:"expected_codes"`
	HTTPMethod         string `json:"http_method"`
	ProvisioningStatus string `json:"provisioning_status"`
	OperatingStatus    string `json:"operating_status"`
	AdminStateUp       bool   `json:"admin_state_up"`
	CreatedAt          string `json:"created_at"`
}

type hmDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &lbHealthMonitorResource{}
}

func (r *lbHealthMonitorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_health_monitor"
}

func (r *lbHealthMonitorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = lbHealthMonitorSchema()
}

func (r *lbHealthMonitorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(plan *lbHealthMonitorResourceModel) hmCreateRequest {
	body := hmCreateRequest{
		Name:       plan.Name.ValueString(),
		PoolID:     plan.PoolID.ValueString(),
		Type:       plan.Type.ValueString(),
		Delay:      plan.Delay.ValueInt64(),
		Timeout:    plan.Timeout.ValueInt64(),
		MaxRetries: plan.MaxRetries.ValueInt64(),
	}
	if !plan.MaxRetriesDown.IsNull() && !plan.MaxRetriesDown.IsUnknown() {
		body.MaxRetriesDown = plan.MaxRetriesDown.ValueInt64()
	}
	if !plan.URLPath.IsNull() && !plan.URLPath.IsUnknown() {
		body.URLPath = plan.URLPath.ValueString()
	}
	if !plan.ExpectedCodes.IsNull() && !plan.ExpectedCodes.IsUnknown() {
		body.ExpectedCodes = plan.ExpectedCodes.ValueString()
	}
	if !plan.HTTPMethod.IsNull() && !plan.HTTPMethod.IsUnknown() {
		body.HTTPMethod = plan.HTTPMethod.ValueString()
	}
	return body
}

func buildUpdateRequest(plan *lbHealthMonitorResourceModel) hmUpdateRequest {
	body := hmUpdateRequest{
		Name:       plan.Name.ValueString(),
		Delay:      plan.Delay.ValueInt64(),
		Timeout:    plan.Timeout.ValueInt64(),
		MaxRetries: plan.MaxRetries.ValueInt64(),
	}
	if !plan.MaxRetriesDown.IsNull() && !plan.MaxRetriesDown.IsUnknown() {
		body.MaxRetriesDown = plan.MaxRetriesDown.ValueInt64()
	}
	if !plan.URLPath.IsNull() && !plan.URLPath.IsUnknown() {
		body.URLPath = plan.URLPath.ValueString()
	}
	if !plan.ExpectedCodes.IsNull() && !plan.ExpectedCodes.IsUnknown() {
		body.ExpectedCodes = plan.ExpectedCodes.ValueString()
	}
	if !plan.HTTPMethod.IsNull() && !plan.HTTPMethod.IsUnknown() {
		body.HTTPMethod = plan.HTTPMethod.ValueString()
	}
	return body
}

func mapAPIResponseToState(model *lbHealthMonitorResourceModel, apiResp *hmAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.PoolID != "" {
		model.PoolID = types.StringValue(apiResp.PoolID)
	}
	if apiResp.Type != "" {
		model.Type = types.StringValue(apiResp.Type)
	}

	model.Delay = types.Int64Value(apiResp.Delay)
	model.Timeout = types.Int64Value(apiResp.Timeout)
	model.MaxRetries = types.Int64Value(apiResp.MaxRetries)

	if apiResp.MaxRetriesDown > 0 {
		model.MaxRetriesDown = types.Int64Value(apiResp.MaxRetriesDown)
	} else if model.MaxRetriesDown.IsNull() {
		model.MaxRetriesDown = types.Int64Null()
	}

	if apiResp.URLPath != "" {
		model.URLPath = types.StringValue(apiResp.URLPath)
	} else if model.URLPath.IsNull() || model.URLPath.IsUnknown() {
		model.URLPath = types.StringNull()
	}

	if apiResp.ExpectedCodes != "" {
		model.ExpectedCodes = types.StringValue(apiResp.ExpectedCodes)
	} else if model.ExpectedCodes.IsNull() || model.ExpectedCodes.IsUnknown() {
		model.ExpectedCodes = types.StringNull()
	}

	if apiResp.HTTPMethod != "" {
		model.HTTPMethod = types.StringValue(apiResp.HTTPMethod)
	} else if model.HTTPMethod.IsNull() || model.HTTPMethod.IsUnknown() {
		model.HTTPMethod = types.StringNull()
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
	model.AdminStateUp = types.BoolValue(apiResp.AdminStateUp)
	if apiResp.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
}

func (r *lbHealthMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lbHealthMonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Wait for pool to be ACTIVE before creating health monitor
	poolID := plan.PoolID.ValueString()
	poolPath := fmt.Sprintf("/cloud/loadbalancers/pools/%s", poolID)
	_, err := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			poolResp, err := r.client.Get(ctx, poolPath, nil)
			if err != nil {
				return nil, err
			}
			var raw map[string]interface{}
			if err := json.Unmarshal(poolResp.Data, &raw); err != nil {
				return nil, err
			}
			status, _ := raw["provisioning_status"].(string)
			return &wait.StatusResult{Status: status}, nil
		},
		TargetStatus: []string{"ACTIVE"},
		ErrorStatus:  []string{"ERROR"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create health monitor", err.Error())
		return
	}

	body := buildCreateRequest(&plan)

	// Retry on "immutable" errors (pool may be temporarily in PENDING_UPDATE)
	var apiResp *client.APIResponse
	err = wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			var postErr error
			apiResp, postErr = r.client.Post(ctx, apiPath, body)
			return postErr
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create health monitor", err.Error())
		return
	}

	var created hmAPIResponse
	if err := json.Unmarshal(apiResp.Data, &created); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	mapAPIResponseToState(&plan, &created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lbHealthMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state lbHealthMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read health monitor", err.Error())
		return
	}

	var hm hmAPIResponse
	if err := json.Unmarshal(apiResp.Data, &hm); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	mapAPIResponseToState(&state, &hm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *lbHealthMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan lbHealthMonitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state lbHealthMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(&plan)

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())

	// Retry on "immutable" / "Duplicate" errors (LB may be in PENDING_UPDATE)
	var apiResp *client.APIResponse
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			var opErr error
			apiResp, opErr = r.client.Put(ctx, path, body)
			return opErr
		},
		RetryableErrors: []string{"immutable", "PENDING_", "Duplicate", "in use"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update health monitor", err.Error())
		return
	}

	var updated hmAPIResponse
	if err := json.Unmarshal(apiResp.Data, &updated); err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}

	mapAPIResponseToState(&plan, &updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lbHealthMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state lbHealthMonitorResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := hmDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	// Retry on "immutable" errors (pool/LB may be in PENDING_UPDATE from other operations)
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, apiPath, body)
			return err
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete health monitor", err.Error())
	}
}
