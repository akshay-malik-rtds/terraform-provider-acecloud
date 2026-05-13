package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/auth"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/client"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/datasources/flavors"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/datasources/images"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/datasources/routers"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/datasources/security_groups"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/datasources/vpcs"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/api_key"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/auto_scaling_deployment"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/auto_scaling_template"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/floating_ip"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/floating_ip_association"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/instance"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/key_pair"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/lb_health_monitor"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/lb_listener"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/lb_pool"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/lb_pool_member"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/load_balancer"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/router"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/router_interface"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/security_group"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/snapshot"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/subnet"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/volume"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/volume_attachment"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/volume_backup"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/vpc"
	"github.com/hashicorp/terraform-plugin-framework-validators/providervalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// defaultAPIURL is the production AceCloud API endpoint used when neither
// the api_url provider argument nor the ACECLOUD_API_URL environment
// variable is set.
const defaultAPIURL = "https://customer.acecloud.ai/api/v1/"

// Ensure AceCloudProvider satisfies the provider interface.
var _ provider.Provider = &AceCloudProvider{}

// AceCloudProvider implements the Ace Cloud Terraform provider.
type AceCloudProvider struct {
	version string
}

// AceCloudProviderModel maps provider schema to Go types.
type AceCloudProviderModel struct {
	APIURL            types.String `tfsdk:"api_url"`
	APIKeyID          types.String `tfsdk:"api_key_id"`
	APIKeySecret      types.String `tfsdk:"api_key_secret"`
	APIKeyServiceName types.String `tfsdk:"api_key_service_name"`
	Region            types.String `tfsdk:"region"`
	ProjectID         types.String `tfsdk:"project_id"`
	ACEConfigPath     types.String `tfsdk:"ace_config_path"`
}

// New returns a provider.Provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AceCloudProvider{
			version: version,
		}
	}
}

func (p *AceCloudProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "acecloud"
	resp.Version = p.version
}

func (p *AceCloudProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Ace Cloud infrastructure. Manages compute instances, volumes, snapshots, backups, VPCs, subnets, routers, security groups, floating IPs, key pairs, and load balancers via the Ace Cloud API.\n\n" +
			"Authenticate by setting `api_key_id`, `api_key_secret`, and `api_key_service_name` in the provider block, or via the matching `ACECLOUD_API_KEY_*` environment variables. Create a key from the AceCloud console.",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Description: "Base URL of the Ace Cloud API. Optional; defaults to `https://customer.acecloud.ai/api/v1/`. Override via the provider block or the `ACECLOUD_API_URL` environment variable when targeting a non-production endpoint.",
				Optional:    true,
			},
			"api_key_id": schema.StringAttribute{
				Description: "API key identifier (the part before the dot in `keyId.secret`). Can also be set via the `ACECLOUD_API_KEY_ID` environment variable. API keys are long-lived credentials suitable for automation; create one with `ace api-key create`. Must be set together with `api_key_secret` and `api_key_service_name`.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_key_secret": schema.StringAttribute{
				Description: "API key secret (the part after the dot in `keyId.secret`). Can also be set via the `ACECLOUD_API_KEY_SECRET` environment variable. The secret is only shown once at key creation; if lost, regenerate via `ace api-key revive <id>`. Must be set together with `api_key_id` and `api_key_service_name`.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_key_service_name": schema.StringAttribute{
				Description: "Service name attached to API key requests. Required when using API key authentication; must match the service name supplied at key creation. Can also be set via the `ACECLOUD_API_KEY_SERVICE_NAME` environment variable.",
				Optional:    true,
			},
			"region": schema.StringAttribute{
				Description: "Cloud region the provider operates against (for example `ap-south-noi-1` or `ap-south-mum-1`). Required. Set via the provider block or the `ACECLOUD_REGION` environment variable. There is no auto-selection — every configuration explicitly names its target region.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "Cloud project UUID the provider operates against. Required. Set via the provider block or the `ACECLOUD_PROJECT_ID` environment variable. There is no auto-selection — every configuration explicitly names its target project.",
				Required:    true,
			},
			"ace_config_path": schema.StringAttribute{
				Description: "Optional path to a local credentials file used as a fallback when API key arguments are not set. Defaults to `~/.ace/config.json`. Most users should leave this unset and configure the API key directly.",
				Optional:    true,
			},
		},
	}
}

// ConfigValidators returns the plan-time validators that enforce the
// paired-credential rules for the provider block. Catching these at plan
// time means the user sees the misconfiguration before any apply runs.
func (p *AceCloudProvider) ConfigValidators(_ context.Context) []provider.ConfigValidator {
	return []provider.ConfigValidator{
		providervalidator.RequiredTogether(
			path.MatchRoot("api_key_id"),
			path.MatchRoot("api_key_secret"),
			path.MatchRoot("api_key_service_name"),
		),
	}
}

