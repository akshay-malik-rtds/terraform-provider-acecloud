# ═══════════════════════════════════════════════════════════════
# Advanced End-User Scenario Tests
# ═══════════════════════════════════════════════════════════════
# Tests real-world usage patterns, cross-resource dependencies,
# and edge cases that end users commonly encounter.
#
# Usage:
#   terraform apply -var="run_advanced_tests=true"
#   terraform plan  -var="run_advanced_tests=true"   # expect: No changes
#   terraform destroy -var="run_advanced_tests=true"

variable "run_advanced_tests" {
  description = "Set to true to run advanced end-user scenario tests"
  type        = bool
  default     = false
}

# ═══════════════════════════════════════════════════════════════
# A1: Full VPC + Multi-Subnet + Router + Gateway Stack
# ═══════════════════════════════════════════════════════════════
# Real-world pattern: User creates a VPC with multiple subnets,
# a router with external gateway, and attaches subnets to it.

resource "acecloud_vpc" "advanced" {
  count = var.run_advanced_tests ? 1 : 0

  name                  = "tf-adv-test-vpc"
  description           = "Advanced test, multi-subnet VPC with router"
  admin_state_up        = true
  port_security_enabled = true

  subnet_name            = "tf-adv-subnet-primary"
  subnet_cidr            = "10.97.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_gateway_ip      = "10.97.0.1"
  subnet_dns_nameservers = ["8.8.8.8", "8.8.4.4"]
}

# Second subnet — application tier
resource "acecloud_subnet" "advanced_app" {
  count = var.run_advanced_tests ? 1 : 0

  name       = "tf-adv-subnet-app"
  cidr       = "10.97.1.0/24"
  vpc_id     = acecloud_vpc.advanced[0].id
  ip_version = 4

  enable_dhcp     = true
  gateway_ip      = "10.97.1.1"
  dns_nameservers = ["8.8.8.8"]

  allocation_pools {
    start = "10.97.1.10"
    end   = "10.97.1.200"
  }
}

# Third subnet — database tier (small range, no external DNS)
resource "acecloud_subnet" "advanced_db" {
  count = var.run_advanced_tests ? 1 : 0

  name       = "tf-adv-subnet-db"
  cidr       = "10.97.2.0/24"
  vpc_id     = acecloud_vpc.advanced[0].id
  ip_version = 4

  enable_dhcp     = true
  gateway_ip      = "10.97.2.1"
  dns_nameservers = ["10.97.0.1"]

  allocation_pools {
    start = "10.97.2.10"
    end   = "10.97.2.50"
  }
}

# Router with external gateway
resource "acecloud_router" "advanced" {
  count = var.run_advanced_tests && var.external_network_id != "" ? 1 : 0

  name                        = "tf-adv-test-router"
  admin_state_up              = true
  external_gateway_network_id = var.external_network_id
}

# Attach app subnet to router
resource "acecloud_router_interface" "advanced_app" {
  count = var.run_advanced_tests && var.external_network_id != "" ? 1 : 0

  router_id = acecloud_router.advanced[0].id
  subnet_id = acecloud_subnet.advanced_app[0].id
}

# Attach db subnet to router (tests multiple interfaces on same router)
resource "acecloud_router_interface" "advanced_db" {
  count = var.run_advanced_tests && var.external_network_id != "" ? 1 : 0

  router_id = acecloud_router.advanced[0].id
  subnet_id = acecloud_subnet.advanced_db[0].id

  depends_on = [acecloud_router_interface.advanced_app]
}

# ═══════════════════════════════════════════════════════════════
# A2: Boot-from-Volume Instance
# ═══════════════════════════════════════════════════════════════
# Real-world pattern: Create a bootable volume from image first,
# then launch an instance from that volume. This is common for
# persistent root disks.

