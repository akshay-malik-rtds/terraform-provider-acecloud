# ═══════════════════════════════════════════════════════════════
# Expanded Test Coverage — gaps not covered by main.tf
# ═══════════════════════════════════════════════════════════════
# Gate: set var.run_expanded_tests = true to create these resources.
# Depends on base resources from main.tf (VPC, subnet, SG, key pair, etc.)

variable "run_expanded_tests" {
  description = "Set to true to run expanded test scenarios"
  type        = bool
  default     = false
}

# ═══════════════════════════════════════════════════════════════
# EX1: Volume -> Snapshot -> New Volume from Snapshot
# ═══════════════════════════════════════════════════════════════
# Tests the snapshot_id source for volume creation and 8GB minimum boundary.

resource "acecloud_volume" "ex1_source" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex1-vol-source"
  size        = 8
  volume_type = "ssd"
  description = "EX1 source volume, 8GB minimum boundary"
}

resource "acecloud_snapshot" "ex1_snap" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex1-snapshot"
  volume_id   = acecloud_volume.ex1_source[0].id
  description = "EX1 snapshot from 8GB source volume"

  depends_on = [acecloud_volume.ex1_source]
}

resource "acecloud_volume" "ex1_from_snapshot" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex1-vol-from-snapshot"
  size        = 8
  volume_type = "ssd"
  snapshot_id = acecloud_snapshot.ex1_snap[0].id
  description = "EX1 volume restored from snapshot"

  depends_on = [acecloud_snapshot.ex1_snap]
}

# ═══════════════════════════════════════════════════════════════
# EX2: Volume -> Backup -> New Volume from Backup
# ═══════════════════════════════════════════════════════════════
# Tests the backup_id source for volume creation.

resource "acecloud_volume" "ex2_source" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex2-vol-source"
  size        = 8
  volume_type = "ssd"
  description = "EX2 source volume for backup chain"
}

resource "acecloud_volume_backup" "ex2_backup" {
  count = var.run_expanded_tests ? 1 : 0

  name      = "tf-ex2-backup"
  volume_id = acecloud_volume.ex2_source[0].id

  depends_on = [acecloud_volume.ex2_source]
}

resource "acecloud_volume" "ex2_from_backup" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex2-vol-from-backup"
  size        = 8
  volume_type = "ssd"
  backup_id   = acecloud_volume_backup.ex2_backup[0].id
  description = "EX2 volume restored from backup"

  depends_on = [acecloud_volume_backup.ex2_backup]
}

# ═══════════════════════════════════════════════════════════════
# EX3: Billing Type Tests
# ═══════════════════════════════════════════════════════════════
# Tests non-default billing types on instance and volume.

# EX3a: Instance with billing_type = "hourly" (default is "monthly")
resource "acecloud_instance" "ex3_hourly" {
  count = var.run_expanded_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-ex3-instance-hourly"
  description = "EX3 instance with hourly billing"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  billing_type          = "hourly"
  delete_on_termination = true

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.test.id]
  security_group_ids = [acecloud_security_group.web.id]
  key_name           = acecloud_key_pair.generated.name
}

# EX3b: Standalone volume with billing_type = "monthly" (default is "hourly")
resource "acecloud_volume" "ex3_monthly" {
  count = var.run_expanded_tests ? 1 : 0

  name         = "tf-ex3-vol-monthly"
  size         = 10
  volume_type  = "ssd"
  billing_type = "monthly"
  description  = "EX3 volume with monthly billing"
}

# EX3c: Instance with boot volume billing_type = "monthly" (default is "hourly" for volumes)
resource "acecloud_instance" "ex3_monthly_vol" {
  count = var.run_expanded_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-ex3-instance-monthly-vol"
  description = "EX3 instance with monthly boot volume billing"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  billing_type          = "monthly"
  delete_on_termination = true

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "monthly"
  }

  network_ids        = [acecloud_vpc.test.id]
  security_group_ids = [acecloud_security_group.web.id]
  key_name           = acecloud_key_pair.generated.name
}

# ═══════════════════════════════════════════════════════════════
# EX4: Instance with Multiple Options
# ═══════════════════════════════════════════════════════════════
# Tests config_drive, user_data, and explicit availability_zone.

