# Create a VPC
resource "acecloud_vpc" "main" {
  name           = "main-vpc"
  description    = "Primary VPC for application"
  admin_state_up = true

  # The backend creates a VPC together with an inline subnet.
  subnet_name        = "main-subnet"
  subnet_cidr        = "10.0.0.0/24"
  subnet_ip_version  = 4
  subnet_enable_dhcp = true
}

# Create a subnet in the VPC
resource "acecloud_subnet" "app" {
  name       = "app-subnet"
  cidr       = "10.0.1.0/24"
  vpc_id     = acecloud_vpc.main.id
  ip_version = 4

  enable_dhcp    = true
  gateway_ip     = "10.0.1.1"
  dns_nameservers = ["8.8.8.8", "8.8.4.4"]

  allocation_pools {
    start = "10.0.1.10"
    end   = "10.0.1.250"
  }
}

output "vpc_id" {
  value = acecloud_vpc.main.id
}

output "subnet_id" {
  value = acecloud_subnet.app.id
}
