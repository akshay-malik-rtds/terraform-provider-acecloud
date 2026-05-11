resource "acecloud_router" "main" {
  name                        = "production-router"
  admin_state_up              = true
  external_gateway_network_id = var.external_network_id
}
