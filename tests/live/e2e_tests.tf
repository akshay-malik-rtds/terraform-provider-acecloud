# ═══════════════════════════════════════════════════════════════
# End-to-End Test Cases — Multi-Resource Dependency Chains
# ═══════════════════════════════════════════════════════════════
# Gate: set var.run_e2e_tests = true to create these resources.
# These tests exercise full dependency chains across resource types.

variable "run_e2e_tests" {
  description = "Set to true to run end-to-end test scenarios"
  type        = bool
  default     = false
}

# ═══════════════════════════════════════════════════════════════
# E2E-1: Full Networking + Compute Stack
# ═══════════════════════════════════════════════════════════════
# VPC -> Router -> Router Interface -> SG -> Instance -> FIP -> FIP Assoc
# Tests a complete production-like networking + compute chain.

resource "acecloud_vpc" "e2e_vpc" {
  count = var.run_e2e_tests ? 1 : 0

  name                  = "tf-e2e-vpc"
  description           = "E2E-1 full networking stack VPC"
  admin_state_up        = true
  port_security_enabled = true

  subnet_name            = "tf-e2e-subnet"
  subnet_cidr            = "10.50.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_gateway_ip      = "10.50.0.1"
  subnet_dns_nameservers = ["8.8.8.8", "8.8.4.4"]
}

resource "acecloud_router" "e2e_router" {
  count = var.run_e2e_tests ? 1 : 0

  name                         = "tf-e2e-router"
  admin_state_up               = true
  external_gateway_network_id  = var.external_network_id
}

resource "acecloud_router_interface" "e2e_ri" {
  count = var.run_e2e_tests ? 1 : 0

  router_id = acecloud_router.e2e_router[0].id
  subnet_id = acecloud_vpc.e2e_vpc[0].subnet_id

  depends_on = [acecloud_router.e2e_router, acecloud_vpc.e2e_vpc]
}

resource "acecloud_security_group" "e2e_sg" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-sg"
  description = "E2E-1 security group with SSH and HTTP"

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

  # All outbound
  rules {
    direction        = "egress"
    protocol         = "tcp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

resource "acecloud_instance" "e2e_instance" {
  count = var.run_e2e_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-e2e-instance"
  description = "E2E-1 instance on dedicated VPC"
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

  network_ids        = [acecloud_vpc.e2e_vpc[0].id]
  security_group_ids = [acecloud_security_group.e2e_sg[0].id]
  key_name           = acecloud_key_pair.generated.name

  depends_on = [acecloud_router_interface.e2e_ri]
}

resource "acecloud_floating_ip" "e2e_fip" {
  count = var.run_e2e_tests && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  description         = "E2E-1 floating IP for instance"
}

resource "acecloud_floating_ip_association" "e2e_fip_assoc" {
  count = var.run_e2e_tests && var.external_network_id != "" && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  floating_ip_address = acecloud_floating_ip.e2e_fip[0].floating_ip_address
  instance_id         = acecloud_instance.e2e_instance[0].id

  depends_on = [acecloud_floating_ip.e2e_fip, acecloud_instance.e2e_instance]
}

# --- E2E-1 Outputs ---

output "e2e1_vpc_id" {
  description = "E2E-1: VPC ID"
  value       = var.run_e2e_tests ? acecloud_vpc.e2e_vpc[0].id : ""
}

output "e2e1_router_id" {
  description = "E2E-1: Router ID"
  value       = var.run_e2e_tests ? acecloud_router.e2e_router[0].id : ""
}

output "e2e1_instance_id" {
  description = "E2E-1: Instance ID"
  value       = var.run_e2e_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.e2e_instance[0].id : ""
}

output "e2e1_fip_address" {
  description = "E2E-1: Floating IP address"
  value       = var.run_e2e_tests && var.external_network_id != "" ? acecloud_floating_ip.e2e_fip[0].floating_ip_address : ""
}

# ═══════════════════════════════════════════════════════════════
# E2E-2: Instance with Boot + Data Volume
# ═══════════════════════════════════════════════════════════════
# Tests multi-volume instance with mixed billing types.

resource "acecloud_instance" "e2e_multi_vol" {
  count = var.run_e2e_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-e2e-multi-vol"
  description = "E2E-2 instance with boot and data volumes"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  billing_type          = "monthly"
  delete_on_termination = true

  # Boot volume — 20GB SSD, hourly billing
  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  # Data volume — 50GB SSD, monthly billing
  volumes {
    size         = 50
    boot         = false
    volume_type  = "ssd"
    billing_type = "monthly"
  }

  network_ids        = [acecloud_vpc.test.id]
  security_group_ids = [acecloud_security_group.web.id]
  key_name           = acecloud_key_pair.generated.name
}

# --- E2E-2 Outputs ---

output "e2e2_instance_id" {
  description = "E2E-2: Multi-volume instance ID"
  value       = var.run_e2e_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.e2e_multi_vol[0].id : ""
}

# ═══════════════════════════════════════════════════════════════
# E2E-3: Snapshot -> Volume -> Boot Instance Chain
# ═══════════════════════════════════════════════════════════════
# Volume -> Snapshot -> Restored Volume (larger) -> Instance booting from volume

resource "acecloud_volume" "e2e_chain_vol" {
  count = var.run_e2e_tests && var.image_id != "" ? 1 : 0

  name        = "tf-e2e-chain-vol"
  size        = 20
  volume_type = "ssd"
  image_ref   = var.image_id
  description = "E2E-3 bootable source volume, created from image"
}

resource "acecloud_snapshot" "e2e_chain_snap" {
  count = var.run_e2e_tests && var.image_id != "" ? 1 : 0

  name        = "tf-e2e-chain-snap"
  volume_id   = acecloud_volume.e2e_chain_vol[0].id
  description = "E2E-3 snapshot from source volume"

  depends_on = [acecloud_volume.e2e_chain_vol]
}

resource "acecloud_volume" "e2e_chain_restored" {
  count = var.run_e2e_tests && var.image_id != "" ? 1 : 0

  name        = "tf-e2e-chain-restored"
  size        = 20
  volume_type = "ssd"
  snapshot_id = acecloud_snapshot.e2e_chain_snap[0].id
  description = "E2E-3 volume restored from snapshot, larger than original"

  depends_on = [acecloud_snapshot.e2e_chain_snap]
}

resource "acecloud_instance" "e2e_chain_instance" {
  count = var.run_e2e_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-e2e-chain-instance"
  description = "E2E-3 instance booting from restored volume"
  flavor_id   = var.flavor_id
  boot_uuid   = acecloud_volume.e2e_chain_restored[0].id
  source_type = "volume"

  billing_type          = "hourly"
  delete_on_termination = false

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.test.id]
  security_group_ids = [acecloud_security_group.web.id]
  key_name           = acecloud_key_pair.generated.name

  depends_on = [acecloud_volume.e2e_chain_restored]
}

