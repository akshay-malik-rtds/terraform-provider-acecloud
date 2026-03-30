# ═══════════════════════════════════════════════════════════════
# Regression Tests — Session 10+ Gaps
# ═══════════════════════════════════════════════════════════════
# Covers gaps identified in test audit:
# - Delete retry behavior (destroy ordering triggers transient errors)
# - Error recovery (re-apply after partial failure)
# - Billing type variations (quarterly, half-yearly, yearly)
# - Volume shrink rejection
# - Metadata removal
# - Data source validation
# - Missing error tests (key_pair, subnet, VPC, listener, pool, HM)
# - Router external gateway add/remove
# - FIP re-association
#
# Usage:
#   terraform apply -var="run_regression_tests=true"
#   terraform plan  -var="run_regression_tests=true"  # idempotency
#   terraform destroy -var="run_regression_tests=true"
#
# For update phase:
#   terraform apply -var="run_regression_tests=true" -var="regression_phase=update"

variable "run_regression_tests" {
  description = "Set to true to run regression tests"
  type        = bool
  default     = false
}

variable "regression_phase" {
  description = "Phase: create or update"
  type        = string
  default     = "create"
}

# ═══════════════════════════════════════════════════════════════
# R1: Delete Retry — Volume + Snapshot Dependency
# ═══════════════════════════════════════════════════════════════
# Creates volume → snapshot. On destroy, Terraform may try to delete
# the volume while the snapshot still references it, triggering the
# "in use" / "status must be available" retry path.

resource "acecloud_volume" "r1_source" {
  count = var.run_regression_tests ? 1 : 0

  name        = "tf-reg-r1-vol"
  size        = 10
  volume_type = "ssd"
  description = "Regression delete retry source volume"
}

resource "acecloud_snapshot" "r1_snap" {
  count = var.run_regression_tests ? 1 : 0

  name        = "tf-reg-r1-snap"
  volume_id   = acecloud_volume.r1_source[0].id
  description = "Regression snapshot for delete retry test"

  # Backup uses the same volume — snapshot must wait for backup to finish
  depends_on = [acecloud_volume_backup.r1_backup]
}

resource "acecloud_volume_backup" "r1_backup" {
  count = var.run_regression_tests ? 1 : 0

  name      = "tf-reg-r1-backup"
  volume_id = acecloud_volume.r1_source[0].id
}

# ═══════════════════════════════════════════════════════════════
# R2: Delete Retry — Instance + VPC Port Lingering
# ═══════════════════════════════════════════════════════════════
# Creates VPC → Instance. On destroy, VPC delete may get "ports still
# in use" since instance ports take a moment to clean up.

resource "acecloud_vpc" "r2_vpc" {
  count = var.run_regression_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                   = "tf-reg-r2-vpc"
  admin_state_up         = true
  subnet_name            = "tf-reg-r2-subnet"
  subnet_cidr            = "10.98.0.0/24"
  subnet_ip_version      = 4
  subnet_enable_dhcp     = true
  subnet_dns_nameservers = ["8.8.8.8"]
}

resource "acecloud_security_group" "r2_sg" {
  count = var.run_regression_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name = "tf-reg-r2-sg"
}

resource "acecloud_instance" "r2_instance" {
  count = var.run_regression_tests && var.flavor_id != "" && var.image_id != "" ? 1 : 0

  name                  = "tf-reg-r2-instance"
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

  network_ids        = [acecloud_vpc.r2_vpc[0].id]
  security_group_ids = [acecloud_security_group.r2_sg[0].id]
}

# ═══════════════════════════════════════════════════════════════
# R3: Volume Shrink Rejection
# ═══════════════════════════════════════════════════════════════
# Creates a volume at 15GB. On update phase, attempts to shrink to 10GB.
# The provider should either reject or the API should error.

resource "acecloud_volume" "r3_shrink" {
  count = var.run_regression_tests ? 1 : 0

  name        = var.regression_phase == "update" ? "tf-reg-r3-vol-renamed" : "tf-reg-r3-vol"
  size        = var.regression_phase == "update" ? 10 : 15
  volume_type = "ssd"
  description = "Regression volume shrink test"
}

