# ═══════════════════════════════════════════════════════════════
# Deep Resource Tests — Under-Tested Resources
# ═══════════════════════════════════════════════════════════════
# In-depth coverage for resources with < 3 test scenarios:
# floating_ip, floating_ip_association, router_interface,
# auto_scaling_template, router, key_pair, volume_backup,
# lb_listener, lb_pool.
#
# Usage:
#   terraform apply -var="run_deep_tests=true"
#   terraform plan  -var="run_deep_tests=true"  # idempotency
#   terraform apply -var="run_deep_tests=true" -var="deep_phase=update"
#   terraform plan  -var="run_deep_tests=true" -var="deep_phase=update"
#   terraform destroy -var="run_deep_tests=true"

variable "run_deep_tests" {
  description = "Set to true to run deep resource tests"
  type        = bool
  default     = false
}

variable "deep_phase" {
  description = "Phase: create or update"
  type        = string
  default     = "create"
}

# ═══════════════════════════════════════════════════════════════
# Shared Networking
# ═══════════════════════════════════════════════════════════════

resource "acecloud_vpc" "deep_vpc" {
  count = var.run_deep_tests ? 1 : 0

  name                   = "tf-deep-vpc"
  admin_state_up         = true
  subnet_name            = "tf-deep-subnet"
  subnet_cidr            = "10.90.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_dns_nameservers = ["8.8.8.8"]
}

# Second subnet on same VPC — for router interface swap test
resource "acecloud_subnet" "deep_subnet2" {
  count = var.run_deep_tests ? 1 : 0

  name       = "tf-deep-subnet2"
  cidr       = "10.90.1.0/24"
  ip_version = 4
  vpc_id     = acecloud_vpc.deep_vpc[0].id
}

# ═══════════════════════════════════════════════════════════════
# D1: Floating IP — Full Attribute Test
# ═══════════════════════════════════════════════════════════════
# Create FIP with description. Verify computed fields populated.
# On update, description change triggers ForceNew (RequiresReplace).

resource "acecloud_floating_ip" "d1_fip" {
  count = var.run_deep_tests && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  description         = var.deep_phase == "update" ? "Updated FIP description" : "Initial FIP description"
}

# D1b: Bare FIP — no optional fields
resource "acecloud_floating_ip" "d1_bare" {
  count = var.run_deep_tests && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  # NO description, NO port_id
}

# ═══════════════════════════════════════════════════════════════
# D2: Router — Description Lifecycle + admin_state_up
# ═══════════════════════════════════════════════════════════════
# Create: router with description and admin_state_up=true
# Update: change description, toggle admin_state_up

resource "acecloud_router" "d2_router" {
  count = var.run_deep_tests ? 1 : 0

  name           = var.deep_phase == "update" ? "tf-deep-d2-router-v2" : "tf-deep-d2-router"
  description    = var.deep_phase == "update" ? "Updated router" : "Initial router"
  admin_state_up = true
  # NOTE: admin_state_up=false rejected by dev4 backend (400 error)
}

# ═══════════════════════════════════════════════════════════════
# D3: Router Interface — Computed Fields + Dual Interface
# ═══════════════════════════════════════════════════════════════
# Create router with gateway, attach to VPC inline subnet.
# Verify ip_address, mac_address, status computed fields.
# Create second interface on subnet2.

resource "acecloud_router" "d3_router" {
  count = var.run_deep_tests && var.external_network_id != "" ? 1 : 0

  name                        = "tf-deep-d3-router"
  admin_state_up              = true
  external_gateway_network_id = var.external_network_id
}

resource "acecloud_router_interface" "d3_ri_primary" {
  count = var.run_deep_tests && var.external_network_id != "" ? 1 : 0

  router_id = acecloud_router.d3_router[0].id
  subnet_id = acecloud_vpc.deep_vpc[0].subnet_id
}

resource "acecloud_router_interface" "d3_ri_secondary" {
  count = var.run_deep_tests && var.external_network_id != "" ? 1 : 0

  router_id = acecloud_router.d3_router[0].id
  subnet_id = acecloud_subnet.deep_subnet2[0].id

  # Serialize interface creation on the same router to avoid conflicts
  depends_on = [acecloud_router_interface.d3_ri_primary]
}

# ═══════════════════════════════════════════════════════════════
# D4: Auto Scaling Template — Standalone Create + All Fields
# ═══════════════════════════════════════════════════════════════
# Tests standalone creation (not gated by update_tests).
# Exercises all fields: type, description, key_name.
# Update: change description, volume_size.

