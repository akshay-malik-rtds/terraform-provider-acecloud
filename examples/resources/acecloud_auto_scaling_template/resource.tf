resource "acecloud_auto_scaling_template" "web" {
  name                   = "web-tier-template"
  type                   = "linux"
  description            = "Launch template for web tier auto scaling"
  volume_size            = 40
  vol_del_on_termination = true
  flavor_id              = var.flavor_id
  image_id               = var.image_id
  is_instance_snapshot   = false
  key_name               = acecloud_key_pair.generated.name
  network_id             = acecloud_vpc.main.id
  subnet_id              = acecloud_vpc.main.subnet_id
  security_groups        = [acecloud_security_group.web.id]
}