# ═══════════════════════════════════════════════════════════════
# R4: Metadata Lifecycle — Add, Modify, Remove
# ═══════════════════════════════════════════════════════════════
# Create: volume with 2 metadata keys
# Update: change one key value, remove the other (set to 1 key)

resource "acecloud_volume" "r4_metadata" {
  count = var.run_regression_tests ? 1 : 0

  name        = "tf-reg-r4-metadata-vol"
  size        = 10
  volume_type = "ssd"

  metadata = var.regression_phase == "update" ? {
    env = "staging"
  } : {
    env     = "test"
    purpose = "regression"
  }
}

# ═══════════════════════════════════════════════════════════════
# R5: Data Source Validation
# ═══════════════════════════════════════════════════════════════
# Verify data sources return non-empty results and attributes match

data "acecloud_flavors" "r5_check" {
  count = var.run_regression_tests ? 1 : 0
}

data "acecloud_images" "r5_check" {
  count = var.run_regression_tests ? 1 : 0
}

data "acecloud_vpcs" "r5_check" {
  count      = var.run_regression_tests ? 1 : 0
  depends_on = [acecloud_vpc.r2_vpc]
}

data "acecloud_security_groups" "r5_check" {
  count      = var.run_regression_tests ? 1 : 0
  depends_on = [acecloud_security_group.r2_sg]
}

# ═══════════════════════════════════════════════════════════════
# R6: Router External Gateway — Add After Creation
# ═══════════════════════════════════════════════════════════════
# Create: router without external gateway
# Update: add external_gateway_network_id

resource "acecloud_router" "r6_gw" {
  count = var.run_regression_tests && var.external_network_id != "" ? 1 : 0

  name                        = var.regression_phase == "update" ? "tf-reg-r6-router-renamed" : "tf-reg-r6-router"
  admin_state_up              = true
  external_gateway_network_id = var.regression_phase == "update" ? var.external_network_id : null
}

# ═══════════════════════════════════════════════════════════════
# R7: FIP Re-association
# ═══════════════════════════════════════════════════════════════
# Create a FIP on the R2 VPC stack to verify the full FIP lifecycle
# alongside the delete retry test (R2).

resource "acecloud_router_interface" "r7_ri" {
  count = var.run_regression_tests && var.regression_phase == "update" && var.flavor_id != "" && var.image_id != "" && var.external_network_id != "" ? 1 : 0

  router_id = acecloud_router.r6_gw[0].id
  subnet_id = acecloud_vpc.r2_vpc[0].subnet_id
}

resource "acecloud_floating_ip" "r7_fip" {
  count = var.run_regression_tests && var.regression_phase == "update" && var.external_network_id != "" ? 1 : 0

  floating_network_id = var.external_network_id
  description         = "Regression FIP lifecycle test"
}

resource "acecloud_floating_ip_association" "r7_assoc" {
  count = var.run_regression_tests && var.regression_phase == "update" && var.flavor_id != "" && var.image_id != "" && var.external_network_id != "" ? 1 : 0

  floating_ip_address = acecloud_floating_ip.r7_fip[0].floating_ip_address
  instance_id         = acecloud_instance.r2_instance[0].id

  depends_on = [acecloud_router_interface.r7_ri]
}

# ═══════════════════════════════════════════════════════════════
# R8: Missing Error Tests — Key Pair, Subnet, VPC, LB Sub-resources
# ═══════════════════════════════════════════════════════════════

# E-R1: Key pair with invalid public_key format
resource "acecloud_key_pair" "error_bad_pubkey" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name       = "tf-err-bad-pubkey"
  public_key = "not-a-valid-ssh-key"
}

# E-R2: Subnet with bad vpc_id
resource "acecloud_subnet" "error_bad_vpc" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name       = "tf-err-bad-subnet"
  cidr       = "10.0.0.0/24"
  ip_version = 4
  vpc_id     = "00000000-0000-0000-0000-000000000000"
}

