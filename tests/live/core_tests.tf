# ═══════════════════════════════════════════════════════════════
# Core Cloud Resource Tests — Compute, Networking, Storage
# ═══════════════════════════════════════════════════════════════
# Covers untested operations on core resources.
# Gate: var.run_core_tests (default false)

variable "run_core_tests" {
  description = "Set to true to run core compute/networking/storage tests"
  type        = bool
  default     = false
}

# ═══════════════════════════════════════════════════════════════
# C1: Volume Clone (source_volid)
# ═══════════════════════════════════════════════════════════════
# Clone an existing volume using source_volid.

resource "acecloud_volume" "c1_source" {
  count = var.run_core_tests ? 1 : 0

  name        = "tf-core-clone-source"
  size        = 10
  volume_type = "ssd"
  description = "C1 source volume for cloning"
}

resource "acecloud_volume" "c1_clone" {
  count = var.run_core_tests ? 1 : 0

  name         = "tf-core-clone-dest"
  size         = 10
  volume_type  = "ssd"
  source_volid = acecloud_volume.c1_source[0].id
  description  = "C1 cloned volume from source_volid"

  depends_on = [acecloud_volume.c1_source]
}

# ═══════════════════════════════════════════════════════════════
# C2: VPC with admin_state_up variations
# ═══════════════════════════════════════════════════════════════

resource "acecloud_vpc" "c2_vpc" {
  count = var.run_core_tests ? 1 : 0

  name                   = "tf-core-vpc-admin"
  admin_state_up         = true
  port_security_enabled  = true
  subnet_name            = "tf-core-subnet-admin"
  subnet_cidr            = "10.60.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_gateway_ip      = "10.60.0.1"
  subnet_dns_nameservers = ["8.8.8.8"]
}

# ═══════════════════════════════════════════════════════════════
# C3: Subnet with description and custom allocation pool
# ═══════════════════════════════════════════════════════════════

resource "acecloud_subnet" "c3_custom" {
  count = var.run_core_tests ? 1 : 0

  name            = "tf-core-subnet-custom"
  vpc_id          = acecloud_vpc.c2_vpc[0].id
  cidr            = "10.60.1.0/24"
  ip_version      = 4
  enable_dhcp     = true
  gateway_ip      = "10.60.1.1"
  dns_nameservers = ["1.1.1.1", "8.8.8.8"]
  description     = "C3 subnet with custom allocation pool and DNS"

  allocation_pools {
    start = "10.60.1.10"
    end   = "10.60.1.100"
  }

  depends_on = [acecloud_vpc.c2_vpc]
}

# ═══════════════════════════════════════════════════════════════
# C4: Subnet with DHCP disabled
# ═══════════════════════════════════════════════════════════════

resource "acecloud_subnet" "c4_no_dhcp" {
  count = var.run_core_tests ? 1 : 0

  name        = "tf-core-subnet-no-dhcp"
  vpc_id      = acecloud_vpc.c2_vpc[0].id
  cidr        = "10.60.2.0/24"
  ip_version  = 4
  enable_dhcp = false
  gateway_ip  = "10.60.2.1"

  depends_on = [acecloud_subnet.c3_custom]
}

# ═══════════════════════════════════════════════════════════════
# C5: Router with external gateway add
# ═══════════════════════════════════════════════════════════════

resource "acecloud_router" "c5_with_gw" {
  count = var.run_core_tests && var.external_network_id != "" ? 1 : 0

  name                        = "tf-core-router-gw"
  admin_state_up              = true
  external_gateway_network_id = var.external_network_id
}

# Router interface to connect VPC subnet to router
resource "acecloud_router_interface" "c5_ri" {
  count = var.run_core_tests && var.external_network_id != "" ? 1 : 0

  router_id = acecloud_router.c5_with_gw[0].id
  subnet_id = acecloud_vpc.c2_vpc[0].subnet_id

  depends_on = [acecloud_router.c5_with_gw, acecloud_vpc.c2_vpc]
}

# ═══════════════════════════════════════════════════════════════
# C6: Security Group — Rule removal test (starts with 3, will update to 1)
# ═══════════════════════════════════════════════════════════════

resource "acecloud_security_group" "c6_mutable" {
  count = var.run_core_tests ? 1 : 0

  name        = "tf-core-sg-mutable"
  description = "C6 SG for rule add-remove testing"

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
}

