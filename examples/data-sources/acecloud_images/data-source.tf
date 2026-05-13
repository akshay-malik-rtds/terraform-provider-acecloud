data "acecloud_images" "all" {}

# Find Ubuntu 22.04 image
output "ubuntu_image_id" {
  value = [for i in data.acecloud_images.all.images : i.id if can(regex("Ubuntu 22.04", i.name))][0]
}
