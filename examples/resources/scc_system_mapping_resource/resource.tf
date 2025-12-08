resource "scc_system_mapping_resource" "scc_smr" {
  region_host  = "cf.eu12.hana.ondemand.com"
  subaccount   = "12345678-90ab-cdef-1234-567890abcdef"
  virtual_host = "virtual.example.com"
  virtual_port = "443"
  url_path     = "/"
  enabled      = true
}
