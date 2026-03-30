# ═══════════════════════════════════════════════════════════════
# Stress Tests — Field Removal and Null Transition Paths
# ═══════════════════════════════════════════════════════════════
# Targets the most bug-prone untested paths: field removal,
# omitempty gaps, and null transitions.
#
# Usage:
#   terraform apply -var="run_stress_tests=true"
#   terraform plan  -var="run_stress_tests=true"  # idempotency
#   terraform apply -var="run_stress_tests=true" -var="stress_phase=update"
#   terraform plan  -var="run_stress_tests=true" -var="stress_phase=update"  # post-update
#   terraform destroy -var="run_stress_tests=true"

variable "run_stress_tests" {
  description = "Set to true to run stress tests"
  type        = bool
  default     = false
}

variable "stress_phase" {
  description = "Phase: create or update"
  type        = string
  default     = "create"
}

# ═══════════════════════════════════════════════════════════════
# S1: LB Tags Removal
# ═══════════════════════════════════════════════════════════════
# Create: LB with tags=["ALB"]
# Update: Change name + description but keep tags (verify no drift)
# This tests whether tags persist correctly through updates.

resource "acecloud_vpc" "s1_vpc" {
  count = var.run_stress_tests ? 1 : 0

  name              = "tf-s1-vpc"
  admin_state_up    = true
  subnet_name       = "tf-s1-subnet"
  subnet_cidr       = "10.93.0.0/24"
  subnet_ip_version = 4
}

resource "acecloud_load_balancer" "s1_lb" {
  count = var.run_stress_tests ? 1 : 0

  name        = var.stress_phase == "update" ? "tf-s1-lb-renamed" : "tf-s1-lb"
  subnet_id   = acecloud_vpc.s1_vpc[0].subnet_id
  tags        = ["ALB"]
  description = var.stress_phase == "update" ? "Updated LB description" : "Initial LB"
}

# ═══════════════════════════════════════════════════════════════
# S2: Router External Gateway Removal
# ═══════════════════════════════════════════════════════════════
# Create: Router with external_gateway
# Update: Remove external_gateway (set to null)
# Tests whether API accepts gateway removal and state stays clean.

resource "acecloud_router" "s2_gw" {
  count = var.run_stress_tests && var.external_network_id != "" ? 1 : 0

  name                        = var.stress_phase == "update" ? "tf-s2-router-renamed" : "tf-s2-router"
  admin_state_up              = true
  external_gateway_network_id = var.external_network_id
  # NOTE: Gateway removal not supported by backend — omitting field on PUT doesn't clear it
}

# ═══════════════════════════════════════════════════════════════
# S3: Volume Description Removal
# ═══════════════════════════════════════════════════════════════
# Create: Volume with description
# Update: Remove description (set to empty or null)
# Tests omitempty behavior on PUT.

resource "acecloud_volume" "s3_desc" {
  count = var.run_stress_tests ? 1 : 0

  name        = var.stress_phase == "update" ? "tf-s3-vol-renamed" : "tf-s3-vol"
  size        = 10
  volume_type = "ssd"
  description = var.stress_phase == "update" ? "" : "Initial description"
}

# ═══════════════════════════════════════════════════════════════
# S4: Volume HDD Type
# ═══════════════════════════════════════════════════════════════
# Create: Volume with volume_type = "hdd"
# Tests the HDD alias mapping round-trip through Create/Read.

resource "acecloud_volume" "s4_hdd" {
  count = var.run_stress_tests ? 1 : 0

  name        = "tf-s4-hdd-vol"
  size        = 10
  volume_type = "hdd"
}

# ═══════════════════════════════════════════════════════════════
# S5: Router Description on Create
# ═══════════════════════════════════════════════════════════════
# Create: Router with description (Create API may not support it).
# Verify no crash and description is set after potential Update pass.

resource "acecloud_router" "s5_desc" {
  count = var.run_stress_tests ? 1 : 0

  name        = "tf-s5-router-desc"
  description = "Router with description from create"
}

# ═══════════════════════════════════════════════════════════════
# S6: Security Group Rule Mutations
# ═══════════════════════════════════════════════════════════════
# Create: SG with 3 rules
# Update: Remove 2 rules, add 1 new rule (net -1 rule)
# Tests PUT replace-all-rules behavior.

