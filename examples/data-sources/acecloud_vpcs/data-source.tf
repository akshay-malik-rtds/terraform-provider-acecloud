data "acecloud_vpcs" "all" {}

output "all_vpc_names" {
  value = [for v in data.acecloud_vpcs.all.vpcs : v.name]
}
