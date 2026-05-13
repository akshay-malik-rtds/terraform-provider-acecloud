package lb_listener

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/loadbalancers/listeners"

var (
	_ resource.Resource              = &lbListenerResource{}
	_ resource.ResourceWithConfigure = &lbListenerResource{}
)

type lbListenerResource struct {
	client *client.Client
}

// --- API types (matching ace-cli/internal/api/loadbalancer.go) ---

type listenerCreateRequest struct {
	Name                 string            `json:"name"`
	Protocol             string            `json:"protocol"`
	ProtocolPort         int64             `json:"protocol_port"`
	LoadBalancerID       string            `json:"loadbalancer_id"`
	Description          string            `json:"description,omitempty"`
	ConnectionLimit      *int64            `json:"connection_limit,omitempty"`
	AllowedCIDRs         []string          `json:"allowed_cidrs,omitempty"`
	InsertHeaders        map[string]string `json:"insert_headers,omitempty"`
	TimeoutClientData    *int64            `json:"timeout_client_data,omitempty"`
	TimeoutMemberConnect *int64            `json:"timeout_member_connect,omitempty"`
	TimeoutMemberData    *int64            `json:"timeout_member_data,omitempty"`
	TLSCiphers           string            `json:"tls_ciphers,omitempty"`
}

type listenerAPIResponse struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Protocol             string            `json:"protocol"`
	ProtocolPort         int64             `json:"protocol_port"`
	LoadBalancerID       string            `json:"loadbalancer_id"`
	Description          string            `json:"description"`
	ConnectionLimit      int64             `json:"connection_limit"`
	AllowedCIDRs         []string          `json:"allowed_cidrs"`
	InsertHeaders        map[string]string `json:"insert_headers"`
	TimeoutClientData    int64             `json:"timeout_client_data"`
	TimeoutMemberConnect int64             `json:"timeout_member_connect"`
	TimeoutMemberData    int64             `json:"timeout_member_data"`
	TLSCiphers           string            `json:"tls_ciphers"`
	ProvisioningStatus   string            `json:"provisioning_status"`
	OperatingStatus      string            `json:"operating_status"`
	DefaultPoolID        string            `json:"default_pool_id"`
	CreatedAt            string            `json:"created_at"`
}

type listenerDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &lbListenerResource{}
}

func (r *lbListenerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_listener"
}

func (r *lbListenerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = lbListenerSchema()
}

