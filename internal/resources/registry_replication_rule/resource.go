package registry_replication_rule

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const basePath = "/ace-registry/replication-rules/policies"

var (
	_ resource.Resource              = &registryReplicationRuleResource{}
	_ resource.ResourceWithConfigure = &registryReplicationRuleResource{}
)

type registryReplicationRuleResource struct {
	client *client.Client
}

// --- API request types ---

type replicationRuleCreateRequest struct {
	Name              string           `json:"name"`
	Description       string           `json:"description,omitempty"`
	Enabled           bool             `json:"enabled"`
	SrcRegistry       *registryRequest `json:"src_registry,omitempty"`
	DestRegistry      *registryRequest `json:"dest_registry,omitempty"`
	DestNamespace     string           `json:"dest_namespace,omitempty"`
	Trigger           *triggerRequest  `json:"trigger,omitempty"`
	Filters           []filterRequest  `json:"filters,omitempty"`
	ReplicateDeletion *bool            `json:"replicate_deletion,omitempty"`
	Override          *bool            `json:"override,omitempty"`
	Speed             *int64           `json:"speed,omitempty"`
}

type registryRequest struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Type       string `json:"type"`
	UpdateTime string `json:"update_time,omitempty"`
}

type triggerRequest struct {
	Type            string                `json:"type"`
	TriggerSettings triggerSettingsRequest `json:"trigger_settings"`
}

type triggerSettingsRequest struct {
	Cron string `json:"cron"`
}

type filterRequest struct {
	Type       string `json:"type"`
	Value      string `json:"value"`
	Decoration string `json:"decoration,omitempty"`
}

// --- API response types ---

type replicationRuleAPIResponse struct {
	ID                int64                `json:"id"`
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	Enabled           bool                 `json:"enabled"`
	SrcRegistry       *registryAPIResponse `json:"src_registry"`
	DestRegistry      *registryAPIResponse `json:"dest_registry"`
	DestNamespace     string               `json:"dest_namespace"`
	Trigger           *triggerAPIResponse  `json:"trigger"`
	Filters           []filterAPIResponse  `json:"filters"`
	ReplicateDeletion bool                 `json:"replicate_deletion"`
	Override          bool                 `json:"override"`
	Speed             int64                `json:"speed"`
	CreatedAt         string               `json:"creation_time"`
	UpdatedAt         string               `json:"update_time"`
}

type registryAPIResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

type triggerAPIResponse struct {
	Type            string                  `json:"type"`
	TriggerSettings *triggerSettingsResponse `json:"trigger_settings"`
}

type triggerSettingsResponse struct {
	Cron string `json:"cron"`
}

type filterAPIResponse struct {
	Type       string `json:"type"`
	Value      string `json:"value"`
	Decoration string `json:"decoration"`
}

func NewResource() resource.Resource {
	return &registryReplicationRuleResource{}
}

func (r *registryReplicationRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry_replication_rule"
}

func (r *registryReplicationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = replicationRuleSchema()
}

