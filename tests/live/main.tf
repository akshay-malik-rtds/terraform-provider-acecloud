terraform {
  required_providers {
    acecloud = {
      source = "registry.terraform.io/acecloud/acecloud"
    }
  }
}

# ─── Provider ────────────────────────────────────────────────
# Supports 3 auth methods (tried in order):
#   1. api_token (static JWT or developer token)
#   2. email + password (POST /auth/login)
#   3. ace-cli config (~/.ace/config.json)

provider "acecloud" {
  api_url    = var.api_url
  api_token  = var.api_token
  email      = var.email
  password   = var.password
  region     = var.region
  project_id = var.project_id
}

# ─── Variables ───────────────────────────────────────────────

variable "api_url" {
  description = "Ace Cloud API base URL (npc-api)"
  type        = string
}

variable "api_token" {
  description = "JWT Bearer token (method 1). Leave empty to use email/password or CLI config."
  type        = string
  default     = ""
  sensitive   = true
}

variable "email" {
  description = "Email for login auth (method 2). Used with password."
  type        = string
  default     = ""
}

variable "password" {
  description = "Password for login auth (method 2). Used with email."
  type        = string
  default     = ""
  sensitive   = true
}

variable "region" {
  description = "Cloud region"
  type        = string
  default     = ""
}

variable "project_id" {
  description = "OpenStack project UUID"
  type        = string
  default     = ""
}

# These are needed for instance creation — you provide them after
# running the data-sources-only test first to discover available values.
variable "flavor_id" {
  description = "Compute flavor UUID (get from data.acecloud_flavors)"
  type        = string
  default     = ""
}

variable "image_id" {
  description = "Boot image UUID (get from data.acecloud_images)"
  type        = string
  default     = ""
}

variable "external_network_id" {
  description = "External/public network UUID for floating IPs"
  type        = string
  default     = ""
}

variable "subnet_id" {
  description = "Existing subnet UUID for LB tests (overrides test subnet if set)"
  type        = string
  default     = ""
}

# ═══════════════════════════════════════════════════════════════
# Phase 1: Data Sources (read-only, always safe)
# ═══════════════════════════════════════════════════════════════

data "acecloud_flavors" "all" {}

data "acecloud_images" "all" {}

data "acecloud_vpcs" "all" {}

data "acecloud_security_groups" "all" {}

data "acecloud_routers" "all" {
  depends_on = [acecloud_router.test]
}

# ═══════════════════════════════════════════════════════════════
# Phase 2: Networking — VPC + Subnet
# ═══════════════════════════════════════════════════════════════
# VPC create requires an inline subnet (backend fails without it).
# The VPC resource creates the initial subnet inline, matching UI behavior.

resource "acecloud_vpc" "test" {
  name                  = "tf-live-test-vpc"
  description           = "Terraform live environment test VPC"
  admin_state_up        = true
  port_security_enabled = true

  # Inline subnet (required — backend requires VPC+subnet together)
  subnet_name            = "tf-live-test-subnet"
  subnet_cidr            = "10.99.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_gateway_ip      = "10.99.0.1"
  subnet_dns_nameservers = ["8.8.8.8", "8.8.4.4"]
}

# Additional subnet on the same VPC (uses the standalone subnet resource)
resource "acecloud_subnet" "test" {
  name       = "tf-live-test-subnet-2"
  cidr       = "10.99.1.0/24"
  vpc_id     = acecloud_vpc.test.id
  ip_version = 4

  enable_dhcp     = true
  gateway_ip      = "10.99.1.1"
  dns_nameservers = ["8.8.8.8", "8.8.4.4"]

  allocation_pools {
    start = "10.99.1.10"
    end   = "10.99.1.250"
  }
}

# ═══════════════════════════════════════════════════════════════
# Phase 3: Security Groups
# ═══════════════════════════════════════════════════════════════

