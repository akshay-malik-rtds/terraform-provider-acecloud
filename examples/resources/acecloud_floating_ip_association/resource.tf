# Floating IP association requires:
#   1. A router with external_gateway_network_id set
#   2. A router interface attaching the router to the same subnet as the instance
# Use depends_on to ensure the route is established before associating.

resource "acecloud_floating_ip_association" "main" {
  floating_ip_address = acecloud_floating_ip.main.floating_ip_address
  instance_id         = acecloud_instance.web.id

  depends_on = [acecloud_router_interface.main]
}
