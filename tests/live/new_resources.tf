# ═══════════════════════════════════════════════════════════════
# Phase 14: Auto Scaling Template
# ═══════════════════════════════════════════════════════════════
# Uses existing VPC, subnet, security group, key pair, and image from Phase 2-4.

resource "acecloud_auto_scaling_template" "test" {
  count = var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                   = "tf-live-test-as-template"
  type                   = "linux"
  description            = "Terraform live test, auto scaling template"
  volume_size            = 40
  vol_del_on_termination = true
  flavor_id              = var.flavor_id
  image_id               = var.image_id
  is_instance_snapshot   = false
  key_name               = acecloud_key_pair.generated.name
  network_id             = acecloud_vpc.test.id
  subnet_id              = acecloud_vpc.test.subnet_id
  security_groups        = [acecloud_security_group.web.id]
}

# ═══════════════════════════════════════════════════════════════
# Phase 15: CaaS Secret (generic type)
# ═══════════════════════════════════════════════════════════════
# Gated by var.run_caas_tests — CaaS backend on dev4 has known issues

variable "run_caas_tests" {
  description = "Set to true to create CaaS resources (may fail on dev4)"
  type        = bool
  default     = false
}

resource "acecloud_caas_secret" "test-generic" {
  count = var.run_caas_tests ? 1 : 0

  name = "tf-test-secret-v5"
  type = "generic"

  data = {
    TEST_KEY   = "test-value"
    ANOTHER_KEY = "another-value"
  }
}

# ═══════════════════════════════════════════════════════════════
# Phase 16: CaaS Deployment (shared, public image)
# ═══════════════════════════════════════════════════════════════

resource "acecloud_caas_deployment" "test" {
  count = var.run_caas_tests ? 1 : 0

  name = "tf-live-test-deploy"
  type = "shared"

  image {
    type      = "public"
    reference = "nginx:latest"
  }

  resources {
    cpu           = 0.25
    memory        = "256Mi"
    replica_count = 1
  }

  networking {
    external_access = true
    endpoint_access = "public"

    port {
      name           = "http"
      protocol       = "TCP"
      container_port = 80
      exposed_port   = 80
    }
  }

  autoscaling {
    enabled = false
  }

  env {
    name  = "APP_ENV"
    value = "test"
  }
}

# ═══════════════════════════════════════════════════════════════
# Phase 16b: CaaS Deployment — Dedicated type with flavor
# ═══════════════════════════════════════════════════════════════
# Tests dedicated deployment type (uses flavor_id instead of cpu/memory)
# NOTE: Currently blocked on dev4 — API returns 500 "Unable to process"

resource "acecloud_caas_deployment" "test_dedicated" {
  count = var.run_caas_tests && var.flavor_id != "" ? 1 : 0

  name = "tf-live-test-deploy-dedicated"
  type = "dedicated"

  image {
    type      = "public"
    reference = "nginx:alpine"
  }

  resources {
    flavor_id     = var.flavor_id
    replica_count = 1
  }

  networking {
    external_access = true
    endpoint_access = "public"

    port {
      name           = "http"
      protocol       = "TCP"
      container_port = 80
      exposed_port   = 8081
    }
  }

  autoscaling {
    enabled = false
  }
}

# ═══════════════════════════════════════════════════════════════
# Phase 16c: CaaS Deployment — Shared with autoscaling + env + volume
# ═══════════════════════════════════════════════════════════════
# Tests more complete shared deployment with autoscaling, multiple envs, volume

resource "acecloud_caas_deployment" "test_shared_full" {
  count = var.run_caas_tests ? 1 : 0

  name = "tf-live-test-deploy-shared-full"
  type = "shared"

  image {
    type      = "public"
    reference = "redis:7-alpine"
  }

  resources {
    cpu           = 0.5
    memory        = "512Mi"
    replica_count = 1
  }

  networking {
    external_access = false
  }

  autoscaling {
    enabled                  = true
    min_replicas             = 1
    max_replicas             = 3
    cpu_target_percentage    = 70.0
    memory_target_percentage = 80.0
  }

  env {
    name  = "REDIS_MAXMEMORY"
    value = "256mb"
  }

  env {
    name  = "REDIS_PASSWORD"
    value = "test-password-123"
  }

  volume {
    name       = "data"
    mount_path = "/data"
    size       = "1Gi"
  }
}

# ═══════════════════════════════════════════════════════════════
# Phase 16d: Auto Scaling Deployment (uses template from Phase 14)
# ═══════════════════════════════════════════════════════════════
# Tests auto scaling deployment with the template created in Phase 14.
# NOTE: Requires auto_scaling_template.test to exist + be active.

resource "acecloud_auto_scaling_deployment" "test" {
  count = var.run_caas_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                  = "tf-live-test-as-deploy"
  description           = "Terraform live test auto scaling deployment"
  template_id           = acecloud_auto_scaling_template.test[0].id
  desired_capacity      = 1
  max_capacity          = 2
  nodes_scale_count     = 1
  scaling_parameter     = "cpu"
  min_threshold         = 40
  max_threshold         = 80
  cool_down_time        = 120
  user_email            = ["test@example.com"]
  is_integrated_with_lb = false

  depends_on = [acecloud_auto_scaling_template.test]
}

# ═══════════════════════════════════════════════════════════════
# Phase 16e: Auto Scaling Deployment with LB integration
# ═══════════════════════════════════════════════════════════════

