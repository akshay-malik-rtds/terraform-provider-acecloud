terraform {
  required_providers {
    acecloud = {
      source  = "registry.terraform.io/acecloud/acecloud"
      version = "~> 0.1"
    }
  }
}

provider "acecloud" {}

resource "acecloud_security_group" "database" {
  name        = "database-sg"
  description = "Security group for database tier"

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 5432
    port_range_max   = 5432
    remote_ip_prefix = "10.0.0.0/8"
    ethertype        = "IPv4"
  }

  rules {
    direction        = "ingress"
    protocol         = "tcp"
    port_range_min   = 3306
    port_range_max   = 3306
    remote_ip_prefix = "10.0.0.0/8"
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

output "security_group_id" {
  value = acecloud_security_group.database.id
}