# --- E2E-3 Outputs ---

output "e2e3_source_vol_id" {
  description = "E2E-3: Source volume ID"
  value       = var.run_e2e_tests && var.image_id != "" ? acecloud_volume.e2e_chain_vol[0].id : ""
}

output "e2e3_snapshot_id" {
  description = "E2E-3: Snapshot ID"
  value       = var.run_e2e_tests && var.image_id != "" ? acecloud_snapshot.e2e_chain_snap[0].id : ""
}

output "e2e3_restored_vol_id" {
  description = "E2E-3: Restored volume ID"
  value       = var.run_e2e_tests && var.image_id != "" ? acecloud_volume.e2e_chain_restored[0].id : ""
}

output "e2e3_instance_id" {
  description = "E2E-3: Boot-from-volume instance ID"
  value       = var.run_e2e_tests && var.flavor_id != "" ? acecloud_instance.e2e_chain_instance[0].id : ""
}

# ═══════════════════════════════════════════════════════════════
# E2E-4: Multi-Tier Security Groups with remote_group_id
# ═══════════════════════════════════════════════════════════════
# Web -> App -> DB tier SG chaining using remote_group_id references.

resource "acecloud_security_group" "e2e_sg_web" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-sg-web"
  description = "E2E-4 web tier, allows HTTP and HTTPS from anywhere"

  # HTTP from anywhere
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 80
    port_range_max   = 80
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # HTTPS from anywhere
  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 443
    port_range_max   = 443
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  # All outbound
  rules {
    direction        = "egress"
    protocol         = "tcp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

resource "acecloud_security_group" "e2e_sg_app" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-sg-app"
  description = "E2E-4 app tier, allows 8080 from web SG only"

  # App port from web tier only (remote_group_id)
  rules {
    direction       = "ingress"
    protocol        = "tcp"
    port_range_min  = 8080
    port_range_max  = 8080
    remote_group_id = acecloud_security_group.e2e_sg_web[0].id
    ethertype       = "IPv4"
  }

  # All outbound
  rules {
    direction        = "egress"
    protocol         = "tcp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

resource "acecloud_security_group" "e2e_sg_db" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-sg-db"
  description = "E2E-4 DB tier, allows 3306 from app SG only"

  # MySQL from app tier only (remote_group_id)
  rules {
    direction       = "ingress"
    protocol        = "tcp"
    port_range_min  = 3306
    port_range_max  = 3306
    remote_group_id = acecloud_security_group.e2e_sg_app[0].id
    ethertype       = "IPv4"
  }

  # All outbound
  rules {
    direction        = "egress"
    protocol         = "tcp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

# --- E2E-4 Outputs ---

output "e2e4_sg_web_id" {
  description = "E2E-4: Web tier SG ID"
  value       = var.run_e2e_tests ? acecloud_security_group.e2e_sg_web[0].id : ""
}

output "e2e4_sg_app_id" {
  description = "E2E-4: App tier SG ID"
  value       = var.run_e2e_tests ? acecloud_security_group.e2e_sg_app[0].id : ""
}

output "e2e4_sg_db_id" {
  description = "E2E-4: DB tier SG ID"
  value       = var.run_e2e_tests ? acecloud_security_group.e2e_sg_db[0].id : ""
}

# ═══════════════════════════════════════════════════════════════
# E2E-5: LB with SOURCE_IP + Weight=0 Drain
# ═══════════════════════════════════════════════════════════════
# LB -> Listener -> Pool (SOURCE_IP) -> Members (active + drained) -> HM

resource "acecloud_load_balancer" "e2e_lb" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-lb-source-ip"
  description = "E2E-5 LB with SOURCE_IP algorithm"
  subnet_id   = var.subnet_id != "" ? var.subnet_id : acecloud_vpc.test.subnet_id
  tags        = ["ALB"]
}

resource "acecloud_lb_listener" "e2e_lb_listener" {
  count = var.run_e2e_tests ? 1 : 0

  name            = "tf-e2e-listener-http"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.e2e_lb[0].id
}

resource "acecloud_lb_pool" "e2e_lb_pool" {
  count = var.run_e2e_tests ? 1 : 0

  name            = "tf-e2e-pool-source-ip"
  protocol        = "HTTP"
  lb_algorithm    = "SOURCE_IP"
  listener_id     = acecloud_lb_listener.e2e_lb_listener[0].id
  loadbalancer_id = acecloud_load_balancer.e2e_lb[0].id

  depends_on = [acecloud_lb_listener.e2e_lb_listener]
}

# Member 1: active (weight=5)
resource "acecloud_lb_pool_member" "e2e_member_active" {
  count = var.run_e2e_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.e2e_lb_pool[0].id
  name          = "tf-e2e-member-active"
  address       = "10.50.0.10"
  protocol_port = 8080
  weight        = 5

  depends_on = [acecloud_lb_pool.e2e_lb_pool]
}

# Member 2: drained (weight=0)
resource "acecloud_lb_pool_member" "e2e_member_drained" {
  count = var.run_e2e_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.e2e_lb_pool[0].id
  name          = "tf-e2e-member-drained"
  address       = "10.50.0.11"
  protocol_port = 8080
  weight        = 0

  depends_on = [acecloud_lb_pool_member.e2e_member_active]
}

# Health Monitor: HTTP HEAD /health
resource "acecloud_lb_health_monitor" "e2e_hm" {
  count = var.run_e2e_tests ? 1 : 0

  name           = "tf-e2e-hm-http"
  pool_id        = acecloud_lb_pool.e2e_lb_pool[0].id
  type           = "HTTP"
  delay          = 10
  timeout        = 5
  max_retries    = 3
  url_path       = "/health"
  expected_codes = "200"
  http_method    = "HEAD"

  depends_on = [acecloud_lb_pool_member.e2e_member_drained]
}

# --- E2E-5 Outputs ---

output "e2e5_lb_id" {
  description = "E2E-5: Load balancer ID"
  value       = var.run_e2e_tests ? acecloud_load_balancer.e2e_lb[0].id : ""
}

output "e2e5_lb_vip" {
  description = "E2E-5: Load balancer VIP address"
  value       = var.run_e2e_tests ? acecloud_load_balancer.e2e_lb[0].vip_address : ""
}

output "e2e5_pool_algorithm" {
  description = "E2E-5: Pool algorithm (should be SOURCE_IP)"
  value       = var.run_e2e_tests ? acecloud_lb_pool.e2e_lb_pool[0].lb_algorithm : ""
}

# ═══════════════════════════════════════════════════════════════
# E2E-6: Multiple Snapshots of Same Volume
# ═══════════════════════════════════════════════════════════════
# Volume -> Snapshot 1 -> Snapshot 2 (sequential)

resource "acecloud_volume" "e2e_multi_snap_vol" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-multi-snap-vol"
  size        = 8
  volume_type = "ssd"
  description = "E2E-6 volume for multiple snapshots"
}

resource "acecloud_snapshot" "e2e_snap_1" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-snap-1"
  volume_id   = acecloud_volume.e2e_multi_snap_vol[0].id
  description = "E2E-6 first snapshot"

  depends_on = [acecloud_volume.e2e_multi_snap_vol]
}

