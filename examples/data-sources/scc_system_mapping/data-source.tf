data "scc_system_mapping" "by_virtual_host_and_virtual_port" {
  region_host  = "cf.eu12.hana.ondemand.com"
  subaccount   = "12345678-90ab-cdef-1234-567890abcdef"
  virtual_host = "virtual.example.com"
  virtual_port = "443"
}