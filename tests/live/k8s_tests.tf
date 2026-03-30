# ═══════════════════════════════════════════════════════════════
# K8s User Scenario Tests
# ═══════════════════════════════════════════════════════════════
# Tests K8s cluster creation and node group operations.
#
# WARNING: K8s cluster creation takes ~25 minutes.
#
# Usage:
#   terraform apply -var="run_k8s_scenario_tests=true"
#   terraform plan  -var="run_k8s_scenario_tests=true"  # idempotency
#   terraform apply -var="run_k8s_scenario_tests=true" -var="k8s_scenario_phase=scale"
#   terraform destroy -var="run_k8s_scenario_tests=true"

variable "run_k8s_scenario_tests" {
  description = "Set to true to run K8s scenario tests (takes ~25 min for cluster)"
  type        = bool
  default     = false
}

variable "k8s_scenario_phase" {
  description = "Phase: create or scale"
  type        = string
  default     = "create"
}

# ═══════════════════════════════════════════════════════════════
# K8S-1: Cluster with Autoscaling
# ═══════════════════════════════════════════════════════════════

resource "acecloud_k8s_cluster" "scenario" {
  count = var.run_k8s_scenario_tests && var.flavor_id != "" ? 1 : 0

  name                  = "tf-k8s-scenario"
  kubernetes_version    = "v1.32.8+rke2r1"
  endpoint_access       = "Public and Private"
  network_isolation     = "Disabled"
  nginx_ingress         = "Enabled"
  nginx_default_backend = "Enabled"
  network_provider      = "Calico"
  snapshot_backup       = "No"
  secrets_encryption    = "Disabled"
  max_worker_nodes      = 5
  autoscale             = true
  autoscale_min         = 1
  autoscale_max         = 5
  worker_node_name      = "default-pool"
  worker_quantity       = 1
  flavor_id             = var.flavor_id
  flavor_name           = "C4i.medium"
  volume_size           = 40
}

# ═══════════════════════════════════════════════════════════════
# K8S-2: Node Group — Create + Scale
# ═══════════════════════════════════════════════════════════════
# Create a second node group on the cluster.
# Scale phase: increase quantity from 1 to 2.

resource "acecloud_security_group" "k8s_ng_sg" {
  count = var.run_k8s_scenario_tests && var.flavor_id != "" ? 1 : 0
  name  = "tf-k8s-ng-sg"
}

resource "acecloud_k8s_node_group" "scenario_pool" {
  count = var.run_k8s_scenario_tests && var.flavor_id != "" ? 1 : 0

  cluster_id   = acecloud_k8s_cluster.scenario[0].id
  name         = "tf-pool-v3"
  quantity     = var.k8s_scenario_phase == "scale" ? 2 : 1
  flavor_id    = var.flavor_id
  volume       = "40"
  sec_group_id = acecloud_security_group.k8s_ng_sg[0].id

  labels = {
    "workload-type" = "general"
    "managed-by"    = "terraform"
  }

  min_node = 1
  max_node = 4
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "k8s_scenario_cluster_id" {
  value = var.run_k8s_scenario_tests && var.flavor_id != "" ? acecloud_k8s_cluster.scenario[0].id : ""
}

output "k8s_scenario_cluster_status" {
  value = var.run_k8s_scenario_tests && var.flavor_id != "" ? acecloud_k8s_cluster.scenario[0].status : ""
}

output "k8s_scenario_node_group_id" {
  value = var.run_k8s_scenario_tests && var.flavor_id != "" ? acecloud_k8s_node_group.scenario_pool[0].id : ""
}

output "k8s_scenario_node_group_state" {
  value = var.run_k8s_scenario_tests && var.flavor_id != "" ? acecloud_k8s_node_group.scenario_pool[0].state : ""
}

output "k8s_scenario_node_group_qty" {
  value = var.run_k8s_scenario_tests && var.flavor_id != "" ? acecloud_k8s_node_group.scenario_pool[0].quantity : 0
}
