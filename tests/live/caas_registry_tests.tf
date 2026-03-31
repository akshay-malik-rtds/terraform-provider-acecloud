# ═══════════════════════════════════════════════════════════════
# CaaS & Registry In-Depth Tests
# ═══════════════════════════════════════════════════════════════
# Comprehensive tests for CaaS secrets, CaaS deployments,
# registry projects, and registry replication rules.
#
# Usage:
#   terraform apply -var="run_caas_registry_tests=true"
#   terraform plan  -var="run_caas_registry_tests=true"  # idempotency
#   terraform apply -var="run_caas_registry_tests=true" -var="cr_phase=update"
#   terraform plan  -var="run_caas_registry_tests=true" -var="cr_phase=update"  # post-update
#   terraform destroy -var="run_caas_registry_tests=true"

variable "run_caas_registry_tests" {
  description = "Set to true to run CaaS + Registry in-depth tests"
  type        = bool
  default     = false
}

variable "cr_phase" {
  description = "Phase: create or update"
  type        = string
  default     = "create"
}

# ═══════════════════════════════════════════════════════════════
# CAAS SECRETS
# ═══════════════════════════════════════════════════════════════

# CR1: Generic Secret — Data Map Lifecycle
# Create: 3 keys. Update: change 1 value, add 1 new key, remove 1 key.
resource "acecloud_caas_secret" "cr1_generic" {
  count = var.run_caas_registry_tests ? 1 : 0

  name = "tf-cr1-generic-secret"
  type = "generic"

  data = var.cr_phase == "update" ? {
    DB_HOST     = "db-v2.internal.example.com"
    DB_PORT     = "5432"
    DB_PASSWORD = "new-password-456"
  } : {
    DB_HOST = "db.internal.example.com"
    DB_PORT = "5432"
    API_KEY = "test-api-key-123"
  }
}

# CR2: Registry Secret — Skipped
# Registry-type secrets require a region-specific AceRegistry URL
# that the CaaS backend validates. Cannot test without knowing
# the exact URL. Generic secrets (CR1, CR3) cover secret lifecycle.

# CR3: Generic Secret — Minimal (single key)
# Tests smallest valid data map.
resource "acecloud_caas_secret" "cr3_minimal" {
  count = var.run_caas_registry_tests ? 1 : 0

  name = "tf-cr3-minimal-secret"
  type = "generic"

  data = {
    SINGLE_KEY = "single-value"
  }
}

# ═══════════════════════════════════════════════════════════════
# CAAS DEPLOYMENTS
# ═══════════════════════════════════════════════════════════════

# CR4: Shared Deployment — Public Image + Env + Port
# Update: scale replicas, change env value, change image tag.
resource "acecloud_caas_deployment" "cr4_shared" {
  count = var.run_caas_registry_tests ? 1 : 0

  name = "tf-cr4-shared-deploy"
  type = "shared"

  image {
    type      = "public"
    reference = var.cr_phase == "update" ? "nginx:1.27-alpine" : "nginx:alpine"
  }

  resources {
    cpu           = var.cr_phase == "update" ? 0.5 : 0.25
    memory        = var.cr_phase == "update" ? "512Mi" : "256Mi"
    replica_count = var.cr_phase == "update" ? 2 : 1
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
    value = var.cr_phase == "update" ? "staging" : "test"
  }

  env {
    name  = "LOG_LEVEL"
    value = "info"
  }
}

# CR5: Shared Deployment — Autoscaling + No External Access
# Tests internal service with autoscaling. Update: toggle autoscaling off.
resource "acecloud_caas_deployment" "cr5_autoscale" {
  count = var.run_caas_registry_tests ? 1 : 0

  name = "tf-cr5-autoscale-deploy"
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
    enabled                  = var.cr_phase == "update" ? false : true
    min_replicas             = var.cr_phase == "update" ? null : 1
    max_replicas             = var.cr_phase == "update" ? null : 3
    cpu_target_percentage    = var.cr_phase == "update" ? null : 70.0
    memory_target_percentage = var.cr_phase == "update" ? null : 80.0
  }
}

# CR6: Deployment with Volume + Multiple Ports + Command
# Tests volume mounts, multi-port config, and custom command.
resource "acecloud_caas_deployment" "cr6_full" {
  count = var.run_caas_registry_tests ? 1 : 0

  name = "tf-cr6-full-deploy"
  type = "shared"

  command = ["nginx", "-g", "daemon off;"]

  image {
    type      = "public"
    reference = "nginx:alpine"
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

    port {
      name           = "https"
      protocol       = "TCP"
      container_port = 443
      exposed_port   = 443
    }
  }

  autoscaling {
    enabled = false
  }

  env {
    name  = "NGINX_HOST"
    value = "localhost"
  }

  volume {
    name       = "config"
    mount_path = "/etc/nginx/conf.d"
    size       = "1Gi"
  }
}

