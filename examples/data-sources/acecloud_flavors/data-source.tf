data "acecloud_flavors" "all" {}

# Reference a specific flavor by name
output "small_flavor_id" {
  value = [for f in data.acecloud_flavors.all.flavors : f.id if f.name == "C4i.medium"][0]
}

# Filter to GPU flavors
output "gpu_flavors" {
  value = [for f in data.acecloud_flavors.all.flavors : f.name if f.is_gpu]
}
