data "acecloud_security_groups" "all" {}

output "all_sg_names" {
  value = [for sg in data.acecloud_security_groups.all.security_groups : sg.name]
}
