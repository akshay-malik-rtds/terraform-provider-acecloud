# ═══════════════════════════════════════════════════════════════
# User Workflow Tests — Real-world operations users perform daily
# ═══════════════════════════════════════════════════════════════
# These test multi-step workflows, update paths, and edge cases
# that real users encounter. Gate: var.run_user_tests (default false)

variable "run_user_tests" {
  description = "Set to true to run user workflow tests"
  type        = bool
  default     = false
}

variable "user_test_phase" {
  description = "Phase: 'create' for initial resources, 'update' to modify them"
  type        = string
  default     = "create"

  validation {
    condition     = contains(["create", "update"], var.user_test_phase)
    error_message = "user_test_phase must be 'create' or 'update'."
  }
}

# ═══════════════════════════════════════════════════════════════
# UW1: Volume resize (grow) — most common storage operation
# ═══════════════════════════════════════════════════════════════

resource "acecloud_volume" "uw1_resize" {
  count = var.run_user_tests ? 1 : 0

  name         = var.user_test_phase == "update" ? "tf-uw1-vol-renamed" : "tf-uw1-vol-resize"
  size         = var.user_test_phase == "update" ? 20 : 10
  volume_type  = "ssd"
  billing_type = "hourly"
  description  = var.user_test_phase == "update" ? "UW1 resized to 20GB and renamed" : "UW1 initial 10GB volume"
}

# ═══════════════════════════════════════════════════════════════
# UW2: Security group rule changes — daily ops workflow
# ═══════════════════════════════════════════════════════════════
# Create phase: SSH + HTTP + HTTPS (3 rules)
# Update phase: SSH only (remove HTTP + HTTPS, tests rule removal)

resource "acecloud_security_group" "uw2_rules" {
  count = var.run_user_tests ? 1 : 0

  name        = var.user_test_phase == "update" ? "tf-uw2-sg-updated" : "tf-uw2-sg-rules"
  description = var.user_test_phase == "update" ? "UW2 updated, SSH only" : "UW2 initial, SSH plus HTTP plus HTTPS"

  # SSH — always present
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = var.user_test_phase == "update" ? "10.0.0.0/8" : "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # HTTP — only in create phase (removed on update)
  dynamic "rules" {
    for_each = var.user_test_phase == "create" ? [1] : []
    content {
      direction        = "ingress"
      protocol         = "tcp"
      port_range_min   = 80
      port_range_max   = 80
      remote_ip_prefix = "0.0.0.0/0"
      ethertype        = "IPv4"
    }
  }

  # HTTPS — only in create phase (removed on update)
  dynamic "rules" {
    for_each = var.user_test_phase == "create" ? [1] : []
    content {
      direction        = "ingress"
      protocol         = "tcp"
      port_range_min   = 443
      port_range_max   = 443
      remote_ip_prefix = "0.0.0.0/0"
      ethertype        = "IPv4"
    }
  }
}

# ═══════════════════════════════════════════════════════════════
# UW3: VPC + Subnet updates — rename, DNS changes
# ═══════════════════════════════════════════════════════════════

resource "acecloud_vpc" "uw3_vpc" {
  count = var.run_user_tests ? 1 : 0

  name                   = var.user_test_phase == "update" ? "tf-uw3-vpc-renamed" : "tf-uw3-vpc"
  admin_state_up         = true
  subnet_name            = "tf-uw3-subnet"
  subnet_cidr            = "10.70.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_gateway_ip      = "10.70.0.1"
  subnet_dns_nameservers = var.user_test_phase == "update" ? ["1.1.1.1", "8.8.8.8"] : ["8.8.8.8", "8.8.4.4"]
}

resource "acecloud_subnet" "uw3_extra" {
  count = var.run_user_tests ? 1 : 0

  name            = var.user_test_phase == "update" ? "tf-uw3-subnet2-renamed" : "tf-uw3-subnet2"
  vpc_id          = acecloud_vpc.uw3_vpc[0].id
  cidr            = "10.70.1.0/24"
  ip_version      = 4
  enable_dhcp     = true
  gateway_ip      = "10.70.1.1"
  dns_nameservers = var.user_test_phase == "update" ? ["1.1.1.1"] : ["8.8.8.8"]
  description     = var.user_test_phase == "update" ? "UW3 subnet updated" : "UW3 extra subnet"

  allocation_pools {
    start = "10.70.1.50"
    end   = "10.70.1.200"
  }

  depends_on = [acecloud_vpc.uw3_vpc]
}

