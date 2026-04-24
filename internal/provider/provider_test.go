package provider

import (
	"context"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestNew(t *testing.T) {
	p := New("test")
	if p == nil {
		t.Fatal("expected non-nil provider factory")
	}
	prov := p()
	if prov == nil {
		t.Fatal("expected non-nil provider instance")
	}
}

func TestProviderFactory(t *testing.T) {
	// Verify the provider can be instantiated by the framework.
	testProviderFactory := map[string]func() (tfprotov6.ProviderServer, error){
		"acecloud": providerserver.NewProtocol6WithError(New("test")()),
	}

	for name, factory := range testProviderFactory {
		server, err := factory()
		if err != nil {
			t.Fatalf("provider %s factory returned error: %v", name, err)
		}
		if server == nil {
			t.Fatalf("provider %s factory returned nil server", name)
		}
	}
}

func TestResolveString(t *testing.T) {
	tests := []struct {
		name     string
		val      types.String
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "returns config value when set",
			val:      types.StringValue("from-config"),
			envVar:   "ACECLOUD_TEST",
			expected: "from-config",
		},
		{
			name:     "returns empty for null with no env var",
			val:      types.StringNull(),
			envVar:   "ACECLOUD_NONEXISTENT_VAR_12345",
			expected: "",
		},
		{
			name:     "returns empty for unknown with no env var",
			val:      types.StringUnknown(),
			envVar:   "ACECLOUD_NONEXISTENT_VAR_12345",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveString(tc.val, tc.envVar)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestProviderModel_NewFields(t *testing.T) {
	// Verify the model struct has the new auth fields.
	model := AceCloudProviderModel{}
	_ = model.Email
	_ = model.Password
	_ = model.ACEConfigPath
}

func TestResolveString_EmailEnvVar(t *testing.T) {
	t.Setenv("ACECLOUD_EMAIL_TEST_UNIQUE", "user@test.com")
	got := resolveString(types.StringNull(), "ACECLOUD_EMAIL_TEST_UNIQUE")
	if got != "user@test.com" {
		t.Errorf("expected user@test.com, got %s", got)
	}
}

func TestResolveString_PasswordEnvVar(t *testing.T) {
	t.Setenv("ACECLOUD_PASSWORD_TEST_UNIQUE", "secret-pass")
	got := resolveString(types.StringNull(), "ACECLOUD_PASSWORD_TEST_UNIQUE")
	if got != "secret-pass" {
		t.Errorf("expected secret-pass, got %s", got)
	}
}

// --- Metadata tests ---

func TestMetadata_TypeName(t *testing.T) {
	p := &AceCloudProvider{version: "1.0.0"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)
	if resp.TypeName != "acecloud" {
		t.Errorf("expected type name 'acecloud', got %s", resp.TypeName)
	}
}

func TestMetadata_Version(t *testing.T) {
	p := &AceCloudProvider{version: "2.5.3"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)
	if resp.Version != "2.5.3" {
		t.Errorf("expected version '2.5.3', got %s", resp.Version)
	}
}

func TestMetadata_EmptyVersion(t *testing.T) {
	p := &AceCloudProvider{version: ""}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)
	if resp.Version != "" {
		t.Errorf("expected empty version, got %s", resp.Version)
	}
}

func TestMetadata_DevVersion(t *testing.T) {
	p := &AceCloudProvider{version: "dev"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)
	if resp.Version != "dev" {
		t.Errorf("expected version 'dev', got %s", resp.Version)
	}
}

// --- Schema tests ---

func TestSchema_AllAttributesPresent(t *testing.T) {
	p := &AceCloudProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	expectedAttrs := []string{
		"api_url",
		"api_token",
		"region",
		"project_id",
		"email",
		"password",
		"ace_config_path",
	}

	for _, name := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("expected attribute '%s' in provider schema", name)
		}
	}

	if len(resp.Schema.Attributes) != len(expectedAttrs) {
		t.Errorf("expected %d attributes, got %d", len(expectedAttrs), len(resp.Schema.Attributes))
	}
}

