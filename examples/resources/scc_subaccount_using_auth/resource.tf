resource "scc_subaccount_using_auth" "scc_sa_auth" {
  authentication_data = file("${path.module}/authentication.data")
  display_name        = "Subaccount_Terraform"
  description         = "Description for Subaccount added via Terraform."
  // Following attributes are applicable for Cloud Connector version 2.19
  auto_certificate_renewal = true
  is_managed               = true
}