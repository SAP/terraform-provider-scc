# terraform import scc_subaccount_k8s_service_channel.<resource_name> '<region_host>,<subaccount>,<id>`

terraform import scc_subaccount_k8s_service_channel.scc_sc 'cf.eu12.hana.ondemand.com,12345678-90ab-cdef-1234-567890abcdef,1'

# terraform import using id attribute in import block
import {
  to = scc_subaccount_k8s_service_channel.<resource_name>
  id = "<region_host>,<subaccount>,<id>"
}

# this resource supports import using identity attribute from Terraform version 1.12 or higher
import {
  to = scc_subaccount_k8s_service_channel.<resource_name>
  identity = {
    region_host = "<region_host>"
    subaccount  = "<subaccount>"
    id          = "<id>"
  }
}