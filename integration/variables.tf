variable "region_host" {
  type        = string
  description = "The host of the Cloud Connector region that specifies the SAP BTP region where the subaccount will be connected."
  default     = "cf.us10.hana.ondemand.com"
}

variable "subaccount" {
  type        = string
  description = "The unique ID (GUID) of the SAP BTP subaccount to be connected via the Cloud Connector."
}

variable "cloud_user" {
  type        = string
  description = "The user ID for the specified SAP BTP subaccount and region, used to authenticate with the Cloud Connector."
}

variable "cloud_password" {
  type        = string
  description = "The password associated with the cloud user for authenticating with the Cloud Connector."
}

variable "virtual_host" {
  type        = string
  description = "The virtual host name as exposed to consumers of the Cloud Connector mapping."
  default     = "s4h"
}

variable "virtual_port" {
  type        = string
  description = "The virtual port number exposed for the mapped system."
  default     = "500"
}

variable "internal_host" {
  type        = string
  description = "The actual IP or hostname of the internal system behind the Cloud Connector."
  default     = "34.32.203.52"
}

variable "internal_port" {
  type        = string
  description = "The port number of the internal system to which the Cloud Connector connects."
  default     = "50000"
}

variable "virtual_domain" {
  type        = string
  description = "The virtual domain name used by external consumers to reach the mapped system."
  default     = "www1.test-system.cloud"
}

variable "internal_domain" {
  type        = string
  description = "The internal domain name of the system as known within the private network."
  default     = "ecc60.mycompany.corp"
}

variable "k8s_cluster_host" {
  type        = string
  description = "Host name to access the Kubernetes cluster."
  default     = "integration_test_cluster_host"
}

variable "k8s_service_id" {
  type        = string
  description = "Host name providing the service inside of Kubernetes cluster."
  default     = "integration_test_service_id"
}

variable "abap_cloud_tenant_host" {
  type        = string
  description = "Host name to access the Host of ABAP Cloud Tenant."
  default     = "integration_test_abap_cloud_tenant_host"
}
