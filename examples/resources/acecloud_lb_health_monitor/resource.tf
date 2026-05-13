resource "acecloud_lb_health_monitor" "http" {
  name           = "http-healthcheck"
  pool_id        = acecloud_lb_pool.http.id
  type           = "HTTP"
  delay          = 10
  timeout        = 5
  max_retries    = 3
  url_path       = "/health"
  expected_codes = "200"
  http_method    = "GET"
}