# ═══════════════════════════════════════════════════════════════
# UW4: Router with gateway — add interface, update name
# ═══════════════════════════════════════════════════════════════

resource "acecloud_router" "uw4_router" {
  count = var.run_user_tests && var.external_network_id != "" ? 1 : 0

  name                        = var.user_test_phase == "update" ? "tf-uw4-router-renamed" : "tf-uw4-router"
  admin_state_up              = true
  external_gateway_network_id = var.external_network_id
}

resource "acecloud_router_interface" "uw4_ri" {
  count = var.run_user_tests && var.external_network_id != "" ? 1 : 0

  router_id = acecloud_router.uw4_router[0].id
  subnet_id = acecloud_vpc.uw3_vpc[0].subnet_id

  depends_on = [acecloud_router.uw4_router, acecloud_vpc.uw3_vpc]
}

# ═══════════════════════════════════════════════════════════════
# UW5: LB update workflow — change name, description, tags
# ═══════════════════════════════════════════════════════════════

resource "acecloud_load_balancer" "uw5_lb" {
  count = var.run_user_tests ? 1 : 0

  name        = var.user_test_phase == "update" ? "tf-uw5-lb-renamed" : "tf-uw5-lb"
  description = var.user_test_phase == "update" ? "UW5 LB updated description" : "UW5 initial LB"
  subnet_id   = acecloud_vpc.uw3_vpc[0].subnet_id
  tags        = var.user_test_phase == "update" ? ["ALB", "production", "updated"] : ["ALB", "staging"]
}

resource "acecloud_lb_listener" "uw5_listener" {
  count = var.run_user_tests ? 1 : 0

  name            = "tf-uw5-listener"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.uw5_lb[0].id
}

resource "acecloud_lb_pool" "uw5_pool" {
  count = var.run_user_tests ? 1 : 0

  name            = "tf-uw5-pool"
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
  listener_id     = acecloud_lb_listener.uw5_listener[0].id
  loadbalancer_id = acecloud_load_balancer.uw5_lb[0].id

  depends_on = [acecloud_lb_listener.uw5_listener]
}

# Members with weight changes on update
resource "acecloud_lb_pool_member" "uw5_m1" {
  count = var.run_user_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.uw5_pool[0].id
  name          = "tf-uw5-member-1"
  address       = "10.70.0.10"
  protocol_port = 8080
  weight        = var.user_test_phase == "update" ? 10 : 5

  depends_on = [acecloud_lb_pool.uw5_pool]
}

resource "acecloud_lb_pool_member" "uw5_m2" {
  count = var.run_user_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.uw5_pool[0].id
  name          = "tf-uw5-member-2"
  address       = "10.70.0.11"
  protocol_port = 8080
  weight        = var.user_test_phase == "update" ? 1 : 5

  depends_on = [acecloud_lb_pool_member.uw5_m1]
}

resource "acecloud_lb_health_monitor" "uw5_hm" {
  count = var.run_user_tests ? 1 : 0

  name           = "tf-uw5-hm"
  pool_id        = acecloud_lb_pool.uw5_pool[0].id
  type           = "HTTP"
  delay          = var.user_test_phase == "update" ? 15 : 10
  timeout        = var.user_test_phase == "update" ? 8 : 5
  max_retries    = var.user_test_phase == "update" ? 5 : 3
  url_path       = var.user_test_phase == "update" ? "/healthz" : "/health"
  expected_codes = var.user_test_phase == "update" ? "200-299" : "200"
  http_method    = "GET"

  depends_on = [acecloud_lb_pool_member.uw5_m2]
}

# ═══════════════════════════════════════════════════════════════
# UW6: Snapshot + Backup update (name, description)
# ═══════════════════════════════════════════════════════════════

resource "acecloud_volume" "uw6_vol" {
  count = var.run_user_tests ? 1 : 0

  name        = "tf-uw6-vol"
  size        = 10
  volume_type = "ssd"
}

resource "acecloud_snapshot" "uw6_snap" {
  count = var.run_user_tests ? 1 : 0

  name        = var.user_test_phase == "update" ? "tf-uw6-snap-renamed" : "tf-uw6-snap"
  volume_id   = acecloud_volume.uw6_vol[0].id
  description = var.user_test_phase == "update" ? "UW6 snapshot renamed" : "UW6 initial snapshot"

  depends_on = [acecloud_volume.uw6_vol]
}

