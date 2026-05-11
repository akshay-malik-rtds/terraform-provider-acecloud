terraform {
  required_providers {
    acecloud = {
      source  = "registry.terraform.io/akshay-malik-rtds/acecloud"
      version = "~> 0.1"
    }
  }
  required_version = ">= 1.6"
}

provider "acecloud" {
  api_url    = var.api_url
  api_token  = var.api_token
  region     = var.region
  project_id = var.project_id
}

# ─── Variables ────────────────────────────────────────────

variable "api_url" {
  description = "Ace Cloud API URL"
  type        = string
}

variable "api_token" {
  description = "Ace Cloud API token"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "Cloud region"
  type        = string
  default     = "mumbai"
}

variable "project_id" {
  description = "AceCloud project UUID"
  type        = string
}

variable "flavor_id" {
  description = "Compute flavor UUID"
  type        = string
}

variable "image_id" {
  description = "Boot image UUID"
  type        = string
}

# ─── Networking ───────────────────────────────────────────

resource "acecloud_vpc" "main" {
  name           = "prod-vpc"
  description    = "Production VPC"
  admin_state_up = true

  # Backend creates the VPC together with an inline subnet.
  subnet_name        = "prod-vpc-subnet"
  subnet_cidr        = "10.0.0.0/24"
  subnet_ip_version  = 4
  subnet_enable_dhcp = true
}

resource "acecloud_subnet" "app" {
  name       = "prod-app-subnet"
  cidr       = "10.0.1.0/24"
  vpc_id     = acecloud_vpc.main.id
  ip_version = 4

  enable_dhcp     = true
  gateway_ip      = "10.0.1.1"
  dns_nameservers = ["8.8.8.8", "8.8.4.4"]

  allocation_pools {
    start = "10.0.1.10"
    end   = "10.0.1.250"
  }
}

# ─── Security ─────────────────────────────────────────────

resource "acecloud_security_group" "web" {
  name        = "prod-web-sg"
  description = "Web tier security group"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 80
    port_range_max   = 80
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 443
    port_range_max   = 443
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "10.0.0.0/8"
    ethertype        = "IPv4"
  }
}

# ─── Key Pair ─────────────────────────────────────────────

resource "acecloud_key_pair" "deploy" {
  name       = "prod-deploy-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

# ─── Compute ──────────────────────────────────────────────

resource "acecloud_instance" "web" {
  count = 2

  name        = "prod-web-${count.index + 1}"
  description = "Production web server ${count.index + 1}"
  flavor_id   = var.flavor_id
  boot_uuid   = var.image_id
  source_type = "image"

  delete_on_termination = true

  volumes {
    size         = 40
    boot         = true
    volume_type  = "ssd"
    billing_type = "monthly"
  }

  network_ids        = [acecloud_vpc.main.id]
  security_group_ids = [acecloud_security_group.web.id]

  key_name = acecloud_key_pair.deploy.name

  metadata = {
    environment = "production"
    service     = "web"
    managed_by  = "terraform"
  }

  tags = ["production", "web"]
}

# ─── Storage ──────────────────────────────────────────────

resource "acecloud_volume" "data" {
  count = 2

  name        = "prod-data-${count.index + 1}"
  size        = 200
  volume_type = "ssd"
  description = "Data volume for web server ${count.index + 1}"

  metadata = {
    environment = "production"
    managed_by  = "terraform"
  }
}

# ─── Floating IPs ─────────────────────────────────────────

resource "acecloud_floating_ip" "web" {
  count = 2

  floating_network_id = acecloud_vpc.main.id
}

# ─── Outputs ──────────────────────────────────────────────

output "instance_ids" {
  description = "Web server instance IDs"
  value       = acecloud_instance.web[*].id
}

output "vpc_id" {
  description = "VPC ID"
  value       = acecloud_vpc.main.id
}

output "subnet_id" {
  description = "App subnet ID"
  value       = acecloud_subnet.app.id
}

output "floating_ips" {
  description = "Floating IP IDs"
  value       = acecloud_floating_ip.web[*].id
}
