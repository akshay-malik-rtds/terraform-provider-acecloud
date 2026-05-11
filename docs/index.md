---
page_title: "acecloud Provider"
description: |-
  Terraform provider for Ace Cloud infrastructure. Manages compute instances, volumes, snapshots, backups, VPCs, subnets, routers, security groups, floating IPs, key pairs, and load balancers via the Ace Cloud API.

Authentication methods (tried in order):
1. API key (api_key_id + api_key_secret, or ACECLOUD_API_KEY_ID + ACECLOUD_API_KEY_SECRET) — recommended for automation
2. Static token via api_token or ACECLOUD_API_TOKEN
3. Email/password login via email + password or ACECLOUD_EMAIL + ACECLOUD_PASSWORD
4. ace-cli config file at ~/.ace/config.json (created by 'ace auth login')
---

# acecloud Provider

The AceCloud Terraform Provider lets you manage AceCloud infrastructure as code — compute instances, block storage, networking, load balancing, auto scaling, and IAM API keys.

## Example Usage

```terraform
terraform {
  required_providers {
    acecloud = {
      source  = "registry.terraform.io/akshay-malik-rtds/acecloud"
      version = "~> 0.1"
    }
  }
  required_version = ">= 1.6"
}

# Authenticate with an API key (recommended for automation).
# Create one with: ace api-key create --service-name terraform-prod
provider "acecloud" {
  api_url              = var.acecloud_api_url
  api_key_id           = var.acecloud_api_key_id
  api_key_secret       = var.acecloud_api_key_secret
  api_key_service_name = "terraform"
  region               = var.acecloud_region
  project_id           = var.acecloud_project_id
}

# Alternative: authenticate with a JWT bearer token.
# Tokens expire after 24 hours; for long-running automation prefer api_key auth.
#
# provider "acecloud" {
#   api_url    = var.acecloud_api_url
#   api_token  = var.acecloud_api_token
#   region     = var.acecloud_region
#   project_id = var.acecloud_project_id
# }

variable "acecloud_api_url" {
  description = "Ace Cloud API base URL"
  type        = string
}

variable "acecloud_api_key_id" {
  description = "Ace Cloud API key identifier"
  type        = string
  sensitive   = true
}

variable "acecloud_api_key_secret" {
  description = "Ace Cloud API key secret"
  type        = string
  sensitive   = true
}

variable "acecloud_api_token" {
  description = "Ace Cloud API JWT token (alternative to API key)"
  type        = string
  default     = ""
  sensitive   = true
}

variable "acecloud_region" {
  description = "Cloud region"
  type        = string
  default     = "mumbai"
}

variable "acecloud_project_id" {
  description = "AceCloud project UUID"
  type        = string
}
```

## Authentication

The provider supports four authentication methods, tried in this order:

1. **API key** (recommended for automation) — long-lived credentials suitable for CI/CD pipelines. Set `api_key_id`, `api_key_secret`, and `api_key_service_name` in the provider block, or use the environment variables `ACECLOUD_API_KEY_ID`, `ACECLOUD_API_KEY_SECRET`, and `ACECLOUD_API_KEY_SERVICE_NAME`. Create a key with `ace api-key create --service-name <name>`.

2. **Static JWT token** — short-lived (24 hour) bearer token. Set `api_token` or `ACECLOUD_API_TOKEN`. Useful for interactive sessions; not recommended for unattended automation.

3. **Email and password** — programmatic login via `POST /auth/login`. Set `email` and `password`, or `ACECLOUD_EMAIL` and `ACECLOUD_PASSWORD`. Not supported for accounts with 2FA enabled.

4. **`ace-cli` config** (zero configuration) — reads `~/.ace/config.json`, the file written by `ace auth login`. Useful for local development when you're already authenticated through the CLI.

If multiple methods are configured, the highest-priority method wins.

## Region and Project ID

The `region` and `project_id` arguments are required for every API call. Set them explicitly in the provider block or via the `ACECLOUD_REGION` / `ACECLOUD_PROJECT_ID` environment variables.

### What happens if you omit them

Behavior depends on which authentication method is active:

| Auth method | If `region` / `project_id` unset | Recommended |
|---|---|---|
| API key | Provider fails to configure with `Missing Region` / `Missing Project ID` errors. | Always set both explicitly. |
| Static JWT token | Same — provider fails to configure. | Always set both explicitly. |
| Email + password | Provider auto-selects an **arbitrary active project** from your account and an **arbitrary region** from that project. A warning is emitted at plan time naming what was picked. | Set both explicitly for any non-interactive use; auto-selection is non-deterministic across runs. |
| `ace-cli` config | Reads `region` and `project_id` from `~/.ace/config.json` (written by `ace auth login` or `ace config set`). | Whatever the CLI was logged into. |

> **Recommendation:** always set `region` and `project_id` explicitly for production / non-interactive use. Auto-selection is a convenience for first-run exploration only and gives different results on different machines.

## Multi-region and multi-project setups

For configurations that span more than one region or more than one project, use **Terraform provider aliases**. This is the same pattern AWS, GCP, and other multi-region providers use.

