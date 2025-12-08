resource "scc_subaccount_k8s_service_channel" "scc_sc" {
  region_host      = "cf.eu12.hana.ondemand.com"
  subaccount       = "12345678-90ab-cdef-1234-567890abcdef"
  k8s_cluster_host = "cp.app.cluster.kyma.ondemand.com"
  k8s_service_id   = "12345678-90ab-cdef-1234-567890abcdef"
  local_port       = 3000
  connections      = 1
  enabled          = true
}