resource "acecloud_instance" "ex4_options" {
  count = var.run_expanded_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-ex4-instance-options"
  description = "EX4 instance with config drive, user data, AZ"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  delete_on_termination = true
  config_drive          = true
  availability_zone     = "nova"

  # Base64 of: #!/bin/bash\necho "hello from terraform" > /tmp/tf-test.txt
  user_data = "IyEvYmluL2Jhc2gKZWNobyAiaGVsbG8gZnJvbSB0ZXJyYWZvcm0iID4gL3RtcC90Zi10ZXN0LnR4dA=="

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.test.id]
  security_group_ids = [acecloud_security_group.web.id]
  key_name           = acecloud_key_pair.generated.name

  metadata = {
    managed_by = "terraform"
    test_case  = "ex4-options"
  }
}

# ═══════════════════════════════════════════════════════════════
# EX5: Security Group Edge Cases
# ═══════════════════════════════════════════════════════════════
# Tests ICMP (no port range), UDP, and wide port range (1-65535).

resource "acecloud_security_group" "ex5_edge" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex5-sg-edge-cases"
  description = "EX5 security group with ICMP, UDP, wide port range"

  # ICMP rule — no port range required
  rules {
    direction        = "ingress"
    protocol         = "icmp"
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # UDP rule with specific port
  rules {
    direction        = "ingress"
    protocol         = "udp"
    port_range_min   = 53
    port_range_max   = 53
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # Wide TCP port range (1-65535)
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "10.0.0.0/8"
    ethertype        = "IPv4"
  }

  # UDP egress full range
  rules {
    direction        = "egress"
    protocol         = "udp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

# ═══════════════════════════════════════════════════════════════
# EX6: Volume Size Boundaries
# ═══════════════════════════════════════════════════════════════
# Tests minimum volume size (8 GB).

resource "acecloud_volume" "ex6_min" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex6-vol-min-size"
  size        = 8
  volume_type = "ssd"
  description = "EX6 minimum 8GB volume boundary test"
}

# ═══════════════════════════════════════════════════════════════
# EX7: LB with Multiple Listeners (HTTP + TCP)
# ═══════════════════════════════════════════════════════════════
# Tests a single LB with two listeners on different ports/protocols.

resource "acecloud_load_balancer" "ex7_multi" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex7-lb-multi-listener"
  description = "EX7 load balancer with HTTP and TCP listeners"
  subnet_id   = var.subnet_id != "" ? var.subnet_id : acecloud_vpc.test.subnet_id
  tags        = ["ALB"]
}

# HTTP listener on port 80
resource "acecloud_lb_listener" "ex7_http" {
  count = var.run_expanded_tests ? 1 : 0

  name            = "tf-ex7-listener-http"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.ex7_multi[0].id
}

# TCP listener on port 443
resource "acecloud_lb_listener" "ex7_tcp" {
  count = var.run_expanded_tests ? 1 : 0

  name            = "tf-ex7-listener-tcp"
  protocol        = "TCP"
  protocol_port   = 443
  loadbalancer_id = acecloud_load_balancer.ex7_multi[0].id

  depends_on = [acecloud_lb_listener.ex7_http]
}

# Pool for HTTP listener
resource "acecloud_lb_pool" "ex7_http_pool" {
  count = var.run_expanded_tests ? 1 : 0

  name            = "tf-ex7-pool-http"
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
  listener_id     = acecloud_lb_listener.ex7_http[0].id
  loadbalancer_id = acecloud_load_balancer.ex7_multi[0].id

  depends_on = [acecloud_lb_listener.ex7_http]
}

# Pool for TCP listener
resource "acecloud_lb_pool" "ex7_tcp_pool" {
  count = var.run_expanded_tests ? 1 : 0

  name            = "tf-ex7-pool-tcp"
  protocol        = "TCP"
  lb_algorithm    = "LEAST_CONNECTIONS"
  listener_id     = acecloud_lb_listener.ex7_tcp[0].id
  loadbalancer_id = acecloud_load_balancer.ex7_multi[0].id

  depends_on = [acecloud_lb_pool.ex7_http_pool]
}