resource "acecloud_snapshot" "e2e_snap_2" {
  count = var.run_e2e_tests ? 1 : 0

  name        = "tf-e2e-snap-2"
  volume_id   = acecloud_volume.e2e_multi_snap_vol[0].id
  description = "E2E-6 second snapshot of same volume"

  depends_on = [acecloud_snapshot.e2e_snap_1]
}

# --- E2E-6 Outputs ---

output "e2e6_volume_id" {
  description = "E2E-6: Source volume ID"
  value       = var.run_e2e_tests ? acecloud_volume.e2e_multi_snap_vol[0].id : ""
}

output "e2e6_snap_1_id" {
  description = "E2E-6: First snapshot ID"
  value       = var.run_e2e_tests ? acecloud_snapshot.e2e_snap_1[0].id : ""
}

output "e2e6_snap_2_id" {
  description = "E2E-6: Second snapshot ID"
  value       = var.run_e2e_tests ? acecloud_snapshot.e2e_snap_2[0].id : ""
}

# ═══════════════════════════════════════════════════════════════
# E2E-7: Data Source -> Resource Chaining
# ═══════════════════════════════════════════════════════════════
# Uses data sources to discover flavor/image, then creates instance.
# Tests the real-world workflow where users don't hardcode UUIDs.

locals {
  # Pick the first available flavor and image from data sources
  e2e_ds_flavor_id = var.run_e2e_tests && length(data.acecloud_flavors.all.flavors) > 0 ? data.acecloud_flavors.all.flavors[0].id : ""
  e2e_ds_image_id  = var.run_e2e_tests && length(data.acecloud_images.all.images) > 0 ? data.acecloud_images.all.images[0].id : ""
}

