# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "scc_subaccount_k8s_service_channel" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  # Filter configuration defined by the provider
  config {
    # Provider-specific filter arguments...
  }
}

# List block to discover all scc subaccount k8s service channel
# Returns only the resource identities (IDs/Labels) by default.
list "scc_subaccount_k8s_service_channel" "all" {
  provider = scc

  # (Required)
  config {
    region_host = "cf.us10.hana.ondemand.com"
    subaccount  = "3ecb7280-c7d4-4db6-b7da-7af3cdb13505"
  }
}

# List block to discover scc subaccount k8s service channel with full resource details
# Setting include_resource = true returns full resource objects 
list "scc_subaccount_k8s_service_channel" "with_resource" {
  provider = scc
  include_resource = true

  # (Required)
  config {
    region_host = "cf.us10.hana.ondemand.com"
    subaccount  = "3ecb7280-c7d4-4db6-b7da-7af3cdb13505"
  }
}