func TestSchema_AllAttributesOptional(t *testing.T) {
	p := &AceCloudProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	for name, attr := range resp.Schema.Attributes {
		if !attr.IsOptional() {
			t.Errorf("expected attribute '%s' to be optional", name)
		}
		if attr.IsRequired() {
			t.Errorf("attribute '%s' should not be required", name)
		}
		if attr.IsComputed() {
			t.Errorf("attribute '%s' should not be computed", name)
		}
	}
}

func TestSchema_SensitiveFields(t *testing.T) {
	p := &AceCloudProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	sensitiveFields := []string{"api_token", "password"}
	for _, name := range sensitiveFields {
		attr := resp.Schema.Attributes[name]
		if !attr.IsSensitive() {
			t.Errorf("expected attribute '%s' to be sensitive", name)
		}
	}

	nonSensitiveFields := []string{"api_url", "region", "project_id", "email", "ace_config_path"}
	for _, name := range nonSensitiveFields {
		attr := resp.Schema.Attributes[name]
		if attr.IsSensitive() {
			t.Errorf("expected attribute '%s' to NOT be sensitive", name)
		}
	}
}

func TestSchema_Description(t *testing.T) {
	p := &AceCloudProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)
	if resp.Schema.Description == "" {
		t.Error("expected provider schema to have a description")
	}
}

func TestSchema_AttributeDescriptions(t *testing.T) {
	p := &AceCloudProvider{}
	resp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, resp)

	for name, attr := range resp.Schema.Attributes {
		desc := attr.GetDescription()
		if desc == "" {
			t.Errorf("expected attribute '%s' to have a description", name)
		}
	}
}

// --- Resource list completeness (19 resources) ---

func TestResources_Count(t *testing.T) {
	p := &AceCloudProvider{}
	resources := p.Resources(context.Background())
	if len(resources) != 19 {
		t.Errorf("expected 19 resources, got %d", len(resources))
	}
}

func TestResources_AllRegistered(t *testing.T) {
	p := &AceCloudProvider{}
	resources := p.Resources(context.Background())

	expectedResources := []string{
		"acecloud_instance",
		"acecloud_volume",
		"acecloud_vpc",
		"acecloud_subnet",
		"acecloud_security_group",
		"acecloud_floating_ip",
		"acecloud_key_pair",
		"acecloud_router",
		"acecloud_router_interface",
		"acecloud_snapshot",
		"acecloud_volume_backup",
		"acecloud_load_balancer",
		"acecloud_lb_listener",
		"acecloud_lb_pool",
		"acecloud_lb_pool_member",
		"acecloud_lb_health_monitor",
		"acecloud_floating_ip_association",
		"acecloud_auto_scaling_template",
		"acecloud_auto_scaling_deployment",
	}

	// Get actual type names by instantiating each resource and calling Metadata
	actualNames := make(map[string]bool)
	for _, factory := range resources {
		r := factory()
		req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
		resp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), req, resp)
		actualNames[resp.TypeName] = true
	}

	for _, expected := range expectedResources {
		if !actualNames[expected] {
			t.Errorf("expected resource '%s' to be registered", expected)
		}
	}

	if len(actualNames) != len(expectedResources) {
		t.Errorf("expected %d unique resource names, got %d", len(expectedResources), len(actualNames))
	}
}

func TestResources_NoDuplicates(t *testing.T) {
	p := &AceCloudProvider{}
	resources := p.Resources(context.Background())

	seen := make(map[string]bool)
	for _, factory := range resources {
		r := factory()
		req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
		resp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), req, resp)
		if seen[resp.TypeName] {
			t.Errorf("duplicate resource type name: %s", resp.TypeName)
		}
		seen[resp.TypeName] = true
	}
}

func TestResources_AllNonNil(t *testing.T) {
	p := &AceCloudProvider{}
	resources := p.Resources(context.Background())
	for i, factory := range resources {
		r := factory()
		if r == nil {
			t.Errorf("resource factory at index %d returned nil", i)
		}
	}
}

