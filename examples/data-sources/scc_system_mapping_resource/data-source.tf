data "scc_system_mapping_resource" "by_resource_id" {
  region_host   = "cf.eu12.hana.ondemand.com"
  subaccount    = "12345678-90ab-cdef-1234-567890abcdef"
  virtual_host  = "virtual.example.com"
  virtual_port  = "443"
}