# terraform import scc_subaccount.<resource_name> '<region_host>,<subaccount>`

terraform import scc_subaccount.scc_sa 'cf.eu12.hana.ondemand.com,12345678-90ab-cdef-1234-567890abcdef'

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
  to = scc_subaccount.<resource_name>
  identity = {
    region_host = "<region_host>"
    subaccount  = "<subaccount>"
  }
}