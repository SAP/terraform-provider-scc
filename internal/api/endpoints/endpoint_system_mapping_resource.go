package endpoints

import "net/url"

func GetSystemMappingResourceBaseEndpoint(regionHost, subaccount, virtualHost, virtualPort string) string {
	return GetSystemMappingEndpoint(regionHost, subaccount, virtualHost, virtualPort) + "/resources"
}

func GetSystemMappingResourceEndpoint(regionHost, subaccount, virtualHost, virtualPort, resourceID string) string {
	return GetSystemMappingResourceBaseEndpoint(regionHost, subaccount, virtualHost, virtualPort) +
		"/" + url.PathEscape(resourceID)
}
