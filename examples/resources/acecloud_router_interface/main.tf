resource "acecloud_router_interface" "main" {
  router_id = acecloud_router.main.id
  subnet_id = acecloud_vpc.main.subnet_id
}
