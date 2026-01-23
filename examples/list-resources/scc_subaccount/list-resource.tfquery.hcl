# This feature requires Terraform v1.14.0 or later
list "scc_subaccount" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name
}

# List block to discover all SCC subaccounts
list "scc_subaccount" "all" {
  provider = scc
}

# List block to discover SCC subaccounts filtered by region host
list "scc_subaccount" "by_region" {
  provider = scc

  # (Optional) Filter configuration
  # If region_host is omitted or empty, subaccounts from all regions are returned
  config {
    region_host = "cf.us10.hana.ondemand.com"
  }
}
