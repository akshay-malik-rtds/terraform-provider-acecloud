package lb_pool_member

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiBasePath = "/cloud/loadbalancers/pools"

var (
	_ resource.Resource              = &lbPoolMemberResource{}
	_ resource.ResourceWithConfigure = &lbPoolMemberResource{}
)

type lbPoolMemberResource struct {
	client *client.Client
}

// --- API types (matching npc-ui useLoadBalancerBackendServers.ts) ---

type memberCreateWrapper struct {
	BackendServers []memberCreateBody `json:"backend_servers"`
}

type memberCreateBody struct {
	Name           string `json:"name,omitempty"`
	Address        string `json:"address"`
	ProtocolPort   int64  `json:"protocol_port"`
	Weight         *int64 `json:"weight,omitempty"`
	MonitorPort    int64  `json:"monitor_port,omitempty"`
	MonitorAddress string `json:"monitor_address,omitempty"`
}

type memberUpdateBody struct {
	Weight         *int64 `json:"weight,omitempty"`
	MonitorPort    int64  `json:"monitor_port,omitempty"`
	MonitorAddress string `json:"monitor_address,omitempty"`
}

type memberAPIResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Address            string `json:"address"`
	ProtocolPort       int64  `json:"protocol_port"`
	Weight             int64  `json:"weight"`
	MonitorPort        int64  `json:"monitor_port"`
	MonitorAddress     string `json:"monitor_address"`
	ProvisioningStatus string `json:"provisioning_status"`
	OperatingStatus    string `json:"operating_status"`
	AdminStateUp       bool   `json:"admin_state_up"`
	CreatedAt          string `json:"created_at"`
}

type memberDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &lbPoolMemberResource{}
}

func (r *lbPoolMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_pool_member"
}

func (r *lbPoolMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = lbPoolMemberSchema()
}

func (r *lbPoolMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(plan *lbPoolMemberResourceModel) memberCreateWrapper {
	member := memberCreateBody{
		Address:      plan.Address.ValueString(),
		ProtocolPort: plan.ProtocolPort.ValueInt64(),
	}
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		member.Name = plan.Name.ValueString()
	}
	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		w := plan.Weight.ValueInt64()
		member.Weight = &w
	}
	if !plan.MonitorPort.IsNull() && !plan.MonitorPort.IsUnknown() {
		member.MonitorPort = plan.MonitorPort.ValueInt64()
	}
	if !plan.MonitorAddress.IsNull() && !plan.MonitorAddress.IsUnknown() {
		member.MonitorAddress = plan.MonitorAddress.ValueString()
	}
	return memberCreateWrapper{
		BackendServers: []memberCreateBody{member},
	}
}

func buildUpdateRequest(plan *lbPoolMemberResourceModel) memberUpdateBody {
	body := memberUpdateBody{}
	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		w := plan.Weight.ValueInt64()
		body.Weight = &w
	}
	if !plan.MonitorPort.IsNull() && !plan.MonitorPort.IsUnknown() {
		body.MonitorPort = plan.MonitorPort.ValueInt64()
	}
	if !plan.MonitorAddress.IsNull() && !plan.MonitorAddress.IsUnknown() {
		body.MonitorAddress = plan.MonitorAddress.ValueString()
	}
	return body
}

func mapAPIResponseToState(model *lbPoolMemberResourceModel, apiResp *memberAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)

	if apiResp.Name != "" {
		model.Name = types.StringValue(apiResp.Name)
	} else if model.Name.IsNull() {
		model.Name = types.StringNull()
	}

	if apiResp.Address != "" {
		model.Address = types.StringValue(apiResp.Address)
	}
	if apiResp.ProtocolPort > 0 {
		model.ProtocolPort = types.Int64Value(apiResp.ProtocolPort)
	}

	model.Weight = types.Int64Value(apiResp.Weight)
	model.AdminStateUp = types.BoolValue(apiResp.AdminStateUp)

	if apiResp.MonitorPort > 0 {
		model.MonitorPort = types.Int64Value(apiResp.MonitorPort)
	} else if model.MonitorPort.IsNull() {
		model.MonitorPort = types.Int64Null()
	}

	if apiResp.MonitorAddress != "" {
		model.MonitorAddress = types.StringValue(apiResp.MonitorAddress)
	} else if model.MonitorAddress.IsNull() {
		model.MonitorAddress = types.StringNull()
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
	if apiResp.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
}