resource "acecloud_volume" "boot_volume" {
  count = var.run_advanced_tests && var.image_id != "" ? 1 : 0

  name        = "tf-adv-boot-volume"
  size        = 30
  volume_type = "ssd"
  description = "Bootable volume from image for instance"
  image_ref   = var.image_id
}

resource "acecloud_instance" "boot_from_volume" {
  count = var.run_advanced_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-adv-boot-from-vol"
  description = "Instance booted from pre-created volume"
  flavor_id   = var.flavor_id
  boot_uuid   = acecloud_volume.boot_volume[0].id
  source_type = "volume"

  delete_on_termination = false

  volumes {
    size         = 30
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = ["8f0b85d7-6517-4a80-8c32-9b7d93006d48"]
  security_group_ids = [acecloud_security_group.advanced_web[0].id]
}

# ═══════════════════════════════════════════════════════════════
# A3: Multi-Tier Security Groups
# ═══════════════════════════════════════════════════════════════
# Real-world pattern: Web tier allows public access,
# App tier only allows from web tier CIDR,
# DB tier only allows from app tier CIDR.

resource "acecloud_security_group" "advanced_web" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-sg-web-tier"
  description = "Web tier, public HTTP HTTPS and SSH"

  # SSH
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # HTTP
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 80
    port_range_max   = 80
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # HTTPS
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 443
    port_range_max   = 443
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # ICMP
  rules {
    direction        = "ingress"
    protocol         = "icmp"
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

resource "acecloud_security_group" "advanced_app" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-sg-app-tier"
  description = "App tier, only from web subnet"

  # App port from web subnet only
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 8080
    port_range_max   = 8080
    remote_ip_prefix = "10.97.0.0/24"
    ethertype        = "IPv4"
  }

  # Health check port from LB
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 8081
    port_range_max   = 8081
    remote_ip_prefix = "10.97.0.0/16"
    ethertype        = "IPv4"
  }
}

resource "acecloud_security_group" "advanced_db" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-sg-db-tier"
  description = "DB tier, only from app subnet"

  # MySQL from app subnet only
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 3306
    port_range_max   = 3306
    remote_ip_prefix = "10.97.1.0/24"
    ethertype        = "IPv4"
  }

  # Redis from app subnet only
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 6379
    port_range_max   = 6379
    remote_ip_prefix = "10.97.1.0/24"
    ethertype        = "IPv4"
  }
}

# ═══════════════════════════════════════════════════════════════
# A4: Full LB Stack with Multiple Backends
# ═══════════════════════════════════════════════════════════════
# Real-world pattern: LB with HTTPS listener, round-robin pool
# with multiple weighted members and a health check.

resource "acecloud_load_balancer" "advanced" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-test-lb"
  description = "Advanced test, multi-member LB"
  subnet_id   = var.subnet_id != "" ? var.subnet_id : acecloud_vpc.advanced[0].subnet_id
  tags        = ["ALB", "production", "terraform"]
}

resource "acecloud_lb_listener" "advanced_http" {
  count = var.run_advanced_tests ? 1 : 0

  name            = "tf-adv-listener-http"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.advanced[0].id
}

resource "acecloud_lb_pool" "advanced" {
  count = var.run_advanced_tests ? 1 : 0

  name            = "tf-adv-pool"
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
  listener_id     = acecloud_lb_listener.advanced_http[0].id
  loadbalancer_id = acecloud_load_balancer.advanced[0].id
}

# Member 1: primary backend (high weight)
resource "acecloud_lb_pool_member" "advanced_primary" {
  count = var.run_advanced_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.advanced[0].id
  address       = "10.97.1.10"
  protocol_port = 8080
  name          = "tf-adv-member-primary"
  weight        = 5
}

# Member 2: secondary backend (lower weight)
resource "acecloud_lb_pool_member" "advanced_secondary" {
  count = var.run_advanced_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.advanced[0].id
  address       = "10.97.1.11"
  protocol_port = 8080
  name          = "tf-adv-member-secondary"
  weight        = 3
}

