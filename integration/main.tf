locals {
  prefix_integration_test            = "integration-test-"
  prefix_integration_test_subaccount = "${local.prefix_integration_test}subaccount"
  disclaimer_description             = "Please don't modify. This is used for integration tests."
}
resource "scc_subaccount" "scc_sa" {
  region_host    = var.region_host
  subaccount     = var.subaccount
  cloud_user     = var.cloud_user
  cloud_password = var.cloud_password
  display_name   = local.prefix_integration_test_subaccount
  description    = local.disclaimer_description
}

resource "scc_system_mapping" "scc_sm" {
  region_host         = scc_subaccount.scc_sa.region_host
  subaccount          = scc_subaccount.scc_sa.subaccount
  virtual_host        = var.virtual_host
  virtual_port        = var.virtual_port
  internal_host       = var.internal_host
  internal_port       = var.internal_port
  protocol            = "HTTP"
  backend_type        = "abapSys"
  authentication_mode = "KERBEROS"
  host_in_header      = "VIRTUAL"
}

resource "scc_system_mapping_resource" "scc_smr" {
  region_host  = scc_subaccount.scc_sa.region_host
  subaccount   = scc_subaccount.scc_sa.subaccount
  virtual_host = scc_system_mapping.scc_sm.virtual_host
  virtual_port = scc_system_mapping.scc_sm.virtual_port
  url_path     = "/"
  enabled      = true
}

resource "scc_domain_mapping" "scc_dm" {
  region_host     = scc_subaccount.scc_sa.region_host
  subaccount      = scc_subaccount.scc_sa.subaccount
  virtual_domain  = var.virtual_domain
  internal_domain = var.internal_domain
}

resource "scc_subaccount_k8s_service_channel" "scc_k8s_sc" {
  region_host      = scc_subaccount.scc_sa.region_host
  subaccount       = scc_subaccount.scc_sa.subaccount
  k8s_cluster_host = var.k8s_cluster_host
  k8s_service_id   = var.k8s_service_id
  local_port       = var.local_port
  connnections     = 1
  enabled          = false
}

resource "scc_subaccount_abap_service_channel" "scc_abap_sc" {
  region_host            = scc_subaccount.scc_sa.region_host
  subaccount             = scc_subaccount.scc_sa.subaccount
  abap_cloud_tenant_host = var.abap_cloud_tenant_host
  instance_number        = 11
  connnections           = 1
  enabled                = false
}