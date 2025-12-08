resource "scc_system_mapping" "scc_sm" {
  region_host         = "cf.eu12.hana.ondemand.com"
  subaccount          = "12345678-90ab-cdef-1234-567890abcdef"
  virtual_host        = "virtual.example.com"
  virtual_port        = "443"
  internal_host       = "internal.example.com"
  internal_port       = "500"
  protocol            = "HTTP"
  backend_type        = "backend"
  authentication_mode = "authentication"
  host_in_header      = "VIRTUAL"
}