# This feature requires Terraform v1.14.0 or later (Stable as of 2026)
# List resources must be defined in .tfquery.hcl files.

# Generic template for a list block
list "scc_system_mapping_resource" "<label_name>" {
  # (Required) Provider instance to use
  provider = provider_name

  # Filter configuration defined by the provider
  config {
    # Provider-specific filter arguments...
  }
}

# List block to discover system mappings resource
list "scc_system_mapping_resource" "all" {
  provider = scc
  include_resource  = true

  # (Required)
  config {
    region_host  = "cf.ap21.hana.ondemand.com"
    subaccount   = "7ecb7280-c7d4-4db6-b7da-7af3cdb13505"
    virtual_host = "virtual.example.com"
    virtual_port = "443"
  }
}
