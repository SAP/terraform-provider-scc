package endpoints

import "net/url"

func GetDomainMappingEndpoint(regionHost, subaccount, internalDomain string) string {
	return GetDomainMappingBaseEndpoint(regionHost, subaccount) + "/" + url.PathEscape(internalDomain)
}

func GetDomainMappingBaseEndpoint(regionHost, subaccount string) string {
	return GetSubaccountBaseEndpoint() + "/" + url.PathEscape(regionHost) + "/" + url.PathEscape(subaccount) + "/domainMappings"
}