func (p *AceCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config AceCloudProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve values from config or environment variables.
	apiURL := resolveString(config.APIURL, "ACECLOUD_API_URL")
	apiKeyID := resolveString(config.APIKeyID, "ACECLOUD_API_KEY_ID")
	apiKeySecret := resolveString(config.APIKeySecret, "ACECLOUD_API_KEY_SECRET")
	apiKeyServiceName := resolveString(config.APIKeyServiceName, "ACECLOUD_API_KEY_SERVICE_NAME")
	region := resolveString(config.Region, "ACECLOUD_REGION")
	projectID := resolveString(config.ProjectID, "ACECLOUD_PROJECT_ID")
	aceConfigPath := resolveString(config.ACEConfigPath, "ACECLOUD_CONFIG_PATH")

	// Paired-credential rules (api_key_*) are enforced at plan time via
	// ConfigValidators. When env vars are the source instead of HCL, re-check
	// here so a half-set env-var combination is still caught.
	if apiKeyID != "" || apiKeySecret != "" || apiKeyServiceName != "" {
		if apiKeyID == "" || apiKeySecret == "" || apiKeyServiceName == "" {
			resp.Diagnostics.AddError(
				"Incomplete API Key Credentials",
				"`api_key_id`, `api_key_secret`, and `api_key_service_name` must all be set "+
					"together. When using environment variables, set `ACECLOUD_API_KEY_ID`, "+
					"`ACECLOUD_API_KEY_SECRET`, and `ACECLOUD_API_KEY_SERVICE_NAME`. Create a "+
					"key with `ace api-key create --service-name <name>`.",
			)
			return
		}
	}

	// Run the auth resolution chain.
	authCfg := auth.AuthConfig{
		APIURL:            apiURL,
		APIKeyID:          apiKeyID,
		APIKeySecret:      apiKeySecret,
		APIKeyServiceName: apiKeyServiceName,
		Region:            region,
		ProjectID:         projectID,
		ACEConfigPath:     aceConfigPath,
	}

	result, err := auth.Resolve(ctx, authCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Authentication Failed",
			fmt.Sprintf("Unable to authenticate with Ace Cloud API: %s", err),
		)
		return
	}

	// api_url chain: explicit HCL/env var → ace-cli config file → built-in
	// production default. Region and project_id are Required at the schema
	// level, so by the time we reach here the framework guarantees both are
	// non-empty.
	if apiURL == "" {
		if cliAPIURL := auth.ReadCLIConfigAPIURL(aceConfigPath); cliAPIURL != "" {
			apiURL = cliAPIURL
		} else {
			apiURL = defaultAPIURL
		}
	}
	resolvedRegion := region
	resolvedProjectID := projectID

	// Log which auth method was used (visible with TF_LOG=DEBUG).
	tflog.Debug(ctx, "Ace Cloud authentication resolved", map[string]interface{}{
		"method": result.Method,
		"region": resolvedRegion,
	})

	var c *client.Client
	if result.APIKeyID != "" && result.APIKeySecret != "" {
		c = client.NewClientWithAPIKey(apiURL, result.APIKeyID, result.APIKeySecret, result.APIKeyServiceName, resolvedRegion, resolvedProjectID)
	} else {
		c = client.NewClient(apiURL, result.Token, resolvedRegion, resolvedProjectID)
	}

	if err := c.ValidateToken(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Authentication Validation Failed",
			fmt.Sprintf("The provided credentials could not be validated: %s\n\n"+
				"Ensure api_key_id and api_key_secret are correct, the key is not disabled or expired, "+
				"and api_key_service_name matches the value used at key creation.", err),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *AceCloudProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		instance.NewResource,
		volume.NewResource,
		vpc.NewResource,
		subnet.NewResource,
		security_group.NewResource,
		floating_ip.NewResource,
		key_pair.NewResource,
		router.NewResource,
		router_interface.NewResource,
		snapshot.NewResource,
		volume_attachment.NewResource,
		volume_backup.NewResource,
		load_balancer.NewResource,
		lb_listener.NewResource,
		lb_pool.NewResource,
		lb_pool_member.NewResource,
		lb_health_monitor.NewResource,
		floating_ip_association.NewResource,
		auto_scaling_template.NewResource,
		auto_scaling_deployment.NewResource,
		api_key.NewResource,
	}
}

func (p *AceCloudProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		flavors.NewDataSource,
		images.NewDataSource,
		vpcs.NewDataSource,
		security_groups.NewDataSource,
		routers.NewDataSource,
	}
}

// resolveString returns the config value if set, otherwise falls back to the env var.
func resolveString(val types.String, envVar string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envVar)
}
