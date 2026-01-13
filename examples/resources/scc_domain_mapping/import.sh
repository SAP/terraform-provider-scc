# terraform import scc_domain_mapping.<resource_name> '<region_host>,<subaccount>,<internal_domain>`

terraform import scc_domain_mapping.scc_dm 'cf.eu12.hana.ondemand.com,12345678-90ab-cdef-1234-567890abcdef,my.internal.domain.com'

# terraform import using id attribute in import block
import {
  to = scc_domain_mapping.<resource_name>
  id = "<region_host>,<subaccount>,<internal_domain>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher
# Import an existing SCC domain mapping into Terraform state
import {
  to = "scc_domain_mapping.<resource_name>"
  identity = {
    region_host     = "<region_host>"
    subaccount      = "<subaccount>"
    internal_domain = "<internal_domain>"
  }
}
