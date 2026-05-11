resource "acecloud_floating_ip" "main" {
  floating_network_id = var.external_network_id
  description         = "Public IP for production web server"
}
