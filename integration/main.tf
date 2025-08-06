locals {
  prefix_integration_test = "integration-test-"
  prefix_integration_test_subaccount = "${local.prefix_integration_test}subaccount"
  disclaimer_description = "Please don't modify. This is used for integration tests."
}
resource "scc_subaccount" "scc_sa" {
  region_host = var.region_host
  subaccount = var.subaccount
  cloud_user = var.cloud_user
  cloud_password = var.cloud_password
  display_name = local.prefix_integration_test_subaccount
  description = local.disclaimer_description
}
 
resource "scc_system_mapping" "scc_sm" {
  region_host = scc_subaccount.scc_sa.region_host
  subaccount = scc_subaccount.scc_sa.subaccount
  virtual_host = var.virtual_host
  virtual_port = var.virtual_port
  local_host = var.local_host
  local_port = var.local_port
  internal_host = var.internal_host
  internal_port = var.internal_port
  protocol = "HTTP"
  backend_type = "abapSys"
  authentication_mode = "KERBEROS"
  host_in_header = "VIRTUAL"
}

resource "scc_system_mapping_resource" "scc_smr" {
    region_host = scc_subaccount.scc_sa.region_host
    subaccount = scc_subaccount.scc_sa.subaccount
    virtual_host = scc_system_mapping.scc_sm.virtual_host
    virtual_port = scc_system_mapping.scc_sm.virtual_port
    id = "/"
    enabled = true
}

resource "scc_domain_mapping" "scc_dm" {
    region_host = scc_subaccount.scc_sa.region_host
    subaccount = scc_subaccount.scc_sa.subaccount
    virtual_domain = var.virtual_domain
    internal_domain = var.internal_domain
}