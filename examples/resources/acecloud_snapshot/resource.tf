resource "acecloud_snapshot" "daily" {
  name        = "daily-backup"
  volume_id   = acecloud_volume.data.id
  description = "Daily snapshot of data volume"
}
