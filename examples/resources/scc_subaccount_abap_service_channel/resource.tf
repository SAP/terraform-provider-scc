resource "scc_subaccount_abap_service_channel" "scc_sc" {
  region_host            = "cf.eu12.hana.ondemand.com"
  subaccount             = "12345678-90ab-cdef-1234-567890abcdef"
  abap_cloud_tenant_host = "<serviceinstanceguid>.abap.<region>.hana.ondemand.com"
  instance_number        = 20
  connections            = 1
  enabled                = true
}