func (r *registryReplicationRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// --- Build helpers ---

func buildRegistryRequest(ctx context.Context, obj types.Object) *registryRequest {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var reg registryModel
	obj.As(ctx, &reg, basetypes.ObjectAsOptions{})
	return &registryRequest{
		ID:         reg.ID.ValueInt64(),
		Name:       reg.Name.ValueString(),
		URL:        reg.URL.ValueString(),
		Type:       reg.Type.ValueString(),
		UpdateTime: "2026-01-01T00:00:00.000Z",
	}
}

func buildTriggerRequest(ctx context.Context, obj types.Object) *triggerRequest {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var trig triggerModel
	obj.As(ctx, &trig, basetypes.ObjectAsOptions{})
	req := &triggerRequest{
		Type:            trig.Type.ValueString(),
		TriggerSettings: triggerSettingsRequest{},
	}
	if !trig.Cron.IsNull() && !trig.Cron.IsUnknown() {
		req.TriggerSettings.Cron = trig.Cron.ValueString()
	}
	return req
}

func buildFilters(ctx context.Context, list types.List) []filterRequest {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var filters []filterModel
	list.ElementsAs(ctx, &filters, false)
	var result []filterRequest
	for _, f := range filters {
		fr := filterRequest{
			Type:  f.Type.ValueString(),
			Value: f.Value.ValueString(),
		}
		if !f.Decoration.IsNull() && !f.Decoration.IsUnknown() {
			fr.Decoration = f.Decoration.ValueString()
		}
		result = append(result, fr)
	}
	return result
}

func buildCreateRequest(ctx context.Context, plan *replicationRuleModel) replicationRuleCreateRequest {
	body := replicationRuleCreateRequest{
		Name:    plan.Name.ValueString(),
		Enabled: plan.Enabled.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}

	body.SrcRegistry = buildRegistryRequest(ctx, plan.SrcRegistry)
	body.DestRegistry = buildRegistryRequest(ctx, plan.DestRegistry)

	if !plan.DestNamespace.IsNull() && !plan.DestNamespace.IsUnknown() {
		body.DestNamespace = plan.DestNamespace.ValueString()
	}

	body.Trigger = buildTriggerRequest(ctx, plan.Trigger)
	body.Filters = buildFilters(ctx, plan.Filter)

	if !plan.ReplicateDeletion.IsNull() && !plan.ReplicateDeletion.IsUnknown() {
		v := plan.ReplicateDeletion.ValueBool()
		body.ReplicateDeletion = &v
	}

	if !plan.Override.IsNull() && !plan.Override.IsUnknown() {
		v := plan.Override.ValueBool()
		body.Override = &v
	}

	if !plan.Speed.IsNull() && !plan.Speed.IsUnknown() {
		v := plan.Speed.ValueInt64()
		body.Speed = &v
	}

	return body
}

func buildUpdateRequest(ctx context.Context, plan *replicationRuleModel) replicationRuleCreateRequest {
	return buildCreateRequest(ctx, plan)
}

// --- Map response to state ---

func mapAPIResponseToState(ctx context.Context, model *replicationRuleModel, apiResp *replicationRuleAPIResponse) {
	model.ID = types.StringValue(fmt.Sprintf("%d", apiResp.ID))
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	model.Enabled = types.BoolValue(apiResp.Enabled)

	// src_registry — preserve user's configured values (API may return different name/url)
	// Only update from API on first Read (when model has no src_registry yet)
	if model.SrcRegistry.IsNull() || model.SrcRegistry.IsUnknown() {
		if apiResp.SrcRegistry != nil {
			srcObj, _ := types.ObjectValueFrom(ctx, registryAttrTypes(), &registryModel{
				ID:   types.Int64Value(apiResp.SrcRegistry.ID),
				Name: types.StringValue(apiResp.SrcRegistry.Name),
				URL:  types.StringValue(apiResp.SrcRegistry.URL),
				Type: types.StringValue(apiResp.SrcRegistry.Type),
			})
			model.SrcRegistry = srcObj
		}
	}

	// dest_registry — NEVER update from API when user didn't configure it.
	// The API always returns a local dest_registry even when not set.
	// Preserve user's configured value (or null if not configured).

	// dest_namespace
	if apiResp.DestNamespace != "" {
		model.DestNamespace = types.StringValue(apiResp.DestNamespace)
	} else if model.DestNamespace.IsNull() {
		model.DestNamespace = types.StringNull()
	}

	// trigger
	if apiResp.Trigger != nil {
		cronVal := types.StringNull()
		if apiResp.Trigger.TriggerSettings != nil && apiResp.Trigger.TriggerSettings.Cron != "" {
			cronVal = types.StringValue(apiResp.Trigger.TriggerSettings.Cron)
		}
		trigObj, _ := types.ObjectValueFrom(ctx, triggerAttrTypes(), &triggerModel{
			Type: types.StringValue(apiResp.Trigger.Type),
			Cron: cronVal,
		})
		model.Trigger = trigObj
	}

	// filter
	if len(apiResp.Filters) > 0 {
		var filters []filterModel
		for _, f := range apiResp.Filters {
			fm := filterModel{
				Type:  types.StringValue(f.Type),
				Value: types.StringValue(f.Value),
			}
			if f.Decoration != "" {
				fm.Decoration = types.StringValue(f.Decoration)
			} else {
				fm.Decoration = types.StringNull()
			}
			filters = append(filters, fm)
		}
		filterList, _ := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: filterAttrTypes()}, filters)
		model.Filter = filterList
	} else {
		model.Filter = types.ListNull(types.ObjectType{AttrTypes: filterAttrTypes()})
	}

	model.ReplicateDeletion = types.BoolValue(apiResp.ReplicateDeletion)
	model.Override = types.BoolValue(apiResp.Override)

	if apiResp.Speed != 0 {
		model.Speed = types.Int64Value(apiResp.Speed)
	} else if model.Speed.IsNull() || model.Speed.IsUnknown() {
		model.Speed = types.Int64Value(0)
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

// --- CRUD ---

func (r *registryReplicationRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan replicationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	_, err := r.client.Post(ctx, basePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create registry replication rule", err.Error())
		return
	}

	// Create response returns {message: "Successfully Created"} with no data.
	// Poll the list endpoint to find the created rule by name.
	targetName := plan.Name.ValueString()
	listResp, err := r.client.Get(ctx, basePath, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list registry replication rules", err.Error())
		return
	}

	var rules []replicationRuleAPIResponse
	if err := json.Unmarshal(listResp.Data, &rules); err != nil {
		resp.Diagnostics.AddError("Failed to parse registry replication rule list", err.Error())
		return
	}

	var fullResp *replicationRuleAPIResponse
	for i, rule := range rules {
		if rule.Name == targetName {
			fullResp = &rules[i]
			break
		}
	}
	if fullResp == nil {
		resp.Diagnostics.AddError("Failed to create registry replication rule",
			fmt.Sprintf("Rule %q was not found in the list after creation", targetName))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", fullResp.ID))

	mapAPIResponseToState(ctx, &plan, fullResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryReplicationRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state replicationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read registry replication rule", err.Error())
		return
	}

	var ruleResp replicationRuleAPIResponse
	if err := json.Unmarshal(apiResp.Data, &ruleResp); err != nil {
		resp.Diagnostics.AddError("Failed to parse registry replication rule response", err.Error())
		return
	}

	mapAPIResponseToState(ctx, &state, &ruleResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *registryReplicationRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan replicationRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state replicationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildUpdateRequest(ctx, &plan)

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	_, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update registry replication rule", err.Error())
		return
	}

	// Read back to get full updated state
	readResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read registry replication rule", err.Error())
		return
	}

	var ruleResp replicationRuleAPIResponse
	if err := json.Unmarshal(readResp.Data, &ruleResp); err != nil {
		resp.Diagnostics.AddError("Failed to parse registry replication rule response", err.Error())
		return
	}

	mapAPIResponseToState(ctx, &plan, &ruleResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *registryReplicationRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state replicationRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	_, err := r.client.Delete(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete registry replication rule", err.Error())
	}
}
