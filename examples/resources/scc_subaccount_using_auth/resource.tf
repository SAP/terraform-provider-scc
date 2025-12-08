resource "scc_subaccount_using_auth" "scc_sa_auth" {
  authentication_data = file("${path.module}/authentication.data")
  display_name        = "Subaccount_Terraform"
  description         = "Description for Subaccount added via Terraform."
}