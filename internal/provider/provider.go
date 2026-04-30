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
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/volume_backup"
	"github.com/akshay-malik-rtds/terraform-provider-acecloud/internal/resources/vpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure AceCloudProvider satisfies the provider interface.
var _ provider.Provider = &AceCloudProvider{}

// AceCloudProvider implements the Ace Cloud Terraform provider.
type AceCloudProvider struct {
	version string
}

// AceCloudProviderModel maps provider schema to Go types.
type AceCloudProviderModel struct {
	APIURL        types.String `tfsdk:"api_url"`
	APIToken      types.String `tfsdk:"api_token"`
	Region        types.String `tfsdk:"region"`
	ProjectID     types.String `tfsdk:"project_id"`
	Email         types.String `tfsdk:"email"`
	Password      types.String `tfsdk:"password"`
	ACEConfigPath types.String `tfsdk:"ace_config_path"`
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
			"Authentication methods (tried in order):\n" +
			"1. Static token via api_token or ACECLOUD_API_TOKEN\n" +
			"2. Email/password login via email + password or ACECLOUD_EMAIL + ACECLOUD_PASSWORD\n" +
			"3. ace-cli config file at ~/.ace/config.json (created by 'ace auth login')",
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Description: "Base URL of the Ace Cloud API (npc-api). Can also be set via ACECLOUD_API_URL environment variable. Falls back to api_base_url from ace-cli config.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "JWT bearer token for Ace Cloud API authentication. Can also be set via ACECLOUD_API_TOKEN environment variable. This is the highest-priority authentication method.",
				Optional:    true,
				Sensitive:   true,
			},
			"region": schema.StringAttribute{
				Description: "Cloud region (e.g. mumbai, noida, atlanta, delhi). Can also be set via ACECLOUD_REGION environment variable. Auto-selected from account if using email/password login.",
				Optional:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "Cloud project UUID. Can also be set via ACECLOUD_PROJECT_ID environment variable. Auto-selected from account if using email/password login.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address for Ace Cloud login authentication. Used with 'password' to obtain a token via POST /auth/login. Can also be set via ACECLOUD_EMAIL environment variable. Not supported with 2FA-enabled accounts.",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Password for Ace Cloud login authentication. Used with 'email' to obtain a token. Can also be set via ACECLOUD_PASSWORD environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"ace_config_path": schema.StringAttribute{
				Description: "Path to the ace-cli configuration file. Defaults to ~/.ace/config.json. The provider reads token, region, and project_id from this file as a fallback authentication method.",
				Optional:    true,
			},
		},
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
	apiToken := resolveString(config.APIToken, "ACECLOUD_API_TOKEN")
	region := resolveString(config.Region, "ACECLOUD_REGION")
	projectID := resolveString(config.ProjectID, "ACECLOUD_PROJECT_ID")
	email := resolveString(config.Email, "ACECLOUD_EMAIL")
	password := resolveString(config.Password, "ACECLOUD_PASSWORD")
	aceConfigPath := resolveString(config.ACEConfigPath, "ACECLOUD_CONFIG_PATH")

	// Validate: email and password must be provided together.
	if (email != "" && password == "") || (email == "" && password != "") {
		resp.Diagnostics.AddError(
			"Incomplete Login Credentials",
			"Both 'email' and 'password' must be provided together for login authentication. "+
				"Set both in the provider configuration or via ACECLOUD_EMAIL and ACECLOUD_PASSWORD "+
				"environment variables.",
		)
		return
	}

	// Run the auth resolution chain.
	authCfg := auth.AuthConfig{
		APIURL:        apiURL,
		APIToken:      apiToken,
		Region:        region,
		ProjectID:     projectID,
		Email:         email,
		Password:      password,
		ACEConfigPath: aceConfigPath,
	}

	result, err := auth.Resolve(ctx, authCfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Authentication Failed",
			fmt.Sprintf("Unable to authenticate with Ace Cloud API: %s", err),
		)
		return
	}

	// If api_url was not set explicitly, try to get it from CLI config.
	if apiURL == "" {
		cliAPIURL := auth.ReadCLIConfigAPIURL(aceConfigPath)
		if cliAPIURL != "" {
			apiURL = cliAPIURL
		}
	}

	// Use auth result for region and project_id (may have been auto-selected).
	resolvedRegion := result.Region
	resolvedProjectID := result.ProjectID

	// Final validation: required fields.
	if apiURL == "" {
		resp.Diagnostics.AddError(
			"Missing API URL",
			"The Ace Cloud API URL must be set in the provider configuration, "+
				"via the ACECLOUD_API_URL environment variable, or in the ace-cli config file (~/.ace/config.json).",
		)
	}
	if resolvedRegion == "" {
		resp.Diagnostics.AddError(
			"Missing Region",
			"The cloud region must be set in the provider configuration, "+
				"via the ACECLOUD_REGION environment variable, or in the ace-cli config file (~/.ace/config.json).",
		)
	}
	if resolvedProjectID == "" {
		resp.Diagnostics.AddError(
			"Missing Project ID",
			"The project ID must be set in the provider configuration, "+
				"via the ACECLOUD_PROJECT_ID environment variable, or in the ace-cli config file (~/.ace/config.json).",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Emit warnings if project/region were auto-selected during login.
	if result.Method == "email_password" {
		if region == "" && resolvedRegion != "" {
			resp.Diagnostics.AddWarning(
				"Region Auto-Selected",
				fmt.Sprintf("No region was explicitly configured. Auto-selected '%s' from your Ace Cloud account. "+
					"Set 'region' in the provider block or ACECLOUD_REGION env var to avoid this.", resolvedRegion),
			)
		}
		if projectID == "" && resolvedProjectID != "" {
			resp.Diagnostics.AddWarning(
				"Project ID Auto-Selected",
				fmt.Sprintf("No project_id was explicitly configured. Auto-selected '%s' from your Ace Cloud account. "+
					"Set 'project_id' in the provider block or ACECLOUD_PROJECT_ID env var to avoid this.", resolvedProjectID),
			)
		}
	}

	// Log which auth method was used (visible with TF_LOG=DEBUG).
	tflog.Debug(ctx, "Ace Cloud authentication resolved", map[string]interface{}{
		"method": result.Method,
		"region": resolvedRegion,
	})

	c := client.NewClient(apiURL, result.Token, resolvedRegion, resolvedProjectID)

	// Validate token for static_token and cli_config methods.
	// Email/password already validated during login.
	if result.Method != "email_password" {
		if err := c.ValidateToken(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Token Validation Failed",
				fmt.Sprintf("The provided token could not be validated (%s method): %s\n\n"+
					"If using a static token, ensure it has not expired.\n"+
					"If using ace-cli config, try re-authenticating: ace auth login", result.Method, err),
			)
			return
		}
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
		volume_backup.NewResource,
		load_balancer.NewResource,
		lb_listener.NewResource,
		lb_pool.NewResource,
		lb_pool_member.NewResource,
		lb_health_monitor.NewResource,
		floating_ip_association.NewResource,
		auto_scaling_template.NewResource,
		auto_scaling_deployment.NewResource,
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