# ═══════════════════════════════════════════════════════════════
# C7: Instance with all optional fields
# ═══════════════════════════════════════════════════════════════
# Tests: description, metadata, config_drive, user_data, key_name, billing_type
# All non-default options that a production user would set.

resource "acecloud_instance" "c7_full_opts" {
  count = var.run_core_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-core-full-opts"
  description = "C7 instance with every optional field set"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  billing_type          = "hourly"
  delete_on_termination = true
  config_drive          = true
  availability_zone     = "nova"

  # Base64 of: #!/bin/bash\necho hello > /tmp/tf-core-test.txt
  user_data = "IyEvYmluL2Jhc2gKZWNobyBoZWxsbyA+IC90bXAvdGYtY29yZS10ZXN0LnR4dA=="

  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  network_ids        = [acecloud_vpc.c2_vpc[0].id]
  security_group_ids = [acecloud_security_group.c6_mutable[0].id]
  key_name           = acecloud_key_pair.generated.name

  metadata = {
    managed_by  = "terraform"
    environment = "core-test"
    purpose     = "full-options-test"
  }

  depends_on = [acecloud_router_interface.c5_ri]
}

# ═══════════════════════════════════════════════════════════════
# C8: Instance with multiple volumes (boot + 2 data)
# ═══════════════════════════════════════════════════════════════

