package load_balancer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/loadbalancers"

var (
	_ resource.Resource              = &loadBalancerResource{}
	_ resource.ResourceWithConfigure = &loadBalancerResource{}
)

type loadBalancerResource struct {
	client *client.Client
}

// --- API types (matching ace-cli/internal/api/loadbalancer.go) ---

type lbCreateRequest struct {
	Name        string   `json:"name"`
	SubnetID    string   `json:"subnet_id"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type lbUpdateRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// lbAPIResponse handles both npc-api create and read response formats.
// Create returns: address, port_id, subnet_id, network_id
// Read returns: vip_address, vip_port_id, vip_subnet_id, vip_network_id
type lbAPIResponse struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	Tags               []string `json:"tags"`
	ProvisioningStatus string   `json:"provisioning_status"`
	OperatingStatus    string   `json:"operating_status"`
	Provider           string   `json:"provider"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	// These fields vary between create and read responses
	SubnetID     string
	VIPAddress   string
	VIPPortID    string
	VIPNetworkID string
}

func parseLBData(data json.RawMessage) (*lbAPIResponse, error) {
	var resp lbAPIResponse
	// Parse known fields
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	// Handle field name variants
	resp.SubnetID = firstStr(raw, "subnet_id", "vip_subnet_id")
	resp.VIPAddress = firstStr(raw, "address", "vip_address")
	resp.VIPPortID = firstStr(raw, "port_id", "vip_port_id")
	resp.VIPNetworkID = firstStr(raw, "network_id", "vip_network_id")
	return &resp, nil
}

func firstStr(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

type lbDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &loadBalancerResource{}
}

func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

func (r *loadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = loadBalancerSchema()
}

func (r *loadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(ctx context.Context, plan *loadBalancerResourceModel) lbCreateRequest {
	body := lbCreateRequest{
		Name:     plan.Name.ValueString(),
		SubnetID: plan.SubnetID.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		plan.Tags.ElementsAs(ctx, &tags, false)
		body.Tags = tags
	}
	return body
}

func buildUpdateRequest(ctx context.Context, plan *loadBalancerResourceModel) lbUpdateRequest {
	body := lbUpdateRequest{
		Name: plan.Name.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		plan.Tags.ElementsAs(ctx, &tags, false)
		body.Tags = tags
	}
	return body
}

func mapAPIResponseToState(model *loadBalancerResourceModel, apiResp *lbAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.SubnetID != "" {
		model.SubnetID = types.StringValue(apiResp.SubnetID)
	}

	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() || model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}

	if len(apiResp.Tags) > 0 {
		// Check if existing tags have the same elements (possibly reordered)
		// If so, preserve the existing order to avoid spurious diffs
		if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
			var existingTags []string
			model.Tags.ElementsAs(context.Background(), &existingTags, false)
			if tagsEqual(existingTags, apiResp.Tags) {
				// Same elements, keep existing order
			} else {
				tagVals := make([]attr.Value, len(apiResp.Tags))
				for i, t := range apiResp.Tags {
					tagVals[i] = types.StringValue(t)
				}
				model.Tags = types.ListValueMust(types.StringType, tagVals)
			}
		} else {
			tagVals := make([]attr.Value, len(apiResp.Tags))
			for i, t := range apiResp.Tags {
				tagVals[i] = types.StringValue(t)
			}
			model.Tags = types.ListValueMust(types.StringType, tagVals)
		}
	} else if model.Tags.IsNull() || model.Tags.IsUnknown() {
		model.Tags = types.ListNull(types.StringType)
	}

	// All computed fields must be resolved (not Unknown) after Create.
	// Set to value if available, otherwise null.
	model.VIPAddress = stringValueOrNull(apiResp.VIPAddress)
	model.VIPPortID = stringValueOrNull(apiResp.VIPPortID)
	model.VIPNetworkID = stringValueOrNull(apiResp.VIPNetworkID)
	model.ProvisioningStatus = stringValueOrNull(apiResp.ProvisioningStatus)
	model.OperatingStatus = stringValueOrNull(apiResp.OperatingStatus)
	model.Provider = stringValueOrNull(apiResp.Provider)
	model.CreatedAt = stringValueOrNull(apiResp.CreatedAt)
	model.UpdatedAt = stringValueOrNull(apiResp.UpdatedAt)
}

func stringValueOrNull(s string) types.String {
	if s != "" {
		return types.StringValue(s)
	}
	return types.StringNull()
}

// tagsEqual returns true if two string slices contain the same elements
// regardless of order.
func tagsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, s := range a {
		counts[s]++
	}
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	return true
}

func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan loadBalancerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	apiResp, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create load balancer", err.Error())
		return
	}

	if len(apiResp.Data) == 0 {
		resp.Diagnostics.AddError(
			"Failed to create load balancer",
			"The API returned an empty response when creating the load balancer",
		)
		return
	}

	created, err := parseLBData(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	// Poll until LB reaches ACTIVE status (needed for dependent resources like listeners)
	if created.ProvisioningStatus != "ACTIVE" && created.ID != "" {
		lbPath := fmt.Sprintf("%s/%s", apiPath, created.ID)
		result, _ := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
			Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
				readResp, err := r.client.Get(ctx, lbPath, nil)
				if err != nil {
					return nil, err
				}
				lb, err := parseLBData(readResp.Data)
				if err != nil {
					return nil, err
				}
				return &wait.StatusResult{Status: lb.ProvisioningStatus, Data: lb}, nil
			},
			TargetStatus: []string{"ACTIVE"},
			ErrorStatus:  []string{"ERROR"},
		})
		if result != nil && result.Data != nil {
			created = result.Data.(*lbAPIResponse)
		}
	}

	mapAPIResponseToState(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state loadBalancerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read load balancer", err.Error())
		return
	}

	lb, err := parseLBData(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	mapAPIResponseToState(&state, lb)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan loadBalancerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state loadBalancerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(ctx, &plan)

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update load balancer", err.Error())
		return
	}

	updated, err := parseLBData(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse update response", err.Error())
		return
	}

	// Save planned tags — API may return them in a different order
	plannedTags := plan.Tags

	mapAPIResponseToState(&plan, updated)

	// Always use plan's tag order since we sent them and the API accepted them.
	// The API may return tags in a different order which causes spurious diffs.
	if !plannedTags.IsNull() && !plannedTags.IsUnknown() {
		plan.Tags = plannedTags
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state loadBalancerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Plain delete — Terraform's dependency graph deletes child resources
	// (listeners, pools, etc.) before the load balancer itself when they are
	// managed by separate Terraform resources, so a cascade hint is not
	// needed. Avoids backend builds that reject the cascade query parameter.
	body := lbDeleteRequest{
		Key:    "id",
		Values: []string{state.ID.ValueString()},
	}

	// Retry on transient errors (LB may still be processing sub-resource
	// deletions when destroy runs, or report associated resources briefly).
	err := wait.RetryOnConflict(ctx, wait.RetryOnConflictOpts{
		Operation: func(ctx context.Context) error {
			_, err := r.client.Delete(ctx, apiPath, body)
			return err
		},
		RetryableErrors: []string{"associated resources", "Pending"},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete load balancer", err.Error())
	}
}
