package caas_secret

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const basePath = "/caas/secrets"

var (
	_ resource.Resource              = &caasSecretResource{}
	_ resource.ResourceWithConfigure = &caasSecretResource{}
)

type caasSecretResource struct {
	client *client.Client
}

// --- API types ---

type secretCreateRequest struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	URL      string            `json:"url,omitempty"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

type secretUpdateRequest struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	URL      string            `json:"url,omitempty"`
	Username string            `json:"username,omitempty"`
	Password string            `json:"password,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
}

type secretAPIResponse struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	URL       string            `json:"url"`
	Username  string            `json:"username"`
	Password  string            `json:"password"`
	Data      map[string]string `json:"data"`
	Status    string            `json:"status"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

type secretDeleteRequest struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

func NewResource() resource.Resource {
	return &caasSecretResource{}
}

func (r *caasSecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caas_secret"
}

func (r *caasSecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = caasSecretSchema()
}

func (r *caasSecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			"Expected *client.Client, got an unexpected type.",
		)
		return
	}
	r.client = c
}

func buildCreateRequest(ctx context.Context, plan *caasSecretModel) secretCreateRequest {
	body := secretCreateRequest{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}

	if !plan.URL.IsNull() && !plan.URL.IsUnknown() {
		body.URL = plan.URL.ValueString()
	}
	if !plan.Username.IsNull() && !plan.Username.IsUnknown() {
		body.Username = plan.Username.ValueString()
	}
	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		body.Password = plan.Password.ValueString()
	}

	if !plan.Data.IsNull() && !plan.Data.IsUnknown() {
		data := make(map[string]string)
		plan.Data.ElementsAs(ctx, &data, false)
		body.Data = data
	}

	return body
}

func buildUpdateRequest(ctx context.Context, plan *caasSecretModel) secretUpdateRequest {
	body := secretUpdateRequest{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}

	if !plan.URL.IsNull() && !plan.URL.IsUnknown() {
		body.URL = plan.URL.ValueString()
	}
	if !plan.Username.IsNull() && !plan.Username.IsUnknown() {
		body.Username = plan.Username.ValueString()
	}
	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		body.Password = plan.Password.ValueString()
	}

	if !plan.Data.IsNull() && !plan.Data.IsUnknown() {
		data := make(map[string]string)
		plan.Data.ElementsAs(ctx, &data, false)
		body.Data = data
	}

	return body
}

func mapAPIResponseToState(ctx context.Context, model *caasSecretModel, apiResp *secretAPIResponse) {
	// CaaS secrets are identified by name
	model.ID = types.StringValue(apiResp.Name)
	model.Name = types.StringValue(apiResp.Name)
	model.Type = types.StringValue(apiResp.Type)

	// Registry fields
	if apiResp.URL != "" {
		model.URL = types.StringValue(apiResp.URL)
	} else if model.URL.IsNull() {
		model.URL = types.StringNull()
	}

	if apiResp.Username != "" {
		model.Username = types.StringValue(apiResp.Username)
	} else if model.Username.IsNull() {
		model.Username = types.StringNull()
	}

	// Note: password is typically not returned by API for security.
	// Preserve the plan/state value. Only update if API returns it.
	if apiResp.Password != "" {
		model.Password = types.StringValue(apiResp.Password)
	}
	// If password was set in plan but API doesn't return it, keep the plan value
	// (it's already in the model from plan/state)

	// Generic data
	if len(apiResp.Data) > 0 {
		dataMap, _ := types.MapValueFrom(ctx, types.StringType, apiResp.Data)
		model.Data = dataMap
	} else if model.Data.IsNull() {
		model.Data = types.MapNull(types.StringType)
	}

	// Computed fields
	if apiResp.Status != "" {
		model.Status = types.StringValue(apiResp.Status)
	} else {
		model.Status = types.StringValue("")
	}

	if apiResp.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.CreatedAt)
	} else {
		model.CreatedAt = types.StringValue("")
	}

	if apiResp.UpdatedAt != "" {
		model.UpdatedAt = types.StringValue(apiResp.UpdatedAt)
	} else {
		model.UpdatedAt = types.StringValue("")
	}
}

func (r *caasSecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan caasSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	_, err := r.client.Post(ctx, basePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create CaaS secret", err.Error())
		return
	}

	// Set ID = name (CaaS secrets are identified by name)
	plan.ID = types.StringValue(plan.Name.ValueString())

	// Wait for secret to become Active
	readPath := fmt.Sprintf("%s/%s", basePath, plan.Name.ValueString())
	result, waitErr := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			readResp, err := r.client.Get(ctx, readPath, nil)
			if err != nil {
				return nil, err
			}
			var secret secretAPIResponse
			if err := json.Unmarshal(readResp.Data, &secret); err != nil {
				return nil, err
			}
			return &wait.StatusResult{Status: secret.Status, Data: &secret}, nil
		},
		TargetStatus: []string{"Active"},
		ErrorStatus:  []string{"OutOfSync"},
	})

	if waitErr != nil {
		plan.Status = types.StringValue("Provisioning")
		plan.CreatedAt = types.StringValue("")
		plan.UpdatedAt = types.StringValue("")
		resp.Diagnostics.AddWarning("CaaS secret may still be provisioning", waitErr.Error())
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if secret, ok := result.Data.(*secretAPIResponse); ok {
		mapAPIResponseToState(ctx, &plan, secret)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *caasSecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state caasSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, state.Name.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read CaaS secret", err.Error())
		return
	}

	var secret secretAPIResponse
	if err := json.Unmarshal(apiResp.Data, &secret); err != nil {
		resp.Diagnostics.AddError("Failed to parse CaaS secret response", err.Error())
		return
	}

	mapAPIResponseToState(ctx, &state, &secret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *caasSecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan caasSecretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state caasSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(ctx, &plan)

	path := fmt.Sprintf("%s/%s", basePath, state.Name.ValueString())
	_, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update CaaS secret", err.Error())
		return
	}

	// Read back to get updated state
	readResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		plan.ID = state.ID
		plan.Status = types.StringValue("")
		plan.CreatedAt = types.StringValue("")
		plan.UpdatedAt = types.StringValue("")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	var secret secretAPIResponse
	if err := json.Unmarshal(readResp.Data, &secret); err != nil {
		plan.ID = state.ID
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	mapAPIResponseToState(ctx, &plan, &secret)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *caasSecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state caasSecretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// CaaS delete uses key/values format (shared DeleteRequestDTO)
	body := secretDeleteRequest{
		Key:    "name",
		Values: []string{state.Name.ValueString()},
	}

	_, err := r.client.Delete(ctx, basePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete CaaS secret", err.Error())
	}
}