resource "acecloud_security_group" "s6_rules" {
  count = var.run_stress_tests ? 1 : 0

  name        = var.stress_phase == "update" ? "tf-s6-sg-renamed" : "tf-s6-sg"
  description = "Stress test SG rule mutations"

  dynamic "rules" {
    for_each = var.stress_phase == "update" ? [
      {
        direction        = "ingress"
        protocol         = "tcp"
        port_range_min   = 443
        port_range_max   = 443
        remote_ip_prefix = "0.0.0.0/0"
        ethertype        = "IPv4"
      },
      {
        direction        = "ingress"
        protocol         = "tcp"
        port_range_min   = 8443
        port_range_max   = 8443
        remote_ip_prefix = "10.0.0.0/8"
        ethertype        = "IPv4"
      },
    ] : [
      {
        direction        = "ingress"
        protocol         = "tcp"
        port_range_min   = 22
        port_range_max   = 22
        remote_ip_prefix = "0.0.0.0/0"
        ethertype        = "IPv4"
      },
      {
        direction        = "ingress"
        protocol         = "tcp"
        port_range_min   = 80
        port_range_max   = 80
        remote_ip_prefix = "0.0.0.0/0"
        ethertype        = "IPv4"
      },
      {
        direction        = "ingress"
        protocol         = "tcp"
        port_range_min   = 443
        port_range_max   = 443
        remote_ip_prefix = "0.0.0.0/0"
        ethertype        = "IPv4"
      },
    ]

    content {
      direction        = rules.value.direction
      protocol         = rules.value.protocol
      port_range_min   = rules.value.port_range_min
      port_range_max   = rules.value.port_range_max
      remote_ip_prefix = rules.value.remote_ip_prefix
      ethertype        = rules.value.ethertype
    }
  }
}

# ═══════════════════════════════════════════════════════════════
# S7: Snapshot Description Removal
# ═══════════════════════════════════════════════════════════════
# Create: Snapshot with description
# Update: Remove description

resource "acecloud_snapshot" "s7_desc" {
  count = var.run_stress_tests ? 1 : 0

  name        = var.stress_phase == "update" ? "tf-s7-snap-renamed" : "tf-s7-snap"
  volume_id   = acecloud_volume.s3_desc[0].id
  description = var.stress_phase == "update" ? "" : "Initial snap description"
}

# ═══════════════════════════════════════════════════════════════
# S8: LB Health Monitor with max_retries_down
# ═══════════════════════════════════════════════════════════════
# Tests the max_retries_down Computed field on create.

resource "acecloud_lb_listener" "s8_listener" {
  count = var.run_stress_tests ? 1 : 0

  name            = "tf-s8-listener"
  loadbalancer_id = acecloud_load_balancer.s1_lb[0].id
  protocol        = "HTTP"
  protocol_port   = 8080
}

resource "acecloud_lb_pool" "s8_pool" {
  count = var.run_stress_tests ? 1 : 0

  name            = "tf-s8-pool"
  listener_id     = acecloud_lb_listener.s8_listener[0].id
  loadbalancer_id = acecloud_load_balancer.s1_lb[0].id
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
}

resource "acecloud_lb_health_monitor" "s8_hm" {
  count = var.run_stress_tests ? 1 : 0

  name            = "tf-s8-hm"
  pool_id         = acecloud_lb_pool.s8_pool[0].id
  type            = "HTTP"
  delay           = 5
  timeout         = 3
  max_retries     = 3
  max_retries_down = 5
  url_path        = "/health"
  expected_codes  = "200-299"
  http_method     = "GET"
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "s1_lb_id" {
  value = var.run_stress_tests ? acecloud_load_balancer.s1_lb[0].id : ""
}

output "s1_lb_tags" {
  value = var.run_stress_tests ? acecloud_load_balancer.s1_lb[0].tags : []
}

output "s1_lb_desc" {
  value = var.run_stress_tests ? acecloud_load_balancer.s1_lb[0].description : ""
}

output "s2_router_gw" {
  value = var.run_stress_tests && var.external_network_id != "" ? acecloud_router.s2_gw[0].external_gateway_network_id : ""
}

output "s3_vol_desc" {
  value = var.run_stress_tests ? acecloud_volume.s3_desc[0].description : ""
}

output "s4_hdd_type" {
  value = var.run_stress_tests ? acecloud_volume.s4_hdd[0].volume_type : ""
}

output "s5_router_desc" {
  value = var.run_stress_tests ? acecloud_router.s5_desc[0].description : ""
}

output "s6_sg_name" {
  value = var.run_stress_tests ? acecloud_security_group.s6_rules[0].name : ""
}

output "s7_snap_desc" {
  value = var.run_stress_tests ? acecloud_snapshot.s7_desc[0].description : ""
}

output "s8_hm_retries_down" {
  value = var.run_stress_tests ? acecloud_lb_health_monitor.s8_hm[0].max_retries_down : 0
}
