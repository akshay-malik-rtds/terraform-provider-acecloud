terraform {
  required_providers {
    acecloud = {
      source = "registry.terraform.io/akshay-malik-rtds/acecloud"
    }
  }
}

provider "acecloud" {
  api_url    = var.api_url
  api_token  = var.api_token
  region     = var.region
  project_id = var.project_id
}

# ── Variables ─────────────────────────────────────────────

variable "api_url" {
  description = "Ace Cloud API base URL (npc-api)"
  type        = string
}

variable "api_token" {
  description = "JWT Bearer token for Ace Cloud API"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "Cloud region (e.g. mumbai, noida, atlanta, delhi)"
  type        = string
  default     = "mumbai"
}

variable "project_id" {
  description = "AceCloud project UUID"
  type        = string
}

# ── Test 1: VPC ───────────────────────────────────────────

resource "acecloud_vpc" "test" {
  name           = "tf-acceptance-test-vpc"
  description    = "Terraform acceptance test"
  admin_state_up = true

  # Backend creates the VPC together with an inline subnet.
  subnet_name        = "tf-acceptance-test-vpc-subnet"
  subnet_cidr        = "192.168.99.0/24"
  subnet_ip_version  = 4
  subnet_enable_dhcp = true
}

# ── Test 2: Subnet ────────────────────────────────────────

resource "acecloud_subnet" "test" {
  name       = "tf-acceptance-test-subnet"
  cidr       = "192.168.100.0/24"
  vpc_id     = acecloud_vpc.test.id
  ip_version = 4

  enable_dhcp     = true
  gateway_ip      = "192.168.100.1"
  dns_nameservers = ["8.8.8.8", "8.8.4.4"]

  allocation_pools {
    start = "192.168.100.10"
    end   = "192.168.100.250"
  }
}

# ── Test 3: Security Group ────────────────────────────────

resource "acecloud_security_group" "test" {
  name        = "tf-acceptance-test-sg"
  description = "Terraform acceptance test security group"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "egress"
    protocol         = "tcp"
    port_range_min   = 1
    port_range_max   = 65535
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

# ── Test 4: Key Pair ──────────────────────────────────────

resource "acecloud_key_pair" "test" {
  name = "tf-acceptance-test-key"
  # Omitting public_key lets the backend generate one
}

# ── Test 5: Volume ────────────────────────────────────────

resource "acecloud_volume" "test" {
  name        = "tf-acceptance-test-vol"
  size        = 10
  volume_type = "ssd"
  description = "Terraform acceptance test volume"

  metadata = {
    managed_by = "terraform"
    test       = "true"
  }
}

# ── Outputs ───────────────────────────────────────────────

output "vpc_id" {
  value = acecloud_vpc.test.id
}

output "subnet_id" {
  value = acecloud_subnet.test.id
}

output "security_group_id" {
  value = acecloud_security_group.test.id
}

output "key_pair_id" {
  value = acecloud_key_pair.test.id
}

output "volume_id" {
  value = acecloud_volume.test.id
}
