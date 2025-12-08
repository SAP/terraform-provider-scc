resource "scc_domain_mapping" "scc_dm" {
  region_host     = "cf.eu12.hana.ondemand.com"
  subaccount      = "12345678-90ab-cdef-1234-567890abcdef"
  virtual_domain  = "my.virtual.domain.com"
  internal_domain = "my.internal.domain.com"
}