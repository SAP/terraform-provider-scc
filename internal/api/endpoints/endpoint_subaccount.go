package endpoints

import "net/url"

func GetSubaccountEndpoint(regionHost, subaccount string) string {
	return GetSubaccountBaseEndpoint() + "/" + url.PathEscape(regionHost) + "/" + url.PathEscape(subaccount)
}

func GetSubaccountBaseEndpoint() string {
	return "/api/v1/configuration/subaccounts"
}
