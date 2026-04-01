package caas_deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/acecloud/terraform-provider-acecloud/internal/client"
	"github.com/acecloud/terraform-provider-acecloud/internal/wait"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

const basePath = "/caas/deployments"

var (
	_ resource.Resource              = &caasDeploymentResource{}
	_ resource.ResourceWithConfigure = &caasDeploymentResource{}
)

type caasDeploymentResource struct {
	client *client.Client
}

// --- API types ---

type deploymentCreateRequest struct {
	Name        string                   `json:"name"`
	Type        string                   `json:"type,omitempty"`
	Image       imageCreateRequest       `json:"image"`
	Command     []string                 `json:"command"`
	Resources   resourcesCreateRequest   `json:"resources"`
	Env         []envCreateRequest       `json:"env,omitempty"`
	EnvSecrets  []string                 `json:"envSecrets,omitempty"`
	Autoscaling *autoscalingCreateRequest `json:"autoscaling,omitempty"`
	Volumes     []volumeCreateRequest    `json:"volumes,omitempty"`
	Networking  networkingCreateRequest  `json:"networking"`
}

type imageCreateRequest struct {
	Type      string   `json:"type"`
	Reference string   `json:"reference"`
	Secrets   []string `json:"secrets,omitempty"`
}

type resourcesCreateRequest struct {
	CPU          *float64 `json:"cpu,omitempty"`
	Memory       string   `json:"memory,omitempty"`
	ReplicaCount int64    `json:"replicaCount"`
	FlavorID     string   `json:"flavorId,omitempty"`
}

type networkingCreateRequest struct {
	ExternalAccess       bool                  `json:"externalAccess"`
	Ports                []portCreateRequest   `json:"ports,omitempty"`
	EndpointAccess       string                `json:"endpointAccess,omitempty"`
	CIDRBlock            []string              `json:"cidrBlock,omitempty"`
	XForwardedFor        bool                  `json:"xForwardedFor"`
	UseExistingNetwork   *bool                 `json:"useExistingNetwork,omitempty"`
	NetworkID            string                `json:"networkId,omitempty"`
	CreateNewNetworkCIDR string                `json:"createNewNetworkCidr,omitempty"`
}

type portCreateRequest struct {
	Name          string `json:"name"`
	Protocol      string `json:"protocol"`
	ContainerPort int64  `json:"containerPort"`
	ExposedPort   *int64 `json:"exposedPort,omitempty"`
}

type autoscalingCreateRequest struct {
	Enabled                bool     `json:"enabled"`
	MinReplicas            *int64   `json:"minReplicas,omitempty"`
	MaxReplicas            *int64   `json:"maxReplicas,omitempty"`
	CPUTargetPercentage    *float64 `json:"cpuTargetPercentage,omitempty"`
	MemoryTargetPercentage *float64 `json:"memoryTargetPercentage,omitempty"`
}

type envCreateRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type volumeCreateRequest struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	Size      string `json:"size"`
}

// --- API response types ---

type deploymentAPIResponse struct {
	Name              string                 `json:"name"`
	Type              string                 `json:"type"`
	DeploymentDetails deploymentDetailsResp  `json:"deploymentDetails"`
}

type deploymentDetailsResp struct {
	Status           string   `json:"status"`
	CreatedAt        string   `json:"createdAt"`
	PrivateEndpoints []string `json:"privateEndpoints"`
	PublicEndpoints  []string `json:"publicEndpoints"`
}

func NewResource() resource.Resource {
	return &caasDeploymentResource{}
}

func (r *caasDeploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caas_deployment"
}

func (r *caasDeploymentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = caasDeploymentSchema()
}