func (r *lbPoolMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lbPoolMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolID := plan.PoolID.ValueString()

	// Wait for pool to be ACTIVE before adding members
	poolPath := fmt.Sprintf("%s/%s", apiBasePath, poolID)
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
		resp.Diagnostics.AddError("Timeout waiting for pool to become active", err.Error())
		return
	}

	body := buildCreateRequest(&plan)

	// POST /cloud/loadbalancers/pools/{pool_id}/backend-servers
	// Retry on "immutable" errors (pool may be temporarily in PENDING_UPDATE)
	createPath := fmt.Sprintf("%s/%s/backend-servers", apiBasePath, poolID)
	err = wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Post(ctx, createPath, body)
			return err
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create pool member", err.Error())
		return
	}

	// Pool member creation is async (API returns data:{}). Poll until it appears.
	targetAddress := plan.Address.ValueString()
	targetPort := plan.ProtocolPort.ValueInt64()
	listPath := fmt.Sprintf("%s/%s/backend-servers", apiBasePath, poolID)

	item, err := wait.PollForResource(ctx, wait.PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			listResp, err := r.client.Get(ctx, listPath, nil)
			if err != nil {
				return nil, err
			}
			var members []memberAPIResponse
			if err := json.Unmarshal(listResp.Data, &members); err != nil {
				return nil, err
			}
			for _, m := range members {
				if m.Address == targetAddress && m.ProtocolPort == targetPort {
					return &m, nil
				}
			}
			return nil, nil // not found yet
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Pool member not found after creation",
			fmt.Sprintf("Member %s:%d was not found in pool %s after polling: %s", targetAddress, targetPort, poolID, err))
		return
	}
	found := item.(*memberAPIResponse)

	mapAPIResponseToState(&plan, found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lbPoolMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state lbPoolMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// GET /cloud/loadbalancers/pools/{pool_id}/backend-servers/{id}
	readPath := fmt.Sprintf("%s/%s/backend-servers/%s", apiBasePath, state.PoolID.ValueString(), state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, readPath, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read pool member", err.Error())
		return
	}

	var member memberAPIResponse
	if err := json.Unmarshal(apiResp.Data, &member); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	mapAPIResponseToState(&state, &member)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *lbPoolMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan lbPoolMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state lbPoolMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(&plan)

	// PUT /cloud/loadbalancers/pools/{pool_id}/backend-servers/{id}
	updatePath := fmt.Sprintf("%s/%s/backend-servers/%s", apiBasePath, state.PoolID.ValueString(), state.ID.ValueString())

	// Retry on "immutable" errors (LB may be in PENDING_UPDATE from other operations)
	var apiResp *client.APIResponse
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			var opErr error
			apiResp, opErr = r.client.Put(ctx, updatePath, body)
			return opErr
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update pool member", err.Error())
		return
	}

	var updated memberAPIResponse
	if err := json.Unmarshal(apiResp.Data, &updated); err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}

	mapAPIResponseToState(&plan, &updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *lbPoolMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state lbPoolMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// DELETE /cloud/loadbalancers/pools/{pool_id}/backend-servers
	deletePath := fmt.Sprintf("%s/%s/backend-servers", apiBasePath, state.PoolID.ValueString())
	body := memberDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	// Retry on "immutable" errors (pool/LB may be in PENDING_UPDATE from other operations)
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, deletePath, body)
			return err
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete pool member", err.Error())
	}
}