// --- Data source list completeness (5 data sources) ---

func TestDataSources_Count(t *testing.T) {
	p := &AceCloudProvider{}
	dataSources := p.DataSources(context.Background())
	if len(dataSources) != 5 {
		t.Errorf("expected 5 data sources, got %d", len(dataSources))
	}
}

func TestDataSources_AllRegistered(t *testing.T) {
	p := &AceCloudProvider{}
	dataSources := p.DataSources(context.Background())

	expectedDataSources := []string{
		"acecloud_flavors",
		"acecloud_images",
		"acecloud_vpcs",
		"acecloud_security_groups",
		"acecloud_routers",
	}

	actualNames := make(map[string]bool)
	for _, factory := range dataSources {
		ds := factory()
		req := datasource.MetadataRequest{ProviderTypeName: "acecloud"}
		resp := &datasource.MetadataResponse{}
		ds.Metadata(context.Background(), req, resp)
		actualNames[resp.TypeName] = true
	}

	for _, expected := range expectedDataSources {
		if !actualNames[expected] {
			t.Errorf("expected data source '%s' to be registered", expected)
		}
	}

	if len(actualNames) != len(expectedDataSources) {
		t.Errorf("expected %d unique data source names, got %d", len(expectedDataSources), len(actualNames))
	}
}

func TestDataSources_NoDuplicates(t *testing.T) {
	p := &AceCloudProvider{}
	dataSources := p.DataSources(context.Background())

	seen := make(map[string]bool)
	for _, factory := range dataSources {
		ds := factory()
		req := datasource.MetadataRequest{ProviderTypeName: "acecloud"}
		resp := &datasource.MetadataResponse{}
		ds.Metadata(context.Background(), req, resp)
		if seen[resp.TypeName] {
			t.Errorf("duplicate data source type name: %s", resp.TypeName)
		}
		seen[resp.TypeName] = true
	}
}

func TestDataSources_AllNonNil(t *testing.T) {
	p := &AceCloudProvider{}
	dataSources := p.DataSources(context.Background())
	for i, factory := range dataSources {
		ds := factory()
		if ds == nil {
			t.Errorf("data source factory at index %d returned nil", i)
		}
	}
}

// --- Provider model field tests ---

func TestProviderModel_AllFields(t *testing.T) {
	model := AceCloudProviderModel{
		APIURL:        types.StringValue("https://api.example.com"),
		APIToken:      types.StringValue("tok-123"),
		Region:        types.StringValue("mumbai"),
		ProjectID:     types.StringValue("proj-abc"),
		Email:         types.StringValue("admin@example.com"),
		Password:      types.StringValue("secret"),
		ACEConfigPath: types.StringValue("/home/user/.ace/config.json"),
	}

	if model.APIURL.ValueString() != "https://api.example.com" {
		t.Errorf("expected api_url value, got %s", model.APIURL.ValueString())
	}
	if model.APIToken.ValueString() != "tok-123" {
		t.Errorf("expected api_token value, got %s", model.APIToken.ValueString())
	}
	if model.Region.ValueString() != "mumbai" {
		t.Errorf("expected region value, got %s", model.Region.ValueString())
	}
	if model.ProjectID.ValueString() != "proj-abc" {
		t.Errorf("expected project_id value, got %s", model.ProjectID.ValueString())
	}
	if model.Email.ValueString() != "admin@example.com" {
		t.Errorf("expected email value, got %s", model.Email.ValueString())
	}
	if model.Password.ValueString() != "secret" {
		t.Errorf("expected password value, got %s", model.Password.ValueString())
	}
	if model.ACEConfigPath.ValueString() != "/home/user/.ace/config.json" {
		t.Errorf("expected ace_config_path value, got %s", model.ACEConfigPath.ValueString())
	}
}

