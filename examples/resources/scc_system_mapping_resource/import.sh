# terraform import scc_system_mapping_resource.<resource_name> '<region_host>,<subaccount>,<virtual_host>,<virtual_port>,<url_path>`

terraform import scc_system_mapping_resource.scc_smr 'cf.eu12.hana.ondemand.com,12345678-90ab-cdef-1234-567890abcdef,virtual.example.com,443,/'

# this resource supports import using identity attribute from Terraform version 1.12 or higher

import {
  to = scc_system_mapping_resource.<resource_name>
  identity = {
    region_host  = "<region_host>"
    subaccount   = "<subaccount>"
    virtual_port = "<virtual_port>"
    virtual_host = "<virtual_host>"
    url_path     = "<url_path>"
  }
}