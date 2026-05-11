data "acecloud_routers" "all" {}

output "all_router_names" {
  value = [for r in data.acecloud_routers.all.routers : r.name]
}