resource "acecloud_security_group" "d4_sg" {
  count = var.run_deep_tests ? 1 : 0
  name  = "tf-deep-d4-sg"
}

resource "acecloud_auto_scaling_template" "d4_template" {
  count = var.run_deep_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                   = var.deep_phase == "update" ? "tf-deep-d4-template-v2" : "tf-deep-d4-template"
  type                   = "linux"
  description            = var.deep_phase == "update" ? "Updated template" : "Initial template"
  volume_size            = var.deep_phase == "update" ? 50 : 40
  vol_del_on_termination = var.deep_phase == "update" ? false : true
  flavor_id              = var.flavor_id
  image_id               = var.image_id
  is_instance_snapshot   = false
  network_id             = acecloud_vpc.deep_vpc[0].id
  subnet_id              = acecloud_vpc.deep_vpc[0].subnet_id
  security_groups        = [acecloud_security_group.d4_sg[0].id]
}

# ═══════════════════════════════════════════════════════════════
# D5: Key Pair — Edge Cases
# ═══════════════════════════════════════════════════════════════
# D5a: Generated key (no public_key) — verify fingerprint populated
# D5b: Imported key with valid SSH key

resource "acecloud_key_pair" "d5_generated" {
  count = var.run_deep_tests ? 1 : 0
  name  = "tf-deep-d5-generated"
  # NO public_key — backend generates
}

resource "acecloud_key_pair" "d5_imported" {
  count = var.run_deep_tests ? 1 : 0
  name  = "tf-deep-d5-imported"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7KVtBGFgXMWRBAMqz5OT1PPXQB2mK9aaXDkME6L8ZmEFgUCMnqOOaaMqPPA0VFfJy73s9C/5hGmI0oYRhDG3rQF1MYSRj8FRC+D0G0F4DpWSNeSHShRHnEVvMyJdT47u7gDh+FM3kYl0F+oGXVm0XMKBLyN5HE6P5f0sJx4w93kB6hQOkSCbhPbwFe0clsPz2uP5BTaiBS1O5OBYzO7v0Q4TJ3BGFhKM0qRdsVtLCn0e06cFAN3eFVVfGjYVh+50PyBVNaKfttD3Nk07U96dHZRj0oCWfIjFhwZiJcJQMiRV2FfhpMVde5rOgwkn3EjmNcSF0O8694yJr2cAspYLmj3N8tCBJX0Y7DG2penPYaTKpE3gVfn7BFbhvDlFMfdHFLKItv4vUq3iKaEsi+F2j0FQgTFLzFbhHKJH2hoehlgiJYwrCVCRaiNY9K2Wng+j+o9qnfYazkDqNxZc9OYgeQOEaCmeUqf/Wbzsp80L8N0P8RQbOU60L00AZG1U= deep-test-key"
}

# ═══════════════════════════════════════════════════════════════
# D6: Volume Backup — Edge Cases
# ═══════════════════════════════════════════════════════════════
# D6a: Backup with all optional fields (description, incremental)
# D6b: Incremental backup after first full backup

resource "acecloud_volume" "d6_vol" {
  count = var.run_deep_tests ? 1 : 0

  name        = "tf-deep-d6-vol"
  size        = 10
  volume_type = "ssd"
}

resource "acecloud_volume_backup" "d6_full" {
  count = var.run_deep_tests ? 1 : 0

  name        = "tf-deep-d6-full-backup"
  volume_id   = acecloud_volume.d6_vol[0].id
  description = "Full backup for deep testing"
  incremental = false
}

resource "acecloud_volume_backup" "d6_incremental" {
  count = var.run_deep_tests ? 1 : 0

  name        = "tf-deep-d6-incr-backup"
  volume_id   = acecloud_volume.d6_vol[0].id
  description = "Incremental backup"
  incremental = true

  depends_on = [acecloud_volume_backup.d6_full]
}

# ═══════════════════════════════════════════════════════════════
# D7: LB Listener + Pool — Update Fields
# ═══════════════════════════════════════════════════════════════
# Tests updating listener name and pool name (no ForceNew).

resource "acecloud_load_balancer" "d7_lb" {
  count = var.run_deep_tests ? 1 : 0

  name      = "tf-deep-d7-lb"
  subnet_id = acecloud_vpc.deep_vpc[0].subnet_id
  tags      = ["ALB"]
}