# ═══════════════════════════════════════════════════════════════
# EX8: Data Source Chaining
# ═══════════════════════════════════════════════════════════════
# Validates data sources return expected data.

output "ex8_flavor_count" {
  description = "EX8: Number of available flavors"
  value       = var.run_expanded_tests ? length(data.acecloud_flavors.all.flavors) : 0
}

output "ex8_image_count" {
  description = "EX8: Number of available images"
  value       = var.run_expanded_tests ? length(data.acecloud_images.all.images) : 0
}

output "ex8_flavor_ids" {
  description = "EX8: All flavor IDs (validates data source returns data)"
  value       = var.run_expanded_tests ? [for f in data.acecloud_flavors.all.flavors : f.id] : []
}

output "ex8_image_ids" {
  description = "EX8: All image IDs (validates data source returns data)"
  value       = var.run_expanded_tests ? [for i in data.acecloud_images.all.images : i.id] : []
}

output "ex8_vpc_count" {
  description = "EX8: Number of VPCs"
  value       = var.run_expanded_tests ? length(data.acecloud_vpcs.all.vpcs) : 0
}

output "ex8_sg_count" {
  description = "EX8: Number of security groups"
  value       = var.run_expanded_tests ? length(data.acecloud_security_groups.all.security_groups) : 0
}

# ═══════════════════════════════════════════════════════════════
# EX9: Description at Max Length (255 characters)
# ═══════════════════════════════════════════════════════════════
# Tests the 255-character description boundary for volume and SG.

resource "acecloud_volume" "ex9_max_desc" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex9-vol-max-desc"
  size        = 8
  volume_type = "ssd"
  description = "This is a maximum length description for testing the 255 character boundary on volume resources. It contains only valid characters including letters, numbers, hyphens, underscores, periods, commas and spaces to pass validation. Padding here to reach limit"
}

