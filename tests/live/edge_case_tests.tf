# ═══════════════════════════════════════════════════════════════
# Edge Case Tests — Bug-Prone Code Paths
# ═══════════════════════════════════════════════════════════════
# Tests for paths most likely to have bugs, identified by
# code audit of all 25 resources.
#
# Covers:
# - EC1: Instance with zero optional fields (no metadata, no tags, no desc)
# - EC2: Subnet with minimal config (Computed fields: description, enable_dhcp, gateway_ip)
# - EC3: Volume metadata removal lifecycle (2 keys → 0 keys → re-add)
# - EC4: LB Health Monitor TCP type (no HTTP-specific fields)
# - EC5: LB Health Monitor PING type
# - EC6: Volume clone without description (source_volid, no desc)
# - EC7: VPC subnet_name update (verify silent drift)
# - EC8: SG with zero rules and no description (double-null)
#
# Usage:
#   terraform apply -var="run_edge_tests=true"
#   terraform plan  -var="run_edge_tests=true"  # idempotency — CRITICAL for these tests
#   terraform apply -var="run_edge_tests=true" -var="edge_phase=update"
#   terraform plan  -var="run_edge_tests=true" -var="edge_phase=update"  # post-update idempotency
#   terraform destroy -var="run_edge_tests=true"

variable "run_edge_tests" {
  description = "Set to true to run edge case tests"
  type        = bool
  default     = false
}

variable "edge_phase" {
  description = "Phase: create or update"
  type        = string
  default     = "create"
}

# ═══════════════════════════════════════════════════════════════
# EC1: Instance with Zero Optional Fields
# ═══════════════════════════════════════════════════════════════
# No description, no metadata, no tags, no key_name, no user_data.
# Tests whether Read handles null gracefully for ALL optional fields
# and whether the API injects metadata that causes "inconsistent result".

resource "acecloud_vpc" "ec1_vpc" {
  count = var.run_edge_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                   = "tf-ec1-vpc"
  admin_state_up         = true
  subnet_name            = "tf-ec1-subnet"
  subnet_cidr            = "10.96.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_dns_nameservers = ["8.8.8.8"]
}

resource "acecloud_security_group" "ec1_sg" {
  count = var.run_edge_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-ec1-sg"
  description = "Edge case minimal SG"
}

resource "acecloud_instance" "ec1_bare" {
  count = var.run_edge_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                  = "tf-ec1-bare-instance"
  flavor_id             = var.flavor_id
  boot_uuid             = var.image_id
  source_type           = "image"
  delete_on_termination = true
  billing_type          = "hourly"

  # NO description, NO metadata, NO tags, NO key_name, NO user_data
  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.ec1_vpc[0].id]
  security_group_ids = [acecloud_security_group.ec1_sg[0].id]
}

# ═══════════════════════════════════════════════════════════════
# EC2: Subnet with Minimal Config
# ═══════════════════════════════════════════════════════════════
# No description, no enable_dhcp, no gateway_ip, no dns_nameservers.
# Tests Computed field defaults after Create.

resource "acecloud_subnet" "ec2_minimal" {
  count = var.run_edge_tests ? 1 : 0

  name       = "tf-ec2-minimal-subnet"
  cidr       = "10.95.0.0/24"
  ip_version = 4
  vpc_id     = var.flavor_id != "" && var.image_id != "" ? acecloud_vpc.ec1_vpc[0].id : null

  # NO description, NO enable_dhcp, NO gateway_ip, NO dns_nameservers
  # All Computed fields should be resolved from API response
}

# ═══════════════════════════════════════════════════════════════
# EC3: Volume Metadata Removal Lifecycle
# ═══════════════════════════════════════════════════════════════
# Create: 2 metadata keys
# Update: metadata = {} (empty map — tests clearing all keys)
# This tests whether the API clears metadata when sent {}

resource "acecloud_volume" "ec3_metadata" {
  count = var.run_edge_tests ? 1 : 0

  name        = "tf-ec3-metadata-vol"
  size        = 10
  volume_type = "ssd"
  description = "Edge case metadata removal"

  metadata = var.edge_phase == "update" ? {} : {
    env     = "test"
    purpose = "edge-case"
  }
}

# ═══════════════════════════════════════════════════════════════
# EC4: LB Health Monitor — TCP Type (No HTTP Fields)
# ═══════════════════════════════════════════════════════════════
# TCP monitor: url_path, expected_codes, http_method should NOT be set.
# Tests whether Computed fields resolve correctly for non-HTTP monitors.

resource "acecloud_load_balancer" "ec4_lb" {
  count = var.run_edge_tests ? 1 : 0

  name      = "tf-ec4-nlb"
  subnet_id = var.flavor_id != "" && var.image_id != "" ? acecloud_vpc.ec1_vpc[0].subnet_id : acecloud_vpc.test.subnet_id
  tags      = ["NLB"]
}

