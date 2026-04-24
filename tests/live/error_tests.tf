# ═══════════════════════════════════════════════════════════════
# Error Sanitization & Message Consistency Tests
# ═══════════════════════════════════════════════════════════════
# These resources intentionally trigger API errors to verify:
# 1. Error messages use "Failed to <verb> <resource>" pattern
# 2. No OpenStack/Neutron/Nova/Cinder/etc. terms leak through
# 3. No raw response bodies appear in error output
#
# Usage:
#   terraform apply -var="run_error_tests=true" -auto-approve 2>&1 | tee /tmp/error_test.log
#   grep -iE "openstack|neutron|nova|cinder|octavia|glance|keystone" /tmp/error_test.log
#   (should produce ZERO matches outside of this comment block)
#
# Each resource is gated by var.run_error_tests (default false).

variable "run_error_tests" {
  description = "Set to true to run error-triggering test resources"
  type        = bool
  default     = false
}

# ─── Test E1: Bad volume type ────────────────────────────────
# Expected: "Failed to create volume" with sanitized message
resource "acecloud_volume" "error_bad_type" {
  count = var.run_error_tests ? 1 : 0

  name        = "tf-error-test-bad-type"
  size        = 10
  volume_type = "nonexistent-type-xyz"
}

# ─── Test E2: Bad flavor ID ─────────────────────────────────
# Expected: "Failed to create instance" with sanitized message
resource "acecloud_instance" "error_bad_flavor" {
  count = var.run_error_tests ? 1 : 0

  name        = "tf-error-test-bad-flavor"
  flavor_id   = "00000000-0000-0000-0000-000000000000"
  boot_uuid   = var.image_id
  source_type = "image"

  delete_on_termination = true

  volumes {
    size        = 20
    boot        = true
    volume_type = "ssd"
  }

  network_ids        = ["8f0b85d7-6517-4a80-8c32-9b7d93006d48"]
  security_group_ids = []
}

# ─── Test E3: Bad image ID ──────────────────────────────────
# Expected: "Failed to create instance" with sanitized message
resource "acecloud_instance" "error_bad_image" {
  count = var.run_error_tests ? 1 : 0

  name        = "tf-error-test-bad-image"
  flavor_id   = var.flavor_id
  boot_uuid   = "00000000-0000-0000-0000-000000000000"
  source_type = "image"

  delete_on_termination = true

  volumes {
    size        = 20
    boot        = true
    volume_type = "ssd"
  }

  network_ids        = ["8f0b85d7-6517-4a80-8c32-9b7d93006d48"]
  security_group_ids = []
}

# ─── Test E4: Bad network ID ────────────────────────────────
# Expected: "Failed to create instance" with sanitized message (no "Neutron")
resource "acecloud_instance" "error_bad_network" {
  count = var.run_error_tests ? 1 : 0

  name        = "tf-error-test-bad-network"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  delete_on_termination = true

  volumes {
    size        = 20
    boot        = true
    volume_type = "ssd"
  }

  network_ids        = ["00000000-0000-0000-0000-000000000000"]
  security_group_ids = []
}

# ─── Test E5: Bad subnet ID for LB ──────────────────────────
# Expected: "Failed to create load balancer" with sanitized message (no "Octavia")
resource "acecloud_load_balancer" "error_bad_subnet" {
  count = var.run_error_tests ? 1 : 0

  name      = "tf-error-test-bad-subnet-lb"
  subnet_id = "00000000-0000-0000-0000-000000000000"
  tags      = ["test"]
}

# ─── Test E6: Bad floating network ID ────────────────────────
# Expected: "Failed to create floating IP" with sanitized message
resource "acecloud_floating_ip" "error_bad_network" {
  count = var.run_error_tests ? 1 : 0

  floating_network_id = "00000000-0000-0000-0000-000000000000"
}

# ═══════════════════════════════════════════════════════════════
# Error Tests for New Resources (Session 5+)
# ═══════════════════════════════════════════════════════════════

# ─── Test E7: Auto Scaling Template — Bad flavor ID ──────────
# Expected: "Failed to create auto scaling template" with sanitized message
resource "acecloud_auto_scaling_template" "error_bad_flavor" {
  count = var.run_error_tests ? 1 : 0

  name                   = "tf-error-test-bad-as-flavor"
  type                   = "linux"
  volume_size            = 20
  vol_del_on_termination = true
  flavor_id              = "00000000-0000-0000-0000-000000000000"
  image_id               = var.image_id
  network_id             = "00000000-0000-0000-0000-000000000001"
  subnet_id              = var.subnet_id
  security_groups        = ["00000000-0000-0000-0000-000000000002"]
  is_instance_snapshot   = false
}

