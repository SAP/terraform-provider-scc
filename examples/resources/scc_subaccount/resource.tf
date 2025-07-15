resource "scc_subaccount" "scc_sa" {
  region_host = "cf.eu12.hana.ondemand.com"
  subaccount = "12345678-90ab-cdef-1234-567890abcdef"
  cloud_user = "Cloud Username"
  cloud_password = "Cloud Password"
  display_name = "Subaccount_Terraform"
  description = "Description for Subaccount added via Terraform."
}