resource "acecloud_instance" "c8_multi_vol" {
  count = var.run_core_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name        = "tf-core-multi-vol-3"
  description = "C8 instance with 1 boot and 2 data volumes"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  billing_type          = "monthly"
  delete_on_termination = true

  # Boot volume
  volumes {
    size         = 20
    boot         = true
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  # Data volume 1 — hourly
  volumes {
    size         = 30
    boot         = false
    volume_type  = "ssd"
    billing_type = "hourly"
  }

  # Data volume 2 — monthly
  volumes {
    size         = 50
    boot         = false
    volume_type  = "ssd"
    billing_type = "monthly"
  }

  network_ids        = [acecloud_vpc.c2_vpc[0].id]
  security_group_ids = [acecloud_security_group.c6_mutable[0].id]
  key_name           = acecloud_key_pair.generated.name

  depends_on = [acecloud_router_interface.c5_ri]
}

# ═══════════════════════════════════════════════════════════════
# C9: Floating IP + Association on core VPC stack
# ═══════════════════════════════════════════════════════════════
# Tests FIP lifecycle on a VPC with proper external gateway routing.

resource "acecloud_floating_ip" "c9_fip" {
  count = var.run_core_tests && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  description         = "C9 floating IP on core VPC with gateway"
}

resource "acecloud_floating_ip_association" "c9_assoc" {
  count = var.run_core_tests && var.external_network_id != "" && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  floating_ip_address = acecloud_floating_ip.c9_fip[0].floating_ip_address
  instance_id         = acecloud_instance.c7_full_opts[0].id

  depends_on = [acecloud_floating_ip.c9_fip, acecloud_instance.c7_full_opts]
}

# ═══════════════════════════════════════════════════════════════
# C10: Volume lifecycle — create, snapshot, backup from same vol
# ═══════════════════════════════════════════════════════════════

resource "acecloud_volume" "c10_lifecycle" {
  count = var.run_core_tests ? 1 : 0

  name         = "tf-core-lifecycle-vol"
  size         = 10
  volume_type  = "ssd"
  billing_type = "monthly"
  description  = "C10 volume for full storage lifecycle"

  metadata = {
    managed_by = "terraform"
    tier       = "production"
  }
}

resource "acecloud_snapshot" "c10_snap" {
  count = var.run_core_tests ? 1 : 0

  name        = "tf-core-lifecycle-snap"
  volume_id   = acecloud_volume.c10_lifecycle[0].id
  description = "C10 snapshot from lifecycle volume"

  depends_on = [acecloud_volume.c10_lifecycle]
}

resource "acecloud_volume_backup" "c10_backup" {
  count = var.run_core_tests ? 1 : 0

  name        = "tf-core-lifecycle-backup"
  volume_id   = acecloud_volume.c10_lifecycle[0].id
  description = "C10 backup from lifecycle volume"

  depends_on = [acecloud_snapshot.c10_snap]
}

# ═══════════════════════════════════════════════════════════════
# C11: LB full stack with weighted members on core VPC
# ═══════════════════════════════════════════════════════════════

resource "acecloud_load_balancer" "c11_lb" {
  count = var.run_core_tests ? 1 : 0

  name        = "tf-core-lb"
  description = "C11 LB on core VPC"
  subnet_id   = acecloud_vpc.c2_vpc[0].subnet_id
  tags        = ["ALB"]
}

resource "acecloud_lb_listener" "c11_http" {
  count = var.run_core_tests ? 1 : 0

  name            = "tf-core-listener"
  protocol        = "HTTP"
  protocol_port   = 80
  loadbalancer_id = acecloud_load_balancer.c11_lb[0].id
}

resource "acecloud_lb_pool" "c11_pool" {
  count = var.run_core_tests ? 1 : 0

  name            = "tf-core-pool"
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
  listener_id     = acecloud_lb_listener.c11_http[0].id
  loadbalancer_id = acecloud_load_balancer.c11_lb[0].id

  depends_on = [acecloud_lb_listener.c11_http]
}

resource "acecloud_lb_pool_member" "c11_m1" {
  count = var.run_core_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.c11_pool[0].id
  name          = "tf-core-member-1"
  address       = "10.60.0.10"
  protocol_port = 8080
  weight        = 10

  depends_on = [acecloud_lb_pool.c11_pool]
}

resource "acecloud_lb_pool_member" "c11_m2" {
  count = var.run_core_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.c11_pool[0].id
  name          = "tf-core-member-2"
  address       = "10.60.0.11"
  protocol_port = 8080
  weight        = 5

  depends_on = [acecloud_lb_pool_member.c11_m1]
}

resource "acecloud_lb_pool_member" "c11_m3" {
  count = var.run_core_tests ? 1 : 0

  pool_id       = acecloud_lb_pool.c11_pool[0].id
  name          = "tf-core-member-3"
  address       = "10.60.0.12"
  protocol_port = 8080
  weight        = 1

  depends_on = [acecloud_lb_pool_member.c11_m2]
}

resource "acecloud_lb_health_monitor" "c11_hm" {
  count = var.run_core_tests ? 1 : 0

  name           = "tf-core-hm"
  pool_id        = acecloud_lb_pool.c11_pool[0].id
  type           = "HTTP"
  delay          = 5
  timeout        = 3
  max_retries    = 3
  url_path       = "/healthz"
  expected_codes = "200-299"
  http_method    = "GET"

  depends_on = [acecloud_lb_pool_member.c11_m3]
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "c1_clone_id" {
  value = var.run_core_tests ? acecloud_volume.c1_clone[0].id : ""
}

output "c2_vpc_id" {
  value = var.run_core_tests ? acecloud_vpc.c2_vpc[0].id : ""
}

output "c3_subnet_id" {
  value = var.run_core_tests ? acecloud_subnet.c3_custom[0].id : ""
}

output "c4_subnet_no_dhcp_id" {
  value = var.run_core_tests ? acecloud_subnet.c4_no_dhcp[0].id : ""
}

output "c5_router_id" {
  value = var.run_core_tests && var.external_network_id != "" ? acecloud_router.c5_with_gw[0].id : ""
}

output "c7_instance_id" {
  value = var.run_core_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.c7_full_opts[0].id : ""
}

output "c7_instance_status" {
  value = var.run_core_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.c7_full_opts[0].status : ""
}

output "c8_instance_id" {
  value = var.run_core_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.c8_multi_vol[0].id : ""
}

output "c9_fip_address" {
  value = var.run_core_tests && var.external_network_id != "" ? acecloud_floating_ip.c9_fip[0].floating_ip_address : ""
}

output "c10_volume_id" {
  value = var.run_core_tests ? acecloud_volume.c10_lifecycle[0].id : ""
}

output "c10_snap_id" {
  value = var.run_core_tests ? acecloud_snapshot.c10_snap[0].id : ""
}

output "c10_backup_id" {
  value = var.run_core_tests ? acecloud_volume_backup.c10_backup[0].id : ""
}

output "c11_lb_id" {
  value = var.run_core_tests ? acecloud_load_balancer.c11_lb[0].id : ""
}
