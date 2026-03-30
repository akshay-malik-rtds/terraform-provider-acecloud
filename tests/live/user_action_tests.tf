# ═══════════════════════════════════════════════════════════════
# User Action Tests — Real-World Day-2 Operations
# ═══════════════════════════════════════════════════════════════
# Tests actual user workflows: scaling, failover, rotation,
# multi-resource updates, and backup-restore chains.
#
# Usage:
#   terraform apply -var="run_action_tests=true"
#   terraform plan  -var="run_action_tests=true"  # idempotency
#   terraform apply -var="run_action_tests=true" -var="action_phase=update"
#   terraform plan  -var="run_action_tests=true" -var="action_phase=update"
#   terraform destroy -var="run_action_tests=true"

variable "run_action_tests" {
  description = "Set to true to run user action tests"
  type        = bool
  default     = false
}

variable "action_phase" {
  description = "Phase: create or update"
  type        = string
  default     = "create"
}

# ═══════════════════════════════════════════════════════════════
# Shared Infrastructure
# ═══════════════════════════════════════════════════════════════

resource "acecloud_vpc" "ua_vpc" {
  count = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                   = "tf-ua-vpc"
  admin_state_up         = true
  subnet_name            = "tf-ua-subnet"
  subnet_cidr            = "10.92.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_dns_nameservers = ["8.8.8.8"]
}

# Two key pairs — for rotation test
resource "acecloud_key_pair" "ua_key_a" {
  count = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0
  name  = "tf-ua-key-a"
}

resource "acecloud_key_pair" "ua_key_b" {
  count = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0
  name  = "tf-ua-key-b"
}

# Two SGs — for SG swap test
resource "acecloud_security_group" "ua_sg_web" {
  count       = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0
  name        = "tf-ua-sg-web"
  description = "Web access"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 80
    port_range_max   = 80
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

resource "acecloud_security_group" "ua_sg_restricted" {
  count       = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0
  name        = "tf-ua-sg-restricted"
  description = "Restricted access"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "10.0.0.0/8"
    ethertype        = "IPv4"
  }
}

# ═══════════════════════════════════════════════════════════════
# UA1: Instance ForceNew — Key Rotation
# ═══════════════════════════════════════════════════════════════
# Create: instance with key_a
# Update: change to key_b → triggers destroy+recreate
# Tests ForceNew lifecycle and dependent resource handling.

resource "acecloud_instance" "ua1_instance" {
  count = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                  = "tf-ua1-instance"
  flavor_id             = var.flavor_id
  boot_uuid             = var.image_id
  source_type           = "image"
  delete_on_termination = true
  billing_type          = "hourly"

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.ua_vpc[0].id]
  security_group_ids = [var.action_phase == "update" ? acecloud_security_group.ua_sg_restricted[0].id : acecloud_security_group.ua_sg_web[0].id]
  key_name           = var.action_phase == "update" ? acecloud_key_pair.ua_key_b[0].name : acecloud_key_pair.ua_key_a[0].name
}

# ═══════════════════════════════════════════════════════════════
# UA2: Volume Snapshot → Restore Chain
# ═══════════════════════════════════════════════════════════════
# Create: volume → snapshot → restored volume from snapshot
# Verifies the full backup-restore workflow. Tests bug #49 fix.

resource "acecloud_volume" "ua2_source" {
  count = var.run_action_tests ? 1 : 0

  name        = "tf-ua2-source-vol"
  size        = 10
  volume_type = "ssd"
  description = "Source volume for snapshot restore test"

  metadata = {
    purpose = "snapshot-source"
  }
}

resource "acecloud_snapshot" "ua2_snap" {
  count = var.run_action_tests ? 1 : 0

  name        = "tf-ua2-snapshot"
  volume_id   = acecloud_volume.ua2_source[0].id
  description = "Snapshot for restore test"
}

resource "acecloud_volume" "ua2_restored" {
  count = var.run_action_tests ? 1 : 0

  name        = "tf-ua2-restored-vol"
  size        = 10
  volume_type = "ssd"
  snapshot_id = acecloud_snapshot.ua2_snap[0].id
  # NO description, NO metadata — tests that source fields are NOT injected
}

# ═══════════════════════════════════════════════════════════════
# UA3: LB Backend Scale Out — Add Second Member
# ═══════════════════════════════════════════════════════════════
# Create: LB + listener + pool + 1 member
# Update: Add a second member to the existing pool

resource "acecloud_load_balancer" "ua3_lb" {
  count = var.run_action_tests ? 1 : 0

  name      = "tf-ua3-lb"
  subnet_id = var.flavor_id != "" && var.image_id != "" ? acecloud_vpc.ua_vpc[0].subnet_id : acecloud_vpc.test.subnet_id
  tags      = ["ALB"]
}

resource "acecloud_lb_listener" "ua3_listener" {
  count = var.run_action_tests ? 1 : 0

  name            = "tf-ua3-http"
  loadbalancer_id = acecloud_load_balancer.ua3_lb[0].id
  protocol        = "HTTP"
  protocol_port   = 80
}

