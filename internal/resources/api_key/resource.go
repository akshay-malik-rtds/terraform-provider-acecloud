package api_key

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
)

const apiKeysBasePath = "/iam/api-keys"

// Ensure the resource implements the expected Plugin Framework interfaces.
var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithConfigure   = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

// NewResource returns a new acecloud_api_key resource factory.
func NewResource() resource.Resource {
	return &apiKeyResource{}
}

type apiKeyResource struct {
	client *client.Client
}

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema()
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}
	r.client = c
}

// createRequest mirrors the npc-api CreateIacUserDto.
type createRequest struct {
	ServiceName string `json:"serviceName"`
	Description string `json:"description,omitempty"`
}

// updateRequest mirrors the npc-api UpdateIacUserDto.
type updateRequest struct {
	ServiceName string `json:"serviceName,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// keyResponse is the shape returned by the npc-api on create / read / update.
// On create, the full credential `<id>.<secret>` is returned in `apiKey`; the
// secret cannot be recovered later.
type keyResponse struct {
	ID          string `json:"_id"`
	APIKey      string `json:"apiKey,omitempty"` // "<id>.<secret>" — only on create + revive
	ServiceName string `json:"serviceName"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := createRequest{
		ServiceName: plan.ServiceName.ValueString(),
		Description: plan.Description.ValueString(),
	}

	apiResp, err := r.client.Post(ctx, apiKeysBasePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create API key", err.Error())
		return
	}

	// Create response shape: {"data": {"apiKey": "<id>.<secret>"}}
	parsed, err := parseKeyResponse(apiResp.Data)
	if err != nil || parsed.APIKey == "" {
		resp.Diagnostics.AddError(
			"API key create response missing credential",
			"The API key secret was not returned. The key may have been created but cannot be stored in Terraform state. "+
				"Check the AceCloud console — if the key exists, regenerate the secret from the console and import it.",
		)
		return
	}

	// apiKey is "<id>.<secret>"; split into separate state fields.
	keyID, secret, ok := strings.Cut(parsed.APIKey, ".")
	if !ok {
		resp.Diagnostics.AddError(
			"Malformed API key returned by backend",
			fmt.Sprintf("Expected format '<id>.<secret>', got: %q", parsed.APIKey),
		)
		return
	}

	// Create returns only the credential. Fetch metadata
	// (serviceName, expiresAt, createdAt) via a follow-up GET.
	state := apiKeyResourceModel{
		ID:          types.StringValue(keyID),
		Secret:      types.StringValue(secret),
		ServiceName: plan.ServiceName,
		Description: plan.Description,
		Enabled:     types.BoolValue(true),
		ExpiresAt:   types.StringValue(""),
		CreatedAt:   types.StringValue(""),
		UpdatedAt:   types.StringValue(""),
	}

	if getResp, getErr := r.client.Get(ctx, fmt.Sprintf("%s/%s", apiKeysBasePath, keyID), nil); getErr == nil {
		if meta, perr := parseKeyResponse(getResp.Data); perr == nil {
			if meta.ServiceName != "" {
				state.ServiceName = types.StringValue(meta.ServiceName)
			}
			if meta.Description != "" {
				state.Description = types.StringValue(meta.Description)
			}
			state.Enabled = types.BoolValue(meta.Enabled)
			state.ExpiresAt = types.StringValue(meta.ExpiresAt)
			state.CreatedAt = types.StringValue(meta.CreatedAt)
			state.UpdatedAt = types.StringValue(meta.UpdatedAt)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.ID.ValueString()
	if keyID == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	apiResp, err := r.client.Get(ctx, fmt.Sprintf("%s/%s", apiKeysBasePath, keyID), nil)
	if err != nil {
		// 404 → resource was deleted out of band.
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read API key", err.Error())
		return
	}

	parsed, err := parseKeyResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse API key read response", err.Error())
		return
	}

	state.ServiceName = types.StringValue(parsed.ServiceName)
	state.Enabled = types.BoolValue(parsed.Enabled)
	if parsed.Description != "" {
		state.Description = types.StringValue(parsed.Description)
	}
	if parsed.ExpiresAt != "" {
		state.ExpiresAt = types.StringValue(parsed.ExpiresAt)
	}
	if parsed.UpdatedAt != "" {
		state.UpdatedAt = types.StringValue(parsed.UpdatedAt)
	}
	// Secret and CreatedAt are preserved from prior state via UseStateForUnknown.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state apiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := updateRequest{}
	if !plan.ServiceName.Equal(state.ServiceName) {
		body.ServiceName = plan.ServiceName.ValueString()
	}
	if !plan.Description.Equal(state.Description) {
		body.Description = plan.Description.ValueString()
	}
	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		body.Enabled = &v
	}

	apiResp, err := r.client.Patch(ctx, fmt.Sprintf("%s/%s", apiKeysBasePath, state.ID.ValueString()), body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update API key", err.Error())
		return
	}

	parsed, err := parseKeyResponse(apiResp.Data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to parse API key update response", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Secret = state.Secret
	plan.CreatedAt = state.CreatedAt
	plan.ServiceName = types.StringValue(parsed.ServiceName)
	plan.Enabled = types.BoolValue(parsed.Enabled)
	if parsed.Description != "" {
		plan.Description = types.StringValue(parsed.Description)
	}
	if parsed.ExpiresAt != "" {
		plan.ExpiresAt = types.StringValue(parsed.ExpiresAt)
	}
	if parsed.UpdatedAt != "" {
		plan.UpdatedAt = types.StringValue(parsed.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keyID := state.ID.ValueString()
	if keyID == "" {
		return
	}

	if _, err := r.client.Delete(ctx, fmt.Sprintf("%s/%s", apiKeysBasePath, keyID), nil); err != nil {
		if isNotFound(err) {
			// Already gone — treat as success.
			return
		}
		resp.Diagnostics.AddError("Failed to delete API key", err.Error())
		return
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// parseKeyResponse parses the npc-api response data into a keyResponse.
// Some endpoints wrap the key under a top-level field; this helper handles
// both shapes.
func parseKeyResponse(data json.RawMessage) (*keyResponse, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response data")
	}
	var key keyResponse
	if err := json.Unmarshal(data, &key); err != nil {
		// Try wrapped: {"apiKey": {...}}
		var wrapped struct {
			APIKey keyResponse `json:"apiKey"`
		}
		if werr := json.Unmarshal(data, &wrapped); werr == nil && wrapped.APIKey.ID != "" {
			return &wrapped.APIKey, nil
		}
		return nil, err
	}
	return &key, nil
}

// isNotFound reports whether err looks like a 404 from the npc-api.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "404") || strings.Contains(strings.ToLower(s), "not found")
}
