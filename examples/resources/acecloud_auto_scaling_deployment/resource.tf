# Auto scaling deployment with embedded load balancer integration.
# AS deployments are async; provisioning typically takes 10+ minutes.

resource "acecloud_auto_scaling_deployment" "web" {
  name                  = "web-tier-deployment"
  description           = "Auto scaling deployment for web tier"
  template_id           = acecloud_auto_scaling_template.web.id
  desired_capacity      = 2
  max_capacity          = 5
  nodes_scale_count     = 1
  scaling_parameter     = "cpu"
  min_threshold         = 30
  max_threshold         = 70
  cool_down_time        = 180
  user_email            = ["ops@example.com"]
  is_integrated_with_lb = true

  lb_data {
    lb_name          = "web-tier-lb"
    tags             = ["ALB"]
    assign_public_ip = false
    is_existing_lb   = false

    listener {
      listener_name          = "web-listener"
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

  depends_on = [acecloud_auto_scaling_template.web]
}
