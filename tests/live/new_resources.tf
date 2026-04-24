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
# Phase 16d: Auto Scaling Deployment (uses template from Phase 14)
# ═══════════════════════════════════════════════════════════════
# Tests auto scaling deployment with the template created in Phase 14.
# NOTE: Requires auto_scaling_template.test to exist + be active.

variable "run_as_deployment_tests" {
  description = "Set to true to create the Auto Scaling deployment (30+ min async)"
  type        = bool
  default     = false
}

resource "acecloud_auto_scaling_deployment" "test" {
  count = var.run_as_deployment_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

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
  count = var.run_as_deployment_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

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
    lb_name          = "tf-as-lb"
    tags             = ["ALB"]
    assign_public_ip = false
    is_existing_lb   = false

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
# Outputs
# ═══════════════════════════════════════════════════════════════

output "auto_scaling_template_id" {
  value = var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_template.test[0].id : "skipped"
}

output "auto_scaling_template_status" {
  value = var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_template.test[0].status : "skipped"
}

output "as_deployment_id" {
  value = var.run_as_deployment_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_deployment.test[0].id : "skipped"
}

output "as_deployment_with_lb_id" {
  value = var.run_as_deployment_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_deployment.test_with_lb[0].id : "skipped"
}
