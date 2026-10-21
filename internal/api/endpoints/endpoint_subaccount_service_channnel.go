package endpoints

import (
	"fmt"
	"net/url"
)

func GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, serviceChannelType string) string {
	return "/api/v1/configuration/subaccounts/" +
		url.PathEscape(regionHost) + "/" +
		url.PathEscape(subaccount) + "/channels/" +
		url.PathEscape(serviceChannelType)
}

func GetSubaccountServiceChannelEndpoint(regionHost, subaccount, serviceChannelType string, id int64) string {
	return fmt.Sprintf(GetSubaccountServiceChannelBaseEndpoint(regionHost, subaccount, serviceChannelType)+"/%d", id)
}
