resource "acecloud_volume_backup" "weekly" {
  name        = "weekly-backup"
  volume_id   = acecloud_volume.data.id
  description = "Weekly portable backup of data volume"
}
