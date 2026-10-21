package endpoints

import "net/url"

func GetSystemMappingBaseEndpoint(regionHost, subaccount string) string {
	return "/api/v1/configuration/subaccounts/" +
		url.PathEscape(regionHost) + "/" +
		url.PathEscape(subaccount) + "/systemMappings"
}

func GetSystemMappingEndpoint(regionHost, subaccount, virtualHost, virtualPort string) string {
	return GetSystemMappingBaseEndpoint(regionHost, subaccount) +
		"/" + url.PathEscape(virtualHost) + ":" + url.PathEscape(virtualPort)
}
