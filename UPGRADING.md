# Upgrading the AceCloud Terraform Provider

## Upgrading to v0.2.4

Bug-fix release for two cosmetic issues from v0.2.3. No state migration required.

- `acecloud_api_key` is now correctly grouped under the **IAM** subcategory in the Registry sidebar (was bucketed under "Resources" in v0.2.3 because of a YAML frontmatter parsing bug).
- GitHub Releases now publish the human-readable summary in the release body (v0.2.3 shipped with an empty body).

## Upgrading to v0.2.3

Documentation and packaging release. No state migration required.

- Provider page H1 on the Registry now reads **AceCloud Provider** (was lowercase `acecloud Provider`). Sidebar entries are unchanged.
- Multi-region and multi-project setup content moved out of the provider overview into a dedicated **Guides** page.
- GitHub Releases now publish a human-readable summary in the release body.

## Upgrading to v0.2.2

v0.2.2 is a documentation-only release with no code or behavior changes. Resource and provider descriptions were rewritten to remove internal implementation details and present only what users need to operate the resource. No state migration is required.

## Upgrading to v0.2.1

v0.2.1 is incremental over v0.2.0:

- **Removed:** the `locked` attribute on `acecloud_instance`. See "Breaking change 3" below for the migration.
- **Changed:** `api_url` now defaults to `https://customer.acecloud.ai/api/v1/`. You can remove it from your provider block if you were pointing at production.
- **Clarified:** new `acecloud_api_key` resources must be created from the AceCloud console. Use `terraform import` to bring an existing key under Terraform management.

If you are upgrading from v0.1.x, all three v0.2.0 breaking changes below also apply.

## Upgrading to v0.2.0

v0.2.0 contains three intentional breaking changes to the provider configuration and resource schemas.

### Breaking change 1: `region` and `project_id` are now Required

In v0.1.x, `region` and `project_id` were optional and could be inferred at runtime. v0.2.0 makes both **Required** so that `terraform plan` produces the same output on every machine. Terraform will fail at plan time if either is missing:

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
  # region / project_id omitted — relied on inferred defaults
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

Find your project UUID and authorized regions in the AceCloud web console.

### Breaking change 2: `api_token`, `email`, and `password` are removed

v0.2.0 supports a single authentication method: API key. The following provider arguments and environment variables are no longer recognised:

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

Switch to an API key created from the AceCloud console, then update your provider block:

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

### Breaking change 3: `locked` attribute removed from `acecloud_instance`

The `locked` attribute on `acecloud_instance` has been removed. Lock state is no longer managed by Terraform.

If your configuration references `locked = true` or `locked = false`, Terraform will fail to parse the resource block:

```text
Error: Unsupported argument

  on main.tf line 8, in resource "acecloud_instance" "web":
   8:   locked = true

An argument named "locked" is not expected here.
```

#### Migration

Remove the `locked` argument from your HCL. If you need to lock or unlock an instance, do it from the AceCloud console.

`terraform destroy` continues to work on locked instances — the provider unlocks them automatically before deletion.

### What stays the same

- `api_url` remains optional. It now defaults to `https://customer.acecloud.ai/api/v1/`. Override only when targeting a non-production endpoint.
- `power_state`, `flavor_id`, `security_group_ids`, and all other `acecloud_instance` attributes are unchanged.
- All other resource and data-source schemas are unchanged. No state migration is required.
