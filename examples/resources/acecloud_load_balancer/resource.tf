resource "acecloud_load_balancer" "main" {
  name        = "app-lb"
  description = "Application load balancer for web tier"
  subnet_id   = acecloud_vpc.main.subnet_id
  tags        = ["ALB"]
}
