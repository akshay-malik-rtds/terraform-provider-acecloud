# Upgrading the AceCloud Terraform Provider

## Upgrading to v0.2.0

v0.2.0 contains two intentional breaking changes to the provider configuration model. Together they make `terraform plan` deterministic and narrow the auth surface to the credential type that's actually appropriate for infrastructure-as-code.

### Breaking change 1: `region` and `project_id` are now Required

In v0.1.x, `region` and `project_id` were optional. When omitted, the provider would either auto-select an arbitrary active project/region from your account (email + password auth) or read them from `~/.ace/config.json` (`ace-cli` auth). Both fallbacks were non-deterministic across machines and across accounts.

In v0.2.0, both arguments are **Required**. Terraform will fail at plan time if either is missing:

```text
Error: Missing required argument

  on main.tf line 3, in provider "acecloud":
   3: provider "acecloud" {

The argument "region" is required, but no definition was found.
```

#### Migration

Set both arguments explicitly in every provider block, either as literals, as variables, or via environment variables.

**Before (v0.1.x):**

```hcl
provider "acecloud" {
  api_key_id           = var.api_key_id
  api_key_secret       = var.api_key_secret
  api_key_service_name = "terraform"
  # region / project_id omitted — relied on auto-select or ace-cli config
}
```

**After (v0.2.0):**

```hcl
provider "acecloud" {
  api_key_id           = var.api_key_id
  api_key_secret       = var.api_key_secret
  api_key_service_name = "terraform"
  region               = var.acecloud_region      # e.g. "ap-south-noi-1"
  project_id           = var.acecloud_project_id
}
```

Or set them via environment variables:

```sh
export ACECLOUD_REGION="ap-south-noi-1"
export ACECLOUD_PROJECT_ID="<your-project-uuid>"
```

To find your project UUID and the regions it's authorized for, run `ace project list` (after `ace auth login-api-key`) or use the AceCloud web console.

### Breaking change 2: `api_token`, `email`, and `password` are removed

In v0.1.x, the provider supported four authentication methods: API key, static JWT token, email + password, and `ace-cli` config. v0.2.0 removes the JWT and email/password methods.

The following provider arguments and environment variables are no longer recognised:

| Removed argument | Removed environment variable |
|---|---|
| `api_token` | `ACECLOUD_API_TOKEN` |
| `email` | `ACECLOUD_EMAIL` |
| `password` | `ACECLOUD_PASSWORD` |

If your configuration references any of these, Terraform will fail to parse the provider block:

```text
Error: Unsupported argument

  on main.tf line 5, in provider "acecloud":
   5:   api_token = var.acecloud_api_token

An argument named "api_token" is not expected here.
```

#### Migration

Switch to an API key. Create one with the CLI:

```sh
ace api-key create --service-name terraform
```

Then update your provider block:

**Before (v0.1.x):**

```hcl
provider "acecloud" {
  api_url    = var.acecloud_api_url
  api_token  = var.acecloud_api_token   # OR email + password
  region     = var.acecloud_region
  project_id = var.acecloud_project_id
}
```

**After (v0.2.0):**

```hcl
provider "acecloud" {
  api_url              = var.acecloud_api_url
  api_key_id           = var.acecloud_api_key_id
  api_key_secret       = var.acecloud_api_key_secret
  api_key_service_name = "terraform"
  region               = var.acecloud_region
  project_id           = var.acecloud_project_id
}
```

If you were authenticating through `ace-cli` and your CLI session was using a legacy JWT, re-authenticate using an API key so the CLI config supplies the right credential type:

```sh
ace api-key create --service-name terraform
ace auth login-api-key --service-name terraform
```

### What stays the same

- `api_url` remains optional. When unset (and `ACECLOUD_API_URL` is not exported), the provider continues to read `api_base_url` from `~/.ace/config.json`. This keeps the zero-config local development path working when you're already authenticated through the CLI.
- The `ace-cli` config fallback still works for credentials — it just supplies an API key (or a legacy JWT, if that's what your CLI session has) instead of being a peer of the removed login methods.
- All resource and data-source schemas are unchanged. No state migration is required.

### Why these changes

**On region/project being Required:** the provider previously had three code paths for resolving `region` and `project_id`. Two of them (auto-select from login response, read from CLI config) were non-deterministic across machines and across CLI sessions. For a tool whose contract is "plan output is the source of truth," that's the wrong default. v0.2.0 makes the deterministic path the only path.

**On removing JWT and email/password auth:** these auth primitives don't fit infrastructure-as-code:

- **Static JWT tokens** expire in 24 hours. Useless for unattended automation; offers no advantage over API keys for interactive use.
- **Email and password** transmits raw user credentials on every Terraform run, can't handle 2FA accounts, and ties Terraform state lifetime to a personal user account. If the user leaves the company, the state becomes unmanageable.
- **API keys** are scoped, rotatable, revocable, and tied to a service identity rather than a person. That's the right primitive for Terraform.
