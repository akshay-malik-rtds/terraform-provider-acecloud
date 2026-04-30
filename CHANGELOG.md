# Changelog

All notable changes to this provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this provider adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-30

### Added

First public release of the AceCloud Terraform Provider.

**Resources (19):**
- `acecloud_instance` — compute instances
- `acecloud_key_pair` — SSH key pairs
- `acecloud_volume` — block storage volumes
- `acecloud_snapshot` — point-in-time volume snapshots
- `acecloud_volume_backup` — volume backups
- `acecloud_vpc` — virtual private clouds with inline subnet
- `acecloud_subnet` — additional subnets
- `acecloud_router` — virtual routers
- `acecloud_router_interface` — subnet-to-router attachments
- `acecloud_security_group` — security groups with inline rules
- `acecloud_floating_ip` — public IP addresses
- `acecloud_floating_ip_association` — instance-to-FIP bindings
- `acecloud_load_balancer` — application load balancers
- `acecloud_lb_listener` — load balancer listeners
- `acecloud_lb_pool` — backend pools
- `acecloud_lb_pool_member` — pool members
- `acecloud_lb_health_monitor` — health checks
- `acecloud_auto_scaling_template` — auto scaling launch templates
- `acecloud_auto_scaling_deployment` — auto scaling deployments

**Data Sources (5):**
- `acecloud_flavors` — list compute flavors
- `acecloud_images` — list boot images
- `acecloud_vpcs` — list VPCs
- `acecloud_security_groups` — list security groups
- `acecloud_routers` — list routers

**Authentication:**
- Static API token
- Email and password login
- `ace-cli` config (`~/.ace/config.json`)
- Environment variables for all credentials

**Features:**
- Three-tier billing model: hourly, monthly, quarterly, half-yearly, yearly (where applicable)
- Cross-resource dependency support (e.g., floating IP association with router interface prerequisite)
- Async resource provisioning with configurable timeouts
- Schema validation matching the AceCloud platform's input rules

[Unreleased]: https://github.com/akshay-malik-rtds/terraform-provider-acecloud/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/akshay-malik-rtds/terraform-provider-acecloud/releases/tag/v0.1.0