# ─── Test E8: Auto Scaling Template — Bad network ID ─────────
# Expected: "Failed to create auto scaling template" with sanitized message
resource "acecloud_auto_scaling_template" "error_bad_network" {
  count = var.run_error_tests ? 1 : 0

  name                   = "tf-error-test-bad-as-network"
  type                   = "linux"
  volume_size            = 20
  vol_del_on_termination = true
  flavor_id              = var.flavor_id
  image_id               = var.image_id
  network_id             = "00000000-0000-0000-0000-000000000000"
  subnet_id              = var.subnet_id
  security_groups        = ["00000000-0000-0000-0000-000000000002"]
  is_instance_snapshot   = false
}

# ─── Test E9: Auto Scaling Deployment — Bad template ID ──────
# Expected: "Failed to create auto scaling deployment" with sanitized message
resource "acecloud_auto_scaling_deployment" "error_bad_template" {
  count = var.run_error_tests ? 1 : 0

  name                   = "tf-error-test-bad-as-deploy"
  template_id            = "00000000-0000-0000-0000-000000000000"
  desired_capacity       = 1
  max_capacity           = 2
  nodes_scale_count      = 1
  scaling_parameter      = "cpu"
  min_threshold          = 40
  max_threshold          = 80
  cool_down_time         = 120
  user_email             = ["test@example.com"]
  is_integrated_with_lb  = false
}

# ═══════════════════════════════════════════════════════════════
# Error Tests E15–E22: Additional Coverage
# ═══════════════════════════════════════════════════════════════

# ─── Test E15: (removed — schema validation failures block entire plan)

# ─── Test E16: Instance with bad network ID ────────────────────
# Expected: "Failed to create instance"
resource "acecloud_instance" "error_bad_net_billing" {
  count = var.run_error_tests ? 1 : 0

  name        = "tf-error-test-bad-net-v2"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  delete_on_termination = true
  billing_type          = "monthly"

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = ["00000000-0000-0000-0000-000000000000"]
  security_group_ids = ["00000000-0000-0000-0000-000000000000"]
}

# ─── Test E17: Instance with bad security group ────────────────
# Expected: "Failed to create instance"
resource "acecloud_instance" "error_bad_sg" {
  count = var.run_error_tests ? 1 : 0

  name        = "tf-error-test-bad-sg"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  delete_on_termination = true
  billing_type          = "monthly"

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = ["8f0b85d7-6517-4a80-8c32-9b7d93006d48"]
  security_group_ids = ["00000000-0000-0000-0000-000000000000"]
}

# ─── Test E18: Snapshot with bad volume ID ─────────────────────
# Expected: "Failed to create snapshot"
resource "acecloud_snapshot" "error_bad_volume" {
  count = var.run_error_tests ? 1 : 0

  name      = "tf-error-test-bad-snap"
  volume_id = "00000000-0000-0000-0000-000000000000"
}

# ─── Test E19: Volume backup with bad volume ID ────────────────
# Expected: "Failed to create volume backup"
resource "acecloud_volume_backup" "error_bad_volume" {
  count = var.run_error_tests ? 1 : 0

  name      = "tf-error-test-bad-backup"
  volume_id = "00000000-0000-0000-0000-000000000000"
}

# ─── Test E20: Router interface with bad subnet ────────────────
# Expected: "Failed to create router interface"
resource "acecloud_router_interface" "error_bad_subnet" {
  count = var.run_error_tests ? 1 : 0

  router_id = "00000000-0000-0000-0000-000000000000"
  subnet_id = "00000000-0000-0000-0000-000000000000"
}

# ─── Test E21: LB pool member with bad pool ID ────────────────
# Expected: "Failed to create pool member"
resource "acecloud_lb_pool_member" "error_bad_pool" {
  count = var.run_error_tests ? 1 : 0

  pool_id       = "00000000-0000-0000-0000-000000000000"
  name          = "tf-error-bad-member"
  address       = "10.0.0.1"
  protocol_port = 80
}

# ─── Test E22: Floating IP association with bad instance ───────
# Expected: "Failed to create floating IP association"
resource "acecloud_floating_ip_association" "error_bad_instance" {
  count = var.run_error_tests ? 1 : 0

  floating_ip_address = "1.2.3.4"
  instance_id         = "00000000-0000-0000-0000-000000000000"
}