> **About credentials across regions/projects:** AceCloud API keys can be scoped to a specific service / account / project. Whether a single key works across the regions or projects you target depends on how the key was issued. The examples below cover both scenarios — **same credentials reused** and **separate credentials per provider alias**.

### Pattern 1: one configuration, two regions, **same credentials**

Use this when your API key is account-scoped and authorized across all regions you target.

```hcl
locals {
  # Single set of credentials reused across regions.
  shared_auth = {
    api_key_id           = var.acecloud_api_key_id
    api_key_secret       = var.acecloud_api_key_secret
    api_key_service_name = "terraform"
  }
}

# Default provider — used by resources that don't specify a `provider =` argument.
provider "acecloud" {
  api_key_id           = local.shared_auth.api_key_id
  api_key_secret       = local.shared_auth.api_key_secret
  api_key_service_name = local.shared_auth.api_key_service_name
  region               = "ap-south-noi-1"
  project_id           = var.primary_project_id
}

# Aliased provider — same credentials, different region.
provider "acecloud" {
  alias                = "mumbai"
  api_key_id           = local.shared_auth.api_key_id
  api_key_secret       = local.shared_auth.api_key_secret
  api_key_service_name = local.shared_auth.api_key_service_name
  region               = "ap-south-mum-1"
  project_id           = var.primary_project_id
}

resource "acecloud_instance" "primary" {
  # Uses the default provider → noida
  name = "primary-app"
  # ... other attributes ...
}

resource "acecloud_instance" "dr" {
  provider = acecloud.mumbai
  name     = "dr-app"
  # ... other attributes ...
}
```

### Pattern 2: one configuration, two regions, **separate credentials per region**

Use this when each region requires its own API key (different account, different team-owned key, or region-scoped key). Every provider alias gets its own credential set; nothing is shared.

```hcl
# Noida — primary region, primary team's key.
provider "acecloud" {
  api_key_id           = var.noida_api_key_id
  api_key_secret       = var.noida_api_key_secret
  api_key_service_name = "terraform-primary"
  region               = "ap-south-noi-1"
  project_id           = var.noida_project_id
}

# Mumbai — DR region, separate team or separate account, different key.
provider "acecloud" {
  alias                = "mumbai"
  api_key_id           = var.mumbai_api_key_id
  api_key_secret       = var.mumbai_api_key_secret
  api_key_service_name = "terraform-dr"
  region               = "ap-south-mum-1"
  project_id           = var.mumbai_project_id
}

resource "acecloud_instance" "primary" {
  name = "primary-app"
  # ... other attributes ...
}

resource "acecloud_instance" "dr" {
  provider = acecloud.mumbai
  name     = "dr-app"
  # ... other attributes ...
}
```

Mixing auth methods per alias is also supported — for example, you might use an API key for the primary region and email/password for an account you only access interactively:

```hcl
provider "acecloud" {
  # Primary region: long-lived API key for unattended automation.
  api_key_id           = var.primary_api_key_id
  api_key_secret       = var.primary_api_key_secret
  api_key_service_name = "terraform"
  region               = "ap-south-noi-1"
  project_id           = var.primary_project_id
}

provider "acecloud" {
  alias      = "audit"
  # Audit account: read-only access via email/password, no automation key issued.
  email      = var.audit_email
  password   = var.audit_password
  region     = "ap-south-noi-1"
  project_id = var.audit_project_id
}
```

### Pattern 3: one configuration, two projects (same region)

Same idea, applied to projects within one region.

```hcl
provider "acecloud" {
  # ... auth for shared services project ...
  region     = "ap-south-noi-1"
  project_id = var.shared_services_project
}

provider "acecloud" {
  alias      = "apps"
  # ... auth that is authorized for the apps project (may be the same or different key) ...
  region     = "ap-south-noi-1"
  project_id = var.apps_project
}

resource "acecloud_vpc" "shared" {
  # Default provider → shared services project
  name              = "shared-vpc"
  subnet_name       = "shared-sub"
  subnet_cidr       = "10.0.0.0/24"
  subnet_ip_version = 4
}

resource "acecloud_instance" "app" {
  provider = acecloud.apps
  name     = "app-01"
  # ... other attributes ...
}
```

> **Securing per-alias credentials:** when each alias has its own credentials, never hard-code them. Use sensitive Terraform variables sourced from environment variables (`TF_VAR_noida_api_key_secret=…`), a CI/CD secret store, or a remote backend with `terraform_remote_state` pulling encrypted values. Avoid committing more than one credential set to the same `.tfvars` file even in encrypted form — segment per-region credentials into separate stores so a leak in one region doesn't expose the others.

### Best practices for managing multi-region infrastructure

For anything beyond two or three resources per region, structure the configuration to keep regional concerns separated:

1. **Terraform modules per region or per environment.** Encapsulate a region's resources in a module, then call the module twice with different providers passed in.

   ```hcl
   module "noida" {
     source     = "./modules/region"
     providers  = { acecloud = acecloud }
     vpc_cidr   = "10.10.0.0/16"
   }

   module "mumbai" {
     source     = "./modules/region"
     providers  = { acecloud = acecloud.mumbai }
     vpc_cidr   = "10.20.0.0/16"
   }
   ```