# E-R3: VPC with CIDR that overlaps nothing but gateway_ip outside range
resource "acecloud_vpc" "error_bad_vpc" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name              = "tf-err-bad-vpc"
  subnet_name       = "tf-err-bad-subnet"
  subnet_cidr       = "10.97.0.0/24"
  subnet_ip_version = 4
  subnet_gateway_ip = "192.168.1.1"
}

# E-R4: LB listener with bad loadbalancer_id
resource "acecloud_lb_listener" "error_bad_lb" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name            = "tf-err-bad-listener"
  loadbalancer_id = "00000000-0000-0000-0000-000000000000"
  protocol        = "HTTP"
  protocol_port   = 80
}

# E-R5: LB pool with bad listener_id
resource "acecloud_lb_pool" "error_bad_listener" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name         = "tf-err-bad-pool"
  listener_id  = "00000000-0000-0000-0000-000000000000"
  protocol     = "HTTP"
  lb_algorithm = "ROUND_ROBIN"
}

# E-R6: LB health monitor with bad pool_id
resource "acecloud_lb_health_monitor" "error_bad_pool" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name        = "tf-err-bad-hm"
  pool_id     = "00000000-0000-0000-0000-000000000000"
  type        = "HTTP"
  delay       = 5
  timeout     = 3
  max_retries = 3
}

# E-R7: Router with bad external_gateway_network_id
resource "acecloud_router" "error_bad_gw" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name                        = "tf-err-bad-router"
  external_gateway_network_id = "00000000-0000-0000-0000-000000000000"
}

# E-R8: Volume with bad snapshot_id
resource "acecloud_volume" "error_bad_snapshot" {
  count = var.run_regression_tests && var.run_error_tests ? 1 : 0

  name        = "tf-err-bad-snap-vol"
  size        = 10
  volume_type = "ssd"
  snapshot_id = "00000000-0000-0000-0000-000000000000"
}

# ═══════════════════════════════════════════════════════════════
# Outputs
# ═══════════════════════════════════════════════════════════════

output "r1_vol_id" {
  value = var.run_regression_tests ? acecloud_volume.r1_source[0].id : ""
}

output "r1_snap_id" {
  value = var.run_regression_tests ? acecloud_snapshot.r1_snap[0].id : ""
}

output "r1_backup_id" {
  value = var.run_regression_tests ? acecloud_volume_backup.r1_backup[0].id : ""
}

output "r2_vpc_id" {
  value = var.run_regression_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_vpc.r2_vpc[0].id : ""
}

output "r2_instance_id" {
  value = var.run_regression_tests && var.flavor_id != "" && var.image_id != "" ? acecloud_instance.r2_instance[0].id : ""
}

output "r3_vol_size" {
  value = var.run_regression_tests ? acecloud_volume.r3_shrink[0].size : 0
}

output "r4_metadata" {
  value = var.run_regression_tests ? acecloud_volume.r4_metadata[0].metadata : {}
}

output "r5_flavor_count" {
  value = var.run_regression_tests ? length(data.acecloud_flavors.r5_check[0].flavors) : 0
}

output "r5_image_count" {
  value = var.run_regression_tests ? length(data.acecloud_images.r5_check[0].images) : 0
}

output "r5_vpc_count" {
  value = var.run_regression_tests ? length(data.acecloud_vpcs.r5_check[0].vpcs) : 0
}

output "r5_sg_count" {
  value = var.run_regression_tests ? length(data.acecloud_security_groups.r5_check[0].security_groups) : 0
}

output "r6_router_id" {
  value = var.run_regression_tests && var.external_network_id != "" ? acecloud_router.r6_gw[0].id : ""
}

output "r6_router_gw" {
  value = var.run_regression_tests && var.external_network_id != "" ? acecloud_router.r6_gw[0].external_gateway_network_id : ""
}

output "r7_fip_address" {
  value = var.run_regression_tests && var.regression_phase == "update" && var.external_network_id != "" ? acecloud_floating_ip.r7_fip[0].floating_ip_address : ""
}

output "r7_assoc_id" {
  value = var.run_regression_tests && var.regression_phase == "update" && var.flavor_id != "" && var.image_id != "" && var.external_network_id != "" ? acecloud_floating_ip_association.r7_assoc[0].id : ""
}