resource "acecloud_lb_listener" "ec4_tcp" {
  count = var.run_edge_tests ? 1 : 0

  name            = "tf-ec4-tcp-listener"
  loadbalancer_id = acecloud_load_balancer.ec4_lb[0].id
  protocol        = "TCP"
  protocol_port   = 3306
}

resource "acecloud_lb_pool" "ec4_tcp" {
  count = var.run_edge_tests ? 1 : 0

  name            = "tf-ec4-tcp-pool"
  listener_id     = acecloud_lb_listener.ec4_tcp[0].id
  loadbalancer_id = acecloud_load_balancer.ec4_lb[0].id
  protocol        = "TCP"
  lb_algorithm    = "LEAST_CONNECTIONS"
}

resource "acecloud_lb_health_monitor" "ec4_tcp" {
  count = var.run_edge_tests ? 1 : 0

  name        = "tf-ec4-tcp-hm"
  pool_id     = acecloud_lb_pool.ec4_tcp[0].id
  type        = "TCP"
  delay       = 10
  timeout     = 5
  max_retries = 3

  # NO url_path, NO expected_codes, NO http_method
  # These are Computed — should resolve to API defaults without drift
}

# EC5: Removed — pool can only have one HM. EC4 TCP covers the non-HTTP Computed field path.

# ═══════════════════════════════════════════════════════════════
# EC6: Volume Clone Without Description
# ═══════════════════════════════════════════════════════════════
# Clone from source_volid without setting description.
# Tests whether the API copies the source vol description and
# causes "inconsistent result after apply".

resource "acecloud_volume" "ec6_source" {
  count = var.run_edge_tests ? 1 : 0

  name        = "tf-ec6-source-vol"
  size        = 10
  volume_type = "ssd"
  description = "Source volume with description"
}

resource "acecloud_volume" "ec6_clone" {
  count = var.run_edge_tests ? 1 : 0

  name         = "tf-ec6-clone-vol"
  size         = 10
  volume_type  = "ssd"
  source_volid = acecloud_volume.ec6_source[0].id
  # NO description — tests if API copies source vol description
}

# ═══════════════════════════════════════════════════════════════
# EC7: VPC Subnet Name Update (Silent Drift Test)
# ═══════════════════════════════════════════════════════════════
# Create VPC, then attempt to change subnet_name on update.
# The subnet won't actually rename (VPC PUT doesn't send subnet fields).
# This verifies the documented behavior.

resource "acecloud_vpc" "ec7_vpc" {
  count = var.run_edge_tests ? 1 : 0

  name              = var.edge_phase == "update" ? "tf-ec7-vpc-renamed" : "tf-ec7-vpc"
  admin_state_up    = true
  subnet_name       = var.edge_phase == "update" ? "tf-ec7-subnet-renamed" : "tf-ec7-subnet"
  subnet_cidr       = "10.94.0.0/24"
  subnet_ip_version = 4
}

# ═══════════════════════════════════════════════════════════════
# EC8: Security Group — No Rules, No Description
# ═══════════════════════════════════════════════════════════════
# Tests bug #58 fix: SG with no description should not drift.
# Also tests empty rules handling.

resource "acecloud_security_group" "ec8_bare" {
  count = var.run_edge_tests ? 1 : 0

  name = "tf-ec8-bare-sg"
  # NO description, NO rules
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "ec1_instance_id" {
  value = var.run_edge_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ec1_bare[0].id : ""
}

output "ec1_instance_status" {
  value = var.run_edge_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ec1_bare[0].status : ""
}

output "ec2_subnet_id" {
  value = var.run_edge_tests ? acecloud_subnet.ec2_minimal[0].id : ""
}

output "ec2_subnet_dhcp" {
  value = var.run_edge_tests ? acecloud_subnet.ec2_minimal[0].enable_dhcp : null
}

output "ec2_subnet_gateway" {
  value = var.run_edge_tests ? acecloud_subnet.ec2_minimal[0].gateway_ip : ""
}

output "ec3_metadata" {
  value = var.run_edge_tests ? acecloud_volume.ec3_metadata[0].metadata : {}
}

output "ec4_tcp_hm_id" {
  value = var.run_edge_tests ? acecloud_lb_health_monitor.ec4_tcp[0].id : ""
}

output "ec5_ping_hm_id" {
  value = "removed"
}

output "ec6_clone_id" {
  value = var.run_edge_tests ? acecloud_volume.ec6_clone[0].id : ""
}

output "ec7_vpc_name" {
  value = var.run_edge_tests ? acecloud_vpc.ec7_vpc[0].name : ""
}

output "ec7_subnet_name" {
  value = var.run_edge_tests ? acecloud_vpc.ec7_vpc[0].subnet_name : ""
}

output "ec8_sg_id" {
  value = var.run_edge_tests ? acecloud_security_group.ec8_bare[0].id : ""
}
