resource "acecloud_lb_pool_member" "backend1" {
  pool_id       = acecloud_lb_pool.http.id
  address       = "10.0.0.10"
  protocol_port = 8080
  name          = "backend-1"
  weight        = 1
}

resource "acecloud_lb_pool_member" "backend2" {
  pool_id       = acecloud_lb_pool.http.id
  address       = "10.0.0.11"
  protocol_port = 8080
  name          = "backend-2"
  weight        = 1
}