# Member 3: canary backend (minimal weight)
resource "acecloud_lb_pool_member" "advanced_canary" {
  count = var.run_advanced_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.advanced[0].id
  address       = "10.97.1.12"
  protocol_port = 8080
  name          = "tf-adv-member-canary"
  weight        = 1
}

# Health monitor with custom settings
resource "acecloud_lb_health_monitor" "advanced" {
  count = var.run_advanced_tests ? 1 : 0

  name           = "tf-adv-health-monitor"
  pool_id        = acecloud_lb_pool.advanced[0].id
  type           = "HTTP"
  delay          = 5
  timeout        = 3
  max_retries    = 2
  url_path       = "/healthz"
  expected_codes = "200-299"
  http_method    = "GET"
}

# ═══════════════════════════════════════════════════════════════
# A5: Floating IP + Association + Disassociation
# ═══════════════════════════════════════════════════════════════
# Real-world pattern: Allocate a floating IP and associate
# it with an instance for public access.

resource "acecloud_floating_ip" "advanced" {
  count = var.run_advanced_tests && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  billing_type        = "hourly"
  description         = "Advanced test: public IP for instance"
}

resource "acecloud_floating_ip_association" "advanced" {
  count = var.run_advanced_tests && var.external_network_id != "" && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  floating_ip_address = acecloud_floating_ip.advanced[0].floating_ip_address
  instance_id         = acecloud_instance.boot_from_volume[0].id
}

# ═══════════════════════════════════════════════════════════════
# A6: Volume Snapshot → Backup Chain
# ═══════════════════════════════════════════════════════════════
# Real-world pattern: Create a volume, take a snapshot,
# then create a backup of the same volume.

resource "acecloud_volume" "advanced_data" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-data-volume"
  size        = 10
  volume_type = "ssd"
  description = "Data volume for snapshot and backup tests"

  metadata = {
    managed_by = "terraform"
    purpose    = "data"
    tier       = "standard"
  }
}

resource "acecloud_snapshot" "advanced" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-snapshot"
  volume_id   = acecloud_volume.advanced_data[0].id
  description = "Snapshot of data volume"
}

resource "acecloud_volume_backup" "advanced" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-backup"
  volume_id   = acecloud_volume.advanced_data[0].id
  description = "Backup of data volume"

  depends_on = [acecloud_snapshot.advanced]
}

# ═══════════════════════════════════════════════════════════════
# A7: Edge Cases — Minimal Resources
# ═══════════════════════════════════════════════════════════════
# Test creating resources with only required fields (no optionals).

# Minimal VPC — just name and required subnet fields
resource "acecloud_vpc" "minimal" {
  count = var.run_advanced_tests ? 1 : 0

  name = "tf-adv-minimal-vpc"

  subnet_name       = "tf-adv-minimal-subnet"
  subnet_cidr       = "10.96.0.0/24"
  subnet_ip_version = 4
}

# Minimal security group — no rules
resource "acecloud_security_group" "minimal" {
  count = var.run_advanced_tests ? 1 : 0

  name = "tf-adv-minimal-sg"
}

# Minimal volume — just name, size, type
resource "acecloud_volume" "minimal" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-minimal-vol"
  size        = 10
  volume_type = "ssd"
}

# Minimal router — just name
resource "acecloud_router" "minimal" {
  count = var.run_advanced_tests ? 1 : 0

  name = "tf-adv-minimal-router"
}

# Minimal key pair — auto-generated
resource "acecloud_key_pair" "minimal" {
  count = var.run_advanced_tests ? 1 : 0

  name = "tf-adv-minimal-kp"
}

# ═══════════════════════════════════════════════════════════════
# A8: TCP Load Balancer (non-HTTP protocol)
# ═══════════════════════════════════════════════════════════════
# Tests LB with TCP protocol instead of HTTP.