resource "acecloud_auto_scaling_deployment" "test_with_lb" {
  count = var.run_caas_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                  = "tf-live-test-as-deploy-lb"
  description           = "Auto scaling deployment with LB integration"
  template_id           = acecloud_auto_scaling_template.test[0].id
  desired_capacity      = 1
  max_capacity          = 3
  nodes_scale_count     = 1
  scaling_parameter     = "ram"
  min_threshold         = 30
  max_threshold         = 70
  cool_down_time        = 180
  user_email            = ["test@example.com", "ops@example.com"]
  is_integrated_with_lb = true

  lb_data {
    lb_name         = "tf-as-lb"
    tags            = ["ALB"]
    assign_public_ip = false
    is_existing_lb  = false

    listener {
      listener_name          = "tf-as-listener"
      listener_protocol      = "HTTP"
      listener_protocol_port = 80
    }

    pool {
      pool_protocol      = "HTTP"
      pool_protocol_port = 8080
      lb_algorithm       = "ROUND_ROBIN"
    }

    health_monitor {
      monitor_protocol    = "HTTP"
      monitor_url_path    = "/health"
      monitor_http_method = "GET"
    }
  }

  depends_on = [acecloud_auto_scaling_template.test]
}

# ═══════════════════════════════════════════════════════════════
# Phase 17: Kubernetes Cluster (long-running, 20min)
# ═══════════════════════════════════════════════════════════════
# Gated by var.run_k8s_tests (default false) — K8s clusters are expensive
# and take ~20 minutes to provision.

variable "run_k8s_tests" {
  description = "Set to true to create a K8s cluster (takes ~20 min)"
  type        = bool
  default     = false
}

resource "acecloud_k8s_cluster" "test" {
  count = var.run_k8s_tests && var.flavor_id != "" ? 1 : 0

  name                  = "tf-live-test-cluster"
  kubernetes_version    = "v1.32.8+rke2r1"
  endpoint_access       = "Public and Private"
  network_isolation     = "Disabled"
  nginx_ingress         = "Enabled"
  nginx_default_backend = "Enabled"
  network_provider      = "Calico"
  snapshot_backup       = "No"
  secrets_encryption    = "Disabled"
  max_worker_nodes      = 3
  autoscale             = false
  autoscale_min         = 0
  autoscale_max         = 0
  worker_node_name      = "tf-worker"
  worker_quantity       = 1
  flavor_id             = var.flavor_id
  flavor_name           = "C4i.medium"
  volume_size           = 40
  # No VPC, SG, or key pair needed — K8s creates its own infra in a separate project
}

# ═══════════════════════════════════════════════════════════════
# Phase 18: Registry Project + Replication Rule
# ═══════════════════════════════════════════════════════════════

variable "run_registry_tests" {
  description = "Set to true to create registry resources"
  type        = bool
  default     = false
}

resource "acecloud_registry_project" "test" {
  count = var.run_registry_tests ? 1 : 0

  registry_name          = "tf-live-test-registry"
  vulnerability_scanning = false
}

resource "acecloud_registry_replication_rule" "test" {
  count = var.run_registry_tests ? 1 : 0

  name    = "tf-live-test-repl-rule"
  enabled = true

  src_registry {
    id   = 1
    name = "local-registry"
    url  = "https://registry.acecloud.ai"
    type = "harbor"
  }

  trigger {
    type = "event_based"
  }

  filter {
    type  = "name"
    value = "tf-live-test-registry-v2/**"
  }

  depends_on = [acecloud_registry_project.test]
}

# ═══════════════════════════════════════════════════════════════
# Outputs for new resources
# ═══════════════════════════════════════════════════════════════

output "auto_scaling_template_id" {
  value = var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_template.test[0].id : "skipped"
}

output "auto_scaling_template_status" {
  value = var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_template.test[0].status : "skipped"
}

output "caas_secret_id" {
  value = var.run_caas_tests ? acecloud_caas_secret.test-generic[0].id : "skipped"
}

output "caas_secret_status" {
  value = var.run_caas_tests ? acecloud_caas_secret.test-generic[0].status : "skipped"
}

output "caas_deployment_id" {
  value = var.run_caas_tests ? acecloud_caas_deployment.test[0].id : "skipped"
}

output "caas_deployment_status" {
  value = var.run_caas_tests ? acecloud_caas_deployment.test[0].status : "skipped"
}

output "registry_project_id" {
  value = var.run_registry_tests ? acecloud_registry_project.test[0].id : "skipped"
}

output "registry_replication_rule_id" {
  value = var.run_registry_tests ? acecloud_registry_replication_rule.test[0].id : "skipped"
}

output "caas_deployment_dedicated_id" {
  value = var.run_caas_tests && var.flavor_id != "" ? acecloud_caas_deployment.test_dedicated[0].id : "skipped"
}

output "caas_deployment_shared_full_id" {
  value = var.run_caas_tests ? acecloud_caas_deployment.test_shared_full[0].id : "skipped"
}

output "as_deployment_id" {
  value = var.run_caas_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_deployment.test[0].id : "skipped"
}

output "as_deployment_with_lb_id" {
  value = var.run_caas_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_deployment.test_with_lb[0].id : "skipped"
}

output "k8s_cluster_id" {
  value = var.run_k8s_tests ? acecloud_k8s_cluster.test[0].id : "skipped"
}

output "k8s_cluster_status" {
  value = var.run_k8s_tests ? acecloud_k8s_cluster.test[0].status : "skipped"
}