# Test 3a: Security group with multiple protocol types
resource "acecloud_security_group" "web" {
  name        = "tf-live-test-sg-web"
  description = "Terraform live test, web server security group"

  # SSH from anywhere
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

  # ICMP (ping)
  rules {
    direction        = "ingress"
    protocol         = "icmp"
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # All outbound TCP
  rules {
    direction        = "egress"
    protocol         = "tcp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # All outbound UDP
  rules {
    direction        = "egress"
    protocol         = "udp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

# Test 3b: Database security group with custom port ranges and CIDR restrictions
resource "acecloud_security_group" "db" {
  name        = "tf-live-test-sg-db"
  description = "Terraform live test, database security group"

  # MySQL from private CIDR only
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 3306
    port_range_max   = 3306
    remote_ip_prefix = "10.0.0.0/8"
    ethertype        = "IPv4"
  }

  # PostgreSQL from private CIDR only
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 5432
    port_range_max   = 5432
    remote_ip_prefix = "10.0.0.0/8"
    ethertype        = "IPv4"
  }

  # Custom port range (e.g. application ports 8000-8100)
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 8000
    port_range_max   = 8100
    remote_ip_prefix = "10.0.0.0/8"
    ethertype        = "IPv4"
  }
}

# ═══════════════════════════════════════════════════════════════
# Phase 4: Key Pairs
# ═══════════════════════════════════════════════════════════════

# Test 4a: Generated key pair (no public_key → backend generates one)
resource "acecloud_key_pair" "generated" {
  name = "tf-live-test-key-generated"
}

# Test 4b: Imported key pair (user provides public_key)
resource "acecloud_key_pair" "imported" {
  name       = "tf-live-test-key-imported"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7KVtBGFgXMWRBAMqz5OT1PPXQB2mK9aaXDkME6L8ZmEFgUCMnqOOaaMqPPA0VFfJy73s9C/5hGmI0oYRhDG3rQF1MYSRj8FRC+D0G0F4DpWSNeSHShRHnEVvMyJdT47u7gDh+FM3kYl0F+oGXVm0XMKBLyN5HE6P5f0sJx4w93kB6hQOkSCbhPbwFe0clsPz2uP5BTaiBS1O5OBYzO7v0Q4TJ3BGFhKM0qRdsVtLCn0e06cFAN3eFVVfGjYVh+50PyBVNaKfttD3Nk07U96dHZRj0oCWfIjFhwZiJcJQMiRV2FfhpMVde5rOgwkn3EjmNcSF0O8694yJr2cAspYLmj3N8tCBJX0Y7DG2penPYaTKpE3gVfn7BFbhvDlFMfdHFLKItv4vUq3iKaEsi+F2j0FQgTFLzFbhHKJH2hoehlgiJYwrCVCRaiNY9K2Wng+j+o9qnfYazkDqNxZc9OYgeQOEaCmeUqf/Wbzsp80L8N0P8RQbOU60L00AZG1U= terraform-test-key"
}

# ═══════════════════════════════════════════════════════════════
# Phase 5: Standalone Volumes
# ═══════════════════════════════════════════════════════════════

# Test 5a: Non-bootable SSD volume (user-friendly type alias)
resource "acecloud_volume" "basic_ssd" {
  name        = "tf-live-test-vol-ssd"
  size        = 10
  volume_type = "ssd"
  description = "Terraform live test, non-bootable SSD volume"

  metadata = {
    managed_by  = "terraform"
    environment = "test"
    purpose     = "live-acceptance-test"
  }
}

# Test 5b: Second SSD volume with different config (no metadata)
resource "acecloud_volume" "basic_ssd2" {
  name         = "tf-live-test-vol-ssd2"
  size         = 15
  volume_type  = "ssd"
  billing_type = "hourly"
  description  = "Terraform live test, second SSD volume without metadata"
}

# Test 5c: Bootable volume from image (requires image_id variable)
resource "acecloud_volume" "bootable" {
  count = var.image_id != "" ? 1 : 0

  name        = "tf-live-test-vol-bootable"
  size        = 20
  volume_type = "ssd"
  description = "Terraform live test, bootable volume from image"
  image_ref   = var.image_id
}

# ═══════════════════════════════════════════════════════════════
# Phase 6: Compute Instance
# ═══════════════════════════════════════════════════════════════
# Only created if flavor_id and image_id are provided.

resource "acecloud_instance" "test" {
  count = var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-live-test-instance"
  description = "Terraform live test compute instance"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  delete_on_termination = true

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = ["8f0b85d7-6517-4a80-8c32-9b7d93006d48"]  # apigw-vpc network
  security_group_ids = [acecloud_security_group.web.id]
  key_name           = acecloud_key_pair.generated.name

  metadata = {
    managed_by  = "terraform"
    environment = "test"
  }

  # NOTE: tags removed — dev4 OpenStack version doesn't support tags on server create
}

# ═══════════════════════════════════════════════════════════════
# Phase 7: Floating IP
# ═══════════════════════════════════════════════════════════════
# Only created if external_network_id is provided.

resource "acecloud_floating_ip" "test" {
  count = var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  description         = "Terraform live test floating IP"
}

# ═══════════════════════════════════════════════════════════════
# Phase 8: Floating IP Association
# ═══════════════════════════════════════════════════════════════
# Requires both a floating IP and an instance to exist.

resource "acecloud_floating_ip_association" "test" {
  count = var.external_network_id != "" && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  floating_ip_address = acecloud_floating_ip.test[0].floating_ip_address
  instance_id         = acecloud_instance.test[0].id
}

# ═══════════════════════════════════════════════════════════════
# Phase 9: Router
# ═══════════════════════════════════════════════════════════════

resource "acecloud_router" "test" {
  name           = "tf-live-test-router"
  admin_state_up = true
}

# ═══════════════════════════════════════════════════════════════
# Phase 10: Router Interface
# ═══════════════════════════════════════════════════════════════
# Uses the second subnet created above (not the inline VPC subnet).

resource "acecloud_router_interface" "test" {
  router_id = acecloud_router.test.id
  subnet_id = acecloud_subnet.test.id
}

# ═══════════════════════════════════════════════════════════════
# Phase 11: Snapshot
# ═══════════════════════════════════════════════════════════════
# Takes a snapshot of the SSD volume.

resource "acecloud_snapshot" "test" {
  name        = "tf-live-test-snapshot"
  volume_id   = acecloud_volume.basic_ssd.id
  description = "Terraform live test, volume snapshot"

  depends_on = [acecloud_volume.basic_ssd]
}

# ═══════════════════════════════════════════════════════════════
# Phase 12: Volume Backup
# ═══════════════════════════════════════════════════════════════
# Backs up the second SSD volume.

resource "acecloud_volume_backup" "test" {
  name        = "tf-live-test-backup"
  volume_id   = acecloud_volume.basic_ssd2.id
  description = "Terraform live test, volume backup"

  depends_on = [acecloud_volume.basic_ssd2]
}

# ═══════════════════════════════════════════════════════════════
# Phase 13: Load Balancer (full stack)
# ═══════════════════════════════════════════════════════════════
# Requires subnet_id. Creates LB → Listener → Pool → Member → Health Monitor.

# Test 13a: Load Balancer
# Uses var.subnet_id if set, otherwise uses the VPC's inline subnet
resource "acecloud_load_balancer" "test" {
  name        = "tf-live-test-lb"
  description = "Terraform live test, application load balancer"
  subnet_id   = var.subnet_id != "" ? var.subnet_id : acecloud_vpc.test.subnet_id
  tags        = ["ALB"]
}

# Test 13b: LB Listener (HTTP on port 80)
resource "acecloud_lb_listener" "test_http" {
  name            = "tf-live-test-listener-http"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.test.id
}

# Test 13c: LB Pool (round-robin HTTP)
resource "acecloud_lb_pool" "test" {
  name            = "tf-live-test-pool"
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
  listener_id     = acecloud_lb_listener.test_http.id
  loadbalancer_id = acecloud_load_balancer.test.id
}

# Test 13d: LB Pool Member (backend server)
resource "acecloud_lb_pool_member" "test" {
  pool_id       = acecloud_lb_pool.test.id
  address       = "10.99.0.100"
  protocol_port = 8080
  name          = "tf-live-test-member"
  weight        = 1
}

# Test 13e: LB Health Monitor (HTTP health check)
resource "acecloud_lb_health_monitor" "test" {
  name           = "tf-live-test-health-monitor"
  pool_id        = acecloud_lb_pool.test.id
  type           = "HTTP"
  delay          = 10
  timeout        = 5
  max_retries    = 3
  url_path       = "/health"
  expected_codes = "200"
  http_method    = "GET"
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

# --- Data Sources ---

output "available_flavors" {
  description = "All available compute flavors"
  value = [for f in data.acecloud_flavors.all.flavors : {
    id     = f.id
    name   = f.name
    vcpus  = f.vcpus
    ram_mb = f.ram
    disk   = f.disk
    gpu    = f.is_gpu
  }]
}

output "available_images" {
  description = "All available boot images"
  value = [for i in data.acecloud_images.all.images : {
    id     = i.id
    name   = i.name
    status = i.status
    format = i.disk_format
  }]
}

output "all_vpcs" {
  description = "All VPCs in the region"
  value = [for v in data.acecloud_vpcs.all.vpcs : {
    id     = v.id
    name   = v.name
    status = v.status
  }]
}

output "all_security_groups" {
  description = "All security groups in the region"
  value = [for sg in data.acecloud_security_groups.all.security_groups : {
    id   = sg.id
    name = sg.name
  }]
}

output "all_routers" {
  description = "All routers in the region"
  value = [for r in data.acecloud_routers.all.routers : {
    id     = r.id
    name   = r.name
    status = r.status
  }]
}

# --- Security Groups ---

output "security_group_web_id" {
  value = acecloud_security_group.web.id
}

output "security_group_db_id" {
  value = acecloud_security_group.db.id
}

# --- Key Pairs ---

output "key_pair_generated_id" {
  value = acecloud_key_pair.generated.id
}

output "key_pair_imported_id" {
  value = acecloud_key_pair.imported.id
}

# --- Volumes ---

output "volume_ssd_id" {
  value = acecloud_volume.basic_ssd.id
}

output "volume_ssd_status" {
  value = acecloud_volume.basic_ssd.status
}

output "volume_ssd2_id" {
  value = acecloud_volume.basic_ssd2.id
}

output "volume_ssd2_status" {
  value = acecloud_volume.basic_ssd2.status
}

output "volume_bootable_id" {
  value = var.image_id != "" ? acecloud_volume.bootable[0].id : "skipped"
}

# --- Instance ---

output "instance_id" {
  value = var.flavor_id != "" && var.image_id != "" ? acecloud_instance.test[0].id : "skipped"
}

output "instance_status" {
  value = var.flavor_id != "" && var.image_id != "" ? acecloud_instance.test[0].status : "skipped"
}

# --- Floating IP ---

output "floating_ip_id" {
  value = var.external_network_id != "" ? acecloud_floating_ip.test[0].id : "skipped"
}

output "floating_ip_address" {
  value = var.external_network_id != "" ? acecloud_floating_ip.test[0].floating_ip_address : "skipped"
}

# --- Floating IP Association ---

output "floating_ip_association_id" {
  value = var.external_network_id != "" && var.flavor_id != "" && var.image_id != "" ? acecloud_floating_ip_association.test[0].id : "skipped"
}

# --- Router ---

output "router_id" {
  value = acecloud_router.test.id
}

output "router_status" {
  value = acecloud_router.test.status
}

# --- VPC + Subnet ---

output "vpc_id" {
  value = acecloud_vpc.test.id
}

output "vpc_subnet_id" {
  value = acecloud_vpc.test.subnet_id
}

output "subnet_id" {
  value = acecloud_subnet.test.id
}

# --- Router Interface ---

output "router_interface_id" {
  value = acecloud_router_interface.test.id
}

# --- Snapshot ---

output "snapshot_id" {
  value = acecloud_snapshot.test.id
}

output "snapshot_status" {
  value = acecloud_snapshot.test.status
}

# --- Volume Backup ---

output "volume_backup_id" {
  value = acecloud_volume_backup.test.id
}

output "volume_backup_status" {
  value = acecloud_volume_backup.test.status
}

# --- Load Balancer Stack ---

output "load_balancer_id" {
  value = acecloud_load_balancer.test.id
}

output "lb_listener_id" {
  value = acecloud_lb_listener.test_http.id
}

output "lb_pool_id" {
  value = acecloud_lb_pool.test.id
}

output "lb_pool_member_id" {
  value = acecloud_lb_pool_member.test.id
}

output "lb_health_monitor_id" {
  value = acecloud_lb_health_monitor.test.id
}
