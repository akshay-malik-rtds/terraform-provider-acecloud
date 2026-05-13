# Look up available flavors and images
data "acecloud_flavors" "all" {}
data "acecloud_images" "all" {}

# Create a security group for the instance
resource "acecloud_security_group" "web" {
  name        = "web-sg"
  description = "Allow HTTP and SSH"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 22
    port_range_max   = 22
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 80
    port_range_max   = 80
    remote_ip_prefix = "0.0.0.0/0"
    ethertype        = "IPv4"
  }
}

# Create an instance
resource "acecloud_instance" "web_server" {
  name        = "web-server-01"
  description = "Web server instance"
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

  network_ids        = [var.network_id]
  security_group_ids = [acecloud_security_group.web.id]

  key_name = var.key_name

  metadata = {
    environment = "dev"
    managed_by  = "terraform"
  }
}

variable "flavor_id" {
  description = "Compute flavor UUID"
  type        = string
}

variable "image_id" {
  description = "Boot image UUID"
  type        = string
}

variable "network_id" {
  description = "Network UUID"
  type        = string
}

variable "key_name" {
  description = "SSH key pair name"
  type        = string
  default     = ""
}

output "instance_id" {
  value = acecloud_instance.web_server.id
}

output "instance_status" {
  value = acecloud_instance.web_server.status
}