resource "acecloud_lb_pool" "ua3_pool" {
  count = var.run_action_tests ? 1 : 0

  name            = "tf-ua3-pool"
  listener_id     = acecloud_lb_listener.ua3_listener[0].id
  loadbalancer_id = acecloud_load_balancer.ua3_lb[0].id
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
}

resource "acecloud_lb_pool_member" "ua3_member_a" {
  count = var.run_action_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.ua3_pool[0].id
  name          = "tf-ua3-backend-a"
  address       = "10.92.0.10"
  protocol_port = 8080
  weight        = 1
}

resource "acecloud_lb_pool_member" "ua3_member_b" {
  count = var.run_action_tests && var.action_phase == "update" ? 1 : 0

  pool_id       = acecloud_lb_pool.ua3_pool[0].id
  name          = "tf-ua3-backend-b"
  address       = "10.92.0.11"
  protocol_port = 8080
  weight        = 1
}

resource "acecloud_lb_health_monitor" "ua3_hm" {
  count = var.run_action_tests ? 1 : 0

  name        = "tf-ua3-hm"
  pool_id     = acecloud_lb_pool.ua3_pool[0].id
  type        = "HTTP"
  delay       = 5
  timeout     = 3
  max_retries = 3
  url_path    = "/health"
}

# ═══════════════════════════════════════════════════════════════
# UA4: Multi-Resource Update — VPC + SG + Volume in One Apply
# ═══════════════════════════════════════════════════════════════
# Create: VPC + SG + Volume with initial config
# Update: Change all three simultaneously in a single apply

resource "acecloud_vpc" "ua4_vpc" {
  count = var.run_action_tests ? 1 : 0

  name                   = var.action_phase == "update" ? "tf-ua4-vpc-v2" : "tf-ua4-vpc"
  admin_state_up         = true
  subnet_name            = var.action_phase == "update" ? "tf-ua4-subnet-v2" : "tf-ua4-subnet"
  subnet_cidr            = "10.91.0.0/24"
  subnet_ip_version      = 4
  subnet_dns_nameservers = var.action_phase == "update" ? ["1.1.1.1", "1.0.0.1"] : ["8.8.8.8", "8.8.4.4"]
}

resource "acecloud_security_group" "ua4_sg" {
  count = var.run_action_tests ? 1 : 0

  name        = var.action_phase == "update" ? "tf-ua4-sg-v2" : "tf-ua4-sg"
  description = var.action_phase == "update" ? "Tightened rules" : "Initial rules"

  dynamic "rules" {
    for_each = var.action_phase == "update" ? [
      { direction = "ingress", protocol = "tcp", port_range_min = 443, port_range_max = 443, remote_ip_prefix = "10.0.0.0/8", ethertype = "IPv4" },
    ] : [
      { direction = "ingress", protocol = "tcp", port_range_min = 22, port_range_max = 22, remote_ip_prefix = "0.0.0.0/0", ethertype = "IPv4" },
      { direction = "ingress", protocol = "tcp", port_range_min = 80, port_range_max = 80, remote_ip_prefix = "0.0.0.0/0", ethertype = "IPv4" },
      { direction = "ingress", protocol = "tcp", port_range_min = 443, port_range_max = 443, remote_ip_prefix = "0.0.0.0/0", ethertype = "IPv4" },
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

resource "acecloud_volume" "ua4_vol" {
  count = var.run_action_tests ? 1 : 0

  name        = var.action_phase == "update" ? "tf-ua4-vol-v2" : "tf-ua4-vol"
  size        = var.action_phase == "update" ? 20 : 10
  volume_type = "ssd"
  description = var.action_phase == "update" ? "Expanded storage" : "Initial storage"

  metadata = var.action_phase == "update" ? {
    version = "v2"
    env     = "staging"
  } : {
    version = "v1"
    env     = "dev"
  }
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "ua1_instance_id" {
  value = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ua1_instance[0].id : ""
}

output "ua1_instance_key" {
  value = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ua1_instance[0].key_name : ""
}

output "ua1_instance_sg" {
  value = var.run_action_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.ua1_instance[0].security_group_ids : []
}

output "ua2_restored_vol_id" {
  value = var.run_action_tests ? acecloud_volume.ua2_restored[0].id : ""
}

output "ua3_lb_id" {
  value = var.run_action_tests ? acecloud_load_balancer.ua3_lb[0].id : ""
}

output "ua3_member_b_id" {
  value = var.run_action_tests && var.action_phase == "update" ? acecloud_lb_pool_member.ua3_member_b[0].id : ""
}

output "ua4_vpc_name" {
  value = var.run_action_tests ? acecloud_vpc.ua4_vpc[0].name : ""
}

output "ua4_sg_name" {
  value = var.run_action_tests ? acecloud_security_group.ua4_sg[0].name : ""
}

output "ua4_vol_size" {
  value = var.run_action_tests ? acecloud_volume.ua4_vol[0].size : 0
}

output "ua4_vol_metadata" {
  value = var.run_action_tests ? acecloud_volume.ua4_vol[0].metadata : {}
}
