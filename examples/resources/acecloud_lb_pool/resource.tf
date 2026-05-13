resource "acecloud_lb_pool" "http" {
  name            = "http-pool"
  protocol        = "HTTP"
  lb_algorithm    = "ROUND_ROBIN"
  listener_id     = acecloud_lb_listener.http.id
  loadbalancer_id = acecloud_load_balancer.main.id
}