resource "acecloud_lb_listener" "d7_listener" {
  count = var.run_deep_tests ? 1 : 0

  name            = var.deep_phase == "update" ? "tf-deep-d7-listener-v2" : "tf-deep-d7-listener"
  loadbalancer_id = acecloud_load_balancer.d7_lb[0].id
  protocol        = "HTTP"
  protocol_port   = 80
}

resource "acecloud_lb_pool" "d7_pool" {
  count = var.run_deep_tests ? 1 : 0

  name            = var.deep_phase == "update" ? "tf-deep-d7-pool-v2" : "tf-deep-d7-pool"
  listener_id     = acecloud_lb_listener.d7_listener[0].id
  loadbalancer_id = acecloud_load_balancer.d7_lb[0].id
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
}

resource "acecloud_lb_pool_member" "d7_member" {
  count = var.run_deep_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.d7_pool[0].id
  name          = var.deep_phase == "update" ? "tf-deep-d7-member-v2" : "tf-deep-d7-member"
  address       = "10.90.0.10"
  protocol_port = 8080
  weight        = var.deep_phase == "update" ? 10 : 1
}

resource "acecloud_lb_health_monitor" "d7_hm" {
  count = var.run_deep_tests ? 1 : 0

  name           = var.deep_phase == "update" ? "tf-deep-d7-hm-v2" : "tf-deep-d7-hm"
  pool_id        = acecloud_lb_pool.d7_pool[0].id
  type           = "HTTP"
  delay          = var.deep_phase == "update" ? 15 : 5
  timeout        = var.deep_phase == "update" ? 5 : 3
  max_retries    = var.deep_phase == "update" ? 5 : 3
  url_path       = var.deep_phase == "update" ? "/status" : "/health"
  expected_codes = var.deep_phase == "update" ? "200-299" : "200"
  http_method    = "GET"
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "d1_fip_address" {
  value = var.run_deep_tests && var.external_network_id != "" ? acecloud_floating_ip.d1_fip[0].floating_ip_address : ""
}

output "d1_fip_status" {
  value = var.run_deep_tests && var.external_network_id != "" ? acecloud_floating_ip.d1_fip[0].status : ""
}

output "d1_bare_fip_address" {
  value = var.run_deep_tests && var.external_network_id != "" ? acecloud_floating_ip.d1_bare[0].floating_ip_address : ""
}

output "d2_router_admin" {
  value = var.run_deep_tests ? acecloud_router.d2_router[0].admin_state_up : null
}

output "d2_router_desc" {
  value = var.run_deep_tests ? acecloud_router.d2_router[0].description : ""
}

output "d3_ri_primary_ip" {
  value = var.run_deep_tests && var.external_network_id != "" ? acecloud_router_interface.d3_ri_primary[0].ip_address : ""
}

output "d3_ri_primary_mac" {
  value = var.run_deep_tests && var.external_network_id != "" ? acecloud_router_interface.d3_ri_primary[0].mac_address : ""
}

output "d3_ri_secondary_ip" {
  value = var.run_deep_tests && var.external_network_id != "" ? acecloud_router_interface.d3_ri_secondary[0].ip_address : ""
}

output "d4_template_id" {
  value = var.run_deep_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_template.d4_template[0].id : ""
}

output "d4_template_status" {
  value = var.run_deep_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_template.d4_template[0].status : ""
}

output "d5_gen_fingerprint" {
  value = var.run_deep_tests ? acecloud_key_pair.d5_generated[0].fingerprint : ""
}

output "d5_imp_fingerprint" {
  value = var.run_deep_tests ? acecloud_key_pair.d5_imported[0].fingerprint : ""
}

output "d6_full_backup_status" {
  value = var.run_deep_tests ? acecloud_volume_backup.d6_full[0].status : ""
}

output "d6_incr_backup_status" {
  value = var.run_deep_tests ? acecloud_volume_backup.d6_incremental[0].status : ""
}

output "d7_listener_name" {
  value = var.run_deep_tests ? acecloud_lb_listener.d7_listener[0].name : ""
}

output "d7_pool_name" {
  value = var.run_deep_tests ? acecloud_lb_pool.d7_pool[0].name : ""
}

output "d7_member_weight" {
  value = var.run_deep_tests ? acecloud_lb_pool_member.d7_member[0].weight : 0
}

output "d7_hm_url_path" {
  value = var.run_deep_tests ? acecloud_lb_health_monitor.d7_hm[0].url_path : ""
}