func (r *caasDeploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func buildCreateRequest(ctx context.Context, plan *caasDeploymentModel) deploymentCreateRequest {
	body := deploymentCreateRequest{
		Name:    plan.Name.ValueString(),
		Type:    plan.Type.ValueString(),
		Command: []string{}, // API requires command to always be present as an array
	}

	// Command
	if !plan.Command.IsNull() && !plan.Command.IsUnknown() {
		var cmds []string
		plan.Command.ElementsAs(ctx, &cmds, false)
		body.Command = cmds
	}

	// EnvSecrets
	if !plan.EnvSecrets.IsNull() && !plan.EnvSecrets.IsUnknown() {
		var secrets []string
		plan.EnvSecrets.ElementsAs(ctx, &secrets, false)
		body.EnvSecrets = secrets
	}

	// Image block
	if !plan.Image.IsNull() && !plan.Image.IsUnknown() {
		var img imageModel
		plan.Image.As(ctx, &img, basetypes.ObjectAsOptions{})
		body.Image = imageCreateRequest{
			Type:      img.Type.ValueString(),
			Reference: img.Reference.ValueString(),
		}
		if !img.Secrets.IsNull() && !img.Secrets.IsUnknown() {
			var secrets []string
			img.Secrets.ElementsAs(ctx, &secrets, false)
			body.Image.Secrets = secrets
		}
	}

	// Resources block
	if !plan.Resources.IsNull() && !plan.Resources.IsUnknown() {
		var res resourcesModel
		plan.Resources.As(ctx, &res, basetypes.ObjectAsOptions{})
		body.Resources = resourcesCreateRequest{
			ReplicaCount: res.ReplicaCount.ValueInt64(),
		}
		if !res.CPU.IsNull() && !res.CPU.IsUnknown() {
			v := res.CPU.ValueFloat64()
			body.Resources.CPU = &v
		}
		if !res.Memory.IsNull() && !res.Memory.IsUnknown() {
			body.Resources.Memory = res.Memory.ValueString()
		}
		if !res.FlavorID.IsNull() && !res.FlavorID.IsUnknown() {
			body.Resources.FlavorID = res.FlavorID.ValueString()
		}
	}

	// Networking block
	if !plan.Networking.IsNull() && !plan.Networking.IsUnknown() {
		var net networkingModel
		plan.Networking.As(ctx, &net, basetypes.ObjectAsOptions{})
		body.Networking = networkingCreateRequest{
			ExternalAccess: net.ExternalAccess.ValueBool(),
		}
		if !net.EndpointAccess.IsNull() && !net.EndpointAccess.IsUnknown() {
			body.Networking.EndpointAccess = net.EndpointAccess.ValueString()
		}
		if !net.CIDRBlock.IsNull() && !net.CIDRBlock.IsUnknown() {
			var cidrs []string
			net.CIDRBlock.ElementsAs(ctx, &cidrs, false)
			body.Networking.CIDRBlock = cidrs
		}
		if !net.XForwardedFor.IsNull() && !net.XForwardedFor.IsUnknown() {
			body.Networking.XForwardedFor = net.XForwardedFor.ValueBool()
		}
		if !net.UseExistingNetwork.IsNull() && !net.UseExistingNetwork.IsUnknown() {
			v := net.UseExistingNetwork.ValueBool()
			body.Networking.UseExistingNetwork = &v
		}
		if !net.NetworkID.IsNull() && !net.NetworkID.IsUnknown() {
			body.Networking.NetworkID = net.NetworkID.ValueString()
		}
		if !net.CreateNewNetworkCIDR.IsNull() && !net.CreateNewNetworkCIDR.IsUnknown() {
			body.Networking.CreateNewNetworkCIDR = net.CreateNewNetworkCIDR.ValueString()
		}

		// Ports
		if !net.Port.IsNull() && !net.Port.IsUnknown() {
			var ports []portModel
			net.Port.ElementsAs(ctx, &ports, false)
			for _, p := range ports {
				pr := portCreateRequest{
					Name:          p.Name.ValueString(),
					Protocol:      p.Protocol.ValueString(),
					ContainerPort: p.ContainerPort.ValueInt64(),
				}
				if !p.ExposedPort.IsNull() && !p.ExposedPort.IsUnknown() {
					v := p.ExposedPort.ValueInt64()
					pr.ExposedPort = &v
				}
				body.Networking.Ports = append(body.Networking.Ports, pr)
			}
		}
	}

	// Autoscaling block
	if !plan.Autoscaling.IsNull() && !plan.Autoscaling.IsUnknown() {
		var as autoscalingModel
		plan.Autoscaling.As(ctx, &as, basetypes.ObjectAsOptions{})
		body.Autoscaling = &autoscalingCreateRequest{
			Enabled: as.Enabled.ValueBool(),
		}
		if !as.MinReplicas.IsNull() && !as.MinReplicas.IsUnknown() {
			v := as.MinReplicas.ValueInt64()
			body.Autoscaling.MinReplicas = &v
		}
		if !as.MaxReplicas.IsNull() && !as.MaxReplicas.IsUnknown() {
			v := as.MaxReplicas.ValueInt64()
			body.Autoscaling.MaxReplicas = &v
		}
		if !as.CPUTargetPercentage.IsNull() && !as.CPUTargetPercentage.IsUnknown() {
			v := as.CPUTargetPercentage.ValueFloat64()
			body.Autoscaling.CPUTargetPercentage = &v
		}
		if !as.MemoryTargetPercentage.IsNull() && !as.MemoryTargetPercentage.IsUnknown() {
			v := as.MemoryTargetPercentage.ValueFloat64()
			body.Autoscaling.MemoryTargetPercentage = &v
		}
	}

	// Env blocks
	if !plan.Env.IsNull() && !plan.Env.IsUnknown() {
		var envs []envModel
		plan.Env.ElementsAs(ctx, &envs, false)
		for _, e := range envs {
			body.Env = append(body.Env, envCreateRequest{
				Name:  e.Name.ValueString(),
				Value: e.Value.ValueString(),
			})
		}
	}

	// Volume blocks
	if !plan.Volume.IsNull() && !plan.Volume.IsUnknown() {
		var vols []volumeModel
		plan.Volume.ElementsAs(ctx, &vols, false)
		for _, v := range vols {
			body.Volumes = append(body.Volumes, volumeCreateRequest{
				Name:      v.Name.ValueString(),
				MountPath: v.MountPath.ValueString(),
				Size:      v.Size.ValueString(),
			})
		}
	}

	return body
}

func mapAPIResponseToState(ctx context.Context, model *caasDeploymentModel, apiResp *deploymentAPIResponse) {
	model.ID = types.StringValue(apiResp.Name)
	model.Name = types.StringValue(apiResp.Name)

	if apiResp.Type != "" {
		model.Type = types.StringValue(apiResp.Type)
	}

	// Computed fields from deploymentDetails
	if apiResp.DeploymentDetails.Status != "" {
		model.Status = types.StringValue(apiResp.DeploymentDetails.Status)
	} else {
		model.Status = types.StringValue("")
	}

	if apiResp.DeploymentDetails.CreatedAt != "" {
		model.CreatedAt = types.StringValue(apiResp.DeploymentDetails.CreatedAt)
	} else {
		model.CreatedAt = types.StringValue("")
	}

	// Endpoints
	if len(apiResp.DeploymentDetails.PrivateEndpoints) > 0 {
		list, _ := types.ListValueFrom(ctx, types.StringType, apiResp.DeploymentDetails.PrivateEndpoints)
		model.PrivateEndpoints = list
	} else {
		model.PrivateEndpoints, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}

	if len(apiResp.DeploymentDetails.PublicEndpoints) > 0 {
		list, _ := types.ListValueFrom(ctx, types.StringType, apiResp.DeploymentDetails.PublicEndpoints)
		model.PublicEndpoints = list
	} else {
		model.PublicEndpoints, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	}
}

func (r *caasDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan caasDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan)

	_, err := r.client.Post(ctx, basePath, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create CaaS deployment", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.Name.ValueString())

	// Wait for deployment to become Active
	readPath := fmt.Sprintf("%s/%s", basePath, url.PathEscape(plan.Name.ValueString()))
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
			return &wait.StatusResult{Status: dep.DeploymentDetails.Status, Data: &dep}, nil
		},
		TargetStatus: []string{"Active"},
		ErrorStatus:  []string{"Error", "DeletionFailed"},
	})

	if waitErr != nil {
		plan.Status = types.StringValue("Provisioning")
		plan.PrivateEndpoints = types.ListNull(types.StringType)
		plan.PublicEndpoints = types.ListNull(types.StringType)
		plan.CreatedAt = types.StringValue("")
		resp.Diagnostics.AddWarning("CaaS deployment may still be provisioning", waitErr.Error())
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if dep, ok := result.Data.(*deploymentAPIResponse); ok {
		mapAPIResponseToState(ctx, &plan, dep)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *caasDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state caasDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, url.PathEscape(state.Name.ValueString()))
	apiResp, err := r.client.Get(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read CaaS deployment", err.Error())
		return
	}

	var dep deploymentAPIResponse
	if err := json.Unmarshal(apiResp.Data, &dep); err != nil {
		resp.Diagnostics.AddError("Failed to parse CaaS deployment response", err.Error())
		return
	}

	mapAPIResponseToState(ctx, &state, &dep)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *caasDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan caasDeploymentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state caasDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := buildCreateRequest(ctx, &plan) // Update uses same body structure

	path := fmt.Sprintf("%s/%s", basePath, url.PathEscape(state.Name.ValueString()))
	_, err := r.client.Put(ctx, path, body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update CaaS deployment", err.Error())
		return
	}

	// Wait for Active after update
	result, waitErr := wait.WaitForStatus(ctx, wait.WaitForStatusOpts{
		Refresh: func(ctx context.Context) (*wait.StatusResult, error) {
			readResp, err := r.client.Get(ctx, path, nil)
			if err != nil {
				return nil, err
			}
			var dep deploymentAPIResponse
			if err := json.Unmarshal(readResp.Data, &dep); err != nil {
				return nil, err
			}
			return &wait.StatusResult{Status: dep.DeploymentDetails.Status, Data: &dep}, nil
		},
		TargetStatus: []string{"Active"},
		ErrorStatus:  []string{"Error", "DeletionFailed"},
	})

	plan.ID = state.ID
	if waitErr != nil {
		plan.Status = types.StringValue("Provisioning")
		plan.PrivateEndpoints = state.PrivateEndpoints
		plan.PublicEndpoints = state.PublicEndpoints
		plan.CreatedAt = state.CreatedAt
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if dep, ok := result.Data.(*deploymentAPIResponse); ok {
		mapAPIResponseToState(ctx, &plan, dep)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *caasDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state caasDeploymentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("%s/%s", basePath, url.PathEscape(state.Name.ValueString()))
	_, err := r.client.Delete(ctx, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete CaaS deployment", err.Error())
	}
}
