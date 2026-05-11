package router

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/os/neutron/routers"

var (
	_ resource.Resource              = &routerResource{}
	_ resource.ResourceWithConfigure = &routerResource{}
)

type routerResource struct {
	client *client.Client
}

// --- API types (matching ace-cli/cmd/vpc/router.go) ---

type routerGatewayInfo struct {
	NetworkID string `json:"network_id"`
}

type routerAPIResponse struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	Description         string             `json:"description"`
	AdminStateUp        bool               `json:"admin_state_up"`
	ExternalGatewayInfo *routerGatewayInfo `json:"external_gateway_info"`
	Status              string             `json:"status"`
	CreatedAt           string             `json:"created_at"`
	UpdatedAt           string             `json:"updated_at"`
}

type routerDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &routerResource{}
}

func (r *routerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_router"
}

func (r *routerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = routerSchema()
}

func (r *routerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// buildCreateRequest converts the Terraform plan to a create API request body.
// Router create API does NOT accept "description" — only name, admin_state_up, external_gateway_info.
// Uses map[string]interface{} to exactly match ace-cli serialization behavior.
func buildCreateRequest(plan *routerResourceModel) map[string]interface{} {
	routerBody := map[string]interface{}{
		"name":           plan.Name.ValueString(),
		"admin_state_up": true,
	}

	if !plan.AdminStateUp.IsNull() && !plan.AdminStateUp.IsUnknown() {
		routerBody["admin_state_up"] = plan.AdminStateUp.ValueBool()
	}

	// External gateway info — matches ace-cli: always send the key,
	// empty object {} when no gateway, object with network_id when set.
	gwInfo := map[string]interface{}{}
	if !plan.ExternalGatewayNetworkID.IsNull() && !plan.ExternalGatewayNetworkID.IsUnknown() {
		gwInfo["network_id"] = plan.ExternalGatewayNetworkID.ValueString()
	}
	routerBody["external_gateway_info"] = gwInfo

	return map[string]interface{}{
		"router": routerBody,
	}
}

// buildUpdateRequest converts the Terraform plan to an update API request body.
// NOTE: npc-api router PUT rejects "description" — only name, admin_state_up,
// and external_gateway_info are accepted on update.
func buildUpdateRequest(plan *routerResourceModel) map[string]interface{} {
	routerBody := map[string]interface{}{
		"name":           plan.Name.ValueString(),
		"admin_state_up": true,
	}

	if !plan.AdminStateUp.IsNull() && !plan.AdminStateUp.IsUnknown() {
		routerBody["admin_state_up"] = plan.AdminStateUp.ValueBool()
	}

	// Only include external_gateway_info when gateway is set.
	// Sending empty {} causes 400 on some backends.
	if !plan.ExternalGatewayNetworkID.IsNull() && !plan.ExternalGatewayNetworkID.IsUnknown() {
		routerBody["external_gateway_info"] = map[string]interface{}{
			"network_id": plan.ExternalGatewayNetworkID.ValueString(),
		}
	}

	return map[string]interface{}{
		"router": routerBody,
	}
}

// mapAPIResponseToState converts an API response to the Terraform state model.
func mapAPIResponseToState(model *routerResourceModel, apiResp *routerAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	model.AdminStateUp = types.BoolValue(apiResp.AdminStateUp)

	if apiResp.ExternalGatewayInfo != nil && apiResp.ExternalGatewayInfo.NetworkID != "" {
		model.ExternalGatewayNetworkID = types.StringValue(apiResp.ExternalGatewayInfo.NetworkID)
	} else if model.ExternalGatewayNetworkID.IsNull() {
		model.ExternalGatewayNetworkID = types.StringNull()
	}

	model.Status = types.StringValue(apiResp.Status)

	if apiResp.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	} else {
		model.CreatedAt = types.StringNull()
	}
	if apiResp.UpdatedAt != "" {
		model.UpdatedAt = types.StringValue(apiResp.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringNull()
	}
}

// parseRouterResponse handles both wrapped {"router": {...}} and direct {...} response formats.
func parseRouterResponse(data json.RawMessage) (*routerAPIResponse, error) {
	// Try wrapped format first: {"router": {...}}
	var wrapped struct {
		Router routerAPIResponse `json:"router"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Router.ID != "" {
		return &wrapped.Router, nil
	}

	// Fallback to direct format
	var direct routerAPIResponse
	if err := json.Unmarshal(data, &direct); err != nil {
		return nil, fmt.Errorf("failed to parse router response: %w", err)
	}
	return &direct, nil
}

func (r *routerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan routerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(&plan)

	apiResp, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create router", err.Error())
		return
	}

	created, err := parseRouterResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	mapAPIResponseToState(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state routerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read router", err.Error())
		return
	}

	router, err := parseRouterResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	mapAPIResponseToState(&state, router)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *routerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan routerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state routerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(&plan)

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update router", err.Error())
		return
	}

	updated, err := parseRouterResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}

	plannedGateway := plan.ExternalGatewayNetworkID
	mapAPIResponseToState(&plan, updated)

	// Preserve plan's gateway value — if user removed gateway (null),
	// keep null even if API still returns the old value (async removal).
	if plannedGateway.IsNull() && !plan.ExternalGatewayNetworkID.IsNull() {
		plan.ExternalGatewayNetworkID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *routerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state routerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := routerDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(_ context.Context) error {
			_, err := r.client.Delete(ctx, apiPath, body)
			return err
		},
		RetryableErrors: []string{"in use", "already in use"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete router", err.Error())
		return
	}
}
