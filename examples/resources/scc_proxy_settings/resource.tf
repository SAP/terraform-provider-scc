resource "scc_proxy_settings" "proxy" {
  host     = "proxy.company.com"
  port     = 8080
  user     = "proxy-user"
  password = "password"
}