func (r *lbListenerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(ctx context.Context, plan *lbListenerResourceModel) listenerCreateRequest {
	body := listenerCreateRequest{
		Name:           plan.Name.ValueString(),
		Protocol:       plan.Protocol.ValueString(),
		ProtocolPort:   plan.ProtocolPort.ValueInt64(),
		LoadBalancerID: plan.LoadBalancerID.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.ConnectionLimit.IsNull() && !plan.ConnectionLimit.IsUnknown() {
		v := plan.ConnectionLimit.ValueInt64()
		body.ConnectionLimit = &v
	}
	if !plan.AllowedCIDRs.IsNull() && !plan.AllowedCIDRs.IsUnknown() {
		var cidrs []string
		plan.AllowedCIDRs.ElementsAs(ctx, &cidrs, false)
		body.AllowedCIDRs = cidrs
	}
	if !plan.InsertHeaders.IsNull() && !plan.InsertHeaders.IsUnknown() {
		var headers map[string]string
		plan.InsertHeaders.ElementsAs(ctx, &headers, false)
		body.InsertHeaders = headers
	}
	if !plan.TimeoutClientData.IsNull() && !plan.TimeoutClientData.IsUnknown() {
		v := plan.TimeoutClientData.ValueInt64()
		body.TimeoutClientData = &v
	}
	if !plan.TimeoutMemberConnect.IsNull() && !plan.TimeoutMemberConnect.IsUnknown() {
		v := plan.TimeoutMemberConnect.ValueInt64()
		body.TimeoutMemberConnect = &v
	}
	if !plan.TimeoutMemberData.IsNull() && !plan.TimeoutMemberData.IsUnknown() {
		v := plan.TimeoutMemberData.ValueInt64()
		body.TimeoutMemberData = &v
	}
	if !plan.TLSCiphers.IsNull() && !plan.TLSCiphers.IsUnknown() {
		body.TLSCiphers = plan.TLSCiphers.ValueString()
	}
	return body
}

func mapAPIResponseToState(ctx context.Context, model *lbListenerResourceModel, apiResp *listenerAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.Protocol != "" {
		model.Protocol = types.StringValue(apiResp.Protocol)
	}
	if apiResp.ProtocolPort > 0 {
		model.ProtocolPort = types.Int64Value(apiResp.ProtocolPort)
	}
	if apiResp.LoadBalancerID != "" {
		model.LoadBalancerID = types.StringValue(apiResp.LoadBalancerID)
	}

	// description: preserve null when API returns "" and user never set it.
	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	model.ConnectionLimit = types.Int64Value(apiResp.ConnectionLimit)

	// allowed_cidrs: only populate when user originally set the field, to avoid
	// drift when API returns nil/empty for an unset list.
	if !model.AllowedCIDRs.IsNull() && !model.AllowedCIDRs.IsUnknown() {
		cidrs, _ := types.ListValueFrom(ctx, types.StringType, apiResp.AllowedCIDRs)
		model.AllowedCIDRs = cidrs
	}

	// insert_headers: only populate when user originally set the field.
	if !model.InsertHeaders.IsNull() && !model.InsertHeaders.IsUnknown() {
		headers, _ := types.MapValueFrom(ctx, types.StringType, apiResp.InsertHeaders)
		model.InsertHeaders = headers
	}

	model.TimeoutClientData = types.Int64Value(apiResp.TimeoutClientData)
	model.TimeoutMemberConnect = types.Int64Value(apiResp.TimeoutMemberConnect)
	model.TimeoutMemberData = types.Int64Value(apiResp.TimeoutMemberData)

	if apiResp.TLSCiphers != "" {
		model.TLSCiphers = types.StringValue(apiResp.TLSCiphers)
	} else if model.TLSCiphers.IsNull() {
		model.TLSCiphers = types.StringNull()
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
	if apiResp.DefaultPoolID != "" {
		model.DefaultPoolID = types.StringValue(apiResp.DefaultPoolID)
	} else {
		model.DefaultPoolID = types.StringNull()
	}
	if apiResp.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
}

func (r *lbListenerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan lbListenerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	_, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create LB listener", err.Error())
		return
	}

	// Listener creation is async (API returns data:{}). Poll until it appears.
	targetName := plan.Name.ValueString()
	targetLBID := plan.LoadBalancerID.ValueString()

	item, err := wait.PollForResource(ctx, wait.PollForResourceOpts{
		List: func(ctx context.Context) (interface{}, error) {
			listResp, err := r.client.Get(ctx, apiPath, nil)
			if err != nil {
				return nil, err
			}
			var rawListeners []map[string]interface{}
			if err := json.Unmarshal(listResp.Data, &rawListeners); err != nil {
				return nil, err
			}
			for _, raw := range rawListeners {
				name, _ := raw["name"].(string)
				if name != targetName {
					continue
				}
				lbs, _ := raw["loadbalancers"].([]interface{})
				for _, lb := range lbs {
					lbObj, _ := lb.(map[string]interface{})
					if lbID, _ := lbObj["id"].(string); lbID == targetLBID {
						rawJSON, _ := json.Marshal(raw)
						var listener listenerAPIResponse
						if err := json.Unmarshal(rawJSON, &listener); err == nil {
							listener.LoadBalancerID = targetLBID
							return &listener, nil
						}
					}
				}
			}
			return nil, nil // not found yet
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create LB listener",
			fmt.Sprintf("Listener %q for LB %s was not found after polling: %s", targetName, targetLBID, err))
		return
	}
	found := item.(*listenerAPIResponse)

	// Wait for listener to reach ACTIVE status (needed for dependent pool creation)
	if found.ProvisioningStatus != "ACTIVE" && found.ID != "" {
		listenerPath := fmt.Sprintf("%s/%s", apiPath, found.ID)
		result, _ := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
			Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
				readResp, err := r.client.Get(ctx, listenerPath, nil)
				if err != nil {
					return nil, err
				}
				listener, err := parseListenerFromRaw(readResp.Data)
				if err != nil {
					return nil, err
				}
				listener.LoadBalancerID = targetLBID
				return &wait.StatusResult{Status: listener.ProvisioningStatus, Data: listener}, nil
			},
			TargetStatus: []string{"ACTIVE"},
			ErrorStatus:  []string{"ERROR"},
		})
		if result != nil && result.Data != nil {
			found = result.Data.(*listenerAPIResponse)
		}
	}

	mapAPIResponseToState(ctx, &plan, found)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func parseListenerFromRaw(data json.RawMessage) (*listenerAPIResponse, error) {
	var listener listenerAPIResponse
	if err := json.Unmarshal(data, &listener); err != nil {
		return nil, err
	}
	// API returns loadbalancers:[{id:...}] not loadbalancer_id
	if listener.LoadBalancerID == "" {
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err == nil {
			if lbs, ok := raw["loadbalancers"].([]interface{}); ok && len(lbs) > 0 {
				if lbObj, ok := lbs[0].(map[string]interface{}); ok {
					listener.LoadBalancerID, _ = lbObj["id"].(string)
				}
			}
		}
	}
	return &listener, nil
}

func (r *lbListenerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state lbListenerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read LB listener", err.Error())
		return
	}

	listener, err := parseListenerFromRaw(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	// Preserve loadbalancer_id from state if API doesn't return it
	if listener.LoadBalancerID == "" {
		listener.LoadBalancerID = state.LoadBalancerID.ValueString()
	}

	mapAPIResponseToState(ctx, &state, listener)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *lbListenerResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Listeners do not support update. Changes trigger destroy and recreate.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"LB listener resources do not support in-place updates. Changes will trigger a destroy and recreate.",
	)
}

func (r *lbListenerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state lbListenerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := listenerDeleteRequest{
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
		resp.Diagnostics.AddError("Failed to delete LB listener", err.Error())
	}
}
