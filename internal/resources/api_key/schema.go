// Package api_key implements the acecloud_api_key resource for managing
// long-lived programmatic credentials via the AceCloud IAM API.
package api_key

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// apiKeyResourceModel maps the resource schema to Go types.
type apiKeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ServiceName types.String `tfsdk:"service_name"`
	Description types.String `tfsdk:"description"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Secret      types.String `tfsdk:"secret"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func resourceSchema() schema.Schema {
	return schema.Schema{
		Description: "Manages an AceCloud IAM API key — a long-lived programmatic credential for automation. The secret is returned only on create; save it immediately.\n\n**Important:** the backend rejects creation and revival of API keys when the caller is itself authenticated with an API key (a security-by-design constraint that prevents key minting from a stolen key). Since v0.2.0 of this provider only accepts API-key authentication, `terraform apply` against `acecloud_api_key` will fail at create time. Create the initial key from the AceCloud console, then `terraform import` it if you want to manage its metadata, updates, or deletion through Terraform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the API key. Combine with `secret` to form the full credential (`<id>.<secret>`).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_name": schema.StringAttribute{
				Description: "Human-readable identifier for the service or automation that will use this key (e.g. `terraform-prod`, `ci-pipeline`). Surfaced in audit logs. Minimum 2 characters.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
					stringvalidator.LengthAtMost(255),
				},
			},
			"description": schema.StringAttribute{
				Description: "Free-form description of the key's purpose. Minimum 3 characters when set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(1024),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the key can be used for authentication. Defaults to `true`. Set to `false` to disable the key without deleting it; flip back to `true` to re-enable.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"secret": schema.StringAttribute{
				Description: "The API key secret. Returned only when the key is created and stored in Terraform state thereafter. Save it to a secret manager immediately — if state is lost, regenerate with `ace api-key revive <id>`. Always referenced as `sensitive`.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expires_at": schema.StringAttribute{
				Description: "RFC3339 timestamp when the key expires (auto-set by the platform).",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "RFC3339 timestamp of key creation.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Description: "RFC3339 timestamp of the last update to this key.",
				Computed:    true,
			},
		},
	}
}
