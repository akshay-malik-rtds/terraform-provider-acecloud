package auto_scaling_deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const basePath = "/auto-scaling/deployments"

var (
	_ resource.Resource              = &autoScalingDeploymentResource{}
	_ resource.ResourceWithConfigure = &autoScalingDeploymentResource{}
)

type autoScalingDeploymentResource struct {
	client *client.Client
}

// --- API types ---

type deploymentCreateRequest struct {
	Name             string              `json:"name"`
	Description      string              `json:"description,omitempty"`
	TemplateID       string              `json:"template_id"`
	DesiredCapacity  int64               `json:"desired_capacity"`
	MaxCapacity      int64               `json:"max_capacity"`
	NodesScaleCount  int64               `json:"nodes_scale_count"`
	ScalingParameter string              `json:"scaling_parameter"`
	MinThreshold     int64               `json:"min_threshold"`
	MaxThreshold     int64               `json:"max_threshold"`
	CoolDownTime     int64               `json:"cool_down_time"`
	UserEmail        []string            `json:"user_email"`
	IsIntegratedLB   bool                `json:"is_integrated_with_lb"`
	LBData           *lbDataCreateRequest `json:"lb_data,omitempty"`
}

type lbDataCreateRequest struct {
	LBName          string                      `json:"lb_name,omitempty"`
	Tags            []string                    `json:"tags,omitempty"`
	AssignPublicIP  bool                        `json:"assign_public_ip"`
	IsExistingLB    bool                        `json:"is_existing_lb"`
	LBID            string                      `json:"lb_id,omitempty"`
	LBVipPortID     string                      `json:"lb_vip_port_id,omitempty"`
	PublicNetworkID string                      `json:"public_network_id,omitempty"`
	Listener        *listenerCreateRequest      `json:"listener,omitempty"`
	Pool            *poolCreateRequest          `json:"pool,omitempty"`
	HealthMonitor   *healthMonitorCreateRequest `json:"health_monitor,omitempty"`
}

type listenerCreateRequest struct {
	ListenerName         string `json:"listener_name"`
	ListenerProtocol     string `json:"listener_protocol"`
	ListenerProtocolPort int64  `json:"listener_protocol_port"`
}

type poolCreateRequest struct {
	PoolProtocol     string `json:"pool_protocol"`
	PoolProtocolPort int64  `json:"pool_protocol_port"`
	LBAlgorithm      string `json:"lb_algorithm"`
}

type healthMonitorCreateRequest struct {
	MonitorProtocol   string `json:"monitor_protocol"`
	MonitorURLPath    string `json:"monitor_url_path"`
	MonitorHTTPMethod string `json:"monitor_http_method,omitempty"`
}

type createResponseID struct {
	ID string `json:"id"`
}

