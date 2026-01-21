# - This feature requires Terraform v1.14.0 or later
list "scc_subaccount" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name
}

# List block to discover all SCC subaccounts
list "scc_subaccount" "all" {
  provider = scc
}
