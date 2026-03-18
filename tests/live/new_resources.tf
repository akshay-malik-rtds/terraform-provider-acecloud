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

  name = "tf-test-secret-v3"
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
  count = var.run_k8s_tests ? 1 : 0

  name                  = "tf-live-test-cluster"
  kubernetes_version    = "v1.32.6"
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
  volume_size           = 40
  sec_group_id          = acecloud_security_group.web.id
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

output "k8s_cluster_id" {
  value = var.run_k8s_tests ? acecloud_k8s_cluster.test[0].id : "skipped"
}

output "k8s_cluster_status" {
  value = var.run_k8s_tests ? acecloud_k8s_cluster.test[0].status : "skipped"
}
