---
page_title: "Multi-region and multi-project setups"
subcategory: "Guides"
description: |-
  Patterns for managing AceCloud infrastructure across multiple regions and projects using Terraform provider aliases.
---

# Multi-region and multi-project setups

For configurations that span more than one region or more than one project, use **Terraform provider aliases**.

> **About credentials across regions/projects:** AceCloud API keys can be scoped to a specific service, account, or project. Whether a single key works across the regions or projects you target depends on how the key was issued. The examples below cover both scenarios — **same credentials reused** and **separate credentials per provider alias**.

## Pattern 1: one configuration, two regions, **same credentials**

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

## Pattern 2: one configuration, two regions, **separate credentials per region**

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

Per-alias keys can also represent different scopes within the same region — for example, one key for production workloads and a separate read-only key for an audit account:

```hcl
provider "acecloud" {
  # Primary: full-access key for the workloads team.
  api_key_id           = var.primary_api_key_id
  api_key_secret       = var.primary_api_key_secret
  api_key_service_name = "terraform"
  region               = "ap-south-noi-1"
  project_id           = var.primary_project_id
}

provider "acecloud" {
  alias                = "audit"
  # Audit: read-only key bound to a separate audit project / service identity.
  api_key_id           = var.audit_api_key_id
  api_key_secret       = var.audit_api_key_secret
  api_key_service_name = "terraform-audit"
  region               = "ap-south-noi-1"
  project_id           = var.audit_project_id
}
```

## Pattern 3: one configuration, two projects (same region)

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

## Best practices for managing multi-region infrastructure

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

## What about per-resource region or project_id attributes?

This provider does **not** expose `region` or `project_id` as resource-level attributes. Use provider aliases (above) for multi-region or multi-project configurations. The reasons:

- Provider aliases are a standard Terraform feature that works with every IDE, CI tool, and policy engine.
- Per-resource region would bloat every schema by two attributes and require threading the override into every API call, with no functional gain over aliases.