resource "acecloud_instance" "e2e_ds_instance" {
  count = var.run_e2e_tests && local.e2e_ds_flavor_id != "" && local.e2e_ds_image_id != "" ? 1 : 0

  name        = "tf-e2e-ds-instance"
  description = "E2E-7 instance from data source discovery"
  flavor_id   = local.e2e_ds_flavor_id
  boot_uuid   = local.e2e_ds_image_id
  source_type = "image"

  billing_type          = "hourly"
  delete_on_termination = true

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.test.id]
  security_group_ids = [acecloud_security_group.web.id]
  key_name           = acecloud_key_pair.generated.name
}

# --- E2E-7 Outputs ---

output "e2e7_discovered_flavor_id" {
  description = "E2E-7: Flavor ID from data source"
  value       = var.run_e2e_tests ? local.e2e_ds_flavor_id : ""
}

output "e2e7_discovered_image_id" {
  description = "E2E-7: Image ID from data source"
  value       = var.run_e2e_tests ? local.e2e_ds_image_id : ""
}

output "e2e7_instance_id" {
  description = "E2E-7: Data-source-driven instance ID"
  value       = var.run_e2e_tests && local.e2e_ds_flavor_id != "" && local.e2e_ds_image_id != "" ? acecloud_instance.e2e_ds_instance[0].id : ""
}

# ═══════════════════════════════════════════════════════════════
# E2E-8: Instance with Multiple Security Groups
# ═══════════════════════════════════════════════════════════════
# Reuses E2E-4 SGs (web, app, db) on a single instance.

resource "acecloud_instance" "e2e_multi_sg" {
  count = var.run_e2e_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-e2e-multi-sg"
  description = "E2E-8 instance with web, app, and db security groups"
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

  network_ids = [acecloud_vpc.test.id]
  security_group_ids = [
    acecloud_security_group.e2e_sg_web[0].id,
    acecloud_security_group.e2e_sg_app[0].id,
    acecloud_security_group.e2e_sg_db[0].id,
  ]
  key_name = acecloud_key_pair.generated.name

  depends_on = [
    acecloud_security_group.e2e_sg_web,
    acecloud_security_group.e2e_sg_app,
    acecloud_security_group.e2e_sg_db,
  ]
}

# --- E2E-8 Outputs ---

output "e2e8_instance_id" {
  description = "E2E-8: Multi-SG instance ID"
  value       = var.run_e2e_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.e2e_multi_sg[0].id : ""
}