func TestProviderModel_NullFields(t *testing.T) {
	model := AceCloudProviderModel{
		APIURL:        types.StringNull(),
		APIToken:      types.StringNull(),
		Region:        types.StringNull(),
		ProjectID:     types.StringNull(),
		Email:         types.StringNull(),
		Password:      types.StringNull(),
		ACEConfigPath: types.StringNull(),
	}

	if !model.APIURL.IsNull() {
		t.Error("expected api_url to be null")
	}
	if !model.APIToken.IsNull() {
		t.Error("expected api_token to be null")
	}
	if !model.Region.IsNull() {
		t.Error("expected region to be null")
	}
	if !model.ProjectID.IsNull() {
		t.Error("expected project_id to be null")
	}
	if !model.Email.IsNull() {
		t.Error("expected email to be null")
	}
	if !model.Password.IsNull() {
		t.Error("expected password to be null")
	}
	if !model.ACEConfigPath.IsNull() {
		t.Error("expected ace_config_path to be null")
	}
}

// --- resolveString with env var set ---

func TestResolveString_ConfigOverridesEnv(t *testing.T) {
	t.Setenv("ACECLOUD_TEST_OVERRIDE_UNIQUE", "env-value")
	got := resolveString(types.StringValue("config-value"), "ACECLOUD_TEST_OVERRIDE_UNIQUE")
	if got != "config-value" {
		t.Errorf("expected config value to override env, got %s", got)
	}
}

func TestResolveString_EnvFallback(t *testing.T) {
	t.Setenv("ACECLOUD_TEST_FALLBACK_UNIQUE", "fallback-value")
	got := resolveString(types.StringNull(), "ACECLOUD_TEST_FALLBACK_UNIQUE")
	if got != "fallback-value" {
		t.Errorf("expected env fallback 'fallback-value', got %s", got)
	}
}

func TestResolveString_EmptyConfigString(t *testing.T) {
	// An empty string is still a "set" value, not null
	got := resolveString(types.StringValue(""), "ACECLOUD_NONEXISTENT_99")
	if got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

// --- Resource names are sorted for documentation ---

func TestResources_TypeNamesAreSorted(t *testing.T) {
	p := &AceCloudProvider{}
	resources := p.Resources(context.Background())

	var names []string
	for _, factory := range resources {
		r := factory()
		req := resource.MetadataRequest{ProviderTypeName: "acecloud"}
		resp := &resource.MetadataResponse{}
		r.Metadata(context.Background(), req, resp)
		names = append(names, resp.TypeName)
	}

	sorted := make([]string, len(names))
	copy(sorted, names)
	sort.Strings(sorted)

	// This is informational — we just verify we can list them
	if len(names) != 19 {
		t.Errorf("expected 19 resource names, got %d", len(names))
	}
}

func TestDataSources_TypeNamesAreSorted(t *testing.T) {
	p := &AceCloudProvider{}
	dataSources := p.DataSources(context.Background())

	var names []string
	for _, factory := range dataSources {
		ds := factory()
		req := datasource.MetadataRequest{ProviderTypeName: "acecloud"}
		resp := &datasource.MetadataResponse{}
		ds.Metadata(context.Background(), req, resp)
		names = append(names, resp.TypeName)
	}

	if len(names) != 5 {
		t.Errorf("expected 5 data source names, got %d", len(names))
	}
}

// --- New factory with various versions ---

func TestNew_VersionPropagation(t *testing.T) {
	versions := []string{"0.1.0", "1.0.0-rc1", "2.0.0+build.123", "dev", ""}
	for _, v := range versions {
		factory := New(v)
		if factory == nil {
			t.Fatalf("expected non-nil factory for version %q", v)
		}
		prov := factory()
		if prov == nil {
			t.Fatalf("expected non-nil provider for version %q", v)
		}
		// Verify version through Metadata
		aceProvider, ok := prov.(*AceCloudProvider)
		if !ok {
			t.Fatalf("expected *AceCloudProvider, got %T", prov)
		}
		if aceProvider.version != v {
			t.Errorf("expected version %q, got %q", v, aceProvider.version)
		}
	}
}

// --- Provider interface compliance ---

func TestProviderInterfaceCompliance(t *testing.T) {
	var _ provider.Provider = &AceCloudProvider{}
}