2. **Separate state files per region.** If your blast radius matters (it usually does for DR/primary pairs), use a separate Terraform configuration root per region with its own backend. This means a corrupted state file or accidental destroy in one region cannot affect the other. Combine with a shared module library if you want config parity. **Bonus benefit:** if each region uses a different API key, isolating per-region state also isolates the credential footprint — a leaked key only puts that region at risk.

3. **A single root module managing everything is fine for small footprints** (≤ 30 resources total, single team). Past that scale, the plan/apply latency and blast radius arguments tilt toward per-region roots.

4. **Pin variable names and tag every resource with its region.** Even if region is in the provider alias, having `tags = ["region:noida"]` or naming conventions like `${var.region_short}-app` makes runbooks easier and prevents cross-region copy-paste mistakes.

5. **Use a workspaces or directory-per-env layout for environment isolation.** Don't mix prod-noida with prod-mumbai with dev-noida in one state.

6. **Run plans against one region at a time when iterating.** With `terraform plan -target=module.noida` you can preview changes scoped to a single region during development, then drop the `-target` for the final apply.

7. **For >2 regions or frequent regional rollouts**, consider [Terragrunt](https://terragrunt.gruntwork.io/) or [Spacelift / env0 / Terraform Cloud workspaces](https://developer.hashicorp.com/terraform/cloud-docs/workspaces) to manage many small per-region states with shared inputs. The native `provider "acecloud" { alias = ... }` approach gets unwieldy past ~3 aliased providers.

### What about per-resource region or project_id attributes?

This provider does **not** expose `region` or `project_id` as resource-level attributes. Use provider aliases (above) for multi-region or multi-project configurations. The reasons:

- Provider aliases are a standard Terraform feature that works with every IDE, CI tool, and policy engine.
- Per-resource region would bloat every schema by two attributes and require threading the override into every API call, with no functional gain over aliases.
- Other mid-size IaaS providers (Hetzner, DigitalOcean) follow the same pattern.

## Environment variables

Every provider argument can be set via an environment variable instead of an HCL attribute. This is the recommended way to keep credentials out of version control.

| Variable | Provider attribute |
|---|---|
| `ACECLOUD_API_URL` | `api_url` |
| `ACECLOUD_API_KEY_ID` | `api_key_id` |
| `ACECLOUD_API_KEY_SECRET` | `api_key_secret` |
| `ACECLOUD_API_KEY_SERVICE_NAME` | `api_key_service_name` |
| `ACECLOUD_API_TOKEN` | `api_token` |
| `ACECLOUD_EMAIL` | `email` |
| `ACECLOUD_PASSWORD` | `password` |
| `ACECLOUD_REGION` | `region` |
| `ACECLOUD_PROJECT_ID` | `project_id` |

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `ace_config_path` (String) Path to the ace-cli configuration file. Defaults to ~/.ace/config.json. The provider reads token or API key, region, and project_id from this file as a fallback authentication method.
- `api_key_id` (String, Sensitive) API key identifier (the part before the dot in keyId.secret). Can also be set via ACECLOUD_API_KEY_ID environment variable. API keys are long-lived credentials suitable for automation; create one with 'ace api-key create'. Must be set together with api_key_secret. This is the highest-priority authentication method.
- `api_key_secret` (String, Sensitive) API key secret (the part after the dot in keyId.secret). Can also be set via ACECLOUD_API_KEY_SECRET environment variable. The secret is only shown once on key creation; if lost, regenerate via 'ace api-key revive <id>'.
- `api_key_service_name` (String) Service name attached to API key requests. **Required when using API key authentication** — must match the service name supplied at key creation time, otherwise the backend rejects the request. Can also be set via ACECLOUD_API_KEY_SERVICE_NAME environment variable. Has no effect when other auth methods are used.
- `api_token` (String, Sensitive) JWT bearer token for Ace Cloud API authentication. Can also be set via ACECLOUD_API_TOKEN environment variable. Short-lived (24h); for automation prefer api_key_id + api_key_secret instead.
- `api_url` (String) Base URL of the Ace Cloud API (npc-api). Can also be set via ACECLOUD_API_URL environment variable. Falls back to api_base_url from ace-cli config.
- `email` (String) Email address for Ace Cloud login authentication. Used with 'password' to obtain a token via POST /auth/login. Can also be set via ACECLOUD_EMAIL environment variable. Not supported with 2FA-enabled accounts.
- `password` (String, Sensitive) Password for Ace Cloud login authentication. Used with 'email' to obtain a token. Can also be set via ACECLOUD_PASSWORD environment variable.
- `project_id` (String) Cloud project UUID. Can also be set via ACECLOUD_PROJECT_ID environment variable. Auto-selected from account if using email/password login.
- `region` (String) Cloud region (e.g. mumbai, noida, atlanta, delhi). Can also be set via ACECLOUD_REGION environment variable. Auto-selected from account if using email/password login.