type deploymentAPIResponse struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	TemplateID       string `json:"template_id"`
	DesiredCapacity  int64  `json:"desired_capacity"`
	MaxCapacity      int64  `json:"max_capacity"`
	NodesScaleCount  int64  `json:"nodes_scale_count"`
	ScalingParameter string `json:"scaling_parameter"`
	MinThreshold     int64  `json:"min_threshold"`
	MaxThreshold     int64  `json:"max_threshold"`
	CoolDownTime     int64  `json:"cool_down_time"`
	Status           string `json:"status"`
	ErrorMessage     string `json:"error_message"`
	PanelURL         string `json:"panel_url"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func NewResource() resource.Resource {
	return &autoScalingDeploymentResource{}
}

func (r *autoScalingDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auto_scaling_deployment"
}

func (r *autoScalingDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = autoScalingDeploymentSchema()
}

func (r *autoScalingDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(ctx context.Context, plan *autoScalingDeploymentModel) deploymentCreateRequest {
	body := deploymentCreateRequest{
		Name:             plan.Name.ValueString(),
		TemplateID:       plan.TemplateID.ValueString(),
		DesiredCapacity:  plan.DesiredCapacity.ValueInt64(),
		MaxCapacity:      plan.MaxCapacity.ValueInt64(),
		NodesScaleCount:  plan.NodesScaleCount.ValueInt64(),
		ScalingParameter: plan.ScalingParameter.ValueString(),
		MinThreshold:     plan.MinThreshold.ValueInt64(),
		MaxThreshold:     plan.MaxThreshold.ValueInt64(),
		CoolDownTime:     plan.CoolDownTime.ValueInt64(),
		IsIntegratedLB:   plan.IsIntegratedLB.ValueBool(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		body.Description = plan.Description.ValueString()
	}

	if !plan.UserEmail.IsNull() && !plan.UserEmail.IsUnknown() {
		var emails []string
		plan.UserEmail.ElementsAs(ctx, &emails, false)
		body.UserEmail = emails
	}

	// Parse LB data from nested block if present
	if !plan.LBData.IsNull() && !plan.LBData.IsUnknown() {
		var lbModel lbDataModel
		plan.LBData.As(ctx, &lbModel, basetypes.ObjectAsOptions{})

		lb := &lbDataCreateRequest{
			AssignPublicIP: lbModel.AssignPublicIP.ValueBool(),
			IsExistingLB:   lbModel.IsExistingLB.ValueBool(),
		}

		if !lbModel.LBName.IsNull() && !lbModel.LBName.IsUnknown() {
			lb.LBName = lbModel.LBName.ValueString()
		}
		if !lbModel.LBID.IsNull() && !lbModel.LBID.IsUnknown() {
			lb.LBID = lbModel.LBID.ValueString()
		}
		if !lbModel.LBVipPortID.IsNull() && !lbModel.LBVipPortID.IsUnknown() {
			lb.LBVipPortID = lbModel.LBVipPortID.ValueString()
		}
		if !lbModel.PublicNetworkID.IsNull() && !lbModel.PublicNetworkID.IsUnknown() {
			lb.PublicNetworkID = lbModel.PublicNetworkID.ValueString()
		}
		if !lbModel.Tags.IsNull() && !lbModel.Tags.IsUnknown() {
			var tags []string
			lbModel.Tags.ElementsAs(ctx, &tags, false)
			lb.Tags = tags
		}

		// Listener sub-block
		if !lbModel.Listener.IsNull() && !lbModel.Listener.IsUnknown() {
			var listener listenerModel
			lbModel.Listener.As(ctx, &listener, basetypes.ObjectAsOptions{})
			lb.Listener = &listenerCreateRequest{
				ListenerName:         listener.ListenerName.ValueString(),
				ListenerProtocol:     listener.ListenerProtocol.ValueString(),
				ListenerProtocolPort: listener.ListenerProtocolPort.ValueInt64(),
			}
		}

		// Pool sub-block
		if !lbModel.Pool.IsNull() && !lbModel.Pool.IsUnknown() {
			var pool poolModel
			lbModel.Pool.As(ctx, &pool, basetypes.ObjectAsOptions{})
			lb.Pool = &poolCreateRequest{
				PoolProtocol:     pool.PoolProtocol.ValueString(),
				PoolProtocolPort: pool.PoolProtocolPort.ValueInt64(),
				LBAlgorithm:      pool.LBAlgorithm.ValueString(),
			}
		}

		// Health monitor sub-block
		if !lbModel.HealthMonitor.IsNull() && !lbModel.HealthMonitor.IsUnknown() {
			var hm healthMonitorModel
			lbModel.HealthMonitor.As(ctx, &hm, basetypes.ObjectAsOptions{})
			lb.HealthMonitor = &healthMonitorCreateRequest{
				MonitorProtocol: hm.MonitorProtocol.ValueString(),
				MonitorURLPath:  hm.MonitorURLPath.ValueString(),
			}
			if !hm.MonitorHTTPMethod.IsNull() && !hm.MonitorHTTPMethod.IsUnknown() {
				lb.HealthMonitor.MonitorHTTPMethod = hm.MonitorHTTPMethod.ValueString()
			}
		}

		body.LBData = lb
	}

	return body
}

func mapAPIResponseToState(model *autoScalingDeploymentModel, apiResp *deploymentAPIResponse) {
	model.ID = types.StringValue(apiResp.ID)
	model.Name = types.StringValue(apiResp.Name)
	model.TemplateID = types.StringValue(apiResp.TemplateID)
	model.DesiredCapacity = types.Int64Value(apiResp.DesiredCapacity)
	model.MaxCapacity = types.Int64Value(apiResp.MaxCapacity)
	model.NodesScaleCount = types.Int64Value(apiResp.NodesScaleCount)
	model.ScalingParameter = types.StringValue(apiResp.ScalingParameter)
	model.MinThreshold = types.Int64Value(apiResp.MinThreshold)
	model.MaxThreshold = types.Int64Value(apiResp.MaxThreshold)
	model.CoolDownTime = types.Int64Value(apiResp.CoolDownTime)

	if apiResp.Description != "" {
		model.Description = types.StringValue(apiResp.Description)
	} else if model.Description.IsNull() {
		model.Description = types.StringNull()
	}

	// Computed fields — always set to known values
	if apiResp.Status != "" {
		model.Status = types.StringValue(apiResp.Status)
	} else {
		model.Status = types.StringValue("")
	}

	if apiResp.ErrorMessage != "" {
		model.ErrorMessage = types.StringValue(apiResp.ErrorMessage)
	} else {
		model.ErrorMessage = types.StringValue("")
	}

	if apiResp.PanelURL != "" {
		model.PanelURL = types.StringValue(apiResp.PanelURL)
	} else {
		model.PanelURL = types.StringValue("")
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

func (r *autoScalingDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan autoScalingDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	apiResp, err := r.client.Post(ctx, basePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create auto scaling deployment", err.Error())
		return
	}

	var created createResponseID
	if err := json.Unmarshal(apiResp.Data, &created); err != nil {
		resp.Diagnostics.AddError("Failed to parse auto scaling deployment response", err.Error())
		return
	}

	if created.ID == "" {
		resp.Diagnostics.AddError("Failed to create auto scaling deployment", "API returned empty ID")
		return
	}

	plan.ID = types.StringValue(created.ID)

	// Wait for deployment to become ACTIVE
	readPath := fmt.Sprintf("%s/%s", basePath, created.ID)
	result, waitErr := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			readResp, err := r.client.Get(ctx, readPath, nil)
			if err != nil {
				return nil, err
			}
			var dep deploymentAPIResponse
			if err := json.Unmarshal(readResp.Data, &dep); err != nil {
				return nil, err
			}
			return &wait.StatusResult{Status: dep.Status, Data: &dep}, nil
		},
		TargetStatus: []string{"ACTIVE", "CREATED"},
		ErrorStatus:  []string{"ERROR"},
		Timeout:      10 * time.Minute,
		PollInterval: 10 * time.Second,
	})

	if waitErr != nil {
		// Still save state with what we have
		plan.Status = types.StringValue("CREATING")
		plan.ErrorMessage = types.StringValue(waitErr.Error())
		plan.PanelURL = types.StringValue("")
		plan.CreatedAt = types.StringValue("")
		plan.UpdatedAt = types.StringValue("")
		resp.Diagnostics.AddWarning("Auto scaling deployment creation may still be in progress", waitErr.Error())
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if dep, ok := result.Data.(*deploymentAPIResponse); ok {
		mapAPIResponseToState(&plan, dep)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *autoScalingDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state autoScalingDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read auto scaling deployment", err.Error())
		return
	}

	var dep deploymentAPIResponse
	if err := json.Unmarshal(apiResp.Data, &dep); err != nil {
		resp.Diagnostics.AddError("Failed to parse auto scaling deployment response", err.Error())
		return
	}

	mapAPIResponseToState(&state, &dep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *autoScalingDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No update endpoint — all fields have ForceNew, so Terraform will
	// destroy + recreate automatically. This method should never be called.
	resp.Diagnostics.AddError(
		"Failed to update auto scaling deployment",
		"Update is not supported. All fields require recreation.",
	)
}

func (r *autoScalingDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state autoScalingDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, state.ID.ValueString())
	_, err := r.client.Delete(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete auto scaling deployment", err.Error())
	}
}
