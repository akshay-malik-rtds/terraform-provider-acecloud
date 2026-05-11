package key_pair

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const apiPath = "/cloud/key-pairs"

// Ensure the resource satisfies the expected interfaces.
var (
	_ resource.Resource              = &keyPairResource{}
	_ resource.ResourceWithConfigure = &keyPairResource{}
)

// keyPairResource is the resource implementation.
type keyPairResource struct {
	client *client.Client
}

// NewResource returns a new key pair resource factory.
func NewResource() resource.Resource {
	return &keyPairResource{}
}

func (r *keyPairResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_pair"
}

func (r *keyPairResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *keyPairResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan keyPairModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	// The npc-api keypair create DTO maps the JSON input field `key` to the
	// internal `public_key` (via `@Expose({ name: 'key' })`). Sending
	// `public_key` is silently dropped, which causes the backend to generate
	// a fresh keypair instead of importing the user's public key.
	if !plan.PublicKey.IsNull() && !plan.PublicKey.IsUnknown() {
		body["key"] = plan.PublicKey.ValueString()
	}

	apiResp, err := r.client.Post(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create key pair", err.Error())
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse create response", err.Error())
		return
	}

	id, ok := result["id"].(string)
	if !ok {
		resp.Diagnostics.AddError("Failed to parse key pair ID", "ID not found in response")
		return
	}
	plan.ID = types.StringValue(id)

	if v, ok := result["fingerprint"].(string); ok {
		plan.Fingerprint = types.StringValue(v)
	} else {
		plan.Fingerprint = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *keyPairResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state keyPairModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", apiPath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read key pair", err.Error())
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		resp.Diagnostics.AddError("Failed to parse read response", err.Error())
		return
	}

	if v, ok := result["name"].(string); ok {
		state.Name = types.StringValue(v)
	}
	// public_key: the npc-api returns this field as `publicKey` (camelCase).
	// Only mirror the API value into state when the user originally provided
	// a public_key in their config. For generated keypairs, the backend
	// returns "...Generated-by-Nova" which would otherwise drift against an
	// empty config and force replacement on the next refresh.
	if !state.PublicKey.IsNull() && !state.PublicKey.IsUnknown() {
		if v, ok := result["publicKey"].(string); ok && v != "" {
			state.PublicKey = types.StringValue(v)
		} else if v, ok := result["public_key"].(string); ok && v != "" {
			state.PublicKey = types.StringValue(v)
		}
	}

	if v, ok := result["fingerprint"].(string); ok {
		state.Fingerprint = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *keyPairResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Key pairs do not support update. Any change triggers destroy and recreate.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Key pair resources do not support in-place updates. Changes will trigger a destroy and recreate.",
	)
}

func (r *keyPairResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state keyPairModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Matches CLI pattern: {"key": "id", "values": [...]}
	body := map[string]interface{}{
		"key":    "id",
		"values": []string{state.ID.ValueString()},
	}

	_, err := r.client.Delete(ctx, apiPath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete key pair", err.Error())
		return
	}
}