resource "acecloud_volume_backup" "uw6_backup" {
  count = var.run_user_tests ? 1 : 0

  name        = "tf-uw6-backup"
  volume_id   = acecloud_volume.uw6_vol[0].id
  description = "UW6 backup for lifecycle testing"

  depends_on = [acecloud_snapshot.uw6_snap]
}

# ═══════════════════════════════════════════════════════════════
# UW7: Instance on user VPC with FIP — full production setup
# ═══════════════════════════════════════════════════════════════

resource "acecloud_instance" "uw7_prod" {
  count = var.run_user_tests && var.flavor_id != "" && var.image_id != "" && var.external_network_id != "" ? 1 : 0

  name        = var.user_test_phase == "update" ? "tf-uw7-prod-renamed" : "tf-uw7-prod"
  description = "UW7 production-like instance"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  billing_type          = "monthly"
  delete_on_termination = true

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.uw3_vpc[0].id]
  security_group_ids = [acecloud_security_group.uw2_rules[0].id]
  key_name           = acecloud_key_pair.generated.name

  metadata = {
    managed_by  = "terraform"
    environment = "user-workflow-test"
  }

  depends_on = [acecloud_router_interface.uw4_ri]
}

resource "acecloud_floating_ip" "uw7_fip" {
  count = var.run_user_tests && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  description         = "UW7 production floating IP"
}

resource "acecloud_floating_ip_association" "uw7_assoc" {
  count = var.run_user_tests && var.flavor_id != "" && var.image_id != "" && var.external_network_id != "" ? 1 : 0

  floating_ip_address = acecloud_floating_ip.uw7_fip[0].floating_ip_address
  instance_id         = acecloud_instance.uw7_prod[0].id

  depends_on = [acecloud_floating_ip.uw7_fip, acecloud_instance.uw7_prod]
}

# ═══════════════════════════════════════════════════════════════
# UW8: Auto Scaling Template update — name, description, vol settings
# ═══════════════════════════════════════════════════════════════

resource "acecloud_auto_scaling_template" "uw8_template" {
  count = var.run_user_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                   = var.user_test_phase == "update" ? "tf-uw8-template-updated" : "tf-uw8-template"
  type                   = "linux"
  description            = var.user_test_phase == "update" ? "UW8 template updated" : "UW8 auto scaling template"
  volume_size            = var.user_test_phase == "update" ? 50 : 40
  vol_del_on_termination = var.user_test_phase == "update" ? false : true
  flavor_id              = var.flavor_id
  image_id               = var.image_id
  is_instance_snapshot   = false
  key_name               = acecloud_key_pair.generated.name
  network_id             = acecloud_vpc.uw3_vpc[0].id
  subnet_id              = acecloud_vpc.uw3_vpc[0].subnet_id
  security_groups        = [acecloud_security_group.uw2_rules[0].id]
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "uw1_vol_id" {
  value = var.run_user_tests ? acecloud_volume.uw1_resize[0].id : ""
}
output "uw1_vol_size" {
  value = var.run_user_tests ? acecloud_volume.uw1_resize[0].size : 0
}
output "uw2_sg_id" {
  value = var.run_user_tests ? acecloud_security_group.uw2_rules[0].id : ""
}
output "uw3_vpc_id" {
  value = var.run_user_tests ? acecloud_vpc.uw3_vpc[0].id : ""
}
output "uw3_subnet_id" {
  value = var.run_user_tests ? acecloud_subnet.uw3_extra[0].id : ""
}
output "uw4_router_id" {
  value = var.run_user_tests && var.external_network_id != "" ? acecloud_router.uw4_router[0].id : ""
}
output "uw5_lb_id" {
  value = var.run_user_tests ? acecloud_load_balancer.uw5_lb[0].id : ""
}
output "uw6_snap_id" {
  value = var.run_user_tests ? acecloud_snapshot.uw6_snap[0].id : ""
}
output "uw6_backup_id" {
  value = var.run_user_tests ? acecloud_volume_backup.uw6_backup[0].id : ""
}
output "uw7_instance_id" {
  value = var.run_user_tests && var.flavor_id != "" && var.image_id != "" && var.external_network_id != "" ? acecloud_instance.uw7_prod[0].id : ""
}
output "uw7_fip_address" {
  value = var.run_user_tests && var.external_network_id != "" ? acecloud_floating_ip.uw7_fip[0].floating_ip_address : ""
}
output "uw8_template_id" {
  value = var.run_user_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_auto_scaling_template.uw8_template[0].id : ""
}
