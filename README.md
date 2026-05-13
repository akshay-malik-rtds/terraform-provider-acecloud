# Terraform Provider for AceCloud

[![Go Reference](https://pkg.go.dev/badge/github.com/akshay-malik-rtds/terraform-provider-acecloud.svg)](https://pkg.go.dev/github.com/akshay-malik-rtds/terraform-provider-acecloud)
[![Go Report Card](https://goreportcard.com/badge/github.com/akshay-malik-rtds/terraform-provider-acecloud)](https://goreportcard.com/report/github.com/akshay-malik-rtds/terraform-provider-acecloud)
[![License: MPL-2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

The AceCloud Terraform Provider lets you manage your [Ace Cloud](https://acecloud.ai) infrastructure as code using HashiCorp [Terraform](https://www.terraform.io).

## Features

- **Compute**: instances, key pairs
- **Block storage**: volumes, snapshots, volume backups
- **Networking**: VPCs, subnets, routers, router interfaces, security groups, floating IPs
- **Load balancing**: load balancers, listeners, pools, pool members, health monitors
- **Auto scaling**: templates, deployments
- **IAM**: API keys (long-lived programmatic credentials)

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.6
- An [AceCloud](https://acecloud.ai) account with API access

## Installation

The provider is published to the [Terraform Registry](https://registry.terraform.io/providers/akshay-malik-rtds/acecloud/latest). Add it to your Terraform configuration:

```hcl
terraform {
  required_providers {
    acecloud = {
      source  = "akshay-malik-rtds/acecloud"
      version = "~> 0.1"
    }
  }
}
```

Run `terraform init` and Terraform will download the provider automatically.

## Quick Start

```hcl
provider "acecloud" {
  # Credentials are read from ~/.ace/config.json by default.
  # See https://registry.terraform.io/providers/akshay-malik-rtds/acecloud/latest/docs
  # for other authentication methods.
}

resource "acecloud_vpc" "main" {
  name                  = "production-vpc"
  admin_state_up        = true
  port_security_enabled = true

  subnet_name        = "primary-subnet"
  subnet_cidr        = "10.0.0.0/24"
  subnet_ip_version  = 4
  subnet_enable_dhcp = true
}

resource "acecloud_security_group" "web" {
  name        = "web"
  description = "Allow HTTP and SSH"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 80
    port_range_max   = 80
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}
```

See the [examples](./examples) directory for more complete configurations.

## Documentation

Full documentation lives on the [Terraform Registry](https://registry.terraform.io/providers/akshay-malik-rtds/acecloud/latest/docs):

- [Getting Started](https://registry.terraform.io/providers/akshay-malik-rtds/acecloud/latest/docs)
- [Provider Configuration](https://registry.terraform.io/providers/akshay-malik-rtds/acecloud/latest/docs)
- Resource reference for every supported resource
- Data source reference for discovery
- Guides on billing, common patterns, and troubleshooting

## Authentication

Authenticate with an API key created from the AceCloud console. Set `api_key_id`, `api_key_secret`, and `api_key_service_name` in the provider block, or export the matching `ACECLOUD_API_KEY_*` environment variables. The service name must match the value supplied at key creation.

API keys are long-lived credentials scoped to a service identity — the recommended primitive for infrastructure-as-code: rotatable, revocable, and not tied to a personal user account.

Example:

```hcl
provider "acecloud" {
  api_key_id           = var.acecloud_api_key_id
  api_key_secret       = var.acecloud_api_key_secret
  api_key_service_name = "terraform"
  region               = "mumbai"
  project_id           = var.acecloud_project_id
}
```

`api_url` is optional and defaults to `https://customer.acecloud.ai/api/v1/`.

See the [Authentication guide](https://registry.terraform.io/providers/akshay-malik-rtds/acecloud/latest/docs) for details.

## Development

This provider is built with the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework). To build from source:

```bash
git clone https://github.com/akshay-malik-rtds/terraform-provider-acecloud
cd terraform-provider-acecloud
make build
make test
```

To use a locally built provider for development, follow [Terraform's `dev_overrides` guide](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers).

## Reporting Issues

- **Bugs**: [open an issue](https://github.com/akshay-malik-rtds/terraform-provider-acecloud/issues/new) using the bug report template
- **Feature requests**: same issues page, use the feature request template

## License

This provider is released under the [Mozilla Public License 2.0](./LICENSE).

Copyright (c) 2026 AceCloud