resource "acecloud_load_balancer" "tcp_test" {
  count = var.run_advanced_tests ? 1 : 0

  name      = "tf-adv-tcp-lb"
  subnet_id = var.subnet_id != "" ? var.subnet_id : acecloud_vpc.advanced[0].subnet_id
  tags      = ["NLB"]
}

resource "acecloud_lb_listener" "tcp_test" {
  count = var.run_advanced_tests ? 1 : 0

  name            = "tf-adv-tcp-listener"
  protocol        = "TCP"
  protocol_port   = 3306
  loadbalancer_id = acecloud_load_balancer.tcp_test[0].id
}

resource "acecloud_lb_pool" "tcp_test" {
  count = var.run_advanced_tests ? 1 : 0

  name            = "tf-adv-tcp-pool"
  protocol        = "TCP"
  lb_algorithm    = "LEAST_CONNECTIONS"
  listener_id     = acecloud_lb_listener.tcp_test[0].id
  loadbalancer_id = acecloud_load_balancer.tcp_test[0].id
}

resource "acecloud_lb_pool_member" "tcp_test" {
  count = var.run_advanced_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.tcp_test[0].id
  address       = "10.97.2.10"
  protocol_port = 3306
  name          = "tf-adv-tcp-member"
  weight        = 1
}

resource "acecloud_lb_health_monitor" "tcp_test" {
  count = var.run_advanced_tests ? 1 : 0

  name        = "tf-adv-tcp-hm"
  pool_id     = acecloud_lb_pool.tcp_test[0].id
  type        = "TCP"
  delay       = 10
  timeout     = 5
  max_retries = 3
}

# ═══════════════════════════════════════════════════════════════
# Advanced Test Outputs
# ═══════════════════════════════════════════════════════════════

output "adv_vpc_id" {
  value = var.run_advanced_tests ? acecloud_vpc.advanced[0].id : "skipped"
}

output "adv_subnet_app_id" {
  value = var.run_advanced_tests ? acecloud_subnet.advanced_app[0].id : "skipped"
}

output "adv_subnet_db_id" {
  value = var.run_advanced_tests ? acecloud_subnet.advanced_db[0].id : "skipped"
}

output "adv_router_id" {
  value = var.run_advanced_tests && var.external_network_id != "" ? acecloud_router.advanced[0].id : "skipped"
}

output "adv_boot_volume_id" {
  value = var.run_advanced_tests && var.image_id != "" ? acecloud_volume.boot_volume[0].id : "skipped"
}

output "adv_instance_id" {
  value = var.run_advanced_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.boot_from_volume[0].id : "skipped"
}

output "adv_floating_ip" {
  value = var.run_advanced_tests && var.external_network_id != "" ? acecloud_floating_ip.advanced[0].floating_ip_address : "skipped"
}

output "adv_lb_id" {
  value = var.run_advanced_tests ? acecloud_load_balancer.advanced[0].id : "skipped"
}

output "adv_lb_members" {
  value = var.run_advanced_tests ? {
    primary   = acecloud_lb_pool_member.advanced_primary[0].id
    secondary = acecloud_lb_pool_member.advanced_secondary[0].id
    canary    = acecloud_lb_pool_member.advanced_canary[0].id
  } : {}
}

output "adv_tcp_lb_id" {
  value = var.run_advanced_tests ? acecloud_load_balancer.tcp_test[0].id : "skipped"
}

output "adv_minimal_vpc_id" {
  value = var.run_advanced_tests ? acecloud_vpc.minimal[0].id : "skipped"
}

output "adv_minimal_sg_id" {
  value = var.run_advanced_tests ? acecloud_security_group.minimal[0].id : "skipped"
}

output "adv_snapshot_id" {
  value = var.run_advanced_tests ? acecloud_snapshot.advanced[0].id : "skipped"
}

output "adv_backup_id" {
  value = var.run_advanced_tests ? acecloud_volume_backup.advanced[0].id : "skipped"
}
