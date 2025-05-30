---
page_title: "scc_domain_mapping Resource - scc"
subcategory: ""
description: |-
  Cloud Connector Domain Mapping Resource.
  Tips:
  You must be assigned to the following roles:
  AdministratorSubaccount Administrator
  Further documentation:
  https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/domain-mappings
---

# scc_domain_mapping (Resource)

Cloud Connector Domain Mapping Resource.

__Tips:__
* You must be assigned to the following roles:
	* Administrator
	* Subaccount Administrator

__Further documentation:__
<https://help.sap.com/docs/connectivity/sap-btp-connectivity-cf/domain-mappings>

## Example Usage

```terraform
resource "scc_domain_mapping" "scc_dm" {
    region_host = "cf.eu12.hana.ondemand.com"
    subaccount = "12345678-90ab-cdef-1234-567890abcdef"
    virtual_domain = "my.virtual.domain.com"
    internal_domain = "my.internal.domain.com"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `internal_domain` (String) Domain used on the on-premise side.
- `region_host` (String) Region Host Name.
- `subaccount` (String) The ID of the subaccount.
- `virtual_domain` (String) Domain used on the cloud side.

## Import

Import is supported using the following syntax:

```terraform
# terraform import scc_domain_mapping.<resource_name> '<region_host>,<subaccount>,<internal_domain>`

terraform import scc_domain_mapping.scc_dm 'cf.eu12.hana.ondemand.com,12345678-90ab-cdef-1234-567890abcdef,my.internal.domain.com'
```
