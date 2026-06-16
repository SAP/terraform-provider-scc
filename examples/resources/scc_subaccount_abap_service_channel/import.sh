# terraform import scc_subaccount_abap_service_channel.<resource_name> '<region_host>,<subaccount>,<type>,<id>`

terraform import scc_subaccount_abap_service_channel.scc_sc 'cf.eu12.hana.ondemand.com,12345678-90ab-cdef-1234-567890abcdef,ABAPCloud,1'

# terraform import using id attribute in import block
import {
  to = scc_subaccount_abap_service_channel.<resource_name>
  id = "<region_host>,<subaccount>,<type>,<id>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher
import {
  to = "scc_subaccount_abap_service_channel.<resource_name>"
  identity = {
    region_host = "<region_host>"
    subaccount  = "<subaccount>"
    type        = "<type>"
    id          = "<id>"
  }
}
