# terraform import scc_subaccount_abap_service_channel.<resource_name> '<region_host>,<subaccount>,<id>`

terraform import scc_subaccount_abap_service_channel.scc_sc 'cf.eu12.hana.ondemand.com,12345678-90ab-cdef-1234-567890abcdef,1'

# Import an existing SCC subaccount ABAP service channel into Terraform state
# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
  to = "scc_subaccount_abap_service_channel.<resource_name>"
  identity = {
    region_host = "'<region_host>"
    subaccount  = "<subaccount>"
    id          = "<id>"
  }
}