resource "acecloud_security_group" "ex9_max_desc" {
  count = var.run_expanded_tests ? 1 : 0

  name        = "tf-ex9-sg-max-desc"
  description = "This is a maximum length description for testing the 255 character boundary on security group resources. It uses only valid characters including letters, numbers, hyphens, underscores, periods, commas and spaces. Padding to approach the limit now, done ok"

  # Minimal rule to make the SG functional
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

# ═══════════════════════════════════════════════════════════════
# EX10: Floating IP with Description Only (no port_id, no billing_type)
# ═══════════════════════════════════════════════════════════════
# Tests bare-minimum floating IP creation with just network + description.

resource "acecloud_floating_ip" "ex10_bare" {
  count = var.run_expanded_tests && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  description         = "EX10 floating IP, description only, no port"
}

# ═══════════════════════════════════════════════════════════════
# Outputs — Expanded Tests
# ═══════════════════════════════════════════════════════════════

# --- EX1: Volume -> Snapshot -> Volume ---

output "ex1_source_volume_id" {
  description = "EX1: Source volume ID"
  value       = var.run_expanded_tests ? acecloud_volume.ex1_source[0].id : "skipped"
}

output "ex1_snapshot_id" {
  description = "EX1: Snapshot ID"
  value       = var.run_expanded_tests ? acecloud_snapshot.ex1_snap[0].id : "skipped"
}

output "ex1_volume_from_snapshot_id" {
  description = "EX1: Volume created from snapshot ID"
  value       = var.run_expanded_tests ? acecloud_volume.ex1_from_snapshot[0].id : "skipped"
}

output "ex1_volume_from_snapshot_status" {
  description = "EX1: Volume from snapshot status"
  value       = var.run_expanded_tests ? acecloud_volume.ex1_from_snapshot[0].status : "skipped"
}

# --- EX2: Volume -> Backup -> Volume ---

output "ex2_source_volume_id" {
  description = "EX2: Source volume ID"
  value       = var.run_expanded_tests ? acecloud_volume.ex2_source[0].id : "skipped"
}

output "ex2_backup_id" {
  description = "EX2: Backup ID"
  value       = var.run_expanded_tests ? acecloud_volume_backup.ex2_backup[0].id : "skipped"
}

output "ex2_volume_from_backup_id" {
  description = "EX2: Volume created from backup ID"
  value       = var.run_expanded_tests ? acecloud_volume.ex2_from_backup[0].id : "skipped"
}

output "ex2_volume_from_backup_status" {
  description = "EX2: Volume from backup status"
  value       = var.run_expanded_tests ? acecloud_volume.ex2_from_backup[0].status : "skipped"
}

# --- EX3: Billing Types ---

output "ex3_instance_hourly_id" {
  description = "EX3: Instance with hourly billing ID"
  value       = var.run_expanded_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ex3_hourly[0].id : "skipped"
}

output "ex3_volume_monthly_id" {
  description = "EX3: Volume with monthly billing ID"
  value       = var.run_expanded_tests ? acecloud_volume.ex3_monthly[0].id : "skipped"
}

output "ex3_instance_monthly_vol_id" {
  description = "EX3: Instance with monthly boot volume ID"
  value       = var.run_expanded_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ex3_monthly_vol[0].id : "skipped"
}

# --- EX4: Instance Options ---

output "ex4_instance_options_id" {
  description = "EX4: Instance with config_drive, user_data, AZ"
  value       = var.run_expanded_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ex4_options[0].id : "skipped"
}

output "ex4_instance_options_status" {
  description = "EX4: Instance with options status"
  value       = var.run_expanded_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ex4_options[0].status : "skipped"
}

# --- EX5: Security Group Edge Cases ---

output "ex5_sg_edge_id" {
  description = "EX5: Security group with ICMP, UDP, wide range"
  value       = var.run_expanded_tests ? acecloud_security_group.ex5_edge[0].id : "skipped"
}

# --- EX6: Volume Size Boundary ---

output "ex6_vol_min_id" {
  description = "EX6: Minimum size (8GB) volume ID"
  value       = var.run_expanded_tests ? acecloud_volume.ex6_min[0].id : "skipped"
}

output "ex6_vol_min_status" {
  description = "EX6: Minimum size volume status"
  value       = var.run_expanded_tests ? acecloud_volume.ex6_min[0].status : "skipped"
}

# --- EX7: LB Multi-Listener ---

output "ex7_lb_id" {
  description = "EX7: Multi-listener LB ID"
  value       = var.run_expanded_tests ? acecloud_load_balancer.ex7_multi[0].id : "skipped"
}

output "ex7_listener_http_id" {
  description = "EX7: HTTP listener ID"
  value       = var.run_expanded_tests ? acecloud_lb_listener.ex7_http[0].id : "skipped"
}

output "ex7_listener_tcp_id" {
  description = "EX7: TCP listener ID"
  value       = var.run_expanded_tests ? acecloud_lb_listener.ex7_tcp[0].id : "skipped"
}

output "ex7_pool_http_id" {
  description = "EX7: HTTP pool ID"
  value       = var.run_expanded_tests ? acecloud_lb_pool.ex7_http_pool[0].id : "skipped"
}

output "ex7_pool_tcp_id" {
  description = "EX7: TCP pool ID"
  value       = var.run_expanded_tests ? acecloud_lb_pool.ex7_tcp_pool[0].id : "skipped"
}

# --- EX9: Max Description ---

output "ex9_vol_max_desc_id" {
  description = "EX9: Volume with 255-char description ID"
  value       = var.run_expanded_tests ? acecloud_volume.ex9_max_desc[0].id : "skipped"
}

output "ex9_sg_max_desc_id" {
  description = "EX9: SG with 255-char description ID"
  value       = var.run_expanded_tests ? acecloud_security_group.ex9_max_desc[0].id : "skipped"
}

# --- EX10: Floating IP Bare ---

output "ex10_floating_ip_id" {
  description = "EX10: Bare floating IP ID (no port)"
  value       = var.run_expanded_tests && var.external_network_id != "" ? acecloud_floating_ip.ex10_bare[0].id : "skipped"
}

output "ex10_floating_ip_address" {
  description = "EX10: Bare floating IP address"
  value       = var.run_expanded_tests && var.external_network_id != "" ? acecloud_floating_ip.ex10_bare[0].floating_ip_address : "skipped"
}

output "ex10_floating_ip_status" {
  description = "EX10: Bare floating IP status"
  value       = var.run_expanded_tests && var.external_network_id != "" ? acecloud_floating_ip.ex10_bare[0].status : "skipped"
}