# CR7: Deployment with env_secrets — References CR1 generic secret
# Tests secret injection into deployment.
resource "acecloud_caas_deployment" "cr7_with_secrets" {
  count = var.run_caas_registry_tests ? 1 : 0

  name        = "tf-cr7-secrets-deploy"
  type        = "shared"
  env_secrets = [acecloud_caas_secret.cr1_generic[0].name]

  image {
    type      = "public"
    reference = "busybox:latest"
  }

  resources {
    cpu           = 0.25
    memory        = "128Mi"
    replica_count = 1
  }

  networking {
    external_access = false
  }

  autoscaling {
    enabled = false
  }

  command = ["sh", "-c", "echo $DB_HOST && sleep 3600"]
}

# ═══════════════════════════════════════════════════════════════
# REGISTRY PROJECTS
# ═══════════════════════════════════════════════════════════════

# CR8: Registry Project — Scanning Off then On
# Update: toggle vulnerability_scanning true.
resource "acecloud_registry_project" "cr8_project" {
  count = var.run_caas_registry_tests ? 1 : 0

  registry_name          = "tf-cr8-registry"
  vulnerability_scanning = var.cr_phase == "update" ? true : false
}

# ═══════════════════════════════════════════════════════════════
# REGISTRY REPLICATION RULES
# ═══════════════════════════════════════════════════════════════

# CR9: Replication Rule — Event Based with Name Filter
# Update: disable rule, rename, add tag filter.
resource "acecloud_registry_replication_rule" "cr9_event" {
  count = var.run_caas_registry_tests ? 1 : 0

  name               = var.cr_phase == "update" ? "tf-cr9-rule-v2" : "tf-cr9-rule"
  enabled            = var.cr_phase == "update" ? false : true
  replicate_deletion = false
  override           = true
  speed              = -1

  src_registry {
    id   = 1
    name = "local-registry"
    url  = "https://registry.acecloud.ai"
  }

  trigger {
    type = "event_based"
  }

  dynamic "filter" {
    for_each = var.cr_phase == "update" ? [
      { type = "name", value = "tf-cr8-registry/**" },
      { type = "tag", value = "v*" },
    ] : [
      { type = "name", value = "tf-cr8-registry/**" },
    ]
    content {
      type  = filter.value.type
      value = filter.value.value
    }
  }

  depends_on = [acecloud_registry_project.cr8_project]
}

# CR10: Replication Rule — Scheduled Trigger
# Tests: scheduled trigger with cron, name + tag filters.
resource "acecloud_registry_replication_rule" "cr10_scheduled" {
  count = var.run_caas_registry_tests ? 1 : 0

  name    = "tf-cr10-sched-rule"
  enabled = true

  src_registry {
    id   = 1
    name = "local-registry"
    url  = "https://registry.acecloud.ai"
  }

  trigger {
    type = "scheduled"
    cron = "0 0 * * *"
  }

  filter {
    type  = "name"
    value = "tf-cr8-registry/**"
  }

  filter {
    type  = "tag"
    value = "release-*"
  }

  depends_on = [acecloud_registry_project.cr8_project]
}

# CR11: Replication Rule — Minimal (no optional fields)
# Tests default values for replicate_deletion, override, speed.
resource "acecloud_registry_replication_rule" "cr11_minimal" {
  count = var.run_caas_registry_tests ? 1 : 0

  name    = "tf-cr11-minimal-rule"
  enabled = true

  src_registry {
    id   = 1
    name = "local-registry"
    url  = "https://registry.acecloud.ai"
  }

  trigger {
    type = "manual"
  }

  # At least one filter is required by the API
  filter {
    type  = "name"
    value = "**"
  }

  depends_on = [acecloud_registry_project.cr8_project]
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

# CaaS Secrets
output "cr1_secret_status" {
  value = var.run_caas_registry_tests ? acecloud_caas_secret.cr1_generic[0].status : ""
}

output "cr2_secret_status" {
  value = "skipped - registry type needs region-specific URL"
}

output "cr3_secret_status" {
  value = var.run_caas_registry_tests ? acecloud_caas_secret.cr3_minimal[0].status : ""
}

# CaaS Deployments
output "cr4_deploy_status" {
  value = var.run_caas_registry_tests ? acecloud_caas_deployment.cr4_shared[0].status : ""
}

output "cr4_deploy_public_endpoints" {
  value = var.run_caas_registry_tests ? acecloud_caas_deployment.cr4_shared[0].public_endpoints : []
}

output "cr5_deploy_status" {
  value = var.run_caas_registry_tests ? acecloud_caas_deployment.cr5_autoscale[0].status : ""
}

output "cr6_deploy_status" {
  value = var.run_caas_registry_tests ? acecloud_caas_deployment.cr6_full[0].status : ""
}

output "cr7_deploy_status" {
  value = var.run_caas_registry_tests ? acecloud_caas_deployment.cr7_with_secrets[0].status : ""
}

# Registry
output "cr8_project_id" {
  value = var.run_caas_registry_tests ? acecloud_registry_project.cr8_project[0].id : ""
}

output "cr9_rule_id" {
  value = var.run_caas_registry_tests ? acecloud_registry_replication_rule.cr9_event[0].id : ""
}

output "cr10_rule_id" {
  value = var.run_caas_registry_tests ? acecloud_registry_replication_rule.cr10_scheduled[0].id : ""
}

output "cr11_rule_id" {
  value = var.run_caas_registry_tests ? acecloud_registry_replication_rule.cr11_minimal[0].id : ""
}
