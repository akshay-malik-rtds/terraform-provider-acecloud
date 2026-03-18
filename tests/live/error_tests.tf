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

# ─── Test E10: CaaS Secret — Registry type missing URL ───────
# Expected: "Failed to create CaaS secret" with sanitized message
resource "acecloud_caas_secret" "error_bad_registry" {
  count = var.run_error_tests ? 1 : 0

  name     = "tf-error-test-bad-secret"
  type     = "registry"
  url      = "https://nonexistent-registry.invalid.example.com"
  username = "baduser"
  password = "badpass"
}

# ─── Test E11: CaaS Deployment — Bad image reference ─────────
# Expected: "Failed to create CaaS deployment" with sanitized message
resource "acecloud_caas_deployment" "error_bad_image" {
  count = var.run_error_tests ? 1 : 0

  name = "tf-error-test-bad-caas"

  image {
    type      = "public"
    reference = "nonexistent-image-xyz:v999.999.999"
  }

  resources {
    replica_count = 1
  }

  networking {
    external_access = false
  }

  autoscaling {
    enabled = false
  }
}

# ─── Test E12: K8s Cluster — Bad flavor ID ───────────────────
# Expected: "Failed to create Kubernetes cluster" with sanitized message
resource "acecloud_k8s_cluster" "error_bad_flavor" {
  count = var.run_error_tests ? 1 : 0

  name                  = "tf-error-test-bad-k8s"
  kubernetes_version    = "v1.32.6+rke2r1"
  endpoint_access       = "Public"
  network_isolation     = "Disabled"
  nginx_ingress         = "Enabled"
  nginx_default_backend = "Enabled"
  network_provider      = "Calico"
  secrets_encryption    = "Disabled"
  max_worker_nodes      = 3
  worker_node_name      = "tf-error-worker"
  worker_quantity       = 1
  flavor_id             = "00000000-0000-0000-0000-000000000000"
  volume_size           = 50
}

# ─── Test E13: K8s Node Group — Bad cluster ID ───────────────
# Expected: "Failed to create Kubernetes node group" with sanitized message
resource "acecloud_k8s_node_group" "error_bad_cluster" {
  count = var.run_error_tests ? 1 : 0

  cluster_id   = "00000000-0000-0000-0000-000000000000"
  sec_group_id = "00000000-0000-0000-0000-000000000001"
  name         = "tf-error-test-bad-ng"
  quantity     = 1
  flavor_id    = var.flavor_id
  volume       = "50"
}

# ─── Test E14: Registry Project — Duplicate name ─────────────
# Expected: "Failed to create registry project" with sanitized message
# Note: Uses obviously invalid name to trigger 400 from Harbor backend
resource "acecloud_registry_project" "error_bad_name" {
  count = var.run_error_tests ? 1 : 0

  registry_name = "!!invalid..project..name!!"
}
