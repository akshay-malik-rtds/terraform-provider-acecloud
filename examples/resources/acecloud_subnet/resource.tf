resource "acecloud_subnet" "secondary" {
  name       = "secondary-subnet"
  cidr       = "10.0.1.0/24"
  vpc_id     = acecloud_vpc.main.id
  ip_version = 4

  enable_dhcp     = true
  gateway_ip      = "10.0.1.1"
  dns_nameservers = ["8.8.8.8", "1.1.1.1"]

  allocation_pools {
    start = "10.0.1.10"
    end   = "10.0.1.250"
  }
}
